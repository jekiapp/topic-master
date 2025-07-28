# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o topic-master *.go

# Run stage
FROM alpine:latest
WORKDIR /app
RUN mkdir -p /app/infra/test_data/
COPY --from=builder /app/topic-master .
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
CMD [] 