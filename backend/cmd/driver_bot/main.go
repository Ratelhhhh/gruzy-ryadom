package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
	"gruzy-ryadom/internal/db"
	"gruzy-ryadom/internal/models"
	"gruzy-ryadom/internal/service"
)

type Bot struct {
	bot     *telebot.Bot
	service *service.Service
}

func NewBot(token string, service *service.Service) (*Bot, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{
		bot:     bot,
		service: service,
	}, nil
}

func (b *Bot) Start() {
	// Driver bot commands
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/orders", b.handleOrders)
	b.bot.Handle("/create_order", b.handleCreateOrder)
	b.bot.Handle("/profile", b.handleProfile)

	// Admin bot commands
	b.bot.Handle("/admin", b.handleAdmin)
	b.bot.Handle("/customers", b.handleCustomers)
	b.bot.Handle("/stats", b.handleStats)

	// Inline handlers
	b.bot.Handle(telebot.OnText, b.handleText)
	b.bot.Handle(telebot.OnCallback, b.handleCallback)

	log.Println("Bot started...")
	b.bot.Start()
}

func (b *Bot) handleStart(c telebot.Context) error {
	user := c.Sender()
	
	// Check if user exists
	customer, err := b.service.GetCustomerByTelegramID(context.Background(), user.ID)
	if err != nil {
		return c.Send("Произошла ошибка при проверке профиля.")
	}

	if customer == nil {
		// Create new customer
		input := models.CreateCustomerInput{
			Name:       user.FirstName + " " + user.LastName,
			Phone:      "", // Will be asked later
			TelegramID: &user.ID,
		}
		if user.Username != "" {
			input.TelegramTag = &user.Username
		}

		_, err = b.service.CreateCustomer(context.Background(), input)
		if err != nil {
			return c.Send("Произошла ошибка при создании профиля.")
		}
	}

	msg := `🚛 Добро пожаловать в "Грузы рядом"!

Доступные команды:
/orders - Посмотреть доступные заказы
/create_order - Создать новый заказ
/profile - Ваш профиль
/help - Помощь`

	return c.Send(msg)
}

func (b *Bot) handleHelp(c telebot.Context) error {
	msg := `📋 Помощь по командам:

/start - Начать работу с ботом
/orders - Посмотреть доступные заказы
/create_order - Создать новый заказ
/profile - Ваш профиль
/help - Показать эту справку

Для создания заказа используйте команду /create_order и следуйте инструкциям.`

	return c.Send(msg)
}

func (b *Bot) handleOrders(c telebot.Context) error {
	filter := models.OrderFilter{
		Page:  1,
		Limit: 10,
	}

	orders, total, err := b.service.ListOrders(context.Background(), filter)
	if err != nil {
		return c.Send("Произошла ошибка при получении заказов.")
	}

	if len(orders) == 0 {
		return c.Send("Пока нет доступных заказов.")
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("📦 Найдено заказов: %d\n\n", total))

	for i, order := range orders {
		msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, order.Title))
		msg.WriteString(fmt.Sprintf("   Вес: %.1f кг\n", order.WeightKg))
		msg.WriteString(fmt.Sprintf("   Цена: %.0f ₽\n", order.Price))
		if order.FromLocation != nil {
			msg.WriteString(fmt.Sprintf("   Откуда: %s\n", *order.FromLocation))
		}
		if order.ToLocation != nil {
			msg.WriteString(fmt.Sprintf("   Куда: %s\n", *order.ToLocation))
		}
		msg.WriteString("\n")
	}

	return c.Send(msg.String())
}

func (b *Bot) handleCreateOrder(c telebot.Context) error {
	msg := `📝 Создание нового заказа

Пожалуйста, отправьте информацию о заказе в следующем формате:

Название заказа
Вес (кг)
Цена (₽)
Откуда
Куда
Описание (необязательно)

Пример:
Перевезти холодильник
70
5000
Москва
Казань
Тонкости: грузить только стоя`

	return c.Send(msg)
}

func (b *Bot) handleProfile(c telebot.Context) error {
	user := c.Sender()
	
	customer, err := b.service.GetCustomerByTelegramID(context.Background(), user.ID)
	if err != nil {
		return c.Send("Произошла ошибка при получении профиля.")
	}

	if customer == nil {
		return c.Send("Профиль не найден. Используйте /start для создания профиля.")
	}

	msg := fmt.Sprintf(`👤 Ваш профиль:

Имя: %s
Телефон: %s`, customer.Name, customer.Phone)

	if customer.TelegramTag != nil {
		msg += fmt.Sprintf("\nTelegram: @%s", *customer.TelegramTag)
	}

	return c.Send(msg)
}

func (b *Bot) handleAdmin(c telebot.Context) error {
	// Simple admin check - in production you should have proper admin management
	adminIDs := []int64{123456789} // Replace with actual admin IDs
	
	isAdmin := false
	for _, id := range adminIDs {
		if c.Sender().ID == id {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("У вас нет доступа к административным функциям.")
	}

	msg := `🔧 Административная панель

Доступные команды:
/customers - Список заказчиков
/stats - Статистика`

	return c.Send(msg)
}

func (b *Bot) handleCustomers(c telebot.Context) error {
	filter := models.CustomerFilter{
		Page:  1,
		Limit: 10,
	}

	customers, total, err := b.service.ListCustomers(context.Background(), filter)
	if err != nil {
		return c.Send("Произошла ошибка при получении списка заказчиков.")
	}

	if len(customers) == 0 {
		return c.Send("Заказчиков пока нет.")
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("👥 Заказчиков: %d\n\n", total))

	for i, customer := range customers {
		msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, customer.Name))
		msg.WriteString(fmt.Sprintf("   Телефон: %s\n", customer.Phone))
		msg.WriteString(fmt.Sprintf("   Telegram: @%s\n", *customer.TelegramTag))
		msg.WriteString("\n")
	}

	return c.Send(msg.String())
}

func (b *Bot) handleStats(c telebot.Context) error {
	// Simple stats - in production you should implement proper statistics
	msg := `📊 Статистика

Заказов: 0
Заказчиков: 0
Активных заказов: 0`

	return c.Send(msg)
}

func (b *Bot) handleText(c telebot.Context) error {
	// Handle text input for order creation
	text := c.Text()
	
	// Simple order creation from text
	// In production you should implement proper state management
	if strings.Contains(text, "кг") || strings.Contains(text, "₽") {
		return c.Send("Получена информация о заказе. Обрабатываем...")
	}

	return nil
}

func (b *Bot) handleCallback(c telebot.Context) error {
	// Handle inline keyboard callbacks
	return nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database, err := db.New(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Service layer
	svc := service.New(database)

	// Bot token
	botToken := os.Getenv("DRIVER_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("DRIVER_BOT_TOKEN is required")
	}

	// Create and start bot
	bot, err := NewBot(botToken, svc)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Start()
}
