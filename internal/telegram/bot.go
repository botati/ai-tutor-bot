package telegram

import (
	"context"
	"log/slog"

	"github.com/cobrich/ai-tutor-bot/internal/service"
	"github.com/cobrich/ai-tutor-bot/internal/stats"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	handler *Handler
	logger  *slog.Logger
}

func NewBot(
	token string,
	tutorService *service.TutorService,
	logger *slog.Logger,
	adminTelegramID int64,
	statsStorage stats.StatsStorage,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	logger.Info("authorized on telegram", "username", api.Self.UserName)

	return &Bot{
		api:     api,
		handler: NewHandler(api, tutorService, logger, adminTelegramID, statsStorage),
		logger:  logger,
	}, nil
}

func (b *Bot) Start(ctx context.Context) {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := b.api.GetUpdatesChan(updateConfig)

	b.logger.Info("bot started, waiting for messages")

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("shutdown signal received, stopping telegram updates")
			b.api.StopReceivingUpdates()
			return

		case update, ok := <-updates:
			if !ok {
				b.logger.Info("telegram updates channel closed")
				return
			}

			b.handler.Handle(update)
		}
	}
}
