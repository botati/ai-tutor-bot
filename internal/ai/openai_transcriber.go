package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"
)

type OpenAITranscriber struct {
	apiKey          string
	transcribeModel string
	httpClient      *http.Client
}

func NewOpenAITranscriber(apiKey string, transcribeModel string) *OpenAIClient {
	return &OpenAIClient{
		apiKey:          apiKey,
		transcribeModel: transcribeModel,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *OpenAITranscriber) TranscribeAudio(ctx context.Context, audio []byte, filename string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("model", c.transcribeModel); err != nil {
		return "", fmt.Errorf("write model field: %w", err)
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}

	if _, err := part.Write(audio); err != nil {
		return "", fmt.Errorf("write audio file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/audio/transcriptions",
		&body,
	)
	if err != nil {
		return "", fmt.Errorf("create transcription request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send transcription request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Text  string `json:"text"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode transcription response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai transcription error: status=%d message=%s", resp.StatusCode, result.Error.Message)
	}

	if result.Text == "" {
		return "", fmt.Errorf("empty transcription result")
	}

	return result.Text, nil
}
