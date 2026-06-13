package limiter

type RateLimiter interface {
	Allow(chatID int64) bool
	Remaining(chatID int64) int
}