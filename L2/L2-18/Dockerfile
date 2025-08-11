FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /calendar ./cmd/calendar/main.go


FROM alpine:latest

COPY --from=builder /calendar /calendar

COPY ./configs /configs

EXPOSE 8081

ENTRYPOINT ["/calendar"]