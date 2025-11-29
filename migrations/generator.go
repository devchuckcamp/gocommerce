package migrations

import (
	"fmt"
	"time"
)

// Generator helps create migration templates.
type Generator struct {
	prefix string
}

// NewGenerator creates a new migration generator.
// prefix is used for migration version naming (e.g., "v1", "20231128")
func NewGenerator(prefix string) *Generator {
	return &Generator{prefix: prefix}
}

// GenerateVersion creates a migration version string.
// Format: {prefix}_{timestamp}_{sequence}
func (g *Generator) GenerateVersion(sequence int) string {
	timestamp := time.Now().Format("20060102_150405")
	if g.prefix != "" {
		return fmt.Sprintf("%s_%s_%03d", g.prefix, timestamp, sequence)
	}
	return fmt.Sprintf("%s_%03d", timestamp, sequence)
}

// GenerateSimpleVersion creates a simple sequential version.
// Format: {prefix}_{sequence}
func (g *Generator) GenerateSimpleVersion(sequence int) string {
	if g.prefix != "" {
		return fmt.Sprintf("%s_%03d", g.prefix, sequence)
	}
	return fmt.Sprintf("%03d", sequence)
}

// NewMigration creates a migration template with generated version.
func (g *Generator) NewMigration(name string, sequence int, up, down MigrationFunc) Migration {
	return Migration{
		Version: g.GenerateVersion(sequence),
		Name:    name,
		Up:      up,
		Down:    down,
	}
}

// NewSimpleMigration creates a migration with simple version numbering.
func (g *Generator) NewSimpleMigration(name string, sequence int, up, down MigrationFunc) Migration {
	return Migration{
		Version: g.GenerateSimpleVersion(sequence),
		Name:    name,
		Up:      up,
		Down:    down,
	}
}
