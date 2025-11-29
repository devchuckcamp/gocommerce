# Migration Examples

This directory contains practical examples of using the gocommerce migration system with different databases.

## 📚 Guides

- **[INTEGRATION.md](INTEGRATION.md)** - Complete guide for integrating migrations into your own project
- **[DOCKER.md](DOCKER.md)** - PostgreSQL setup with Docker Compose
- **[postgresql/README.md](postgresql/README.md)** - Detailed PostgreSQL example walkthrough

## PostgreSQL Example

A complete, runnable example using PostgreSQL:

```bash
cd examples/postgresql
go run main.go
```

This will:
1. Connect to the PostgreSQL database
2. Check migration status
3. Run any pending migrations
4. Display results

### First Run Output

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

## Database Seeding

After running migrations, seed the database with test data:

```bash
go run seed-products.go
```

### Seeding Output

```
🌱 GoCommerce Database Seeder
===============================

📡 Connecting to PostgreSQL...
✓ Connected to PostgreSQL

🌱 Running all seeds...

✓ brand_seeder completed (8 brands)
✓ category_seeder completed (8 categories)
✓ product_seeder completed (22 products)
✓ random_product_seeder completed (50 products)

✅ All seeds completed successfully!

📊 Database Summary:
   Total Products: 72
   Active Products: 57
   
   Sample Products:
   - MacBook Pro 16" M3 Max ($2,799.00)
   - Dell XPS 15 9530 ($1,899.99)
   - iPhone 15 Pro Max ($1,199.00)
   - iPad Pro 12.9" M2 ($1,099.00)
   - Logitech MX Master 3S ($99.99)
```

### View Seeded Data

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U edomain -d edomain

# View brands
SELECT * FROM brands;

# View categories
SELECT * FROM categories ORDER BY parent_id NULLS FIRST, display_order;

# View products with brands and categories
SELECT 
    p.name,
    p.base_price_amount / 100.0 as price,
    b.name as brand,
    c.name as category,
    p.status
FROM products p
JOIN brands b ON p.brand_id = b.id
JOIN categories c ON p.category_id = c.id
ORDER BY p.name
LIMIT 10;
```

### Seeded Data Includes

- **8 Brands**: Apple, Dell, Lenovo, HP, Samsung, Logitech, Sony, Bose
- **8 Categories**: Electronics, Computers, Accessories, etc. (with hierarchy)
- **72 Products**: 
  - 22 curated products (MacBook, Dell XPS, iPhone, iPad, etc.)
  - 50 random products for load testing

## Key Concepts Demonstrated

### 1. Executor Implementation

The `PostgreSQLExecutor` implements the `migrations.Executor` interface:

```go
type PostgreSQLExecutor struct {
    db *sql.DB
    tx *sql.Tx
}

func (e *PostgreSQLExecutor) Exec(ctx context.Context, query string, args ...interface{}) error
func (e *PostgreSQLExecutor) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
func (e *PostgreSQLExecutor) Begin(ctx context.Context) (migrations.Executor, error)
func (e *PostgreSQLExecutor) Commit(ctx context.Context) error
func (e *PostgreSQLExecutor) Rollback(ctx context.Context) error
```

### 2. Migration Definition

Migrations are defined with version, name, and Up/Down functions:

```go
migrations.Migration{
    Version: "001",
    Name:    "create_products_table",
    Up: func(ctx context.Context, exec migrations.Executor) error {
        return exec.Exec(ctx, `CREATE TABLE products (...)`)
    },
    Down: func(ctx context.Context, exec migrations.Executor) error {
        return exec.Exec(ctx, "DROP TABLE IF EXISTS products")
    },
}
```

### 3. Manager Usage

```go
// Create manager
manager := migrations.NewManager(repo, executor)

// Register migrations
manager.RegisterMultiple(allMigrations)

// Check status
status, err := manager.Status(ctx)

// Run migrations
err := manager.Up(ctx)
```

## Adapting for Other Databases

### MySQL

Use MySQL-specific syntax:

```go
// Use ? for parameters
return exec.Exec(ctx, "INSERT INTO products (id, name) VALUES (?, ?)", id, name)

// Use MySQL-specific types
CREATE TABLE products (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

## Testing Migrations

Run migrations, then verify tables exist:

```bash
# Run migrations
go run main.go

# Verify with PostgreSQL CLI
docker-compose exec postgres psql -U edomain -d edomain
> \dt
> \d products
```

## Rollback Example

To test rollbacks, modify `main.go`:

```go
// After running Up, test rollback
fmt.Println("\n⬇️  Testing rollback...")
if err := manager.Down(ctx); err != nil {
    log.Fatal(err)
}

status, _ = manager.Status(ctx)
fmt.Printf("After rollback: %d applied, %d pending\n", 
    len(status.Applied), len(status.Pending))
```

## Integration with Sample Project

To integrate with the sample-project API:

1. Add database connection to `sample-project/main.go`
2. Run migrations at startup
3. Replace `MemoryStore` with database implementations

```go
// In sample-project/main.go
func main() {
    // Initialize database
    connStr := "host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    
    // Run migrations
    executor := NewPostgreSQLExecutor(db)
    repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
    manager := migrations.NewManager(repo, executor)
    manager.RegisterMultiple(allMigrations)
    
    if err := manager.Up(context.Background()); err != nil {
        log.Fatal(err)
    }
    
    // Continue with server setup...
}
```
