package telegram

import "strings"

func sanitizeAnswer(text string) string {
	replacer := strings.NewReplacer(
		"<b>", "",
		"</b>", "",
		"<i>", "",
		"</i>", "",
		"<code>", "",
		"</code>", "",
		"<pre>", "",
		"</pre>", "",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
	)

	return replacer.Replace(text)
}