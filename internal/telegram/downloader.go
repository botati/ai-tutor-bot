package telegram

import (
	"fmt"
	"io"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handler) downloadTelegramFile(fileID string) ([]byte, string, error) {
	file, err := h.api.GetFile(tgbotapi.FileConfig{
		FileID: fileID,
	})
	if err != nil {
		return nil, "", fmt.Errorf("get telegram file: %w", err)
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read telegram file: %w", err)
	}

	mimeType := http.DetectContentType(data)

	return data, mimeType, nil
}