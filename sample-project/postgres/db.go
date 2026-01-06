package postgres

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

const defaultConnStr = "host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable"

func Open() (*sql.DB, error) {
	connStr := connectionStringFromEnv()
	return sql.Open("postgres", connStr)
}

func connectionStringFromEnv() string {
	// Prefer an explicit DSN.
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	// Otherwise build from parts.
	driver := os.Getenv("DB_DRIVER")
	if driver != "" && driver != "postgres" {
		// This package is Postgres-only; fall back to defaults.
		return defaultConnStr
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" && port == "" && user == "" && password == "" && name == "" {
		return defaultConnStr
	}
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if name == "" {
		name = "postgres"
	}

	// lib/pq supports keyword/value connection strings.
	conn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable", host, port, user, name)
	if password != "" {
		conn += " password=" + password
	}
	return conn
}
