-- ORDER BY句のファイルソートを回避するためのソート用インデックス
-- 各ソートフィールド（product_name以外）に対応

-- shipped_status でのソート用（ASC/DESC両対応）
ALTER TABLE orders ADD INDEX idx_user_status_order_asc (user_id, shipped_status ASC, order_id ASC);
ALTER TABLE orders ADD INDEX idx_user_status_order_desc (user_id, shipped_status DESC, order_id ASC);

-- created_at でのソート用（ASC/DESC両対応）
ALTER TABLE orders ADD INDEX idx_user_created_order_asc (user_id, created_at ASC, order_id ASC);
ALTER TABLE orders ADD INDEX idx_user_created_order_desc (user_id, created_at DESC, order_id ASC);

-- arrived_at でのソート用（ASC/DESC両対応）
ALTER TABLE orders ADD INDEX idx_user_arrived_order_asc (user_id, arrived_at ASC, order_id ASC);
ALTER TABLE orders ADD INDEX idx_user_arrived_order_desc (user_id, arrived_at DESC, order_id ASC);

-- order_id でのソート用（ASC/DESC両対応）
ALTER TABLE orders ADD INDEX idx_user_order_asc (user_id, order_id ASC);
ALTER TABLE orders ADD INDEX idx_user_order_desc (user_id, order_id DESC);
