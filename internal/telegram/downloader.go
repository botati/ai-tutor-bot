package telegram

import (
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) downloadTelegramFile(fileID string, maxSizeBytes int) ([]byte, string, error) {
	file, err := h.api.GetFile(tgbotapi.FileConfig{
		FileID: fileID,
	})
	if err != nil {
		return nil, "", fmt.Errorf("get telegram file: %w", err)
	}

	if file.FileSize > maxSizeBytes {
		return nil, "", fmt.Errorf("file too large: size=%d limit=%d", file.FileSize, maxSizeBytes)
	}

	fileURL := file.Link(h.api.Token)

	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, "", fmt.Errorf("download telegram file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("telegram file download status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxSizeBytes)+1))
	if err != nil {
		return nil, "", fmt.Errorf("read telegram file: %w", err)
	}

	if len(data) > maxSizeBytes {
		return nil, "", fmt.Errorf("downloaded file too large: size=%d limit=%d", len(data), maxSizeBytes)
	}

	mimeType := http.DetectContentType(data)

	return data, mimeType, nil
}
