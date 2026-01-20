# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 使用国内 Go 代理加速依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy dependency files first (利用 Docker 层缓存)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .
# Copy prompts
COPY --from=builder /app/prompts ./prompts
# Copy Docker-specific config (not the local one)
COPY --from=builder /app/config/config.docker.yaml ./config/config.yaml

EXPOSE 8088

CMD ["./main"]
