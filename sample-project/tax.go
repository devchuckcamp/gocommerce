package main

import (
	"context"

	"github.com/devchuckcamp/gocommerce/money"
	"github.com/devchuckcamp/gocommerce/tax"
)

// SimpleTaxCalculator implements a basic tax calculator
type SimpleTaxCalculator struct {
	defaultRate float64
}

func NewSimpleTaxCalculator(rate float64) *SimpleTaxCalculator {
	return &SimpleTaxCalculator{defaultRate: rate}
}

func (c *SimpleTaxCalculator) Calculate(ctx context.Context, req tax.CalculationRequest) (*tax.CalculationResult, error) {
	// Calculate subtotal
	currency := "USD"
	if len(req.LineItems) > 0 {
		currency = req.LineItems[0].Amount.Currency
	}
	
	subtotal := money.Zero(currency)
	for _, item := range req.LineItems {
		if item.IsTaxable {
			itemTotal := item.Amount.MultiplyInt(item.Quantity)
			subtotal, _ = subtotal.Add(itemTotal)
		}
	}
	
	// Add shipping to taxable amount
	taxableAmount, _ := subtotal.Add(req.ShippingCost)
	
	// Calculate tax
	taxAmount := taxableAmount.Multiply(c.defaultRate)
	
	// Create line item taxes
	lineItemTaxes := make([]tax.LineItemTax, len(req.LineItems))
	for i, item := range req.LineItems {
		if item.IsTaxable {
			itemSubtotal := item.Amount.MultiplyInt(item.Quantity)
			itemTax := itemSubtotal.Multiply(c.defaultRate)
			lineItemTaxes[i] = tax.LineItemTax{
				LineItemID: item.ID,
				TaxAmount:  itemTax,
				TaxRates: []tax.AppliedTaxRate{
					{
						Name:         "Sales Tax",
						Rate:         c.defaultRate,
						Amount:       itemTax,
						Jurisdiction: req.Address.State,
						TaxType:      tax.TaxTypeSales,
					},
				},
			}
		} else {
			lineItemTaxes[i] = tax.LineItemTax{
				LineItemID: item.ID,
				TaxAmount:  money.Zero(currency),
			}
		}
	}
	
	return &tax.CalculationResult{
		TotalTax: taxAmount,
		TaxRates: []tax.AppliedTaxRate{
			{
				Name:         "Sales Tax",
				Rate:         c.defaultRate,
				Amount:       taxAmount,
				Jurisdiction: req.Address.State,
				TaxType:      tax.TaxTypeSales,
			},
		},
		LineItemTaxes: lineItemTaxes,
		ShippingTax:   req.ShippingCost.Multiply(c.defaultRate),
	}, nil
}

func (c *SimpleTaxCalculator) GetRatesForAddress(ctx context.Context, address tax.Address) ([]tax.TaxRate, error) {
	return []tax.TaxRate{
		{
			ID:           "rate-1",
			Name:         "Sales Tax",
			Rate:         c.defaultRate,
			Country:      address.Country,
			State:        address.State,
			TaxType:      tax.TaxTypeSales,
			IsCompound:   false,
			Priority:     1,
		},
	}, nil
}
