# ---------- Build stage ----------
FROM golang:1.21-bullseye AS builder

WORKDIR /app

# 安装构建依赖（Debian 正确姿势）
RUN apt-get update && apt-get install -y \
    git \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o rxprobe ./cmd/server


# ---------- Runtime stage ----------
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/rxprobe .
COPY --from=builder /app/configs ./configs

ENV TZ=Asia/Shanghai

EXPOSE 8088

CMD ["./rxprobe", "-config", "configs/config.yaml"]