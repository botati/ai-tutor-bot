-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS conversations (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_chat_id_created_at
ON conversations(chat_id, created_at DESC);

CREATE TABLE IF NOT EXISTS user_limits (
    chat_id BIGINT PRIMARY KEY,
    request_count INTEGER NOT NULL DEFAULT 0,
    last_reset DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE IF NOT EXISTS bot_users (
    chat_id BIGINT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bot_stats (
    key TEXT PRIMARY KEY,
    value BIGINT NOT NULL DEFAULT 0
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS bot_stats;
DROP TABLE IF EXISTS bot_users;
DROP TABLE IF EXISTS user_limits;
DROP TABLE IF EXISTS conversations;

-- +goose StatementEnd