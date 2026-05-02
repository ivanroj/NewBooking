ALTER TABLE users
    ADD COLUMN password_hash TEXT;

INSERT INTO users (telegram_id, role, email, password_hash)
VALUES (
        NULL,
        'admin',
        'admin@cowork.local',
        '$2a$10$qOpS3BvVS4.oxT0KysuUY.eXnKrCldmcnSNgrK1.irqAE2cYRUwU7C'
    );
