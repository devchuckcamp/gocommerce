# E-Commerce Domain Library

A pure Go domain library for e-commerce applications. No HTTP frameworks, no database dependencies - just clean domain logic.

## 🎯 Goals

- **Pure Domain Logic**: Only business rules, no infrastructure code
- **Framework Agnostic**: Use with any HTTP framework (Gin, Echo, Chi, net/http)
- **Database Agnostic**: Plug in any database via repository interfaces
- **DDD Patterns**: Entities, value objects, aggregates, domain services
- **Production Ready**: Handles edge cases, validation, money calculations correctly
- **Reusable**: Build monoliths, microservices, or serverless functions

## 📦 Package Structure

```
github.com/myorg/ecommerce-domain/
├── money/          # Money value object (no floating-point errors!)
├── catalog/        # Products, variants, categories, brands
├── cart/           # Shopping cart with CartService
├── pricing/        # Pricing engine (discounts, tax, shipping)
├── orders/         # Order management with OrderService
├── inventory/      # Stock management interfaces
├── payments/       # Payment gateway interfaces
├── shipping/       # Shipping rate calculation interfaces
├── tax/            # Tax calculation interfaces
├── user/           # User profiles and addresses
└── examples/       # Usage examples and HTTP handler patterns
```

## ✨ Key Features

### Domain Services Implemented

✅ **CartService**: Add items, update quantities, merge carts, stock validation  
✅ **PricingService**: Calculate totals with promotions, tax, shipping  
✅ **OrderService**: Create orders from carts, manage status transitions  

### Repository Interfaces Defined

You implement these for your database:
- Product, Variant, Category, Brand repositories
- Cart, Order repositories
- Promotion repository
- User, Address repositories

### Service Interfaces Defined

You implement these for external services:
- `inventory.Service` - Stock management
- `payments.Gateway` - Payment processing (Stripe, PayPal, etc.)
- `shipping.RateCalculator` - Shipping rates (FedEx, UPS, etc.)
- `tax.Calculator` - Tax calculation (TaxJar, Avalara, or simple)

## 🚀 Quick Start

### Installation

```bash
go get github.com/devchuckcamp/gocommerce
```

### Basic Usage

```go
import (
    "github.com/devchuckcamp/gocommerce/cart"
    "github.com/devchuckcamp/gocommerce/money"
    "github.com/devchuckcamp/gocommerce/orders"
    "github.com/devchuckcamp/gocommerce/pricing"
)

// 1. Work with money (no floating-point errors)
price, _ := money.NewFromFloat(19.99, "USD")
discount := price.Multiply(0.10)
final, _ := price.Subtract(discount)

// 2. Create cart service
cartService := cart.NewCartService(
    cartRepo,        // Your implementation
    productRepo,     // Your implementation
    variantRepo,     // Your implementation
    inventoryService, // Your implementation (optional)
    generateID,
)

// 3. Add items to cart
shoppingCart, _ := cartService.GetOrCreateCart(ctx, userID, "")
shoppingCart, _ = cartService.AddItem(ctx, shoppingCart.ID, cart.AddItemRequest{
    ProductID: "prod-123",
    Quantity:  2,
})

// 4. Calculate pricing with discounts
pricingService := pricing.NewPricingService(promotionRepo, taxCalc, shippingCalc)
result, _ := pricingService.PriceCart(ctx, pricing.PriceCartRequest{
    Cart:           shoppingCart,
    PromotionCodes: []string{"SAVE10"},
    ShippingAddress: address,
})

// 5. Create order
orderService := orders.NewOrderService(orderRepo, pricingService, inventoryService, paymentGateway, genOrderNum, genID)
order, _ := orderService.CreateFromCart(ctx, orders.CreateOrderRequest{
    Cart:            shoppingCart,
    ShippingAddress: address,
    PaymentMethodID: "pm_123",
})
```

## 📚 Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Get started in 5 minutes
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Detailed design patterns and decisions
- **[PACKAGE_SUMMARY.md](PACKAGE_SUMMARY.md)** - Visual guide to all packages
- **[examples/usage.go](examples/usage.go)** - Domain usage examples
- **[examples/http_handlers.go](examples/http_handlers.go)** - HTTP integration patterns

## 🏗️ Architecture Highlights

### Value Objects

```go
type Money struct {
    Amount   int64  // Cents, not dollars (avoid float errors)
    Currency string
}
```

### Aggregates

```go
type Cart struct {
    ID    string
    Items []CartItem  // Only Cart can modify items
}

func (c *Cart) AddItem(item CartItem)  // Enforces invariants
```

### Domain Services

```go
type CartService struct {
    repo        Repository
    productRepo catalog.ProductRepository
    inventory   inventory.Service
}

func (s *CartService) AddItem(ctx, cartID, req) (*Cart, error) {
    // Validates product, checks stock, updates cart
}
```

### Repository Pattern

```go
type ProductRepository interface {
    FindByID(ctx context.Context, id string) (*Product, error)
    Save(ctx context.Context, product *Product) error
}
```

## 🔌 Integration

### Monolith

```go
func main() {
    db := postgres.Connect()
    cartRepo := postgres.NewCartRepository(db)
    cartService := cart.NewCartService(cartRepo, ...)
    
    http.HandleFunc("/cart/items", handlers.AddToCart(cartService))
    http.ListenAndServe(":8080", nil)
}
```

### Microservices

```go
// Cart Service
func main() {
    cartService := cart.NewCartService(
        postgres.NewCartRepo(db),
        grpc.NewProductClient(),  // RPC to catalog service
        grpc.NewInventoryClient(), // RPC to inventory service
        generateID,
    )
    grpc.ServeCartService(cartService)
}
```

## ✅ What You Get

- **11 domain packages** with complete e-commerce logic
- **Zero external dependencies** (only Go standard library)
- **Production-ready** money handling (no float errors)
- **Flexible architecture** (monolith or microservices)
- **Clean separation** of concerns (domain vs. infrastructure)
- **Easy testing** (mock interfaces, test domain logic)
- **Well-documented** (godoc comments, examples, guides)

## 📋 Design Principles

- **Pure Domain Logic**: No external dependencies, only standard library
- **Interface-Driven**: Repository and service interfaces for flexibility
- **Value Objects**: Immutable types like Money for correctness
- **DDD Patterns**: Entities, aggregates, domain services
- **Reusable**: Works in monoliths or microservices

## 🧪 Testing

```go
// Unit test domain logic
func TestCart_AddItem(t *testing.T) {
    cart := &cart.Cart{Items: []cart.CartItem{}}
    cart.AddItem(item)
    assert.Equal(t, 1, len(cart.Items))
}

// Integration test with mocks
func TestCartService_AddItem(t *testing.T) {
    service := cart.NewCartService(mockRepo, mockProductRepo, nil, genID)
    cart, err := service.AddItem(ctx, cartID, req)
    assert.NoError(t, err)
}
```

## 🤝 Contributing

This is a reference implementation. Feel free to:
- Use it as-is in your projects
- Modify for your needs
- Extract patterns for your own domain libraries

## 📄 License

MIT License - Use freely in commercial or personal projects

## 🎓 Learn More

This library demonstrates:
- Domain-Driven Design (DDD) in Go
- Hexagonal Architecture (Ports & Adapters)
- Repository Pattern
- Service Layer Pattern
- Value Objects and Aggregates
- Interface Segregation Principle

Perfect for building production e-commerce systems or learning Go architecture patterns!
