package limiter

import (
	"sync"
	"time"
)

type userLimit struct {
	Count int
	Day   int
}

type MemoryRateLimiter struct {
	mu       sync.RWMutex
	limit    int
	requests map[int64]userLimit
}

func NewMemoryRateLimiter(limit int) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		limit:    limit,
		requests: make(map[int64]userLimit),
	}
}

func (l *MemoryRateLimiter) Allow(chatID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	today := time.Now().YearDay()
	current := l.requests[chatID]

	if current.Day != today {
		current = userLimit{
			Count: 0,
			Day:   today,
		}
	}

	if current.Count >= l.limit {
		l.requests[chatID] = current
		return false
	}

	current.Count++
	l.requests[chatID] = current

	return true
}

func (l *MemoryRateLimiter) Remaining(chatID int64) int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	today := time.Now().YearDay()
	current := l.requests[chatID]

	if current.Day != today {
		return l.limit
	}

	remaining := l.limit - current.Count
	if remaining < 0 {
		return 0
	}

	return remaining
}