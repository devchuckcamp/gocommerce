# Quick Start Guide

## Prerequisites

- Go 1.21 or higher
- Docker (for PostgreSQL database)

## Database Setup (Recommended First Step)

```bash
# 1. Start PostgreSQL
cd migrations/examples && docker-compose up -d

# 2. Run migrations
cd postgresql
cd migrations/examples/postgresql
go run main.go

# 3. Seed database with test data
go run seed-products.go
```

This creates:
- 6 tables (brands, categories, products, carts, cart_items, orders)
- 8 brands (Apple, Dell, Lenovo, HP, Samsung, Logitech, Sony, Bose)
- 8 categories (Electronics, Computers, Laptops, Accessories, etc.)
- 72 products (22 curated + 50 random)

## Library Installation

```bash
go get github.com/devchuckcamp/gocommerce
```

## Basic Usage

### 1. Working with Money

```go
import "github.com/devchuckcamp/gocommerce/money"

// Create money values
price, _ := money.NewFromFloat(19.99, "USD")
tax, _ := money.NewFromFloat(1.60, "USD")

// Calculate total
total, _ := price.Add(tax)
fmt.Println(total.String()) // "USD 21.59"

// Apply discount
discount := total.Multiply(0.10)  // 10% off
final, _ := total.Subtract(discount)
```

### 2. Add Items to Cart

```go
import (
    "github.com/devchuckcamp/gocommerce/cart"
    "github.com/devchuckcamp/gocommerce/catalog"
)

// Setup services (you provide implementations)
var (
    cartRepo     cart.Repository           // Your implementation
    productRepo  catalog.ProductRepository // Your implementation
    inventoryService inventory.Service     // Your implementation
)

// Create cart service
cartService := cart.NewCartService(
    cartRepo,
    productRepo,
    variantRepo,
    inventoryService,
    generateID,
)

// Get or create cart
ctx := context.Background()
shoppingCart, err := cartService.GetOrCreateCart(ctx, "user-123", "")

// Add item
updatedCart, err := cartService.AddItem(ctx, shoppingCart.ID, cart.AddItemRequest{
    ProductID: "product-456",
    Quantity:  2,
})

fmt.Printf("Cart has %d items, subtotal: %s\n", 
    updatedCart.ItemCount(), 
    updatedCart.Subtotal().String())
```

### 3. Calculate Pricing

```go
import "github.com/devchuckcamp/gocommerce/pricing"

// Setup pricing service
pricingService := pricing.NewPricingService(
    promotionRepo,  // Your implementation
    taxCalculator,  // Your implementation
    shippingCalc,   // Your implementation
)

// Calculate cart pricing
result, err := pricingService.PriceCart(ctx, pricing.PriceCartRequest{
    Cart: shoppingCart,
    PromotionCodes: []string{"SAVE10"},
    ShippingMethodID: &shippingMethodID,
    ShippingAddress: &pricing.Address{
        Country:    "US",
        State:      "CA",
        PostalCode: "94102",
    },
})

fmt.Printf("Subtotal: %s\n", result.Subtotal.String())
fmt.Printf("Discount: %s\n", result.DiscountTotal.String())
fmt.Printf("Tax:      %s\n", result.TaxTotal.String())
fmt.Printf("Shipping: %s\n", result.ShippingTotal.String())
fmt.Printf("Total:    %s\n", result.Total.String())
```

### 4. Create an Order

```go
import "github.com/devchuckcamp/gocommerce/orders"

// Setup order service
orderService := orders.NewOrderService(
    orderRepo,        // Your implementation
    pricingService,   // From above
    inventoryService, // Your implementation
    paymentGateway,   // Your implementation
    generateOrderNumber,
    generateID,
)

// Create order from cart
order, err := orderService.CreateFromCart(ctx, orders.CreateOrderRequest{
    Cart:   shoppingCart,
    UserID: "user-123",
    ShippingAddress: orders.Address{
        FirstName:    "John",
        LastName:     "Doe",
        AddressLine1: "123 Main St",
        City:         "San Francisco",
        State:        "CA",
        PostalCode:   "94102",
        Country:      "US",
        Phone:        "+1-555-0100",
    },
    BillingAddress:   shippingAddress, // or different
    PaymentMethodID:  "pm_stripe_123",
    PromotionCodes:   []string{"SAVE10"},
    ShippingMethodID: "standard",
})

fmt.Printf("Order created: %s\n", order.OrderNumber)
fmt.Printf("Status: %s\n", order.Status)
fmt.Printf("Total: %s\n", order.Total.String())
```

## Implementing Required Interfaces

You need to provide implementations for repository interfaces. Here's an example:

### Example: PostgreSQL Cart Repository

```go
package postgres

import (
    "context"
    "database/sql"
    "encoding/json"
    
    "github.com/devchuckcamp/gocommerce/cart"
)

type CartRepository struct {
    db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
    return &CartRepository{db: db}
}

func (r *CartRepository) FindByID(ctx context.Context, id string) (*cart.Cart, error) {
    var (
        cartJSON []byte
        cart     cart.Cart
    )
    
    err := r.db.QueryRowContext(ctx, 
        "SELECT data FROM carts WHERE id = $1", id,
    ).Scan(&cartJSON)
    
    if err == sql.ErrNoRows {
        return nil, cart.ErrCartNotFound
    }
    if err != nil {
        return nil, err
    }
    
    if err := json.Unmarshal(cartJSON, &cart); err != nil {
        return nil, err
    }
    
    return &cart, nil
}

func (r *CartRepository) Save(ctx context.Context, c *cart.Cart) error {
    data, err := json.Marshal(c)
    if err != nil {
        return err
    }
    
    _, err = r.db.ExecContext(ctx,
        `INSERT INTO carts (id, user_id, data, updated_at) 
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (id) DO UPDATE 
         SET data = $3, updated_at = NOW()`,
        c.ID, c.UserID, data,
    )
    
    return err
}

func (r *CartRepository) FindByUserID(ctx context.Context, userID string) (*cart.Cart, error) {
    // Similar to FindByID
}
```

### Example: Simple Tax Calculator

```go
package simpletax

import (
    "context"
    
    "github.com/devchuckcamp/gocommerce/money"
    "github.com/devchuckcamp/gocommerce/tax"
)

type SimpleTaxCalculator struct {
    defaultRate float64 // e.g., 0.08 for 8%
}

func NewSimpleTaxCalculator(rate float64) *SimpleTaxCalculator {
    return &SimpleTaxCalculator{defaultRate: rate}
}

func (c *SimpleTaxCalculator) Calculate(
    ctx context.Context, 
    req tax.CalculationRequest,
) (*tax.CalculationResult, error) {
    // Calculate subtotal
    subtotal := money.Zero(req.LineItems[0].Amount.Currency)
    for _, item := range req.LineItems {
        itemTotal := item.Amount.MultiplyInt(item.Quantity)
        subtotal, _ = subtotal.Add(itemTotal)
    }
    
    // Add shipping to taxable amount
    taxableAmount, _ := subtotal.Add(req.ShippingCost)
    
    // Calculate tax
    taxAmount := taxableAmount.Multiply(c.defaultRate)
    
    return &tax.CalculationResult{
        TotalTax: taxAmount,
        TaxRates: []tax.AppliedTaxRate{
            {
                Name:   "Sales Tax",
                Rate:   c.defaultRate,
                Amount: taxAmount,
            },
        },
    }, nil
}

func (c *SimpleTaxCalculator) GetRatesForAddress(
    ctx context.Context, 
    address tax.Address,
) ([]tax.TaxRate, error) {
    return []tax.TaxRate{
        {
            Name: "Default Sales Tax",
            Rate: c.defaultRate,
        },
    }, nil
}
```

## HTTP Integration Example

See `examples/http_handlers.go` for complete examples of HTTP handler integration.

### Minimal HTTP Server

```go
package main

import (
    "net/http"
    
    "github.com/myorg/ecommerce-domain/cart"
    "github.com/myorg/ecommerce-domain/orders"
    "github.com/myorg/ecommerce-domain/pricing"
)

func main() {
    // Initialize database
    db := initDB()
    
    // Create repositories (your implementations)
    cartRepo := postgres.NewCartRepository(db)
    productRepo := postgres.NewProductRepository(db)
    orderRepo := postgres.NewOrderRepository(db)
    promotionRepo := postgres.NewPromotionRepository(db)
    
    // Create services
    cartService := cart.NewCartService(
        cartRepo, productRepo, variantRepo, nil, generateID,
    )
    
    pricingService := pricing.NewPricingService(
        promotionRepo,
        simpletax.NewSimpleTaxCalculator(0.08),
        nil,
    )
    
    orderService := orders.NewOrderService(
        orderRepo, pricingService, nil, nil,
        generateOrderNumber, generateID,
    )
    
    // Setup HTTP handlers
    cartHandler := NewCartHandler(cartService, pricingService)
    orderHandler := NewOrderHandler(cartService, orderService, pricingService)
    
    http.HandleFunc("/cart", cartHandler.HandleGetCart)
    http.HandleFunc("/cart/items", cartHandler.HandleAddToCart)
    http.HandleFunc("/orders", orderHandler.HandleCreateOrder)
    
    http.ListenAndServe(":8080", nil)
}
```

## Testing

### Unit Test Example

```go
package cart_test

import (
    "testing"
    
    "github.com/devchuckcamp/gocommerce/cart"
    "github.com/devchuckcamp/gocommerce/money"
)

func TestCart_AddItem(t *testing.T) {
    c := &cart.Cart{
        ID:    "cart-1",
        Items: []cart.CartItem{},
    }
    
    price, _ := money.New(1999, "USD")
    item := cart.CartItem{
        ID:        "item-1",
        ProductID: "prod-1",
        SKU:       "SKU-001",
        Name:      "Test Product",
        Price:     price,
        Quantity:  2,
    }
    
    c.AddItem(item)
    
    if len(c.Items) != 1 {
        t.Errorf("expected 1 item, got %d", len(c.Items))
    }
    
    if c.ItemCount() != 2 {
        t.Errorf("expected item count 2, got %d", c.ItemCount())
    }
    
    expectedSubtotal, _ := money.New(3998, "USD")
    if !c.Subtotal().Equals(expectedSubtotal) {
        t.Errorf("expected subtotal %s, got %s", 
            expectedSubtotal.String(), c.Subtotal().String())
    }
}
```

## Next Steps

1. **Read `ARCHITECTURE.md`** for detailed design patterns
2. **Check `examples/usage.go`** for more usage patterns
3. **Implement repository interfaces** for your database
4. **Implement service interfaces** for external services (payment, shipping, tax)
5. **Create HTTP handlers** using the domain services
6. **Write tests** for your implementations

## Common Patterns

### Pattern 1: Guest to User Cart Migration

```go
// When user logs in, merge guest cart
func handleLogin(userID, sessionID string) error {
    guestCart, _ := cartService.GetOrCreateCart(ctx, "", sessionID)
    userCart, _ := cartService.GetOrCreateCart(ctx, userID, "")
    
    if !guestCart.IsEmpty() {
        _, err := cartService.MergeCarts(ctx, guestCart.ID, userCart.ID)
        return err
    }
    
    return nil
}
```

### Pattern 2: Checkout Preview (Before Order Creation)

```go
// Show user total before they confirm
func previewCheckout(cart *cart.Cart, address pricing.Address) {
    result, _ := pricingService.PriceCart(ctx, pricing.PriceCartRequest{
        Cart:            cart,
        ShippingAddress: &address,
    })
    
    // Display to user
    fmt.Printf("Your total will be: %s\n", result.Total.String())
}
```

### Pattern 3: Order Status Updates

```go
// Fulfillment worker updates order status
func fulfillOrder(orderID string) error {
    order, _ := orderService.GetOrder(ctx, orderID)
    
    // Ship the order
    order, _ = orderService.UpdateStatus(ctx, orderID, orders.OrderStatusShipped)
    
    // Later: mark as delivered
    order, _ = orderService.UpdateStatus(ctx, orderID, orders.OrderStatusDelivered)
    
    return nil
}
```

## Get Help

- Read the godoc comments in each package
- Check the examples directory
- Review ARCHITECTURE.md for design decisions

Happy building! 🚀
