package catalog

import (
	"context"
)

// ProductRepository defines methods for product persistence.
type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*Product, error)
	FindBySKU(ctx context.Context, sku string) (*Product, error)
	FindByCategory(ctx context.Context, categoryID string, filter ProductFilter) ([]*Product, error)
	FindByBrand(ctx context.Context, brandID string, filter ProductFilter) ([]*Product, error)
	Search(ctx context.Context, query string, filter ProductFilter) ([]*Product, error)
	Save(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id string) error
}

// VariantRepository defines methods for variant persistence.
type VariantRepository interface {
	FindByID(ctx context.Context, id string) (*Variant, error)
	FindBySKU(ctx context.Context, sku string) (*Variant, error)
	FindByProductID(ctx context.Context, productID string) ([]*Variant, error)
	Save(ctx context.Context, variant *Variant) error
	Delete(ctx context.Context, id string) error
}

// CategoryRepository defines methods for category persistence.
type CategoryRepository interface {
	FindByID(ctx context.Context, id string) (*Category, error)
	FindBySlug(ctx context.Context, slug string) (*Category, error)
	FindChildren(ctx context.Context, parentID string) ([]*Category, error)
	FindRoots(ctx context.Context) ([]*Category, error)
	FindAll(ctx context.Context) ([]*Category, error)
	Save(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id string) error
}

// BrandRepository defines methods for brand persistence.
type BrandRepository interface {
	FindByID(ctx context.Context, id string) (*Brand, error)
	FindBySlug(ctx context.Context, slug string) (*Brand, error)
	FindAll(ctx context.Context) ([]*Brand, error)
	Save(ctx context.Context, brand *Brand) error
	Delete(ctx context.Context, id string) error
}

// ProductFilter defines query filters for products.
type ProductFilter struct {
	Status       *ProductStatus
	MinPrice     *int64 // in cents
	MaxPrice     *int64
	BrandIDs     []string
	CategoryIDs  []string
	Attributes   map[string]string
	IsAvailable  *bool
	Limit        int
	Offset       int
	SortBy       string // e.g., "price_asc", "name", "created_at_desc"
}
