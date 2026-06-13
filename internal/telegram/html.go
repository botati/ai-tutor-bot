package telegram

import "html"

func escapeHTML(text string) string {
	return html.EscapeString(text)
}
