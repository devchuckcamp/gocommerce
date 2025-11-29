package orders

import (
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

// Order represents a customer order.
type Order struct {
	ID              string
	OrderNumber     string // Human-readable order number
	UserID          string
	Status          OrderStatus
	Items           []OrderItem
	ShippingAddress Address
	BillingAddress  Address
	PaymentMethodID string
	
	// Pricing
	Subtotal      money.Money
	DiscountTotal money.Money
	TaxTotal      money.Money
	ShippingTotal money.Money
	Total         money.Money
	
	// Metadata
	Notes         string
	IPAddress     string
	UserAgent     string
	
	// Timestamps
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
	CanceledAt  *time.Time
}

// OrderItem represents an item in an order.
type OrderItem struct {
	ID            string
	ProductID     string
	VariantID     *string
	SKU           string
	Name          string
	UnitPrice     money.Money
	Quantity      int
	DiscountAmount money.Money
	TaxAmount     money.Money
	Total         money.Money
	Attributes    map[string]string
}

// OrderStatus represents the state of an order.
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCanceled   OrderStatus = "canceled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// Address represents a shipping or billing address.
type Address struct {
	FirstName   string
	LastName    string
	Company     string
	AddressLine1 string
	AddressLine2 string
	City        string
	State       string
	PostalCode  string
	Country     string
	Phone       string
}

// FullName returns the full name from the address.
func (a Address) FullName() string {
	return a.FirstName + " " + a.LastName
}

// IsComplete checks if address has required fields.
func (a Address) IsComplete() bool {
	return a.FirstName != "" &&
		a.LastName != "" &&
		a.AddressLine1 != "" &&
		a.City != "" &&
		a.PostalCode != "" &&
		a.Country != ""
}

// CanTransitionTo checks if an order can transition to a new status.
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending: {
			OrderStatusPaid,
			OrderStatusCanceled,
		},
		OrderStatusPaid: {
			OrderStatusProcessing,
			OrderStatusCanceled,
			OrderStatusRefunded,
		},
		OrderStatusProcessing: {
			OrderStatusShipped,
			OrderStatusCanceled,
		},
		OrderStatusShipped: {
			OrderStatusDelivered,
		},
		OrderStatusDelivered: {
			OrderStatusRefunded,
		},
	}
	
	allowedTransitions, exists := transitions[o.Status]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	
	return false
}

// UpdateStatus updates the order status if transition is valid.
func (o *Order) UpdateStatus(newStatus OrderStatus) bool {
	if !o.CanTransitionTo(newStatus) {
		return false
	}
	
	o.Status = newStatus
	o.UpdatedAt = time.Now()
	
	if newStatus == OrderStatusDelivered {
		now := time.Now()
		o.CompletedAt = &now
	}
	
	if newStatus == OrderStatusCanceled {
		now := time.Now()
		o.CanceledAt = &now
	}
	
	return true
}

// IsCancelable returns true if the order can be canceled.
func (o *Order) IsCancelable() bool {
	return o.Status == OrderStatusPending ||
		o.Status == OrderStatusPaid ||
		o.Status == OrderStatusProcessing
}

// IsRefundable returns true if the order can be refunded.
func (o *Order) IsRefundable() bool {
	return o.Status == OrderStatusPaid ||
		o.Status == OrderStatusDelivered
}

// ItemCount returns the total number of items.
func (o *Order) ItemCount() int {
	count := 0
	for _, item := range o.Items {
		count += item.Quantity
	}
	return count
}
