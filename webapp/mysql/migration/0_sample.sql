-- このファイルに記述されたSQLコマンドが、マイグレーション時に実行されます。
ALTER TABLE products ADD INDEX idx_name (name, product_id);
ALTER TABLE products ADD INDEX idx_value (value, product_id);
ALTER TABLE products ADD INDEX idx_weight (weight, product_id);

-- DESC（降順）用のインデックスを追加
ALTER TABLE products ADD INDEX idx_name_desc (name DESC, product_id ASC);
ALTER TABLE products ADD INDEX idx_value_desc (value DESC, product_id ASC);
ALTER TABLE products ADD INDEX idx_weight_desc (weight DESC, product_id ASC);

-- ========================================
-- /api/robot/delivery-plan 最適化用インデックス
-- ========================================
-- shipped_statusでのフィルタリングとproduct_idでのJOINを同時に最適化
-- WHERE shipped_status = 'shipping' AND JOIN ON product_id の両方をカバー
ALTER TABLE orders ADD INDEX idx_orders_status_product (shipped_status, product_id);
ALTER TABLE products ADD FULLTEXT INDEX idx_fulltext_name_description (name, description);
