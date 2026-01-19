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
	{
		Version: "007",
		Name:    "extend_orders_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				ALTER TABLE orders
					ADD COLUMN IF NOT EXISTS discount_currency VARCHAR(3),
					ADD COLUMN IF NOT EXISTS tax_currency VARCHAR(3),
					ADD COLUMN IF NOT EXISTS shipping_currency VARCHAR(3),
					ADD COLUMN IF NOT EXISTS total_currency VARCHAR(3),
					ADD COLUMN IF NOT EXISTS shipping_address JSONB,
					ADD COLUMN IF NOT EXISTS billing_address JSONB,
					ADD COLUMN IF NOT EXISTS payment_method_id VARCHAR(255),
					ADD COLUMN IF NOT EXISTS notes TEXT,
					ADD COLUMN IF NOT EXISTS ip_address VARCHAR(100),
					ADD COLUMN IF NOT EXISTS user_agent TEXT,
					ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP,
					ADD COLUMN IF NOT EXISTS canceled_at TIMESTAMP;
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			// Best-effort rollback (columns may contain data). This intentionally does not drop columns.
			return nil
		},
	},
	{
		Version: "008",
		Name:    "create_order_items_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS order_items (
					id VARCHAR(255) PRIMARY KEY,
					order_id VARCHAR(255) NOT NULL,
					product_id VARCHAR(255) NOT NULL,
					variant_id VARCHAR(255),
					sku VARCHAR(255) NOT NULL,
					name VARCHAR(255) NOT NULL,
					unit_price_amount BIGINT NOT NULL,
					unit_price_currency VARCHAR(3) NOT NULL,
					quantity INT NOT NULL,
					discount_amount BIGINT NOT NULL,
					discount_currency VARCHAR(3) NOT NULL,
					tax_amount BIGINT NOT NULL,
					tax_currency VARCHAR(3) NOT NULL,
					total_amount BIGINT NOT NULL,
					total_currency VARCHAR(3) NOT NULL,
					attributes JSONB,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
				);
				CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS order_items CASCADE")
		},
	},
	{
		Version: "009",
		Name:    "create_promotions_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS promotions (
					id VARCHAR(255) PRIMARY KEY,
					code VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					description TEXT,
					discount_type VARCHAR(50) NOT NULL,
					value DOUBLE PRECISION NOT NULL,
					min_purchase_amount BIGINT,
					min_purchase_currency VARCHAR(3),
					max_discount_amount BIGINT,
					max_discount_currency VARCHAR(3),
					valid_from TIMESTAMP,
					valid_to TIMESTAMP,
					is_active BOOLEAN NOT NULL DEFAULT true,
					usage_limit INT NOT NULL DEFAULT 0,
					usage_count INT NOT NULL DEFAULT 0,
					applicable_product_ids JSONB,
					applicable_category_ids JSONB,
					excluded_product_ids JSONB,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_promotions_code ON promotions(code);
				CREATE INDEX IF NOT EXISTS idx_promotions_is_active ON promotions(is_active);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS promotions CASCADE")
		},
	},
	{
		Version: "010",
		Name:    "create_variants_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS variants (
					id VARCHAR(255) PRIMARY KEY,
					product_id VARCHAR(255) NOT NULL,
					sku VARCHAR(255) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					price_amount BIGINT NOT NULL,
					price_currency VARCHAR(3) NOT NULL,
					attributes JSONB,
					images JSONB,
					is_available BOOLEAN NOT NULL DEFAULT true,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
				);
				CREATE INDEX IF NOT EXISTS idx_variants_product_id ON variants(product_id);
				CREATE INDEX IF NOT EXISTS idx_variants_sku ON variants(sku);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS variants CASCADE")
		},
	},
	{
		Version: "011",
		Name:    "extend_cart_items_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				ALTER TABLE cart_items
					ADD COLUMN IF NOT EXISTS attributes JSONB;
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			// Best-effort rollback; keep column.
			return nil
		},
	},
	{
		Version: "012",
		Name:    "create_product_prices_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS product_prices (
					id VARCHAR(255) PRIMARY KEY,
					product_id VARCHAR(255) NOT NULL,
					variant_id VARCHAR(255),
					price_amount BIGINT NOT NULL,
					price_currency VARCHAR(3) NOT NULL,
					valid_from TIMESTAMP,
					valid_to TIMESTAMP,
					priority INT NOT NULL DEFAULT 0,
					price_type VARCHAR(50) NOT NULL DEFAULT 'regular',
					is_active BOOLEAN NOT NULL DEFAULT true,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
					FOREIGN KEY (variant_id) REFERENCES variants(id) ON DELETE CASCADE
				);
				CREATE INDEX IF NOT EXISTS idx_product_prices_product_id ON product_prices(product_id);
				CREATE INDEX IF NOT EXISTS idx_product_prices_variant_id ON product_prices(variant_id);
				CREATE INDEX IF NOT EXISTS idx_product_prices_valid_from ON product_prices(valid_from);
				CREATE INDEX IF NOT EXISTS idx_product_prices_valid_to ON product_prices(valid_to);
				CREATE INDEX IF NOT EXISTS idx_product_prices_is_active ON product_prices(is_active);
				CREATE INDEX IF NOT EXISTS idx_product_prices_priority ON product_prices(priority);

				-- Composite index for efficient price lookup
				CREATE INDEX IF NOT EXISTS idx_product_prices_lookup
					ON product_prices(product_id, variant_id, is_active, valid_from, valid_to, priority DESC);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS product_prices CASCADE")
		},
	},
	// ============================================================
	// Inventory Workflow Migrations (v013-v017)
	// ============================================================
	{
		Version: "013",
		Name:    "create_suppliers_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS suppliers (
					id VARCHAR(255) PRIMARY KEY,
					code VARCHAR(50) UNIQUE NOT NULL,
					name VARCHAR(255) NOT NULL,
					contact_name VARCHAR(255),
					contact_email VARCHAR(255),
					contact_phone VARCHAR(50),
					address TEXT,
					website VARCHAR(500),
					notes TEXT,
					is_active BOOLEAN NOT NULL DEFAULT true,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
				CREATE INDEX IF NOT EXISTS idx_suppliers_code ON suppliers(code);
				CREATE INDEX IF NOT EXISTS idx_suppliers_name ON suppliers(name);
				CREATE INDEX IF NOT EXISTS idx_suppliers_is_active ON suppliers(is_active);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS suppliers CASCADE")
		},
	},
	{
		Version: "014",
		Name:    "create_product_suppliers_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS product_suppliers (
					id VARCHAR(255) PRIMARY KEY,
					product_id VARCHAR(255) NOT NULL,
					supplier_id VARCHAR(255) NOT NULL,
					supplier_sku VARCHAR(100),
					cost_price_amount BIGINT,
					cost_price_currency VARCHAR(3),
					lead_time_days INT NOT NULL DEFAULT 0,
					min_order_qty INT NOT NULL DEFAULT 1,
					is_primary BOOLEAN NOT NULL DEFAULT false,
					is_active BOOLEAN NOT NULL DEFAULT true,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
					FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE CASCADE,
					UNIQUE(product_id, supplier_id)
				);
				CREATE INDEX IF NOT EXISTS idx_product_suppliers_product_id ON product_suppliers(product_id);
				CREATE INDEX IF NOT EXISTS idx_product_suppliers_supplier_id ON product_suppliers(supplier_id);
				CREATE INDEX IF NOT EXISTS idx_product_suppliers_is_primary ON product_suppliers(is_primary);
				CREATE INDEX IF NOT EXISTS idx_product_suppliers_is_active ON product_suppliers(is_active);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS product_suppliers CASCADE")
		},
	},
	{
		Version: "015",
		Name:    "create_inventory_levels_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS inventory_levels (
					id VARCHAR(255) PRIMARY KEY,
					sku VARCHAR(100) UNIQUE NOT NULL,
					product_id VARCHAR(255),
					variant_id VARCHAR(255),
					quantity_on_hand INT NOT NULL DEFAULT 0,
					quantity_reserved INT NOT NULL DEFAULT 0,
					quantity_available INT GENERATED ALWAYS AS (quantity_on_hand - quantity_reserved) STORED,
					reorder_point INT NOT NULL DEFAULT 10,
					reorder_quantity INT NOT NULL DEFAULT 50,
					location VARCHAR(100),
					bin_location VARCHAR(50),
					is_active BOOLEAN NOT NULL DEFAULT true,
					last_counted_at TIMESTAMP,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL,
					FOREIGN KEY (variant_id) REFERENCES variants(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_sku ON inventory_levels(sku);
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_product_id ON inventory_levels(product_id);
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_variant_id ON inventory_levels(variant_id);
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_location ON inventory_levels(location);
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_is_active ON inventory_levels(is_active);

				-- Partial index for low stock alerts (items at or below reorder point)
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_low_stock
					ON inventory_levels(sku, quantity_on_hand, reorder_point)
					WHERE is_active = true;

				-- Composite index for stock availability checks
				CREATE INDEX IF NOT EXISTS idx_inventory_levels_lookup
					ON inventory_levels(sku, is_active, quantity_on_hand, quantity_reserved);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS inventory_levels CASCADE")
		},
	},
	{
		Version: "016",
		Name:    "create_inventory_suppliers_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				CREATE TABLE IF NOT EXISTS inventory_suppliers (
					id VARCHAR(255) PRIMARY KEY,
					inventory_level_id VARCHAR(255) NOT NULL,
					supplier_id VARCHAR(255) NOT NULL,
					quantity_from_supplier INT NOT NULL DEFAULT 0,
					last_restock_at TIMESTAMP,
					next_restock_at TIMESTAMP,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (inventory_level_id) REFERENCES inventory_levels(id) ON DELETE CASCADE,
					FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE CASCADE,
					UNIQUE(inventory_level_id, supplier_id)
				);
				CREATE INDEX IF NOT EXISTS idx_inventory_suppliers_inventory_id ON inventory_suppliers(inventory_level_id);
				CREATE INDEX IF NOT EXISTS idx_inventory_suppliers_supplier_id ON inventory_suppliers(supplier_id);
				CREATE INDEX IF NOT EXISTS idx_inventory_suppliers_next_restock ON inventory_suppliers(next_restock_at);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, "DROP TABLE IF EXISTS inventory_suppliers CASCADE")
		},
	},
	{
		Version: "017",
		Name:    "create_inventory_activities_table",
		Up: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				-- Create enum type for activity types
				DO $$ BEGIN
					CREATE TYPE inventory_activity_type AS ENUM (
						'stock_in',
						'stock_out',
						'adjustment',
						'reservation',
						'reservation_release',
						'reservation_commit',
						'transfer_in',
						'transfer_out',
						'cycle_count',
						'damage',
						'return'
					);
				EXCEPTION
					WHEN duplicate_object THEN null;
				END $$;

				CREATE TABLE IF NOT EXISTS inventory_activities (
					id VARCHAR(255) PRIMARY KEY,
					sku VARCHAR(100) NOT NULL,
					activity_type inventory_activity_type NOT NULL,
					quantity INT NOT NULL,
					quantity_before INT NOT NULL,
					quantity_after INT NOT NULL,
					reference_type VARCHAR(50),
					reference_id VARCHAR(255),
					supplier_id VARCHAR(255),
					location VARCHAR(100),
					reason TEXT,
					performed_by VARCHAR(255),
					metadata JSONB,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE SET NULL
				);
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_sku ON inventory_activities(sku);
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_type ON inventory_activities(activity_type);
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_created_at ON inventory_activities(created_at);
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_reference ON inventory_activities(reference_type, reference_id);
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_supplier ON inventory_activities(supplier_id);
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_location ON inventory_activities(location);

				-- Composite index for audit queries (SKU + time range)
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_audit
					ON inventory_activities(sku, created_at DESC);

				-- Composite index for reference lookups
				CREATE INDEX IF NOT EXISTS idx_inventory_activities_ref_lookup
					ON inventory_activities(reference_type, reference_id, created_at DESC);
			`)
		},
		Down: func(ctx context.Context, exec Executor) error {
			return exec.Exec(ctx, `
				DROP TABLE IF EXISTS inventory_activities CASCADE;
				DROP TYPE IF EXISTS inventory_activity_type;
			`)
		},
	},
}
