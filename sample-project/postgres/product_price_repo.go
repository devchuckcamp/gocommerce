package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/devchuckcamp/gocommerce/pricing"
)

// ProductPriceRepository implements pricing.ProductPriceRepository for PostgreSQL.
type ProductPriceRepository struct {
	db *sql.DB
}

// NewProductPriceRepository creates a new ProductPriceRepository.
func NewProductPriceRepository(db *sql.DB) *ProductPriceRepository {
	return &ProductPriceRepository{db: db}
}

// FindByID returns a price by its ID.
func (r *ProductPriceRepository) FindByID(ctx context.Context, id string) (*pricing.ProductPrice, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, variant_id, price_amount, price_currency,
			   valid_from, valid_to, priority, price_type, is_active,
			   created_at, updated_at
		FROM product_prices
		WHERE id = $1
	`, id)

	return r.scanProductPrice(row)
}

// FindActiveForProduct returns all active prices for a product.
func (r *ProductPriceRepository) FindActiveForProduct(ctx context.Context, productID string) ([]*pricing.ProductPrice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, variant_id, price_amount, price_currency,
			   valid_from, valid_to, priority, price_type, is_active,
			   created_at, updated_at
		FROM product_prices
		WHERE product_id = $1 AND variant_id IS NULL AND is_active = true
		ORDER BY priority DESC, valid_from DESC NULLS LAST
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProductPrices(rows)
}

// FindActiveForVariant returns all active prices for a variant.
func (r *ProductPriceRepository) FindActiveForVariant(ctx context.Context, variantID string) ([]*pricing.ProductPrice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, variant_id, price_amount, price_currency,
			   valid_from, valid_to, priority, price_type, is_active,
			   created_at, updated_at
		FROM product_prices
		WHERE variant_id = $1 AND is_active = true
		ORDER BY priority DESC, valid_from DESC NULLS LAST
	`, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanProductPrices(rows)
}

// FindEffectivePrice returns the effective price for a product/variant at a given time.
// Returns nil if no specific price found (caller should use BasePrice).
// Priority order:
// 1. Date-bounded prices (valid_from AND valid_to both set) matching the time
// 2. Semi-bounded prices (only valid_from OR valid_to set) matching the time
// 3. Unbounded/default prices (both valid_from AND valid_to are NULL)
// Within each category, higher priority wins, then most recently created.
func (r *ProductPriceRepository) FindEffectivePrice(
	ctx context.Context,
	productID string,
	variantID *string,
	at time.Time,
) (*pricing.ProductPrice, error) {
	// If variant is specified, first try variant-specific price
	if variantID != nil {
		row := r.db.QueryRowContext(ctx, `
			SELECT id, product_id, variant_id, price_amount, price_currency,
				   valid_from, valid_to, priority, price_type, is_active,
				   created_at, updated_at
			FROM product_prices
			WHERE variant_id = $1
			  AND is_active = true
			  AND (valid_from IS NULL OR valid_from <= $2)
			  AND (valid_to IS NULL OR valid_to >= $2)
			ORDER BY
				-- Prioritize date-bounded prices over unbounded
				CASE
					WHEN valid_from IS NOT NULL AND valid_to IS NOT NULL THEN 0
					WHEN valid_from IS NOT NULL OR valid_to IS NOT NULL THEN 1
					ELSE 2
				END,
				priority DESC,
				created_at DESC
			LIMIT 1
		`, *variantID, at)

		price, err := r.scanProductPrice(row)
		if err == nil && price != nil {
			return price, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		// If no variant price, fall through to product price
	}

	// Try product-level price
	row := r.db.QueryRowContext(ctx, `
		SELECT id, product_id, variant_id, price_amount, price_currency,
			   valid_from, valid_to, priority, price_type, is_active,
			   created_at, updated_at
		FROM product_prices
		WHERE product_id = $1
		  AND variant_id IS NULL
		  AND is_active = true
		  AND (valid_from IS NULL OR valid_from <= $2)
		  AND (valid_to IS NULL OR valid_to >= $2)
		ORDER BY
			CASE
				WHEN valid_from IS NOT NULL AND valid_to IS NOT NULL THEN 0
				WHEN valid_from IS NOT NULL OR valid_to IS NOT NULL THEN 1
				ELSE 2
			END,
			priority DESC,
			created_at DESC
		LIMIT 1
	`, productID, at)

	price, err := r.scanProductPrice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No price found, caller should use BasePrice
		}
		return nil, err
	}
	return price, nil
}

// FindEffectivePrices returns effective prices for multiple products at once.
func (r *ProductPriceRepository) FindEffectivePrices(
	ctx context.Context,
	productIDs []string,
	at time.Time,
) (map[string]*pricing.ProductPrice, error) {
	result := make(map[string]*pricing.ProductPrice)

	// For simplicity, iterate. In production, use a more efficient batch query with DISTINCT ON.
	for _, productID := range productIDs {
		price, err := r.FindEffectivePrice(ctx, productID, nil, at)
		if err != nil {
			return nil, err
		}
		if price != nil {
			result[productID] = price
		}
	}

	return result, nil
}

// Save creates or updates a product price.
func (r *ProductPriceRepository) Save(ctx context.Context, p *pricing.ProductPrice) error {
	if p == nil {
		return errors.New("product price is nil")
	}
	if p.ID == "" {
		return errors.New("product price ID is required")
	}
	if p.ProductID == "" {
		return errors.New("product ID is required")
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO product_prices (
			id, product_id, variant_id, price_amount, price_currency,
			valid_from, valid_to, priority, price_type, is_active,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			COALESCE($11, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP
		)
		ON CONFLICT (id) DO UPDATE SET
			product_id = EXCLUDED.product_id,
			variant_id = EXCLUDED.variant_id,
			price_amount = EXCLUDED.price_amount,
			price_currency = EXCLUDED.price_currency,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to,
			priority = EXCLUDED.priority,
			price_type = EXCLUDED.price_type,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
	`,
		p.ID,
		p.ProductID,
		p.VariantID,
		p.Price.Amount,
		p.Price.Currency,
		p.ValidFrom,
		p.ValidTo,
		p.Priority,
		string(p.PriceType),
		p.IsActive,
		nullTime(p.CreatedAt),
	)
	return err
}

// Delete removes a product price.
func (r *ProductPriceRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM product_prices WHERE id = $1`, id)
	return err
}

// DeleteByProductID removes all prices for a product.
func (r *ProductPriceRepository) DeleteByProductID(ctx context.Context, productID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
	return err
}

// scanProductPrice scans a single product price from a row.
func (r *ProductPriceRepository) scanProductPrice(row *sql.Row) (*pricing.ProductPrice, error) {
	var p pricing.ProductPrice
	var variantID sql.NullString
	var amount int64
	var currency string
	var validFrom, validTo sql.NullTime
	var priceType string
	var createdAt, updatedAt time.Time

	if err := row.Scan(
		&p.ID,
		&p.ProductID,
		&variantID,
		&amount,
		&currency,
		&validFrom,
		&validTo,
		&p.Priority,
		&priceType,
		&p.IsActive,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	if variantID.Valid {
		p.VariantID = &variantID.String
	}

	m, err := moneyFrom(amount, currency)
	if err != nil {
		return nil, err
	}
	p.Price = m

	p.ValidFrom = scanNullTime(validFrom)
	p.ValidTo = scanNullTime(validTo)
	p.PriceType = pricing.PriceType(priceType)
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	return &p, nil
}

// scanProductPrices scans multiple product prices from rows.
func (r *ProductPriceRepository) scanProductPrices(rows *sql.Rows) ([]*pricing.ProductPrice, error) {
	prices := make([]*pricing.ProductPrice, 0)
	for rows.Next() {
		var p pricing.ProductPrice
		var variantID sql.NullString
		var amount int64
		var currency string
		var validFrom, validTo sql.NullTime
		var priceType string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&p.ID,
			&p.ProductID,
			&variantID,
			&amount,
			&currency,
			&validFrom,
			&validTo,
			&p.Priority,
			&priceType,
			&p.IsActive,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}

		if variantID.Valid {
			p.VariantID = &variantID.String
		}

		m, err := moneyFrom(amount, currency)
		if err != nil {
			return nil, err
		}
		p.Price = m

		p.ValidFrom = scanNullTime(validFrom)
		p.ValidTo = scanNullTime(validTo)
		p.PriceType = pricing.PriceType(priceType)
		p.CreatedAt = createdAt
		p.UpdatedAt = updatedAt

		prices = append(prices, &p)
	}
	return prices, rows.Err()
}
