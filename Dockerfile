FROM golang:1.22.3-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY .env .

RUN CGO_ENABLED=0 GOOS=linux go build -o tg-bot ./cmd/telegram-bot-go/main.go

CMD ["./tg-bot"]



