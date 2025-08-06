# Docker Setup для "Грузы рядом"

## 🚀 Быстрый старт

### Для Production (используя GitHub Container Registry)

1. **Настройте переменные окружения:**
   ```bash
   cp env.example .env
   # Отредактируйте .env файл, добавив ваши токены ботов
   ```

2. **Запустите приложение:**
   ```bash
   # Установите GITHUB_REPOSITORY в .env или экспортируйте переменную
   export GITHUB_REPOSITORY=your-username/gruzy-ryadom
   
   docker-compose up -d
   ```

### Для разработки (локальная сборка)

```bash
docker-compose -f docker-compose.dev.yml up -d
```

## 📦 GitHub Container Registry

### Автоматическая сборка

При каждом push в ветки `main` или `develop`, или при создании тегов `v*`, GitHub Actions автоматически:

1. Собирает Docker образы
2. Публикует их в GitHub Container Registry
3. Присваивает соответствующие теги

### Образы

- **Backend**: `ghcr.io/your-username/gruzy-ryadom/backend:main`
- **Frontend**: `ghcr.io/your-username/gruzy-ryadom/frontend:main`

### Теги

- `main` - последняя версия из ветки main
- `develop` - последняя версия из ветки develop  
- `v1.0.0` - версии по семантическому версионированию
- `main-abc123` - коммиты с SHA

## 🔧 Конфигурация

### Переменные окружения

В `.env` файле должны быть только токены ботов:
```bash
DRIVER_BOT_TOKEN=your_driver_bot_token_here
ADMIN_BOT_TOKEN=your_admin_bot_token_here
```

### Конфигурационные файлы

Все остальные настройки хранятся в `config.yaml`:
- Настройки базы данных
- Настройки сервера
- Переменные окружения

## 🏗️ Архитектура

### Backend Dockerfile

- **Multi-stage build** для оптимизации размера
- **Security**: запуск от непривилегированного пользователя
- **Health check**: автоматическая проверка здоровья
- **Optimization**: статическая компиляция с флагами оптимизации

### Frontend Dockerfile

- **Nginx**: легковесный веб-сервер
- **Static files**: статические файлы приложения
- **Custom config**: кастомная конфигурация nginx

## 🚀 Деплой

### Production

```bash
# Клонируйте репозиторий
git clone https://github.com/your-username/gruzy-ryadom.git
cd gruzy-ryadom

# Настройте переменные окружения
cp env.example .env
# Отредактируйте .env

# Запустите
docker-compose up -d
```

### Обновление

```bash
# Остановите контейнеры
docker-compose down

# Получите последние изменения
git pull

# Запустите с новыми образами
docker-compose up -d
```

## 🔍 Мониторинг

### Логи

```bash
# Все сервисы
docker-compose logs -f

# Только backend
docker-compose logs -f backend

# Только frontend
docker-compose logs -f frontend
```

### Health Check

Backend автоматически проверяет здоровье на `/health` endpoint.

### Статус контейнеров

```bash
docker-compose ps
```

## 🛠️ Разработка

### Локальная разработка

```bash
# Запуск с локальной сборкой
docker-compose -f docker-compose.dev.yml up -d

# Пересборка после изменений
docker-compose -f docker-compose.dev.yml up -d --build
```

### Отладка

```bash
# Войти в контейнер backend
docker-compose exec backend sh

# Посмотреть логи в реальном времени
docker-compose logs -f backend
``` 