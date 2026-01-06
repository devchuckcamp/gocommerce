package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/devchuckcamp/gocommerce/migrations"
	_ "github.com/lib/pq"
)

// PostgreSQLExecutor implements migrations.Executor for PostgreSQL database.
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
	
	return scanRows(rows)
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

// scanRows converts sql.Rows to []map[string]interface{}
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	result := make([]map[string]interface{}, 0)
	
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			
			// Handle []byte conversion
			if b, ok := val.([]byte); ok {
				v = string(b)
			} else {
				v = val
			}
			
			rowMap[col] = v
		}
		
		result = append(result, rowMap)
	}
	
	return result, rows.Err()
}

func main() {
	ctx := context.Background()
	
	connStr := os.Getenv("DB_DSN")
	if connStr == "" {
		connStr = os.Getenv("DATABASE_URL")
	}
	if connStr == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "edomain"
		}
		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "edomain"
		}
		name := os.Getenv("DB_NAME")
		if name == "" {
			name = "edomain"
		}
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, name)
	}
	
	fmt.Println("🔄 GoCommerce PostgreSQL Migration Tool")
	fmt.Println("========================================\n")
	
	// Connect to PostgreSQL
	fmt.Println("📡 Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v\n\nMake sure PostgreSQL is running:\n  docker-compose up -d\n", err)
	}
	fmt.Println("✓ Connected to PostgreSQL\n")
	
	// Create executor and repository
	executor := NewPostgreSQLExecutor(db)
	repo := migrations.NewPostgreSQLRepository(executor, "gocommerce_migrations")
	
	// Create migration manager
	manager := migrations.NewManager(repo, executor)
	
	// Register PostgreSQL migrations
	if err := manager.RegisterMultiple(migrations.PostgreSQLExampleMigrations); err != nil {
		log.Fatal(err)
	}
	
	// Check current status
	fmt.Println("📊 Current Migration Status:")
	status, err := manager.Status(ctx)
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Printf("   Applied: %d migrations\n", len(status.Applied))
	for _, m := range status.Applied {
		fmt.Printf("   ✓ %s - %s (applied: %v)\n", m.Version, m.Name, m.AppliedAt.Format("2006-01-02 15:04:05"))
	}
	
	fmt.Printf("   Pending: %d migrations\n", len(status.Pending))
	for _, m := range status.Pending {
		fmt.Printf("   ○ %s - %s\n", m.Version, m.Name)
	}
	fmt.Println()
	
	if len(status.Pending) > 0 {
		// Run migrations
		fmt.Println("⬆️  Running pending migrations...")
		if err := manager.Up(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		
		// Show updated status
		status, err = manager.Status(ctx)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Println("✅ Migrations completed!")
		fmt.Printf("   Total applied: %d migrations\n\n", len(status.Applied))
		for _, m := range status.Applied {
			fmt.Printf("   ✓ %s - %s\n", m.Version, m.Name)
		}
		
		fmt.Println("\n📋 Created Tables:")
		fmt.Println("   - brands")
		fmt.Println("   - categories")
		fmt.Println("   - products")
		fmt.Println("   - variants")
		fmt.Println("   - carts")
		fmt.Println("   - cart_items")
		fmt.Println("   - orders")
		fmt.Println("   - order_items")
		fmt.Println("   - promotions")
		
		fmt.Println("\n💡 Verify with:")
		fmt.Println("   docker-compose exec postgres psql -U edomain -d edomain -c '\\dt'")
	} else {
		fmt.Println("✅ Database is up to date!")
		
		fmt.Println("\n💡 View tables:")
		fmt.Println("   docker-compose exec postgres psql -U edomain -d edomain -c '\\dt'")
		
		fmt.Println("\n💡 Rollback last migration:")
		fmt.Println("   go run main.go down")
	}
}
