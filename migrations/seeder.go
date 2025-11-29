package migrations

import (
	"context"
	"fmt"
)

// Seeder represents a database seeder that can populate tables with test/mock data.
type Seeder struct {
	executor Executor
	seeds    []Seed
}

// Seed represents a single seeding operation.
type Seed struct {
	Name        string
	Description string
	Run         func(ctx context.Context, exec Executor) error
}

// NewSeeder creates a new seeder instance.
func NewSeeder(executor Executor) *Seeder {
	return &Seeder{
		executor: executor,
		seeds:    make([]Seed, 0),
	}
}

// Register adds a seed to the seeder.
func (s *Seeder) Register(seed Seed) {
	s.seeds = append(s.seeds, seed)
}

// RegisterMultiple adds multiple seeds to the seeder.
func (s *Seeder) RegisterMultiple(seeds []Seed) {
	s.seeds = append(s.seeds, seeds...)
}

// Run executes all registered seeds.
func (s *Seeder) Run(ctx context.Context) error {
	for _, seed := range s.seeds {
		if err := s.runSeed(ctx, seed); err != nil {
			return fmt.Errorf("seed '%s' failed: %w", seed.Name, err)
		}
	}
	return nil
}

// RunSingle executes a specific seed by name.
func (s *Seeder) RunSingle(ctx context.Context, seedName string) error {
	for _, seed := range s.seeds {
		if seed.Name == seedName {
			return s.runSeed(ctx, seed)
		}
	}
	return fmt.Errorf("seed not found: %s", seedName)
}

// runSeed executes a single seed within a transaction.
func (s *Seeder) runSeed(ctx context.Context, seed Seed) error {
	// Start transaction
	txExec, err := s.executor.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Run seed
	if err := seed.Run(ctx, txExec); err != nil {
		if rbErr := txExec.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("seed failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := txExec.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// List returns all registered seeds.
func (s *Seeder) List() []Seed {
	return s.seeds
}
