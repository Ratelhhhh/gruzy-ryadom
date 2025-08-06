# Инструкция по деплою "Грузы рядом"

## Быстрый старт

### 1. Подготовка

1. Скопируйте файл с переменными окружения:
```bash
cp env.example .env
```

2. Отредактируйте `.env` файл:
```bash
nano .env
```

Добавьте ваши токены ботов (получите у @BotFather в Telegram):
- `DRIVER_BOT_TOKEN` - токен для бота водителей
- `ADMIN_BOT_TOKEN` - токен для админского бота

### 2. Запуск

```bash
# Собрать и запустить все сервисы
make start

# Или по шагам:
make build
make up
```

### 3. Проверка

После запуска приложение будет доступно по адресам:
- **Frontend**: http://localhost
- **Backend API**: http://localhost:8080
- **Database**: localhost:5432

## Управление приложением

```bash
# Показать статус сервисов
make status

# Показать логи
make logs

# Остановить приложение
make down

# Перезапустить
make restart

# Обновить и перезапустить
make update

# Очистить все (контейнеры, образы, данные)
make clean
```

## Структура приложения

Теперь у вас **одно приложение**, которое включает:
- ✅ REST API сервер
- ✅ Бот для водителей
- ✅ Админский бот
- ✅ База данных PostgreSQL
- ✅ Frontend (Nginx + статические файлы)

## Деплой на сервер

### Вариант 1: Простой VPS

1. Установите Docker и Docker Compose на сервер
2. Скопируйте проект
3. Настройте `.env` файл
4. Запустите: `make start`

### Вариант 2: Облачный провайдер

#### DigitalOcean App Platform
1. Подключите GitHub репозиторий
2. Настройте переменные окружения
3. Укажите команду запуска: `./main`

#### Heroku
1. Создайте `heroku.yml`:
```yaml
build:
  docker:
    web: backend/Dockerfile
```

2. Настройте переменные окружения в панели

#### Railway
1. Подключите GitHub репозиторий
2. Настройте переменные окружения
3. Railway автоматически определит Dockerfile

### Вариант 3: Kubernetes

Создайте `k8s/` директорию с манифестами:
- `deployment.yaml` - для приложения
- `service.yaml` - для сетевого доступа
- `configmap.yaml` - для конфигурации
- `secret.yaml` - для токенов ботов

## Мониторинг

### Health Check
```bash
curl http://localhost:8080/health
```

### Логи
```bash
# Все сервисы
make logs

# Только приложение
docker-compose logs app

# Только база данных
docker-compose logs postgres
```

## Безопасность

1. **Никогда не коммитьте `.env` файл**
2. Используйте сильные пароли для базы данных
3. Настройте firewall на сервере
4. Используйте HTTPS в продакшене

## Масштабирование

Для увеличения нагрузки:
1. Добавьте больше реплик приложения
2. Настройте балансировщик нагрузки
3. Используйте Redis для кэширования
4. Рассмотрите микросервисную архитектуру

## Резервное копирование

```bash
# Бэкап базы данных
docker-compose exec postgres pg_dump -U postgres gruzy_ryadom > backup.sql

# Восстановление
docker-compose exec -T postgres psql -U postgres gruzy_ryadom < backup.sql
``` 