# ERP SPPG - Setup Guide

Panduan lengkap untuk setup dan menjalankan Sistem ERP SPPG.

## Daftar Isi

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Backend Setup](#backend-setup)
4. [Web App Setup](#web-app-setup)
5. [PWA Setup](#pwa-setup)
6. [Firebase Configuration](#firebase-configuration)
7. [Database Setup](#database-setup)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

Pastikan sistem Anda memiliki software berikut:

### Required
- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Node.js 18+** - [Download](https://nodejs.org/)
- **PostgreSQL 15+** - [Download](https://www.postgresql.org/download/)
- **Firebase Account** - [Console](https://console.firebase.google.com/)

### Optional
- **Git** - untuk version control
- **Docker** - untuk containerized deployment
- **Redis** - untuk caching (production)

## Quick Start

### Automated Setup

Gunakan script setup otomatis:

```bash
chmod +x setup.sh
./setup.sh
```

Pilih opsi 4 untuk setup semua komponen sekaligus.

### Manual Setup

Jika prefer setup manual, ikuti langkah-langkah di bawah untuk setiap komponen.

## Backend Setup

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit file `.env` dengan konfigurasi Anda:

```env
# Server
PORT=8080
GIN_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=erp_sppg
DB_SSLMODE=disable

# JWT
JWT_SECRET=your-secret-key-change-this
JWT_EXPIRY_HOURS=24

# Firebase
FIREBASE_CREDENTIALS_PATH=./firebase-credentials.json
FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
STORAGE_BUCKET=your-project.appspot.com

# CORS
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174
```

### 3. Setup Firebase Credentials

1. Buka [Firebase Console](https://console.firebase.google.com/)
2. Pilih project Anda
3. Go to Project Settings → Service Accounts
4. Click "Generate New Private Key"
5. Save file sebagai `backend/firebase-credentials.json`

### 4. Create Database

```bash
createdb erp_sppg
```

Atau menggunakan psql:

```sql
CREATE DATABASE erp_sppg;
```

### 5. Run Backend

```bash
go run cmd/server/main.go
```

Backend akan berjalan di `http://localhost:8080`

Test dengan:
```bash
curl http://localhost:8080/health
```

## Web App Setup

### 1. Install Dependencies

```bash
cd web
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit file `.env`:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Firebase Configuration
VITE_FIREBASE_API_KEY=your-api-key
VITE_FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
VITE_FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
VITE_FIREBASE_PROJECT_ID=your-project-id
VITE_FIREBASE_STORAGE_BUCKET=your-project.appspot.com
VITE_FIREBASE_MESSAGING_SENDER_ID=your-sender-id
VITE_FIREBASE_APP_ID=your-app-id
```

### 3. Run Web App

```bash
npm run dev
```

Web app akan berjalan di `http://localhost:5173`

### 4. Build for Production

```bash
npm run build
npm run preview
```

## PWA Setup

### 1. Install Dependencies

```bash
cd pwa
npm install
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit file `.env` (sama seperti Web App):

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1

# Firebase Configuration (sama dengan Web App)
VITE_FIREBASE_API_KEY=your-api-key
# ... dst
```

### 3. Run PWA

```bash
npm run dev
```

PWA akan berjalan di `http://localhost:5174`

### 4. Build for Production

```bash
npm run build
npm run preview
```

## Firebase Configuration

### 1. Create Firebase Project

1. Buka [Firebase Console](https://console.firebase.google.com/)
2. Click "Add Project"
3. Ikuti wizard setup

### 2. Enable Services

Enable services berikut:

- **Realtime Database** - untuk real-time sync
- **Cloud Storage** - untuk upload foto dan file
- **Authentication** (optional) - jika ingin gunakan Firebase Auth

### 3. Get Configuration

#### For Backend (Service Account):
- Project Settings → Service Accounts
- Generate New Private Key
- Save sebagai `backend/firebase-credentials.json`

#### For Web/PWA (Web Config):
- Project Settings → General
- Scroll ke "Your apps" section
- Click Web icon (</>) untuk add web app
- Copy configuration values ke `.env` file

### 4. Setup Realtime Database Rules

```json
{
  "rules": {
    ".read": "auth != null",
    ".write": "auth != null",
    "kds": {
      ".read": true,
      ".write": "auth != null"
    },
    "dashboard": {
      ".read": "auth != null",
      ".write": "auth != null"
    }
  }
}
```

### 5. Setup Storage Rules

```
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    match /{allPaths=**} {
      allow read: if request.auth != null;
      allow write: if request.auth != null;
    }
  }
}
```

## Database Setup

### Create Database

```bash
createdb erp_sppg
```

### Run Migrations

Migrations akan dijalankan otomatis saat backend start (menggunakan GORM AutoMigrate).

Untuk production, gunakan migration tools seperti:
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [goose](https://github.com/pressly/goose)

### Seed Data (Optional)

Untuk development, Anda bisa create seed data:

```bash
cd backend
go run cmd/seed/main.go
```

(Script seed akan dibuat pada task berikutnya)

## Running All Services

### Development

Buka 3 terminal windows:

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/server/main.go
```

**Terminal 2 - Web App:**
```bash
cd web
npm run dev
```

**Terminal 3 - PWA:**
```bash
cd pwa
npm run dev
```

### Access URLs

- Backend API: http://localhost:8080
- Web App: http://localhost:5173
- PWA: http://localhost:5174
- API Health Check: http://localhost:8080/health

## Troubleshooting

### Backend Issues

**Error: "failed to connect to database"**
- Pastikan PostgreSQL running: `pg_ctl status`
- Check database credentials di `.env`
- Pastikan database sudah dibuat: `createdb erp_sppg`

**Error: "failed to initialize Firebase"**
- Pastikan `firebase-credentials.json` ada dan valid
- Check Firebase configuration di `.env`
- Pastikan Firebase project sudah dibuat

### Web/PWA Issues

**Error: "Network Error" atau "CORS Error"**
- Pastikan backend running
- Check `VITE_API_BASE_URL` di `.env`
- Check CORS configuration di backend

**Error: "Firebase not initialized"**
- Check Firebase configuration di `.env`
- Pastikan semua Firebase env variables terisi

### Database Issues

**Error: "database does not exist"**
```bash
createdb erp_sppg
```

**Error: "password authentication failed"**
- Check DB_USER dan DB_PASSWORD di backend/.env
- Reset PostgreSQL password jika perlu

## Next Steps

Setelah infrastructure setup selesai, lanjutkan ke:

1. **Task 2**: Implement Database Schema and Migrations
2. **Task 3**: Implement Authentication & Authorization Module

Lihat [tasks.md](.kiro/specs/erp-sppg-system/tasks.md) untuk detail implementasi.

## Support

Untuk pertanyaan atau issues:
- Check [Requirements](.kiro/specs/erp-sppg-system/requirements.md)
- Check [Design Document](.kiro/specs/erp-sppg-system/design.md)
- Review [Tasks](.kiro/specs/erp-sppg-system/tasks.md)
