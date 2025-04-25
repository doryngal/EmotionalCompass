FROM golang:1.21.4

WORKDIR /app

# Копируем go.mod и go.sum и качаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Копируем скрипт ожидания БД
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# Собираем приложение
RUN go build -o bot ./cmd/main.go

# Заменим CMD на ENTRYPOINT для использования wait-for-it в docker-compose
ENTRYPOINT ["/wait-for-it.sh"]
CMD ["db:5432", "--timeout=30", "--strict", "--", "./bot"]