# PostgreSQL Migration Example

Complete example using PostgreSQL with Docker Compose.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21 or later

## Quick Start

### 1. Start PostgreSQL

From the project root:

```bash
cd migrations/examples && docker-compose up -d
```

Wait a few seconds for PostgreSQL to be ready.

### 2. Install Dependencies

```bash
cd migrations/examples/postgresql
go mod download
```

### 3. Run Migrations

```bash
go run main.go
```

## Expected Output

### First Run

```
🔄 GoCommerce PostgreSQL Migration Tool
========================================

📡 Connecting to PostgreSQL...
✓ Connected to PostgreSQL

📊 Current Migration Status:
   Applied: 0 migrations
   Pending: 6 migrations
   ○ 001 - create_brands_table
   ○ 002 - create_categories_table
   ○ 003 - create_products_table
   ○ 004 - create_carts_table
   ○ 005 - create_cart_items_table
   ○ 006 - create_orders_table

⬆️  Running pending migrations...
✅ Migrations completed!
   Total applied: 6 migrations

   ✓ 001 - create_brands_table
   ✓ 002 - create_categories_table
   ✓ 003 - create_products_table
   ✓ 004 - create_carts_table
   ✓ 005 - create_cart_items_table
   ✓ 006 - create_orders_table

📋 Created Tables:
   - brands
   - categories
   - products (with brand_id, category_id, images, attributes)
   - carts
   - cart_items
   - orders

💡 Verify with:
   docker-compose exec postgres psql -U edomain -d edomain -c '\dt'
```

### Subsequent Runs

```
🔄 GoCommerce PostgreSQL Migration Tool
========================================

📡 Connecting to PostgreSQL...
✓ Connected to PostgreSQL

📊 Current Migration Status:
   Applied: 6 migrations
   ✓ 001 - create_brands_table (applied: 2025-11-28 18:14:40)
   ✓ 002 - create_categories_table (applied: 2025-11-28 18:14:40)
   ✓ 003 - create_products_table (applied: 2025-11-28 18:14:40)
   ✓ 004 - create_carts_table (applied: 2025-11-28 18:14:40)
   ✓ 005 - create_cart_items_table (applied: 2025-11-28 18:14:40)
   ✓ 006 - create_orders_table (applied: 2025-11-28 18:14:40)
   Pending: 0 migrations

✅ Database is up to date!
```

## Connection Details

The example uses these connection settings (from migrations/examples/docker-compose.yml):

- **Host**: localhost
- **Port**: 5432
- **Database**: edomain
- **Username**: edomain
- **Password**: edomain

Connection string:
```
host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable
```

## Verify Migrations

### View Tables

```bash
docker-compose exec postgres psql -U edomain -d edomain -c '\dt'
```

### View Migration History

```bash
docker-compose exec postgres psql -U edomain -d edomain -c 'SELECT * FROM gocommerce_migrations;'
```

### View Table Schema

```bash
docker-compose exec postgres psql -U edomain -d edomain -c '\d products'
```

### Query Data

```bash
docker-compose exec postgres psql -U edomain -d edomain
```

Then run SQL:
```sql
SELECT * FROM products;
SELECT * FROM carts;
```

## Troubleshooting

### Connection Refused

Make sure PostgreSQL is running:
```bash
docker-compose ps
```

Should show:
```
NAME                  STATUS
gocommerce-postgres   Up X minutes (healthy)
```

If not running:
```bash
docker-compose up -d
docker-compose logs postgres
```

### Port Already in Use

If port 5432 is in use, edit `migrations/examples/docker-compose.yml`:
```yaml
ports:
  - "5432:5432"
```

Then update connection string in `main.go`:
```go
connStr := "host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable"
```

### Permission Denied

Check PostgreSQL logs:
```bash
docker-compose logs postgres
```

### Reset Database

Stop and remove volumes:
```bash
cd ../../..  # Back to project root
docker-compose down -v
docker-compose up -d
cd migrations/examples/postgresql
go run main.go
```

## PostgreSQL-Specific Features

This example uses PostgreSQL-specific syntax:

### Parameterized Queries
```go
// PostgreSQL uses $1, $2, etc.
exec.Exec(ctx, "INSERT INTO products (id, name) VALUES ($1, $2)", id, name)
```

### Indexes
```sql
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
```

### Foreign Keys with CASCADE
```sql
FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE
```

### Timestamps
```sql
created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
```

## Seeding Data

After running migrations, populate the database with mock data:

```bash
go run seed-products.go
```

This will seed:
- **8 Brands**: Apple, Dell, Lenovo, HP, Samsung, Logitech, Sony, Bose
- **8 Categories**: Electronics, Computers, Accessories, Audio, Storage, Networking, Office, Furniture
- **72 Products**: 22 curated products + 50 random products for testing

### Expected Seeder Output

```
🌱 GoCommerce Database Seeder
========================================

📡 Connecting to PostgreSQL...
✓ Connected to PostgreSQL

📋 Available Seeds:
   1. brand_seeder
      Seeds the brands table with product brands
   2. category_seeder
      Seeds the categories table with product categories
   3. product_seeder
      Seeds the products table with realistic mock product data
   4. random_product_seeder
      Seeds the products table with randomly generated products for load testing

📊 Products before seeding: 0

🌱 Running seeds...

✅ Seeding completed!
   Products after seeding: 72
   New products added: 72
```

### View Seeded Data

```bash
# View brands
docker-compose exec postgres psql -U edomain -d edomain -c "SELECT * FROM brands;"

# View categories with hierarchy
docker-compose exec postgres psql -U edomain -d edomain -c "SELECT id, name, parent_id FROM categories ORDER BY display_order;"

# View products with brand and category
docker-compose exec postgres psql -U edomain -d edomain -c "SELECT p.name, b.name as brand, c.name as category, p.base_price_amount/100.0 as price FROM products p LEFT JOIN brands b ON p.brand_id = b.id LEFT JOIN categories c ON p.category_id = c.id LIMIT 10;"
```

## Integration with Sample Project

To use PostgreSQL with the sample-project:

1. Update `sample-project/main.go`:

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    // Connect to PostgreSQL
    db, err := sql.Open("postgres", 
        "host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Run migrations
    executor := NewPostgreSQLExecutor(db)
    repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
    manager := migrations.NewManager(repo, executor)
    manager.RegisterMultiple(migrations.PostgreSQLExampleMigrations)
    manager.Up(context.Background())
    
    // Replace MemoryStore with PostgreSQL repositories
    productRepo := postgres.NewProductRepository(db)
    cartRepo := postgres.NewCartRepository(db)
    // ...
}
```

2. Implement repository interfaces for PostgreSQL
3. Start the API server

## Project Structure

```
migrations/examples/postgresql/
├── main.go              # Migration runner
├── seed-products.go     # Database seeder
├── go.mod              # Go module file
├── go.sum              # Dependency checksums
└── README.md           # This file
```

## Schema Overview

The migrations create the following schema:

### Core Tables
- **brands**: Product brands (Apple, Dell, Samsung, etc.)
- **categories**: Hierarchical product categories
- **products**: Products with brand/category relationships, images, and attributes

### E-commerce Tables
- **carts**: Shopping carts for users/sessions
- **cart_items**: Items in shopping carts
- **orders**: Customer orders

### Product Schema

Matches the `catalog.Product` domain model:
```go
type Product struct {
    ID          string
    SKU         string
    Name        string
    Description string
    BrandID     string          // Links to brands table
    CategoryID  string          // Links to categories table
    BasePrice   money.Money     // Stored as amount (cents) + currency
    Status      ProductStatus   // active, draft, discontinued
    Images      []string        // Stored as JSON array
    Attributes  map[string]string // Stored as JSON object
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

## Next Steps

- ✅ Migrations with brands and categories
- ✅ Database seeder with realistic mock data
- ✅ Complete schema matching domain models
- 🔲 Implement repository interfaces for PostgreSQL
- 🔲 Integrate with the sample-project API
- 🔲 Add more entity migrations (variants, inventory, shipping)
- 🔲 Set up connection pooling
- 🔲 Add monitoring and logging
