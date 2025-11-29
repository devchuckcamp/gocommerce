package cart

import (
	"context"
	"errors"
	"time"

	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/inventory"
	"github.com/devchuckcamp/gocommerce/money"
)

var (
	ErrCartNotFound     = errors.New("cart not found")
	ErrItemNotFound     = errors.New("item not found")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrOutOfStock       = errors.New("product out of stock")
)

// Repository defines methods for cart persistence.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Cart, error)
	FindByUserID(ctx context.Context, userID string) (*Cart, error)
	FindBySessionID(ctx context.Context, sessionID string) (*Cart, error)
	Save(ctx context.Context, cart *Cart) error
	Delete(ctx context.Context, id string) error
}

// Service provides cart business logic.
type Service interface {
	GetCart(ctx context.Context, cartID string) (*Cart, error)
	GetOrCreateCart(ctx context.Context, userID, sessionID string) (*Cart, error)
	AddItem(ctx context.Context, cartID string, req AddItemRequest) (*Cart, error)
	UpdateItemQuantity(ctx context.Context, cartID, itemID string, quantity int) (*Cart, error)
	RemoveItem(ctx context.Context, cartID, itemID string) (*Cart, error)
	Clear(ctx context.Context, cartID string) (*Cart, error)
	MergeCarts(ctx context.Context, sourceCartID, targetCartID string) (*Cart, error)
}

// AddItemRequest contains data needed to add an item to cart.
type AddItemRequest struct {
	ProductID  string
	VariantID  *string
	Quantity   int
	Attributes map[string]string
}

// CartService implements the Service interface.
type CartService struct {
	repo             Repository
	productRepo      catalog.ProductRepository
	variantRepo      catalog.VariantRepository
	inventoryService inventory.Service
	idGenerator      func() string
}

// NewCartService creates a new cart service.
func NewCartService(
	repo Repository,
	productRepo catalog.ProductRepository,
	variantRepo catalog.VariantRepository,
	inventoryService inventory.Service,
	idGenerator func() string,
) *CartService {
	return &CartService{
		repo:             repo,
		productRepo:      productRepo,
		variantRepo:      variantRepo,
		inventoryService: inventoryService,
		idGenerator:      idGenerator,
	}
}

// GetCart retrieves a cart by ID.
func (s *CartService) GetCart(ctx context.Context, cartID string) (*Cart, error) {
	return s.repo.FindByID(ctx, cartID)
}

// GetOrCreateCart gets an existing cart or creates a new one.
func (s *CartService) GetOrCreateCart(ctx context.Context, userID, sessionID string) (*Cart, error) {
	var cart *Cart
	var err error
	
	if userID != "" {
		cart, err = s.repo.FindByUserID(ctx, userID)
	} else if sessionID != "" {
		cart, err = s.repo.FindBySessionID(ctx, sessionID)
	} else {
		return nil, errors.New("userID or sessionID required")
	}
	
	if err == nil && cart != nil {
		return cart, nil
	}
	
	// Create new cart
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	cart = &Cart{
		ID:        s.idGenerator(),
		UserID:    userID,
		SessionID: sessionID,
		Items:     []CartItem{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ExpiresAt: &expiresAt,
	}
	
	err = s.repo.Save(ctx, cart)
	if err != nil {
		return nil, err
	}
	
	return cart, nil
}

// AddItem adds a product to the cart with stock validation.
func (s *CartService) AddItem(ctx context.Context, cartID string, req AddItemRequest) (*Cart, error) {
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	
	cart, err := s.repo.FindByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	
	// Fetch product
	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	
	if !product.IsActive() {
		return nil, errors.New("product not available")
	}
	
	// Check inventory if service available
	var sku string
	var price money.Money
	
	if req.VariantID != nil {
		variant, err := s.variantRepo.FindByID(ctx, *req.VariantID)
		if err != nil {
			return nil, err
		}
		sku = variant.SKU
		price = variant.Price
		
		if !variant.IsAvailable {
			return nil, errors.New("variant not available")
		}
	} else {
		sku = product.SKU
		price = product.BasePrice
	}
	
	// Check stock availability
	if s.inventoryService != nil {
		available, err := s.inventoryService.GetAvailableStock(ctx, sku)
		if err == nil && available < req.Quantity {
			return nil, ErrOutOfStock
		}
	}
	
	// Add item to cart
	item := CartItem{
		ID:         s.idGenerator(),
		ProductID:  req.ProductID,
		VariantID:  req.VariantID,
		SKU:        sku,
		Name:       product.Name,
		Price:      price,
		Quantity:   req.Quantity,
		Attributes: req.Attributes,
		AddedAt:    time.Now(),
	}
	
	cart.AddItem(item)
	
	err = s.repo.Save(ctx, cart)
	if err != nil {
		return nil, err
	}
	
	return cart, nil
}

// UpdateItemQuantity updates the quantity of a cart item.
func (s *CartService) UpdateItemQuantity(ctx context.Context, cartID, itemID string, quantity int) (*Cart, error) {
	cart, err := s.repo.FindByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	
	item := cart.FindItem(itemID)
	if item == nil {
		return nil, ErrItemNotFound
	}
	
	// Check stock if increasing quantity
	if quantity > item.Quantity && s.inventoryService != nil {
		available, err := s.inventoryService.GetAvailableStock(ctx, item.SKU)
		if err == nil && available < quantity {
			return nil, ErrOutOfStock
		}
	}
	
	cart.UpdateItemQuantity(itemID, quantity)
	
	err = s.repo.Save(ctx, cart)
	if err != nil {
		return nil, err
	}
	
	return cart, nil
}

// RemoveItem removes an item from the cart.
func (s *CartService) RemoveItem(ctx context.Context, cartID, itemID string) (*Cart, error) {
	cart, err := s.repo.FindByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	
	if !cart.RemoveItem(itemID) {
		return nil, ErrItemNotFound
	}
	
	err = s.repo.Save(ctx, cart)
	if err != nil {
		return nil, err
	}
	
	return cart, nil
}

// Clear removes all items from the cart.
func (s *CartService) Clear(ctx context.Context, cartID string) (*Cart, error) {
	cart, err := s.repo.FindByID(ctx, cartID)
	if err != nil {
		return nil, err
	}
	
	cart.Clear()
	
	err = s.repo.Save(ctx, cart)
	if err != nil {
		return nil, err
	}
	
	return cart, nil
}

// MergeCarts merges source cart into target cart (e.g., guest -> user cart).
func (s *CartService) MergeCarts(ctx context.Context, sourceCartID, targetCartID string) (*Cart, error) {
	sourceCart, err := s.repo.FindByID(ctx, sourceCartID)
	if err != nil {
		return nil, err
	}
	
	targetCart, err := s.repo.FindByID(ctx, targetCartID)
	if err != nil {
		return nil, err
	}
	
	targetCart.Merge(sourceCart)
	
	err = s.repo.Save(ctx, targetCart)
	if err != nil {
		return nil, err
	}
	
	// Optionally delete source cart
	_ = s.repo.Delete(ctx, sourceCartID)
	
	return targetCart, nil
}
