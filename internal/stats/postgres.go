package stats

import (
	"context"
	"fmt"
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
	s.incrementUserStat(chatID, "text_requests")
	s.increment("text_requests")
	s.increment("total_requests")
}

func (s *PostgresStatsStorage) TrackImageRequest(chatID int64) {
	s.TrackUser(chatID)
	s.incrementUserStat(chatID, "image_requests")
	s.increment("image_requests")
	s.increment("total_requests")
}

func (s *PostgresStatsStorage) TrackVoiceRequest(chatID int64) {
	s.TrackUser(chatID)
	s.incrementUserStat(chatID, "voice_requests")
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

	var positiveFeedback int
	var negativeFeedback int

	err = s.pool.QueryRow(
		ctx,
		`
	SELECT COUNT(*)
	FROM feedback
	WHERE value = 'positive'
	`,
	).Scan(&positiveFeedback)

	if err != nil {
		s.logger.Error("failed to count positive feedback", "error", err)
	}

	err = s.pool.QueryRow(
		ctx,
		`
	SELECT COUNT(*)
	FROM feedback
	WHERE value = 'negative'
	`,
	).Scan(&negativeFeedback)

	if err != nil {
		s.logger.Error("failed to count negative feedback", "error", err)
	}

	return Snapshot{
		Users:            users,
		TextRequests:     values["text_requests"],
		ImageRequests:    values["image_requests"],
		VoiceRequests:    values["voice_requests"],
		TotalRequests:    values["total_requests"],
		PositiveFeedback: positiveFeedback,
		NegativeFeedback: negativeFeedback,
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

func (s *PostgresStatsStorage) UserProfile(chatID int64) UserProfile {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var profile UserProfile
	profile.ChatID = chatID

	err := s.pool.QueryRow(
		ctx,
		`
		SELECT
			u.chat_id,
			TO_CHAR(u.created_at + INTERVAL '5 hours', 'YYYY-MM-DD HH24:MI'),
			TO_CHAR(u.last_seen_at + INTERVAL '5 hours', 'YYYY-MM-DD HH24:MI'),
			COALESCE(st.text_requests, 0),
			COALESCE(st.image_requests, 0),
			COALESCE(st.voice_requests, 0)
		FROM bot_users u
		LEFT JOIN user_stats st ON st.chat_id = u.chat_id
		WHERE u.chat_id = $1
		`,
		chatID,
	).Scan(
		&profile.ChatID,
		&profile.CreatedAt,
		&profile.LastSeenAt,
		&profile.TextRequests,
		&profile.ImageRequests,
		&profile.VoiceRequests,
	)

	if err != nil {
		s.logger.Error("failed to load user profile", "chat_id", chatID, "error", err)
		return profile
	}

	profile.TotalRequests = profile.TextRequests + profile.ImageRequests + profile.VoiceRequests

	return profile
}

func (s *PostgresStatsStorage) incrementUserStat(chatID int64, column string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if column != "text_requests" &&
		column != "image_requests" &&
		column != "voice_requests" {
		return
	}

	query := fmt.Sprintf(
		`
		INSERT INTO user_stats (chat_id, %s)
		VALUES ($1, 1)
		ON CONFLICT (chat_id)
		DO UPDATE SET
			%s = user_stats.%s + 1,
			updated_at = NOW()
		`,
		column,
		column,
		column,
	)

	if _, err := s.pool.Exec(ctx, query, chatID); err != nil {
		s.logger.Error("failed to increment user stat", "chat_id", chatID, "column", column, "error", err)
	}
}

func (s *PostgresStatsStorage) TrackFeedback(chatID int64, value string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if value != "positive" && value != "negative" {
		return
	}

	_, err := s.pool.Exec(
		ctx,
		`
		INSERT INTO feedback (chat_id, value)
		VALUES ($1, $2)
		`,
		chatID,
		value,
	)

	if err != nil {
		s.logger.Error(
			"failed to track feedback",
			"chat_id", chatID,
			"value", value,
			"error", err,
		)
	}
}
