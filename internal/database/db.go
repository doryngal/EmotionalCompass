package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"telegram-bot/internal/config"
)

type DB struct {
	*sql.DB
}

// InitDB подключается к базе данных и возвращает объект *sql.DB
func InitDB(cfg config.DatabaseConfig) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия соединения с БД: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения с БД: %w", err)
	}

	log.Println("✅ Успешное подключение к БД")
	return &DB{db}, nil
}

// User operations
func (db *DB) GetOrCreateUser(userID int64, username string) (*User, error) {
	user := &User{ID: userID, Username: username}

	// Проверяем существование пользователя
	err := db.QueryRow("SELECT username, is_premium FROM users WHERE id = $1", userID).
		Scan(&user.Username, &user.IsPremium)

	if err == sql.ErrNoRows {
		// Пользователь не существует, создаем нового
		_, err = db.Exec(
			"INSERT INTO users (id, username, is_premium) VALUES ($1, $2, $3)",
			userID, username, false,
		)
		return user, err
	}

	return user, err
}

func (db *DB) UpdateUserPremium(userID int64, isPremium bool) error {
	_, err := db.Exec("UPDATE users SET is_premium = $1 WHERE id = $2", isPremium, userID)
	return err
}

func (db *DB) GetUserState(userID int64) (string, error) {
	var state string
	err := db.QueryRow("SELECT current_state FROM user_states WHERE user_id = $1", userID).Scan(&state)
	if err == sql.ErrNoRows {
		return "start", nil
	}
	return state, err
}

func (db *DB) SetUserState(userID int64, state string) error {
	_, err := db.Exec(
		`INSERT INTO user_states (user_id, current_state) 
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET current_state = $2`,
		userID, state,
	)
	return err
}

func (db *DB) GetUserData(userID int64) (map[string]string, error) {
	rows, err := db.Query("SELECT key, value FROM user_data WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		data[key] = value
	}
	return data, nil
}

func (db *DB) SetUserData(userID int64, key, value string) error {
	_, err := db.Exec(
		`INSERT INTO user_data (user_id, key, value) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, key) 
		DO UPDATE SET value = $3`,
		userID, key, value,
	)
	return err
}

// Message operations
func (db *DB) SaveMessage(userID int64, text string) error {
	_, err := db.Exec("INSERT INTO messages (user_id, text) VALUES ($1, $2)", userID, text)
	return err
}

func (db *DB) GetMessages(userID int64) ([]Message, error) {
	rows, err := db.Query(
		"SELECT id, user_id, text, timestamp FROM messages WHERE user_id = $1 ORDER BY timestamp",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.UserID, &msg.Text, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
