-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS user_access (
    chat_id BIGINT PRIMARY KEY,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    approved_at TIMESTAMP,
    rejected_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_access_status
ON user_access(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_access;

-- +goose StatementEnd