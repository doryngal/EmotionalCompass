-- Создаем таблицу users
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    username TEXT,
    is_premium BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Создаем таблицу user_states для хранения текущего состояния пользователей
CREATE TABLE user_states (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    current_state TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Создаем таблицу user_data для хранения дополнительных данных пользователей
CREATE TABLE user_data (
    user_id BIGINT REFERENCES users(id),
    key TEXT NOT NULL,
    value TEXT,
    PRIMARY KEY (user_id, key)
);

-- Создаем таблицу messages для хранения сообщений
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    text TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW()
);