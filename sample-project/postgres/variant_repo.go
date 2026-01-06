package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/devchuckcamp/gocommerce/catalog"
)

type VariantRepository struct {
	db *sql.DB
}

func NewVariantRepository(db *sql.DB) *VariantRepository {
	return &VariantRepository{db: db}
}

func (r *VariantRepository) FindByID(ctx context.Context, id string) (*catalog.Variant, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, sku, name, price_amount, price_currency,
			COALESCE(attributes, '{}'::jsonb), COALESCE(images, '[]'::jsonb),
			is_available, created_at, updated_at
		FROM variants
		WHERE id = $1
	`, id)

	var v catalog.Variant
	var amount int64
	var currency string
	var attrsRaw, imagesRaw []byte
	if err := row.Scan(
		&v.ID,
		&v.ProductID,
		&v.SKU,
		&v.Name,
		&amount,
		&currency,
		&attrsRaw,
		&imagesRaw,
		&v.IsAvailable,
		&v.CreatedAt,
		&v.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("variant not found")
		}
		return nil, err
	}

	m, err := moneyFrom(amount, currency)
	if err != nil {
		return nil, err
	}
	v.Price = m
	_ = fromJSONB(attrsRaw, &v.Attributes)
	_ = fromJSONB(imagesRaw, &v.Images)
	return &v, nil
}

func (r *VariantRepository) FindBySKU(ctx context.Context, sku string) (*catalog.Variant, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id FROM variants WHERE sku = $1`, sku)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("variant not found")
		}
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *VariantRepository) FindByProductID(ctx context.Context, productID string) ([]*catalog.Variant, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM variants WHERE product_id = $1 ORDER BY created_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]*catalog.Variant, 0, len(ids))
	for _, id := range ids {
		v, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *VariantRepository) Save(ctx context.Context, v *catalog.Variant) error {
	if v == nil {
		return errors.New("variant is nil")
	}
	attrs, err := toJSONB(v.Attributes)
	if err != nil {
		return err
	}
	images, err := toJSONB(v.Images)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO variants (
			id, product_id, sku, name, price_amount, price_currency,
			attributes, images, is_available, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9, COALESCE($10, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP
		)
		ON CONFLICT (id) DO UPDATE SET
			product_id = EXCLUDED.product_id,
			sku = EXCLUDED.sku,
			name = EXCLUDED.name,
			price_amount = EXCLUDED.price_amount,
			price_currency = EXCLUDED.price_currency,
			attributes = EXCLUDED.attributes,
			images = EXCLUDED.images,
			is_available = EXCLUDED.is_available,
			updated_at = CURRENT_TIMESTAMP
	`,
		v.ID,
		v.ProductID,
		v.SKU,
		v.Name,
		v.Price.Amount,
		v.Price.Currency,
		attrs,
		images,
		v.IsAvailable,
		nullTime(v.CreatedAt),
	)
	return err
}

func (r *VariantRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM variants WHERE id = $1`, id)
	return err
}
