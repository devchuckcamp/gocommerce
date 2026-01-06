package postgres

import (
	"context"
	"database/sql"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/orders"
	"github.com/devchuckcamp/gocommerce/pricing"
)

// Store bundles Postgres-backed repository implementations for the sample-project.
type Store struct {
	DB *sql.DB

	Products   *ProductRepository
	Variants   *VariantRepository
	Carts      *CartRepository
	Orders     *OrderRepository
	Promotions *PromotionRepository
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		DB:         db,
		Products:   NewProductRepository(db),
		Variants:   NewVariantRepository(db),
		Carts:      NewCartRepository(db),
		Orders:     NewOrderRepository(db),
		Promotions: NewPromotionRepository(db),
	}
}

func (s *Store) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}

// Convenience accessors (helps satisfy sample-project wiring).
func (s *Store) CartRepo() cart.Repository { return s.Carts }
func (s *Store) ProductRepo() catalog.ProductRepository { return s.Products }
func (s *Store) OrderRepo() orders.Repository { return s.Orders }
func (s *Store) PromotionRepo() pricing.PromotionRepository { return s.Promotions }

// ProductStore-like helpers for the HTTP API.
func (s *Store) ListProducts(ctx context.Context) ([]*catalog.Product, error) {
	return s.Products.ListProducts(ctx, catalog.ProductFilter{Limit: 1000, Offset: 0, SortBy: "created_at_desc"})
}

func (s *Store) FindProductByID(ctx context.Context, id string) (*catalog.Product, error) {
	return s.Products.FindByID(ctx, id)
}
