package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/cobrich/ai-tutor-bot/internal/entity"
)

type OpenAIClient struct {
	apiKey          string
	model           string
	httpClient      *http.Client
	transcribeModel string
}

func NewOpenAIClient(apiKey string, model string, transcribeModel string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		transcribeModel: transcribeModel,
	}
}

func (c *OpenAIClient) AskText(ctx context.Context, question string, history []entity.Message) (string, error) {
	input := make([]inputMessage, 0, len(history)+1)

	for _, msg := range history {
		contentType := "input_text"

		if msg.Role == "assistant" {
			contentType = "output_text"
		}

		input = append(input, inputMessage{
			Role: msg.Role,
			Content: []inputContent{
				{
					Type: contentType,
					Text: msg.Content,
				},
			},
		})
	}

	input = append(input, inputMessage{
		Role: "user",
		Content: []inputContent{
			{
				Type: "input_text",
				Text: question,
			},
		},
	})

	reqBody := responsesRequest{
		Model:        c.model,
		Instructions: TutorSystemPrompt,
		Input:        input,
	}

	return c.sendRequest(ctx, reqBody)
}

func (c *OpenAIClient) AskWithImage(ctx context.Context, question string, image []byte, mimeType string) (string, error) {
	if question == "" {
		question = "Распознай задачу на картинке и объясни решение простыми словами."
	}

	imageBase64 := base64.StdEncoding.EncodeToString(image)
	imageURL := fmt.Sprintf("data:%s;base64,%s", mimeType, imageBase64)

	reqBody := responsesRequest{
		Model:        c.model,
		Instructions: TutorSystemPrompt,
		Input: []inputMessage{
			{
				Role: "user",
				Content: []inputContent{
					{
						Type: "input_text",
						Text: question,
					},
					{
						Type:     "input_image",
						ImageURL: imageURL,
					},
				},
			},
		},
	}

	return c.sendRequest(ctx, reqBody)
}

func (c *OpenAIClient) sendRequest(ctx context.Context, reqBody responsesRequest) (string, error) {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/responses",
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	var result responsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai error: status=%d message=%s", resp.StatusCode, result.Error.Message)
	}

	if result.OutputText != "" {
		return result.OutputText, nil
	}

	for _, out := range result.Output {
		for _, content := range out.Content {
			if content.Text != "" {
				return content.Text, nil
			}
		}
	}

	return "", fmt.Errorf("empty openai response")
}

type responsesRequest struct {
	Model        string         `json:"model"`
	Instructions string         `json:"instructions"`
	Input        []inputMessage `json:"input"`
}

type inputMessage struct {
	Role    string         `json:"role"`
	Content []inputContent `json:"content"`
}

type inputContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

type responsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *OpenAIClient) TranscribeAudio(ctx context.Context, audio []byte, filename string) (string, error) {
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

func (c *OpenAIClient) DetectSubject(
	ctx context.Context,
	question string,
	image []byte,
	mimeType string,
) (entity.Subject, error) {
	var s entity.Subject
	return s, nil
}
