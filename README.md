# Sistem ERP SPPG

Platform manajemen operasional terintegrasi untuk mengelola produksi, distribusi, dan pelaporan program pemenuhan gizi.

## Struktur Project

```
erp-sppg/
├── backend/          # Golang REST API Server
├── web/             # Vue 3 Web Application (Desktop)
├── pwa/             # Vue 3 Progressive Web App (Mobile)
└── README.md
```

## Komponen Sistem

### Backend (erp-sppg-be)
- **Tech Stack**: Golang 1.21+, Gin Framework, PostgreSQL, Firebase Admin SDK
- **Port**: 8080
- **Database**: PostgreSQL 15+

### Web App (erp-sppg-web)
- **Tech Stack**: Vue 3, TypeScript, Pinia, Ant Design Vue
- **Port**: 5173

### PWA (erp-sppg-pwa)
- **Tech Stack**: Vue 3, TypeScript, Pinia, Vant UI, Workbox
- **Port**: 5174

## Quick Start

### Backend
```bash
cd backend
go mod download
cp .env.example .env
# Edit .env dengan konfigurasi database dan Firebase
go run cmd/server/main.go
```

### Web App
```bash
cd web
npm install
cp .env.example .env
npm run dev
```

### PWA
```bash
cd pwa
npm install
cp .env.example .env
npm run dev
```

## Environment Variables

Lihat file `.env.example` di masing-masing direktori untuk konfigurasi yang diperlukan.

## Documentation

- [Requirements](.kiro/specs/erp-sppg-system/requirements.md)
- [Design](.kiro/specs/erp-sppg-system/design.md)
- [Tasks](.kiro/specs/erp-sppg-system/tasks.md)
