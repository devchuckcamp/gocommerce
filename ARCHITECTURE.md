# E-Commerce Domain Library Architecture

## Overview

This library provides pure domain logic for e-commerce applications using Domain-Driven Design (DDD) principles. It contains no HTTP, no database code, no external dependencies - just business logic and interfaces.

## Package Structure

```
github.com/devchuckcamp/gocommerce/
├── money/          # Money value object (handles currency correctly)
├── catalog/        # Product catalog domain
├── cart/           # Shopping cart domain
├── pricing/        # Pricing engine (discounts, tax, shipping)
├── orders/         # Order management domain
├── inventory/      # Inventory management interfaces
├── payments/       # Payment gateway interfaces
├── shipping/       # Shipping rate calculation interfaces
├── tax/            # Tax calculation interfaces
├── user/           # User profile and address domain
└── examples/       # Usage examples (not part of library)
```

## Core Concepts

### 1. Money Value Object (`money/`)

**Why**: Avoid floating-point errors in financial calculations.

```go
// Always use Money for prices, never float64
price, _ := money.NewFromFloat(19.99, "USD")
discount := price.Multiply(0.10)  // 10% off
final, _ := price.Subtract(discount)
```

**Key features**:
- Stores amounts in cents (int64) to avoid floating-point issues
- Currency-safe: can't add USD + EUR
- Allocation method for splitting amounts (handles remainders correctly)
- Immutable value object

### 2. Catalog Domain (`catalog/`)

**Entities**: Product, Variant, Category, Brand

**Purpose**: Manage product information.

```go
type Product struct {
    ID          string
    SKU         string
    Name        string
    BasePrice   money.Money
    Status      ProductStatus
    BrandID     string
    CategoryID  string
    // ...
}
```

**Repositories** (interfaces only):
- `ProductRepository`: Find products by ID, SKU, category, search
- `VariantRepository`: Manage product variants
- `CategoryRepository`: Category tree operations
- `BrandRepository`: Brand management

### 3. Cart Domain (`cart/`)

**Aggregate Root**: Cart

**Purpose**: Shopping cart management with business rules.

```go
type Cart struct {
    ID         string
    UserID     string
    Items      []CartItem
    // ...
}

// Domain methods
func (c *Cart) AddItem(item CartItem)
func (c *Cart) UpdateItemQuantity(itemID string, qty int)
func (c *Cart) Merge(other *Cart)  // For guest->user migration
func (c *Cart) Subtotal() money.Money
```

**Service**: `CartService`
- Orchestrates cart operations
- Validates stock availability
- Manages guest/user cart transitions

### 4. Pricing Domain (`pricing/`)

**Purpose**: Calculate totals with discounts, tax, and shipping.

**Core Service**: `PricingService.PriceCart()`

```go
type PricingResult struct {
    Subtotal         money.Money
    DiscountTotal    money.Money
    TaxTotal         money.Money
    ShippingTotal    money.Money
    Total            money.Money
    AppliedDiscounts []AppliedDiscount
    TaxLines         []TaxLine
}
```

**Promotion Engine**:
- Percentage discounts
- Fixed amount discounts
- Buy X Get Y
- Minimum purchase requirements
- Product/category restrictions

**Dependencies**:
- `TaxCalculator` interface (from `tax/`)
- `ShippingRateCalculator` interface (from `shipping/`)
- `PromotionRepository` interface

### 5. Orders Domain (`orders/`)

**Aggregate Root**: Order

**Purpose**: Order lifecycle management.

```go
type Order struct {
    ID              string
    OrderNumber     string
    Status          OrderStatus
    Items           []OrderItem
    ShippingAddress Address
    Total           money.Money
    // ...
}

// Status transitions with validation
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool
func (o *Order) UpdateStatus(newStatus OrderStatus) bool
```

**Service**: `OrderService`
- Creates orders from carts
- Reserves inventory
- Processes payments
- Manages status transitions

**Status Flow**:
```
Pending → Paid → Processing → Shipped → Delivered
   ↓        ↓         ↓
Canceled  Refunded  Canceled
```

### 6. Inventory Domain (`inventory/`)

**Purpose**: Stock management with reservations.

```go
type Service interface {
    GetAvailableStock(ctx, sku string) (int, error)
    Reserve(ctx, sku string, qty int, refID string) error
    Release(ctx, sku string, qty int, refID string) error
    Commit(ctx, referenceID string) error
}
```

**Reservation Pattern**:
1. Reserve stock when adding to cart or creating order
2. Release if cart abandoned or order canceled
3. Commit when order is paid/confirmed

### 7. Payments Domain (`payments/`)

**Purpose**: Payment gateway abstraction.

```go
type Gateway interface {
    CreateIntent(ctx, req IntentRequest) (*PaymentIntent, error)
    CaptureIntent(ctx, intentID string) (*PaymentIntent, error)
    CreateRefund(ctx, req RefundRequest) (*Refund, error)
}
```

**Payment Intent Pattern**:
- Create intent (authorize)
- Capture later (two-phase commit)
- Support for refunds

**Implementations**: Stripe, PayPal, etc. (in separate packages)

### 8. Shipping Domain (`shipping/`)

**Purpose**: Shipping rate calculation.

```go
type RateCalculator interface {
    GetRate(ctx, req RateRequest) (*ShippingRate, error)
    GetAvailableRates(ctx, req RateRequest) ([]*ShippingRate, error)
}

type ShippingRate struct {
    MethodID      string
    MethodName    string
    Cost          money.Money
    EstimatedDays int
    Carrier       string
}
```

**Implementations**: FedEx, UPS, flat-rate, etc. (in separate packages)

### 9. Tax Domain (`tax/`)

**Purpose**: Tax calculation (sales tax, VAT, GST).

```go
type Calculator interface {
    Calculate(ctx, req CalculationRequest) (*CalculationResult, error)
    GetRatesForAddress(ctx, address Address) ([]TaxRate, error)
}
```

**Features**:
- Tax-inclusive and tax-exclusive pricing
- Multiple tax rates (state, county, city)
- Compound taxes
- Line-item and shipping tax

**Implementations**: TaxJar, Avalara, simple table-based, etc. (in separate packages)

### 10. User Domain (`user/`)

**Purpose**: User profiles and saved addresses.

```go
type UserProfile struct {
    ID        string
    Email     string
    FirstName string
    LastName  string
}

type Address struct {
    ID           string
    UserID       string
    Label        string  // "Home", "Work"
    AddressLine1 string
    IsDefault    bool
}
```

**Note**: Authentication is handled separately (not in this library).

## Design Patterns

### 1. Repository Pattern

All persistence is abstracted through interfaces:

```go
type ProductRepository interface {
    FindByID(ctx context.Context, id string) (*Product, error)
    Save(ctx context.Context, product *Product) error
}
```

**Benefits**:
- Swap database implementations (Postgres, MySQL, MongoDB)
- Easy testing with mocks
- No SQL in domain logic

### 2. Service Pattern

Complex operations that span multiple aggregates:

```go
type CartService struct {
    repo             Repository
    productRepo      catalog.ProductRepository
    inventoryService inventory.Service
}

func (s *CartService) AddItem(ctx, cartID string, req AddItemRequest) (*Cart, error) {
    // Fetch cart, validate product, check stock, update cart
}
```

### 3. Value Objects

Immutable types that represent concepts:

```go
type Money struct {
    Amount   int64   // Immutable
    Currency string  // Immutable
}

// Returns new Money, doesn't mutate
func (m Money) Add(other Money) (Money, error)
```

### 4. Aggregate Roots

Entities that control access to related entities:

```go
type Cart struct {  // Aggregate root
    Items []CartItem  // Only Cart can modify Items
}

// Public methods enforce invariants
func (c *Cart) AddItem(item CartItem) {
    // Business rules enforced here
}
```

## Usage Flow

### Complete Checkout Example

```go
// 1. User adds items to cart
cart, _ := cartService.GetOrCreateCart(ctx, userID, "")
cart, _ = cartService.AddItem(ctx, cart.ID, AddItemRequest{
    ProductID: "prod_123",
    Quantity: 2,
})

// 2. Calculate pricing
pricing, _ := pricingService.PriceCart(ctx, PriceCartRequest{
    Cart: cart,
    PromotionCodes: []string{"SAVE10"},
    ShippingAddress: address,
})

// 3. Create order
order, _ := orderService.CreateFromCart(ctx, CreateOrderRequest{
    Cart: cart,
    UserID: userID,
    ShippingAddress: address,
    PaymentMethodID: "pm_123",
})

// 4. Order service handles:
//    - Inventory reservation
//    - Payment processing
//    - Order creation
//    - Cart clearing
```

## Integration Points

### In a Monolith

```go
// main.go
func main() {
    // Database layer
    db := postgres.Connect()
    productRepo := postgres.NewProductRepository(db)
    cartRepo := postgres.NewCartRepository(db)
    
    // Domain services
    cartService := cart.NewCartService(cartRepo, productRepo, ...)
    pricingService := pricing.NewPricingService(...)
    orderService := orders.NewOrderService(...)
    
    // HTTP layer
    handlers.SetupRoutes(cartService, orderService, pricingService)
    
    http.ListenAndServe(":8080", nil)
}
```

### In Microservices

**Cart Service**:
```go
// services/cart-service/main.go
func main() {
    db := postgres.Connect()
    cartRepo := postgres.NewCartRepository(db)
    productClient := grpc.NewProductClient()  // RPC to catalog service
    
    cartService := cart.NewCartService(
        cartRepo,
        productClient,  // Implements catalog.ProductRepository
        nil,            // No inventory service in this service
        generateID,
    )
    
    grpc.ServeCartService(cartService)
}
```

**Order Service**:
```go
// services/order-service/main.go
func main() {
    orderRepo := postgres.NewOrderRepository(db)
    cartClient := grpc.NewCartClient()
    pricingClient := grpc.NewPricingClient()
    
    orderService := orders.NewOrderService(
        orderRepo,
        pricingClient,  // Implements pricing.Service
        nil,
        nil,
        generateOrderNumber,
        generateID,
    )
}
```

## Testing

### Unit Testing Domain Logic

```go
func TestCart_AddItem(t *testing.T) {
    cart := &cart.Cart{ID: "1", Items: []cart.CartItem{}}
    
    item := cart.CartItem{
        ID:       "item1",
        ProductID: "prod1",
        Quantity: 2,
    }
    
    cart.AddItem(item)
    
    assert.Equal(t, 1, len(cart.Items))
    assert.Equal(t, 2, cart.ItemCount())
}
```

### Testing Services with Mocks

```go
type mockProductRepo struct{}

func (m *mockProductRepo) FindByID(ctx context.Context, id string) (*catalog.Product, error) {
    return &catalog.Product{ID: id, Name: "Test Product"}, nil
}

func TestCartService_AddItem(t *testing.T) {
    mockRepo := &mockCartRepo{}
    mockProductRepo := &mockProductRepo{}
    
    service := cart.NewCartService(mockRepo, mockProductRepo, nil, generateID)
    
    cart, err := service.AddItem(ctx, "cart1", cart.AddItemRequest{
        ProductID: "prod1",
        Quantity: 1,
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, cart)
}
```

## Extension Points

### Custom Tax Calculator

```go
type MyTaxCalculator struct {
    apiKey string
}

func (c *MyTaxCalculator) Calculate(ctx context.Context, req tax.CalculationRequest) (*tax.CalculationResult, error) {
    // Call external tax API
    // Or implement custom logic
}

// Use in pricing service
pricingService := pricing.NewPricingService(
    promotionRepo,
    &MyTaxCalculator{apiKey: "..."},
    shippingCalc,
)
```

### Custom Payment Gateway

```go
type StripeGateway struct {
    client *stripe.Client
}

func (g *StripeGateway) CreateIntent(ctx context.Context, req payments.IntentRequest) (*payments.PaymentIntent, error) {
    // Implement Stripe integration
}

// Use in order service
orderService := orders.NewOrderService(
    orderRepo,
    pricingService,
    inventoryService,
    &StripeGateway{client: stripeClient},
    generateOrderNumber,
    generateID,
)
```

## Best Practices

1. **Always use `money.Money`** for prices, never `float64`
2. **Pass `context.Context`** as first parameter to all repository methods
3. **Use interfaces** for all external dependencies (repositories, services)
4. **Keep entities simple** - business logic goes in domain services
5. **Validate at service boundaries** - don't trust input
6. **Use value objects** for concepts like Money, Address
7. **Enforce invariants** in aggregate roots
8. **Return errors, don't panic** - errors are part of the domain
9. **Make zero values useful** when possible
10. **Document public APIs** with godoc comments

## Migration Guide

### From Existing Monolith

1. **Identify domain boundaries** in current code
2. **Extract value objects** first (Money, Address)
3. **Define repository interfaces** for current data access
4. **Move business logic** to domain services
5. **Keep HTTP handlers thin** - just request/response conversion
6. **Test incrementally** - one domain at a time

### To Microservices

1. **Each service owns its database** - no shared DB
2. **Implement gRPC adapters** that implement repository interfaces
3. **Use events** for cross-service communication (not covered here)
4. **Handle eventual consistency** - compensating transactions
5. **Service boundaries** should match domain boundaries

## Summary

This library provides:

✅ Pure domain logic - reusable across any architecture  
✅ Interface-driven design - swap implementations easily  
✅ DDD patterns - entities, value objects, aggregates, services  
✅ No dependencies - only standard library  
✅ Production-ready - handles edge cases, validation, errors  
✅ Testable - easy to mock and unit test  
✅ Documented - godoc comments and examples  

You can use this in:
- Monolithic applications
- Microservices
- Serverless functions
- CLI tools
- Background workers

Just plug in your infrastructure (HTTP, DB, messaging) around it!
