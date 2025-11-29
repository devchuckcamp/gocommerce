package migrations

import (
	"context"
	"fmt"
	"time"
)

// TableName is the name of the table used to track migration state.
// Using a domain-specific prefix to avoid conflicts with existing projects.
const TableName = "gocommerce_migrations"

// SQLRepository implements Repository for SQL databases.
type SQLRepository struct {
	executor  Executor
	tableName string
}

// NewSQLRepository creates a new SQL-based migration repository.
// If tableName is empty, it uses the default "gocommerce_migrations".
func NewSQLRepository(executor Executor, tableName string) *SQLRepository {
	if tableName == "" {
		tableName = TableName
	}
	return &SQLRepository{
		executor:  executor,
		tableName: tableName,
	}
}

// InitializeSchema creates the migration tracking table if it doesn't exist.
func (r *SQLRepository) InitializeSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`, r.tableName)
	
	return r.executor.Exec(ctx, query)
}

// GetAppliedMigrations returns all migrations that have been applied.
func (r *SQLRepository) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	query := fmt.Sprintf(`
		SELECT version, name, applied_at 
		FROM %s 
		ORDER BY version ASC
	`, r.tableName)
	
	rows, err := r.executor.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	
	migrations := make([]Migration, 0, len(rows))
	for _, row := range rows {
		migration := Migration{
			Version: row["version"].(string),
			Name:    row["name"].(string),
		}
		
		if appliedAt, ok := row["applied_at"].(time.Time); ok {
			migration.AppliedAt = &appliedAt
		}
		
		migrations = append(migrations, migration)
	}
	
	return migrations, nil
}

// RecordMigration records that a migration was applied.
func (r *SQLRepository) RecordMigration(ctx context.Context, migration Migration) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, name, applied_at)
		VALUES (?, ?, ?)
	`, r.tableName)
	
	appliedAt := time.Now()
	if migration.AppliedAt != nil {
		appliedAt = *migration.AppliedAt
	}
	
	return r.executor.Exec(ctx, query, migration.Version, migration.Name, appliedAt)
}

// RemoveMigration removes a migration record (for rollback).
func (r *SQLRepository) RemoveMigration(ctx context.Context, version string) error {
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = ?
	`, r.tableName)
	
	return r.executor.Exec(ctx, query, version)
}

// PostgreSQLRepository implements Repository for PostgreSQL databases.
type PostgreSQLRepository struct {
	executor  Executor
	tableName string
}

// NewPostgreSQLRepository creates a PostgreSQL-specific migration repository.
func NewPostgreSQLRepository(executor Executor, tableName string) *PostgreSQLRepository {
	if tableName == "" {
		tableName = TableName
	}
	return &PostgreSQLRepository{
		executor:  executor,
		tableName: tableName,
	}
}

// InitializeSchema creates the migration tracking table if it doesn't exist.
func (r *PostgreSQLRepository) InitializeSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`, r.tableName)
	
	return r.executor.Exec(ctx, query)
}

// GetAppliedMigrations returns all migrations that have been applied.
func (r *PostgreSQLRepository) GetAppliedMigrations(ctx context.Context) ([]Migration, error) {
	query := fmt.Sprintf(`
		SELECT version, name, applied_at 
		FROM %s 
		ORDER BY version ASC
	`, r.tableName)
	
	rows, err := r.executor.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	
	migrations := make([]Migration, 0, len(rows))
	for _, row := range rows {
		migration := Migration{
			Version: row["version"].(string),
			Name:    row["name"].(string),
		}
		
		if appliedAt, ok := row["applied_at"].(time.Time); ok {
			migration.AppliedAt = &appliedAt
		}
		
		migrations = append(migrations, migration)
	}
	
	return migrations, nil
}

// RecordMigration records that a migration was applied.
func (r *PostgreSQLRepository) RecordMigration(ctx context.Context, migration Migration) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, name, applied_at)
		VALUES ($1, $2, $3)
	`, r.tableName)
	
	appliedAt := time.Now()
	if migration.AppliedAt != nil {
		appliedAt = *migration.AppliedAt
	}
	
	return r.executor.Exec(ctx, query, migration.Version, migration.Name, appliedAt)
}

// RemoveMigration removes a migration record (for rollback).
func (r *PostgreSQLRepository) RemoveMigration(ctx context.Context, version string) error {
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = $1
	`, r.tableName)
	
	return r.executor.Exec(ctx, query, version)
}
