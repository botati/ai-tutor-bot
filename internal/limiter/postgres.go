package limiter

import (
	"context"
	"log/slog"
	"time"

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

	today := time.Now().Format("2006-01-02")

	tx, err := l.pool.Begin(ctx)
	if err != nil {
		l.logger.Error("failed to begin limiter tx", "chat_id", chatID, "error", err)
		return false
	}
	defer tx.Rollback(ctx)

	var count int
	var lastReset time.Time

	err = tx.QueryRow(
		ctx,
		`
		SELECT request_count, last_reset
		FROM user_limits
		WHERE chat_id = $1
		FOR UPDATE
		`,
		chatID,
	).Scan(&count, &lastReset)

	if err != nil {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO user_limits (
				chat_id,
				request_count,
				last_reset
			)
			VALUES ($1, 1, CURRENT_DATE)
			`,
			chatID,
		)
		if err != nil {
			l.logger.Error("failed to insert user limit", "chat_id", chatID, "error", err)
			return false
		}

		if err := tx.Commit(ctx); err != nil {
			l.logger.Error("failed to commit limiter tx", "chat_id", chatID, "error", err)
			return false
		}

		return true
	}

	if lastReset.Format("2006-01-02") != today {
		count = 0
		lastReset = time.Now()
	}

	if count >= l.limit {
		if err := tx.Commit(ctx); err != nil {
			l.logger.Error("failed to commit limiter blocked", "chat_id", chatID, "error", err)
		}
		return false
	}

	count++

	_, err = tx.Exec(
		ctx,
		`
		UPDATE user_limits
		SET request_count = $1,
		    last_reset = CURRENT_DATE
		WHERE chat_id = $2
		`,
		count,
		chatID,
	)
	if err != nil {
		l.logger.Error("failed to update user limit", "chat_id", chatID, "error", err)
		return false
	}

	if err := tx.Commit(ctx); err != nil {
		l.logger.Error("failed to commit limiter update", "chat_id", chatID, "error", err)
		return false
	}

	return true
}

func (l *PostgresRateLimiter) Remaining(chatID int64) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	today := time.Now().Format("2006-01-02")

	var count int
	var lastReset time.Time

	err := l.pool.QueryRow(
		ctx,
		`
		SELECT request_count, last_reset
		FROM user_limits
		WHERE chat_id = $1
		`,
		chatID,
	).Scan(&count, &lastReset)

	if err != nil {
		return l.limit
	}

	if lastReset.Format("2006-01-02") != today {
		return l.limit
	}

	remaining := l.limit - count
	if remaining < 0 {
		return 0
	}

	return remaining
}
