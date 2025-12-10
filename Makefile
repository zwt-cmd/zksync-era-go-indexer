.PHONY: help up down logs db redis clean

help:
	@echo "SyncSwap 扫链项目"
	@echo "================="
	@echo ""
	@echo "Docker 命令:"
	@echo "  make up       - 启动服务"
	@echo "  make down     - 停止服务"
	@echo "  make logs     - 查看日志"
	@echo "  make db       - 进入MySQL"
	@echo "  make redis    - 进入Redis"
	@echo "  make clean    - 删除所有数据"
	@echo ""
	@echo "Go 命令:"
	@echo "  make deps     - 安装依赖"
	@echo "  make run      - 运行程序"
	@echo ""

# 启动 Docker（自动检测平台）
up:
	@echo "🚀 启动 Docker 服务..."
	@if [ "$$(uname -s)" = "Darwin" ] && [ "$$(uname -m)" = "arm64" ]; then \
		echo "检测到 Mac ARM (M1/M2/M3)"; \
		docker-compose -f docker/docker-compose.mac-arm.yml up -d; \
	elif [ "$$(uname -s)" = "Linux" ]; then \
		echo "检测到 Linux"; \
		docker-compose -f docker/docker-compose.yml up -d; \
	else \
		echo "检测到 Mac Intel / 其他"; \
		docker-compose -f docker/docker-compose.yml up -d; \
	fi
	@echo ""
	@echo "✅ 启动完成！"
	@echo ""
	@echo "MySQL:  localhost:3307 (用户: scanner / 密码: scannerpass)"
	@echo "Redis:  localhost:6380"
	@echo ""
	@echo "💡 使用 Navicat 连接 MySQL, Another Redis 连接 Redis"

# 停止服务
down:
	@echo "⏹️  停止服务..."
	@docker-compose -f docker/docker-compose.yml down 2>/dev/null || true
	@docker-compose -f docker/docker-compose.mac-arm.yml down 2>/dev/null || true
	@docker-compose -f docker/docker-compose.windows.yml down 2>/dev/null || true

# 查看日志
logs:
	@docker logs -f syncswap_mysql syncswap_redis 2>/dev/null || docker-compose logs -f

# 进入 MySQL
db:
	@docker exec -it syncswap_mysql mysql -uscanner -pscannerpass syncswap

# 进入 Redis
redis:
	@docker exec -it syncswap_redis redis-cli

# 删除所有数据
clean:
	@echo "⚠️  警告：将删除所有数据！"
	@read -p "确认删除？[y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose -f docker/docker-compose.yml down -v 2>/dev/null || true; \
		docker-compose -f docker/docker-compose.mac-arm.yml down -v 2>/dev/null || true; \
		docker-compose -f docker/docker-compose.windows.yml down -v 2>/dev/null || true; \
		echo "✅ 已清理"; \
	fi

# 安装 Go 依赖
deps:
	@echo "📦 安装依赖..."
	@go mod download
	@go mod tidy

# 运行程序
run:
	@go run main.go

