package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cobrich/ai-tutor-bot/internal/service"
	"github.com/cobrich/ai-tutor-bot/internal/stats"
	"github.com/cobrich/ai-tutor-bot/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	api             *tgbotapi.BotAPI
	tutorService    *service.TutorService
	logger          *slog.Logger
	adminTelegramID int64
	statsStorage    stats.StatsStorage
}

func NewHandler(
	api *tgbotapi.BotAPI,
	tutorService *service.TutorService,
	logger *slog.Logger,
	adminTelegramID int64,
	statsStorage stats.StatsStorage,
) *Handler {
	return &Handler{
		api:             api,
		tutorService:    tutorService,
		logger:          logger,
		adminTelegramID: adminTelegramID,
		statsStorage:    statsStorage,
	}
}
func (h *Handler) Handle(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	message := update.Message
	chatID := message.Chat.ID

	if message.IsCommand() {
		h.handleCommand(message)
		return
	}

	if len(message.Photo) > 0 {
		h.handlePhoto(message)
		return
	}

	if message.Voice != nil {
		h.handleVoice(message)
		return
	}

	if message.Text != "" {
		h.handleText(chatID, message.Text)
		return
	}

	h.sendMessage(chatID, "Пока я умею отвечать на текст и фото 🙂 Скоро добавим голос.")
}

func (h *Handler) handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	switch message.Command() {
	case "start":
		h.sendMessage(chatID, utils.START_TEXT)
	case "help":
		h.sendMessage(chatID, utils.HELP_TEXT)
	case "reset":
		h.tutorService.ResetHistory(chatID)
		h.sendMessage(chatID, "История диалога очищена ✅ Можем начать заново.")
	case "limit":
		remaining := h.tutorService.RemainingLimit(chatID)

		text := fmt.Sprintf(
			"На сегодня осталось запросов: <b>%d</b> ✅",
			remaining,
		)

		h.sendMessage(chatID, text)
	case "stats":
		if chatID != h.adminTelegramID {
			h.sendMessage(chatID, "Эта команда доступна только администратору.")
			return
		}

		snapshot := h.statsStorage.Snapshot()

		text := fmt.Sprintf(
			"<b>Статистика бота</b>\n\n"+
				"Пользователи: <b>%d</b>\n"+
				"Всего запросов: <b>%d</b>\n"+
				"Текстовые: <b>%d</b>\n"+
				"Фото: <b>%d</b>\n"+
				"Голосовые: <b>%d</b>",
			snapshot.Users,
			snapshot.TotalRequests,
			snapshot.TextRequests,
			snapshot.ImageRequests,
			snapshot.VoiceRequests,
		)

		h.sendMessage(chatID, text)
	default:
		h.sendMessage(chatID, "Я пока не знаю такую команду 😅 Напиши /help")
	}
}

func (h *Handler) handleText(chatID int64, text string) {
	cleanText := strings.TrimSpace(text)

	if cleanText == "" {
		h.sendMessage(chatID, "Напиши вопрос текстом 🙂")
		return
	}

	h.statsStorage.TrackTextRequest(chatID)

	h.sendChatAction(chatID, tgbotapi.ChatTyping)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	answer, err := h.tutorService.AnswerText(ctx, chatID, cleanText)
	if err != nil {
		h.logger.Error(
			"failed to answer text",
			"chat_id", chatID,
			"error", err,
		)

		h.sendMessage(chatID, "Не получилось получить ответ от ИИ 😔 Попробуй ещё раз.")
		return
	}

	h.sendMessage(chatID, answer)
}

func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"

	if _, err := h.api.Send(msg); err != nil {
		h.logger.Error(
			"failed to send html message",
			"chat_id", chatID,
			"error", err,
		)

		plainMsg := tgbotapi.NewMessage(chatID, text)
		if _, err := h.api.Send(plainMsg); err != nil {
			h.logger.Error(
				"failed to send html message",
				"chat_id", chatID,
				"error", err,
			)
		}
	}
}

func (h *Handler) handlePhoto(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	h.sendChatAction(chatID, tgbotapi.ChatUploadPhoto)

	photos := message.Photo
	bestPhoto := photos[len(photos)-1]

	image, mimeType, err := h.downloadTelegramFile(bestPhoto.FileID)
	if err != nil {
		h.logger.Error(
			"failed to download photo",
			"chat_id", chatID,
			"error", err,
		)

		h.sendMessage(chatID, "Не получилось скачать фото 😔 Попробуй отправить ещё раз.")
		return
	}

	h.statsStorage.TrackImageRequest(chatID)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	answer, err := h.tutorService.AnswerImage(ctx, chatID, message.Caption, image, mimeType)
	if err != nil {
		h.logger.Error(
			"failed to answer image",
			"chat_id", chatID,
			"error", err,
		)
		h.sendMessage(chatID, "Не получилось разобрать фото 😔 Попробуй отправить более чёткое изображение.")
		return
	}

	h.sendMessage(chatID, answer)
}

func (h *Handler) handleVoice(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	h.sendChatAction(chatID, tgbotapi.ChatTyping)

	audio, _, err := h.downloadTelegramFile(message.Voice.FileID)
	if err != nil {
		h.logger.Error(
			"failed to download voice",
			"chat_id", chatID,
			"error", err,
		)
		h.sendMessage(chatID, "Не получилось скачать голосовое 😔 Попробуй ещё раз.")
		return
	}

	h.statsStorage.TrackVoiceRequest(chatID)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	transcript, answer, err := h.tutorService.AnswerVoice(ctx, chatID, audio, "voice.ogg")
	if err != nil {
		h.logger.Error(
			"failed to answer voice",
			"chat_id", chatID,
			"error", err,
		)
		h.sendMessage(chatID, "Не получилось разобрать голосовое 😔 Попробуй сказать чуть чётче.")
		return
	}

	response := "<b>Я услышал:</b>\n" +
		"<i>" + escapeHTML(transcript) + "</i>\n\n" +
		answer

	h.sendMessage(chatID, response)
}

func (h *Handler) sendChatAction(chatID int64, action string) {
	chatAction := tgbotapi.NewChatAction(chatID, action)

	if _, err := h.api.Send(chatAction); err != nil {
		h.logger.Error(
			"failed to send chat action",
			"chat_id", chatID,
			"error", err,
		)
	}
}
