CREATE TABLE global_settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO global_settings (key, value) VALUES ('booking_limit', '3');
