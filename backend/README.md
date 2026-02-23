# ERP SPPG Backend

REST API Server untuk Sistem ERP SPPG menggunakan Golang, Gin Framework, PostgreSQL, dan Firebase.

## Tech Stack

- **Language**: Golang 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+
- **ORM**: GORM
- **Real-time**: Firebase Admin SDK
- **Authentication**: JWT

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point aplikasi
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── database/
│   │   └── database.go          # Database initialization
│   ├── firebase/
│   │   └── firebase.go          # Firebase initialization
│   ├── middleware/
│   │   └── cors.go              # CORS middleware
│   ├── router/
│   │   └── router.go            # Route definitions
│   ├── models/                  # Database models (to be added)
│   ├── handlers/                # HTTP handlers (to be added)
│   ├── services/                # Business logic (to be added)
│   └── utils/                   # Utility functions (to be added)
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

## Setup

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Firebase project with credentials

### Installation

1. Install dependencies:
```bash
go mod download
```

2. Setup environment variables:
```bash
cp .env.example .env
# Edit .env dengan konfigurasi yang sesuai
```

3. Setup PostgreSQL database:
```bash
createdb erp_sppg
```

4. Place Firebase credentials:
```bash
# Download serviceAccountKey.json dari Firebase Console
# Simpan sebagai firebase-credentials.json di root backend/
```

### Running the Server

Development mode:
```bash
go run cmd/server/main.go
```

Build and run:
```bash
go build -o server cmd/server/main.go
./server
```

## API Endpoints

### Health Check
```
GET /health
```

### API v1
Base URL: `/api/v1`

Endpoints akan ditambahkan sesuai dengan implementasi modul-modul berikutnya.

## Environment Variables

Lihat `.env.example` untuk daftar lengkap environment variables yang diperlukan.

## Development

### Adding New Models
Models ditambahkan di `internal/models/` dan di-migrate di `internal/database/database.go`

### Adding New Routes
Routes ditambahkan di `internal/router/router.go`

### Adding New Handlers
Handlers ditambahkan di `internal/handlers/`

### Adding New Services
Business logic ditambahkan di `internal/services/`
