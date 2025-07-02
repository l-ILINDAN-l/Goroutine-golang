# --- СТАДИЯ 1: Сборщик ---
# Меняем образ на Debian-based (bookworm) для совместимости с C-библиотеками
FROM golang:1.24.3-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем приложение
RUN go build -o /app/main ./cmd/app/main.go


# --- СТАДИЯ 2: Финальный образ ---
# Финальный образ можно оставить на Alpine, он маленький
FROM alpine:latest

# Устанавливаем только саму библиотеку librdkafka для запуска
RUN apk add --no-cache librdkafka

# ... остальной код с созданием пользователя и копированием ...
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/static ./static

# Устанавливаем пользователя после всех операций с файлами
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["./main"]