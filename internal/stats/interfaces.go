package stats

type StatsStorage interface {
	TrackUser(chatID int64)
	TrackTextRequest(chatID int64)
	TrackImageRequest(chatID int64)
	TrackVoiceRequest(chatID int64)
	Snapshot() Snapshot
	UserProfile(chatID int64) UserProfile
}

type Snapshot struct {
	Users         int
	TextRequests  int
	ImageRequests int
	VoiceRequests int
	TotalRequests int
}

type UserProfile struct {
	ChatID        int64
	CreatedAt     string
	LastSeenAt    string
	TotalRequests int
	TextRequests  int
	ImageRequests int
	VoiceRequests int
}
