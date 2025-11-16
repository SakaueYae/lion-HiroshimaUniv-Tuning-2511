-- 商品名検索の高速化のためのインデックス
-- 注文検索（キーワード検索）と商品検索で使用される products.name に対するインデックス
-- 
-- 使用箇所:
-- - /api/v1/orders (注文検索: キーワード部分一致・前方一致)
-- - /api/v1/product (商品検索: キーワード部分一致)
--
-- パフォーマンス改善:
-- - LIKE 'keyword%' (前方一致): インデックスを使用可能
-- - LIKE '%keyword%' (部分一致): フルスキャンだが、インデックスでソート済みデータを高速スキャン
--
-- E2Eテスト対象:
-- - tests/02_orders.test.ts: キーワード検索（部分一致・前方一致）
-- - tests/04_product.test.ts: 商品部分一致検索

ALTER TABLE products ADD INDEX idx_product_name (name);
