package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TelegramClient struct {
	bot *tgbotapi.BotAPI
}

func NewTelegramClient(token string) (*TelegramClient, error) {
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &TelegramClient{bot: botAPI}, nil
}

func (c *TelegramClient) GetUpdatesChan(u tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return c.bot.GetUpdatesChan(u)
}

func (c *TelegramClient) AnswerCallbackQuery(callbackID string, text ...string) {
	cfg := tgbotapi.CallbackConfig{CallbackQueryID: callbackID}
	if len(text) > 0 {
		cfg.Text = text[0]
		cfg.ShowAlert = true
	}
	c.bot.Request(cfg)
}

func (c *TelegramClient) SendHTMLMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	_, err := c.bot.Send(msg)
	return err
}

func (c *TelegramClient) SendMessageWithMarkup(msg tgbotapi.MessageConfig) {
	c.bot.Send(msg)
}

func (c *TelegramClient) SendPhoto(photo tgbotapi.PhotoConfig) error {
	_, err := c.bot.Send(photo)
	return err
}
