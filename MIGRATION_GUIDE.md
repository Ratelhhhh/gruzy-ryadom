# Руководство по миграции конфигурации

## Что изменилось

Приложение теперь использует гибридную систему конфигурации:
- **Основные настройки** → `config.yaml`
- **Токены ботов** → `.env` (как и раньше)

## Шаги миграции

### 1. Обновите .env файл
Удалите из `.env` все настройки кроме токенов ботов:

**Было:**
```bash
# Database
DATABASE_URL=postgres://postgres:password@localhost:5432/gruzy_ryadom?sslmode=disable

# Server
PORT=8080

# Bot Tokens
DRIVER_BOT_TOKEN=your_driver_bot_token_here
ADMIN_BOT_TOKEN=your_admin_bot_token_here

# Environment
ENV=development
```

**Стало:**
```bash
# Bot Tokens (получите у @BotFather в Telegram)
DRIVER_BOT_TOKEN=your_driver_bot_token_here
ADMIN_BOT_TOKEN=your_admin_bot_token_here
```

### 2. Проверьте config.yaml
Убедитесь, что `config.yaml` содержит правильные настройки:

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

### 3. Проверьте конфигурацию
```bash
make config
```

### 4. Запустите приложение
```bash
make start
```

## Проверка работоспособности

1. **Проверьте конфигурацию:**
   ```bash
   make config
   ```

2. **Запустите приложение:**
   ```bash
   make start
   ```

3. **Проверьте доступность:**
   - Frontend: http://localhost
   - API: http://localhost:8080/health

## Обратная совместимость

- API остался без изменений
- База данных не изменилась
- Боты работают как раньше
- Изменилась только система конфигурации

## Проблемы?

1. **Ошибка "config.yaml not found"**
   - Убедитесь, что файл `config.yaml` существует в корне проекта

2. **Ошибка "DRIVER_BOT_TOKEN is required"**
   - Проверьте, что в `.env` файле есть токены ботов

3. **Ошибка "failed to parse config.yaml"**
   - Проверьте синтаксис YAML в `config.yaml`

## Дополнительная документация

- [Подробная документация по конфигурации](docs/CONFIGURATION.md)
- [Быстрый старт](QUICKSTART.md)
- [Инструкция по деплою](DEPLOY.md) 