package telegram

import "strings"

const telegramMessageLimit = 4000

func splitMessage(text string) []string {
	text = strings.TrimSpace(text)

	if len(text) <= telegramMessageLimit {
		return []string{text}
	}

	var parts []string

	for len(text) > telegramMessageLimit {
		splitAt := strings.LastIndex(text[:telegramMessageLimit], "\n")

		if splitAt == -1 || splitAt < 1000 {
			splitAt = telegramMessageLimit
		}

		parts = append(parts, strings.TrimSpace(text[:splitAt]))
		text = strings.TrimSpace(text[splitAt:])
	}

	if text != "" {
		parts = append(parts, text)
	}

	return parts
}