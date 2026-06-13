-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS user_stats (
    chat_id BIGINT PRIMARY KEY REFERENCES bot_users(chat_id) ON DELETE CASCADE,
    text_requests BIGINT NOT NULL DEFAULT 0,
    image_requests BIGINT NOT NULL DEFAULT 0,
    voice_requests BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_stats;

-- +goose StatementEnd