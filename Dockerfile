FROM golang:1.24.1

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Строим приложение
RUN go build -o bot ./cmd/main.go

# Запускаем приложение
CMD ["./bot"]
