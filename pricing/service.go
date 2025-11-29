package pricing

import (
	"context"
	"time"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/money"
	"github.com/devchuckcamp/gocommerce/shipping"
	"github.com/devchuckcamp/gocommerce/tax"
)

// Service defines the pricing service interface.
type Service interface {
	PriceCart(ctx context.Context, req PriceCartRequest) (*PricingResult, error)
	PriceLineItems(ctx context.Context, req PriceLineItemsRequest) (*PricingResult, error)
	ValidatePromotion(ctx context.Context, code string, cartTotal money.Money) (*Promotion, error)
}

// PriceCartRequest contains data needed to price a cart.
type PriceCartRequest struct {
	Cart             *cart.Cart
	PromotionCodes   []string
	ShippingMethodID *string
	ShippingAddress  *Address // For tax calculation
	TaxInclusive     bool
}

// PriceLineItemsRequest prices arbitrary line items.
type PriceLineItemsRequest struct {
	Items            []LineItem
	PromotionCodes   []string
	ShippingCost     *money.Money
	ShippingAddress  *Address
	TaxInclusive     bool
}

// Address represents a shipping/billing address (minimal for pricing).
type Address struct {
	Country     string
	State       string
	City        string
	PostalCode  string
}

// PromotionRepository defines methods for promotion persistence.
type PromotionRepository interface {
	FindByCode(ctx context.Context, code string) (*Promotion, error)
	FindActive(ctx context.Context) ([]*Promotion, error)
	Save(ctx context.Context, promotion *Promotion) error
}

// PricingService implements the Service interface.
type PricingService struct {
	promotionRepo    PromotionRepository
	taxCalculator    tax.Calculator
	shippingCalc     shipping.RateCalculator
}

// NewPricingService creates a new pricing service.
func NewPricingService(
	promotionRepo PromotionRepository,
	taxCalculator tax.Calculator,
	shippingCalc shipping.RateCalculator,
) *PricingService {
	return &PricingService{
		promotionRepo: promotionRepo,
		taxCalculator: taxCalculator,
		shippingCalc:  shippingCalc,
	}
}

// PriceCart calculates the complete pricing for a cart.
func (s *PricingService) PriceCart(ctx context.Context, req PriceCartRequest) (*PricingResult, error) {
	if req.Cart == nil || req.Cart.IsEmpty() {
		return nil, nil
	}
	
	// Convert cart items to line items
	lineItems := make([]LineItem, len(req.Cart.Items))
	for i, item := range req.Cart.Items {
		lineItems[i] = LineItem{
			ID:         item.ID,
			ProductID:  item.ProductID,
			VariantID:  item.VariantID,
			SKU:        item.SKU,
			Name:       item.Name,
			UnitPrice:  item.Price,
			Quantity:   item.Quantity,
			Attributes: item.Attributes,
		}
	}
	
	// Calculate subtotal
	currency := req.Cart.Items[0].Price.Currency
	subtotal := money.Zero(currency)
	lineItemPrices := make([]LineItemPrice, len(lineItems))
	
	for i, item := range lineItems {
		itemSubtotal := item.UnitPrice.MultiplyInt(item.Quantity)
		lineItemPrices[i] = LineItemPrice{
			LineItemID:     item.ID,
			Subtotal:       itemSubtotal,
			DiscountAmount: money.Zero(currency),
			TaxAmount:      money.Zero(currency),
			Total:          itemSubtotal,
		}
		subtotal, _ = subtotal.Add(itemSubtotal)
	}
	
	// Apply promotions
	appliedDiscounts, err := s.applyPromotions(ctx, lineItems, lineItemPrices, req.PromotionCodes)
	if err != nil {
		return nil, err
	}
	
	// Calculate total discount
	discountTotal := money.Zero(currency)
	for _, discount := range appliedDiscounts {
		discountTotal, _ = discountTotal.Add(discount.Amount)
	}
	
	// Calculate shipping
	shippingTotal := money.Zero(currency)
	if req.ShippingMethodID != nil && s.shippingCalc != nil {
		shippingRate, err := s.shippingCalc.GetRate(ctx, shipping.RateRequest{
			Items:            convertToShippingItems(lineItems),
			DestinationAddress: convertToShippingAddress(req.ShippingAddress),
			ShippingMethodID: *req.ShippingMethodID,
		})
		if err == nil && shippingRate != nil {
			shippingTotal = shippingRate.Cost
		}
	}
	
	// Calculate tax
	var taxLines []TaxLine
	taxTotal := money.Zero(currency)
	
	if req.ShippingAddress != nil && s.taxCalculator != nil {
		taxReq := tax.CalculationRequest{
			LineItems:       convertToTaxableItems(lineItems, lineItemPrices),
			ShippingCost:    shippingTotal,
			Address:         convertToTaxAddress(req.ShippingAddress),
			TaxInclusive:    req.TaxInclusive,
		}
		
		taxResult, err := s.taxCalculator.Calculate(ctx, taxReq)
		if err == nil {
			taxLines = convertTaxLines(taxResult)
			taxTotal = taxResult.TotalTax
			
			// Update line item tax amounts
			for i, taxLine := range taxResult.LineItemTaxes {
				if i < len(lineItemPrices) {
					lineItemPrices[i].TaxAmount = taxLine.TaxAmount
				}
			}
		}
	}
	
	// Calculate totals
	subtotalAfterDiscount, _ := subtotal.Subtract(discountTotal)
	total := subtotalAfterDiscount
	total, _ = total.Add(taxTotal)
	total, _ = total.Add(shippingTotal)
	
	// Update line item totals
	for i := range lineItemPrices {
		itemTotal := lineItemPrices[i].Subtotal
		itemTotal, _ = itemTotal.Subtract(lineItemPrices[i].DiscountAmount)
		itemTotal, _ = itemTotal.Add(lineItemPrices[i].TaxAmount)
		lineItemPrices[i].Total = itemTotal
	}
	
	return &PricingResult{
		Subtotal:         subtotal,
		DiscountTotal:    discountTotal,
		TaxTotal:         taxTotal,
		ShippingTotal:    shippingTotal,
		Total:            total,
		LineItemPrices:   lineItemPrices,
		AppliedDiscounts: appliedDiscounts,
		TaxLines:         taxLines,
		Currency:         currency,
		CalculatedAt:     time.Now(),
	}, nil
}

// PriceLineItems prices arbitrary line items.
func (s *PricingService) PriceLineItems(ctx context.Context, req PriceLineItemsRequest) (*PricingResult, error) {
	// Similar to PriceCart but works with generic line items
	// Implementation follows same pattern as PriceCart
	return s.PriceCart(ctx, PriceCartRequest{
		Cart: &cart.Cart{
			Items: convertLineItemsToCartItems(req.Items),
		},
		PromotionCodes:   req.PromotionCodes,
		ShippingAddress:  req.ShippingAddress,
		TaxInclusive:     req.TaxInclusive,
	})
}

// ValidatePromotion validates a promotion code.
func (s *PricingService) ValidatePromotion(ctx context.Context, code string, cartTotal money.Money) (*Promotion, error) {
	promotion, err := s.promotionRepo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	
	if !promotion.IsValid(time.Now()) {
		return nil, ErrPromotionInvalid
	}
	
	if promotion.MinPurchase != nil {
		isGreater, _ := cartTotal.GreaterThan(*promotion.MinPurchase)
		if !isGreater {
			return nil, ErrMinPurchaseNotMet
		}
	}
	
	return promotion, nil
}

// applyPromotions applies promotions to line items.
func (s *PricingService) applyPromotions(
	ctx context.Context,
	lineItems []LineItem,
	lineItemPrices []LineItemPrice,
	codes []string,
) ([]AppliedDiscount, error) {
	appliedDiscounts := []AppliedDiscount{}
	
	for _, code := range codes {
		promotion, err := s.promotionRepo.FindByCode(ctx, code)
		if err != nil || !promotion.IsValid(time.Now()) {
			continue
		}
		
		discount := s.calculateDiscount(promotion, lineItems, lineItemPrices)
		if discount != nil {
			appliedDiscounts = append(appliedDiscounts, *discount)
		}
	}
	
	return appliedDiscounts, nil
}

// calculateDiscount calculates discount for a promotion.
func (s *PricingService) calculateDiscount(
	promotion *Promotion,
	lineItems []LineItem,
	lineItemPrices []LineItemPrice,
) *AppliedDiscount {
	if len(lineItems) == 0 {
		return nil
	}
	
	currency := lineItems[0].UnitPrice.Currency
	totalDiscount := money.Zero(currency)
	appliedToItems := []string{}
	
	for i, item := range lineItems {
		if !promotion.CanApplyToProduct(item.ProductID) {
			continue
		}
		
		var itemDiscount money.Money
		
		switch promotion.DiscountType {
		case DiscountTypePercentage:
			itemDiscount = lineItemPrices[i].Subtotal.Multiply(promotion.Value)
		case DiscountTypeFixedAmount:
			discountMoney, _ := money.New(int64(promotion.Value), currency)
			itemDiscount = discountMoney
		}
		
		// Apply max discount if set
		if promotion.MaxDiscount != nil {
			isGreater, _ := itemDiscount.GreaterThan(*promotion.MaxDiscount)
			if isGreater {
				itemDiscount = *promotion.MaxDiscount
			}
		}
		
		lineItemPrices[i].DiscountAmount, _ = lineItemPrices[i].DiscountAmount.Add(itemDiscount)
		totalDiscount, _ = totalDiscount.Add(itemDiscount)
		appliedToItems = append(appliedToItems, item.ID)
	}
	
	if totalDiscount.IsZero() {
		return nil
	}
	
	return &AppliedDiscount{
		PromotionID:    promotion.ID,
		Code:           promotion.Code,
		Name:           promotion.Name,
		DiscountType:   promotion.DiscountType,
		Amount:         totalDiscount,
		AppliedToItems: appliedToItems,
	}
}

// Helper conversion functions

func convertLineItemsToCartItems(items []LineItem) []cart.CartItem {
	cartItems := make([]cart.CartItem, len(items))
	for i, item := range items {
		cartItems[i] = cart.CartItem{
			ID:         item.ID,
			ProductID:  item.ProductID,
			VariantID:  item.VariantID,
			SKU:        item.SKU,
			Name:       item.Name,
			Price:      item.UnitPrice,
			Quantity:   item.Quantity,
			Attributes: item.Attributes,
		}
	}
	return cartItems
}

func convertToShippingItems(items []LineItem) []shipping.ShippableItem {
	// Stub - would convert to shipping items
	return []shipping.ShippableItem{}
}

func convertToShippingAddress(addr *Address) shipping.Address {
	if addr == nil {
		return shipping.Address{}
	}
	return shipping.Address{
		Country:    addr.Country,
		State:      addr.State,
		City:       addr.City,
		PostalCode: addr.PostalCode,
	}
}

func convertToTaxableItems(items []LineItem, prices []LineItemPrice) []tax.TaxableItem {
	taxItems := make([]tax.TaxableItem, len(items))
	for i, item := range items {
		taxItems[i] = tax.TaxableItem{
			ID:       item.ID,
			Amount:   prices[i].Subtotal,
			Quantity: item.Quantity,
		}
	}
	return taxItems
}

func convertToTaxAddress(addr *Address) tax.Address {
	if addr == nil {
		return tax.Address{}
	}
	return tax.Address{
		Country:    addr.Country,
		State:      addr.State,
		City:       addr.City,
		PostalCode: addr.PostalCode,
	}
}

func convertTaxLines(result *tax.CalculationResult) []TaxLine {
	lines := make([]TaxLine, len(result.TaxRates))
	for i, rate := range result.TaxRates {
		lines[i] = TaxLine{
			Name:         rate.Name,
			Rate:         rate.Rate,
			Amount:       rate.Amount,
			Jurisdiction: rate.Jurisdiction,
		}
	}
	return lines
}

var (
	ErrPromotionInvalid   = DiscountError{Message: "promotion code is invalid"}
	ErrMinPurchaseNotMet  = DiscountError{Message: "minimum purchase not met"}
)

type DiscountError struct {
	Message string
}

func (e DiscountError) Error() string {
	return e.Message
}
