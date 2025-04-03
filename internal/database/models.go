package database

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	IsPremium bool   `json:"is_premium"`
}

type Message struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}
