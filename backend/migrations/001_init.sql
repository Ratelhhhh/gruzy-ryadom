-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create customers table
CREATE TABLE customers (
  uuid           UUID      PRIMARY KEY DEFAULT uuid_generate_v4(),
  name           TEXT      NOT NULL,    -- ФИО или название
  phone          TEXT      NOT NULL,
  telegram_id    BIGINT,                 -- числовой ID в Telegram
  telegram_tag   TEXT,                   -- @username
  created_at     TIMESTAMP NOT NULL DEFAULT now()
);

-- Create orders table
CREATE TABLE orders (
  uuid           UUID      PRIMARY KEY DEFAULT uuid_generate_v4(),
  customer_uuid  UUID      NOT NULL REFERENCES customers(uuid) ON DELETE CASCADE,
  title          TEXT      NOT NULL,
  description    TEXT,
  weight_kg      NUMERIC   NOT NULL CHECK(weight_kg >= 0),
  length_cm      NUMERIC             CHECK(length_cm >= 0),
  width_cm       NUMERIC             CHECK(width_cm >= 0),
  height_cm      NUMERIC             CHECK(height_cm >= 0),
  from_location  TEXT,
  to_location    TEXT,
  tags           TEXT[]    NOT NULL DEFAULT '{}',
  price          NUMERIC   NOT NULL CHECK(price >= 0),
  available_from DATE,
  created_at     TIMESTAMP NOT NULL DEFAULT now()
);

-- Create indexes
CREATE INDEX idx_orders_tags   ON orders USING GIN(tags);
CREATE INDEX idx_orders_price  ON orders(price);
CREATE INDEX idx_orders_weight ON orders(weight_kg);
CREATE INDEX idx_orders_customer ON orders(customer_uuid);
CREATE INDEX idx_customers_telegram_id ON customers(telegram_id);
