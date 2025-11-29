package migrations

import (
	"context"
)

// ExampleMigrations demonstrates how to define migrations for gocommerce.
var ExampleMigrations = []Migration{
	{
		Version: "001",
		Name:    "create_products_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE products (
					id VARCHAR(255) PRIMARY KEY,
					sku VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					description TEXT,
					base_price_amount BIGINT NOT NULL,
					base_price_currency VARCHAR(3) NOT NULL,
					status VARCHAR(50) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_sku (sku),
					INDEX idx_status (status)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS products")
		},
	},
	{
		Version: "002",
		Name:    "create_carts_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE carts (
					id VARCHAR(255) PRIMARY KEY,
					user_id VARCHAR(255),
					session_id VARCHAR(255),
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP,
					INDEX idx_user_id (user_id),
					INDEX idx_session_id (session_id)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS carts")
		},
	},
	{
		Version: "003",
		Name:    "create_cart_items_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE cart_items (
					id VARCHAR(255) PRIMARY KEY,
					cart_id VARCHAR(255) NOT NULL,
					product_id VARCHAR(255) NOT NULL,
					variant_id VARCHAR(255),
					sku VARCHAR(255) NOT NULL,
					name VARCHAR(255) NOT NULL,
					price_amount BIGINT NOT NULL,
					price_currency VARCHAR(3) NOT NULL,
					quantity INT NOT NULL,
					added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
					INDEX idx_cart_id (cart_id)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS cart_items")
		},
	},
	{
		Version: "004",
		Name:    "create_orders_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE orders (
					id VARCHAR(255) PRIMARY KEY,
					order_number VARCHAR(255) UNIQUE NOT NULL,
					user_id VARCHAR(255) NOT NULL,
					status VARCHAR(50) NOT NULL,
					subtotal_amount BIGINT NOT NULL,
					subtotal_currency VARCHAR(3) NOT NULL,
					discount_amount BIGINT NOT NULL,
					tax_amount BIGINT NOT NULL,
					shipping_amount BIGINT NOT NULL,
					total_amount BIGINT NOT NULL,
					payment_status VARCHAR(50),
					fulfillment_status VARCHAR(50),
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_order_number (order_number),
					INDEX idx_user_id (user_id),
					INDEX idx_status (status),
					INDEX idx_created_at (created_at)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS orders")
		},
	},
	{
		Version: "005",
		Name:    "create_order_items_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE order_items (
					id VARCHAR(255) PRIMARY KEY,
					order_id VARCHAR(255) NOT NULL,
					product_id VARCHAR(255) NOT NULL,
					variant_id VARCHAR(255),
					sku VARCHAR(255) NOT NULL,
					name VARCHAR(255) NOT NULL,
					price_amount BIGINT NOT NULL,
					price_currency VARCHAR(3) NOT NULL,
					quantity INT NOT NULL,
					subtotal_amount BIGINT NOT NULL,
					discount_amount BIGINT NOT NULL,
					tax_amount BIGINT NOT NULL,
					total_amount BIGINT NOT NULL,
					FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
					INDEX idx_order_id (order_id)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS order_items")
		},
	},
	{
		Version: "006",
		Name:    "create_promotions_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE promotions (
					id VARCHAR(255) PRIMARY KEY,
					code VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					description TEXT,
					discount_type VARCHAR(50) NOT NULL,
					discount_value BIGINT NOT NULL,
					min_purchase_amount BIGINT,
					max_discount_amount BIGINT,
					is_active BOOLEAN NOT NULL DEFAULT TRUE,
					starts_at TIMESTAMP,
					ends_at TIMESTAMP,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_code (code),
					INDEX idx_is_active (is_active)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS promotions")
		},
	},
	{
		Version: "007",
		Name:    "create_addresses_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE addresses (
					id VARCHAR(255) PRIMARY KEY,
					user_id VARCHAR(255) NOT NULL,
					first_name VARCHAR(255) NOT NULL,
					last_name VARCHAR(255) NOT NULL,
					company VARCHAR(255),
					address_line_1 VARCHAR(255) NOT NULL,
					address_line_2 VARCHAR(255),
					city VARCHAR(255) NOT NULL,
					state VARCHAR(255),
					postal_code VARCHAR(50) NOT NULL,
					country VARCHAR(2) NOT NULL,
					phone VARCHAR(50),
					is_default BOOLEAN NOT NULL DEFAULT FALSE,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					INDEX idx_user_id (user_id)
				)
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS addresses")
		},
	},
}

// PostgreSQLExampleMigrations are PostgreSQL-specific migrations.
var PostgreSQLExampleMigrations = []Migration{
	{
		Version: "001",
		Name:    "create_brands_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS brands (
					id VARCHAR(255) PRIMARY KEY,
					name VARCHAR(255) NOT NULL,
					slug VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					logo_url VARCHAR(500),
					is_active BOOLEAN NOT NULL DEFAULT true,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_brands_slug ON brands(slug);
				CREATE INDEX IF NOT EXISTS idx_brands_is_active ON brands(is_active);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS brands CASCADE")
		},
	},
	{
		Version: "002",
		Name:    "create_categories_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS categories (
					id VARCHAR(255) PRIMARY KEY,
					parent_id VARCHAR(255),
					name VARCHAR(255) NOT NULL,
					slug VARCHAR(255) UNIQUE NOT NULL,
					description TEXT,
					image_url VARCHAR(500),
					is_active BOOLEAN NOT NULL DEFAULT true,
					display_order INT NOT NULL DEFAULT 0,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug);
				CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
				CREATE INDEX IF NOT EXISTS idx_categories_is_active ON categories(is_active);
				CREATE INDEX IF NOT EXISTS idx_categories_display_order ON categories(display_order);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS categories CASCADE")
		},
	},
	{
		Version: "003",
		Name:    "create_products_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS products (
					id VARCHAR(255) PRIMARY KEY,
					sku VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					description TEXT,
					brand_id VARCHAR(255),
					category_id VARCHAR(255),
					base_price_amount BIGINT NOT NULL,
					base_price_currency VARCHAR(3) NOT NULL,
					status VARCHAR(50) NOT NULL,
					images TEXT,
					attributes TEXT,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
				CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
				CREATE INDEX IF NOT EXISTS idx_products_brand_id ON products(brand_id);
				CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS products CASCADE")
		},
	},
	{
		Version: "004",
		Name:    "create_carts_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS carts (
					id VARCHAR(255) PRIMARY KEY,
					user_id VARCHAR(255),
					session_id VARCHAR(255),
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_carts_user_id ON carts(user_id);
				CREATE INDEX IF NOT EXISTS idx_carts_session_id ON carts(session_id);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS carts CASCADE")
		},
	},
	{
		Version: "005",
		Name:    "create_cart_items_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS cart_items (
					id VARCHAR(255) PRIMARY KEY,
					cart_id VARCHAR(255) NOT NULL,
					product_id VARCHAR(255) NOT NULL,
					variant_id VARCHAR(255),
					sku VARCHAR(255) NOT NULL,
					name VARCHAR(255) NOT NULL,
					price_amount BIGINT NOT NULL,
					price_currency VARCHAR(3) NOT NULL,
					quantity INT NOT NULL,
					added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE
				);
				CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items(cart_id);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS cart_items CASCADE")
		},
	},
	{
		Version: "006",
		Name:    "create_orders_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS orders (
					id VARCHAR(255) PRIMARY KEY,
					order_number VARCHAR(255) UNIQUE NOT NULL,
					user_id VARCHAR(255) NOT NULL,
					status VARCHAR(50) NOT NULL,
					subtotal_amount BIGINT NOT NULL,
					subtotal_currency VARCHAR(3) NOT NULL,
					discount_amount BIGINT NOT NULL,
					tax_amount BIGINT NOT NULL,
					shipping_amount BIGINT NOT NULL,
					total_amount BIGINT NOT NULL,
					payment_status VARCHAR(50),
					fulfillment_status VARCHAR(50),
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
				CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
				CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
				CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS orders CASCADE")
		},
	},
}
