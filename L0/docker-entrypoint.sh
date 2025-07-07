#!/bin/sh
# Прерываем выполнение при любой ошибке
set -e

# --- Шаг 1: Ждём, пока брокер Kafka станет доступен по сети ---
echo "Waiting for Kafka broker to be available..."
until kcat -b kafka:9092 -L; do
  >&2 echo "Kafka broker is unavailable - sleeping"
  sleep 5
done
>&2 echo "Kafka broker is up."

# --- Шаг 2: Ждём, пока топик 'orders' станет доступен для записи ---
echo "Waiting for topic 'orders' to be ready..."
# Мы пытаемся отправить пустое сообщение в топик. Цикл будет повторяться,
# пока Kafka не будет готов принять сообщение (и автоматически создать топик).
until echo "init" | kcat -b kafka:9092 -t orders -P; do
    >&2 echo "Topic 'orders' not ready yet - sleeping"
    sleep 2
done
>&2 echo "Topic 'orders' is ready."


# --- Шаг 3: Выполняем миграции для баз данных ---
echo "Running database migrations..."
# ... ваш код для migrate ...
migrate -path /app/migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres-lookup:5432/orders_db?sslmode=disable" up
migrate -path /app/migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres-shard1:5432/orders_db?sslmode=disable" up
migrate -path /app/migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres-shard2:5432/orders_db?sslmode=disable" up


# --- Шаг 4: Запускаем основное приложение ---
echo "Migrations finished. Starting application..."
exec "$@"