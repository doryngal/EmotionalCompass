package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config содержит конфигурацию приложения
type Config struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Database DatabaseConfig `yaml:"database"`
	Bot      BotConfig      `yaml:"bot"`
}

type TelegramConfig struct {
	Token string `yaml:"token"`
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

type BotConfig struct {
	StateFiles []string `yaml:"state_files"`
}

// LoadConfig загружает конфигурацию из config.yaml или переменных окружения
func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	// Читаем конфиг-файл
	file, err := os.ReadFile(path)
	if err != nil {
		log.Println("Не удалось прочитать config.yaml, используем переменные окружения")
		config.Telegram.Token = os.Getenv("TELEGRAM_BOT_TOKEN")
		config.Database = DatabaseConfig{
			Host:     os.Getenv("DB_HOST"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Database: os.Getenv("DB_NAME"),
			SSLMode:  os.Getenv("DB_SSLMODE"),
		}
		// Для StateFiles из переменных окружения можно использовать разделитель
		if stateFiles := os.Getenv("STATE_FILES"); stateFiles != "" {
			config.Bot.StateFiles = strings.Split(stateFiles, ",")
		}
		return config, nil
	}

	// Разбираем YAML
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора config.yaml: %v", err)
	}

	return config, nil
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
