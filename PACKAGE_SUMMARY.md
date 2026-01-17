# Package Structure Summary

## Visual Package Layout

```
github.com/devchuckcamp/gocommerce/
│
├── money/                          # Value Object Package
│   └── money.go                    # Money type, operations
│
├── catalog/                        # Product Catalog Domain
│   ├── product.go                  # Product, Variant, Category, Brand entities
│   └── repository.go               # Repository interfaces
│
├── cart/                           # Shopping Cart Domain
│   ├── cart.go                     # Cart aggregate, CartItem
│   └── service.go                  # CartService implementation
│
├── pricing/                        # Pricing domain (unit pricing + totals)
│   ├── pricing.go                  # PricingResult, Promotion types
│   ├── service.go                  # PricingService implementation
│   ├── product_price.go            # ProductPrice + ProductPriceRepository
│   └── price_resolver.go           # PriceResolverService + cart adapter
│
├── orders/                         # Order Management Domain
│   ├── order.go                    # Order aggregate, OrderItem
│   └── service.go                  # OrderService implementation
│
├── inventory/                      # Inventory Management
│   └── inventory.go                # Service interface, StockLevel, Reservation
│
├── payments/                       # Payment Gateway Abstraction
│   └── payments.go                 # Gateway interface, PaymentIntent, Refund
│
├── shipping/                       # Shipping Rate Calculation
│   └── shipping.go                 # RateCalculator interface, ShippingRate
│
├── tax/                           # Tax Calculation
│   └── tax.go                     # Calculator interface, TaxRate
│
├── user/                          # User Domain
│   └── user.go                    # UserProfile, Address, repositories
│
├── migrations/                    # Database Migration System
│   ├── migrations.go              # Core migration manager
│   ├── repository.go              # SQL/PostgreSQL repositories
│   ├── generator.go               # Version generator utilities
│   ├── examples.go                # Pre-built migrations (6 migrations)
│   ├── seeder.go                  # Seeding framework
│   ├── seeds.go                   # Built-in seeds (brands, categories, products)
│   ├── README.md                  # Migration system documentation
│   ├── SUMMARY.md                 # Quick reference guide
│   └── examples/
│       ├── DOCKER.md              # PostgreSQL setup guide
│       ├── docker-compose.yml     # PostgreSQL configuration
│       ├── README.md              # Examples documentation
│       └── postgresql/
│           ├── main.go            # PostgreSQL migration runner
│           ├── seed-products.go   # Database seeder
│           └── README.md          # PostgreSQL example documentation
│
├── sample-project/                # Complete Working API
│   ├── main.go                    # HTTP server & handlers
│   ├── store.go                   # In-memory repositories
│   ├── tax.go                     # Tax calculator implementation
│   ├── README.md                  # API documentation
│   └── test-client/
│       └── main.go                # Automated test client
│
├── examples/                      # Usage Examples (NOT part of library)
│   ├── usage.go                   # Domain usage examples
│   └── http_handlers.go           # HTTP integration examples
│
├── go.mod                         # Go module definition
├── README.md                      # Project overview
├── QUICKSTART.md                  # Quick start guide
├── ARCHITECTURE.md                # Detailed architecture guide
└── PACKAGE_SUMMARY.md             # This file
```

## Core Types by Package

### 💰 money/

**Value Objects:**
- `Money` - Monetary value with currency

**Key Methods:**
- `New(amount int64, currency string)`
- `Add(other Money)`, `Subtract(other Money)`
- `Multiply(factor float64)`
- `Allocate(n int)` - Split money correctly

---

### 📦 catalog/

**Entities:**
- `Product` - Product with base price
- `Variant` - Product variant (size, color)
- `Category` - Product category tree
- `Brand` - Product brand

**Interfaces:**
- `ProductRepository`
- `VariantRepository`
- `CategoryRepository`
- `BrandRepository`

---

### 🛒 cart/

**Aggregate:**
- `Cart` - Shopping cart with items

**Value Objects:**
- `CartItem` - Item in cart

**Service:**
- `CartService` - Add, update, remove, merge operations

**Interfaces:**
- `Repository` - Cart persistence

---

### 💲 pricing/

**Types:**
- `PricingResult` - Complete pricing breakdown
- `Promotion` - Discount promotion
- `LineItem` - Item to be priced
- `AppliedDiscount` - Discount that was applied
- `ProductPrice` - Product/variant price record with optional validity window

**Service:**
- `PricingService` - Calculate totals with discounts, tax, shipping
- `PriceResolverService` - Resolve effective unit prices (time-aware)

**Interfaces:**
- `PromotionRepository`
- `ProductPriceRepository`
- `PriceResolver` (for effective unit pricing)

**Dependencies:**
- Uses `tax.Calculator`
- Uses `shipping.RateCalculator`
- Uses `catalog.ProductRepository` and `catalog.VariantRepository` (for product/variant fallback)

---

### 📋 orders/

**Aggregate:**
- `Order` - Customer order with items

**Value Objects:**
- `OrderItem` - Item in order
- `Address` - Shipping/billing address

**Enums:**
- `OrderStatus` - Pending, Paid, Processing, Shipped, Delivered, etc.

**Service:**
- `OrderService` - Create orders, manage status transitions

**Interfaces:**
- `Repository` - Order persistence

**Dependencies:**
- Uses `pricing.Service`
- Uses `inventory.Service`
- Uses `payments.Gateway`

---

### 📊 inventory/

**Interfaces:**
- `Service` - Get stock, reserve, release, commit

**Types:**
- `StockLevel` - Stock information
- `Reservation` - Stock reservation
- `ReservationStatus`

**Repository:**
- `Repository` - Inventory persistence

---

### 💳 payments/

**Interfaces:**
- `Gateway` - Payment processing interface

**Types:**
- `PaymentIntent` - Authorization/charge
- `Refund` - Payment refund
- `IntentStatus`, `RefundStatus`

**Repository:**
- `Repository` - Payment data persistence

---

### 🚚 shipping/

**Interfaces:**
- `RateCalculator` - Calculate shipping rates

**Types:**
- `ShippingRate` - Cost and delivery estimate
- `ShippingMethod` - Carrier and service level
- `ShippableItem` - Item dimensions/weight

**Repository:**
- `Repository` - Shipping method persistence

---

### 🧾 tax/

**Interfaces:**
- `Calculator` - Calculate tax

**Types:**
- `TaxRate` - Tax rate configuration
- `CalculationResult` - Tax calculation result
- `TaxableItem` - Item subject to tax
- `AppliedTaxRate` - Tax rate that was applied

**Repository:**
- `Repository` - Tax rate persistence

---

### 👤 user/

**Entities:**
- `UserProfile` - User profile information
- `Address` - Saved user address

**Interfaces:**
- `ProfileRepository`
- `AddressRepository`

---

## Service Dependencies

```
┌─────────────┐
│ CartService │
└──────┬──────┘
       │ depends on
       ├─→ catalog.ProductRepository
       ├─→ catalog.VariantRepository
   ├─→ inventory.Service
   └─→ cart.PriceResolver (optional)
         │ implemented by
         └─→ pricing.CartPriceResolverAdapter → pricing.PriceResolverService

┌──────────────────┐
│ PricingService   │
└────────┬─────────┘
         │ depends on
         ├─→ PromotionRepository
         ├─→ tax.Calculator
         └─→ shipping.RateCalculator

┌──────────────────────┐
│ PriceResolverService │
└──────────┬───────────┘
       │ depends on
       ├─→ ProductPriceRepository
       ├─→ catalog.ProductRepository
       └─→ catalog.VariantRepository

┌──────────────┐
│ OrderService │
└──────┬───────┘
       │ depends on
       ├─→ pricing.Service
       ├─→ inventory.Service
       └─→ payments.Gateway
```

## Flow: Cart → Order

```
1. User adds items to Cart
   ↓
   CartService.AddItem()
   - Validates product exists
   - Checks inventory
   - Resolves unit price (optional PriceResolver)
   - Updates cart

2. User proceeds to checkout
   ↓
   PricingService.PriceCart()
   - Applies promotions
   - Calculates tax
   - Calculates shipping
   - Returns total

3. User confirms order
   ↓
   OrderService.CreateFromCart()
   - Prices the cart
   - Reserves inventory
   - Processes payment
   - Creates order
   - Clears cart
```

## Interface Implementation Strategy

### Your Infrastructure Layer Implements These:

```go
// Repository Interfaces (your DB layer)
✓ cart.Repository
✓ catalog.ProductRepository
✓ catalog.VariantRepository
✓ catalog.CategoryRepository
✓ catalog.BrandRepository
✓ orders.Repository
✓ pricing.PromotionRepository
✓ pricing.ProductPriceRepository
✓ user.ProfileRepository
✓ user.AddressRepository
✓ inventory.Repository
✓ payments.Repository

// Service Interfaces (your external integrations)
✓ inventory.Service
✓ payments.Gateway
✓ shipping.RateCalculator
✓ tax.Calculator
```

### Domain Library Provides:

```go
// Domain Services (ready to use)
✓ cart.CartService
✓ pricing.PricingService
✓ orders.OrderService

// Domain Entities & Value Objects
✓ All types in each package
✓ Business logic methods
✓ Validation rules
```

## Key Design Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Money** | Value object with int64 cents | Avoid floating-point errors |
| **Repositories** | Interfaces only | Allow any database implementation |
| **Services** | Interface + implementation | CartService, etc. provide business logic |
| **Dependencies** | Through interfaces | Easy testing and swapping |
| **Context** | First parameter everywhere | Standard Go practice |
| **Errors** | Return values, not panic | Idiomatic Go |
| **Immutability** | Value objects are immutable | Prevent bugs |
| **Aggregates** | Control access to children | Enforce invariants |

## When to Use Each Package

| Use Case | Packages Needed |
|----------|----------------|
| Product browsing | `catalog`, `money` |
| Shopping cart | `cart`, `catalog`, `money`, `inventory` |
| Checkout preview | `pricing`, `cart`, `tax`, `shipping` |
| Order creation | `orders`, `cart`, `pricing`, `inventory`, `payments` |
| Order fulfillment | `orders`, `inventory`, `shipping` |
| Refunds | `orders`, `payments` |
| User profile | `user` |

## Extension Points

You can extend the library by:

1. **Implementing Interfaces** - Provide your own repositories, calculators, gateways
2. **Custom Promotions** - Add new discount types to `pricing.Promotion`
3. **Custom Product Pricing** - Implement `pricing.ProductPriceRepository` and populate `pricing.ProductPrice` rules
4. **Custom Tax Logic** - Implement `tax.Calculator` with your rules
5. **Payment Providers** - Implement `payments.Gateway` for Stripe, PayPal, etc.
6. **Shipping Carriers** - Implement `shipping.RateCalculator` for FedEx, UPS, etc.

## Testing Strategy

```
Unit Tests (Domain Logic)
├── money operations
├── cart operations
├── order status transitions
└── promotion calculations

Integration Tests (Services)
├── CartService with mock repos
├── PricingService with mock calculators
└── OrderService end-to-end

Repository Tests
├── Test your DB implementations
└── Use real database (or testcontainers)

E2E Tests
└── Full checkout flow with all services
```

---

## Summary

This library gives you:

✅ **11 domain packages** - Comprehensive e-commerce logic  
✅ **Zero external dependencies** - Only Go standard library  
✅ **Interface-driven** - Plug in any infrastructure  
✅ **DDD patterns** - Entities, aggregates, value objects, services  
✅ **Production-ready** - Validation, error handling, edge cases  
✅ **Well-documented** - Godoc, examples, architecture guide  

Start with the packages you need, implement the required interfaces, and build your application!
