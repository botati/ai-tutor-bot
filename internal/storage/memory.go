package storage

import (
	"sync"

	"github.com/cobrich/ai-tutor-bot/internal/entity"
)

type MemoryHistoryStorage struct {
	mu       sync.RWMutex
	messages map[int64][]entity.Message
}

func NewMemoryHistoryStorage() *MemoryHistoryStorage {
	return &MemoryHistoryStorage{
		messages: make(map[int64][]entity.Message),
	}
}

func (s *MemoryHistoryStorage) Add(chatID int64, message entity.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages[chatID] = append(s.messages[chatID], message)

	if len(s.messages[chatID]) > 20 {
		s.messages[chatID] = s.messages[chatID][len(s.messages[chatID])-20:]
	}
}

func (s *MemoryHistoryStorage) GetRecent(chatID int64, limit int) []entity.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := s.messages[chatID]

	if len(history) <= limit {
		return append([]entity.Message(nil), history...)
	}

	return append([]entity.Message(nil), history[len(history)-limit:]...)
}

func (s *MemoryHistoryStorage) Clear(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.messages, chatID)
}
