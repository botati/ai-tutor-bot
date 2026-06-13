package ai

import (
	"context"

	"github.com/cobrich/ai-tutor-bot/internal/entity"
)

type Client interface {
	AskText(ctx context.Context, question string, history []entity.Message) (string, error)
	AskWithImage(ctx context.Context, question string, image []byte, mimeType string) (string, error)
	TranscribeAudio(ctx context.Context, audio []byte, filename string) (string, error)
}

type LLM interface {
	AskText(
		ctx context.Context,
		question string,
		history []entity.Message,
	) (string, error)

	AskWithImage(
		ctx context.Context,
		question string,
		image []byte,
		mimeType string,
	) (string, error)

	DetectSubject(
		ctx context.Context,
		question string,
		image []byte,
		mimeType string,
	) (entity.Subject, error)
}

type Transcriber interface {
	TranscribeAudio(
		ctx context.Context,
		audio []byte,
		filename string,
	) (string, error)
}
