FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/main.go
RUN go build -o firego ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/firego .

EXPOSE 54321

CMD ["./firego"]