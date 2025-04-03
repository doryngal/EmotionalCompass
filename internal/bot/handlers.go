// handlers.go
package bot

import (
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
	"telegram-bot/pkg/telegram"
	"time"
)

var db *sql.DB

func SetDatabase(database *sql.DB) {
	db = database
}

// Проверяем, есть ли у пользователя подписка
func isUserPremium(userID int64) (bool, error) {
	var isPremium bool
	fmt.Println(userID)
	err := db.QueryRow("SELECT is_premium FROM users WHERE id = $1", userID).Scan(&isPremium)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Пользователь не найден, считаем что без подписки
		}
		return false, err
	}
	return isPremium, nil
}

// HandleState отправляет сообщения с проверкой подписки
func HandleState(botAPI *telegram.TelegramClient, chatID int64, state string) {
	s, exists := GetState(state)
	if !exists {
		botAPI.SendHTMLMessage(chatID, "Неизвестное состояние. Попробуйте снова.")
		return
	}

	// Обновляем состояние пользователя
	userStates[chatID] = state

	// Создаем кнопки с учетом подписки
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	if len(s.Buttons) > 0 {
		for _, btn := range s.Buttons {
			// Если кнопка требует подписки, проверяем статус
			if btn.RequiresPremium {
				premium, err := isUserPremium(chatID)
				fmt.Println(premium)
				if err != nil {
					log.Printf("Ошибка проверки подписки: %v", err)
					continue
				}

				if !premium {
					// Заменяем на fallback-состояние если нет подписки
					btn.NextState = btn.FallbackState
				}
			}

			inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.NextState),
			))
		}
	}

	// Отправляем сообщения
	for i, msgPart := range s.Message {
		// Применяем задержку перед отправкой (кроме первого сообщения)
		if i > 0 && s.Message[i-1].Sleep > 0 {
			time.Sleep(time.Duration(s.Message[i-1].Sleep * float64(time.Second)))
		}

		text := replaceTemplates(msgPart.Text, chatID)

		// Настраиваем клавиатуру только для последнего сообщения
		var replyMarkup interface{}
		if i == len(s.Message)-1 && len(inlineKeyboard) > 0 {
			replyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
		}

		// Отправка изображений с подписью
		if i == 0 && len(s.Images) > 0 {
			for _, imgPath := range s.Images {
				if _, err := os.Stat(imgPath); err == nil {
					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
					photo.Caption = text
					photo.ParseMode = "HTML"
					photo.ReplyMarkup = replyMarkup

					if err := botAPI.SendPhoto(photo); err != nil {
						log.Printf("Ошибка отправки фото: %v", err)
					}
					time.Sleep(300 * time.Millisecond)
				}
			}
		} else {
			// Обычное текстовое сообщение
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = replyMarkup

			botAPI.SendMessageWithMarkup(msg)
		}
	}
}

// HandleCallbackQuery обрабатывает нажатие на кнопку с проверкой подписки
func HandleCallbackQuery(botAPI *telegram.TelegramClient, callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	nextState := callback.Data

	// Ищем нажатие на премиум-кнопку
	for _, state := range states {
		for _, btn := range state.Buttons {
			if btn.NextState == nextState && btn.RequiresPremium {
				premium, err := isUserPremium(chatID)
				if err != nil {
					botAPI.AnswerCallbackQuery(callback.ID, "⚠️ Ошибка проверки подписки")
					return
				}

				if !premium {
					nextState = btn.FallbackState
					botAPI.AnswerCallbackQuery(callback.ID, "🔒 Требуется подписка")
				}
				break
			}
		}
	}

	botAPI.AnswerCallbackQuery(callback.ID)
	HandleState(botAPI, chatID, nextState)
}

// replaceTemplates остается без изменений
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
