-- このファイルに記述されたSQLコマンドが、マイグレーション時に実行されます。
ALTER TABLE products ADD INDEX idx_name (name, product_id);
ALTER TABLE products ADD INDEX idx_value (value, product_id);
ALTER TABLE products ADD INDEX idx_weight (weight, product_id);

-- DESC（降順）用のインデックスは削除（マイグレーション時間短縮のため）
-- MySQLはASCインデックスを逆順スキャンできるため、DESC専用インデックスは不要
-- ALTER TABLE products ADD INDEX idx_name_desc (name DESC, product_id ASC);
-- ALTER TABLE products ADD INDEX idx_value_desc (value DESC, product_id ASC);
-- ALTER TABLE products ADD INDEX idx_weight_desc (weight DESC, product_id ASC);

-- ========================================
-- /api/robot/delivery-plan 最適化用インデックス
-- ========================================
-- shipped_statusでのフィルタリングとproduct_idでのJOINを同時に最適化
-- WHERE shipped_status = 'shipping' AND JOIN ON product_id の両方をカバー
ALTER TABLE orders ADD INDEX idx_orders_status_product (shipped_status, product_id);
ALTER TABLE products ADD FULLTEXT INDEX idx_fulltext_name_description (name, description) WITH PARSER ngram;

ALTER TABLE users ADD INDEX idx_user_name (user_name);
-- すでにuser_idのインデックスがあるため、重複を避ける
-- ALTER TABLE orders ADD INDEX idx_user_id (user_id);
ALTER TABLE orders ADD INDEX idx_created_at (user_id, created_at);
-- 複合インデックスは削除（マイグレーション時間短縮のため）
-- expires_at単独のインデックス（idx_user_sessions_expires_at）のみで十分
-- ALTER TABLE user_sessions ADD INDEX idx_expires_at (session_uuid, expires_at);

-- ========================================
-- セッションクリーンアップ最適化用インデックス
-- ========================================
-- DELETE FROM user_sessions WHERE expires_at < NOW() を高速化
-- 複合インデックス idx_expires_at (session_uuid, expires_at) では
-- expires_at が第2カラムのため、expires_at 単独の検索には非効率
ALTER TABLE user_sessions ADD INDEX idx_user_sessions_expires_at (expires_at);
