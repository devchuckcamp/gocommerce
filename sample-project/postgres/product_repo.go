package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devchuckcamp/gocommerce/catalog"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) FindByID(ctx context.Context, id string) (*catalog.Product, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, sku, name, COALESCE(description,''), COALESCE(brand_id,''), COALESCE(category_id,''),
			base_price_amount, base_price_currency, status, COALESCE(images,'[]'), COALESCE(attributes,'{}'),
			created_at, updated_at
		FROM products
		WHERE id = $1
	`, id)

	var p catalog.Product
	var brandID, categoryID sql.NullString
	var amount int64
	var currency string
	var status string
	var imagesRaw, attrsRaw []byte
	var createdAt, updatedAt time.Time

	if err := row.Scan(
		&p.ID,
		&p.SKU,
		&p.Name,
		&p.Description,
		&brandID,
		&categoryID,
		&amount,
		&currency,
		&status,
		&imagesRaw,
		&attrsRaw,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}

	p.BrandID = scanNullString(brandID)
	p.CategoryID = scanNullString(categoryID)
	m, err := moneyFrom(amount, currency)
	if err != nil {
		return nil, err
	}
	p.BasePrice = m
	p.Status = catalog.ProductStatus(status)
	_ = fromJSONB(imagesRaw, &p.Images)
	_ = fromJSONB(attrsRaw, &p.Attributes)
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt
	return &p, nil
}

func (r *ProductRepository) FindBySKU(ctx context.Context, sku string) (*catalog.Product, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id FROM products WHERE sku = $1`, sku)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	return r.FindByID(ctx, id)
}

func (r *ProductRepository) FindByCategory(ctx context.Context, categoryID string, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	q := `SELECT id FROM products WHERE category_id = $1`
	args := []any{categoryID}
	q, args = applyProductFilter(q, args, filter)
	return r.listByQuery(ctx, q, args...)
}

func (r *ProductRepository) FindByBrand(ctx context.Context, brandID string, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	q := `SELECT id FROM products WHERE brand_id = $1`
	args := []any{brandID}
	q, args = applyProductFilter(q, args, filter)
	return r.listByQuery(ctx, q, args...)
}

func (r *ProductRepository) Search(ctx context.Context, query string, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	q := `SELECT id FROM products WHERE (name ILIKE $1 OR sku ILIKE $1)`
	args := []any{"%" + query + "%"}
	q, args = applyProductFilter(q, args, filter)
	return r.listByQuery(ctx, q, args...)
}

func (r *ProductRepository) Save(ctx context.Context, product *catalog.Product) error {
	if product == nil {
		return errors.New("product is nil")
	}
	if product.ID == "" {
		return errors.New("product ID is required")
	}
	if product.SKU == "" {
		return errors.New("product SKU is required")
	}

	images, err := toJSONB(product.Images)
	if err != nil {
		return err
	}
	attrs, err := toJSONB(product.Attributes)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO products (
			id, sku, name, description, brand_id, category_id,
			base_price_amount, base_price_currency, status, images, attributes,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,NULLIF($5,''),NULLIF($6,''),
			$7,$8,$9,$10,$11,
			COALESCE($12, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP
		)
		ON CONFLICT (id) DO UPDATE SET
			sku = EXCLUDED.sku,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			brand_id = EXCLUDED.brand_id,
			category_id = EXCLUDED.category_id,
			base_price_amount = EXCLUDED.base_price_amount,
			base_price_currency = EXCLUDED.base_price_currency,
			status = EXCLUDED.status,
			images = EXCLUDED.images,
			attributes = EXCLUDED.attributes,
			updated_at = CURRENT_TIMESTAMP
	`,
		product.ID,
		product.SKU,
		product.Name,
		product.Description,
		product.BrandID,
		product.CategoryID,
		product.BasePrice.Amount,
		product.BasePrice.Currency,
		string(product.Status),
		images,
		attrs,
		nullTime(product.CreatedAt),
	)
	return err
}

func (r *ProductRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (r *ProductRepository) ListProducts(ctx context.Context, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	q := `SELECT id FROM products WHERE 1=1`
	args := []any{}
	q, args = applyProductFilter(q, args, filter)
	return r.listByQuery(ctx, q, args...)
}

func (r *ProductRepository) listByQuery(ctx context.Context, q string, args ...any) ([]*catalog.Product, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
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

	products := make([]*catalog.Product, 0, len(ids))
	for _, id := range ids {
		p, err := r.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}

func applyProductFilter(base string, args []any, filter catalog.ProductFilter) (string, []any) {
	q := base

	if filter.Status != nil {
		args = append(args, string(*filter.Status))
		q += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if filter.MinPrice != nil {
		args = append(args, *filter.MinPrice)
		q += fmt.Sprintf(" AND base_price_amount >= $%d", len(args))
	}
	if filter.MaxPrice != nil {
		args = append(args, *filter.MaxPrice)
		q += fmt.Sprintf(" AND base_price_amount <= $%d", len(args))
	}

	// Sorting (keep it minimal and safe)
	switch strings.ToLower(filter.SortBy) {
	case "price_asc":
		q += " ORDER BY base_price_amount ASC"
	case "price_desc":
		q += " ORDER BY base_price_amount DESC"
	case "name":
		q += " ORDER BY name ASC"
	case "created_at_desc":
		q += " ORDER BY created_at DESC"
	default:
		q += " ORDER BY created_at DESC"
	}

	if filter.Limit > 0 {
		args = append(args, filter.Limit)
		q += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if filter.Offset > 0 {
		args = append(args, filter.Offset)
		q += fmt.Sprintf(" OFFSET $%d", len(args))
	}
	return q, args
}
