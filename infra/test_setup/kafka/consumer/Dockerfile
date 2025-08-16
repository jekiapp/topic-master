# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Install librdkafka for confluent-kafka-go
RUN apt-get update && apt-get install -y librdkafka-dev

RUN go build -o kafka-consumer consume_topics.go

# Final stage
FROM debian:bookworm-slim

WORKDIR /root/

RUN apt-get update && apt-get install -y librdkafka1 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/kafka-consumer .
COPY topics.txt .

CMD ["./kafka-consumer"] 