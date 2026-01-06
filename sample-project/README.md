# Sample E-Commerce API

A complete RESTful API demonstrating the **gocommerce** domain library in action. This sample project shows how to build a real e-commerce backend using pure domain logic with an HTTP API layer.

## 🎯 What This Demonstrates

- **Clean Architecture**: HTTP handlers separate from domain logic
- **Repository Pattern**: In-memory implementations of repository interfaces
- **Service Layer**: Cart, Pricing, and Order services orchestrating domain operations
- **Tax Calculation**: Custom tax calculator implementation
- **Product Catalog**: Sample products with prices and attributes
- **Shopping Cart**: Add, update, remove items with automatic cart management
- **Checkout Flow**: Preview totals before order creation
- **Order Management**: Create orders from carts with proper status tracking

## 📁 Project Structure

```
sample-project/
├── main.go           # HTTP server and route handlers
├── store.go          # In-memory repository implementations
├── tax.go            # Simple tax calculator implementation
├── test-client/      # Automated test client
│   └── main.go
└── README.md
```

## 🚀 Quick Start

### Option 1: With PostgreSQL Database (Recommended)

```bash
# 1. Start PostgreSQL
cd migrations/examples && docker-compose up -d

# 2. Run migrations
cd postgresql
cd migrations/examples/postgresql
go run main.go

# 3. Seed database
go run seed-products.go

# 4. Start API server (update to use PostgreSQL)
cd ../../sample-project

# Option A: use default local docker settings
USE_POSTGRES=1 go run .

# Option B: provide an explicit connection string
# DATABASE_URL="host=localhost port=5432 user=edomain password=edomain dbname=edomain sslmode=disable" \
# USE_POSTGRES=1 go run .

# Optional: run migrations automatically on startup
# RUN_MIGRATIONS=1 USE_POSTGRES=1 go run .
```

This gives you:
- Persistent data storage
- 8 brands (Apple, Dell, Lenovo, etc.)
- 8 categories (Electronics, Computers, etc.)
- 72 products (22 curated + 50 random)
- Full e-commerce schema

### Option 2: In-Memory Only (Demo)

```bash
cd sample-project
go run .
```

The API will start on `http://localhost:8080` with 4 sample products in memory

### 2. Run the Test Client

In a new terminal:

```bash
cd sample-project/test-client
go run main.go
```

This will run through a complete e-commerce flow:
1. List products
2. Add items to cart
3. View cart contents
4. Preview checkout totals
5. Create an order
6. Verify cart was cleared

## 📡 API Endpoints

### Products
- `GET /products` - List all products
- `GET /products/:id` - Get product details

### Shopping Cart
- `GET /cart` - Get cart (requires `user-id` header)
- `POST /cart/items` - Add item to cart
- `PUT /cart/items/:id` - Update item quantity
- `DELETE /cart/items/:id` - Remove item from cart
- `DELETE /cart` - Clear cart

### Checkout & Orders
- `POST /checkout/preview` - Preview order totals (tax, shipping, etc.)
- `POST /orders` - Create order from cart

## 💡 Usage Examples

### Add item to cart
```bash
curl -X POST http://localhost:8080/cart/items \
  -H "Content-Type: application/json" \
  -H "user-id: user-123" \
  -d '{
    "product_id": "prod-1",
    "quantity": 2
  }'
```

### Get cart
```bash
curl http://localhost:8080/cart \
  -H "user-id: user-123"
```

### Preview checkout
```bash
curl -X POST http://localhost:8080/checkout/preview \
  -H "Content-Type: application/json" \
  -H "user-id: user-123" \
  -d '{
    "shipping_address": {
      "country": "US",
      "state": "CA",
      "city": "San Francisco",
      "postal_code": "94102"
    }
  }'
```

### Create order
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -H "user-id: user-123" \
  -d '{
    "shipping_address": {
      "first_name": "John",
      "last_name": "Doe",
      "address_line_1": "123 Main St",
      "city": "San Francisco",
      "state": "CA",
      "postal_code": "94102",
      "country": "US",
      "phone": "+1-555-0100"
    },
    "payment_method_id": "pm_test_123"
  }'
```

## 🏗️ How It Works

### Repository Pattern

The `MemoryStore` implements all repository interfaces:
- `catalog.ProductRepository`
- `cart.Repository` 
- `orders.Repository`
- `pricing.PromotionRepository`

This demonstrates how to satisfy the domain's repository contracts with in-memory storage. In production, you'd swap these with database implementations.

### Service Orchestration

The API creates domain services with their dependencies:

```go
cartService := cart.NewCartService(
    cartRepo,
    productRepo,
    variantRepo,
    inventoryService,  // nil for demo
    generateID,
)

pricingService := pricing.NewPricingService(
    promotionRepo,
    taxCalculator,
    shippingCalculator,  // nil for demo
)

orderService := orders.NewOrderService(
    orderRepo,
    pricingService,
    inventoryService,   // nil for demo
    paymentGateway,     // nil for demo
    generateOrderNumber,
    generateID,
)
```

### HTTP Handlers

Simple HTTP handlers convert requests/responses and delegate to domain services:

```go
func (api *API) handleCartItems(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req AddItemRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    // Call domain service
    cart, err := api.cartService.AddItem(ctx, cartID, req)
    
    // Return response
    respondJSON(w, cart)
}
```

## 🎓 Learning Points

1. **Domain Independence**: The gocommerce library has zero knowledge of HTTP, JSON, or databases
2. **Interface Segregation**: Each service depends only on the interfaces it needs
3. **Dependency Injection**: All dependencies passed in at service creation
4. **Pure Functions**: Domain logic is testable without mocking infrastructure
5. **Type Safety**: Strong typing catches errors at compile time

## 🔧 Extending This Sample

### Add Database Persistence

The gocommerce library includes a complete migration system. See `migrations/` directory.

```bash
# 1. Start PostgreSQL
cd migrations/examples && docker-compose up -d

# 2. Run migrations (creates 6 tables)
cd postgresql
cd migrations/examples/postgresql
go run main.go

# 3. Seed test data (8 brands, 8 categories, 72 products)
go run seed-products.go
```

Then implement PostgreSQL repositories:
```go
productRepo := postgres.NewProductRepository(db)
cartRepo := postgres.NewCartRepository(db)
```

See `migrations/README.md` for complete documentation

### Add Payment Processing
Implement the `payments.Gateway` interface:
```go
stripeGateway := stripe.NewGateway(apiKey)
orderService := orders.NewOrderService(
    orderRepo,
    pricingService,
    inventoryService,
    stripeGateway,  // Real payment gateway
    generateOrderNumber,
    generateID,
)
```

### Add Shipping Calculation
Implement the `shipping.RateCalculator` interface:
```go
shippoCalculator := shippo.NewRateCalculator(apiKey)
pricingService := pricing.NewPricingService(
    promotionRepo,
    taxCalculator,
    shippoCalculator,  // Real shipping rates
)
```

## 📦 Sample Products

### In-Memory Mode
The API comes pre-loaded with 4 products:
- Blue T-Shirt (Medium) - $49.99
- Black Jeans (32) - $79.99
- White Sneakers (Size 10) - $89.99
- Gray Hoodie (Large) - $59.99

### Database Mode
With PostgreSQL migrations + seeding:
- **8 Brands**: Apple, Dell, Lenovo, HP, Samsung, Logitech, Sony, Bose
- **8 Categories**: Electronics, Computers, Laptops, Accessories, Audio, Storage, Input Devices, Peripherals
- **72 Products**: Including MacBook Pro, Dell XPS, iPhone, iPad, Logitech peripherals, Samsung SSDs, and more

## ⚙️ Configuration

- **Tax Rate**: 8.75% (California sales tax)
- **Currency**: USD
- **Cart Expiration**: 30 days
- **Storage**: In-memory (data lost on restart)

## 📝 Notes

- This is a **demo project** - not production-ready
- Uses in-memory storage (data clears on restart)
- No authentication/authorization (uses simple user-id header)
- No inventory management (always in stock)
- No real payment processing
- No shipping rate calculation
- Tax is calculated but not applied to checkout preview (demo limitation)

## 🚀 Next Steps

Check out the main [gocommerce documentation](../README.md) to learn more about:
- Domain-Driven Design principles
- Repository patterns
- Service implementation
- Testing strategies
- Production deployment
