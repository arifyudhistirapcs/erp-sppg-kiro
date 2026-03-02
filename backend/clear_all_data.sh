#!/bin/bash

echo "=== Clearing All KDS and Delivery Data ==="
echo ""

# Step 1: Clear database
echo "Step 1: Clearing database..."
psql -U arifyudhistira -d erp_sppg -f clear_kds_delivery_data.sql
echo ""

# Step 2: Login to get token
echo "Step 2: Getting authentication token..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "Error: Failed to get authentication token"
  echo "Please make sure the server is running and credentials are correct"
  exit 1
fi

echo "Token obtained successfully"
echo ""

# Step 3: Clear Firebase
echo "Step 3: Clearing Firebase KDS data..."
curl -X POST http://localhost:8080/api/v1/dashboard/clear-firebase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo ""
echo "=== All data cleared successfully! ==="
echo "You can now start fresh with KDS and delivery tasks."
