package storage

import "github.com/cobrich/ai-tutor-bot/internal/entity"

type HistoryStorage interface {
	Add(chatID int64, message entity.Message)
	GetRecent(chatID int64, limit int) []entity.Message
	Clear(chatID int64)
}
