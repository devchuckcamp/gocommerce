#!/bin/bash

# Sample E-Commerce API Test Script

BASE_URL="http://localhost:8080"
USER_ID="user-123"

echo "🧪 Testing E-Commerce API"
echo "=========================="
echo ""

# Test 1: List Products
echo "1. Listing all products..."
curl -s "$BASE_URL/products" | jq '.'
echo ""

# Test 2: Get single product
echo "2. Getting product details..."
curl -s "$BASE_URL/products/prod-1" | jq '.'
echo ""

# Test 3: Add items to cart
echo "3. Adding items to cart..."
curl -s -X POST "$BASE_URL/cart/items" \
  -H "Content-Type: application/json" \
  -H "user-id: $USER_ID" \
  -d '{"product_id": "prod-1", "quantity": 2}' | jq '.'
echo ""

curl -s -X POST "$BASE_URL/cart/items" \
  -H "Content-Type: application/json" \
  -H "user-id: $USER_ID" \
  -d '{"product_id": "prod-2", "quantity": 1}' | jq '.'
echo ""

# Test 4: Get cart
echo "4. Getting cart contents..."
curl -s "$BASE_URL/cart" \
  -H "user-id: $USER_ID" | jq '.'
echo ""

# Test 5: Update item quantity
echo "5. Updating item quantity..."
ITEM_ID=$(curl -s "$BASE_URL/cart" -H "user-id: $USER_ID" | jq -r '.Items[0].ID')
curl -s -X PUT "$BASE_URL/cart/items/$ITEM_ID" \
  -H "Content-Type: application/json" \
  -H "user-id: $USER_ID" \
  -d '{"quantity": 3}' | jq '.'
echo ""

# Test 6: Preview checkout
echo "6. Previewing checkout totals..."
curl -s -X POST "$BASE_URL/checkout/preview" \
  -H "Content-Type: application/json" \
  -H "user-id: $USER_ID" \
  -d '{
    "shipping_address": {
      "country": "US",
      "state": "CA",
      "city": "San Francisco",
      "postal_code": "94102"
    }
  }' | jq '.'
echo ""

# Test 7: Create order
echo "7. Creating order..."
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -H "user-id: $USER_ID" \
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
  }' | jq '.'
echo ""

# Test 8: Verify cart is cleared
echo "8. Verifying cart was cleared after order..."
curl -s "$BASE_URL/cart" \
  -H "user-id: $USER_ID" | jq '.'
echo ""

echo "✅ All tests completed!"
