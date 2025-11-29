package catalog

import (
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

// Product represents a product in the catalog.
type Product struct {
	ID          string
	SKU         string
	Name        string
	Description string
	BrandID     string
	CategoryID  string
	BasePrice   money.Money
	Status      ProductStatus
	Images      []string
	Attributes  map[string]string // e.g., "material": "cotton"
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductStatus string

const (
	ProductStatusDraft       ProductStatus = "draft"
	ProductStatusActive      ProductStatus = "active"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

// Variant represents a product variant (size, color, etc.).
type Variant struct {
	ID          string
	ProductID   string
	SKU         string
	Name        string
	Price       money.Money
	Attributes  map[string]string // e.g., "size": "L", "color": "blue"
	Images      []string
	IsAvailable bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Category represents a product category.
type Category struct {
	ID          string
	ParentID    *string // nil for root categories
	Name        string
	Slug        string
	Description string
	ImageURL    string
	IsActive    bool
	DisplayOrder int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Brand represents a product brand.
type Brand struct {
	ID          string
	Name        string
	Slug        string
	Description string
	LogoURL     string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsActive returns true if the product is active.
func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}

// GetEffectivePrice returns the variant price if available, otherwise base price.
func (p *Product) GetEffectivePrice(variant *Variant) money.Money {
	if variant != nil && !variant.Price.IsZero() {
		return variant.Price
	}
	return p.BasePrice
}
