package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/devchuckcamp/gocommerce/cart"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) FindByID(ctx context.Context, id string) (*cart.Cart, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(user_id,''), COALESCE(session_id,''), created_at, updated_at, expires_at
		FROM carts
		WHERE id = $1
	`, id)

	var c cart.Cart
	var expiresAt sql.NullTime
	if err := row.Scan(&c.ID, &c.UserID, &c.SessionID, &c.CreatedAt, &c.UpdatedAt, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cart.ErrCartNotFound
		}
		return nil, err
	}
	c.ExpiresAt = scanNullTime(expiresAt)

	items, err := r.findItems(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Items = items
	return &c, nil
}

func (r *CartRepository) FindByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id FROM carts WHERE user_id = $1 ORDER BY updated_at DESC LIMIT 1`, userID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cart.ErrCartNotFound
		}
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *CartRepository) FindBySessionID(ctx context.Context, sessionID string) (*cart.Cart, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id FROM carts WHERE session_id = $1 ORDER BY updated_at DESC LIMIT 1`, sessionID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, cart.ErrCartNotFound
		}
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *CartRepository) Save(ctx context.Context, c *cart.Cart) error {
	if c == nil {
		return errors.New("cart is nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO carts (id, user_id, session_id, created_at, updated_at, expires_at)
		VALUES ($1, NULLIF($2,''), NULLIF($3,''), COALESCE($4, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP, $5)
		ON CONFLICT (id) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			session_id = EXCLUDED.session_id,
			updated_at = CURRENT_TIMESTAMP,
			expires_at = EXCLUDED.expires_at
	`, c.ID, c.UserID, c.SessionID, nullTime(c.CreatedAt), c.ExpiresAt)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, c.ID)
	if err != nil {
		return err
	}

	for _, item := range c.Items {
		attrs, err := toJSONB(item.Attributes)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO cart_items (
				id, cart_id, product_id, variant_id, sku, name,
				price_amount, price_currency, quantity, added_at, attributes
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
			)
		`,
			item.ID,
			c.ID,
			item.ProductID,
			item.VariantID,
			item.SKU,
			item.Name,
			item.Price.Amount,
			item.Price.Currency,
			item.Quantity,
			nullTime(item.AddedAt),
			attrs,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *CartRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM carts WHERE id = $1`, id)
	return err
}

func (r *CartRepository) findItems(ctx context.Context, cartID string) ([]cart.CartItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, variant_id, sku, name, price_amount, price_currency, quantity, added_at, COALESCE(attributes,'{}')
		FROM cart_items
		WHERE cart_id = $1
		ORDER BY added_at ASC
	`, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]cart.CartItem, 0)
	for rows.Next() {
		var item cart.CartItem
		var variantID sql.NullString
		var amount int64
		var currency string
		var addedAt time.Time
		var attrsRaw []byte

		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&variantID,
			&item.SKU,
			&item.Name,
			&amount,
			&currency,
			&item.Quantity,
			&addedAt,
			&attrsRaw,
		); err != nil {
			return nil, err
		}
		if variantID.Valid {
			v := variantID.String
			item.VariantID = &v
		}
		m, err := moneyFrom(amount, currency)
		if err != nil {
			return nil, err
		}
		item.Price = m
		item.AddedAt = addedAt
		_ = fromJSONB(attrsRaw, &item.Attributes)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
