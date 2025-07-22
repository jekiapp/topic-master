FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY consumer/* .
RUN go mod tidy
RUN go build -o consumer main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache curl
COPY --from=builder /app/consumer .
COPY topics.txt .
COPY consumer/start_consumer.sh .
RUN chmod +x start_consumer.sh
ENTRYPOINT ["/app/start_consumer.sh"] 