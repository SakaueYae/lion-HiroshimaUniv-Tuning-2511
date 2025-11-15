-- このファイルに記述されたSQLコマンドが、マイグレーション時に実行されます。
ALTER TABLE products ADD INDEX idx_name (name, product_id);
ALTER TABLE products ADD INDEX idx_value (value, product_id);
ALTER TABLE products ADD INDEX idx_weight (weight, product_id);

-- DESC（降順）用のインデックスを追加
ALTER TABLE products ADD INDEX idx_name_desc (name DESC, product_id ASC);
ALTER TABLE products ADD INDEX idx_value_desc (value DESC, product_id ASC);
ALTER TABLE products ADD INDEX idx_weight_desc (weight DESC, product_id ASC);

ALTER TABLE products ADD FULLTEXT INDEX idx_fulltext_name_description (name, description);