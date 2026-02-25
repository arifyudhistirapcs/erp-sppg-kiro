# Test Login dari Frontend

## Problem
Frontend di `http://localhost:5175` tidak bisa login karena CORS error.

## Solution
Sudah ditambahkan `http://localhost:5175` ke `ALLOWED_ORIGINS` di `.env` dan server sudah di-restart.

## Test Login

### 1. Test dari Browser Console

Buka browser console (F12) dan jalankan:

```javascript
// Test login
fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    identifier: 'kepala.sppg@sppg.com',
    password: 'password123'
  })
})
.then(res => res.json())
.then(data => console.log('Login response:', data))
.catch(err => console.error('Login error:', err));
```

### 2. Test dari Terminal

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:5175" \
  -d '{"identifier":"kepala.sppg@sppg.com","password":"password123"}' \
  -v
```

## Expected Response

```json
{
  "success": true,
  "message": "Login berhasil",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "nik": "12345678901",
    "email": "kepala.sppg@sppg.com",
    "full_name": "Kepala SPPG",
    "role": "kepala_sppg"
  }
}
```

## User Accounts untuk Testing

| Email | Password | Role |
|-------|----------|------|
| kepala.sppg@sppg.com | password123 | Kepala SPPG |
| kepala.yayasan@sppg.com | password123 | Kepala Yayasan |
| akuntan@sppg.com | password123 | Akuntan |
| ahli.gizi@sppg.com | password123 | Ahli Gizi |
| chef@sppg.com | password123 | Chef |
| driver1@sppg.com | password123 | Driver |

## Troubleshooting

### Masih CORS Error?

1. **Cek server sudah restart:**
   ```bash
   # Lihat log server, harus ada "Starting server on port 8080"
   ```

2. **Cek ALLOWED_ORIGINS di .env:**
   ```bash
   cat backend/.env | grep ALLOWED_ORIGINS
   # Harus ada: http://localhost:5175
   ```

3. **Cek browser console untuk error detail**

4. **Clear browser cache dan reload**

### Login Failed?

1. **Cek credentials:**
   - Email: `kepala.sppg@sppg.com`
   - Password: `password123`

2. **Cek database:**
   ```bash
   cd backend
   ./scripts/verify_dashboard_data_pg.sh
   ```

3. **Cek backend logs:**
   - Lihat terminal backend untuk error messages

### Network Error?

1. **Cek backend server running:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Cek port tidak bentrok:**
   ```bash
   lsof -ti:8080
   ```

## Frontend Configuration

Pastikan frontend menggunakan base URL yang benar:

**File: `web/src/services/api.js` atau `web/src/config.js`**

```javascript
const API_BASE_URL = 'http://localhost:8080/api/v1';
```

## Next Steps

Setelah login berhasil:
1. Token akan disimpan di localStorage/sessionStorage
2. Redirect ke dashboard
3. Dashboard akan fetch data dari `/api/v1/dashboard/kepala-sppg`

## Debug Mode

Untuk melihat request/response detail, buka Network tab di browser DevTools:
1. Tekan F12
2. Pilih tab "Network"
3. Coba login
4. Klik request "login" untuk melihat detail
5. Cek Headers, Payload, dan Response

---

**Status:** CORS sudah diperbaiki, server sudah restart

**Last Updated:** 2026-02-25 01:54
