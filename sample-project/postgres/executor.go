package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/devchuckcamp/gocommerce/migrations"
)

// Executor implements migrations.Executor for a PostgreSQL database.
type Executor struct {
	db *sql.DB
	tx *sql.Tx
}

func NewExecutor(db *sql.DB) *Executor {
	return &Executor{db: db}
}

func (e *Executor) Exec(ctx context.Context, query string, args ...interface{}) error {
	if e.tx != nil {
		_, err := e.tx.ExecContext(ctx, query, args...)
		return err
	}
	_, err := e.db.ExecContext(ctx, query, args...)
	return err
}

func (e *Executor) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
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

		rowMap := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		result = append(result, rowMap)
	}
	return result, rows.Err()
}

func (e *Executor) Begin(ctx context.Context) (migrations.Executor, error) {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Executor{db: e.db, tx: tx}, nil
}

func (e *Executor) Commit(ctx context.Context) error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to commit")
	}
	return e.tx.Commit()
}

func (e *Executor) Rollback(ctx context.Context) error {
	if e.tx == nil {
		return fmt.Errorf("no transaction to rollback")
	}
	return e.tx.Rollback()
}
