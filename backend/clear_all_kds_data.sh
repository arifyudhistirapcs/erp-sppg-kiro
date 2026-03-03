#!/bin/bash

# Script to clear all KDS and delivery data from both PostgreSQL and Firebase
# This is useful for testing - it resets the system to a clean state

set -e  # Exit on error

echo "🧹 Clearing all KDS and delivery data..."
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Step 1: Clear PostgreSQL database
echo "📊 Step 1: Clearing PostgreSQL database..."
psql -h ${DB_HOST:-localhost} -U ${DB_USER:-arifyudhistira} -d ${DB_NAME:-erp_sppg} -f clear_kds_delivery_data.sql
echo "✓ Database cleared"
echo ""

# Step 2: Clear Firebase
echo "🔥 Step 2: Clearing Firebase data..."
cd cmd/clear_firebase
export FIREBASE_DATABASE_URL="${FIREBASE_DATABASE_URL}"
go run main.go
cd ../..
echo ""

echo "✅ All KDS and delivery data cleared successfully!"
echo ""
echo "📝 Note: Master data (schools, recipes, ingredients, users, menu_plans) preserved."
echo "You can now start fresh with cooking/packing/delivery/cleaning activities."
