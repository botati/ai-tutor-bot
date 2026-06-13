package storage

import (
	"context"
	"log/slog"

	"github.com/cobrich/ai-tutor-bot/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresHistoryStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresHistoryStorage(pool *pgxpool.Pool, logger *slog.Logger) *PostgresHistoryStorage {
	return &PostgresHistoryStorage{
		pool:   pool,
		logger: logger,
	}
}

func (s *PostgresHistoryStorage) Add(
	chatID int64,
	message entity.Message,
) {
	_, err := s.pool.Exec(
		context.Background(),
		`
		INSERT INTO conversations (
			chat_id,
			role,
			content
		)
		VALUES ($1, $2, $3)
		`,
		chatID,
		message.Role,
		message.Content,
	)

	if err != nil {
		s.logger.Error("failed to save conversation", "chat_id", chatID, "error", err)
	}
}

func (s *PostgresHistoryStorage) GetRecent(
	chatID int64,
	limit int,
) []entity.Message {

	rows, err := s.pool.Query(
		context.Background(),
		`
		SELECT role, content
		FROM conversations
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2
		`,
		chatID,
		limit,
	)

	if err != nil {
		s.logger.Error("failed to load history", "chat_id", chatID, "error", err)
		return nil
	}

	defer rows.Close()

	var messages []entity.Message

	for rows.Next() {
		var msg entity.Message

		if err := rows.Scan(
			&msg.Role,
			&msg.Content,
		); err != nil {
			continue
		}

		messages = append(messages, msg)
	}

	// переворачиваем порядок
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages
}

func (s *PostgresHistoryStorage) Clear(chatID int64) {
	_, err := s.pool.Exec(
		context.Background(),
		`
		DELETE FROM conversations
		WHERE chat_id = $1
		`,
		chatID,
	)

	if err != nil {
		s.logger.Error("failed to clear history", "chat_id", chatID, "error", err)
	}
}
