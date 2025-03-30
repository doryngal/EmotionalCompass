package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramBotToken string `yaml:"telegram_bot_token"`
	DBConnection     string `yaml:"db_connection"`
}

// LoadConfig загружает конфигурацию из config.yaml или переменных окружения
func LoadConfig() *Config {
	config := &Config{}

	// Читаем конфиг-файл
	file, err := os.ReadFile("internal/config/config.yaml")
	if err != nil {
		log.Println("Не удалось прочитать config.yaml, используем переменные окружения")
		config.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
		config.DBConnection = os.Getenv("DB_CONNECTION")
		return config
	}

	// Разбираем YAML
	err = yaml.Unmarshal(file, config)
	if err != nil {
		log.Fatalf("Ошибка разбора config.yaml: %v", err)
	}

	return config
}
