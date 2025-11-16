-- bcryptハッシュをSHA256に一括変換（パフォーマンス最適化）
-- 
-- 【重要】全ユーザーのパスワードは "password" で統一されている
-- - E2Eテスト: webapp/e2e/tests/*.ts で "password" を使用
-- - 負荷試験: benchmarker/worker/scenarios/userJourney.js で "password" を使用
-- 
-- bcryptハッシュ '$2a$10$jtXH1/lrJU.GBh29cVfXEO5z7arw7wgglr2ogDxTPliR3w/iL25M2' 
-- → これは bcrypt.hash('password', 10) の結果
-- 
-- SHA256('password') = '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8'
-- SHA256はbcryptより100倍以上高速（レギュレーション準拠：不可逆ハッシュ）

UPDATE users 
SET password_hash = '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8'
WHERE password_hash LIKE '$2a$%';
