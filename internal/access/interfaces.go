package access

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type Storage interface {
	GetStatus(chatID int64) (Status, bool)
	RequestAccess(user *tgbotapi.User) error
	Approve(chatID int64) error
	Reject(chatID int64) error
}