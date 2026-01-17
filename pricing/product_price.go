package pricing

import (
	"context"
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

// ProductPrice represents a price for a product/variant valid during a date range.
// If no date range matches, the product's BasePrice is used as fallback.
type ProductPrice struct {
	ID        string
	ProductID string
	VariantID *string     // nil means applies to product base price
	Price     money.Money
	ValidFrom *time.Time  // nil means "since beginning of time" (default price)
	ValidTo   *time.Time  // nil means "until end of time" (default price)
	Priority  int         // Higher priority wins when ranges overlap
	PriceType PriceType   // sale, clearance, regular, etc.
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PriceType categorizes the price (useful for display/reporting).
type PriceType string

const (
	PriceTypeRegular   PriceType = "regular"
	PriceTypeSale      PriceType = "sale"
	PriceTypeClearance PriceType = "clearance"
	PriceTypePromo     PriceType = "promotional"
)

// IsValidAt checks if this price is valid at a given point in time.
func (p *ProductPrice) IsValidAt(t time.Time) bool {
	if !p.IsActive {
		return false
	}
	if p.ValidFrom != nil && t.Before(*p.ValidFrom) {
		return false
	}
	if p.ValidTo != nil && t.After(*p.ValidTo) {
		return false
	}
	return true
}

// IsDefaultPrice returns true if this is a price with no date boundaries.
func (p *ProductPrice) IsDefaultPrice() bool {
	return p.ValidFrom == nil && p.ValidTo == nil
}

// IsDateBounded returns true if this price has specific date boundaries.
func (p *ProductPrice) IsDateBounded() bool {
	return p.ValidFrom != nil || p.ValidTo != nil
}

// ProductPriceRepository defines methods for product price persistence.
type ProductPriceRepository interface {
	// FindByID returns a price by its ID.
	FindByID(ctx context.Context, id string) (*ProductPrice, error)

	// FindActiveForProduct returns all active prices for a product.
	FindActiveForProduct(ctx context.Context, productID string) ([]*ProductPrice, error)

	// FindActiveForVariant returns all active prices for a variant.
	FindActiveForVariant(ctx context.Context, variantID string) ([]*ProductPrice, error)

	// FindEffectivePrice returns the effective price for a product/variant at a given time.
	// Returns nil if no specific price found (caller should use BasePrice).
	FindEffectivePrice(ctx context.Context, productID string, variantID *string, at time.Time) (*ProductPrice, error)

	// FindEffectivePrices returns effective prices for multiple products at once (batch operation).
	FindEffectivePrices(ctx context.Context, productIDs []string, at time.Time) (map[string]*ProductPrice, error)

	// Save creates or updates a product price.
	Save(ctx context.Context, price *ProductPrice) error

	// Delete removes a product price.
	Delete(ctx context.Context, id string) error

	// DeleteByProductID removes all prices for a product.
	DeleteByProductID(ctx context.Context, productID string) error
}
