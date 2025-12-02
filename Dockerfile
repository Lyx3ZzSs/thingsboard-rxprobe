# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git gcc musl-dev

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=1 GOOS=linux go build -o rxprobe ./cmd/server

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制二进制文件和配置
COPY --from=builder /app/rxprobe .
COPY --from=builder /app/configs ./configs

# 设置时区
ENV TZ=Asia/Shanghai

EXPOSE 8088

CMD ["./rxprobe", "-config", "configs/config.yaml"]

