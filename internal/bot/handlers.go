package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram-bot/pkg/telegram"
)

// HandleState отправляет сообщение с кнопками на основе состояния
func HandleState(botAPI *telegram.TelegramClient, chatID int64, state string) {
	s, exists := states[state]
	if !exists {
		botAPI.SendMessage(chatID, "Неизвестное состояние. Попробуйте снова.")
		return
	}

	// Создаем кнопки
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, btn := range s.Buttons {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.NextState),
		))
	}

	msg := tgbotapi.NewMessage(chatID, s.Message)
	if len(keyboard) > 0 {
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}

	botAPI.SendMessageWithMarkup(chatID, msg)
}

// HandleCallbackQuery обрабатывает нажатие на кнопку и переключает состояние
func HandleCallbackQuery(botAPI *telegram.TelegramClient, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	nextState := callback.Data

	// Отвечаем на callback, чтобы убрать "часики"
	botAPI.AnswerCallbackQuery(callback.ID)

	HandleState(botAPI, chatID, nextState)
}
