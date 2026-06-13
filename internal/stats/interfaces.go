package stats

type StatsStorage interface {
	TrackUser(chatID int64)
	TrackTextRequest(chatID int64)
	TrackImageRequest(chatID int64)
	TrackVoiceRequest(chatID int64)
	Snapshot() Snapshot
}

type Snapshot struct {
	Users         int
	TextRequests  int
	ImageRequests int
	VoiceRequests int
	TotalRequests int
}