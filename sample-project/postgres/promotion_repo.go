package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/devchuckcamp/gocommerce/pricing"
)

type PromotionRepository struct {
	db *sql.DB
}

func NewPromotionRepository(db *sql.DB) *PromotionRepository {
	return &PromotionRepository{db: db}
}

func (r *PromotionRepository) FindByCode(ctx context.Context, code string) (*pricing.Promotion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, name, COALESCE(description,''), discount_type, value,
			min_purchase_amount, min_purchase_currency,
			max_discount_amount, max_discount_currency,
			COALESCE(valid_from, CURRENT_TIMESTAMP), COALESCE(valid_to, CURRENT_TIMESTAMP),
			is_active, usage_limit, usage_count,
			COALESCE(applicable_product_ids, '[]'::jsonb),
			COALESCE(applicable_category_ids, '[]'::jsonb),
			COALESCE(excluded_product_ids, '[]'::jsonb)
		FROM promotions
		WHERE code = $1
	`, code)

	var p pricing.Promotion
	var discountType string
	var minAmount, maxAmount sql.NullInt64
	var minCur, maxCur sql.NullString
	var validFrom, validTo time.Time
	var applicableProducts, applicableCategories, excludedProducts []byte

	if err := row.Scan(
		&p.ID,
		&p.Code,
		&p.Name,
		&p.Description,
		&discountType,
		&p.Value,
		&minAmount,
		&minCur,
		&maxAmount,
		&maxCur,
		&validFrom,
		&validTo,
		&p.IsActive,
		&p.UsageLimit,
		&p.UsageCount,
		&applicableProducts,
		&applicableCategories,
		&excludedProducts,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("promotion not found")
		}
		return nil, err
	}

	p.DiscountType = pricing.DiscountType(discountType)
	p.ValidFrom = validFrom
	p.ValidTo = validTo

	if minAmount.Valid {
		m, err := moneyFrom(minAmount.Int64, scanNullString(minCur))
		if err == nil {
			p.MinPurchase = &m
		}
	}
	if maxAmount.Valid {
		m, err := moneyFrom(maxAmount.Int64, scanNullString(maxCur))
		if err == nil {
			p.MaxDiscount = &m
		}
	}

	_ = fromJSONB(applicableProducts, &p.ApplicableProductIDs)
	_ = fromJSONB(applicableCategories, &p.ApplicableCategoryIDs)
	_ = fromJSONB(excludedProducts, &p.ExcludedProductIDs)

	return &p, nil
}

func (r *PromotionRepository) FindActive(ctx context.Context) ([]*pricing.Promotion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT code FROM promotions WHERE is_active = true`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	promos := make([]*pricing.Promotion, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		p, err := r.FindByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (r *PromotionRepository) Save(ctx context.Context, p *pricing.Promotion) error {
	if p == nil {
		return errors.New("promotion is nil")
	}

	applicableProducts, err := toJSONB(p.ApplicableProductIDs)
	if err != nil {
		return err
	}
	applicableCategories, err := toJSONB(p.ApplicableCategoryIDs)
	if err != nil {
		return err
	}
	excludedProducts, err := toJSONB(p.ExcludedProductIDs)
	if err != nil {
		return err
	}

	var minAmt any
	var minCur any
	if p.MinPurchase != nil {
		minAmt = p.MinPurchase.Amount
		minCur = p.MinPurchase.Currency
	}
	var maxAmt any
	var maxCur any
	if p.MaxDiscount != nil {
		maxAmt = p.MaxDiscount.Amount
		maxCur = p.MaxDiscount.Currency
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO promotions (
			id, code, name, description, discount_type, value,
			min_purchase_amount, min_purchase_currency,
			max_discount_amount, max_discount_currency,
			valid_from, valid_to, is_active, usage_limit, usage_count,
			applicable_product_ids, applicable_category_ids, excluded_product_ids,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,
			$11,$12,$13,$14,$15,
			$16,$17,$18,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON CONFLICT (id) DO UPDATE SET
			code = EXCLUDED.code,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			discount_type = EXCLUDED.discount_type,
			value = EXCLUDED.value,
			min_purchase_amount = EXCLUDED.min_purchase_amount,
			min_purchase_currency = EXCLUDED.min_purchase_currency,
			max_discount_amount = EXCLUDED.max_discount_amount,
			max_discount_currency = EXCLUDED.max_discount_currency,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to,
			is_active = EXCLUDED.is_active,
			usage_limit = EXCLUDED.usage_limit,
			usage_count = EXCLUDED.usage_count,
			applicable_product_ids = EXCLUDED.applicable_product_ids,
			applicable_category_ids = EXCLUDED.applicable_category_ids,
			excluded_product_ids = EXCLUDED.excluded_product_ids,
			updated_at = CURRENT_TIMESTAMP
	`,
		p.ID,
		p.Code,
		p.Name,
		p.Description,
		string(p.DiscountType),
		p.Value,
		minAmt,
		minCur,
		maxAmt,
		maxCur,
		p.ValidFrom,
		p.ValidTo,
		p.IsActive,
		p.UsageLimit,
		p.UsageCount,
		applicableProducts,
		applicableCategories,
		excludedProducts,
	)
	return err
}
