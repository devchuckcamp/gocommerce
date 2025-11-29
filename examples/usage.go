package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/inventory"
	"github.com/devchuckcamp/gocommerce/money"
	"github.com/devchuckcamp/gocommerce/orders"
	"github.com/devchuckcamp/gocommerce/pricing"
)

// Example 1: Add product to cart
// This shows how an HTTP handler might use the cart service
func ExampleAddToCart(
	ctx context.Context,
	cartService cart.Service,
	userID string,
	productID string,
	variantID *string,
	quantity int,
) error {
	// Get or create cart for user
	shoppingCart, err := cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		return fmt.Errorf("failed to get cart: %w", err)
	}

	// Add item to cart
	req := cart.AddItemRequest{
		ProductID: productID,
		VariantID: variantID,
		Quantity:  quantity,
		Attributes: map[string]string{
			"gift_wrap": "yes",
		},
	}

	updatedCart, err := cartService.AddItem(ctx, shoppingCart.ID, req)
	if err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	fmt.Printf("Cart updated. Total items: %d, Subtotal: %s\n",
		updatedCart.ItemCount(),
		updatedCart.Subtotal().String())

	return nil
}

// Example 2: Price a cart with promotions
// This shows how to calculate complete pricing with discounts and tax
func ExamplePriceCart(
	ctx context.Context,
	pricingService pricing.Service,
	shoppingCart *cart.Cart,
	promoCode string,
	shippingAddress *pricing.Address,
) (*pricing.PricingResult, error) {
	// Calculate pricing with promotion
	shippingMethodID := "standard-shipping"
	result, err := pricingService.PriceCart(ctx, pricing.PriceCartRequest{
		Cart:             shoppingCart,
		PromotionCodes:   []string{promoCode},
		ShippingMethodID: &shippingMethodID,
		ShippingAddress:  shippingAddress,
		TaxInclusive:     false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to price cart: %w", err)
	}

	// Display pricing breakdown
	fmt.Printf("Pricing Breakdown:\n")
	fmt.Printf("  Subtotal:        %s\n", result.Subtotal.String())
	fmt.Printf("  Discount:       -%s\n", result.DiscountTotal.String())
	fmt.Printf("  Shipping:        %s\n", result.ShippingTotal.String())
	fmt.Printf("  Tax:             %s\n", result.TaxTotal.String())
	fmt.Printf("  Total:           %s\n", result.Total.String())
	fmt.Printf("\n")

	// Show applied discounts
	for _, discount := range result.AppliedDiscounts {
		fmt.Printf("Applied: %s (%s) - %s\n",
			discount.Name,
			discount.Code,
			discount.Amount.String())
	}

	return result, nil
}

// Example 3: Create an order from cart
// This shows the complete order creation flow
func ExampleCreateOrder(
	ctx context.Context,
	orderService orders.Service,
	shoppingCart *cart.Cart,
	userID string,
	shippingAddress orders.Address,
	paymentMethodID string,
) (*orders.Order, error) {
	// Create order from cart
	req := orders.CreateOrderRequest{
		Cart:   shoppingCart,
		UserID: userID,
		ShippingAddress: shippingAddress,
		BillingAddress:  shippingAddress, // Use same as shipping
		PaymentMethodID: paymentMethodID,
		PromotionCodes:  []string{"SAVE10"},
		ShippingMethodID: "standard-shipping",
		Notes:           "Please leave at front door",
		IPAddress:       "192.168.1.1",
		UserAgent:       "Mozilla/5.0...",
	}

	order, err := orderService.CreateFromCart(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	fmt.Printf("Order created successfully!\n")
	fmt.Printf("  Order Number: %s\n", order.OrderNumber)
	fmt.Printf("  Status:       %s\n", order.Status)
	fmt.Printf("  Total:        %s\n", order.Total.String())
	fmt.Printf("  Items:        %d\n", order.ItemCount())

	return order, nil
}

// Example 4: Build a complete checkout flow
// This shows how all services work together
func ExampleCompleteCheckout(
	ctx context.Context,
	cartService cart.Service,
	pricingService pricing.Service,
	orderService orders.Service,
	userID string,
	productIDs []string,
	promoCode string,
) error {
	// Step 1: Create cart and add items
	shoppingCart, err := cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		return err
	}

	for _, productID := range productIDs {
		_, err = cartService.AddItem(ctx, shoppingCart.ID, cart.AddItemRequest{
			ProductID: productID,
			Quantity:  1,
		})
		if err != nil {
			return fmt.Errorf("failed to add product %s: %w", productID, err)
		}
	}

	fmt.Printf("✓ Added %d items to cart\n", len(productIDs))

	// Step 2: Calculate pricing
	shippingAddr := &pricing.Address{
		Country:    "US",
		State:      "CA",
		City:       "San Francisco",
		PostalCode: "94102",
	}

	pricingResult, err := ExamplePriceCart(ctx, pricingService, shoppingCart, promoCode, shippingAddr)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Calculated pricing: %s\n", pricingResult.Total.String())

	// Step 3: Create order
	orderAddr := orders.Address{
		FirstName:    "John",
		LastName:     "Doe",
		AddressLine1: "123 Main St",
		City:         "San Francisco",
		State:        "CA",
		PostalCode:   "94102",
		Country:      "US",
		Phone:        "+1-555-0100",
	}

	order, err := ExampleCreateOrder(
		ctx,
		orderService,
		shoppingCart,
		userID,
		orderAddr,
		"pm_test_123",
	)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Order created: %s\n", order.OrderNumber)

	// Step 4: Clear cart after successful order
	_, err = cartService.Clear(ctx, shoppingCart.ID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	fmt.Printf("✓ Cart cleared\n")
	fmt.Printf("\nCheckout complete! Order %s is being processed.\n", order.OrderNumber)

	return nil
}

// Example 5: Merge guest cart to user cart after login
func ExampleMergeGuestCart(
	ctx context.Context,
	cartService cart.Service,
	guestSessionID string,
	userID string,
) error {
	// Get guest cart
	guestCart, err := cartService.GetOrCreateCart(ctx, "", guestSessionID)
	if err != nil {
		return err
	}

	if guestCart.IsEmpty() {
		return nil // Nothing to merge
	}

	// Get or create user cart
	userCart, err := cartService.GetOrCreateCart(ctx, userID, "")
	if err != nil {
		return err
	}

	// Merge guest cart into user cart
	mergedCart, err := cartService.MergeCarts(ctx, guestCart.ID, userCart.ID)
	if err != nil {
		return fmt.Errorf("failed to merge carts: %w", err)
	}

	fmt.Printf("Merged guest cart into user cart. Total items: %d\n",
		mergedCart.ItemCount())

	return nil
}

// Example 6: Update order status (for order fulfillment)
func ExampleFulfillOrder(
	ctx context.Context,
	orderService orders.Service,
	orderID string,
) error {
	// Get order
	order, err := orderService.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	fmt.Printf("Processing order %s (status: %s)\n", order.OrderNumber, order.Status)

	// Transition through fulfillment statuses
	statuses := []orders.OrderStatus{
		orders.OrderStatusProcessing,
		orders.OrderStatusShipped,
		orders.OrderStatusDelivered,
	}

	for _, status := range statuses {
		order, err = orderService.UpdateStatus(ctx, orderID, status)
		if err != nil {
			return fmt.Errorf("failed to update status to %s: %w", status, err)
		}

		fmt.Printf("✓ Order status updated to: %s\n", status)
		time.Sleep(1 * time.Second) // Simulate processing time
	}

	fmt.Printf("Order %s delivered successfully!\n", order.OrderNumber)

	return nil
}

// Example 7: Working with Money value objects
func ExampleMoneyOperations() {
	// Create money values
	price1, _ := money.NewFromFloat(19.99, "USD")
	price2, _ := money.NewFromFloat(29.99, "USD")

	fmt.Printf("Price 1: %s\n", price1.String())
	fmt.Printf("Price 2: %s\n", price2.String())

	// Add prices
	total, _ := price1.Add(price2)
	fmt.Printf("Total: %s\n", total.String())

	// Apply discount (20% off)
	discount := total.Multiply(0.20)
	fmt.Printf("20%% discount: %s\n", discount.String())

	finalPrice, _ := total.Subtract(discount)
	fmt.Printf("Final price: %s\n", finalPrice.String())

	// Compare prices
	isGreater, _ := finalPrice.GreaterThan(price1)
	fmt.Printf("Final price > Price 1: %t\n", isGreater)

	// Allocate money across 3 items (handles remainders correctly)
	allocated := total.Allocate(3)
	fmt.Printf("Split across 3 items:\n")
	for i, amount := range allocated {
		fmt.Printf("  Item %d: %s\n", i+1, amount.String())
	}
}

// Example 8: Catalog operations
func ExampleCatalogOperations(
	ctx context.Context,
	productRepo catalog.ProductRepository,
	categoryID string,
) error {
	// Search products in category
	filter := catalog.ProductFilter{
		Status:   &[]catalog.ProductStatus{catalog.ProductStatusActive}[0],
		MinPrice: &[]int64{1000}[0], // $10.00
		MaxPrice: &[]int64{5000}[0], // $50.00
		SortBy:   "price_asc",
		Limit:    20,
		Offset:   0,
	}

	products, err := productRepo.FindByCategory(ctx, categoryID, filter)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d products in category\n", len(products))

	for _, product := range products {
		fmt.Printf("  - %s: %s (SKU: %s)\n",
			product.Name,
			product.BasePrice.String(),
			product.SKU)
	}

	return nil
}

// Example 9: Inventory check before adding to cart
func ExampleCheckInventory(
	ctx context.Context,
	inventoryService inventory.Service,
	cartService cart.Service,
	cartID string,
	productID string,
	sku string,
	requestedQty int,
) error {
	// Check available stock
	available, err := inventoryService.GetAvailableStock(ctx, sku)
	if err != nil {
		return fmt.Errorf("failed to check inventory: %w", err)
	}

	fmt.Printf("Requested: %d, Available: %d\n", requestedQty, available)

	if available < requestedQty {
		return fmt.Errorf("insufficient stock: only %d available", available)
	}

	// Add to cart
	_, err = cartService.AddItem(ctx, cartID, cart.AddItemRequest{
		ProductID: productID,
		Quantity:  requestedQty,
	})
	if err != nil {
		return err
	}

	fmt.Printf("✓ Added %d items to cart\n", requestedQty)

	return nil
}
