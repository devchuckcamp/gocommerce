# Project Files Index

## 📂 Complete Project Structure

```
gocommerce/
│
├── 📄 go.mod                      # Go module definition
├── 📖 README.md                   # Project overview and quick start
├── 📖 QUICKSTART.md               # 5-minute getting started guide
├── 📖 ARCHITECTURE.md             # Detailed architecture documentation
├── 📖 PACKAGE_SUMMARY.md          # Visual package guide
├── 📖 INDEX.md                    # This file
│
├── 💰 money/                      # Money Value Object
│   └── money.go                   # (190 lines) Money type with currency-safe operations
│
├── 📦 catalog/                    # Product Catalog Domain
│   ├── product.go                 # (99 lines) Product, Variant, Category, Brand
│   └── repository.go              # (58 lines) Repository interfaces
│
├── 🛒 cart/                       # Shopping Cart Domain
│   ├── cart.go                    # (134 lines) Cart aggregate with business logic
│   └── service.go                 # (235 lines) CartService with validation
│
├── 💲 pricing/                    # Pricing Engine
│   ├── pricing.go                 # (102 lines) PricingResult, Promotion, discounts
│   └── service.go                 # (280 lines) PricingService with tax & shipping
│
├── 📋 orders/                     # Order Management
│   ├── order.go                   # (171 lines) Order aggregate with status machine
│   └── service.go                 # (241 lines) OrderService with payment integration
│
├── 📊 inventory/                  # Inventory Management
│   └── inventory.go               # (92 lines) Service interface, reservations
│
├── 💳 payments/                   # Payment Gateway
│   └── payments.go                # (141 lines) Gateway interface, intents, refunds
│
├── 🚚 shipping/                   # Shipping Rates
│   └── shipping.go                # (95 lines) RateCalculator interface, methods
│
├── 🧾 tax/                        # Tax Calculation
│   └── tax.go                     # (113 lines) Calculator interface, tax rates
│
├── 👤 user/                       # User Domain
│   └── user.go                    # (82 lines) UserProfile, Address
│
└── 📚 examples/                   # Usage Examples (NOT part of library)
    ├── usage.go                   # (323 lines) 9 domain usage examples
    └── http_handlers.go           # (424 lines) HTTP integration patterns
```

## 📊 Statistics

**Total Lines of Code**: ~2,600 lines
**Packages**: 11 domain packages
**Services Implemented**: 3 (Cart, Pricing, Order)
**Repository Interfaces**: 10+
**Service Interfaces**: 4 (Inventory, Payments, Shipping, Tax)
**Documentation Files**: 4 comprehensive guides
**Examples**: 9 usage examples + HTTP handlers

## 🗂️ File Descriptions

### Core Documentation

| File | Purpose | Key Content |
|------|---------|-------------|
| `README.md` | Project overview | Quick start, features, integration examples |
| `QUICKSTART.md` | Getting started | Installation, basic usage, testing |
| `ARCHITECTURE.md` | Design guide | DDD patterns, flows, best practices |
| `PACKAGE_SUMMARY.md` | Visual guide | Package dependencies, when to use what |

### Domain Packages (Core Library)

#### money/ - Value Object Package
- **money.go**: Money value object with currency-safe operations
  - Stores amounts as int64 cents (no floating-point errors)
  - Add, Subtract, Multiply operations
  - Allocate method for splitting amounts
  - Currency validation

#### catalog/ - Product Catalog
- **product.go**: Core product entities
  - `Product` - Main product entity
  - `Variant` - Product variants (size, color, etc.)
  - `Category` - Category tree
  - `Brand` - Product brands
- **repository.go**: Repository interfaces
  - `ProductRepository` - CRUD and search
  - `VariantRepository` - Variant management
  - `CategoryRepository` - Category tree operations
  - `BrandRepository` - Brand management
  - `ProductFilter` - Query filters

#### cart/ - Shopping Cart
- **cart.go**: Cart aggregate
  - `Cart` - Shopping cart entity
  - `CartItem` - Item in cart
  - Domain methods: AddItem, RemoveItem, UpdateQuantity, Merge, Subtotal
- **service.go**: Cart business logic
  - `CartService` - Orchestrates cart operations
  - Stock validation
  - Product fetching
  - Guest/user cart transitions

#### pricing/ - Pricing Engine
- **pricing.go**: Pricing types
  - `PricingResult` - Complete pricing breakdown
  - `Promotion` - Discount promotion
  - `LineItem` - Item to be priced
  - `AppliedDiscount` - Discount application result
- **service.go**: Pricing calculations
  - `PricingService` - Main pricing service
  - Applies promotions (percentage, fixed, buy-x-get-y)
  - Integrates with tax calculator
  - Integrates with shipping calculator
  - Handles min purchase, max discount

#### orders/ - Order Management
- **order.go**: Order aggregate
  - `Order` - Order entity
  - `OrderItem` - Item in order
  - `OrderStatus` - Status enum
  - `Address` - Shipping/billing address
  - Status transition validation
- **service.go**: Order operations
  - `OrderService` - Order lifecycle management
  - Creates orders from carts
  - Reserves inventory
  - Processes payments
  - Status management

#### inventory/ - Inventory Management
- **inventory.go**: Stock management
  - `Service` interface - Get stock, reserve, release, commit
  - `StockLevel` - Inventory information
  - `Reservation` - Stock reservation
  - `ReservationStatus` - Reservation states

#### payments/ - Payment Gateway
- **payments.go**: Payment abstraction
  - `Gateway` interface - Create, capture, refund
  - `PaymentIntent` - Payment authorization
  - `Refund` - Payment refund
  - `IntentStatus`, `RefundStatus` - States

#### shipping/ - Shipping Rates
- **shipping.go**: Shipping calculation
  - `RateCalculator` interface - Get rates
  - `ShippingRate` - Cost and delivery estimate
  - `ShippingMethod` - Carrier configuration
  - `ShippableItem` - Item dimensions

#### tax/ - Tax Calculation
- **tax.go**: Tax computation
  - `Calculator` interface - Calculate tax
  - `TaxRate` - Tax rate configuration
  - `CalculationResult` - Tax breakdown
  - Handles tax-inclusive and tax-exclusive pricing

#### user/ - User Domain
- **user.go**: User profile and addresses
  - `UserProfile` - User information
  - `Address` - Saved address
  - `ProfileRepository` interface
  - `AddressRepository` interface

### Examples (Not Part of Library)

#### examples/ - Usage Patterns
- **usage.go**: Domain usage examples
  1. Add product to cart
  2. Price cart with promotions
  3. Create order from cart
  4. Complete checkout flow
  5. Merge guest cart to user cart
  6. Update order status
  7. Money operations
  8. Catalog operations
  9. Inventory check

- **http_handlers.go**: HTTP integration examples
  - `CartHandler` - GET cart, POST add item
  - `OrderHandler` - POST create order
  - `CheckoutHandler` - POST checkout preview
  - Request/Response types
  - Shows how HTTP layer uses domain services

## 🎯 Where to Start

### For Learning
1. Start with `README.md` for overview
2. Read `QUICKSTART.md` for basic usage
3. Look at `examples/usage.go` for patterns
4. Study `ARCHITECTURE.md` for deep dive

### For Implementation
1. Read `PACKAGE_SUMMARY.md` to understand structure
2. Implement repository interfaces for your database
3. Implement service interfaces for external services
4. Use domain services in your HTTP handlers
5. Reference `examples/http_handlers.go` for patterns

### For Specific Features

| Need | Start With |
|------|------------|
| Product catalog | `catalog/product.go`, `catalog/repository.go` |
| Shopping cart | `cart/cart.go`, `cart/service.go` |
| Pricing & discounts | `pricing/pricing.go`, `pricing/service.go` |
| Order management | `orders/order.go`, `orders/service.go` |
| Money handling | `money/money.go` |
| HTTP integration | `examples/http_handlers.go` |
| Complete flows | `examples/usage.go` |

## 🔍 Key Concepts by File

### money/money.go
- Value object pattern
- Currency-safe operations
- Allocation algorithm (remainder handling)
- Immutability

### cart/cart.go
- Aggregate root pattern
- Entity encapsulation
- Business invariants

### cart/service.go
- Service layer pattern
- Dependency injection
- Stock validation
- Cart merging logic

### pricing/service.go
- Pricing calculation engine
- Promotion application
- Tax integration
- Shipping integration
- Line item pricing

### orders/service.go
- Order creation flow
- Inventory reservation
- Payment processing
- Status transitions
- Compensation logic (rollback)

## 📦 Dependencies Between Packages

```
cart/service.go depends on:
  ├─→ catalog.ProductRepository
  ├─→ catalog.VariantRepository
  └─→ inventory.Service

pricing/service.go depends on:
  ├─→ tax.Calculator
  ├─→ shipping.RateCalculator
  └─→ cart.Cart

orders/service.go depends on:
  ├─→ pricing.Service
  ├─→ inventory.Service
  ├─→ payments.Gateway
  └─→ cart.Cart
```

## 🧪 Testing Strategy

### Unit Tests (Test Domain Logic)
- `money/money.go` - Money operations
- `cart/cart.go` - Cart methods
- `orders/order.go` - Status transitions
- `pricing/pricing.go` - Promotion validation

### Integration Tests (Test Services)
- `cart/service.go` - With mock repos
- `pricing/service.go` - With mock calculators
- `orders/service.go` - With mock dependencies

### Example Test Locations
```go
// money/money_test.go
func TestMoney_Add(t *testing.T) { ... }

// cart/cart_test.go
func TestCart_AddItem(t *testing.T) { ... }

// cart/service_test.go
func TestCartService_AddItem(t *testing.T) { ... }
```

## 📝 Code Statistics

| Package | Lines | Files | Key Types |
|---------|-------|-------|-----------|
| money | 190 | 1 | Money |
| catalog | 157 | 2 | Product, Variant, Category, Brand |
| cart | 369 | 2 | Cart, CartItem, CartService |
| pricing | 382 | 2 | PricingResult, Promotion, PricingService |
| orders | 412 | 2 | Order, OrderItem, OrderService |
| inventory | 92 | 1 | StockLevel, Reservation |
| payments | 141 | 1 | PaymentIntent, Refund |
| shipping | 95 | 1 | ShippingRate, ShippingMethod |
| tax | 113 | 1 | TaxRate, CalculationResult |
| user | 82 | 1 | UserProfile, Address |
| **Total** | **2,033** | **14** | **25+ types** |
| examples | 747 | 2 | Usage patterns (not counted in library) |
| docs | ~2,500 | 4 | Comprehensive guides |

## 🎓 Learning Path

### Beginner
1. Read `README.md` introduction
2. Review `money/money.go` for value object pattern
3. Study `cart/cart.go` for aggregate pattern
4. Look at `examples/usage.go` example 1 (add to cart)

### Intermediate
1. Read `ARCHITECTURE.md` patterns section
2. Study `cart/service.go` for service pattern
3. Review `pricing/service.go` for complex service
4. Implement a mock repository

### Advanced
1. Read full `ARCHITECTURE.md`
2. Study `orders/service.go` for orchestration
3. Review `pricing/service.go` for integration points
4. Design your own domain package
5. Implement production repositories

## 🚀 Implementation Checklist

- [ ] Read README.md
- [ ] Read QUICKSTART.md
- [ ] Implement cart.Repository (your database)
- [ ] Implement catalog repositories (your database)
- [ ] Implement orders.Repository (your database)
- [ ] Implement tax.Calculator (simple or external API)
- [ ] Implement shipping.RateCalculator (flat rate or external API)
- [ ] Implement inventory.Service (optional)
- [ ] Implement payments.Gateway (Stripe, PayPal, etc.)
- [ ] Create HTTP handlers using domain services
- [ ] Write unit tests for your implementations
- [ ] Write integration tests for full flows

## 📮 Package Import Paths

```go
import (
    "github.com/devchuckcamp/gocommerce/money"
    "github.com/devchuckcamp/gocommerce/catalog"
    "github.com/devchuckcamp/gocommerce/cart"
    "github.com/devchuckcamp/gocommerce/pricing"
    "github.com/devchuckcamp/gocommerce/orders"
    "github.com/devchuckcamp/gocommerce/inventory"
    "github.com/devchuckcamp/gocommerce/payments"
    "github.com/devchuckcamp/gocommerce/shipping"
    "github.com/devchuckcamp/gocommerce/tax"
    "github.com/devchuckcamp/gocommerce/user"
)
```

---

## Summary

This library provides a complete, production-ready e-commerce domain layer with:

✅ **2,000+ lines** of pure domain logic  
✅ **11 packages** covering all e-commerce subdomains  
✅ **3 implemented services** (Cart, Pricing, Order)  
✅ **14+ interfaces** for your implementations  
✅ **25+ types** with business logic  
✅ **9 examples** showing real usage  
✅ **2,500+ lines** of documentation  

Use this as-is, extend it, or learn from it. Built with Go best practices and DDD principles.
