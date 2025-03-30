package bot

import (
	"strings"
)

func ExtractKeyPoints(text string) string {
	// Пример логики извлечения ключевых моментов.
	words := strings.Fields(text)
	if len(words) > 10 {
		return strings.Join(words[:10], " ") + "..."
	}
	return text
}
