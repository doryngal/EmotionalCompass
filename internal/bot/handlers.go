// handlers.go
package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
	"telegram-bot/pkg/telegram"
	"time"
)

// HandleState отправляет сообщения с изображениями и текстом
func HandleState(botAPI *telegram.TelegramClient, chatID int64, state string) {
	s, exists := GetState(state)
	if !exists {
		botAPI.SendHTMLMessage(chatID, "Неизвестное состояние. Попробуйте снова.")
		return
	}

	// Обновляем состояние пользователя
	userStates[chatID] = state

	// Создаем обычные inline-кнопки
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	if len(s.Buttons) > 0 {
		for _, btn := range s.Buttons {
			inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.NextState),
			))
		}
	}

	var quickReplyButtons [][]tgbotapi.KeyboardButton
	if len(s.QuickReplies) > 0 {
		var row []tgbotapi.KeyboardButton
		for _, btn := range s.QuickReplies {
			row = append(row, tgbotapi.NewKeyboardButton(btn.Text))
		}
		quickReplyButtons = append(quickReplyButtons, row)
	}

	// Создаем кнопки (если они есть)
	var keyboard [][]tgbotapi.InlineKeyboardButton
	if len(s.Buttons) > 0 {
		for _, btn := range s.Buttons {
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.NextState),
			))
		}
	}

	// Отправляем все сообщения с изображениями
	for i, msgPart := range s.Message {
		text := replaceTemplates(msgPart.Text, chatID)

		if i > 0 && s.Message[i-1].Sleep > 0 {
			duration := time.Duration(s.Message[i-1].Sleep * float64(time.Second))
			time.Sleep(duration)
		}

		// Настраиваем клавиатуру для этого сообщения
		var replyMarkup interface{}

		// Для последнего сообщения добавляем Inline кнопки
		if i == len(s.Message)-1 && len(inlineKeyboard) > 0 {
			replyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
		} else if len(quickReplyButtons) > 0 {
			replyMarkup = tgbotapi.NewReplyKeyboard(quickReplyButtons...)
		}

		// Если есть изображения и это первое сообщение, отправляем их с текстом
		if i == 0 && len(s.Images) > 0 {
			for _, imgPath := range s.Images {
				if _, err := os.Stat(imgPath); err == nil {
					// Создаем сообщение с фото и подписью
					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
					photo.Caption = text
					photo.ParseMode = "HTML"
					photo.ReplyMarkup = replyMarkup

					// Добавляем кнопки только к последнему сообщению с фото
					if i == len(s.Message)-1 && len(keyboard) > 0 {
						photo.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
					}

					err := botAPI.SendPhoto(photo)
					if err != nil {
						log.Printf("Ошибка отправки фото: %v", err)
					}

					// Пауза между сообщениями с фото
					time.Sleep(300 * time.Millisecond)
				}
			}
		} else {
			// Обычное текстовое сообщение
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = replyMarkup

			// Добавляем кнопки только к последнему сообщению
			if i == len(s.Message)-1 && len(keyboard) > 0 {
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
			}

			botAPI.SendMessageWithMarkup(msg)
		}
	}
}

// HandleCallbackQuery обрабатывает нажатие на кнопку и переключает состояние
func HandleCallbackQuery(botAPI *telegram.TelegramClient, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	nextState := callback.Data

	// Отвечаем на callback, чтобы убрать "часики"
	botAPI.AnswerCallbackQuery(callback.ID)

	HandleState(botAPI, chatID, nextState)
}

// replaceTemplates заменяет шаблоны в сообщении на данные пользователя
func replaceTemplates(text string, chatID int64) string {
	data, exists := userData[chatID]
	if !exists {
		return text
	}

	for key, value := range data {
		template := "{{" + key + "}}"
		text = strings.ReplaceAll(text, template, value)
	}

	return text
}
