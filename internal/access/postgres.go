package access

import (
	"context"
	"log/slog"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresStorage(pool *pgxpool.Pool, logger *slog.Logger) *PostgresStorage {
	return &PostgresStorage{
		pool:   pool,
		logger: logger,
	}
}

func (s *PostgresStorage) GetStatus(chatID int64) (Status, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var status Status

	err := s.pool.QueryRow(
		ctx,
		`
		SELECT status
		FROM user_access
		WHERE chat_id = $1
		`,
		chatID,
	).Scan(&status)

	if err != nil {
		return "", false
	}

	return status, true
}

func (s *PostgresStorage) RequestAccess(user *tgbotapi.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO user_access (
			chat_id,
			username,
			first_name,
			last_name,
			status,
			requested_at
		)
		VALUES ($1, $2, $3, $4, 'pending', NOW())
		ON CONFLICT (chat_id)
		DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			status = CASE
				WHEN user_access.status = 'approved' THEN 'approved'
				ELSE 'pending'
			END,
			requested_at = CASE
				WHEN user_access.status = 'approved' THEN user_access.requested_at
				ELSE NOW()
			END
		`,
		user.ID,
		user.UserName,
		user.FirstName,
		user.LastName,
	)

	if err != nil {
		s.logger.Error("failed to request access", "chat_id", user.ID, "error", err)
	}

	return err
}

func (s *PostgresStorage) Approve(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(
		ctx,
		`
		UPDATE user_access
		SET status = 'approved',
		    approved_at = NOW(),
		    rejected_at = NULL
		WHERE chat_id = $1
		`,
		chatID,
	)

	if err != nil {
		s.logger.Error("failed to approve user", "chat_id", chatID, "error", err)
	}

	return err
}

func (s *PostgresStorage) Reject(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(
		ctx,
		`
		UPDATE user_access
		SET status = 'rejected',
		    rejected_at = NOW()
		WHERE chat_id = $1
		`,
		chatID,
	)

	if err != nil {
		s.logger.Error("failed to reject user", "chat_id", chatID, "error", err)
	}

	return err
}