package models

import (
	"time"
	"github.com/google/uuid"
)

// Customer represents a customer in the system
type Customer struct {
	UUID        uuid.UUID `json:"uuid" db:"uuid"`
	Name        string    `json:"name" db:"name"`
	Phone       string    `json:"phone" db:"phone"`
	TelegramID  *int64    `json:"telegram_id,omitempty" db:"telegram_id"`
	TelegramTag *string   `json:"telegram_tag,omitempty" db:"telegram_tag"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Order represents an order in the system
type Order struct {
	UUID          uuid.UUID `json:"uuid" db:"uuid"`
	CustomerUUID  uuid.UUID `json:"customer_uuid" db:"customer_uuid"`
	Title         string    `json:"title" db:"title"`
	Description   *string   `json:"description,omitempty" db:"description"`
	WeightKg      float64   `json:"weight_kg" db:"weight_kg"`
	LengthCm      *float64  `json:"length_cm,omitempty" db:"length_cm"`
	WidthCm       *float64  `json:"width_cm,omitempty" db:"width_cm"`
	HeightCm      *float64  `json:"height_cm,omitempty" db:"height_cm"`
	FromLocation  *string   `json:"from_location,omitempty" db:"from_location"`
	ToLocation    *string   `json:"to_location,omitempty" db:"to_location"`
	Tags          []string  `json:"tags" db:"tags"`
	Price         float64   `json:"price" db:"price"`
	AvailableFrom *time.Time `json:"available_from,omitempty" db:"available_from"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	Customer      *Customer `json:"customer,omitempty"`
}

// OrderFilter represents filters for listing orders
type OrderFilter struct {
	MinWeight, MaxWeight   float64
	MinLength, MaxLength   float64
	MinWidth, MaxWidth     float64
	MinHeight, MaxHeight   float64
	MinPrice, MaxPrice     float64
	Tags                   []string
	From, To               string
	Page, Limit            int
	SortBy, SortOrder      string
}

// CustomerFilter represents filters for listing customers
type CustomerFilter struct {
	Name, Phone, TelegramTag string
	TelegramID               int64
	Page, Limit              int
	SortBy, SortOrder        string
}

// CreateOrderInput represents input for creating an order
type CreateOrderInput struct {
	CustomerUUID  uuid.UUID
	Title         string
	Description   *string
	WeightKg      float64
	LengthCm      *float64
	WidthCm       *float64
	HeightCm      *float64
	FromLocation  *string
	ToLocation    *string
	Tags          []string
	Price         float64
	AvailableFrom *time.Time
}

// UpdateOrderInput represents input for updating an order
type UpdateOrderInput struct {
	Title         *string
	Description   *string
	WeightKg      *float64
	LengthCm      *float64
	WidthCm       *float64
	HeightCm      *float64
	FromLocation  *string
	ToLocation    *string
	Tags          *[]string
	Price         *float64
	AvailableFrom *time.Time
}

// CreateCustomerInput represents input for creating a customer
type CreateCustomerInput struct {
	Name        string
	Phone       string
	TelegramID  *int64
	TelegramTag *string
}

// UpdateCustomerInput represents input for updating a customer
type UpdateCustomerInput struct {
	Name        *string
	Phone       *string
	TelegramID  *int64
	TelegramTag *string
}

// OrdersResponse represents the response for listing orders
type OrdersResponse struct {
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Total  int     `json:"total"`
	Orders []Order `json:"orders"`
}
