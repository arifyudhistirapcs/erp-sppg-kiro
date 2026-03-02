#!/bin/bash

# Script to clear Firebase KDS data
# This will call the API endpoint to clear all KDS-related data from Firebase

echo "Clearing Firebase KDS data..."

# Call the API endpoint
curl -X POST http://localhost:8080/api/v1/dashboard/clear-firebase \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

echo ""
echo "Done!"
