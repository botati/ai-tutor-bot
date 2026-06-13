package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/cobrich/ai-tutor-bot/internal/ai"
	"github.com/cobrich/ai-tutor-bot/internal/entity"
	"github.com/cobrich/ai-tutor-bot/internal/limiter"
	"github.com/cobrich/ai-tutor-bot/internal/storage"
)

type TutorService struct {
	aiClient    ai.Client
	history     storage.HistoryStorage
	rateLimiter limiter.RateLimiter
}

func NewTutorService(
	aiClient ai.Client,
	history storage.HistoryStorage,
	rateLimiter limiter.RateLimiter,
) *TutorService {
	return &TutorService{
		aiClient:    aiClient,
		history:     history,
		rateLimiter: rateLimiter,
	}
}

func (s *TutorService) AnswerText(ctx context.Context, chatID int64, question string) (string, error) {
	if !s.rateLimiter.Allow(chatID) {
		return "Ты достиг дневного лимита запросов 😔 Попробуй снова завтра.", nil
	}

	question = strings.TrimSpace(question)
	if question == "" {
		return "Напиши вопрос текстом 🙂", nil
	}

	recentHistory := s.history.GetRecent(chatID, 10)

	answer, err := s.aiClient.AskText(ctx, question, recentHistory)
	if err != nil {
		return "", fmt.Errorf("failed to answer text: %w", err)
	}

	s.history.Add(chatID, entity.Message{
		Role:    "user",
		Content: question,
	})

	s.history.Add(chatID, entity.Message{
		Role:    "assistant",
		Content: answer,
	})

	return answer, nil
}

func (s *TutorService) AnswerImage(ctx context.Context, chatID int64, caption string, image []byte, mimeType string) (string, error) {
	if !s.rateLimiter.Allow(chatID) {
		return "Ты достиг дневного лимита запросов 😔 Попробуй снова завтра.", nil
	}

	caption = strings.TrimSpace(caption)

	answer, err := s.aiClient.AskWithImage(ctx, caption, image, mimeType)
	if err != nil {
		return "", fmt.Errorf("failed to answer image: %w", err)
	}

	userText := caption
	if userText == "" {
		userText = "[Пользователь отправил фото задачи]"
	}

	s.history.Add(chatID, entity.Message{
		Role:    "user",
		Content: userText,
	})

	s.history.Add(chatID, entity.Message{
		Role:    "assistant",
		Content: answer,
	})

	return answer, nil
}

func (s *TutorService) AnswerVoice(ctx context.Context, chatID int64, audio []byte, filename string) (string, string, error) {
	transcript, err := s.aiClient.TranscribeAudio(ctx, audio, filename)
	if err != nil {
		return "", "", fmt.Errorf("failed to transcribe voice: %w", err)
	}

	// s.stats.TrackVoiceRequest(chatID)

	answer, err := s.AnswerText(ctx, chatID, transcript)
	if err != nil {
		return transcript, "", fmt.Errorf("failed to answer transcribed text: %w", err)
	}

	return transcript, answer, nil
}

func (s *TutorService) ResetHistory(chatID int64) {
	s.history.Clear(chatID)
}

func (s *TutorService) RemainingLimit(chatID int64) int {
	return s.rateLimiter.Remaining(chatID)
}
