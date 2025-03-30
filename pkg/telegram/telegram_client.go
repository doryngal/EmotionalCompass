package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func (c *TelegramClient) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := c.bot.Send(msg)
	return err
}

func (c *TelegramClient) GetUpdatesChan(u tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return c.bot.GetUpdatesChan(u)
}

func (c *TelegramClient) SendMessageWithMarkup(chatID int64, msg tgbotapi.MessageConfig) {
	c.bot.Send(msg)
}

func (c *TelegramClient) AnswerCallbackQuery(callbackID string) {
	c.bot.Request(tgbotapi.CallbackConfig{CallbackQueryID: callbackID})
}
