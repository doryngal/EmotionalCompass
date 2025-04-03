package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config содержит конфигурацию приложения
type Config struct {
	TelegramBotToken string         `yaml:"telegram_bot_token"`
	Database         DatabaseConfig `yaml:"database"`
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

// LoadConfig загружает конфигурацию из config.yaml или переменных окружения
func LoadConfig() *Config {
	config := &Config{}

	// Читаем конфиг-файл
	file, err := os.ReadFile("internal/config/config.yaml")
	if err != nil {
		log.Println("Не удалось прочитать config.yaml, используем переменные окружения")
		config.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
		config.Database = DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		}
		return config
	}

	// Разбираем YAML
	err = yaml.Unmarshal(file, config)
	if err != nil {
		log.Fatalf("Ошибка разбора config.yaml: %v", err)
	}

	return config
}

// getEnvAsInt получает переменную окружения как int
func getEnvAsInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultVal
}
