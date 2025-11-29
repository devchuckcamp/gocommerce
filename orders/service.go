package orders

import (
	"context"
	"errors"
	"time"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/inventory"
	"github.com/devchuckcamp/gocommerce/payments"
	"github.com/devchuckcamp/gocommerce/pricing"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidStatus      = errors.New("invalid status transition")
	ErrEmptyCart          = errors.New("cart is empty")
	ErrInvalidAddress     = errors.New("invalid address")
	ErrPaymentFailed      = errors.New("payment failed")
)

// Repository defines methods for order persistence.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
	FindByUserID(ctx context.Context, userID string, filter OrderFilter) ([]*Order, error)
	Save(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id string) error
}

// OrderFilter defines query filters for orders.
type OrderFilter struct {
	Status    *OrderStatus
	DateFrom  *time.Time
	DateTo    *time.Time
	Limit     int
	Offset    int
}

// Service provides order business logic.
type Service interface {
	CreateFromCart(ctx context.Context, req CreateOrderRequest) (*Order, error)
	GetOrder(ctx context.Context, id string) (*Order, error)
	GetUserOrders(ctx context.Context, userID string, filter OrderFilter) ([]*Order, error)
	UpdateStatus(ctx context.Context, orderID string, status OrderStatus) (*Order, error)
	CancelOrder(ctx context.Context, orderID string, reason string) (*Order, error)
}

// CreateOrderRequest contains data needed to create an order.
type CreateOrderRequest struct {
	Cart            *cart.Cart
	UserID          string
	ShippingAddress Address
	BillingAddress  Address
	PaymentMethodID string
	PromotionCodes  []string
	ShippingMethodID string
	Notes           string
	IPAddress       string
	UserAgent       string
}

// OrderService implements the Service interface.
type OrderService struct {
	repo              Repository
	pricingService    pricing.Service
	inventoryService  inventory.Service
	paymentGateway    payments.Gateway
	orderNumberGen    func() string
	idGenerator       func() string
}

// NewOrderService creates a new order service.
func NewOrderService(
	repo Repository,
	pricingService pricing.Service,
	inventoryService inventory.Service,
	paymentGateway payments.Gateway,
	orderNumberGen func() string,
	idGenerator func() string,
) *OrderService {
	return &OrderService{
		repo:             repo,
		pricingService:   pricingService,
		inventoryService: inventoryService,
		paymentGateway:   paymentGateway,
		orderNumberGen:   orderNumberGen,
		idGenerator:      idGenerator,
	}
}

// CreateFromCart creates an order from a cart.
func (s *OrderService) CreateFromCart(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	if req.Cart == nil || req.Cart.IsEmpty() {
		return nil, ErrEmptyCart
	}
	
	if !req.ShippingAddress.IsComplete() {
		return nil, ErrInvalidAddress
	}
	
	if !req.BillingAddress.IsComplete() {
		req.BillingAddress = req.ShippingAddress
	}
	
	// Calculate pricing
	shippingMethodID := req.ShippingMethodID
	pricingResult, err := s.pricingService.PriceCart(ctx, pricing.PriceCartRequest{
		Cart:             req.Cart,
		PromotionCodes:   req.PromotionCodes,
		ShippingMethodID: &shippingMethodID,
		ShippingAddress: &pricing.Address{
			Country:    req.ShippingAddress.Country,
			State:      req.ShippingAddress.State,
			City:       req.ShippingAddress.City,
			PostalCode: req.ShippingAddress.PostalCode,
		},
		TaxInclusive: false,
	})
	if err != nil {
		return nil, err
	}
	
	// Reserve inventory
	if s.inventoryService != nil {
		reservationID := s.idGenerator()
		for _, item := range req.Cart.Items {
			err := s.inventoryService.Reserve(ctx, item.SKU, item.Quantity, reservationID)
			if err != nil {
				// Rollback previous reservations
				s.rollbackInventory(ctx, reservationID)
				return nil, err
			}
		}
	}
	
	// Create order items
	orderItems := make([]OrderItem, len(req.Cart.Items))
	for i, cartItem := range req.Cart.Items {
		var itemPrice pricing.LineItemPrice
		if i < len(pricingResult.LineItemPrices) {
			itemPrice = pricingResult.LineItemPrices[i]
		}
		
		orderItems[i] = OrderItem{
			ID:             s.idGenerator(),
			ProductID:      cartItem.ProductID,
			VariantID:      cartItem.VariantID,
			SKU:            cartItem.SKU,
			Name:           cartItem.Name,
			UnitPrice:      cartItem.Price,
			Quantity:       cartItem.Quantity,
			DiscountAmount: itemPrice.DiscountAmount,
			TaxAmount:      itemPrice.TaxAmount,
			Total:          itemPrice.Total,
			Attributes:     cartItem.Attributes,
		}
	}
	
	// Create order
	order := &Order{
		ID:              s.idGenerator(),
		OrderNumber:     s.orderNumberGen(),
		UserID:          req.UserID,
		Status:          OrderStatusPending,
		Items:           orderItems,
		ShippingAddress: req.ShippingAddress,
		BillingAddress:  req.BillingAddress,
		PaymentMethodID: req.PaymentMethodID,
		Subtotal:        pricingResult.Subtotal,
		DiscountTotal:   pricingResult.DiscountTotal,
		TaxTotal:        pricingResult.TaxTotal,
		ShippingTotal:   pricingResult.ShippingTotal,
		Total:           pricingResult.Total,
		Notes:           req.Notes,
		IPAddress:       req.IPAddress,
		UserAgent:       req.UserAgent,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// Save order
	err = s.repo.Save(ctx, order)
	if err != nil {
		return nil, err
	}
	
	// Process payment if gateway available
	if s.paymentGateway != nil {
		intent, err := s.paymentGateway.CreateIntent(ctx, payments.IntentRequest{
			Amount:          order.Total,
			Currency:        order.Total.Currency,
			PaymentMethodID: req.PaymentMethodID,
			OrderID:         order.ID,
			Description:     "Order " + order.OrderNumber,
		})
		if err != nil {
			return nil, ErrPaymentFailed
		}
		
		if intent.Status == payments.IntentStatusSucceeded {
			order.UpdateStatus(OrderStatusPaid)
			s.repo.Save(ctx, order)
		}
	}
	
	return order, nil
}

// GetOrder retrieves an order by ID.
func (s *OrderService) GetOrder(ctx context.Context, id string) (*Order, error) {
	return s.repo.FindByID(ctx, id)
}

// GetUserOrders retrieves orders for a user.
func (s *OrderService) GetUserOrders(ctx context.Context, userID string, filter OrderFilter) ([]*Order, error) {
	return s.repo.FindByUserID(ctx, userID, filter)
}

// UpdateStatus updates the order status.
func (s *OrderService) UpdateStatus(ctx context.Context, orderID string, status OrderStatus) (*Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	if !order.UpdateStatus(status) {
		return nil, ErrInvalidStatus
	}
	
	err = s.repo.Save(ctx, order)
	if err != nil {
		return nil, err
	}
	
	return order, nil
}

// CancelOrder cancels an order.
func (s *OrderService) CancelOrder(ctx context.Context, orderID string, reason string) (*Order, error) {
	order, err := s.repo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	
	if !order.IsCancelable() {
		return nil, errors.New("order cannot be canceled")
	}
	
	// Release inventory
	if s.inventoryService != nil {
		for _, item := range order.Items {
			_ = s.inventoryService.Release(ctx, item.SKU, item.Quantity, order.ID)
		}
	}
	
	order.UpdateStatus(OrderStatusCanceled)
	order.Notes = order.Notes + "\nCanceled: " + reason
	
	err = s.repo.Save(ctx, order)
	if err != nil {
		return nil, err
	}
	
	return order, nil
}

// rollbackInventory releases reserved inventory.
func (s *OrderService) rollbackInventory(ctx context.Context, reservationID string) {
	if s.inventoryService != nil {
		_ = s.inventoryService.Release(ctx, "", 0, reservationID)
	}
}
