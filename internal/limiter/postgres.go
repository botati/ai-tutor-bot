package limiter

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRateLimiter struct {
	pool   *pgxpool.Pool
	limit  int
	logger *slog.Logger
}

func NewPostgresRateLimiter(pool *pgxpool.Pool, limit int, logger *slog.Logger) *PostgresRateLimiter {
	return &PostgresRateLimiter{
		pool:   pool,
		limit:  limit,
		logger: logger,
	}
}

func (l *PostgresRateLimiter) Allow(chatID int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	err := l.pool.QueryRow(
		ctx,
		`
		INSERT INTO user_limits (
			chat_id,
			request_count,
			last_reset
		)
		VALUES ($1, 1, CURRENT_DATE)
		ON CONFLICT (chat_id)
		DO UPDATE SET
			request_count = CASE
				WHEN user_limits.last_reset < CURRENT_DATE THEN 1
				ELSE user_limits.request_count + 1
			END,
			last_reset = CURRENT_DATE
		WHERE
			user_limits.last_reset < CURRENT_DATE
			OR user_limits.request_count < $2
		RETURNING request_count
		`,
		chatID,
		l.limit,
	).Scan(&count)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false
		}

		l.logger.Error(
			"failed to check rate limit",
			"chat_id", chatID,
			"error", err,
		)

		return false
	}

	return true
}

func (l *PostgresRateLimiter) Remaining(chatID int64) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var remaining int

	err := l.pool.QueryRow(
		ctx,
		`
		SELECT
			CASE
				WHEN last_reset < CURRENT_DATE THEN $2
				ELSE GREATEST($2 - request_count, 0)
			END
		FROM user_limits
		WHERE chat_id = $1
		`,
		chatID,
		l.limit,
	).Scan(&remaining)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return l.limit
		}

		l.logger.Error(
			"failed to get remaining limit",
			"chat_id", chatID,
			"error", err,
		)

		return 0
	}

	return remaining
}
