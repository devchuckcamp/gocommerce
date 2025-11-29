package migrations

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// Migration represents a single database migration.
type Migration struct {
	// Version is the unique migration identifier (e.g., "20231128_001", "v1.0.0")
	Version string
	
	// Name is a human-readable description
	Name string
	
	// Up applies the migration
	Up MigrationFunc
	
	// Down reverts the migration
	Down MigrationFunc
	
	// AppliedAt is when the migration was executed (set by the system)
	AppliedAt *time.Time
}

// MigrationFunc is a function that executes migration logic.
// It receives a context and an Executor for running queries.
type MigrationFunc func(ctx context.Context, exec Executor) error

// Executor defines the interface for executing migrations.
// This abstraction allows the migration system to work with any database.
type Executor interface {
	// Exec executes a query without returning rows
	Exec(ctx context.Context, query string, args ...interface{}) error
	
	// Query executes a query that returns rows
	Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error)
	
	// Begin starts a transaction and returns a transactional executor
	Begin(ctx context.Context) (Executor, error)
	
	// Commit commits the current transaction
	Commit(ctx context.Context) error
	
	// Rollback rolls back the current transaction
	Rollback(ctx context.Context) error
}

// Repository manages migration state persistence.
type Repository interface {
	// GetAppliedMigrations returns all migrations that have been applied
	GetAppliedMigrations(ctx context.Context) ([]Migration, error)
	
	// RecordMigration records that a migration was applied
	RecordMigration(ctx context.Context, migration Migration) error
	
	// RemoveMigration removes a migration record (for rollback)
	RemoveMigration(ctx context.Context, version string) error
	
	// InitializeSchema creates the migration tracking table if it doesn't exist
	InitializeSchema(ctx context.Context) error
}

// Manager orchestrates migration execution.
type Manager struct {
	repo       Repository
	executor   Executor
	migrations []Migration
}

// NewManager creates a new migration manager.
func NewManager(repo Repository, executor Executor) *Manager {
	return &Manager{
		repo:       repo,
		executor:   executor,
		migrations: make([]Migration, 0),
	}
}

// Register adds a migration to the manager.
func (m *Manager) Register(migration Migration) error {
	// Validate migration
	if migration.Version == "" {
		return fmt.Errorf("migration version cannot be empty")
	}
	if migration.Name == "" {
		return fmt.Errorf("migration name cannot be empty")
	}
	if migration.Up == nil {
		return fmt.Errorf("migration %s: Up function cannot be nil", migration.Version)
	}
	
	// Check for duplicate versions
	for _, existing := range m.migrations {
		if existing.Version == migration.Version {
			return fmt.Errorf("migration version %s already registered", migration.Version)
		}
	}
	
	m.migrations = append(m.migrations, migration)
	return nil
}

// RegisterMultiple adds multiple migrations at once.
func (m *Manager) RegisterMultiple(migrations []Migration) error {
	for _, migration := range migrations {
		if err := m.Register(migration); err != nil {
			return err
		}
	}
	return nil
}

// Up runs all pending migrations.
func (m *Manager) Up(ctx context.Context) error {
	// Initialize migration tracking schema
	if err := m.repo.InitializeSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize migration schema: %w", err)
	}
	
	// Get applied migrations
	applied, err := m.repo.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	// Build set of applied versions
	appliedVersions := make(map[string]bool)
	for _, m := range applied {
		appliedVersions[m.Version] = true
	}
	
	// Sort migrations by version
	pending := m.getPendingMigrations(appliedVersions)
	
	if len(pending) == 0 {
		return nil
	}
	
	// Execute pending migrations
	for _, migration := range pending {
		if err := m.executeMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to execute migration %s (%s): %w", 
				migration.Version, migration.Name, err)
		}
	}
	
	return nil
}

// UpTo runs migrations up to (and including) a specific version.
func (m *Manager) UpTo(ctx context.Context, targetVersion string) error {
	if err := m.repo.InitializeSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize migration schema: %w", err)
	}
	
	applied, err := m.repo.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	appliedVersions := make(map[string]bool)
	for _, m := range applied {
		appliedVersions[m.Version] = true
	}
	
	pending := m.getPendingMigrations(appliedVersions)
	
	// Execute migrations up to target version
	for _, migration := range pending {
		if err := m.executeMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to execute migration %s (%s): %w", 
				migration.Version, migration.Name, err)
		}
		
		// Stop if we've reached the target version
		if migration.Version == targetVersion {
			break
		}
	}
	
	return nil
}

// Down rolls back the most recent migration.
func (m *Manager) Down(ctx context.Context) error {
	applied, err := m.repo.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	if len(applied) == 0 {
		return fmt.Errorf("no migrations to roll back")
	}
	
	// Sort by version descending to get the most recent
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version > applied[j].Version
	})
	
	mostRecent := applied[0]
	
	// Find the migration definition
	var migration *Migration
	for i := range m.migrations {
		if m.migrations[i].Version == mostRecent.Version {
			migration = &m.migrations[i]
			break
		}
	}
	
	if migration == nil {
		return fmt.Errorf("migration %s not found in registered migrations", mostRecent.Version)
	}
	
	if migration.Down == nil {
		return fmt.Errorf("migration %s has no Down function", migration.Version)
	}
	
	return m.rollbackMigration(ctx, *migration)
}

// DownTo rolls back migrations down to (but not including) a specific version.
func (m *Manager) DownTo(ctx context.Context, targetVersion string) error {
	applied, err := m.repo.GetAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	// Sort by version descending
	sort.Slice(applied, func(i, j int) bool {
		return applied[i].Version > applied[j].Version
	})
	
	// Roll back migrations until we reach the target
	for _, appliedMig := range applied {
		if appliedMig.Version == targetVersion {
			break
		}
		
		// Find the migration definition
		var migration *Migration
		for i := range m.migrations {
			if m.migrations[i].Version == appliedMig.Version {
				migration = &m.migrations[i]
				break
			}
		}
		
		if migration == nil {
			return fmt.Errorf("migration %s not found in registered migrations", appliedMig.Version)
		}
		
		if migration.Down == nil {
			return fmt.Errorf("migration %s has no Down function", migration.Version)
		}
		
		if err := m.rollbackMigration(ctx, *migration); err != nil {
			return err
		}
	}
	
	return nil
}

// Status returns the current migration status.
func (m *Manager) Status(ctx context.Context) (*Status, error) {
	if err := m.repo.InitializeSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize migration schema: %w", err)
	}
	
	applied, err := m.repo.GetAppliedMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	appliedVersions := make(map[string]bool)
	for _, m := range applied {
		appliedVersions[m.Version] = true
	}
	
	pending := m.getPendingMigrations(appliedVersions)
	
	return &Status{
		Applied: applied,
		Pending: pending,
	}, nil
}

// Status represents the current state of migrations.
type Status struct {
	Applied []Migration
	Pending []Migration
}

// executeMigration runs a single migration within a transaction.
func (m *Manager) executeMigration(ctx context.Context, migration Migration) error {
	// Start transaction
	tx, err := m.executor.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Execute Up function
	if err := migration.Up(ctx, tx); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("migration failed: %w", err)
	}
	
	// Record migration
	migration.AppliedAt = timePtr(time.Now())
	if err := m.repo.RecordMigration(ctx, migration); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to record migration: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// rollbackMigration reverts a single migration within a transaction.
func (m *Manager) rollbackMigration(ctx context.Context, migration Migration) error {
	// Start transaction
	tx, err := m.executor.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	// Execute Down function
	if err := migration.Down(ctx, tx); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("rollback failed: %w", err)
	}
	
	// Remove migration record
	if err := m.repo.RemoveMigration(ctx, migration.Version); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("failed to remove migration record: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// getPendingMigrations returns migrations that haven't been applied, sorted by version.
func (m *Manager) getPendingMigrations(appliedVersions map[string]bool) []Migration {
	pending := make([]Migration, 0)
	
	for _, migration := range m.migrations {
		if !appliedVersions[migration.Version] {
			pending = append(pending, migration)
		}
	}
	
	// Sort by version
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].Version < pending[j].Version
	})
	
	return pending
}

func timePtr(t time.Time) *time.Time {
	return &t
}
