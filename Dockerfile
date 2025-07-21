# Build stage
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o topic-master ./cmd/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
RUN mkdir -p /app/infra/test_data/
COPY --from=builder /app/topic-master .
CMD ["./topic-master"] 