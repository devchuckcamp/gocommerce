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

func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, rows.Err()
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

	fmt.Println("🌱 GoCommerce Database Seeder")
	fmt.Println("========================================")
	fmt.Println()

	// Connect to database
	fmt.Println("📡 Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v\n"+
			"Make sure PostgreSQL is running:\n"+
			"  docker-compose up -d", err)
	}
	fmt.Println("✓ Connected to PostgreSQL")
	fmt.Println()

	// Create executor
	executor := NewPostgreSQLExecutor(db)

	// Create seeder
	seeder := migrations.NewSeeder(executor)

	// Register all seeds (brands and categories must come first)
	seeder.RegisterMultiple(migrations.AllSeeds)

	// Display available seeds
	fmt.Println("📋 Available Seeds:")
	for i, seed := range seeder.List() {
		fmt.Printf("   %d. %s\n", i+1, seed.Name)
		fmt.Printf("      %s\n", seed.Description)
	}
	fmt.Println()

	// Check current product count
	var countBefore int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&countBefore)
	fmt.Printf("📊 Products before seeding: %d\n", countBefore)
	fmt.Println()

	// Run seeds
	fmt.Println("🌱 Running seeds...")
	if err := seeder.Run(ctx); err != nil {
		log.Fatalf("Failed to run seeds: %v", err)
	}
	fmt.Println()

	// Check final product count
	var countAfter int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&countAfter)
	fmt.Printf("✅ Seeding completed!\n")
	fmt.Printf("   Products after seeding: %d\n", countAfter)
	fmt.Printf("   New products added: %d\n", countAfter-countBefore)
	fmt.Println()

	// Display sample products
	fmt.Println("📦 Sample Products:")
	rows, err := db.Query(`
		SELECT name, sku, base_price_amount/100.0 as price, status 
		FROM products 
		ORDER BY created_at DESC 
		LIMIT 10
	`)
	if err != nil {
		log.Fatalf("Failed to query products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, sku, status string
		var price float64
		if err := rows.Scan(&name, &sku, &price, &status); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		fmt.Printf("   • %s (SKU: %s) - $%.2f [%s]\n", name, sku, price, status)
	}
	fmt.Println()

	fmt.Println("💡 View all products:")
	fmt.Println("   docker-compose exec postgres psql -U edomain -d edomain -c \"SELECT id, name, sku, base_price_amount/100.0 as price FROM products LIMIT 20;\"")
	fmt.Println()

	fmt.Println("💡 To run specific seeds only:")
	fmt.Println("   seeder.RunSingle(ctx, \"product_seeder\")")
}
