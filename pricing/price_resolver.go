package pricing

import (
	"context"
	"time"

	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/money"
)

// PriceResolver resolves the effective price for products/variants.
type PriceResolver interface {
	// GetEffectivePrice returns the price for a product/variant at the given time.
	// If at is nil, uses current time.
	// Returns the resolved price, the ProductPrice record (if from pricing table), and any error.
	GetEffectivePrice(ctx context.Context, product *catalog.Product, variant *catalog.Variant, at *time.Time) (money.Money, *ProductPrice, error)

	// GetEffectivePriceByID returns the price for a product ID at the given time.
	// Falls back to base price if no specific pricing rule exists.
	GetEffectivePriceByID(ctx context.Context, productID string, variantID *string, at *time.Time) (money.Money, *ProductPrice, error)

	// GetEffectivePrices returns prices for multiple products (batch optimization).
	GetEffectivePrices(ctx context.Context, productIDs []string, at *time.Time) (map[string]money.Money, error)
}

// PriceResolverService implements PriceResolver.
type PriceResolverService struct {
	productPriceRepo ProductPriceRepository
	productRepo      catalog.ProductRepository
	variantRepo      catalog.VariantRepository
}

// NewPriceResolverService creates a new price resolver.
func NewPriceResolverService(
	productPriceRepo ProductPriceRepository,
	productRepo catalog.ProductRepository,
	variantRepo catalog.VariantRepository,
) *PriceResolverService {
	return &PriceResolverService{
		productPriceRepo: productPriceRepo,
		productRepo:      productRepo,
		variantRepo:      variantRepo,
	}
}

// GetEffectivePrice resolves the price for a product/variant.
// Priority: 1) Date-range price, 2) Default price from pricing table, 3) Product/Variant BasePrice
func (s *PriceResolverService) GetEffectivePrice(
	ctx context.Context,
	product *catalog.Product,
	variant *catalog.Variant,
	at *time.Time,
) (money.Money, *ProductPrice, error) {
	resolveTime := time.Now()
	if at != nil {
		resolveTime = *at
	}

	var variantID *string
	if variant != nil {
		variantID = &variant.ID
	}

	// Try to find an effective price from the pricing table
	effectivePrice, err := s.productPriceRepo.FindEffectivePrice(ctx, product.ID, variantID, resolveTime)
	if err != nil {
		return money.Money{}, nil, err
	}

	// If found a specific price, use it
	if effectivePrice != nil {
		return effectivePrice.Price, effectivePrice, nil
	}

	// Fall back to base price from product/variant
	return product.GetEffectivePrice(variant), nil, nil
}

// GetEffectivePriceByID resolves the price for a product by ID.
func (s *PriceResolverService) GetEffectivePriceByID(
	ctx context.Context,
	productID string,
	variantID *string,
	at *time.Time,
) (money.Money, *ProductPrice, error) {
	resolveTime := time.Now()
	if at != nil {
		resolveTime = *at
	}

	// Try to find an effective price from the pricing table
	effectivePrice, err := s.productPriceRepo.FindEffectivePrice(ctx, productID, variantID, resolveTime)
	if err != nil {
		return money.Money{}, nil, err
	}

	// If found a specific price, use it
	if effectivePrice != nil {
		return effectivePrice.Price, effectivePrice, nil
	}

	// Fall back to base price from product/variant
	if variantID != nil {
		variant, err := s.variantRepo.FindByID(ctx, *variantID)
		if err != nil {
			return money.Money{}, nil, err
		}
		if !variant.Price.IsZero() {
			return variant.Price, nil, nil
		}
	}

	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return money.Money{}, nil, err
	}

	return product.BasePrice, nil, nil
}

// GetEffectivePrices resolves prices for multiple products.
func (s *PriceResolverService) GetEffectivePrices(
	ctx context.Context,
	productIDs []string,
	at *time.Time,
) (map[string]money.Money, error) {
	resolveTime := time.Now()
	if at != nil {
		resolveTime = *at
	}

	// Get prices from pricing table
	pricingTablePrices, err := s.productPriceRepo.FindEffectivePrices(ctx, productIDs, resolveTime)
	if err != nil {
		return nil, err
	}

	result := make(map[string]money.Money)

	for _, productID := range productIDs {
		if price, ok := pricingTablePrices[productID]; ok {
			result[productID] = price.Price
		} else {
			// Fall back to base price
			product, err := s.productRepo.FindByID(ctx, productID)
			if err != nil {
				return nil, err
			}
			result[productID] = product.BasePrice
		}
	}

	return result, nil
}

// CartPriceResolverAdapter adapts PriceResolverService to work with cart.PriceResolver interface.
// This adapter is necessary because cart.PriceResolver has a simpler return signature
// to avoid import cycles between cart and pricing packages.
type CartPriceResolverAdapter struct {
	resolver *PriceResolverService
}

// NewCartPriceResolverAdapter creates a new adapter that implements cart.PriceResolver.
func NewCartPriceResolverAdapter(resolver *PriceResolverService) *CartPriceResolverAdapter {
	return &CartPriceResolverAdapter{resolver: resolver}
}

// GetEffectivePrice implements the cart.PriceResolver interface.
func (a *CartPriceResolverAdapter) GetEffectivePrice(
	ctx context.Context,
	product *catalog.Product,
	variant *catalog.Variant,
	at *time.Time,
) (money.Money, error) {
	price, _, err := a.resolver.GetEffectivePrice(ctx, product, variant, at)
	return price, err
}
