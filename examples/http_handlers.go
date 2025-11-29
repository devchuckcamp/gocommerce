package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/orders"
	"github.com/devchuckcamp/gocommerce/pricing"
)

// This file shows how HTTP handlers would use the domain services.
// Note: This is NOT part of the domain library - just examples!

// CartHandler demonstrates how an HTTP API would use cart services.
type CartHandler struct {
	cartService    cart.Service
	pricingService pricing.Service
}

// AddToCartRequest is the HTTP request body.
type AddToCartRequest struct {
	ProductID  string            `json:"product_id"`
	VariantID  *string           `json:"variant_id,omitempty"`
	Quantity   int               `json:"quantity"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// CartResponse is the HTTP response body.
type CartResponse struct {
	ID         string              `json:"id"`
	ItemCount  int                 `json:"item_count"`
	Subtotal   string              `json:"subtotal"`
	Items      []CartItemResponse  `json:"items"`
}

type CartItemResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SKU      string `json:"sku"`
	Price    string `json:"price"`
	Quantity int    `json:"quantity"`
}

// HandleAddToCart shows how a POST /cart/items endpoint would work.
func (h *CartHandler) HandleAddToCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Extract user ID from auth context (handled by auth middleware)
	userID := getUserIDFromContext(ctx) // Your auth logic
	
	// Parse request
	var req AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Validate
	if req.ProductID == "" || req.Quantity <= 0 {
		http.Error(w, "Invalid product or quantity", http.StatusBadRequest)
		return
	}
	
	// Get or create cart
	shoppingCart, err := h.cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}
	
	// Add item using domain service
	updatedCart, err := h.cartService.AddItem(ctx, shoppingCart.ID, cart.AddItemRequest{
		ProductID:  req.ProductID,
		VariantID:  req.VariantID,
		Quantity:   req.Quantity,
		Attributes: req.Attributes,
	})
	if err != nil {
		if err == cart.ErrOutOfStock {
			http.Error(w, "Product out of stock", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Convert domain model to response
	response := convertCartToResponse(updatedCart)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetCart shows how a GET /cart endpoint would work.
func (h *CartHandler) HandleGetCart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	
	// Get cart
	shoppingCart, err := h.cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}
	
	response := convertCartToResponse(shoppingCart)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// OrderHandler demonstrates how an HTTP API would use order services.
type OrderHandler struct {
	cartService    cart.Service
	orderService   orders.Service
	pricingService pricing.Service
}

// CreateOrderRequest is the HTTP request body.
type CreateOrderRequest struct {
	ShippingAddress  AddressRequest `json:"shipping_address"`
	BillingAddress   *AddressRequest `json:"billing_address,omitempty"`
	PaymentMethodID  string         `json:"payment_method_id"`
	PromotionCodes   []string       `json:"promotion_codes,omitempty"`
	ShippingMethodID string         `json:"shipping_method_id"`
	Notes            string         `json:"notes,omitempty"`
}

type AddressRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	AddressLine1 string `json:"address_line_1"`
	AddressLine2 string `json:"address_line_2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	Phone        string `json:"phone"`
}

type OrderResponse struct {
	ID           string  `json:"id"`
	OrderNumber  string  `json:"order_number"`
	Status       string  `json:"status"`
	Total        string  `json:"total"`
	Subtotal     string  `json:"subtotal"`
	Tax          string  `json:"tax"`
	Shipping     string  `json:"shipping"`
	ItemCount    int     `json:"item_count"`
	CreatedAt    string  `json:"created_at"`
}

// HandleCreateOrder shows how a POST /orders endpoint would work.
func (h *OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	
	// Parse request
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.PaymentMethodID == "" || req.ShippingMethodID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	
	// Get user's cart
	shoppingCart, err := h.cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}
	
	if shoppingCart.IsEmpty() {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}
	
	// Convert request addresses to domain addresses
	shippingAddr := orders.Address{
		FirstName:    req.ShippingAddress.FirstName,
		LastName:     req.ShippingAddress.LastName,
		AddressLine1: req.ShippingAddress.AddressLine1,
		AddressLine2: req.ShippingAddress.AddressLine2,
		City:         req.ShippingAddress.City,
		State:        req.ShippingAddress.State,
		PostalCode:   req.ShippingAddress.PostalCode,
		Country:      req.ShippingAddress.Country,
		Phone:        req.ShippingAddress.Phone,
	}
	
	billingAddr := shippingAddr
	if req.BillingAddress != nil {
		billingAddr = orders.Address{
			FirstName:    req.BillingAddress.FirstName,
			LastName:     req.BillingAddress.LastName,
			AddressLine1: req.BillingAddress.AddressLine1,
			AddressLine2: req.BillingAddress.AddressLine2,
			City:         req.BillingAddress.City,
			State:        req.BillingAddress.State,
			PostalCode:   req.BillingAddress.PostalCode,
			Country:      req.BillingAddress.Country,
			Phone:        req.BillingAddress.Phone,
		}
	}
	
	// Create order using domain service
	order, err := h.orderService.CreateFromCart(ctx, orders.CreateOrderRequest{
		Cart:             shoppingCart,
		UserID:           userID,
		ShippingAddress:  shippingAddr,
		BillingAddress:   billingAddr,
		PaymentMethodID:  req.PaymentMethodID,
		PromotionCodes:   req.PromotionCodes,
		ShippingMethodID: req.ShippingMethodID,
		Notes:            req.Notes,
		IPAddress:        getIPAddress(r),
		UserAgent:        r.UserAgent(),
	})
	if err != nil {
		if err == orders.ErrEmptyCart {
			http.Error(w, "Cart is empty", http.StatusBadRequest)
			return
		}
		if err == orders.ErrPaymentFailed {
			http.Error(w, "Payment failed", http.StatusPaymentRequired)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Clear cart after successful order
	_, _ = h.cartService.Clear(ctx, shoppingCart.ID)
	
	// Convert to response
	response := OrderResponse{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		Status:      string(order.Status),
		Total:       order.Total.String(),
		Subtotal:    order.Subtotal.String(),
		Tax:         order.TaxTotal.String(),
		Shipping:    order.ShippingTotal.String(),
		ItemCount:   order.ItemCount(),
		CreatedAt:   order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CheckoutHandler shows a complete checkout flow.
type CheckoutHandler struct {
	cartService    cart.Service
	pricingService pricing.Service
}

type CheckoutPreviewRequest struct {
	PromotionCodes   []string       `json:"promotion_codes,omitempty"`
	ShippingMethodID string         `json:"shipping_method_id"`
	ShippingAddress  AddressRequest `json:"shipping_address"`
}

type CheckoutPreviewResponse struct {
	Subtotal      string                  `json:"subtotal"`
	Discount      string                  `json:"discount"`
	Tax           string                  `json:"tax"`
	Shipping      string                  `json:"shipping"`
	Total         string                  `json:"total"`
	Discounts     []DiscountResponse      `json:"discounts"`
	TaxLines      []TaxLineResponse       `json:"tax_lines"`
}

type DiscountResponse struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Amount string `json:"amount"`
}

type TaxLineResponse struct {
	Name   string  `json:"name"`
	Rate   float64 `json:"rate"`
	Amount string  `json:"amount"`
}

// HandleCheckoutPreview calculates pricing before order creation.
// This is useful for showing the user the total before they confirm.
func (h *CheckoutHandler) HandleCheckoutPreview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	
	var req CheckoutPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	// Get cart
	shoppingCart, err := h.cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		http.Error(w, "Failed to get cart", http.StatusInternalServerError)
		return
	}
	
	// Calculate pricing
	shippingAddr := &pricing.Address{
		Country:    req.ShippingAddress.Country,
		State:      req.ShippingAddress.State,
		City:       req.ShippingAddress.City,
		PostalCode: req.ShippingAddress.PostalCode,
	}
	
	pricingResult, err := h.pricingService.PriceCart(ctx, pricing.PriceCartRequest{
		Cart:             shoppingCart,
		PromotionCodes:   req.PromotionCodes,
		ShippingMethodID: &req.ShippingMethodID,
		ShippingAddress:  shippingAddr,
		TaxInclusive:     false,
	})
	if err != nil {
		http.Error(w, "Failed to calculate pricing", http.StatusInternalServerError)
		return
	}
	
	// Convert to response
	response := CheckoutPreviewResponse{
		Subtotal: pricingResult.Subtotal.String(),
		Discount: pricingResult.DiscountTotal.String(),
		Tax:      pricingResult.TaxTotal.String(),
		Shipping: pricingResult.ShippingTotal.String(),
		Total:    pricingResult.Total.String(),
		Discounts: make([]DiscountResponse, len(pricingResult.AppliedDiscounts)),
		TaxLines:  make([]TaxLineResponse, len(pricingResult.TaxLines)),
	}
	
	for i, d := range pricingResult.AppliedDiscounts {
		response.Discounts[i] = DiscountResponse{
			Code:   d.Code,
			Name:   d.Name,
			Amount: d.Amount.String(),
		}
	}
	
	for i, t := range pricingResult.TaxLines {
		response.TaxLines[i] = TaxLineResponse{
			Name:   t.Name,
			Rate:   t.Rate,
			Amount: t.Amount.String(),
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper functions (these would be in your HTTP/auth layer)

func getUserIDFromContext(ctx context.Context) string {
	// Extract from JWT or session
	// This is handled by your auth middleware
	return "user-123"
}

func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header, etc.
	return r.RemoteAddr
}

func convertCartToResponse(c *cart.Cart) CartResponse {
	items := make([]CartItemResponse, len(c.Items))
	for i, item := range c.Items {
		items[i] = CartItemResponse{
			ID:       item.ID,
			Name:     item.Name,
			SKU:      item.SKU,
			Price:    item.Price.String(),
			Quantity: item.Quantity,
		}
	}
	
	return CartResponse{
		ID:        c.ID,
		ItemCount: c.ItemCount(),
		Subtotal:  c.Subtotal().String(),
		Items:     items,
	}
}

// Example: Setting up HTTP routes (pseudo-code)
func ExampleHTTPRoutes() {
	fmt.Print(`
// In your HTTP server setup (e.g., with Gin, Echo, or net/http):

func SetupRoutes(
	cartService cart.Service,
	orderService orders.Service,
	pricingService pricing.Service,
) {
	cartHandler := &CartHandler{
		cartService:    cartService,
		pricingService: pricingService,
	}
	
	orderHandler := &OrderHandler{
		cartService:    cartService,
		orderService:   orderService,
		pricingService: pricingService,
	}
	
	checkoutHandler := &CheckoutHandler{
		cartService:    cartService,
		pricingService: pricingService,
	}
	
	// Cart routes
	http.HandleFunc("GET /cart", cartHandler.HandleGetCart)
	http.HandleFunc("POST /cart/items", cartHandler.HandleAddToCart)
	
	// Checkout routes
	http.HandleFunc("POST /checkout/preview", checkoutHandler.HandleCheckoutPreview)
	
	// Order routes
	http.HandleFunc("POST /orders", orderHandler.HandleCreateOrder)
}
`)
}
