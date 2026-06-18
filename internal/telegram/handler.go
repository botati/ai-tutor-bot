package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/cobrich/ai-tutor-bot/internal/access"
	"github.com/cobrich/ai-tutor-bot/internal/metrics"
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
	accessStorage   access.Storage
}

func NewHandler(
	api *tgbotapi.BotAPI,
	tutorService *service.TutorService,
	logger *slog.Logger,
	adminTelegramID int64,
	statsStorage stats.StatsStorage,
	accessStorage access.Storage,
) *Handler {
	return &Handler{
		api:             api,
		tutorService:    tutorService,
		logger:          logger,
		adminTelegramID: adminTelegramID,
		statsStorage:    statsStorage,
		accessStorage:   accessStorage,
	}
}
func (h *Handler) Handle(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
		return
	}

	if update.Message == nil {
		return
	}

	message := update.Message
	chatID := message.Chat.ID

	if message.IsCommand() {
		h.handleCommand(message)
		return
	}

	if !h.isApproved(chatID) {
		h.requestAccess(message)
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
	metrics.TelegramRequestsTotal.WithLabelValues("command").Inc()

	chatID := message.Chat.ID

	switch message.Command() {
	case "start":
		h.sendMessage(chatID, "Сәлем! AI Ustaz дайын ✅\n\nСұрағыңызды мәтін, фото немесе дауыспен жіберіңіз.")
		// if h.isApproved(chatID) {
		// 	h.sendMessage(chatID, "Сәлем! AI Ustaz дайын ✅\n\nСұрағыңызды мәтін, фото немесе дауыспен жіберіңіз.")
		// 	return
		// }

		// h.requestAccess(message)
	case "help":
		h.sendMessage(chatID, utils.HELP_TEXT)
	case "reset":
		h.tutorService.ResetHistory(chatID)
		h.sendMessage(chatID, utils.RESET_TEXT)
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
				"Голосовые: <b>%d</b>\n"+
				"👍 Полезно: %d\n"+
				"👎 Не помогло: %d\n",
			snapshot.Users,
			snapshot.TotalRequests,
			snapshot.TextRequests,
			snapshot.ImageRequests,
			snapshot.VoiceRequests,
			snapshot.PositiveFeedback,
			snapshot.NegativeFeedback,
		)

		h.sendMessage(chatID, text)
	case "profile":
		profile := h.statsStorage.UserProfile(chatID)
		remaining := h.tutorService.RemainingLimit(chatID)

		text := fmt.Sprintf(
			"👤 Профиль\n\n"+
				"Осталось запросов сегодня: %d\n\n"+
				"📊 Мои запросы:\n"+
				"Всего: %d\n"+
				"Текст: %d\n"+
				"Фото: %d\n"+
				"Голос: %d\n\n"+
				"Дата регистрации: %s\n"+
				"Последняя активность: %s",
			remaining,
			profile.TotalRequests,
			profile.TextRequests,
			profile.ImageRequests,
			profile.VoiceRequests,
			profile.CreatedAt,
			profile.LastSeenAt,
		)

		h.sendMessage(chatID, text)
	default:
		h.sendMessage(chatID, "Я пока не знаю такую команду 😅 Напиши /help")
	}
}

func (h *Handler) handleText(chatID int64, text string) {
	metrics.TelegramRequestsTotal.WithLabelValues("text").Inc()

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

	h.sendMessageWithFeedback(chatID, answer)
}

func (h *Handler) handlePhoto(message *tgbotapi.Message) {
	metrics.TelegramRequestsTotal.WithLabelValues("photo").Inc()

	chatID := message.Chat.ID

	h.sendChatAction(chatID, tgbotapi.ChatUploadPhoto)

	photos := message.Photo
	bestPhoto := photos[len(photos)-1]

	image, mimeType, err := h.downloadTelegramFile(bestPhoto.FileID, MaxTelegramPhotoSizeBytes)
	if err != nil {
		h.logger.Error(
			"failed to download photo",
			"chat_id", chatID,
			"error", err,
		)

		h.sendMessage(chatID, "Не получилось скачать фото 😔\n\nПроверь, что фото не слишком большое и попробуй ещё раз.")
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

	h.sendMessageWithFeedback(chatID, answer)
}

func (h *Handler) handleVoice(message *tgbotapi.Message) {
	metrics.TelegramRequestsTotal.WithLabelValues("voice").Inc()

	chatID := message.Chat.ID

	h.sendChatAction(chatID, tgbotapi.ChatTyping)

	audio, _, err := h.downloadTelegramFile(message.Voice.FileID, MaxTelegramVoiceSizeBytes)
	if err != nil {
		h.logger.Error(
			"failed to download voice",
			"chat_id", chatID,
			"error", err,
		)
		h.sendMessage(chatID, "Не получилось скачать голосовое 😔\n\nПроверь, что оно не слишком длинное и попробуй ещё раз.")
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

	response := "🎧 Я услышал:\n" +
		transcript + "\n\n" +
		answer

	h.sendMessage(chatID, response)
}

func (h *Handler) sendChatAction(chatID int64, action string) {
	_, err := h.api.Request(
		tgbotapi.NewChatAction(chatID, action),
	)

	if err != nil {
		h.logger.Error(
			"failed to send chat action",
			"chat_id", chatID,
			"error", err,
		)
	}
}

func (h *Handler) sendMessage(chatID int64, text string) {
	text = sanitizeAnswer(text)

	parts := splitMessage(text)

	for _, part := range parts {
		msg := tgbotapi.NewMessage(chatID, part)

		if _, err := h.api.Send(msg); err != nil {
			h.logger.Error(
				"failed to send message",
				"chat_id", chatID,
				"error", err,
			)
		}
	}
}

func (h *Handler) sendMessageWithFeedback(chatID int64, text string) {
	text = sanitizeAnswer(text)

	msg := tgbotapi.NewMessage(chatID, text)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👍 Полезно", "feedback:positive"),
			tgbotapi.NewInlineKeyboardButtonData("👎 Не помогло", "feedback:negative"),
		),
	)

	msg.ReplyMarkup = keyboard

	if _, err := h.api.Send(msg); err != nil {
		h.logger.Error("failed to send message with feedback", "chat_id", chatID, "error", err)
	}
}

func (h *Handler) handleCallback(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	data := callback.Data

	if strings.HasPrefix(data, "access:") {
		h.handleAccessCallback(callback)
		return
	}

	switch data {
	case "feedback:positive":
		h.statsStorage.TrackFeedback(chatID, "positive")
		h.answerCallback(callback.ID, "Спасибо за оценку 👍")
		h.replaceFeedbackButtons(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
		)

	case "feedback:negative":
		h.statsStorage.TrackFeedback(chatID, "negative")
		h.answerCallback(callback.ID, "Спасибо! Я учту это 👎")
		h.replaceFeedbackButtons(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
		)
	case "feedback:already":
		h.answerCallback(callback.ID, "Отзыв уже принят ✅")

	default:
		h.answerCallback(callback.ID, "Неизвестное действие")
	}
}

func (h *Handler) answerCallback(callbackID string, text string) {
	callback := tgbotapi.NewCallback(callbackID, text)

	if _, err := h.api.Request(callback); err != nil {
		h.logger.Error("failed to answer callback", "error", err)
	}
}

func (h *Handler) replaceFeedbackButtons(chatID int64, messageID int) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Спасибо за отзыв", "feedback:already"),
		),
	)

	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, keyboard)

	if _, err := h.api.Request(edit); err != nil {
		h.logger.Error("failed to replace feedback buttons", "error", err)
	}
}

func (h *Handler) isApproved(chatID int64) bool {
	if chatID == h.adminTelegramID {
		return true
	}

	status, ok := h.accessStorage.GetStatus(chatID)
	return ok && status == access.StatusApproved
}

func (h *Handler) requestAccess(message *tgbotapi.Message) {
	chatID := message.Chat.ID

	status, ok := h.accessStorage.GetStatus(chatID)
	if ok && status == access.StatusPending {
		h.sendMessage(chatID, "Сіздің өтінішіңіз админге жіберілді. Рұқсат күтіңіз 🙂")
		return
	}

	if ok && status == access.StatusRejected {
		h.sendMessage(chatID, "Кешіріңіз, сізге ботты қолдануға рұқсат берілмеді.")
		return
	}

	if err := h.accessStorage.RequestAccess(message.From); err != nil {
		h.sendMessage(chatID, "Қате шықты. Кейінірек қайталап көріңіз.")
		return
	}

	h.sendMessage(chatID, "Сіздің өтінішіңіз админге жіберілді. Рұқсат күтіңіз 🙂")
	h.notifyAdminAboutAccessRequest(message.From)
}

func (h *Handler) notifyAdminAboutAccessRequest(user *tgbotapi.User) {
	text := fmt.Sprintf(
		"🔐 Жаңа қолданушы рұқсат сұрады\n\n"+
			"ID: %d\n"+
			"Username: @%s\n"+
			"First name: %s\n"+
			"Last name: %s",
		user.ID,
		user.UserName,
		user.FirstName,
		user.LastName,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Рұқсат беру", fmt.Sprintf("access:approve:%d", user.ID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Қабылдамау", fmt.Sprintf("access:reject:%d", user.ID)),
		),
	)

	msg := tgbotapi.NewMessage(h.adminTelegramID, text)
	msg.ReplyMarkup = keyboard

	if _, err := h.api.Send(msg); err != nil {
		h.logger.Error("failed to notify admin about access request", "error", err)
	}
}

func (h *Handler) handleAccessCallback(callback *tgbotapi.CallbackQuery) {
	if callback.From.ID != h.adminTelegramID {
		h.answerCallback(callback.ID, "Бұл әрекет тек админге арналған.")
		return
	}

	parts := strings.Split(callback.Data, ":")
	if len(parts) != 3 {
		h.answerCallback(callback.ID, "Қате callback.")
		return
	}

	action := parts[1]
	userID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		h.answerCallback(callback.ID, "Қате user ID.")
		return
	}

	switch action {
	case "approve":
		if err := h.accessStorage.Approve(userID); err != nil {
			h.answerCallback(callback.ID, "Қате шықты.")
			return
		}

		h.answerCallback(callback.ID, "Қолданушыға рұқсат берілді ✅")
		h.sendMessage(userID, "Сізге AI Ustaz қолдануға рұқсат берілді ✅\n\nЕнді сұрағыңызды жібере аласыз.")

	case "reject":
		if err := h.accessStorage.Reject(userID); err != nil {
			h.answerCallback(callback.ID, "Қате шықты.")
			return
		}

		h.answerCallback(callback.ID, "Қолданушы қабылданбады ❌")
		h.sendMessage(userID, "Кешіріңіз, сізге AI Ustaz қолдануға рұқсат берілмеді.")
	}

	h.removeInlineKeyboard(callback.Message.Chat.ID, callback.Message.MessageID)
}

func (h *Handler) removeInlineKeyboard(chatID int64, messageID int) {
	edit := tgbotapi.NewEditMessageReplyMarkup(
		chatID,
		messageID,
		tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		},
	)

	if _, err := h.api.Request(edit); err != nil {
		h.logger.Error(
			"failed to remove inline keyboard",
			"chat_id", chatID,
			"message_id", messageID,
			"error", err,
		)
	}
}
