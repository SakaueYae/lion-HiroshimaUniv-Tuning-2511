-- ORDER BY句のファイルソートを回避するためのソート用インデックス
-- マイグレーション時間短縮のため、最も使用頻度の高い2パターンのみに削減
--
-- 削減による影響:
-- - 他のソートパターンではファイルソートが発生する可能性あり
-- - ただし、user_idでの絞り込み後のデータ量は限定的なため影響は軽微
--
-- 残したインデックス:
-- 1. created_at DESC: 最新の注文から表示する最も一般的なパターン
-- 2. order_id ASC: デフォルトのソート順

-- created_at でのソート用（DESC: 最新順）
ALTER TABLE orders ADD INDEX idx_user_created_order_desc (user_id, created_at DESC, order_id ASC);

-- order_id でのソート用（ASC: デフォルト順）
ALTER TABLE orders ADD INDEX idx_user_order_asc (user_id, order_id ASC);
