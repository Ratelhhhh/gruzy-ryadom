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

type AdminBot struct {
	bot     *telebot.Bot
	service *service.Service
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewAdminBot(token string, service *service.Service) (*AdminBot, error) {
	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &AdminBot{
		bot:     bot,
		service: service,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

func (b *AdminBot) Start() {
	// Admin commands
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/customers", b.handleCustomers)
	b.bot.Handle("/orders", b.handleOrders)
	b.bot.Handle("/stats", b.handleStats)
	b.bot.Handle("/broadcast", b.handleBroadcast)

	log.Println("Admin Bot started...")
	b.bot.Start()
}

func (b *AdminBot) Stop() {
	b.cancel()
	b.bot.Stop()
}

func (b *AdminBot) handleStart(c telebot.Context) error {
	user := c.Sender()
	
	// Admin check
	adminIDs := []int64{123456789} // Replace with actual admin IDs
	
	isAdmin := false
	for _, id := range adminIDs {
		if user.ID == id {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("⛔ У вас нет доступа к административной панели.")
	}

	msg := `🔧 Административная панель "Грузы рядом"

Доступные команды:
/customers - Список заказчиков
/orders - Список заказов
/stats - Статистика
/broadcast - Отправить сообщение всем пользователям
/help - Помощь`

	return c.Send(msg)
}

func (b *AdminBot) handleHelp(c telebot.Context) error {
	msg := `📋 Административные команды:

/start - Главное меню
/customers - Просмотр списка заказчиков
/orders - Просмотр списка заказов
/stats - Статистика системы
/broadcast - Массовая рассылка
/help - Показать эту справку`

	return c.Send(msg)
}

func (b *AdminBot) handleCustomers(c telebot.Context) error {
	// Admin check
	adminIDs := []int64{123456789}
	isAdmin := false
	for _, id := range adminIDs {
		if c.Sender().ID == id {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("⛔ Доступ запрещен.")
	}

	filter := models.CustomerFilter{
		Page:  1,
		Limit: 20,
	}

	customers, total, err := b.service.ListCustomers(b.ctx, filter)
	if err != nil {
		return c.Send("❌ Ошибка при получении списка заказчиков.")
	}

	if len(customers) == 0 {
		return c.Send("📭 Заказчиков пока нет.")
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("👥 Заказчиков: %d\n\n", total))

	for i, customer := range customers {
		msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, customer.Name))
		msg.WriteString(fmt.Sprintf("   📞 %s\n", customer.Phone))
		if customer.TelegramTag != nil {
			msg.WriteString(fmt.Sprintf("   📱 @%s\n", *customer.TelegramTag))
		}
		msg.WriteString(fmt.Sprintf("   📅 %s\n", customer.CreatedAt.Format("02.01.2006")))
		msg.WriteString("\n")
	}

	return c.Send(msg.String())
}

func (b *AdminBot) handleOrders(c telebot.Context) error {
	// Admin check
	adminIDs := []int64{123456789}
	isAdmin := false
	for _, id := range adminIDs {
		if c.Sender().ID == id {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("⛔ Доступ запрещен.")
	}

	filter := models.OrderFilter{
		Page:  1,
		Limit: 20,
	}

	orders, total, err := b.service.ListOrders(b.ctx, filter)
	if err != nil {
		return c.Send("❌ Ошибка при получении списка заказов.")
	}

	if len(orders) == 0 {
		return c.Send("📦 Заказов пока нет.")
	}

	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("📦 Заказов: %d\n\n", total))

	for i, order := range orders {
		msg.WriteString(fmt.Sprintf("%d. %s\n", i+1, order.Title))
		msg.WriteString(fmt.Sprintf("   ⚖️ %.1f кг\n", order.WeightKg))
		msg.WriteString(fmt.Sprintf("   💰 %.0f ₽\n", order.Price))
		if order.FromLocation != nil {
			msg.WriteString(fmt.Sprintf("   📍 Откуда: %s\n", *order.FromLocation))
		}
		if order.ToLocation != nil {
			msg.WriteString(fmt.Sprintf("   🎯 Куда: %s\n", *order.ToLocation))
		}
		msg.WriteString(fmt.Sprintf("   📅 %s\n", order.CreatedAt.Format("02.01.2006")))
		msg.WriteString("\n")
	}

	return c.Send(msg.String())
}

func (b *AdminBot) handleStats(c telebot.Context) error {
	// Admin check
	adminIDs := []int64{123456789}
	isAdmin := false
	for _, id := range adminIDs {
		if c.Sender().ID == id {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("⛔ Доступ запрещен.")
	}

	// Get basic stats
	customers, _, err := b.service.ListCustomers(b.ctx, models.CustomerFilter{})
	if err != nil {
		return c.Send("❌ Ошибка при получении статистики.")
	}

	orders, _, err := b.service.ListOrders(b.ctx, models.OrderFilter{})
	if err != nil {
		return c.Send("❌ Ошибка при получении статистики.")
	}

	msg := fmt.Sprintf(`📊 Статистика системы

👥 Заказчиков: %d
📦 Заказов: %d
📅 Дата: %s`, len(customers), len(orders), time.Now().Format("02.01.2006 15:04"))

	return c.Send(msg)
}

func (b *AdminBot) handleBroadcast(c telebot.Context) error {
	// Admin check
	adminIDs := []int64{123456789}
	isAdmin := false
	for _, id := range adminIDs {
		if c.Sender().ID == id {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		return c.Send("⛔ Доступ запрещен.")
	}

	msg := `📢 Массовая рассылка

Отправьте сообщение, которое будет разослано всем заказчикам.

Для отмены отправьте /cancel`

	return c.Send(msg)
} 