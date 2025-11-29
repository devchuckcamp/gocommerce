package main

import (
	"fmt"

	"github.com/devchuckcamp/gocommerce/cart"
	"github.com/devchuckcamp/gocommerce/catalog"
	"github.com/devchuckcamp/gocommerce/money"
	"github.com/devchuckcamp/gocommerce/orders"
)

func main() {
	fmt.Println("🛒 E-Commerce Domain Library Demo")
	fmt.Println("====================================\n")

	// Demo 1: Money Operations
	fmt.Println("💰 Demo 1: Money Value Object")
	fmt.Println("-------------------------------")
	price1, _ := money.NewFromFloat(19.99, "USD")
	price2, _ := money.NewFromFloat(29.99, "USD")
	fmt.Printf("Price 1: %s\n", price1.String())
	fmt.Printf("Price 2: %s\n", price2.String())

	total, _ := price1.Add(price2)
	fmt.Printf("Total: %s\n", total.String())

	discount := total.Multiply(0.20)
	fmt.Printf("20%% discount: %s\n", discount.String())

	finalPrice, _ := total.Subtract(discount)
	fmt.Printf("Final price: %s\n\n", finalPrice.String())

	// Demo 2: Product Catalog
	fmt.Println("📦 Demo 2: Product Catalog")
	fmt.Println("---------------------------")
	basePrice, _ := money.NewFromFloat(49.99, "USD")
	product := &catalog.Product{
		ID:          "prod-001",
		SKU:         "TSHIRT-BLU-M",
		Name:        "Blue T-Shirt (Medium)",
		Description: "Comfortable cotton t-shirt",
		BrandID:     "brand-001",
		CategoryID:  "cat-clothing",
		BasePrice:   basePrice,
		Status:      catalog.ProductStatusActive,
		Attributes: map[string]string{
			"material": "cotton",
			"color":    "blue",
		},
	}
	fmt.Printf("Product: %s\n", product.Name)
	fmt.Printf("SKU: %s\n", product.SKU)
	fmt.Printf("Price: %s\n", product.BasePrice.String())
	fmt.Printf("Active: %t\n\n", product.IsActive())

	// Demo 3: Shopping Cart
	fmt.Println("🛒 Demo 3: Shopping Cart")
	fmt.Println("------------------------")
	shoppingCart := &cart.Cart{
		ID:     "cart-001",
		UserID: "user-123",
		Items:  []cart.CartItem{},
	}

	item1Price, _ := money.NewFromFloat(49.99, "USD")
	item1 := cart.CartItem{
		ID:        "item-001",
		ProductID: "prod-001",
		SKU:       "TSHIRT-BLU-M",
		Name:      "Blue T-Shirt (Medium)",
		Price:     item1Price,
		Quantity:  2,
	}

	item2Price, _ := money.NewFromFloat(79.99, "USD")
	item2 := cart.CartItem{
		ID:        "item-002",
		ProductID: "prod-002",
		SKU:       "JEANS-BLK-32",
		Name:      "Black Jeans (32)",
		Price:     item2Price,
		Quantity:  1,
	}

	shoppingCart.AddItem(item1)
	shoppingCart.AddItem(item2)

	fmt.Printf("Cart ID: %s\n", shoppingCart.ID)
	fmt.Printf("Items: %d\n", len(shoppingCart.Items))
	fmt.Printf("Total quantity: %d\n", shoppingCart.ItemCount())
	fmt.Printf("Subtotal: %s\n\n", shoppingCart.Subtotal().String())

	// Demo 4: Cart Operations
	fmt.Println("🔄 Demo 4: Cart Operations")
	fmt.Println("--------------------------")
	fmt.Println("Updating item quantity from 2 to 3...")
	shoppingCart.UpdateItemQuantity("item-001", 3)
	fmt.Printf("New subtotal: %s\n", shoppingCart.Subtotal().String())

	fmt.Println("Removing an item...")
	shoppingCart.RemoveItem("item-002")
	fmt.Printf("Items remaining: %d\n", len(shoppingCart.Items))
	fmt.Printf("New subtotal: %s\n\n", shoppingCart.Subtotal().String())

	// Demo 5: Order Status Transitions
	fmt.Println("📋 Demo 5: Order Status Transitions")
	fmt.Println("------------------------------------")
	orderTotal, _ := money.NewFromFloat(149.97, "USD")
	order := &orders.Order{
		ID:          "order-001",
		OrderNumber: "ORD-2025-001",
		UserID:      "user-123",
		Status:      orders.OrderStatusPending,
		Total:       orderTotal,
		Items: []orders.OrderItem{
			{
				ID:        "oi-001",
				ProductID: "prod-001",
				SKU:       "TSHIRT-BLU-M",
				Name:      "Blue T-Shirt (Medium)",
				UnitPrice: item1Price,
				Quantity:  3,
			},
		},
	}

	fmt.Printf("Order Number: %s\n", order.OrderNumber)
	fmt.Printf("Status: %s\n", order.Status)
	fmt.Printf("Total: %s\n", order.Total.String())
	fmt.Printf("Item count: %d\n\n", order.ItemCount())

	fmt.Println("Status transitions:")
	statuses := []orders.OrderStatus{
		orders.OrderStatusPaid,
		orders.OrderStatusProcessing,
		orders.OrderStatusShipped,
		orders.OrderStatusDelivered,
	}

	for _, status := range statuses {
		canTransition := order.CanTransitionTo(status)
		if canTransition {
			order.UpdateStatus(status)
			fmt.Printf("✓ Transitioned to: %s\n", status)
		} else {
			fmt.Printf("✗ Cannot transition to: %s\n", status)
		}
	}

	fmt.Printf("\nFinal order status: %s\n", order.Status)
	fmt.Printf("Cancelable: %t\n", order.IsCancelable())
	fmt.Printf("Refundable: %t\n\n", order.IsRefundable())

	// Demo 6: Money Allocation (splitting amounts)
	fmt.Println("💵 Demo 6: Money Allocation")
	fmt.Println("---------------------------")
	amount, _ := money.NewFromFloat(100.00, "USD")
	fmt.Printf("Splitting %s across 3 people:\n", amount.String())
	
	shares := amount.Allocate(3)
	for i, share := range shares {
		fmt.Printf("Person %d: %s\n", i+1, share.String())
	}
	
	// Verify the sum
	sum := money.Zero("USD")
	for _, share := range shares {
		sum, _ = sum.Add(share)
	}
	fmt.Printf("Total: %s (correct allocation)\n\n", sum.String())

	// Demo 7: Address Validation
	fmt.Println("📍 Demo 7: Address Validation")
	fmt.Println("------------------------------")
	validAddress := orders.Address{
		FirstName:    "John",
		LastName:     "Doe",
		AddressLine1: "123 Main St",
		City:         "San Francisco",
		State:        "CA",
		PostalCode:   "94102",
		Country:      "US",
		Phone:        "+1-555-0100",
	}

	incompleteAddress := orders.Address{
		FirstName: "Jane",
		LastName:  "Smith",
		City:      "Los Angeles",
	}

	fmt.Printf("Valid address (%s): %t\n", validAddress.FullName(), validAddress.IsComplete())
	fmt.Printf("Incomplete address (%s): %t\n\n", incompleteAddress.FullName(), incompleteAddress.IsComplete())

	// Summary
	fmt.Println("✅ Demo Complete!")
	fmt.Println("=================")
	fmt.Println("\nThis demo showed:")
	fmt.Println("• Money value object with currency-safe operations")
	fmt.Println("• Product catalog entities")
	fmt.Println("• Shopping cart management")
	fmt.Println("• Cart operations (add, update, remove)")
	fmt.Println("• Order status state machine")
	fmt.Println("• Money allocation algorithm")
	fmt.Println("• Address validation")
	fmt.Println("\nFor more examples, see:")
	fmt.Println("• examples/usage.go - Domain usage patterns")
	fmt.Println("• examples/http_handlers.go - HTTP integration")
	fmt.Println("• QUICKSTART.md - Implementation guide")
}
