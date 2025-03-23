# Используем официальный образ Go
FROM golang:1.24-alpine as builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o backend .

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates netcat-openbsd

COPY --from=builder /app/backend /usr/local/bin/backend
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

EXPOSE 8080

# Изменим команду запуска: ждем PostgreSQL перед стартом backend
CMD ["/wait-for-it.sh", "postgres", "5432", "--", "backend"]
