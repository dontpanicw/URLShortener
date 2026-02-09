# Сборка
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o urlshortener ./cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/urlshortener .

# Копируем веб-файлы
COPY --from=builder /app/web ./web

EXPOSE 8080

CMD ["./urlshortener"]
