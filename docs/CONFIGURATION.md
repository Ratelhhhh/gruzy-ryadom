# Конфигурация приложения

## Обзор

Приложение использует гибридную систему конфигурации для обеспечения безопасности и гибкости:

- **`config.yaml`** - основные настройки приложения
- **`.env`** - токены ботов (чувствительные данные)

## Структура файлов

### config.yaml
Основной файл конфигурации с настройками приложения:

```yaml
# Database configuration
database:
  url: "postgres://postgres:password@localhost:5432/gruzy_ryadom?sslmode=disable"

# Server configuration
server:
  port: "8080"

# Environment
env: "development"
```

### .env
Файл с токенами ботов (не коммитится в git):

```bash
# Bot Tokens (получите у @BotFather в Telegram)
DRIVER_BOT_TOKEN=your_driver_bot_token_here
ADMIN_BOT_TOKEN=your_admin_bot_token_here
```

## Настройка

### 1. Копирование файлов
```bash
# Скопируйте примеры конфигурации
cp env.example .env
```

### 2. Настройка токенов ботов
Отредактируйте `.env` файл и добавьте реальные токены:
```bash
nano .env
```

### 3. Настройка основных параметров (опционально)
При необходимости отредактируйте `config.yaml`:
```bash
nano config.yaml
```

## Переменные окружения

### Обязательные
- `DRIVER_BOT_TOKEN` - токен бота для водителей
- `ADMIN_BOT_TOKEN` - токен административного бота

### Опциональные
Все остальные настройки находятся в `config.yaml` и имеют значения по умолчанию.

## Окружения

### Разработка
Используйте `config.yaml` с настройками для разработки.

### Продакшен
Создайте `config.production.yaml` с продакшен настройками:

```yaml
database:
  url: "postgres://user:pass@prod-db:5432/gruzy_ryadom?sslmode=require"

server:
  port: "8080"

env: "production"
```

## Docker

### Локальная разработка
```bash
make start
```

### Продакшен
```bash
# Используйте продакшен конфигурацию
cp config.production.yaml config.yaml
make start
```

## Проверка конфигурации

```bash
make config
```

Эта команда проверит наличие всех необходимых файлов конфигурации.

## Безопасность

1. **Никогда не коммитьте `.env` файл**
2. Используйте разные токены для разных окружений
3. Регулярно обновляйте токены ботов
4. Используйте сильные пароли для базы данных

## Примеры

### Локальная разработка
```yaml
# config.yaml
database:
  url: "postgres://postgres:password@localhost:5432/gruzy_ryadom?sslmode=disable"
server:
  port: "8080"
env: "development"
```

```bash
# .env
DRIVER_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
ADMIN_BOT_TOKEN=0987654321:ZYXwvuTSRqpoNMLkjihGFEdcba
```

### Продакшен
```yaml
# config.yaml
database:
  url: "postgres://prod_user:strong_password@prod-db:5432/gruzy_ryadom?sslmode=require"
server:
  port: "8080"
env: "production"
```

```bash
# .env
DRIVER_BOT_TOKEN=prod_driver_token_here
ADMIN_BOT_TOKEN=prod_admin_token_here
``` 