package inventory

import (
	"context"
	"errors"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidSKU        = errors.New("invalid SKU")
	ErrReservationFailed = errors.New("reservation failed")
)

// Service defines the inventory service interface.
type Service interface {
	GetAvailableStock(ctx context.Context, sku string) (int, error)
	GetReservedStock(ctx context.Context, sku string) (int, error)
	Reserve(ctx context.Context, sku string, quantity int, referenceID string) error
	Release(ctx context.Context, sku string, quantity int, referenceID string) error
	Commit(ctx context.Context, referenceID string) error
	AdjustStock(ctx context.Context, sku string, quantity int, reason string) error
}

// StockLevel represents inventory stock information.
type StockLevel struct {
	SKU              string
	QuantityOnHand   int
	QuantityReserved int
	QuantityAvailable int
	ReorderPoint     int
	ReorderQuantity  int
}

// IsInStock returns true if the SKU has available stock.
func (s *StockLevel) IsInStock() bool {
	return s.QuantityAvailable > 0
}

// NeedsReorder returns true if stock is at or below reorder point.
func (s *StockLevel) NeedsReorder() bool {
	return s.QuantityAvailable <= s.ReorderPoint
}

// Reservation represents a stock reservation.
type Reservation struct {
	ID          string
	SKU         string
	Quantity    int
	ReferenceID string // Order ID, cart ID, etc.
	Status      ReservationStatus
	ExpiresAt   int64 // Unix timestamp
}

// ReservationStatus represents the state of a reservation.
type ReservationStatus string

const (
	ReservationStatusActive   ReservationStatus = "active"
	ReservationStatusCommitted ReservationStatus = "committed"
	ReservationStatusReleased  ReservationStatus = "released"
	ReservationStatusExpired   ReservationStatus = "expired"
)

// Repository defines methods for inventory persistence.
type Repository interface {
	GetStockLevel(ctx context.Context, sku string) (*StockLevel, error)
	UpdateStockLevel(ctx context.Context, level *StockLevel) error
	GetReservation(ctx context.Context, id string) (*Reservation, error)
	GetReservationsByReference(ctx context.Context, referenceID string) ([]*Reservation, error)
	SaveReservation(ctx context.Context, reservation *Reservation) error
	DeleteReservation(ctx context.Context, id string) error
	GetExpiredReservations(ctx context.Context) ([]*Reservation, error)
}

// StockAdjustment represents a stock level change.
type StockAdjustment struct {
	ID         string
	SKU        string
	Quantity   int    // Positive for increase, negative for decrease
	Reason     string // e.g., "restock", "damage", "correction"
	ReferenceID string
	CreatedAt  int64
}
