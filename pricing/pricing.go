package pricing

import (
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

// LineItem represents an item to be priced.
type LineItem struct {
	ID         string
	ProductID  string
	VariantID  *string
	SKU        string
	Name       string
	UnitPrice  money.Money
	Quantity   int
	Attributes map[string]string
}

// PricingResult contains the complete pricing breakdown.
type PricingResult struct {
	Subtotal       money.Money
	DiscountTotal  money.Money
	TaxTotal       money.Money
	ShippingTotal  money.Money
	Total          money.Money
	LineItemPrices []LineItemPrice
	AppliedDiscounts []AppliedDiscount
	TaxLines       []TaxLine
	Currency       string
	CalculatedAt   time.Time
}

// LineItemPrice contains pricing details for a single line item.
type LineItemPrice struct {
	LineItemID     string
	Subtotal       money.Money
	DiscountAmount money.Money
	TaxAmount      money.Money
	Total          money.Money
}

// AppliedDiscount represents a discount that was applied.
type AppliedDiscount struct {
	PromotionID   string
	Code          string
	Name          string
	DiscountType  DiscountType
	Amount        money.Money
	AppliedToItems []string // Line item IDs
}

// TaxLine represents a tax calculation.
type TaxLine struct {
	Name       string
	Rate       float64 // e.g., 0.08 for 8%
	Amount     money.Money
	Jurisdiction string // e.g., "CA", "NY"
}

// DiscountType defines types of discounts.
type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "percentage"
	DiscountTypeFixedAmount DiscountType = "fixed_amount"
	DiscountTypeBuyXGetY    DiscountType = "buy_x_get_y"
	DiscountTypeFreeShipping DiscountType = "free_shipping"
)

// Promotion represents a discount promotion.
type Promotion struct {
	ID           string
	Code         string
	Name         string
	Description  string
	DiscountType DiscountType
	Value        float64 // percentage (0.10 = 10%) or fixed amount in cents
	MinPurchase  *money.Money
	MaxDiscount  *money.Money
	ValidFrom    time.Time
	ValidTo      time.Time
	IsActive     bool
	UsageLimit   int
	UsageCount   int
	// Additional rules
	ApplicableProductIDs  []string
	ApplicableCategoryIDs []string
	ExcludedProductIDs    []string
}

// IsValid checks if a promotion can be used.
func (p *Promotion) IsValid(at time.Time) bool {
	if !p.IsActive {
		return false
	}
	if at.Before(p.ValidFrom) || at.After(p.ValidTo) {
		return false
	}
	if p.UsageLimit > 0 && p.UsageCount >= p.UsageLimit {
		return false
	}
	return true
}

// CanApplyToProduct checks if promotion applies to a product.
func (p *Promotion) CanApplyToProduct(productID string) bool {
	// Check exclusions
	for _, excludedID := range p.ExcludedProductIDs {
		if excludedID == productID {
			return false
		}
	}
	
	// If specific products listed, check if product is in list
	if len(p.ApplicableProductIDs) > 0 {
		for _, applicableID := range p.ApplicableProductIDs {
			if applicableID == productID {
				return true
			}
		}
		return false
	}
	
	// If no specific products, applies to all
	return true
}
