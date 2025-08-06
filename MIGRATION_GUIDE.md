# 🔄 Руководство по миграции

## От отдельных приложений к единому приложению

### Что изменилось

**Было:**
- `cmd/server/main.go` - REST API сервер
- `cmd/admin_bot/main.go` - Административный бот
- `cmd/driver_bot/main.go` - Бот для водителей
- Три отдельных Docker контейнера

**Стало:**
- `cmd/app/main.go` - Единое приложение (API + оба бота)
- Один Docker контейнер
- Graceful shutdown для всех компонентов

### Преимущества

✅ **Простота деплоя** - один контейнер вместо трех  
✅ **Меньше ресурсов** - общая база данных и сервисы  
✅ **Простое управление** - одна точка входа  
✅ **Graceful shutdown** - корректное завершение всех компонентов  
✅ **Единые логи** - все логи в одном месте  

### Миграция

#### 1. Остановите старые сервисы
```bash
# Если запущены старые контейнеры
docker-compose down

# Удалите старые образы (опционально)
docker system prune -f
```

#### 2. Обновите конфигурацию
```bash
# Проверьте .env файл
cp env.example .env
# Добавьте токены ботов в .env

# Проверьте config.yaml
nano config.yaml
```

#### 3. Запустите новое приложение
```bash
# Собрать и запустить
make start

# Или по шагам
make build
make up
```

#### 4. Проверьте работоспособность
```bash
# Проверка всех сервисов
make check

# Проверка статуса
make status

# Просмотр логов
make logs
```

### Проверка миграции

#### API сервер
```bash
curl http://localhost:8080/health
# Должен вернуть: OK
```

#### Боты
- Отправьте `/start` в админский бот
- Отправьте `/start` в бот для водителей
- Проверьте, что боты отвечают

#### База данных
```bash
docker-compose exec postgres psql -U postgres -d gruzy_ryadom -c "SELECT version();"
```

### Откат (если что-то пошло не так)

Если нужно вернуться к старой архитектуре:

1. **Восстановите старые main.go файлы** из git истории
2. **Обновите docker-compose.yml** для отдельных сервисов
3. **Пересоберите и запустите** старые контейнеры

```bash
# Восстановить из git
git checkout HEAD~1 -- backend/cmd/server/main.go
git checkout HEAD~1 -- backend/cmd/admin_bot/main.go
git checkout HEAD~1 -- backend/cmd/driver_bot/main.go
git checkout HEAD~1 -- docker-compose.yml

# Пересобрать
make clean
make build
make up
```

### Структура нового приложения

```
backend/cmd/app/main.go
├── Application struct
│   ├── server     *http.Server    # REST API
│   ├── adminBot   *bots.AdminBot  # Админский бот
│   ├── driverBot  *bots.DriverBot # Бот для водителей
│   ├── service    *service.Service # Бизнес-логика
│   ├── database   *db.DB          # База данных
│   └── ctx/cancel                 # Graceful shutdown
├── NewApplication()               # Инициализация
├── Start()                        # Запуск всех компонентов
└── Stop()                         # Корректное завершение
```

### Мониторинг

#### Логи единого приложения
```bash
# Все логи
make logs

# Только приложение
docker-compose logs app

# Следить за логами
docker-compose logs -f app
```

#### Health check
```bash
# Проверка API
curl http://localhost:8080/health

# Проверка базы данных
docker-compose exec postgres pg_isready -U postgres
```

#### Статус сервисов
```bash
make status
```

### Производительность

Новое единое приложение должно работать эффективнее:

- **Меньше накладных расходов** на Docker контейнеры
- **Общая память** для всех компонентов
- **Единое подключение** к базе данных
- **Меньше сетевых вызовов** между компонентами

### Безопасность

- Все токены ботов остаются в `.env` файле
- База данных изолирована в отдельном контейнере
- CORS настройки сохранены
- Graceful shutdown предотвращает потерю данных

---

🎉 **Миграция завершена!** Теперь у вас единое приложение, которое проще в управлении и эффективнее в использовании. 