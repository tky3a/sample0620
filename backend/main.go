package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "demo_user"
	password = "demo_password"
	dbname   = "demo_db"
)

type QueryResponse struct {
	ExecutionTimeMs int64       `json:"executionTimeMs"`
	Plan            interface{} `json:"plan"`
	Error           string      `json:"error,omitempty"`
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// MySQLの接続情報
	mysqlDSN := "demo_user:demo_password@tcp(127.0.0.1:3306)/demo_db?parseTime=true"
	dbMysql, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("Error opening MySQL database: %v", err)
	}
	defer dbMysql.Close()

	http.HandleFunc("/api/query", func(w http.ResponseWriter, r *http.Request) {
		// CORS設定（フロントエンドとバックエンドでポートが異なる可能性があるため）
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 実行計画をJSON形式で取得するためのクエリ
		query := "EXPLAIN (FORMAT JSON, ANALYZE) SELECT * FROM user_logs WHERE status = 'error' ORDER BY created_at DESC LIMIT 50;"

		start := time.Now()
		rows, err := db.Query(query)
		if err != nil {
			json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
			return
		}
		defer rows.Close()

		var planJSON string
		if rows.Next() {
			if err := rows.Scan(&planJSON); err != nil {
				json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
				return
			}
		}
		duration := time.Since(start).Milliseconds()

		var plan interface{}
		if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
			json.NewEncoder(w).Encode(QueryResponse{Error: fmt.Sprintf("Failed to parse JSON plan: %v", err)})
			return
		}

		resp := QueryResponse{
			ExecutionTimeMs: duration,
			Plan:            plan,
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/api/clearcache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// デモ用にキャッシュ（共有バッファ）をクリアするため、コンテナごと再起動する
		cmd := exec.Command("podman", "restart", "demo_postgres")
		if err := cmd.Run(); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to restart container: %v", err)})
			return
		}

		// さらに、Podmanの仮想マシンのOSページキャッシュ（ディスクキャッシュ）も完全に破棄する
		// これにより、ディスクI/Oが確実に発生し「重いクエリ」を再現しやすくなります
		cmdCache := exec.Command("podman", "machine", "ssh", "sync && sudo tee /proc/sys/vm/drop_caches <<< 3")
		cmdCache.Run()

		// DBが完全に立ち上がるまで少し待機
		time.Sleep(3 * time.Second)

		json.NewEncoder(w).Encode(map[string]string{"message": "Cache cleared successfully"})
	})

	// === Demo 2: MySQL Implicit Cast Trap ===
	http.HandleFunc("/api/demo2/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 意図的に型変換の罠を起こすため、Goからは int 型で検索する
		// （テーブルの末尾に50件だけ存在する '99' を検索させることでフルスキャンを強制）
		var targetType int = 99

		// 1. 実際の実行時間を計測
		queryStr := "SELECT id, device_type, data FROM devices WHERE device_type = ? LIMIT 50"

		start := time.Now()
		rows, err := dbMysql.Query(queryStr, targetType)
		if err != nil {
			json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
			return
		}
		for rows.Next() {
		} // 結果を読み捨てる
		rows.Close()
		duration := time.Since(start).Milliseconds()

		// 2. AIに分析させるための実行計画（JSON）を取得する
		explainRows, err := dbMysql.Query("EXPLAIN FORMAT=JSON "+queryStr, targetType)
		if err != nil {
			json.NewEncoder(w).Encode(QueryResponse{Error: err.Error()})
			return
		}

		var explainJSON string
		if explainRows.Next() {
			explainRows.Scan(&explainJSON)
		}
		explainRows.Close()

		// MySQL 8.0ではEXPLAIN JSONに型変換の警告が含まれず SHOW WARNINGS に出るため、それも取得して追記する
		var warnings []string
		warningRows, err := dbMysql.Query("SHOW WARNINGS")
		if err == nil {
			for warningRows.Next() {
				var level, code, message string
				if err := warningRows.Scan(&level, &code, &message); err == nil {
					// EXPLAINの警告（型変換エラー等）を抽出
					if level == "Warning" {
						warnings = append(warnings, fmt.Sprintf("[%s %s] %s", level, code, message))
					}
				}
			}
			warningRows.Close()
		}

		var plan map[string]interface{}
		if err := json.Unmarshal([]byte(explainJSON), &plan); err != nil {
			json.NewEncoder(w).Encode(QueryResponse{Error: fmt.Sprintf("Failed to parse JSON plan: %v", err)})
			return
		}

		if len(warnings) > 0 {
			plan["optimizer_warnings"] = warnings
		}

		resp := QueryResponse{
			Plan:            plan,
			ExecutionTimeMs: duration,
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/api/demo2/clearcache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// MySQLコンテナの再起動でキャッシュ（InnoDB Buffer Pool）をクリア
		cmd := exec.Command("podman", "restart", "demo_mysql")
		if err := cmd.Run(); err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to restart container: %v", err)})
			return
		}

		// VM側のディスクキャッシュもクリア
		cmdCache := exec.Command("podman", "machine", "ssh", "sync && sudo tee /proc/sys/vm/drop_caches <<< 3")
		cmdCache.Run()

		time.Sleep(3 * time.Second)
		json.NewEncoder(w).Encode(map[string]string{"message": "Cache cleared successfully"})
	})

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
