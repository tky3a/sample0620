CREATE TABLE devices (
    id INT AUTO_INCREMENT PRIMARY KEY,
    device_type VARCHAR(10) NOT NULL,
    data TEXT
);

-- フルテーブルスキャンを確実に遅くするため、約300万件のダミーデータを生成
DELIMITER $$
CREATE PROCEDURE InsertDummyData()
BEGIN
    DECLARE i INT DEFAULT 0;
    -- 1000行 × 3000回 = 300万件
    WHILE i < 3000 DO
        INSERT INTO devices (device_type, data)
        WITH RECURSIVE seq AS (
            SELECT 1 AS val UNION ALL SELECT val + 1 FROM seq WHERE val < 1000
        )
        SELECT 
            CAST(FLOOR(1 + RAND() * 5) AS CHAR), -- '1'〜'5'
            CONCAT('Dummy device data padding: ', REPEAT('x', 50))
        FROM seq;
        SET i = i + 1;
    END WHILE;

    -- ★罠のポイント★
    -- 検索対象となる '99' のデータを、テーブルの「一番最後」に50件だけ挿入する。
    -- LIMIT 50 があっても、フルスキャン時に全件舐めないと50件見つからないため確実に数秒かかる。
    SET i = 0;
    WHILE i < 50 DO
        INSERT INTO devices (device_type, data) VALUES ('99', 'Target data');
        SET i = i + 1;
    END WHILE;
END$$
DELIMITER ;

CALL InsertDummyData();

-- DDLで定義された VARCHAR(10) カラムに対する単一インデックス
-- ※Goなどのプログラム側から `int` 型の数値で検索してしまうと暗黙の型変換が走り、
-- このインデックスが使われずにフルテーブルスキャンになってしまうという「罠」の状態。
CREATE INDEX idx_device_type ON devices(device_type);
