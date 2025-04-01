package main

import (
	"log"
	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"telegram-bot/pkg/telegram"
)

func main() {
	log.Println("[INFO] Запуск бота...")

	// Загрузка конфигурации.
	cfg := config.LoadConfig()
	if cfg == nil {
		log.Fatal("[ERROR] Ошибка загрузки конфигурации")
	}
	log.Println("[INFO] Конфигурация загружена успешно")

	// Инициализация Telegram API.
	botAPI, err := telegram.NewTelegramClient(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("[ERROR] Ошибка подключения к Telegram API: %v", err)
	}
	log.Println("[INFO] Подключение к Telegram API успешно")

	stateConfig, err := bot.LoadConfig("./config/main.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Загрузка состояний
	if err := bot.LoadStates(stateConfig); err != nil {
		log.Fatalf("[ERROR] Ошибка загрузки состояний: %v", err)
	}
	log.Println("[INFO] Состояния загружены успешно")

	log.Println("[INFO] Запуск обработчика бота...")
	bot.StartBot(botAPI)
}
