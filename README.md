# AI時代のチューニング方法 デモ環境

社内技術共有会用のデモアプリです。
「ORDER BYの罠」をテーマに、AIを用いて実行計画（JSON）を分析し、最適な複合インデックスを提案させるデモを実演します。

## 概要
- **DB**: PostgreSQL 16
- **バックエンド**: Go
- **フロントエンド**: HTML / CSS / Vanilla JS
- **対象データ**: 100万件のログデータ (`user_logs` テーブル)

## 起動手順（データ投入〜画面表示まで）

### 1. データベースの起動とデータ投入
Podman (またはDocker) を使用してPostgreSQLおよびMySQLコンテナを起動します。
初回起動時に `init.sql` と `mysql_init.sql` が読み込まれ、それぞれに大量のダミーデータと罠インデックスが生成されます。

> **Tips: MacでPodmanを使用する場合**
> 仮想マシン（VM）が起動していないと `returned non-zero exit status 125` などのエラーが発生します。その場合は、事前に以下のコマンドでVMを起動してください。
> ```bash
> podman machine start
> ```

```bash
# プロジェクトルートに移動
cd /demo

# コンテナの起動（バックグラウンド）
podman-compose up -d
# ※Dockerを使用している場合は docker-compose up -d
```

> **注意:** 初回起動時は両方のデータベースで合計数百万件のデータ生成が行われるため、APIが応答するまで数分程度かかる場合があります。

### 2. バックエンドAPIサーバーの起動
Goで作成したAPIサーバーを起動します。このサーバーはフロントエンドからのリクエストを受け、DBに `EXPLAIN (FORMAT JSON, ANALYZE)` クエリを発行し、結果を返却します。

別のターミナルタブを開き、以下のコマンドを実行します。

```bash
cd /demo/backend

# サーバー起動
go run main.go
```

ターミナルに `Server is running on http://localhost:8080` と表示されればバックエンドの準備は完了です。

### 3. フロントエンド（デモ画面）の表示
APIサーバーを立ち上げたまま、ブラウザでフロントエンドの画面を開きます。
Macの場合、新しいターミナルタブから以下のコマンドを実行すると、デフォルトのブラウザで開くことができます。

```bash
open /demo/frontend/index.html
```

---

## デモの実施シナリオ（参考）

1. **画面の操作**
   ブラウザで開いたデモ画面の **「Run Query & Get Plan」** ボタンをクリックします。
2. **実行計画の取得**
   データベースでの処理が完了すると、画面上に「実行時間（数百ms〜数秒）」と「実行計画（JSON）」が表示されます。
3. **AIによる分析デモ**
   画面に表示されたJSON文字列をすべてコピーし、ChatGPT等のAIに以下のように入力します。
   > 「以下のPostgreSQLの実行計画（JSON）を分析して、クエリが遅い原因と、最適なDDL（インデックス追加など）を提案してください。」
   > (ここにコピーしたJSONを貼り付ける)
4. **解決策の検証**
   AIが「`Sort`ノードでファイルソートが発生しているため、`CREATE INDEX idx_status_created_at ON user_logs(status, created_at DESC);` のような複合インデックスが必要です」と回答します。
   実際にDBにログインし、インデックスを追加してから再度デモ画面でボタンを押すことで、処理時間が激減することを実演できます。

```bash
# デモ中のインデックス追加用コマンド
podman exec -it demo_postgres psql -U demo_user -d demo_db

# プロンプトが切り替わったら以下を実行
CREATE INDEX idx_status_created_at ON user_logs(status, created_at DESC);
```

> **Tips: デモ1をやり直したい場合（インデックスの削除）**
> 追加した複合インデックスを削除して元の「遅い状態」に戻したい場合は、同じく `psql` のプロンプト内で以下のコマンドを実行してください。その後、画面の「Clear DB Cache」ボタンを押すことで、完全に重い状態に戻ります。
> ```sql
> DROP INDEX idx_status_created_at;
> ```

---

### Demo 2: 暗黙の型変換トラップ（MySQL）の実施シナリオ

1. **AIによる分析**
   取得したJSONをAIに入力すると、「`devices` テーブルの `device_type` が文字列型（VARCHAR）であるのに対し、INT型として結合・検索しているため暗黙の型変換（CAST）が発生し、インデックスが無視されています」といった回答が得られます。
2. **解決策の検証（型の修正）**
   AIの提案に従い、`devices` テーブルの `device_type` カラムを `VARCHAR(10)` から正しい `INT` 型に修正します。これによりインデックスが使用可能になり、処理時間が劇的に改善します。

```bash
# MySQLコンテナに接続
podman exec -it demo_mysql mysql -u demo_user -pdemo_password demo_db

# demo 2 は　Go側のソースコードを変更して型定義を合わせにいく想定

---

## 終了手順

デモが終了したら、以下の手順でコンテナやサーバーを停止してください。

### 1. バックエンドサーバーの停止
APIサーバーを起動しているターミナルで `Ctrl + C` を押してサーバーを停止します。

### 2. データベース（コンテナ）の停止
PostgreSQLのコンテナを停止・削除します。

```bash
cd /demo
podman-compose down
# ※Dockerを使用している場合は docker-compose down
```

> **Tips: MacでPodmanを使用する場合**
> Podmanの仮想マシン（VM）も完全に停止してリソースを解放したい場合は、以下のコマンドを実行します。
> ```bash
> podman machine stop
> ```
