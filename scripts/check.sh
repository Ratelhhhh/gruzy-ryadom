#!/bin/bash

echo "🔍 Проверка единого приложения 'Грузы рядом'"
echo "=========================================="

# Проверка переменных окружения
echo "📋 Проверка переменных окружения..."
if [ ! -f ".env" ]; then
    echo "❌ Файл .env не найден"
    echo "💡 Скопируйте env.example в .env и настройте токены ботов"
    exit 1
fi

# Проверка токенов
source .env
if [ -z "$DRIVER_BOT_TOKEN" ]; then
    echo "❌ DRIVER_BOT_TOKEN не настроен"
    exit 1
fi

if [ -z "$ADMIN_BOT_TOKEN" ]; then
    echo "❌ ADMIN_BOT_TOKEN не настроен"
    exit 1
fi

echo "✅ Переменные окружения настроены"

# Проверка конфигурации
echo "📋 Проверка конфигурации..."
if [ ! -f "config.yaml" ]; then
    echo "❌ Файл config.yaml не найден"
    exit 1
fi

echo "✅ Конфигурация найдена"

# Проверка Docker
echo "🐳 Проверка Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose не установлен"
    exit 1
fi

echo "✅ Docker и Docker Compose установлены"

# Сборка и запуск
echo "🚀 Сборка и запуск приложения..."
make build

if [ $? -ne 0 ]; then
    echo "❌ Ошибка сборки"
    exit 1
fi

echo "✅ Сборка завершена"

# Запуск
echo "🚀 Запуск приложения..."
make up

# Ждем запуска
echo "⏳ Ожидание запуска сервисов..."
sleep 10

# Проверка статуса
echo "📊 Проверка статуса сервисов..."
make status

# Проверка health check
echo "🏥 Проверка health check..."
for i in {1..5}; do
    if curl -f http://localhost:8080/health &> /dev/null; then
        echo "✅ API сервер работает"
        break
    else
        echo "⏳ Попытка $i/5..."
        sleep 2
    fi
done

# Проверка базы данных
echo "🗄️ Проверка базы данных..."
if docker-compose exec -T postgres pg_isready -U postgres &> /dev/null; then
    echo "✅ База данных работает"
else
    echo "❌ Проблемы с базой данных"
fi

echo ""
echo "🎉 Проверка завершена!"
echo ""
echo "📱 Доступные сервисы:"
echo "   Frontend: http://localhost"
echo "   API: http://localhost:8080"
echo "   Health: http://localhost:8080/health"
echo ""
echo "📋 Полезные команды:"
echo "   make logs     - Показать логи"
echo "   make status   - Статус сервисов"
echo "   make down     - Остановить"
echo "   make restart  - Перезапустить" 