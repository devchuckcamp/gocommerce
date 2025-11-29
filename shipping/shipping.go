package shipping

import (
	"context"

	"github.com/devchuckcamp/gocommerce/money"
)

// RateCalculator defines the shipping rate calculator interface.
type RateCalculator interface {
	GetRate(ctx context.Context, req RateRequest) (*ShippingRate, error)
	GetAvailableRates(ctx context.Context, req RateRequest) ([]*ShippingRate, error)
}

// RateRequest contains data needed to calculate shipping rates.
type RateRequest struct {
	Items              []ShippableItem
	SourceAddress      Address
	DestinationAddress Address
	ShippingMethodID   string
}

// ShippableItem represents an item that can be shipped.
type ShippableItem struct {
	SKU            string
	Quantity       int
	WeightGrams    int
	LengthCm       int
	WidthCm        int
	HeightCm       int
	IsFragile      bool
	RequiresColdChain bool
}

// Address represents a shipping address.
type Address struct {
	Country    string
	State      string
	City       string
	PostalCode string
}

// ShippingRate represents the cost to ship with a specific method.
type ShippingRate struct {
	MethodID          string
	MethodName        string
	Cost              money.Money
	EstimatedDays     int
	EstimatedDaysMin  int
	EstimatedDaysMax  int
	Carrier           string
	ServiceLevel      string
	IsGuaranteed      bool
}

// ShippingMethod represents a shipping method/carrier.
type ShippingMethod struct {
	ID          string
	Name        string
	Description string
	Carrier     string
	ServiceLevel string
	IsActive    bool
	// Rate calculation rules
	FlatRate         *money.Money
	RatePerWeightKg  *money.Money
	FreeShippingMin  *money.Money
}

// Repository defines methods for shipping data persistence.
type Repository interface {
	FindMethod(ctx context.Context, id string) (*ShippingMethod, error)
	FindActiveMethods(ctx context.Context) ([]*ShippingMethod, error)
	SaveMethod(ctx context.Context, method *ShippingMethod) error
}

// Shipment represents a package shipment.
type Shipment struct {
	ID              string
	OrderID         string
	Carrier         string
	ServiceLevel    string
	TrackingNumber  string
	TrackingURL     string
	LabelURL        string
	Status          ShipmentStatus
	ShippedAt       int64
	DeliveredAt     *int64
	EstimatedDelivery int64
}

// ShipmentStatus represents the state of a shipment.
type ShipmentStatus string

const (
	ShipmentStatusPending    ShipmentStatus = "pending"
	ShipmentStatusInTransit  ShipmentStatus = "in_transit"
	ShipmentStatusDelivered  ShipmentStatus = "delivered"
	ShipmentStatusFailed     ShipmentStatus = "failed"
	ShipmentStatusReturned   ShipmentStatus = "returned"
)
