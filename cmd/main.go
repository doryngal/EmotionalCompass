package main

import (
	"log"
	"telegram-bot/internal/bot"
	"telegram-bot/internal/config"
	"telegram-bot/internal/database"
	"telegram-bot/pkg/telegram"
)

func main() {
	log.Println("[INFO] Запуск бота...")

	// Загрузка конфигурации.
	cfg, err := config.LoadConfig("./internal/config/config.yaml")
	if err != nil {
		log.Fatal("[ERROR] Ошибка загрузки конфигурации: ", err)
	}
	log.Println("[INFO] Конфигурация загружена успешно")

	// Инициализация базы данных
	db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("[ERROR] Ошибка подключения к БД: %v", err)
	}
	defer db.Close()

	// Инициализация Telegram API.
	tgClient, err := telegram.NewTelegramClient(cfg.Telegram.Token)
	if err != nil {
		log.Fatalf("[ERROR] Ошибка подключения к Telegram API: %v", err)
	}
	log.Println("[INFO] Подключение к Telegram API успешно")

	// Создание бота
	bot := bot.NewBot(tgClient, db)

	if err := bot.LoadStates(cfg); err != nil {
		log.Fatalf("[ERROR] Ошибка загрузки состояний: %v", err)
	}
	log.Println("[INFO] Состояния загружены успешно")

	log.Println("[INFO] Запуск обработчика бота...")
	bot.Start()
}
