# 🚀 Настройка CI/CD для "Грузы рядом"

## ✅ Что уже настроено

1. **Упрощена структура Dockerfile'ов**:
   - Удалены лишние `Dockerfile.simple`, `Dockerfile.api-only`, `Dockerfile.test`
   - Удалена api-only версия приложения
   - Оставлен один оптимизированный `backend/Dockerfile`
   - Обновлен `frontend/Dockerfile`

2. **GitHub Actions workflow** (`.github/workflows/docker-build.yml`):
   - Автоматическая сборка при push в `main`/`develop`
   - Публикация в GitHub Container Registry
   - Поддержка тегов версий

3. **Docker Compose файлы**:
   - `docker-compose.yml` - для production (использует GitHub Container Registry)
   - `docker-compose.dev.yml` - для разработки (локальная сборка)

4. **Оптимизация**:
   - `.dockerignore` файлы для быстрой сборки
   - Multi-stage build для backend
   - Health checks
   - Security (non-root user)

## 🔧 Что нужно сделать

### 1. Настройка GitHub Repository

```bash
# Убедитесь, что репозиторий публичный или настройте доступ к packages
# В Settings -> Actions -> General -> Workflow permissions:
# ✅ Read and write permissions
# ✅ Allow GitHub Actions to create and approve pull requests
```

### 2. Первый запуск

```bash
# Клонируйте репозиторий
git clone https://github.com/your-username/gruzy-ryadom.git
cd gruzy-ryadom

# Настройте переменные окружения
cp env.example .env
# Отредактируйте .env и добавьте токены ботов

# Запустите приложение
make run
```

### 3. Проверка работы

```bash
# Проверьте статус
make status

# Посмотрите логи
make logs

# Проверьте health check
curl http://localhost:8080/health
```

## 📦 Образы в GitHub Container Registry

После первого push в `main` ветку, образы будут доступны по адресам:

- **Backend**: `ghcr.io/your-username/gruzy-ryadom/backend:main`
- **Frontend**: `ghcr.io/your-username/gruzy-ryadom/frontend:main`

## 🔄 Обновление приложения

```bash
# Получите последние изменения
git pull

# Обновите приложение
make update
```

## 🛠️ Разработка

```bash
# Локальная разработка
make dev

# Пересборка после изменений
make build
make dev
```

## 📋 Команды Makefile

- `make help` - показать все команды
- `make run` - запуск production версии
- `make dev` - запуск development версии
- `make down` - остановка приложения
- `make logs` - просмотр логов
- `make clean` - очистка Docker ресурсов
- `make update` - обновление приложения
- `make status` - статус контейнеров

## 🎯 Результат

Теперь у вас есть:
- ✅ Один оптимизированный Dockerfile для backend
- ✅ Автоматическая сборка в GitHub Actions
- ✅ Публикация в GitHub Container Registry
- ✅ Простое развертывание с `docker-compose`
- ✅ Разделение конфигурации и переменных окружения
- ✅ Health checks и security 