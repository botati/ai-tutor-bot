package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cobrich/ai-tutor-bot/internal/entity"
	"github.com/cobrich/ai-tutor-bot/internal/metrics"
)

type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *slog.Logger
}

func NewGeminiClient(apiKey string, model string, logger *slog.Logger) *GeminiClient {
	transport := &http.Transport{
		ForceAttemptHTTP2: true,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout:   90 * time.Second,
			Transport: transport,
		},
		logger: logger,
	}
}

func (c *GeminiClient) DetectSubject(
	ctx context.Context,
	question string,
	image []byte,
	mimeType string,
) (entity.Subject, error) {
	if question == "" {
		question = "Определи предмет по изображению."
	}

	parts := []geminiPart{
		{Text: question},
	}

	if len(image) > 0 {
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{
				MimeType: mimeType,
				Data:     base64.StdEncoding.EncodeToString(image),
			},
		})
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: SubjectClassifierPrompt},
			},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: parts,
			},
		},
	}

	metrics.AIRequestsTotal.WithLabelValues("gemini", "classifier").Inc()

	result, err := c.sendGenerateContent(ctx, reqBody)
	if err != nil {
		metrics.AIErrorsTotal.WithLabelValues("gemini", "classifier").Inc()
		return entity.SubjectUnknown, err
	}

	subject := entity.SubjectUnknown

	switch strings.TrimSpace(strings.ToLower(result)) {

	case "math":
		subject = entity.SubjectMath

	case "english":
		subject = entity.SubjectEnglish

	case "history":
		subject = entity.SubjectHistory

	case "chemistry":
		subject = entity.SubjectChemistry

	case "physics":
		subject = entity.SubjectPhysics

	case "biology":
		subject = entity.SubjectBiology

	case "geography":
		subject = entity.SubjectGeography
	}
	c.logger.Info(
		"subject detected",
		"subject", subject,
		"classifier_result", result,
	)
	return subject, nil
}

func (c *GeminiClient) AskText(ctx context.Context, question string, history []entity.Message) (string, error) {
	contents := c.buildContents(question, history)

	subject, err := c.DetectSubject(ctx, question, nil, "")
	if err != nil {
		subject = DetectSubject(question)
	}

	prompt := GetSubjectPrompt(subject)

	metrics.AIRequestsTotal.WithLabelValues("gemini", "text").Inc()

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: prompt},
			},
		},
		Contents: contents,
	}

	answer, err := c.sendGenerateContent(ctx, reqBody)
	if err != nil {
		metrics.AIErrorsTotal.WithLabelValues("gemini", "text").Inc()
		return "", err
	}

	return answer, nil
}

func (c *GeminiClient) AskWithImage(ctx context.Context, question string, image []byte, mimeType string) (string, error) {
	subject, err := c.DetectSubject(ctx, question, image, mimeType)
	if err != nil {
		subject = DetectSubject(question)
	}

	c.logger.Info(
		"using subject prompt",
		"subject", subject,
	)

	prompt := GetSubjectPrompt(subject)

	if question == "" {
		question = "Распознай задачу на картинке и объясни решение простыми словами."
	}

	metrics.AIRequestsTotal.WithLabelValues("gemini", "image").Inc()

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{
				{Text: prompt},
			},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: question},
					{
						InlineData: &geminiInlineData{
							MimeType: mimeType,
							Data:     base64.StdEncoding.EncodeToString(image),
						},
					},
				},
			},
		},
	}

	answer, err := c.sendGenerateContent(ctx, reqBody)
	if err != nil {
		metrics.AIErrorsTotal.WithLabelValues("gemini", "image").Inc()
		return "", err
	}

	return answer, nil
}

func (c *GeminiClient) buildContents(question string, history []entity.Message) []geminiContent {
	contents := make([]geminiContent, 0, len(history)+1)

	for _, msg := range history {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: msg.Content},
			},
		})
	}

	contents = append(contents, geminiContent{
		Role: "user",
		Parts: []geminiPart{
			{Text: question},
		},
	})

	return contents
}

func (c *GeminiClient) sendGenerateContent(ctx context.Context, reqBody geminiRequest) (string, error) {
	start := time.Now()

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model,
		c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create gemini request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send gemini request: %w", err)
	}
	defer resp.Body.Close()

	var result geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode gemini response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.Error(
			"gemini request failed",
			"status", resp.StatusCode,
			"message", result.Error.Message,
			"duration_ms", time.Since(start).Milliseconds(),
		)

		return "", fmt.Errorf("gemini error: status=%d message=%s", resp.StatusCode, result.Error.Message)
	}

	for _, candidate := range result.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				c.logger.Info(
					"gemini request completed",
					"model", c.model,
					"duration_ms", time.Since(start).Milliseconds(),
				)

				return part.Text, nil
			}
		}
	}

	return "", fmt.Errorf("empty gemini response")
}

type geminiRequest struct {
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	Contents          []geminiContent `json:"contents"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`

	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}
