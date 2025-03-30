package bot

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"telegram-bot/pkg/telegram"
)

type State struct {
	Message string   `json:"message"`
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Text      string `json:"text"`
	NextState string `json:"next_state"`
}

var states map[string]State

func LoadStates(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &states)
}

func StartBot(botAPI *telegram.TelegramClient) {
	updates := botAPI.GetUpdatesChan(tgbotapi.NewUpdate(0)) // Теперь метод есть

	for update := range updates {
		if update.Message != nil {
			HandleState(botAPI, update.Message.Chat.ID, "start")
		} else if update.CallbackQuery != nil {
			HandleCallbackQuery(botAPI, update.CallbackQuery)
		}
	}
}
