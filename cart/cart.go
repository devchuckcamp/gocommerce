package cart

import (
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

// Cart represents a shopping cart.
type Cart struct {
	ID         string
	UserID     string    // Empty for guest carts
	SessionID  string    // For guest carts
	Items      []CartItem
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExpiresAt  *time.Time
}

// CartItem represents an item in the cart.
type CartItem struct {
	ID         string
	ProductID  string
	VariantID  *string // Optional variant
	SKU        string
	Name       string
	Price      money.Money // Price at time of adding
	Quantity   int
	Attributes map[string]string // Selected options
	AddedAt    time.Time
}

// AddItem adds an item to the cart or increases quantity if it already exists.
func (c *Cart) AddItem(item CartItem) {
	for i, existing := range c.Items {
		if existing.ProductID == item.ProductID && 
		   existing.VariantID == item.VariantID {
			c.Items[i].Quantity += item.Quantity
			c.UpdatedAt = time.Now()
			return
		}
	}
	c.Items = append(c.Items, item)
	c.UpdatedAt = time.Now()
}

// RemoveItem removes an item from the cart by ID.
func (c *Cart) RemoveItem(itemID string) bool {
	for i, item := range c.Items {
		if item.ID == itemID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// UpdateItemQuantity updates the quantity of a cart item.
func (c *Cart) UpdateItemQuantity(itemID string, quantity int) bool {
	if quantity <= 0 {
		return c.RemoveItem(itemID)
	}
	
	for i, item := range c.Items {
		if item.ID == itemID {
			c.Items[i].Quantity = quantity
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// Clear removes all items from the cart.
func (c *Cart) Clear() {
	c.Items = []CartItem{}
	c.UpdatedAt = time.Now()
}

// IsEmpty returns true if the cart has no items.
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// ItemCount returns the total number of items (sum of quantities).
func (c *Cart) ItemCount() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

// Subtotal calculates the subtotal (before discounts/tax).
func (c *Cart) Subtotal() money.Money {
	if len(c.Items) == 0 {
		return money.Zero("USD")
	}
	
	currency := c.Items[0].Price.Currency
	total := money.Zero(currency)
	
	for _, item := range c.Items {
		itemTotal := item.Price.MultiplyInt(item.Quantity)
		total, _ = total.Add(itemTotal)
	}
	
	return total
}

// FindItem finds a cart item by ID.
func (c *Cart) FindItem(itemID string) *CartItem {
	for i := range c.Items {
		if c.Items[i].ID == itemID {
			return &c.Items[i]
		}
	}
	return nil
}

// Merge merges another cart into this one (useful for guest->user cart migration).
func (c *Cart) Merge(other *Cart) {
	for _, otherItem := range other.Items {
		found := false
		for i, existing := range c.Items {
			if existing.ProductID == otherItem.ProductID && 
			   existing.VariantID == otherItem.VariantID {
				c.Items[i].Quantity += otherItem.Quantity
				found = true
				break
			}
		}
		if !found {
			c.Items = append(c.Items, otherItem)
		}
	}
	c.UpdatedAt = time.Now()
}
