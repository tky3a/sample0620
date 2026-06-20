-- テーブル作成
CREATE TABLE user_logs (
    id SERIAL PRIMARY KEY,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    message TEXT
);

-- ダミーデータの生成 (約500万件に増やしてクエリを遅くする)
-- statusは 'info', 'warn', 'error' のいずれか。'error'の割合を少し増やして（20%）ヒット件数を増やす。
-- created_at は過去1年間のランダムな日時
-- message に余分なテキストを入れてレコードサイズを大きくし、スキャンを重くする
INSERT INTO user_logs (status, created_at, message)
SELECT 
    CASE 
        WHEN random() < 0.20 THEN 'error'
        WHEN random() < 0.40 THEN 'warn'
        ELSE 'info'
    END,
    NOW() - (random() * interval '365 days'),
    'Dummy message details: ' || repeat('x', 200) || ' ' || i
FROM generate_series(1, 5000000) AS s(i);

-- 問題の単一インデックス（「WHERE句の条件カラムをインデックスに追加した」という状況）
-- これにより、WHERE句の絞り込みにはインデックスが使われますが、
-- その後の ORDER BY created_at DESC でソートノード（ファイルソート等）が発生し遅くなります。
CREATE INDEX idx_status ON user_logs(status);

-- 【デモ用メモ（解決策）】
-- デモの終盤で以下の複合インデックスを貼ることで速度が劇的に改善します。
-- CREATE INDEX idx_status_created_at ON user_logs(status, created_at DESC);
