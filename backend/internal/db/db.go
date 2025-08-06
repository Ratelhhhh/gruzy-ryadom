package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gruzy-ryadom/internal/models"
)

type DB struct {
	*sql.DB
}

func New(dsn string) (*DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{db}, nil
}

// Orders methods
func (db *DB) ListOrders(ctx context.Context, filter models.OrderFilter) ([]models.Order, int, error) {
	query := `
		SELECT 
			o.uuid, o.customer_uuid, o.title, o.description, o.weight_kg,
			o.length_cm, o.width_cm, o.height_cm, o.from_location, o.to_location,
			o.tags, o.price, o.available_from, o.created_at,
			c.uuid, c.name, c.phone, c.telegram_id, c.telegram_tag, c.created_at
		FROM orders o
		JOIN customers c ON o.customer_uuid = c.uuid
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	// Add filters
	if filter.MinWeight > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.weight_kg >= $%d", argCount)
		args = append(args, filter.MinWeight)
	}
	if filter.MaxWeight > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.weight_kg <= $%d", argCount)
		args = append(args, filter.MaxWeight)
	}
	if filter.MinLength > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.length_cm >= $%d", argCount)
		args = append(args, filter.MinLength)
	}
	if filter.MaxLength > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.length_cm <= $%d", argCount)
		args = append(args, filter.MaxLength)
	}
	if filter.MinWidth > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.width_cm >= $%d", argCount)
		args = append(args, filter.MinWidth)
	}
	if filter.MaxWidth > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.width_cm <= $%d", argCount)
		args = append(args, filter.MaxWidth)
	}
	if filter.MinHeight > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.height_cm >= $%d", argCount)
		args = append(args, filter.MinHeight)
	}
	if filter.MaxHeight > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.height_cm <= $%d", argCount)
		args = append(args, filter.MaxHeight)
	}
	if filter.MinPrice > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.price >= $%d", argCount)
		args = append(args, filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.price <= $%d", argCount)
		args = append(args, filter.MaxPrice)
	}
	if len(filter.Tags) > 0 {
		argCount++
		query += fmt.Sprintf(" AND o.tags && $%d", argCount)
		args = append(args, pq.Array(filter.Tags))
	}
	if filter.From != "" {
		argCount++
		query += fmt.Sprintf(" AND o.from_location ILIKE $%d", argCount)
		args = append(args, "%"+filter.From+"%")
	}
	if filter.To != "" {
		argCount++
		query += fmt.Sprintf(" AND o.to_location ILIKE $%d", argCount)
		args = append(args, "%"+filter.To+"%")
	}

	// Add sorting
	if filter.SortBy != "" {
		query += " ORDER BY "
		switch filter.SortBy {
		case "price":
			query += "o.price"
		case "weight":
			query += "o.weight_kg"
		case "price/weight":
			query += "o.price / o.weight_kg"
		default:
			query += "o.created_at"
		}
		if filter.SortOrder == "desc" {
			query += " DESC"
		} else {
			query += " ASC"
		}
	} else {
		query += " ORDER BY o.created_at DESC"
	}

	// Add pagination
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, filter.Limit)
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		var customer models.Customer
		var description, fromLocation, toLocation sql.NullString
		var lengthCm, widthCm, heightCm sql.NullFloat64
		var availableFrom sql.NullTime
		var telegramID sql.NullInt64
		var telegramTag sql.NullString

		err := rows.Scan(
			&order.UUID, &order.CustomerUUID, &order.Title, &description, &order.WeightKg,
			&lengthCm, &widthCm, &heightCm, &fromLocation, &toLocation,
			pq.Array(&order.Tags), &order.Price, &availableFrom, &order.CreatedAt,
			&customer.UUID, &customer.Name, &customer.Phone, &telegramID, &telegramTag, &customer.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		if description.Valid {
			order.Description = &description.String
		}
		if fromLocation.Valid {
			order.FromLocation = &fromLocation.String
		}
		if toLocation.Valid {
			order.ToLocation = &toLocation.String
		}
		if lengthCm.Valid {
			order.LengthCm = &lengthCm.Float64
		}
		if widthCm.Valid {
			order.WidthCm = &widthCm.Float64
		}
		if heightCm.Valid {
			order.HeightCm = &heightCm.Float64
		}
		if availableFrom.Valid {
			order.AvailableFrom = &availableFrom.Time
		}
		if telegramID.Valid {
			customer.TelegramID = &telegramID.Int64
		}
		if telegramTag.Valid {
			customer.TelegramTag = &telegramTag.String
		}

		order.Customer = &customer
		orders = append(orders, order)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM orders o
		WHERE 1=1
	`
	countArgs := []interface{}{}
	argCount = 0

	// Add same filters for count
	if filter.MinWeight > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.weight_kg >= $%d", argCount)
		countArgs = append(countArgs, filter.MinWeight)
	}
	if filter.MaxWeight > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.weight_kg <= $%d", argCount)
		countArgs = append(countArgs, filter.MaxWeight)
	}
	if filter.MinLength > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.length_cm >= $%d", argCount)
		countArgs = append(countArgs, filter.MinLength)
	}
	if filter.MaxLength > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.length_cm <= $%d", argCount)
		countArgs = append(countArgs, filter.MaxLength)
	}
	if filter.MinWidth > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.width_cm >= $%d", argCount)
		countArgs = append(countArgs, filter.MinWidth)
	}
	if filter.MaxWidth > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.width_cm <= $%d", argCount)
		countArgs = append(countArgs, filter.MaxWidth)
	}
	if filter.MinHeight > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.height_cm >= $%d", argCount)
		countArgs = append(countArgs, filter.MinHeight)
	}
	if filter.MaxHeight > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.height_cm <= $%d", argCount)
		countArgs = append(countArgs, filter.MaxHeight)
	}
	if filter.MinPrice > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.price >= $%d", argCount)
		countArgs = append(countArgs, filter.MinPrice)
	}
	if filter.MaxPrice > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.price <= $%d", argCount)
		countArgs = append(countArgs, filter.MaxPrice)
	}
	if len(filter.Tags) > 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND o.tags && $%d", argCount)
		countArgs = append(countArgs, pq.Array(filter.Tags))
	}
	if filter.From != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND o.from_location ILIKE $%d", argCount)
		countArgs = append(countArgs, "%"+filter.From+"%")
	}
	if filter.To != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND o.to_location ILIKE $%d", argCount)
		countArgs = append(countArgs, "%"+filter.To+"%")
	}

	var total int
	err = db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return orders, total, nil
}

func (db *DB) CreateOrder(ctx context.Context, input models.CreateOrderInput) (models.Order, error) {
	query := `
		INSERT INTO orders (
			customer_uuid, title, description, weight_kg, length_cm, width_cm, height_cm,
			from_location, to_location, tags, price, available_from
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING uuid, customer_uuid, title, description, weight_kg, length_cm, width_cm, height_cm,
			from_location, to_location, tags, price, available_from, created_at
	`

	var order models.Order
	var description, fromLocation, toLocation sql.NullString
	var lengthCm, widthCm, heightCm sql.NullFloat64
	var availableFrom sql.NullTime

	err := db.QueryRowContext(ctx, query,
		input.CustomerUUID, input.Title, input.Description, input.WeightKg,
		input.LengthCm, input.WidthCm, input.HeightCm, input.FromLocation, input.ToLocation,
		pq.Array(input.Tags), input.Price, input.AvailableFrom,
	).Scan(
		&order.UUID, &order.CustomerUUID, &order.Title, &description, &order.WeightKg,
		&lengthCm, &widthCm, &heightCm, &fromLocation, &toLocation,
		pq.Array(&order.Tags), &order.Price, &availableFrom, &order.CreatedAt,
	)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to create order: %w", err)
	}

	if description.Valid {
		order.Description = &description.String
	}
	if fromLocation.Valid {
		order.FromLocation = &fromLocation.String
	}
	if toLocation.Valid {
		order.ToLocation = &toLocation.String
	}
	if lengthCm.Valid {
		order.LengthCm = &lengthCm.Float64
	}
	if widthCm.Valid {
		order.WidthCm = &widthCm.Float64
	}
	if heightCm.Valid {
		order.HeightCm = &heightCm.Float64
	}
	if availableFrom.Valid {
		order.AvailableFrom = &availableFrom.Time
	}

	return order, nil
}

func (db *DB) UpdateOrder(ctx context.Context, id uuid.UUID, input models.UpdateOrderInput) (models.Order, error) {
	query := "UPDATE orders SET "
	args := []interface{}{}
	argCount := 0

	updates := []string{}

	if input.Title != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("title = $%d", argCount))
		args = append(args, *input.Title)
	}
	if input.Description != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("description = $%d", argCount))
		args = append(args, *input.Description)
	}
	if input.WeightKg != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("weight_kg = $%d", argCount))
		args = append(args, *input.WeightKg)
	}
	if input.LengthCm != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("length_cm = $%d", argCount))
		args = append(args, *input.LengthCm)
	}
	if input.WidthCm != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("width_cm = $%d", argCount))
		args = append(args, *input.WidthCm)
	}
	if input.HeightCm != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("height_cm = $%d", argCount))
		args = append(args, *input.HeightCm)
	}
	if input.FromLocation != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("from_location = $%d", argCount))
		args = append(args, *input.FromLocation)
	}
	if input.ToLocation != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("to_location = $%d", argCount))
		args = append(args, *input.ToLocation)
	}
	if input.Tags != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("tags = $%d", argCount))
		args = append(args, pq.Array(*input.Tags))
	}
	if input.Price != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("price = $%d", argCount))
		args = append(args, *input.Price)
	}
	if input.AvailableFrom != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("available_from = $%d", argCount))
		args = append(args, *input.AvailableFrom)
	}

	if len(updates) == 0 {
		return models.Order{}, fmt.Errorf("no fields to update")
	}

	query += strings.Join(updates, ", ")
	argCount++
	query += fmt.Sprintf(" WHERE uuid = $%d", argCount)
	args = append(args, id)

	query += " RETURNING uuid, customer_uuid, title, description, weight_kg, length_cm, width_cm, height_cm, from_location, to_location, tags, price, available_from, created_at"

	var order models.Order
	var description, fromLocation, toLocation sql.NullString
	var lengthCm, widthCm, heightCm sql.NullFloat64
	var availableFrom sql.NullTime

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&order.UUID, &order.CustomerUUID, &order.Title, &description, &order.WeightKg,
		&lengthCm, &widthCm, &heightCm, &fromLocation, &toLocation,
		pq.Array(&order.Tags), &order.Price, &availableFrom, &order.CreatedAt,
	)
	if err != nil {
		return models.Order{}, fmt.Errorf("failed to update order: %w", err)
	}

	if description.Valid {
		order.Description = &description.String
	}
	if fromLocation.Valid {
		order.FromLocation = &fromLocation.String
	}
	if toLocation.Valid {
		order.ToLocation = &toLocation.String
	}
	if lengthCm.Valid {
		order.LengthCm = &lengthCm.Float64
	}
	if widthCm.Valid {
		order.WidthCm = &widthCm.Float64
	}
	if heightCm.Valid {
		order.HeightCm = &heightCm.Float64
	}
	if availableFrom.Valid {
		order.AvailableFrom = &availableFrom.Time
	}

	return order, nil
}

// Customers methods
func (db *DB) ListCustomers(ctx context.Context, filter models.CustomerFilter) ([]models.Customer, int, error) {
	query := "SELECT uuid, name, phone, telegram_id, telegram_tag, created_at FROM customers WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if filter.Name != "" {
		argCount++
		query += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		args = append(args, "%"+filter.Name+"%")
	}
	if filter.Phone != "" {
		argCount++
		query += fmt.Sprintf(" AND phone ILIKE $%d", argCount)
		args = append(args, "%"+filter.Phone+"%")
	}
	if filter.TelegramTag != "" {
		argCount++
		query += fmt.Sprintf(" AND telegram_tag ILIKE $%d", argCount)
		args = append(args, "%"+filter.TelegramTag+"%")
	}
	if filter.TelegramID != 0 {
		argCount++
		query += fmt.Sprintf(" AND telegram_id = $%d", argCount)
		args = append(args, filter.TelegramID)
	}

	if filter.SortBy != "" {
		query += " ORDER BY " + filter.SortBy
		if filter.SortOrder == "desc" {
			query += " DESC"
		} else {
			query += " ASC"
		}
	} else {
		query += " ORDER BY created_at DESC"
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit
	argCount++
	query += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, filter.Limit)
	argCount++
	query += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		var telegramID sql.NullInt64
		var telegramTag sql.NullString

		err := rows.Scan(
			&customer.UUID, &customer.Name, &customer.Phone, &telegramID, &telegramTag, &customer.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}

		if telegramID.Valid {
			customer.TelegramID = &telegramID.Int64
		}
		if telegramTag.Valid {
			customer.TelegramTag = &telegramTag.String
		}

		customers = append(customers, customer)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM customers WHERE 1=1"
	countArgs := []interface{}{}
	argCount = 0

	if filter.Name != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND name ILIKE $%d", argCount)
		countArgs = append(countArgs, "%"+filter.Name+"%")
	}
	if filter.Phone != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND phone ILIKE $%d", argCount)
		countArgs = append(countArgs, "%"+filter.Phone+"%")
	}
	if filter.TelegramTag != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND telegram_tag ILIKE $%d", argCount)
		countArgs = append(countArgs, "%"+filter.TelegramTag+"%")
	}
	if filter.TelegramID != 0 {
		argCount++
		countQuery += fmt.Sprintf(" AND telegram_id = $%d", argCount)
		countArgs = append(countArgs, filter.TelegramID)
	}

	var total int
	err = db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return customers, total, nil
}

func (db *DB) CreateCustomer(ctx context.Context, input models.CreateCustomerInput) (models.Customer, error) {
	query := `
		INSERT INTO customers (name, phone, telegram_id, telegram_tag)
		VALUES ($1, $2, $3, $4)
		RETURNING uuid, name, phone, telegram_id, telegram_tag, created_at
	`

	var customer models.Customer
	var telegramID sql.NullInt64
	var telegramTag sql.NullString

	err := db.QueryRowContext(ctx, query,
		input.Name, input.Phone, input.TelegramID, input.TelegramTag,
	).Scan(
		&customer.UUID, &customer.Name, &customer.Phone, &telegramID, &telegramTag, &customer.CreatedAt,
	)
	if err != nil {
		return models.Customer{}, fmt.Errorf("failed to create customer: %w", err)
	}

	if telegramID.Valid {
		customer.TelegramID = &telegramID.Int64
	}
	if telegramTag.Valid {
		customer.TelegramTag = &telegramTag.String
	}

	return customer, nil
}

func (db *DB) UpdateCustomer(ctx context.Context, id uuid.UUID, input models.UpdateCustomerInput) (models.Customer, error) {
	query := "UPDATE customers SET "
	args := []interface{}{}
	argCount := 0

	updates := []string{}

	if input.Name != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("name = $%d", argCount))
		args = append(args, *input.Name)
	}
	if input.Phone != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("phone = $%d", argCount))
		args = append(args, *input.Phone)
	}
	if input.TelegramID != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("telegram_id = $%d", argCount))
		args = append(args, *input.TelegramID)
	}
	if input.TelegramTag != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("telegram_tag = $%d", argCount))
		args = append(args, *input.TelegramTag)
	}

	if len(updates) == 0 {
		return models.Customer{}, fmt.Errorf("no fields to update")
	}

	query += strings.Join(updates, ", ")
	argCount++
	query += fmt.Sprintf(" WHERE uuid = $%d", argCount)
	args = append(args, id)

	query += " RETURNING uuid, name, phone, telegram_id, telegram_tag, created_at"

	var customer models.Customer
	var telegramID sql.NullInt64
	var telegramTag sql.NullString

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&customer.UUID, &customer.Name, &customer.Phone, &telegramID, &telegramTag, &customer.CreatedAt,
	)
	if err != nil {
		return models.Customer{}, fmt.Errorf("failed to update customer: %w", err)
	}

	if telegramID.Valid {
		customer.TelegramID = &telegramID.Int64
	}
	if telegramTag.Valid {
		customer.TelegramTag = &telegramTag.String
	}

	return customer, nil
}

func (db *DB) GetCustomerByTelegramID(ctx context.Context, telegramID int64) (*models.Customer, error) {
	query := `
		SELECT uuid, name, phone, telegram_id, telegram_tag, created_at
		FROM customers WHERE telegram_id = $1
	`

	var customer models.Customer
	var telegramIDNull sql.NullInt64
	var telegramTag sql.NullString

	err := db.QueryRowContext(ctx, query, telegramID).Scan(
		&customer.UUID, &customer.Name, &customer.Phone, &telegramIDNull, &telegramTag, &customer.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by telegram ID: %w", err)
	}

	if telegramIDNull.Valid {
		customer.TelegramID = &telegramIDNull.Int64
	}
	if telegramTag.Valid {
		customer.TelegramTag = &telegramTag.String
	}

	return &customer, nil
}
