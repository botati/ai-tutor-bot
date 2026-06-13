package stats

import "sync"

type MemoryStatsStorage struct {
	mu sync.RWMutex

	users map[int64]struct{}

	textRequests  int
	imageRequests int
	voiceRequests int
}

func NewMemoryStatsStorage() *MemoryStatsStorage {
	return &MemoryStatsStorage{
		users: make(map[int64]struct{}),
	}
}

func (s *MemoryStatsStorage) TrackUser(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[chatID] = struct{}{}
}

func (s *MemoryStatsStorage) TrackTextRequest(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[chatID] = struct{}{}
	s.textRequests++
}

func (s *MemoryStatsStorage) TrackImageRequest(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[chatID] = struct{}{}
	s.imageRequests++
}

func (s *MemoryStatsStorage) TrackVoiceRequest(chatID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[chatID] = struct{}{}
	s.voiceRequests++
}

func (s *MemoryStatsStorage) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.textRequests + s.imageRequests + s.voiceRequests

	return Snapshot{
		Users:         len(s.users),
		TextRequests:  s.textRequests,
		ImageRequests: s.imageRequests,
		VoiceRequests: s.voiceRequests,
		TotalRequests: total,
	}
}