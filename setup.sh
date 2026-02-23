#!/bin/bash

# ERP SPPG Setup Script
# This script helps initialize all three components of the system

set -e

echo "=========================================="
echo "ERP SPPG - Project Setup"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "Checking prerequisites..."
echo ""

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go installed: $GO_VERSION"
else
    echo -e "${YELLOW}✗${NC} Go not found. Please install Go 1.21 or higher"
    echo "  Visit: https://golang.org/dl/"
fi

# Check Node.js
if command_exists node; then
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}✓${NC} Node.js installed: $NODE_VERSION"
else
    echo -e "${YELLOW}✗${NC} Node.js not found. Please install Node.js 18 or higher"
    echo "  Visit: https://nodejs.org/"
fi

# Check PostgreSQL
if command_exists psql; then
    PSQL_VERSION=$(psql --version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} PostgreSQL installed: $PSQL_VERSION"
else
    echo -e "${YELLOW}✗${NC} PostgreSQL not found. Please install PostgreSQL 15 or higher"
    echo "  Visit: https://www.postgresql.org/download/"
fi

echo ""
echo "=========================================="
echo "Setup Options"
echo "=========================================="
echo "1. Setup Backend only"
echo "2. Setup Web App only"
echo "3. Setup PWA only"
echo "4. Setup All (Backend + Web + PWA)"
echo "5. Exit"
echo ""
read -p "Choose option (1-5): " choice

case $choice in
    1)
        echo ""
        echo "Setting up Backend..."
        cd backend
        cp .env.example .env
        echo "✓ Created .env file (please edit with your configuration)"
        go mod download
        echo "✓ Downloaded Go dependencies"
        echo ""
        echo "Next steps:"
        echo "1. Edit backend/.env with your database and Firebase configuration"
        echo "2. Create PostgreSQL database: createdb erp_sppg"
        echo "3. Place Firebase credentials as backend/firebase-credentials.json"
        echo "4. Run: cd backend && go run cmd/server/main.go"
        ;;
    2)
        echo ""
        echo "Setting up Web App..."
        cd web
        cp .env.example .env
        echo "✓ Created .env file (please edit with your configuration)"
        npm install
        echo "✓ Installed npm dependencies"
        echo ""
        echo "Next steps:"
        echo "1. Edit web/.env with your API and Firebase configuration"
        echo "2. Run: cd web && npm run dev"
        ;;
    3)
        echo ""
        echo "Setting up PWA..."
        cd pwa
        cp .env.example .env
        echo "✓ Created .env file (please edit with your configuration)"
        npm install
        echo "✓ Installed npm dependencies"
        echo ""
        echo "Next steps:"
        echo "1. Edit pwa/.env with your API and Firebase configuration"
        echo "2. Run: cd pwa && npm run dev"
        ;;
    4)
        echo ""
        echo "Setting up all components..."
        
        # Backend
        echo ""
        echo "1/3 Setting up Backend..."
        cd backend
        cp .env.example .env
        echo "✓ Created backend/.env"
        go mod download
        echo "✓ Downloaded Go dependencies"
        cd ..
        
        # Web
        echo ""
        echo "2/3 Setting up Web App..."
        cd web
        cp .env.example .env
        echo "✓ Created web/.env"
        npm install
        echo "✓ Installed Web dependencies"
        cd ..
        
        # PWA
        echo ""
        echo "3/3 Setting up PWA..."
        cd pwa
        cp .env.example .env
        echo "✓ Created pwa/.env"
        npm install
        echo "✓ Installed PWA dependencies"
        cd ..
        
        echo ""
        echo "=========================================="
        echo "Setup Complete!"
        echo "=========================================="
        echo ""
        echo "Next steps:"
        echo "1. Edit all .env files with your configuration"
        echo "2. Create PostgreSQL database: createdb erp_sppg"
        echo "3. Place Firebase credentials as backend/firebase-credentials.json"
        echo ""
        echo "To start the applications:"
        echo "  Backend: cd backend && go run cmd/server/main.go"
        echo "  Web App: cd web && npm run dev"
        echo "  PWA:     cd pwa && npm run dev"
        ;;
    5)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid option"
        exit 1
        ;;
esac

echo ""
echo "=========================================="
echo "Setup completed successfully!"
echo "=========================================="
