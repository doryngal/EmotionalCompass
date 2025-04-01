package bot

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"path/filepath"
	"sync"
	"telegram-bot/pkg/telegram"
)

type MessagePart struct {
	Text  string  `json:"text"`
	Sleep float64 `json:"sleep"`
}

type State struct {
	Message      []MessagePart `json:"message"`
	Buttons      []Button      `json:"buttons"`
	Images       []string      `json:"images"`
	QuickReplies []QuickReply  `json:"quick_replies"`
}

type QuickReply struct {
	Text      string `json:"text"`
	NextState string `json:"next_state"`
}

type Button struct {
	Text      string `json:"text"`
	NextState string `json:"next_state"`
}

type Config struct {
	StateFiles []string `json:"state_files"`
}

var (
	userStates = make(map[int64]string)
	userData   = make(map[int64]map[string]string)

	states     = make(map[string]State)
	statesLock sync.RWMutex
)

// LoadConfig загружает основной конфиг
func LoadConfig(configPath string) (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	return &config, err
}

// LoadStates загружает все состояния из указанных файлов
func LoadStates(config *Config) error {
	statesLock.Lock()
	defer statesLock.Unlock()

	// Очищаем текущие состояния
	states = make(map[string]State)

	// Загружаем каждый файл состояний
	for _, filePath := range config.StateFiles {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return err
		}

		file, err := os.ReadFile(absPath)
		if err != nil {
			return err
		}

		var fileStates map[string]State
		if err := json.Unmarshal(file, &fileStates); err != nil {
			return err
		}

		// Объединяем состояния
		for k, v := range fileStates {
			states[k] = v
		}
	}

	return nil
}

// GetState безопасно возвращает состояние по ключу
func GetState(key string) (State, bool) {
	statesLock.RLock()
	defer statesLock.RUnlock()
	s, exists := states[key]
	return s, exists
}

func StartBot(botAPI *telegram.TelegramClient) {
	updates := botAPI.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for update := range updates {
		if update.Message != nil {
			handleMessage(botAPI, update.Message)
		} else if update.CallbackQuery != nil {
			HandleCallbackQuery(botAPI, update.CallbackQuery)
		}
	}
}

func handleMessage(botAPI *telegram.TelegramClient, message *tgbotapi.Message) {
	chatID := message.Chat.ID

	for _, s := range states {
		for _, btn := range s.QuickReplies {
			if message.Text == btn.Text {
				HandleState(botAPI, chatID, btn.NextState)
				return
			}
		}
	}

	currentState, exists := userStates[chatID]

	if !exists {
		// Новый пользователь - начинаем диалог
		userStates[chatID] = "start"
		userData[chatID] = make(map[string]string)
		HandleState(botAPI, chatID, "start")
		return
	}

	// Обработка ввода пользователя в зависимости от текущего состояния
	switch currentState {
	case "start":
		// Сохраняем имя пользователя
		userData[chatID]["Username"] = message.Text
		userStates[chatID] = "user_name"
		HandleState(botAPI, chatID, "user_name")
	default:
		// Для других состояний просто продолжаем диалог
		HandleState(botAPI, chatID, currentState)
	}
}
