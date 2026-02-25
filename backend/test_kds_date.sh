#!/bin/bash

# Test KDS date filtering
# First, get a token

echo "=== Testing KDS Date Filtering ==="
echo ""

# Login to get token
echo "1. Getting auth token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"chef@sppg.com","password":"password123"}')

TOKEN=$(echo $TOKEN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Failed to get token. Response:"
  echo $TOKEN_RESPONSE
  exit 1
fi

echo "Token obtained: ${TOKEN:0:20}..."
echo ""

# Test without date parameter (should default to today)
echo "2. Testing without date parameter (default to today)..."
curl -s -X GET "http://localhost:8080/api/v1/kds/cooking/today" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

# Test with specific date
echo "3. Testing with date=2026-02-25..."
curl -s -X GET "http://localhost:8080/api/v1/kds/cooking/today?date=2026-02-25" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

# Test with invalid date format
echo "4. Testing with invalid date format..."
curl -s -X GET "http://localhost:8080/api/v1/kds/cooking/today?date=25-02-2026" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

echo "=== Test Complete ==="
