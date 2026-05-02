CREATE TABLE rooms (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);

CREATE TABLE workspaces (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    name TEXT NOT NULL
);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id TEXT UNIQUE,
    role TEXT NOT NULL CHECK (role IN ('student', 'admin')),
    email TEXT UNIQUE
);

CREATE INDEX idx_users_role ON users (role);

CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    workspace_id BIGINT NOT NULL REFERENCES workspaces (id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'canceled', 'completed')),
    CONSTRAINT bookings_time_order CHECK (end_time > start_time)
);

CREATE INDEX idx_bookings_workspace_time ON bookings (workspace_id, start_time, end_time);
CREATE INDEX idx_bookings_user_status ON bookings (user_id, status);

CREATE TABLE booking_settings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    max_active INT NOT NULL DEFAULT 3 CHECK (max_active >= 0)
);
