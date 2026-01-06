package postgres

import (
	"context"
	"database/sql"

	"github.com/devchuckcamp/gocommerce/migrations"
)

func RunMigrations(ctx context.Context, db *sql.DB) error {
	exec := NewExecutor(db)
	repo := migrations.NewPostgreSQLRepository(exec, migrations.TableName)
	mgr := migrations.NewManager(repo, exec)
	if err := mgr.RegisterMultiple(migrations.PostgreSQLExampleMigrations); err != nil {
		return err
	}
	return mgr.Up(ctx)
}
