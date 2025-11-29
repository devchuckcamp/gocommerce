package tax

import (
	"context"

	"github.com/devchuckcamp/gocommerce/money"
)

// Calculator defines the tax calculator interface.
type Calculator interface {
	Calculate(ctx context.Context, req CalculationRequest) (*CalculationResult, error)
	GetRatesForAddress(ctx context.Context, address Address) ([]TaxRate, error)
}

// CalculationRequest contains data needed to calculate tax.
type CalculationRequest struct {
	LineItems    []TaxableItem
	ShippingCost money.Money
	Address      Address
	TaxInclusive bool // Whether prices already include tax
}

// TaxableItem represents an item subject to tax.
type TaxableItem struct {
	ID         string
	Amount     money.Money
	Quantity   int
	TaxCode    string // Optional product tax code
	IsTaxable  bool
}

// CalculationResult contains the tax calculation results.
type CalculationResult struct {
	TotalTax       money.Money
	TaxRates       []AppliedTaxRate
	LineItemTaxes  []LineItemTax
	ShippingTax    money.Money
}

// AppliedTaxRate represents a tax rate that was applied.
type AppliedTaxRate struct {
	Name         string
	Rate         float64 // e.g., 0.08 for 8%
	Amount       money.Money
	Jurisdiction string
	TaxType      TaxType
}

// LineItemTax contains tax for a specific line item.
type LineItemTax struct {
	LineItemID string
	TaxAmount  money.Money
	TaxRates   []AppliedTaxRate
}

// TaxType represents the type of tax.
type TaxType string

const (
	TaxTypeSales TaxType = "sales"
	TaxTypeVAT   TaxType = "vat"
	TaxTypeGST   TaxType = "gst"
	TaxTypeHST   TaxType = "hst"
)

// Address represents an address for tax calculation.
type Address struct {
	Country    string
	State      string
	City       string
	PostalCode string
	Street     string
}

// TaxRate represents a tax rate configuration.
type TaxRate struct {
	ID           string
	Name         string
	Rate         float64
	Country      string
	State        string
	City         string
	PostalCode   string
	TaxType      TaxType
	IsCompound   bool // Compound tax calculated on subtotal + other taxes
	Priority     int  // Order in which to apply (for compound taxes)
}

// AppliesTo checks if a tax rate applies to an address.
func (tr *TaxRate) AppliesTo(addr Address) bool {
	if tr.Country != "" && tr.Country != addr.Country {
		return false
	}
	if tr.State != "" && tr.State != addr.State {
		return false
	}
	if tr.City != "" && tr.City != addr.City {
		return false
	}
	if tr.PostalCode != "" && tr.PostalCode != addr.PostalCode {
		return false
	}
	return true
}

// Repository defines methods for tax data persistence.
type Repository interface {
	FindRatesByAddress(ctx context.Context, address Address) ([]*TaxRate, error)
	FindRateByID(ctx context.Context, id string) (*TaxRate, error)
	SaveRate(ctx context.Context, rate *TaxRate) error
	DeleteRate(ctx context.Context, id string) error
}
