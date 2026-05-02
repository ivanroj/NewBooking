DELETE FROM users WHERE email = 'admin@cowork.local' AND role = 'admin';

ALTER TABLE users DROP COLUMN password_hash;
