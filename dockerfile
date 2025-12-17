# Stage 1: Build
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app .

# Runtime
FROM alpine:latest

# Tambahkan timezone database
RUN apk add --no-cache tzdata

# Set timezone environment
ENV TZ=Asia/Jakarta

WORKDIR /app
COPY --from=builder /app/app .

CMD ["./app"]
