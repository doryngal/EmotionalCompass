package bot

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"path/filepath"
	"sync"
	"telegram-bot/internal/config"
	"telegram-bot/internal/database"
	"telegram-bot/pkg/telegram"
)

type MessagePart struct {
	Text  string  `json:"text"`
	Sleep float64 `json:"sleep"`
}

type QuickReplyRow struct {
	Buttons []QuickReply `json:"buttons"`
}

type State struct {
	Message      []MessagePart   `json:"message"`
	Buttons      []Button        `json:"buttons"`
	Images       []string        `json:"images"`
	Audio        []string        `json:"audio"`
	QuickReplies []QuickReplyRow `json:"quick_replies"` // Изменили на массив рядов
}

type QuickReply struct {
	Text      string `json:"text"`
	NextState string `json:"next_state"`
}

type Button struct {
	Text            string `json:"text"`
	NextState       string `json:"next_state"`
	RequiresPremium bool   `json:"requires_premium,omitempty"`
	FallbackState   string `json:"fallback_state,omitempty"`
}

type Bot struct {
	API      *telegram.TelegramClient
	DB       *database.DB
	states   map[string]State
	statesMu sync.RWMutex
}

func NewBot(api *telegram.TelegramClient, db *database.DB) *Bot {
	return &Bot{
		API:    api,
		DB:     db,
		states: make(map[string]State),
	}
}

// LoadConfig загружает основной конфиг
func (b *Bot) LoadConfig(configPath string) (*config.Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config config.Config
	err = json.Unmarshal(file, &config)
	return &config, err
}

// LoadStates загружает все состояния из указанных файлов
func (b *Bot) LoadStates(config *config.Config) error {
	b.statesMu.Lock()
	defer b.statesMu.Unlock()

	// Очищаем текущие состояния
	b.states = make(map[string]State)

	// Загружаем каждый файл состояний
	for _, filePath := range config.Bot.StateFiles {
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
			b.states[k] = v
		}
	}

	return nil
}

// GetState безопасно возвращает состояние по ключу
func (b *Bot) GetState(key string) (State, bool) {
	b.statesMu.RLock()
	defer b.statesMu.RUnlock()
	s, exists := b.states[key]
	return s, exists
}

func (b *Bot) Start() {
	updates := b.API.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.HandleCallbackQuery(update.CallbackQuery)
		}
	}
}
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	username := message.From.UserName

	// Получаем или создаем пользователя
	_, err := b.DB.GetOrCreateUser(chatID, username)
	if err != nil {
		b.API.SendHTMLMessage(chatID, "Ошибка обработки запроса. Попробуйте позже.")
		return
	}

	// Проверяем команду /start
	if message.IsCommand() && message.Command() == "start" {
		if err := b.DB.SetUserState(chatID, "start"); err != nil {
			b.API.SendHTMLMessage(chatID, "Ошибка обработки запроса. Попробуйте позже.")
			return
		}
		b.HandleState(chatID, "start")
		return
	}

	// Сначала проверяем все возможные quick replies
	b.statesMu.RLock()
	defer b.statesMu.RUnlock()

	// Собираем все quick replies из всех состояний
	var allQuickReplies []QuickReply
	for _, state := range b.states {
		for _, row := range state.QuickReplies {
			allQuickReplies = append(allQuickReplies, row.Buttons...)
		}
	}

	// Добавляем стандартные quick replies
	allQuickReplies = append(allQuickReplies, []QuickReply{
		{Text: "🧰 Галерея эмоций", NextState: "all_emotions"},
		{Text: "📚 Дневники", NextState: "diaries"},
		{Text: "🧘‍♂️ Медитации", NextState: "meditations"},
		{Text: "Купить полный доступ 🚀", NextState: "buy_access"},
	}...)

	// Проверяем, соответствует ли текст сообщения какой-либо quick reply
	for _, btn := range allQuickReplies {
		if message.Text == btn.Text {
			b.HandleState(chatID, btn.NextState)
			return
		}
	}

	// Если quick reply не найдена, обрабатываем как обычное сообщение
	currentState, err := b.DB.GetUserState(chatID)
	if err != nil {
		b.API.SendHTMLMessage(chatID, "Ошибка обработки запроса. Попробуйте позже.")
		return
	}

	// Обработка ввода пользователя в зависимости от текущего состояния
	switch currentState {
	case "start":
		// Сохраняем имя пользователя
		if err := b.DB.SetUserData(chatID, "Username", message.Text); err != nil {
			b.API.SendHTMLMessage(chatID, "Ошибка сохранения данных. Попробуйте позже.")
			return
		}
		if err := b.DB.SetUserState(chatID, "user_name"); err != nil {
			b.API.SendHTMLMessage(chatID, "Ошибка обработки запроса. Попробуйте позже.")
			return
		}
		b.HandleState(chatID, "user_name")
	default:
		// Для других состояний просто продолжаем диалог
		b.HandleState(chatID, currentState)
	}
}
