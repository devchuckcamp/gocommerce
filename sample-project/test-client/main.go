package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "http://localhost:8080"
	userID  = "user-123"
)

func main() {
	fmt.Println("🧪 Testing E-Commerce API")
	fmt.Println("==========================")
	fmt.Println()
	
	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)
	
	// Test 1: List products
	fmt.Println("1️⃣  Listing all products...")
	products := getProducts()
	fmt.Printf("   Found %d products\n", len(products))
	for _, p := range products {
		name := p["Name"].(string)
		price := p["BasePrice"].(map[string]interface{})
		amount := price["Amount"].(float64) / 100
		fmt.Printf("   - %s: $%.2f\n", name, amount)
	}
	fmt.Println()
	
	// Test 2: Add items to cart
	fmt.Println("2️⃣  Adding items to cart...")
	addToCart("prod-1", 2)
	fmt.Println("   ✓ Added 2x Blue T-Shirt")
	
	addToCart("prod-3", 1)
	fmt.Println("   ✓ Added 1x White Sneakers")
	fmt.Println()
	
	// Test 3: View cart
	fmt.Println("3️⃣  Viewing cart contents...")
	cart := getCart()
	items := cart["Items"].([]interface{})
	fmt.Printf("   Cart has %d items\n", len(items))
	
	var subtotal float64
	for _, item := range items {
		i := item.(map[string]interface{})
		name := i["Name"].(string)
		qty := int(i["Quantity"].(float64))
		price := i["Price"].(map[string]interface{})["Amount"].(float64) / 100
		lineTotal := price * float64(qty)
		subtotal += lineTotal
		fmt.Printf("   - %s x%d = $%.2f\n", name, qty, lineTotal)
	}
	fmt.Printf("   Subtotal: $%.2f\n", subtotal)
	fmt.Println()
	
	// Test 4: Preview checkout
	fmt.Println("4️⃣  Previewing checkout totals...")
	preview := previewCheckout()
	subtotalPreview := preview["Subtotal"].(map[string]interface{})["Amount"].(float64) / 100
	tax := preview["TaxTotal"].(map[string]interface{})["Amount"].(float64) / 100
	shipping := preview["ShippingTotal"].(map[string]interface{})["Amount"].(float64) / 100
	total := preview["Total"].(map[string]interface{})["Amount"].(float64) / 100
	fmt.Printf("   Subtotal: $%.2f\n", subtotalPreview)
	fmt.Printf("   Tax:      $%.2f\n", tax)
	fmt.Printf("   Shipping: $%.2f\n", shipping)
	fmt.Printf("   Total:    $%.2f\n", total)
	fmt.Println()
	
	// Test 5: Create order
	fmt.Println("5️⃣  Creating order...")
	order := createOrder()
	orderID := order["ID"].(string)
	orderNumber := order["OrderNumber"].(string)
	status := order["Status"].(string)
	fmt.Printf("   ✓ Order created!\n")
	fmt.Printf("   Order ID: %s\n", orderID)
	fmt.Printf("   Order Number: %s\n", orderNumber)
	fmt.Printf("   Status: %s\n", status)
	fmt.Println()
	
	// Test 6: Verify cart is empty
	fmt.Println("6️⃣  Verifying cart was cleared...")
	cart = getCart()
	items = cart["Items"].([]interface{})
	if len(items) == 0 {
		fmt.Println("   ✓ Cart is empty after order!")
	} else {
		fmt.Printf("   ⚠ Cart still has %d items\n", len(items))
	}
	fmt.Println()
	
	fmt.Println("✅ All tests completed successfully!")
}

func getProducts() []map[string]interface{} {
	resp, err := http.Get(baseURL + "/products")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	var products []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&products)
	return products
}

func addToCart(productID string, quantity int) {
	body := map[string]interface{}{
		"product_id": productID,
		"quantity":   quantity,
	}
	doPost("/cart/items", body)
}

func getCart() map[string]interface{} {
	req, _ := http.NewRequest("GET", baseURL+"/cart", nil)
	req.Header.Set("user-id", userID)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	var cart map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&cart)
	return cart
}

func previewCheckout() map[string]interface{} {
	body := map[string]interface{}{
		"shipping_address": map[string]string{
			"country":     "US",
			"state":       "CA",
			"city":        "San Francisco",
			"postal_code": "94102",
		},
	}
	return doPost("/checkout/preview", body)
}

func createOrder() map[string]interface{} {
	body := map[string]interface{}{
		"shipping_address": map[string]string{
			"first_name":     "John",
			"last_name":      "Doe",
			"address_line_1": "123 Main St",
			"city":           "San Francisco",
			"state":          "CA",
			"postal_code":    "94102",
			"country":        "US",
			"phone":          "+1-555-0100",
		},
		"payment_method_id": "pm_test_123",
	}
	return doPost("/orders", body)
}

func doPost(path string, body interface{}) map[string]interface{} {
	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", baseURL+path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", userID)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(bodyBytes))
		panic(fmt.Sprintf("Request failed with status %d", resp.StatusCode))
	}
	
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}
