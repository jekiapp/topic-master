FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init testscript || true
RUN go mod tidy || true
RUN go build -o test-script main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/test-script .
CMD ["./test-script"] 