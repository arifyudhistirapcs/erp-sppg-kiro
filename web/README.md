# ERP SPPG - Web Application

Aplikasi web desktop untuk staff kantor SPPG (Satuan Pelayanan Pemenuhan Gizi).

## Tech Stack

- **Framework**: Vue 3 dengan Composition API
- **Build Tool**: Vite
- **State Management**: Pinia
- **UI Library**: Ant Design Vue
- **HTTP Client**: Axios
- **Real-time**: Firebase
- **Router**: Vue Router

## Project Structure

```
web/
├── src/
│   ├── assets/          # Static assets (images, fonts, etc.)
│   ├── components/      # Reusable Vue components
│   ├── composables/     # Vue composables (reusable logic)
│   ├── layouts/         # Layout components
│   ├── router/          # Vue Router configuration
│   ├── services/        # API services
│   ├── stores/          # Pinia stores
│   ├── utils/           # Utility functions
│   ├── views/           # Page components
│   ├── App.vue          # Root component
│   └── main.js          # Application entry point
├── .env                 # Environment variables
├── .env.example         # Environment variables template
├── index.html           # HTML entry point
├── package.json         # Dependencies
└── vite.config.js       # Vite configuration
```

## Setup

1. Install dependencies:
```bash
npm install
```

2. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your actual configuration
```

3. Start development server:
```bash
npm run dev
```

4. Build for production:
```bash
npm run build
```

## Features Implemented (Task 15)

### 15.1 Project Setup ✅
- Vue 3 project with Vite
- All required dependencies installed
- Folder structure created
- Environment configuration

### 15.2 Authentication Store & Service ✅
- Pinia store for authentication state
- Auth service with login, logout, refresh token, get current user
- JWT token storage in localStorage
- Axios interceptor to attach token to requests
- Automatic token refresh on 401 responses

### 15.3 Login Page ✅
- Professional login form with validation
- NIK/Email and password input
- Error messages in Bahasa Indonesia
- Role-based redirect after successful login
- Beautiful gradient design

### 15.4 Main Layout with Navigation ✅
- Responsive sidebar navigation
- Role-based menu items (RBAC)
- Header with user info and logout
- Notification bell with unread count
- Collapsible sidebar
- Footer with copyright

### 15.5 Route Guards ✅
- Authentication check before accessing protected routes
- Role-based access control (RBAC)
- Automatic redirect to login if not authenticated
- Redirect to dashboard if already logged in
- Permission utilities and composables

## Authentication Flow

1. User enters NIK/Email and password on login page
2. Credentials sent to backend API `/api/v1/auth/login`
3. Backend validates and returns user data + JWT token
4. Token stored in localStorage
5. User redirected to appropriate page based on role
6. All subsequent API requests include JWT token in Authorization header
7. If token expires (401), user redirected to login

## Role-Based Access Control (RBAC)

The system implements RBAC with 8 roles:

1. **Kepala SPPG** - Full access to all modules
2. **Kepala Yayasan** - Dashboard and financial reports
3. **Akuntan** - Financial, HRM, inventory
4. **Ahli Gizi** - Recipe management, menu planning, KDS
5. **Pengadaan** - Supply chain, logistics, inventory
6. **Chef** - Kitchen Display System (cooking)
7. **Packing** - Kitchen Display System (packing)
8. **Driver** - Delivery tasks (PWA only)
9. **Asisten Lapangan** - Delivery tasks (PWA only)

### Using Permissions in Components

```vue
<script setup>
import { usePermissions } from '@/composables/usePermissions'

const { can, isRole, roleLabel } = usePermissions()
</script>

<template>
  <div>
    <p>Role: {{ roleLabel }}</p>
    
    <!-- Show button only if user has permission -->
    <a-button v-if="can('RECIPE_CREATE')">
      Tambah Resep
    </a-button>
    
    <!-- Show section only for specific role -->
    <div v-if="isRole('kepala_sppg')">
      Admin controls
    </div>
  </div>
</template>
```

## API Configuration

The application connects to the backend API at the URL specified in `.env`:

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

All API calls are made through the `api` service which automatically:
- Adds the base URL
- Attaches JWT token to requests
- Handles 401 responses (redirect to login)
- Provides consistent error handling

## Firebase Configuration

Firebase is used for real-time updates (KDS, dashboard, notifications). Configure in `.env`:

```
VITE_FIREBASE_API_KEY=your-api-key
VITE_FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
VITE_FIREBASE_DATABASE_URL=https://your-project.firebaseio.com
VITE_FIREBASE_PROJECT_ID=your-project-id
VITE_FIREBASE_STORAGE_BUCKET=your-project.appspot.com
VITE_FIREBASE_MESSAGING_SENDER_ID=your-sender-id
VITE_FIREBASE_APP_ID=your-app-id
```

## Next Steps

The following modules will be implemented in subsequent tasks:

- Task 16: Recipe & Menu Planning Module
- Task 17: Kitchen Display System
- Task 18: Supply Chain & Inventory Module
- Task 19: Logistics & Distribution Module
- Task 20: HRM Module
- Task 21: Financial & Asset Module
- Task 22: Executive Dashboard Module
- Task 23: Audit Trail & System Config

## Development Guidelines

### Code Style
- Use Composition API with `<script setup>`
- Use TypeScript-style JSDoc comments for better IDE support
- Follow Vue 3 best practices
- Use Ant Design Vue components consistently

### Naming Conventions
- Components: PascalCase (e.g., `RecipeListView.vue`)
- Composables: camelCase with `use` prefix (e.g., `usePermissions.js`)
- Services: camelCase with service suffix (e.g., `authService.js`)
- Stores: camelCase (e.g., `auth.js`)

### State Management
- Use Pinia for global state
- Use composables for reusable logic
- Keep component state local when possible

### API Calls
- Always use the `api` service from `@/services/api.js`
- Create dedicated service files for each module
- Handle errors gracefully with user-friendly messages in Bahasa Indonesia

### UI/UX
- All text must be in professional Bahasa Indonesia
- Use Ant Design Vue components for consistency
- Ensure responsive design (mobile, tablet, desktop)
- Provide loading states for async operations
- Show clear error messages

## Troubleshooting

### Cannot connect to backend
- Ensure backend is running on `http://localhost:8080`
- Check `VITE_API_BASE_URL` in `.env`
- Check browser console for CORS errors

### Login not working
- Verify backend `/api/v1/auth/login` endpoint is working
- Check credentials are correct
- Check browser console for errors
- Verify JWT token is being stored in localStorage

### Routes not accessible
- Check if user is authenticated
- Verify user role has permission for the route
- Check route configuration in `router/index.js`

## License

Copyright © 2024 SPPG - Satuan Pelayanan Pemenuhan Gizi
