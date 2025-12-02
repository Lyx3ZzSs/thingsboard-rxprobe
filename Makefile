.PHONY: build run test clean docker-build docker-run help

# 变量
APP_NAME := rxprobe
MAIN_FILE := ./cmd/server/main.go
BUILD_DIR := ./build
CONFIG_FILE := ./configs/config.yaml

# Go 相关
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

# 默认目标
.DEFAULT_GOAL := help

## help: 显示帮助信息
help:
	@echo "Thingsboard RxProbe 探针系统"
	@echo ""
	@echo "使用方法:"
	@echo "  make <target>"
	@echo ""
	@echo "目标:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: 编译项目
build:
	@echo "正在编译..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "编译完成: $(BUILD_DIR)/$(APP_NAME)"

## run: 运行项目
run:
	@echo "正在启动..."
	$(GORUN) $(MAIN_FILE) -config $(CONFIG_FILE)

## test: 运行测试
test:
	@echo "正在运行测试..."
	$(GOTEST) -v ./...

## deps: 下载依赖
deps:
	@echo "正在下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

## clean: 清理编译产物
clean:
	@echo "正在清理..."
	@rm -rf $(BUILD_DIR)
	@rm -f rxprobe.db
	@echo "清理完成"

## docker-build: 构建 Docker 镜像
docker-build:
	@echo "正在构建 Docker 镜像..."
	docker build -t $(APP_NAME):latest .

## docker-run: 运行 Docker 容器
docker-run:
	@echo "正在启动 Docker 容器..."
	docker-compose up -d

## docker-stop: 停止 Docker 容器
docker-stop:
	@echo "正在停止 Docker 容器..."
	docker-compose down

## docker-logs: 查看 Docker 日志
docker-logs:
	docker-compose logs -f

## lint: 代码检查
lint:
	@echo "正在检查代码..."
	golangci-lint run ./...

## fmt: 格式化代码
fmt:
	@echo "正在格式化代码..."
	$(GOCMD) fmt ./...

## dev: 开发模式（热重载）
dev:
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	air -c .air.toml

