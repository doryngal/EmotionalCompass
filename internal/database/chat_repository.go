package database

import (
	"database/sql"
)

func SaveMessage(db *sql.DB, userID int64, text string) error {
	query := `INSERT INTO messages (user_id, text) VALUES ($1, $2)`
	_, err := db.Exec(query, userID, text)
	return err
}

func GetMessages(db *sql.DB, userID int64) ([]string, error) {
	query := `SELECT text FROM messages WHERE user_id = $1`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return nil, err
		}
		messages = append(messages, text)
	}
	return messages, nil
}
