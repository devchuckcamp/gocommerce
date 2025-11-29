package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/money"
	"github.com/devchuckcamp/gocommerce/orders"
	"github.com/devchuckcamp/gocommerce/pricing"
)

// MemoryStore implements all repository interfaces using in-memory storage
type MemoryStore struct {
	products   map[string]*catalog.Product
	variants   map[string]*catalog.Variant
	carts      map[string]*cart.Cart
	orders     map[string]*orders.Order
	promotions map[string]*pricing.Promotion
	mu         sync.RWMutex
	
	// Separate repo instances to satisfy different interfaces
	cartRepo      cartRepository
	variantRepo   variantRepository
	orderRepo     orderRepository
	promotionRepo promotionRepository
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		products:   make(map[string]*catalog.Product),
		variants:   make(map[string]*catalog.Variant),
		carts:      make(map[string]*cart.Cart),
		orders:     make(map[string]*orders.Order),
		promotions: make(map[string]*pricing.Promotion),
	}
	s.cartRepo = cartRepository{store: s}
	s.variantRepo = variantRepository{store: s}
	s.orderRepo = orderRepository{store: s}
	s.promotionRepo = promotionRepository{store: s}
	return s
}

// Wrapper types to implement specific repository interfaces
type cartRepository struct{ store *MemoryStore }
type variantRepository struct{ store *MemoryStore }
type orderRepository struct{ store *MemoryStore }
type promotionRepository struct{ store *MemoryStore }

// Product Repository

func (s *MemoryStore) FindProductByID(ctx context.Context, id string) (*catalog.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	product, ok := s.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (s *MemoryStore) FindByID(ctx context.Context, id string) (*catalog.Product, error) {
	return s.FindProductByID(ctx, id)
}

func (s *MemoryStore) FindBySKU(ctx context.Context, sku string) (*catalog.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, p := range s.products {
		if p.SKU == sku {
			return p, nil
		}
	}
	return nil, errors.New("product not found")
}

func (s *MemoryStore) FindByCategory(ctx context.Context, categoryID string, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	return nil, errors.New("not implemented")
}

func (s *MemoryStore) FindByBrand(ctx context.Context, brandID string, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	return nil, errors.New("not implemented")
}

func (s *MemoryStore) Search(ctx context.Context, query string, filter catalog.ProductFilter) ([]*catalog.Product, error) {
	return nil, errors.New("not implemented")
}

func (s *MemoryStore) Save(ctx context.Context, product *catalog.Product) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.products[product.ID] = product
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.products, id)
	return nil
}

func (s *MemoryStore) ListProducts() []*catalog.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	products := make([]*catalog.Product, 0, len(s.products))
	for _, p := range s.products {
		products = append(products, p)
	}
	return products
}

// Variant Repository implementation

func (r *variantRepository) FindByID(ctx context.Context, id string) (*catalog.Variant, error) {
	return nil, errors.New("not implemented")
}

func (r *variantRepository) FindBySKU(ctx context.Context, sku string) (*catalog.Variant, error) {
	return nil, errors.New("not implemented")
}

func (r *variantRepository) FindByProductID(ctx context.Context, productID string) ([]*catalog.Variant, error) {
	return nil, nil
}

func (r *variantRepository) Save(ctx context.Context, variant *catalog.Variant) error {
	return nil
}

func (r *variantRepository) Delete(ctx context.Context, id string) error {
	return nil
}

// Cart Repository implementation

func (r *cartRepository) FindByID(ctx context.Context, id string) (*cart.Cart, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	
	c, ok := r.store.carts[id]
	if !ok {
		return nil, cart.ErrCartNotFound
	}
	return c, nil
}

func (r *cartRepository) FindByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	
	for _, c := range r.store.carts {
		if c.UserID == userID {
			return c, nil
		}
	}
	return nil, cart.ErrCartNotFound
}

func (r *cartRepository) FindBySessionID(ctx context.Context, sessionID string) (*cart.Cart, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	
	for _, c := range r.store.carts {
		if c.SessionID == sessionID {
			return c, nil
		}
	}
	return nil, cart.ErrCartNotFound
}

func (r *cartRepository) Save(ctx context.Context, c *cart.Cart) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	
	r.store.carts[c.ID] = c
	return nil
}

func (r *cartRepository) Delete(ctx context.Context, id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	
	delete(r.store.carts, id)
	return nil
}

// Order Repository implementation

func (r *orderRepository) FindByID(ctx context.Context, id string) (*orders.Order, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	
	order, ok := r.store.orders[id]
	if !ok {
		return nil, orders.ErrOrderNotFound
	}
	return order, nil
}

func (r *orderRepository) FindByOrderNumber(ctx context.Context, orderNumber string) (*orders.Order, error) {
	return nil, errors.New("not implemented")
}

func (r *orderRepository) FindByUserID(ctx context.Context, userID string, filter orders.OrderFilter) ([]*orders.Order, error) {
	return nil, errors.New("not implemented")
}

func (r *orderRepository) Save(ctx context.Context, order *orders.Order) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	
	r.store.orders[order.ID] = order
	return nil
}

func (r *orderRepository) Delete(ctx context.Context, id string) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	
	delete(r.store.orders, id)
	return nil
}

// Promotion Repository implementation

func (r *promotionRepository) FindByCode(ctx context.Context, code string) (*pricing.Promotion, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	
	for _, p := range r.store.promotions {
		if p.Code == code {
			return p, nil
		}
	}
	return nil, errors.New("promotion not found")
}

func (r *promotionRepository) FindActive(ctx context.Context) ([]*pricing.Promotion, error) {
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	
	promotions := make([]*pricing.Promotion, 0)
	for _, p := range r.store.promotions {
		if p.IsActive {
			promotions = append(promotions, p)
		}
	}
	return promotions, nil
}

func (r *promotionRepository) Save(ctx context.Context, promotion *pricing.Promotion) error {
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	
	r.store.promotions[promotion.ID] = promotion
	return nil
}

// Seed sample products
func seedProducts(store *MemoryStore) {
	products := []*catalog.Product{
		{
			ID:          "prod-1",
			SKU:         "TSHIRT-BLU-M",
			Name:        "Blue T-Shirt (Medium)",
			Description: "Comfortable cotton t-shirt in blue",
			BasePrice:   mustMoney(49.99, "USD"),
			Status:      catalog.ProductStatusActive,
			Images:      []string{"/images/tshirt-blue.jpg"},
			Attributes: map[string]string{
				"color":    "blue",
				"size":     "M",
				"material": "cotton",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "prod-2",
			SKU:         "JEANS-BLK-32",
			Name:        "Black Jeans (32)",
			Description: "Classic black denim jeans",
			BasePrice:   mustMoney(79.99, "USD"),
			Status:      catalog.ProductStatusActive,
			Images:      []string{"/images/jeans-black.jpg"},
			Attributes: map[string]string{
				"color":    "black",
				"size":     "32",
				"material": "denim",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "prod-3",
			SKU:         "SNEAKERS-WHT-10",
			Name:        "White Sneakers (Size 10)",
			Description: "Comfortable white sneakers",
			BasePrice:   mustMoney(89.99, "USD"),
			Status:      catalog.ProductStatusActive,
			Images:      []string{"/images/sneakers-white.jpg"},
			Attributes: map[string]string{
				"color": "white",
				"size":  "10",
				"type":  "casual",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "prod-4",
			SKU:         "HOODIE-GRY-L",
			Name:        "Gray Hoodie (Large)",
			Description: "Cozy gray hoodie with front pocket",
			BasePrice:   mustMoney(59.99, "USD"),
			Status:      catalog.ProductStatusActive,
			Images:      []string{"/images/hoodie-gray.jpg"},
			Attributes: map[string]string{
				"color":    "gray",
				"size":     "L",
				"material": "cotton blend",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	for _, p := range products {
		store.Save(context.Background(), p)
	}
}

func mustMoney(amount float64, currency string) money.Money {
	m, _ := money.NewFromFloat(amount, currency)
	return m
}
