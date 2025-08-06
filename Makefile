.PHONY: help build run dev down logs clean test

# Default target
help:
	@echo "Доступные команды:"
	@echo "  build    - Собрать Docker образы локально"
	@echo "  run      - Запустить приложение (production)"
	@echo "  dev      - Запустить приложение (development)"
	@echo "  down     - Остановить приложение"
	@echo "  logs     - Показать логи"
	@echo "  clean    - Очистить Docker ресурсы"
	@echo "  test     - Запустить тесты"

# Build Docker images locally
build:
	docker-compose -f docker-compose.dev.yml build

# Run in production mode (using GitHub Container Registry)
run:
	@if [ ! -f .env ]; then \
		echo "Создайте .env файл из env.example"; \
		cp env.example .env; \
		echo "Отредактируйте .env файл и добавьте токены ботов"; \
		exit 1; \
	fi
	docker-compose up -d

# Run in development mode (local build)
dev:
	@if [ ! -f .env ]; then \
		echo "Создайте .env файл из env.example"; \
		cp env.example .env; \
		echo "Отредактируйте .env файл и добавьте токены ботов"; \
		exit 1; \
	fi
	docker-compose -f docker-compose.dev.yml up -d

# Stop application
down:
	docker-compose down
	docker-compose -f docker-compose.dev.yml down

# Show logs
logs:
	docker-compose logs -f

# Clean Docker resources
clean:
	docker-compose down -v --remove-orphans
	docker-compose -f docker-compose.dev.yml down -v --remove-orphans
	docker system prune -f

# Run tests
test:
	cd backend && go test ./...

# Update application
update:
	git pull
	docker-compose down
	docker-compose up -d

# Show status
status:
	docker-compose ps
