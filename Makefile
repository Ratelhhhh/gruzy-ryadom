.PHONY: help build up down logs clean

# Default target
help:
	@echo "Доступные команды:"
	@echo "  build    - Собрать все контейнеры"
	@echo "  up       - Запустить приложение"
	@echo "  down     - Остановить приложение"
	@echo "  logs     - Показать логи"
	@echo "  clean    - Очистить все контейнеры и образы"
	@echo "  clean-bin - Очистить бинарные файлы"
	@echo "  dev      - Запустить в режиме разработки"
	@echo "  prod     - Запустить в продакшн режиме"
	@echo "  config   - Проверить конфигурацию"
	@echo "  check    - Проверить работоспособность"

# Build all containers
build:
	docker-compose build

# Start the application
up:
	docker-compose up -d

# Stop the application
down:
	docker-compose down

# Show logs
logs:
	docker-compose logs -f

# Clean everything
clean:
	docker-compose down -v --rmi all
	docker system prune -f

# Clean binaries
clean-bin:
	@echo "Очистка бинарных файлов..."
	@find . -name "main" -type f -delete 2>/dev/null || true
	@find . -name "app" -type f -delete 2>/dev/null || true
	@find . -name "server" -type f -delete 2>/dev/null || true
	@find . -name "driver_bot" -type f -delete 2>/dev/null || true
	@find . -name "admin_bot" -type f -delete 2>/dev/null || true
	@find . -name "*.exe" -type f -delete 2>/dev/null || true
	@find . -name "*.test" -type f -delete 2>/dev/null || true
	@find . -name "*.out" -type f -delete 2>/dev/null || true
	@echo "✅ Бинарные файлы удалены"

# Development mode (with volume mounts for hot reload)
dev:
	docker-compose -f docker-compose.dev.yml up

# Production mode
prod:
	docker-compose -f docker-compose.prod.yml up -d

# Quick start (build and run)
start: config build up
	@echo "Приложение запущено!"
	@echo "Frontend: http://localhost"
	@echo "Backend API: http://localhost:8080"
	@echo "Database: localhost:5432"

# Check status
status:
	docker-compose ps

# Restart services
restart:
	docker-compose restart

# Update and restart
update: down build up
	@echo "Приложение обновлено и перезапущено!"

# Check configuration
config:
	@echo "Проверка конфигурации..."
	@if [ ! -f ".env" ]; then \
		echo "❌ Файл .env не найден. Скопируйте env.example в .env"; \
		exit 1; \
	fi
	@if [ ! -f "config.yaml" ]; then \
		echo "❌ Файл config.yaml не найден"; \
		exit 1; \
	fi
	@echo "✅ Конфигурация в порядке"
	@echo "📝 config.yaml - основные настройки"
	@echo "🔐 .env - токены ботов"

# Check application health
check:
	@echo "Проверка работоспособности приложения..."
	@if [ -f "scripts/check.sh" ]; then \
		./scripts/check.sh; \
	else \
		echo "❌ Скрипт проверки не найден"; \
		exit 1; \
	fi
