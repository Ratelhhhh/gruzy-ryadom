package bots

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/telebot.v3"
	"gruzy-ryadom/internal/models"
	"gruzy-ryadom/internal/service"
)

type DriverBot struct {
	bot     *telebot.Bot
	service *service.Service
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewDriverBot(token string, service *service.Service) (*DriverBot, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DriverBot{
		bot:     bot,
		service: service,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (b *DriverBot) Start() {
	// Driver bot commands
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/orders", b.handleOrders)
	b.bot.Handle("/create_order", b.handleCreateOrder)
	b.bot.Handle("/profile", b.handleProfile)

	// Inline handlers
	b.bot.Handle(telebot.OnText, b.handleText)
	b.bot.Handle(telebot.OnCallback, b.handleCallback)

	log.Println("Driver Bot started...")
	b.bot.Start()
}

func (b *DriverBot) Stop() {
	b.cancel()
	b.bot.Stop()
}

func (b *DriverBot) handleStart(c telebot.Context) error {
	user := c.Sender()
	
	// Check if user exists
	customer, err := b.service.GetCustomerByTelegramID(b.ctx, user.ID)
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

		_, err = b.service.CreateCustomer(b.ctx, input)
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

func (b *DriverBot) handleHelp(c telebot.Context) error {
	msg := `📋 Помощь по командам:

/start - Начать работу с ботом
/orders - Посмотреть доступные заказы
/create_order - Создать новый заказ
/profile - Ваш профиль
/help - Показать эту справку

Для создания заказа используйте команду /create_order и следуйте инструкциям.`

	return c.Send(msg)
}

func (b *DriverBot) handleOrders(c telebot.Context) error {
	filter := models.OrderFilter{
		Page:  1,
		Limit: 10,
	}

	orders, total, err := b.service.ListOrders(b.ctx, filter)
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

func (b *DriverBot) handleCreateOrder(c telebot.Context) error {
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

func (b *DriverBot) handleProfile(c telebot.Context) error {
	user := c.Sender()
	
	customer, err := b.service.GetCustomerByTelegramID(b.ctx, user.ID)
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

func (b *DriverBot) handleText(c telebot.Context) error {
	// Handle text input for order creation
	text := c.Text()
	
	// Simple order creation from text
	// In production you should implement proper state management
	if strings.Contains(text, "кг") || strings.Contains(text, "₽") {
		return c.Send("Получена информация о заказе. Обрабатываем...")
	}

	return nil
}

func (b *DriverBot) handleCallback(c telebot.Context) error {
	// Handle inline keyboard callbacks
	return nil
} 