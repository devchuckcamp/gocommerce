package supplier

import (
	"context"
	"errors"
	"time"

	"github.com/devchuckcamp/gocommerce/money"
)

var (
	ErrSupplierNotFound = errors.New("supplier not found")
	ErrDuplicateCode    = errors.New("supplier code already exists")
	ErrInvalidSupplier  = errors.New("invalid supplier data")
)

// Supplier represents a product supplier or vendor.
type Supplier struct {
	ID           string
	Code         string // Unique supplier code (e.g., "SUP-001")
	Name         string
	ContactName  string
	ContactEmail string
	ContactPhone string
	Address      string
	Website      string
	Notes        string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SupplierStatus represents the status of a supplier.
type SupplierStatus string

const (
	SupplierStatusActive   SupplierStatus = "active"
	SupplierStatusInactive SupplierStatus = "inactive"
	SupplierStatusPending  SupplierStatus = "pending"
)

// ProductSupplier represents the relationship between a product and its supplier.
type ProductSupplier struct {
	ID           string
	ProductID    string
	SupplierID   string
	SupplierSKU  string      // Supplier's product code
	CostPrice    money.Money // Purchase cost from supplier
	LeadTimeDays int         // Days to receive stock
	MinOrderQty  int         // Minimum order quantity
	IsPrimary    bool        // Primary supplier for this product
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InventorySupplier represents supplier contribution to inventory levels (multi-warehouse).
type InventorySupplier struct {
	ID               string
	InventoryLevelID string
	SupplierID       string
	QuantityFromSupplier int
	LastRestockAt    *time.Time
	NextRestockAt    *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SupplierRepository defines methods for supplier persistence.
type SupplierRepository interface {
	FindByID(ctx context.Context, id string) (*Supplier, error)
	FindByCode(ctx context.Context, code string) (*Supplier, error)
	FindAll(ctx context.Context, filter SupplierFilter) ([]*Supplier, error)
	FindActive(ctx context.Context) ([]*Supplier, error)
	Save(ctx context.Context, supplier *Supplier) error
	Delete(ctx context.Context, id string) error
}

// ProductSupplierRepository defines methods for product-supplier relationship persistence.
type ProductSupplierRepository interface {
	FindByID(ctx context.Context, id string) (*ProductSupplier, error)
	FindByProductID(ctx context.Context, productID string) ([]*ProductSupplier, error)
	FindBySupplierID(ctx context.Context, supplierID string) ([]*ProductSupplier, error)
	FindPrimaryByProductID(ctx context.Context, productID string) (*ProductSupplier, error)
	Save(ctx context.Context, ps *ProductSupplier) error
	Delete(ctx context.Context, id string) error
}

// InventorySupplierRepository defines methods for inventory-supplier relationship persistence.
type InventorySupplierRepository interface {
	FindByID(ctx context.Context, id string) (*InventorySupplier, error)
	FindByInventoryLevelID(ctx context.Context, inventoryLevelID string) ([]*InventorySupplier, error)
	FindBySupplierID(ctx context.Context, supplierID string) ([]*InventorySupplier, error)
	Save(ctx context.Context, is *InventorySupplier) error
	Delete(ctx context.Context, id string) error
}

// SupplierFilter defines query filters for suppliers.
type SupplierFilter struct {
	IsActive *bool
	Search   string // Search in name, code
	Limit    int
	Offset   int
	SortBy   string // e.g., "name", "code", "created_at_desc"
}

// IsValid returns true if the supplier has required fields.
func (s *Supplier) IsValid() bool {
	return s.Code != "" && s.Name != ""
}

// CanSupply returns true if the supplier is active and can fulfill orders.
func (s *Supplier) CanSupply() bool {
	return s.IsActive
}

// GetEstimatedDelivery returns the expected delivery date based on lead time.
func (ps *ProductSupplier) GetEstimatedDelivery(orderDate time.Time) time.Time {
	return orderDate.AddDate(0, 0, ps.LeadTimeDays)
}

// CalculateOrderCost calculates the total cost for a given quantity.
func (ps *ProductSupplier) CalculateOrderCost(quantity int) money.Money {
	return ps.CostPrice.Multiply(float64(quantity))
}
