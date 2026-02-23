# ERP SPPG Progressive Web App

Aplikasi mobile untuk driver dan karyawan lapangan SPPG menggunakan Vue 3, Pinia, dan Vant UI.

## Tech Stack

- **Framework**: Vue 3
- **State Management**: Pinia
- **UI Library**: Vant (Mobile UI)
- **Build Tool**: Vite
- **PWA**: vite-plugin-pwa + Workbox
- **HTTP Client**: Axios
- **Offline Storage**: Dexie (IndexedDB)
- **Real-time**: Firebase JavaScript SDK

## Project Structure

```
pwa/
├── src/
│   ├── assets/              # Static assets
│   ├── components/          # Reusable components (to be added)
│   ├── views/               # Page components
│   ├── router/              # Vue Router configuration
│   ├── stores/              # Pinia stores
│   ├── services/            # API, Firebase, and IndexedDB services
│   ├── utils/               # Utility functions (to be added)
│   ├── App.vue              # Root component
│   └── main.js              # Application entry point
├── public/                  # Public static files
├── index.html
├── vite.config.js
├── package.json
├── .env.example
└── README.md
```

## Setup

### Prerequisites

- Node.js 18 or higher
- npm or yarn

### Installation

1. Install dependencies:
```bash
npm install
```

2. Setup environment variables:
```bash
cp .env.example .env
# Edit .env dengan konfigurasi yang sesuai
```

### Running the Application

Development mode:
```bash
npm run dev
```

Build for production:
```bash
npm run build
```

Preview production build:
```bash
npm run preview
```

## Features

### Implemented
- Basic project structure
- Vue Router with authentication guard
- Pinia store for authentication
- Axios API client with interceptors
- Firebase integration setup
- IndexedDB setup with Dexie
- PWA configuration with offline support
- Vant UI mobile components
- Bottom navigation

### To Be Implemented
- Authentication module
- Delivery tasks list
- Electronic Proof of Delivery (e-POD)
- Camera integration for photos
- Signature capture
- GPS geotagging
- Offline data sync
- Attendance with Wi-Fi validation

## PWA Features

- **Offline Support**: App dapat berfungsi tanpa koneksi internet
- **Install to Home Screen**: Dapat diinstall seperti aplikasi native
- **Background Sync**: Data akan di-sync otomatis saat online
- **Cache Strategy**: API responses di-cache untuk akses offline

## Development

### Adding New Views
Views ditambahkan di `src/views/` dan didaftarkan di `src/router/index.js`

### Adding New Stores
Pinia stores ditambahkan di `src/stores/`

### Adding New Components
Reusable components ditambahkan di `src/components/`

### Offline Storage
IndexedDB tables didefinisikan di `src/services/db.js` menggunakan Dexie

### API Integration
API calls menggunakan `src/services/api.js` yang sudah dikonfigurasi dengan interceptors
