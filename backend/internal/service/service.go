package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gruzy-ryadom/internal/db"
	"gruzy-ryadom/internal/models"
)

type Service struct {
	db *db.DB
}

func New(db *db.DB) *Service {
	return &Service{db: db}
}

// Orders methods
func (s *Service) ListOrders(ctx context.Context, filter models.OrderFilter) ([]models.Order, int, error) {
	return s.db.ListOrders(ctx, filter)
}

func (s *Service) CreateOrder(ctx context.Context, input models.CreateOrderInput) (models.Order, error) {
	return s.db.CreateOrder(ctx, input)
}

func (s *Service) UpdateOrder(ctx context.Context, id string, input models.UpdateOrderInput) (models.Order, error) {
	uuid, err := parseUUID(id)
	if err != nil {
		return models.Order{}, err
	}
	return s.db.UpdateOrder(ctx, uuid, input)
}

// Customers methods
func (s *Service) ListCustomers(ctx context.Context, filter models.CustomerFilter) ([]models.Customer, int, error) {
	return s.db.ListCustomers(ctx, filter)
}

func (s *Service) CreateCustomer(ctx context.Context, input models.CreateCustomerInput) (models.Customer, error) {
	return s.db.CreateCustomer(ctx, input)
}

func (s *Service) UpdateCustomer(ctx context.Context, id string, input models.UpdateCustomerInput) (models.Customer, error) {
	uuid, err := parseUUID(id)
	if err != nil {
		return models.Customer{}, err
	}
	return s.db.UpdateCustomer(ctx, uuid, input)
}

func (s *Service) GetCustomerByTelegramID(ctx context.Context, telegramID int64) (*models.Customer, error) {
	return s.db.GetCustomerByTelegramID(ctx, telegramID)
}

// Helper functions
func parseUUID(id string) (uuid.UUID, error) {
	// Parse UUID string to UUID type
	return uuid.Parse(id)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseTags(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}
