FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o worker ./internal/adapters/pubsub/worker.go

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/worker .

CMD ["./worker"]
