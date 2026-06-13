package ai

import (
	"fmt"
	"log/slog"

	"github.com/cobrich/ai-tutor-bot/internal/config"
)

func NewLLM(cfg config.Config, logger *slog.Logger) (LLM, error) {
	switch cfg.AIProvider {
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required")
		}

		return NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel, logger), nil

	case "openai":
		if cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required")
		}

		return NewOpenAIClient(
			cfg.OpenAIAPIKey,
			cfg.OpenAIModel,
			cfg.OpenAITranscribeModel,
		), nil

	default:
		return nil, fmt.Errorf("unknown AI_PROVIDER: %s", cfg.AIProvider)
	}
}

func NewTranscriber(cfg config.Config) (Transcriber, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required for voice transcription")
	}

	return NewOpenAITranscriber(
		cfg.OpenAIAPIKey,
		cfg.OpenAITranscribeModel,
	), nil
}
