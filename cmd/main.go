package main

import (
	"fmt"
	"log"
	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"telegram-bot/pkg/telegram"
)

func main() {
	// Загрузка конфигурации.
	cfg := config.LoadConfig()

	fmt.Println(cfg)
	// Инициализация базы данных.
	//db, err := database.InitDB(cfg.DBConnection)
	//if err != nil {
	//	log.Fatalf("Ошибка подключения к базе данных: %v", err)
	//}
	//defer db.Close()

	// Инициализация Telegram API.

	botAPI, err := telegram.NewTelegramClient(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("Ошибка подключения к Telegram API: %v", err)
	}

	if err := bot.LoadStates("./internal/config/states.json"); err != nil {
		log.Fatalf("Ошибка загрузки состояний: %v", err)
	}

	bot.StartBot(botAPI)
}
