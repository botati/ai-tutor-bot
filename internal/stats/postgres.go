package stats

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStatsStorage struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresStatsStorage(pool *pgxpool.Pool, logger *slog.Logger) *PostgresStatsStorage {
	return &PostgresStatsStorage{
		pool:   pool,
		logger: logger,
	}
}

func (s *PostgresStatsStorage) TrackUser(chatID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO bot_users (chat_id)
		VALUES ($1)
		ON CONFLICT (chat_id)
		DO UPDATE SET last_seen_at = NOW()
		`,
		chatID,
	)

	if err != nil {
		s.logger.Error("failed to track user", "chat_id", chatID, "error", err)
	}
}

func (s *PostgresStatsStorage) TrackTextRequest(chatID int64) {
	s.TrackUser(chatID)
	s.increment("text_requests")
	s.increment("total_requests")
}

func (s *PostgresStatsStorage) TrackImageRequest(chatID int64) {
	s.TrackUser(chatID)
	s.increment("image_requests")
	s.increment("total_requests")
}

func (s *PostgresStatsStorage) TrackVoiceRequest(chatID int64) {
	s.TrackUser(chatID)
	s.increment("voice_requests")
	s.increment("total_requests")
}

func (s *PostgresStatsStorage) Snapshot() Snapshot {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var users int

	err := s.pool.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM bot_users
		`,
	).Scan(&users)

	if err != nil {
		s.logger.Error("failed to count users", "error", err)
	}

	values := map[string]int{
		"text_requests":  0,
		"image_requests": 0,
		"voice_requests": 0,
		"total_requests": 0,
	}

	rows, err := s.pool.Query(
		ctx,
		`
		SELECT key, value
		FROM bot_stats
		WHERE key IN (
			'text_requests',
			'image_requests',
			'voice_requests',
			'total_requests'
		)
		`,
	)

	if err != nil {
		s.logger.Error("failed to load bot stats", "error", err)
		return Snapshot{
			Users: users,
		}
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value int

		if err := rows.Scan(&key, &value); err != nil {
			continue
		}

		values[key] = value
	}

	return Snapshot{
		Users:         users,
		TextRequests:  values["text_requests"],
		ImageRequests: values["image_requests"],
		VoiceRequests: values["voice_requests"],
		TotalRequests: values["total_requests"],
	}
}

func (s *PostgresStatsStorage) increment(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO bot_stats (key, value)
		VALUES ($1, 1)
		ON CONFLICT (key)
		DO UPDATE SET value = bot_stats.value + 1
		`,
		key,
	)

	if err != nil {
		s.logger.Error("failed to increment bot stat", "error", err)
	}
}
