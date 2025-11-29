# Using GoCommerce Migrations in Your Project

This guide shows how to integrate the gocommerce migration system into your own project.

## Quick Start

### 1. Install the Package

```bash
go get github.com/devchuckcamp/gocommerce
```

### 2. Copy the Executor Implementation

Copy the PostgreSQL executor from `migrations/examples/postgresql/main.go` into your project:

```go
// internal/database/executor.go
package database

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/devchuckcamp/gocommerce/migrations"
)

type PostgreSQLExecutor struct {
    db *sql.DB
    tx *sql.Tx
}

func NewPostgreSQLExecutor(db *sql.DB) *PostgreSQLExecutor {
    return &PostgreSQLExecutor{db: db}
}

func (e *PostgreSQLExecutor) Exec(ctx context.Context, query string, args ...interface{}) error {
    if e.tx != nil {
        _, err := e.tx.ExecContext(ctx, query, args...)
        return err
    }
    _, err := e.db.ExecContext(ctx, query, args...)
    return err
}

func (e *PostgreSQLExecutor) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
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
    
    cols, err := rows.Columns()
    if err != nil {
        return nil, err
    }
    
    var results []map[string]interface{}
    for rows.Next() {
        values := make([]interface{}, len(cols))
        valuePtrs := make([]interface{}, len(cols))
        for i := range values {
            valuePtrs[i] = &values[i]
        }
        
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
        }
        
        row := make(map[string]interface{})
        for i, col := range cols {
            row[col] = values[i]
        }
        results = append(results, row)
    }
    
    return results, rows.Err()
}

func (e *PostgreSQLExecutor) Begin(ctx context.Context) (migrations.Executor, error) {
    tx, err := e.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    return &PostgreSQLExecutor{db: e.db, tx: tx}, nil
}

func (e *PostgreSQLExecutor) Commit(ctx context.Context) error {
    if e.tx == nil {
        return fmt.Errorf("no transaction to commit")
    }
    return e.tx.Commit()
}

func (e *PostgreSQLExecutor) Rollback(ctx context.Context) error {
    if e.tx == nil {
        return fmt.Errorf("no transaction to rollback")
    }
    return e.tx.Rollback()
}
```

### 3. Create Migration Runner

Create a function to run migrations in your project:

```go
// internal/database/migrate.go
package database

import (
    "context"
    "database/sql"
    "log"
    
    "github.com/devchuckcamp/gocommerce/migrations"
)

func RunMigrations(db *sql.DB) error {
    executor := NewPostgreSQLExecutor(db)
    repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
    manager := migrations.NewManager(repo, executor)
    
    // Use pre-built migrations
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

func SeedDatabase(db *sql.DB) error {
    executor := NewPostgreSQLExecutor(db)
    seeder := migrations.NewSeeder(executor)
    
    // Use pre-built seeds
    seeder.RegisterMultiple(migrations.AllSeeds)
    
    ctx := context.Background()
    if err := seeder.Run(ctx); err != nil {
        return err
    }
    
    log.Println("✅ Database seeded with test data!")
    return nil
}
```

## Integration Patterns

### Pattern 1: Run at Startup (Recommended for Development)

```go
// main.go
package main

import (
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "yourapp/internal/database"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", getDatabaseURL())
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Run migrations automatically
    if err := database.RunMigrations(db); err != nil {
        log.Fatal("Migration failed:", err)
    }
    
    // Start application
    startServer(db)
}
```

### Pattern 2: Separate CLI Tool (Recommended for Production)

```go
// cmd/migrate/main.go
package main

import (
    "database/sql"
    "flag"
    "fmt"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    "yourapp/internal/database"
)

func main() {
    var (
        up   = flag.Bool("up", false, "Run pending migrations")
        down = flag.Bool("down", false, "Rollback last migration")
        seed = flag.Bool("seed", false, "Seed database")
        dsn  = flag.String("dsn", os.Getenv("DATABASE_URL"), "Database DSN")
    )
    flag.Parse()
    
    db, err := sql.Open("postgres", *dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    switch {
    case *up:
        if err := database.RunMigrations(db); err != nil {
            log.Fatal(err)
        }
    case *seed:
        if err := database.SeedDatabase(db); err != nil {
            log.Fatal(err)
        }
    case *down:
        // Implement rollback logic
        fmt.Println("Rollback not implemented")
    default:
        flag.Usage()
    }
}
```

Usage:
```bash
# Run migrations
go run cmd/migrate/main.go -up

# Seed database
go run cmd/migrate/main.go -seed

# Or build and use
go build -o migrate cmd/migrate/main.go
./migrate -up
./migrate -seed
```

### Pattern 3: Docker Entrypoint

```dockerfile
FROM golang:1.21-alpine

WORKDIR /app
COPY . .

RUN go build -o /app/migrate cmd/migrate/main.go
RUN go build -o /app/server cmd/server/main.go

# Run migrations then start server
CMD ./migrate -up && ./server
```

### Pattern 4: Kubernetes Init Container

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      initContainers:
      - name: migrations
        image: myapp:latest
        command: ["/app/migrate", "-up"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
      containers:
      - name: app
        image: myapp:latest
        command: ["/app/server"]
```

## Environment-Specific Configuration

### Development
```go
// Run migrations + seeds automatically
if os.Getenv("APP_ENV") == "development" {
    database.RunMigrations(db)
    database.SeedDatabase(db)
}
```

### Staging/Production
```bash
# Run migrations manually before deployment
./migrate -up -dsn="$DATABASE_URL"

# Or use CI/CD pipeline
- name: Run Migrations
  run: go run cmd/migrate/main.go -up
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

## What Gets Created

Running `migrations.PostgreSQLExampleMigrations` creates these tables:

### 1. brands
```sql
CREATE TABLE brands (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    logo_url TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. categories
```sql
CREATE TABLE categories (
    id VARCHAR(255) PRIMARY KEY,
    parent_id VARCHAR(255) REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    is_active BOOLEAN DEFAULT true,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 3. products
```sql
CREATE TABLE products (
    id VARCHAR(255) PRIMARY KEY,
    sku VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    brand_id VARCHAR(255) REFERENCES brands(id),
    category_id VARCHAR(255) REFERENCES categories(id),
    base_price_amount BIGINT NOT NULL,
    base_price_currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'draft',
    images TEXT,
    attributes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 4. carts
```sql
CREATE TABLE carts (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 5. cart_items
```sql
CREATE TABLE cart_items (
    id VARCHAR(255) PRIMARY KEY,
    cart_id VARCHAR(255) REFERENCES carts(id),
    product_id VARCHAR(255) REFERENCES products(id),
    quantity INTEGER NOT NULL,
    unit_price_amount BIGINT NOT NULL,
    unit_price_currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 6. orders
```sql
CREATE TABLE orders (
    id VARCHAR(255) PRIMARY KEY,
    order_number VARCHAR(255) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    total_amount BIGINT NOT NULL,
    total_currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Seeding Test Data

The package includes built-in seeds:

```go
// Seed database with test data
executor := NewPostgreSQLExecutor(db)
seeder := migrations.NewSeeder(executor)
seeder.RegisterMultiple(migrations.AllSeeds)

if err := seeder.Run(context.Background()); err != nil {
    log.Fatal(err)
}
```

This creates:
- **8 brands**: Apple, Dell, Lenovo, HP, Samsung, Logitech, Sony, Bose
- **8 categories**: Electronics, Computers, Laptops, Accessories, Audio, Storage, Input Devices, Peripherals
- **72 products**: 22 curated products (MacBook Pro, Dell XPS, iPhone, etc.) + 50 random products

## Custom Migrations

If you need to add custom tables beyond the pre-built ones:

```go
var customMigration = migrations.Migration{
    Version: "007",
    Name:    "create_reviews_table",
    Up: func(ctx context.Context, exec migrations.Executor) error {
        return exec.Exec(ctx, `
            CREATE TABLE reviews (
                id VARCHAR(255) PRIMARY KEY,
                product_id VARCHAR(255) REFERENCES products(id),
                user_id VARCHAR(255) NOT NULL,
                rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
                comment TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        `)
    },
    Down: func(ctx context.Context, exec migrations.Executor) error {
        return exec.Exec(ctx, "DROP TABLE IF EXISTS reviews")
    },
}

// Register both pre-built and custom migrations
manager.RegisterMultiple(migrations.PostgreSQLExampleMigrations)
manager.Register(customMigration)
```

## Troubleshooting

### Connection Issues
```go
// Test database connection first
if err := db.Ping(); err != nil {
    log.Fatal("Database connection failed:", err)
}
```

### Migration Table Already Exists
The system uses a custom table name `gocommerce_migrations` to avoid conflicts. If you need a different name:

```go
repo := migrations.NewPostgreSQLRepository(executor, "my_custom_migrations_table")
```

### Migrations Out of Order
Always run migrations in order. The system tracks which migrations have been applied.

```go
// Check status before running
status, _ := manager.Status(ctx)
fmt.Printf("Applied: %d, Pending: %d\n", len(status.Applied), len(status.Pending))
```

## Next Steps

1. Copy the executor code into your project
2. Create a migration runner function
3. Choose an integration pattern (startup, CLI, or Docker)
4. Run migrations in your development environment
5. Add migration step to your deployment pipeline

See [migrations/README.md](../README.md) for complete API documentation.
