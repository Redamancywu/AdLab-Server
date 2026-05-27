.PHONY: build run run-sdkapi docker-up-sdkapi docker-up-sdkapi-lite smoke-sdkapi test test-short lint migrate seed clean dev frontend-install frontend-dev frontend-build help

# 二进制输出路径
BINARY := bin/adlab-server
# 主程序入口
MAIN := cmd/server/main.go
# SDK API 入口
SDKAPI_MAIN := cmd/sdkapi/main.go
# 配置文件路径
CONFIG := config/config.yaml
LDFLAGS := -ldflags "-X adlab-server/internal/buildinfo.Version=$${ADLAB_VERSION:-dev} -X adlab-server/internal/buildinfo.GitSHA=$${ADLAB_GIT_SHA:-unknown}"

## build: 编译 Go 二进制
build:
	@echo ">>> 编译 adlab-server..."
	@mkdir -p bin
	eval "go build $(LDFLAGS) -o $(BINARY) ./$(MAIN)"
	@echo ">>> 编译完成: $(BINARY)"

## run: 直接运行（不编译到文件）
run:
	@echo ">>> 启动 adlab-server..."
	eval "go run $(LDFLAGS) ./$(MAIN)"

## run-sdkapi: 运行独立 SDK API 入口
run-sdkapi:
	@echo ">>> 启动 adlab-sdkapi..."
	eval "go run $(LDFLAGS) ./$(SDKAPI_MAIN)"

## docker-up-sdkapi: 启动 postgres + sdkapi（适合联调）
docker-up-sdkapi:
	@echo ">>> 启动 postgres + sdkapi..."
	docker compose up -d postgres sdkapi

## docker-up-sdkapi-lite: 使用独立 compose 文件启动纯 postgres + sdkapi
docker-up-sdkapi-lite:
	@echo ">>> 启动纯 sdkapi 形态..."
	docker compose -f docker-compose.sdkapi.yml up -d

## smoke-sdkapi: 对独立 sdkapi 做最小 smoke 检查
smoke-sdkapi:
	@echo ">>> 运行 sdkapi smoke 检查..."
	sh scripts/smoke_sdkapi.sh

## test: 运行所有测试
test:
	@echo ">>> 运行测试..."
	go test -v -race ./...

## test-short: 快速运行测试（跳过耗时属性测试）
test-short:
	@echo ">>> 快速测试..."
	go test -short ./...

## lint: 运行 golangci-lint
lint:
	@echo ">>> 运行 lint..."
	golangci-lint run ./...

## migrate: 执行数据库迁移（通过启动服务触发 AutoMigrate）
migrate:
	@echo ">>> 执行数据库迁移..."
	go run ./$(MAIN) --migrate-only || true

## seed: 初始化种子数据（广告位、广告源、DSP 配置、素材）
seed:
	@echo ">>> 初始化种子数据..."
	go run scripts/seed.go
	@echo ">>> 种子数据初始化完成"

## tidy: 整理依赖
tidy:
	@echo ">>> 整理 Go 依赖..."
	go mod tidy

## clean: 清理编译产物
clean:
	@echo ">>> 清理..."
	rm -rf bin/
	rm -f adlab.db

## dev: 初始化数据库并启动服务（开发快速启动）
dev: seed run

## frontend-install: 安装前端依赖
frontend-install:
	@echo ">>> 安装前端依赖..."
	cd admin-frontend && npm install

## frontend-dev: 启动前端开发服务器
frontend-dev:
	@echo ">>> 启动前端开发服务器（http://localhost:3000）..."
	cd admin-frontend && npm run dev

## frontend-build: 构建前端生产包
frontend-build:
	@echo ">>> 构建前端..."
	cd admin-frontend && npm run build

## help: 显示帮助信息
help:
	@echo "可用目标："
	@grep -E '^## ' Makefile | sed 's/## /  /'
