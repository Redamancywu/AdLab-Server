.PHONY: build run test migrate seed clean

# 二进制输出路径
BINARY := bin/adlab-server
# 主程序入口
MAIN := cmd/server/main.go
# 配置文件路径
CONFIG := config/config.yaml

## build: 编译 Go 二进制
build:
	@echo ">>> 编译 adlab-server..."
	@mkdir -p bin
	go build -o $(BINARY) ./$(MAIN)
	@echo ">>> 编译完成: $(BINARY)"

## run: 直接运行（不编译到文件）
run:
	@echo ">>> 启动 adlab-server..."
	go run ./$(MAIN)

## test: 运行所有测试
test:
	@echo ">>> 运行测试..."
	go test -v -race ./...

## test-short: 快速运行测试（跳过耗时属性测试）
test-short:
	@echo ">>> 快速测试..."
	go test -short ./...

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
