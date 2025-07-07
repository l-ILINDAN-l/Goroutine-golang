# --- СТАДИЯ 1: Сборщик ---
FROM golang:1.24.3-bookworm AS builder

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем приложение
RUN go build -o /app/main ./cmd/app/main.go


# --- СТАДИЯ 2: Финальный образ ---


FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y kcat && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY docker-entrypoint.sh .
COPY --from=builder /go/bin/migrate /usr/local/bin/

COPY --from=builder /app/main .

COPY --from=builder /app/migrations ./migrations

# Копируем бинарный файл и нужные папки
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/static ./static

# Создаём пользователя с ограниченными правами
RUN addgroup --system appgroup && adduser --system --ingroup appgroup appuser
USER appuser


ENTRYPOINT ["/app/docker-entrypoint.sh"]

# Указываем команду для запуска
CMD ["./main"]

