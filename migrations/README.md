# GoCommerce Migrations

A flexible, database-agnostic migration system for the gocommerce domain library.

## Features

- ✅ **Database Agnostic** - Works with any SQL database via the `Executor` interface
- ✅ **Transaction Safety** - All migrations run within transactions
- ✅ **Rollback Support** - Reverse migrations with `Down` functions
- ✅ **Version Tracking** - Custom table name (`gocommerce_migrations`) to avoid conflicts
- ✅ **Flexible Versioning** - Support for semantic versions, timestamps, or custom schemes
- ✅ **Status Reporting** - Check applied and pending migrations
- ✅ **Partial Migrations** - Run migrations up to a specific version
- ✅ **Database Seeding** - Built-in seeder system for mock/test data
- ✅ **Complete Schema** - Pre-built migrations for brands, categories, products, carts, orders

## Quick Start

### 1. Define Migrations

```go
package main

import (
    "context"
    "github.com/devchuckcamp/gocommerce/migrations"
)

func createProductsTable(ctx context.Context, exec migrations.Executor) error {
    return exec.Exec(ctx, `
        CREATE TABLE products (
            id VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            price BIGINT NOT NULL,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        )
    `)
}

func dropProductsTable(ctx context.Context, exec migrations.Executor) error {
    return exec.Exec(ctx, "DROP TABLE products")
}

var migration001 = migrations.Migration{
    Version: "001",
    Name:    "create_products_table",
    Up:      createProductsTable,
    Down:    dropProductsTable,
}
```

### 2. Register and Run Migrations

```go
// Create executor (implement migrations.Executor interface)
executor := NewDatabaseExecutor(db)

// Create repository
repo := migrations.NewSQLRepository(executor, "gocommerce_migrations")

// Create manager
manager := migrations.NewManager(repo, executor)

// Register migrations
manager.Register(migration001)
manager.Register(migration002)
// ... or use RegisterMultiple

// Run all pending migrations
if err := manager.Up(ctx); err != nil {
    log.Fatal(err)
}
```

### 3. Check Migration Status

```go
status, err := manager.Status(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Applied: %d migrations\n", len(status.Applied))
fmt.Printf("Pending: %d migrations\n", len(status.Pending))
```

## Migration Versioning

### Timestamp-Based (Recommended)

```go
var migration = migrations.Migration{
    Version: "20231128_001",
    Name:    "create_products_table",
    Up:      createProductsTable,
    Down:    dropProductsTable,
}
```

### Sequential Numbering

```go
var migration = migrations.Migration{
    Version: "001",
    Name:    "create_products_table",
    Up:      createProductsTable,
    Down:    dropProductsTable,
}
```

### Semantic Versioning

```go
var migration = migrations.Migration{
    Version: "v1.0.0",
    Name:    "initial_schema",
    Up:      createInitialSchema,
    Down:    dropInitialSchema,
}
```

### Using Generator

```go
gen := migrations.NewGenerator("gocommerce")

migration := gen.NewMigration("create_products_table", 1, 
    createProductsTable, 
    dropProductsTable,
)
// Version: gocommerce_20231128_150405_001
```

## Operations

### Run All Pending Migrations

```go
err := manager.Up(ctx)
```

### Run Migrations Up To Version

```go
err := manager.UpTo(ctx, "20231128_005")
```

### Rollback Last Migration

```go
err := manager.Down(ctx)
```

### Rollback To Version

```go
err := manager.DownTo(ctx, "20231128_003")
```

### Check Status

```go
status, err := manager.Status(ctx)
for _, m := range status.Applied {
    fmt.Printf("✓ %s - %s (applied: %v)\n", 
        m.Version, m.Name, m.AppliedAt)
}
for _, m := range status.Pending {
    fmt.Printf("○ %s - %s (pending)\n", 
        m.Version, m.Name)
}
```

## Implementing the Executor Interface

The migration system requires an `Executor` implementation for your database:

```go
type DatabaseExecutor struct {
    db *sql.DB
    tx *sql.Tx
}

func (e *DatabaseExecutor) Exec(ctx context.Context, query string, args ...interface{}) error {
    if e.tx != nil {
        _, err := e.tx.ExecContext(ctx, query, args...)
        return err
    }
    _, err := e.db.ExecContext(ctx, query, args...)
    return err
}

func (e *DatabaseExecutor) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
    var rows *sql.Rows
    var err error
    
    if e.tx != nil {
        rows, err = e.tx.QueryContext(ctx, query, args...)
    } else {
        rows, err = e.db.QueryContext(ctx, query, args...)
    }
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // Convert rows to []map[string]interface{}
    return scanRows(rows)
}

func (e *DatabaseExecutor) Begin(ctx context.Context) (migrations.Executor, error) {
    tx, err := e.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return &DatabaseExecutor{db: e.db, tx: tx}, nil
}

func (e *DatabaseExecutor) Commit(ctx context.Context) error {
    if e.tx == nil {
        return fmt.Errorf("no transaction to commit")
    }
    return e.tx.Commit()
}

func (e *DatabaseExecutor) Rollback(ctx context.Context) error {
    if e.tx == nil {
        return fmt.Errorf("no transaction to rollback")
    }
    return e.tx.Rollback()
}
```

## Custom Table Name

To avoid conflicts with existing migration systems:

```go
// Use custom table name
repo := migrations.NewSQLRepository(executor, "my_app_migrations")
```

Default table name is `gocommerce_migrations`.

## Migration Best Practices

### 1. Always Provide Down Functions

```go
// ✅ Good
var migration = migrations.Migration{
    Version: "001",
    Name:    "add_products",
    Up:      createProductsTable,
    Down:    dropProductsTable,  // Rollback supported
}

// ❌ Avoid
var migration = migrations.Migration{
    Version: "001",
    Name:    "add_products",
    Up:      createProductsTable,
    Down:    nil,  // Cannot rollback!
}
```

### 2. Use Transactions for Data Changes

```go
func seedProducts(ctx context.Context, exec migrations.Executor) error {
    products := []struct{ id, name string }{
        {"prod-1", "T-Shirt"},
        {"prod-2", "Jeans"},
    }
    
    for _, p := range products {
        err := exec.Exec(ctx, 
            "INSERT INTO products (id, name) VALUES (?, ?)",
            p.id, p.name,
        )
        if err != nil {
            return err  // Transaction will rollback automatically
        }
    }
    
    return nil
}
```

### 3. Make Migrations Idempotent

```go
func createProductsTable(ctx context.Context, exec migrations.Executor) error {
    // Use IF NOT EXISTS for safety
    return exec.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS products (
            id VARCHAR(255) PRIMARY KEY,
            name VARCHAR(255) NOT NULL
        )
    `)
}
```

### 4. Test Migrations

```go
func TestMigration001(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    
    executor := NewDatabaseExecutor(db)
    repo := migrations.NewSQLRepository(executor, "test_migrations")
    manager := migrations.NewManager(repo, executor)
    
    // Register and run migration
    manager.Register(migration001)
    
    if err := manager.Up(ctx); err != nil {
        t.Fatalf("migration failed: %v", err)
    }
    
    // Verify table exists
    // ... assertions
    
    // Test rollback
    if err := manager.Down(ctx); err != nil {
        t.Fatalf("rollback failed: %v", err)
    }
    
    // Verify table is gone
    // ... assertions
}
```

## Example: Complete Schema Migration

```go
package migrations

import (
    "context"
    "github.com/devchuckcamp/gocommerce/migrations"
)

// All migrations for gocommerce domain
var AllMigrations = []migrations.Migration{
    {
        Version: "001",
        Name:    "create_products_table",
        Up: func(ctx context.Context, exec migrations.Executor) error {
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
                    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                )
            `)
        },
        Down: func(ctx context.Context, exec migrations.Executor) error {
            return exec.Exec(ctx, "DROP TABLE products")
        },
    },
    {
        Version: "002",
        Name:    "create_carts_table",
        Up: func(ctx context.Context, exec migrations.Executor) error {
            return exec.Exec(ctx, `
                CREATE TABLE carts (
                    id VARCHAR(255) PRIMARY KEY,
                    user_id VARCHAR(255),
                    session_id VARCHAR(255),
                    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    expires_at TIMESTAMP
                )
            `)
        },
        Down: func(ctx context.Context, exec migrations.Executor) error {
            return exec.Exec(ctx, "DROP TABLE carts")
        },
    },
    {
        Version: "003",
        Name:    "create_orders_table",
        Up: func(ctx context.Context, exec migrations.Executor) error {
            return exec.Exec(ctx, `
                CREATE TABLE orders (
                    id VARCHAR(255) PRIMARY KEY,
                    order_number VARCHAR(255) UNIQUE NOT NULL,
                    user_id VARCHAR(255) NOT NULL,
                    status VARCHAR(50) NOT NULL,
                    subtotal_amount BIGINT NOT NULL,
                    subtotal_currency VARCHAR(3) NOT NULL,
                    tax_amount BIGINT NOT NULL,
                    shipping_amount BIGINT NOT NULL,
                    total_amount BIGINT NOT NULL,
                    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                )
            `)
        },
        Down: func(ctx context.Context, exec migrations.Executor) error {
            return exec.Exec(ctx, "DROP TABLE orders")
        },
    },
}
```

## Database Support

The migration system is **database-agnostic** but includes helpers for:

- **MySQL/MariaDB** - Use `SQLRepository`
- **PostgreSQL** - Use `PostgreSQLRepository` (uses `$1` placeholders)
- **Custom** - Implement `Executor` and `Repository` interfaces

## Architecture

```
┌─────────────────────────────────┐
│   Manager (Orchestrator)        │
│   - Register migrations          │
│   - Up/Down/Status operations    │
└────────────┬────────────────────┘
             │
    ┌────────┴────────┐
    │                 │
    ▼                 ▼
┌─────────┐    ┌──────────┐
│Executor │    │Repository│
│Interface│    │Interface │
└─────────┘    └──────────┘
    │              │
    ▼              ▼
┌─────────┐    ┌──────────┐
│Database │    │Migration │
│Driver   │    │Tracking  │
└─────────┘    └──────────┘
```

## Why "gocommerce_migrations"?

Using a domain-specific table name (`gocommerce_migrations`) instead of generic names like `schema_migrations` prevents conflicts when:

- Integrating into existing projects that already have migration systems
- Using multiple migration systems in the same database
- Deploying alongside other packages/frameworks

You can customize the table name:

```go
repo := migrations.NewSQLRepository(executor, "my_custom_name")
```

## Error Handling

All migration operations are transactional:

```go
err := manager.Up(ctx)
if err != nil {
    // Migration failed and was rolled back
    // Database is in consistent state
    log.Printf("Migration failed: %v", err)
}
```

## Database Seeding

The migration system includes a built-in seeder for populating tables with test/mock data:

```go
// Create seeder
seeder := migrations.NewSeeder(executor)

// Register seeds
seeder.RegisterMultiple(migrations.AllSeeds)

// Run all seeds
if err := seeder.Run(ctx); err != nil {
    log.Fatal(err)
}

// Or run a specific seed
if err := seeder.RunSingle(ctx, "product_seeder"); err != nil {
    log.Fatal(err)
}
```

### Built-in Seeds

- **BrandSeed** - 8 realistic brands (Apple, Dell, Lenovo, HP, Samsung, Logitech, Sony, Bose)
- **CategorySeed** - 8 hierarchical categories (Electronics, Computers, Accessories, etc.)
- **ProductSeed** - 22 curated products with full details
- **RandomProductSeed** - 50 random products for load testing

All seeds are transaction-safe and idempotent (safe to run multiple times).

### Custom Seeds

Define your own seeds:

```go
var CustomSeed = migrations.Seed{
    Name:        "custom_seeder",
    Description: "Seeds custom data",
    Run: func(ctx context.Context, exec migrations.Executor) error {
        // Your seeding logic here
        return exec.Exec(ctx, "INSERT INTO ...")
    },
}

seeder.Register(CustomSeed)
```

## Complete Example with PostgreSQL

See `migrations/examples/postgresql/` for a complete working example:

```bash
# Start PostgreSQL
cd migrations/examples && docker-compose up -d

# Run migrations
cd postgresql
cd migrations/examples/postgresql
go run main.go

# Seed database
go run seed-products.go
```

This creates:
- 6 tables (brands, categories, products, carts, cart_items, orders)
- 8 brands with realistic data
- 8 categories with parent-child relationships  
- 72 products (22 curated + 50 random)

## Using Pre-Built Migrations in Your Project

The gocommerce package includes pre-built migrations for a complete e-commerce schema. You can use them directly in your project:

### Option 1: Use Pre-Built Migrations (Recommended)

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "github.com/devchuckcamp/gocommerce/migrations"
)

func main() {
    // Connect to your database
    db, err := sql.Open("postgres", "your-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create executor (you'll need to implement this - see examples)
    executor := NewPostgreSQLExecutor(db)
    
    // Create repository with custom table name
    repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
    
    // Create manager
    manager := migrations.NewManager(repo, executor)
    
    // Use pre-built PostgreSQL migrations
    manager.RegisterMultiple(migrations.PostgreSQLExampleMigrations)
    
    // Run migrations
    ctx := context.Background()
    if err := manager.Up(ctx); err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ Migrations completed!")
    
    // Optional: Seed database with test data
    seeder := migrations.NewSeeder(executor)
    seeder.RegisterMultiple(migrations.AllSeeds)
    if err := seeder.Run(ctx); err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ Database seeded!")
}
```

### Option 2: Run Migrations from CLI Tool

Create a simple migration tool in your project:

```go
// cmd/migrate/main.go
package main

import (
    "context"
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    "github.com/devchuckcamp/gocommerce/migrations"
)

func main() {
    var (
        up   = flag.Bool("up", false, "Run pending migrations")
        down = flag.Bool("down", false, "Rollback last migration")
        seed = flag.Bool("seed", false, "Seed database with test data")
        dsn  = flag.String("dsn", os.Getenv("DATABASE_URL"), "Database connection string")
    )
    flag.Parse()
    
    // Connect to database
    db, err := sql.Open("postgres", *dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Setup migration manager
    executor := NewPostgreSQLExecutor(db)
    repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
    manager := migrations.NewManager(repo, executor)
    manager.RegisterMultiple(migrations.PostgreSQLExampleMigrations)
    
    ctx := context.Background()
    
    // Run commands
    switch {
    case *up:
        if err := manager.Up(ctx); err != nil {
            log.Fatal(err)
        }
        fmt.Println("✅ Migrations completed!")
        
    case *down:
        if err := manager.Down(ctx); err != nil {
            log.Fatal(err)
        }
        fmt.Println("✅ Rollback completed!")
        
    case *seed:
        seeder := migrations.NewSeeder(executor)
        seeder.RegisterMultiple(migrations.AllSeeds)
        if err := seeder.Run(ctx); err != nil {
            log.Fatal(err)
        }
        fmt.Println("✅ Database seeded!")
        
    default:
        status, _ := manager.Status(ctx)
        fmt.Printf("Applied: %d migrations\n", len(status.Applied))
        fmt.Printf("Pending: %d migrations\n", len(status.Pending))
    }
}
```

Then run:
```bash
# Run migrations
go run cmd/migrate/main.go -up -dsn="postgres://user:pass@localhost/dbname"

# Seed database
go run cmd/migrate/main.go -seed -dsn="postgres://user:pass@localhost/dbname"

# Check status
go run cmd/migrate/main.go -dsn="postgres://user:pass@localhost/dbname"
```

### Option 3: Run Migrations at Application Startup

```go
// main.go
package main

import (
    "context"
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "github.com/devchuckcamp/gocommerce/migrations"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", getDatabaseURL())
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Run migrations automatically on startup
    if err := runMigrations(db); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Start your application...
    startServer(db)
}

func runMigrations(db *sql.DB) error {
    executor := NewPostgreSQLExecutor(db)
    repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
    manager := migrations.NewManager(repo, executor)
    
    // Register pre-built migrations
    manager.RegisterMultiple(migrations.PostgreSQLExampleMigrations)
    
    ctx := context.Background()
    
    // Check status
    status, err := manager.Status(ctx)
    if err != nil {
        return err
    }
    
    if len(status.Pending) == 0 {
        log.Println("✅ Database is up to date")
        return nil
    }
    
    log.Printf("⬆️  Running %d pending migrations...\n", len(status.Pending))
    
    // Run migrations
    if err := manager.Up(ctx); err != nil {
        return err
    }
    
    log.Println("✅ Migrations completed!")
    return nil
}
```

### What Tables Are Created?

The pre-built migrations create these tables:

1. **brands** - Product brands (Apple, Dell, etc.)
2. **categories** - Hierarchical product categories  
3. **products** - Products with brand/category relationships
4. **carts** - Shopping carts
5. **cart_items** - Cart line items
6. **orders** - Customer orders

### Implementing the Executor Interface

You need to implement the `migrations.Executor` interface for your database. See `migrations/examples/postgresql/main.go` for a complete PostgreSQL example:

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

## Integration Example

See the `sample-project` directory for a complete example integrating migrations with the gocommerce domain library.
