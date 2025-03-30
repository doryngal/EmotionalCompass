package ai

func SummarizeMessage(text string) string {
	//TODO Логика извлечения ключевых моментов
	return text[:min(len(text), 100)] // Пример: возвращаем первые 100 символов
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
