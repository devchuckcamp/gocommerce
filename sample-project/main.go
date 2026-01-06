package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/orders"
	"github.com/devchuckcamp/gocommerce/pricing"

	"sample-ecommerce-api/postgres"
)

func main() {
	usePostgres := os.Getenv("USE_POSTGRES") == "1" || os.Getenv("DATABASE_URL") != ""
	runMigrations := os.Getenv("RUN_MIGRATIONS") == "1"

	var productStore ProductStore
	var cartRepo cart.Repository
	var productRepo catalog.ProductRepository
	var variantRepo catalog.VariantRepository
	var orderRepo orders.Repository
	var promotionRepo pricing.PromotionRepository

	if usePostgres {
		db, err := postgres.Open()
		if err != nil {
			log.Fatalf("failed to open postgres: %v", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}

		if runMigrations {
			if err := postgres.RunMigrations(context.Background(), db); err != nil {
				log.Fatalf("failed to run migrations: %v", err)
			}
		}

		pg := postgres.NewStore(db)
		defer pg.Close()
		productStore = pg
		cartRepo = pg.Carts
		productRepo = pg.Products
		variantRepo = pg.Variants
		orderRepo = pg.Orders
		promotionRepo = pg.Promotions
	} else {
		store := NewMemoryStore()
		seedProducts(store)
		productStore = store
		cartRepo = &store.cartRepo
		productRepo = store
		variantRepo = &store.variantRepo
		orderRepo = &store.orderRepo
		promotionRepo = &store.promotionRepo
	}

	// Create domain services
	cartService := cart.NewCartService(
		cartRepo,
		productRepo,
		variantRepo,
		nil, // No inventory service for demo
		generateID,
	)

	pricingService := pricing.NewPricingService(
		promotionRepo,
		NewSimpleTaxCalculator(0.0875), // 8.75% tax
		nil, // No shipping calculator for demo
	)

	orderService := orders.NewOrderService(
		orderRepo,
		pricingService,
		nil, // No inventory service
		nil, // No payment gateway
		generateOrderNumber,
		generateID,
	)
	
	// Create HTTP handlers
	api := &API{
		store:          productStore,
		cartService:    cartService,
		pricingService: pricingService,
		orderService:   orderService,
	}
	
	// Setup routes
	http.HandleFunc("/products", api.handleProducts)
	http.HandleFunc("/products/", api.handleProductDetail)
	http.HandleFunc("/cart", api.handleCart)
	http.HandleFunc("/cart/items", api.handleCartItems)
	http.HandleFunc("/cart/items/", api.handleCartItem)
	http.HandleFunc("/checkout/preview", api.handleCheckoutPreview)
	http.HandleFunc("/orders", api.handleOrders)
	
	// Start server
	fmt.Println("🚀 E-Commerce API Server")
	fmt.Println("========================")
	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET    /products")
	fmt.Println("  GET    /products/:id")
	fmt.Println("  GET    /cart")
	fmt.Println("  POST   /cart/items")
	fmt.Println("  PUT    /cart/items/:id")
	fmt.Println("  DELETE /cart/items/:id")
	fmt.Println("  DELETE /cart")
	fmt.Println("  POST   /checkout/preview")
	fmt.Println("  POST   /orders")
	fmt.Println("\nAdd header: user-id: user-123")
	fmt.Println()
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type API struct {
	store          ProductStore
	cartService    cart.Service
	pricingService pricing.Service
	orderService   orders.Service
}

// Product handlers

func (api *API) handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	products, err := api.store.ListProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondJSON(w, products)
}

func (api *API) handleProductDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	product, err := api.store.FindProductByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	
	respondJSON(w, product)
}

// Cart handlers

func (api *API) handleCart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user-id")
	if userID == "" {
		http.Error(w, "user-id header required", http.StatusBadRequest)
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		shoppingCart, err := api.cartService.GetOrCreateCart(r.Context(), userID, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, shoppingCart)
		
	case http.MethodDelete:
		shoppingCart, err := api.cartService.GetOrCreateCart(r.Context(), userID, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = api.cartService.Clear(r.Context(), shoppingCart.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		respondJSON(w, map[string]string{"message": "Cart cleared"})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (api *API) handleCartItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID := r.Header.Get("user-id")
	if userID == "" {
		http.Error(w, "user-id header required", http.StatusBadRequest)
		return
	}
	
	var req struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	shoppingCart, err := api.cartService.GetOrCreateCart(r.Context(), userID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	updatedCart, err := api.cartService.AddItem(r.Context(), shoppingCart.ID, cart.AddItemRequest{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	respondJSON(w, updatedCart)
}

func (api *API) handleCartItem(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("user-id")
	if userID == "" {
		http.Error(w, "user-id header required", http.StatusBadRequest)
		return
	}
	
	itemID := strings.TrimPrefix(r.URL.Path, "/cart/items/")
	
	shoppingCart, err := api.cartService.GetOrCreateCart(r.Context(), userID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	switch r.Method {
	case http.MethodPut:
		var req struct {
			Quantity int `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		updatedCart, err := api.cartService.UpdateItemQuantity(r.Context(), shoppingCart.ID, itemID, req.Quantity)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondJSON(w, updatedCart)
		
	case http.MethodDelete:
		updatedCart, err := api.cartService.RemoveItem(r.Context(), shoppingCart.ID, itemID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		respondJSON(w, updatedCart)
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Checkout handlers

func (api *API) handleCheckoutPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID := r.Header.Get("user-id")
	if userID == "" {
		http.Error(w, "user-id header required", http.StatusBadRequest)
		return
	}
	
	var req struct {
		ShippingAddress struct {
			Country    string `json:"country"`
			State      string `json:"state"`
			City       string `json:"city"`
			PostalCode string `json:"postal_code"`
		} `json:"shipping_address"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	shoppingCart, err := api.cartService.GetOrCreateCart(r.Context(), userID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if shoppingCart.IsEmpty() {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}
	
	result, err := api.pricingService.PriceCart(r.Context(), pricing.PriceCartRequest{
		Cart: shoppingCart,
		ShippingAddress: &pricing.Address{
			Country:    req.ShippingAddress.Country,
			State:      req.ShippingAddress.State,
			City:       req.ShippingAddress.City,
			PostalCode: req.ShippingAddress.PostalCode,
		},
		TaxInclusive: false,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, result)
}

// Order handlers

func (api *API) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userID := r.Header.Get("user-id")
	if userID == "" {
		http.Error(w, "user-id header required", http.StatusBadRequest)
		return
	}
	
	var req struct {
		ShippingAddress struct {
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			AddressLine1 string `json:"address_line_1"`
			AddressLine2 string `json:"address_line_2"`
			City         string `json:"city"`
			State        string `json:"state"`
			PostalCode   string `json:"postal_code"`
			Country      string `json:"country"`
			Phone        string `json:"phone"`
		} `json:"shipping_address"`
		PaymentMethodID string `json:"payment_method_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	shoppingCart, err := api.cartService.GetOrCreateCart(r.Context(), userID, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	if shoppingCart.IsEmpty() {
		http.Error(w, "Cart is empty", http.StatusBadRequest)
		return
	}
	
	order, err := api.orderService.CreateFromCart(r.Context(), orders.CreateOrderRequest{
		Cart:   shoppingCart,
		UserID: userID,
		ShippingAddress: orders.Address{
			FirstName:    req.ShippingAddress.FirstName,
			LastName:     req.ShippingAddress.LastName,
			AddressLine1: req.ShippingAddress.AddressLine1,
			AddressLine2: req.ShippingAddress.AddressLine2,
			City:         req.ShippingAddress.City,
			State:        req.ShippingAddress.State,
			PostalCode:   req.ShippingAddress.PostalCode,
			Country:      req.ShippingAddress.Country,
			Phone:        req.ShippingAddress.Phone,
		},
		BillingAddress:  orders.Address{},
		PaymentMethodID: req.PaymentMethodID,
		IPAddress:       r.RemoteAddr,
		UserAgent:       r.UserAgent(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Clear cart after successful order
	_, _ = api.cartService.Clear(r.Context(), shoppingCart.ID)
	
	respondJSON(w, order)
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func generateID() string {
	return fmt.Sprintf("id-%d", time.Now().UnixNano())
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}
