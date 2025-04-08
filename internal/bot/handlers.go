package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strings"
	"time"
)

// HandleState отправляет сообщения с проверкой подписки
func (b *Bot) HandleState(chatID int64, state string) {
	s, exists := b.GetState(state)
	if !exists {
		b.API.SendHTMLMessage(chatID, "Неизвестное состояние. Попробуйте снова.")
		return
	}

	// Обновляем состояние пользователя
	if err := b.DB.SetUserState(chatID, state); err != nil {
		b.API.SendHTMLMessage(chatID, "Ошибка обработки запроса. Попробуйте позже.")
		return
	}

	// Получаем данные пользователя
	userData, err := b.DB.GetUserData(chatID)
	if err != nil {
		b.API.SendHTMLMessage(chatID, "Ошибка обработки запроса. Попробуйте позже.")
		return
	}

	// Проверяем подписку пользователя
	isPremium, err := b.DB.CheckUserPremium(chatID)
	if err != nil {
		log.Printf("Ошибка проверки подписки: %v", err)
		isPremium = false
	}

	// Создаем inline-кнопки с учетом подписки
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton
	if len(s.Buttons) > 0 {
		for _, btn := range s.Buttons {
			// Если кнопка требует подписки, проверяем статус
			if btn.RequiresPremium && !isPremium {
				btn.Text += " 💎"
				// Заменяем на fallback-состояние если нет подписки
				btn.NextState = btn.FallbackState
			}

			inlineKeyboard = append(inlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.NextState),
			))
		}
	}

	// Определяем, нужно ли показывать quick replies
	showQuickReplies := !(state == "start" || state == "user_name")

	// Создаем клавиатуру для quick replies (если нужно)
	var replyKeyboardMarkup *tgbotapi.ReplyKeyboardMarkup
	if len(s.QuickReplies) > 0 || showQuickReplies {
		quickReplies := s.QuickReplies
		if len(s.QuickReplies) > 0 {
			keyboard := createQuickReplyKeyboard(s.QuickReplies)
			replyMarkup := tgbotapi.NewReplyKeyboard(keyboard...)
			replyMarkup.ResizeKeyboard = true

			msg := tgbotapi.NewMessage(chatID, "🪷")
			msg.ReplyMarkup = replyMarkup
			b.API.SendMessageWithMarkup(msg)
		} else if len(quickReplies) == 0 {
			// Стандартные quick replies (в новом формате)
			quickReplies = []QuickReplyRow{
				{
					Buttons: []QuickReply{
						{Text: "🫣 Галерея эмоций", NextState: "all_emotions"},
					},
				},
				{
					Buttons: []QuickReply{
						{Text: "📚 Дневники", NextState: "diaries"},
						{Text: "🧘‍♂️ Медитации", NextState: "meditations"},
					},
				},
			}
			if !isPremium {
				quickReplies = append(quickReplies, QuickReplyRow{
					Buttons: []QuickReply{
						{Text: "Купить полный доступ 🚀", NextState: "buy_access"},
					},
				})
			}
		}
		keyboard := createQuickReplyKeyboard(quickReplies)
		replyMarkup := tgbotapi.NewReplyKeyboard(keyboard...)
		replyMarkup.ResizeKeyboard = true
		replyKeyboardMarkup = &replyMarkup
	}

	// Отправляем сообщения
	for i, msgPart := range s.Message {
		// Применяем задержку перед отправкой (кроме первого сообщения)
		if i > 0 && s.Message[i].Sleep > 0 {
			time.Sleep(time.Duration(s.Message[i].Sleep * float64(time.Second)))
		}

		text := b.replaceTemplates(msgPart.Text, userData)

		// Настраиваем клавиатуру только для последнего сообщения
		var finalReplyMarkup interface{}
		if i == len(s.Message)-1 {
			if len(inlineKeyboard) > 0 {
				// Для inline-кнопок создаем отдельную клавиатуру
				finalReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)
			}
		}
		// Отправка изображений с подписью
		if i == 0 && len(s.Images) > 0 {
			for _, imgPath := range s.Images {
				if _, err := os.Stat(imgPath); err == nil {
					photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
					photo.Caption = text
					photo.ParseMode = "HTML"
					photo.ReplyMarkup = finalReplyMarkup

					if err := b.API.SendPhoto(photo); err != nil {
						log.Printf("Ошибка отправки фото: %v", err)
					}
					time.Sleep(300 * time.Millisecond)
				}
			}
		} else if i == 0 && len(s.Audio) > 0 {
			for _, audioPath := range s.Audio {
				fmt.Println(audioPath)
				if _, err := os.Stat(audioPath); err == nil {
					audio := tgbotapi.NewAudio(chatID, tgbotapi.FilePath(audioPath))
					audio.Caption = text
					audio.ParseMode = "HTML"
					audio.ReplyMarkup = finalReplyMarkup
					audio.Title = "Медитация" // Можно динамически задавать название
					if err := b.API.SendAudio(audio); err != nil {

					}
					time.Sleep(300 * time.Millisecond)
				} else {
					log.Printf("Аудиофайл не найден: %s", audioPath)
				}
			}
		} else {
			// Обычное текстовое сообщение
			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "HTML"
			msg.ReplyMarkup = finalReplyMarkup

			b.API.SendMessageWithMarkup(msg)
		}
	}

	// Если есть quick replies, отправляем их отдельным сообщением
	if replyKeyboardMarkup != nil {
		msg := tgbotapi.NewMessage(chatID, "🪷")
		msg.ReplyMarkup = replyKeyboardMarkup
		b.API.SendMessageWithMarkup(msg)
	}
}

// createQuickReplyKeyboard создает клавиатуру для quick replies с учетом структуры рядов
func createQuickReplyKeyboard(replyRows []QuickReplyRow) [][]tgbotapi.KeyboardButton {
	var keyboard [][]tgbotapi.KeyboardButton

	for _, row := range replyRows {
		var keyboardRow []tgbotapi.KeyboardButton
		for _, btn := range row.Buttons {
			keyboardRow = append(keyboardRow, tgbotapi.NewKeyboardButton(btn.Text))
		}
		if len(keyboardRow) > 0 {
			keyboard = append(keyboard, keyboardRow)
		}
	}

	return keyboard
}

// HandleCallbackQuery обрабатывает нажатие на кнопку с проверкой подписки
func (b *Bot) HandleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	chatID := callback.Message.Chat.ID
	nextState := callback.Data

	// Проверяем подписку пользователя
	isPremium, err := b.DB.CheckUserPremium(chatID)
	if err != nil {
		b.API.AnswerCallbackQuery(callback.ID, "⚠️ Ошибка проверки подписки")
		return
	}

	// Ищем нажатие на премиум-кнопку
	b.statesMu.RLock()
	defer b.statesMu.RUnlock()

	for _, state := range b.states {
		for _, btn := range state.Buttons {
			if btn.NextState == nextState && btn.RequiresPremium && !isPremium {
				nextState = btn.FallbackState
				b.API.AnswerCallbackQuery(callback.ID, "🔒 Требуется подписка")
				break
			}
		}
	}

	b.API.AnswerCallbackQuery(callback.ID)
	b.HandleState(chatID, nextState)
}

func (b *Bot) replaceTemplates(text string, data map[string]string) string {
	for key, value := range data {
		template := "{{" + key + "}}"
		text = strings.ReplaceAll(text, template, value)
	}
	return text
}
