# Troubleshooting Login Issue

## Status Saat Ini

✅ **Backend:** Running di port 8080
✅ **CORS:** Sudah dikonfigurasi untuk `http://localhost:5175`
✅ **API Test:** Login berhasil dari curl
✅ **Frontend Config:** API URL sudah benar

## Langkah Troubleshooting

### 1. Restart Frontend

Frontend perlu di-restart agar membaca perubahan CORS:

```bash
# Stop frontend (Ctrl+C di terminal frontend)
# Kemudian start lagi:
cd web
npm run dev
```

### 2. Clear Browser Cache

1. Buka DevTools (F12)
2. Klik kanan pada tombol Refresh
3. Pilih "Empty Cache and Hard Reload"

ATAU

1. Buka browser dalam Incognito/Private mode
2. Coba login lagi

### 3. Cek Network Tab

1. Buka DevTools (F12)
2. Pilih tab "Network"
3. Coba login
4. Lihat request ke `/auth/login`
5. Cek:
   - **Status Code:** Harus 200
   - **Response Headers:** Harus ada `Access-Control-Allow-Origin`
   - **Response Body:** Harus ada `token` dan `user`

### 4. Cek Console Tab

1. Buka DevTools (F12)
2. Pilih tab "Console"
3. Coba login
4. Lihat apakah ada error messages

### 5. Test Login Manual dari Console

Buka Console (F12) dan jalankan:

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
.then(data => {
  console.log('✅ Login Success:', data);
  if (data.token) {
    console.log('Token:', data.token);
    console.log('User:', data.user);
  }
})
.catch(err => console.error('❌ Login Error:', err));
```

## Common Issues & Solutions

### Issue 1: CORS Error

**Symptoms:**
```
Access to XMLHttpRequest at 'http://localhost:8080/api/v1/auth/login' 
from origin 'http://localhost:5175' has been blocked by CORS policy
```

**Solution:**
✅ Sudah diperbaiki! CORS sudah dikonfigurasi untuk port 5175.

Jika masih error:
1. Restart backend server
2. Clear browser cache
3. Restart frontend

### Issue 2: Network Error

**Symptoms:**
```
Network Error
Failed to load resource: net::ERR_FAILED
```

**Solution:**
1. Cek backend server running:
   ```bash
   curl http://localhost:8080/api/v1/auth/login
   ```

2. Cek tidak ada firewall blocking

3. Cek port 8080 tidak digunakan process lain:
   ```bash
   lsof -ti:8080
   ```

### Issue 3: 401 Unauthorized

**Symptoms:**
```
{
  "error_code": "UNAUTHORIZED",
  "message": "Token autentikasi tidak ditemukan"
}
```

**Solution:**
Ini normal untuk endpoint yang memerlukan auth. Login endpoint tidak memerlukan token.

### Issue 4: 400 Bad Request / Validation Error

**Symptoms:**
```
{
  "error_code": "VALIDATION_ERROR",
  "message": "Data tidak valid"
}
```

**Solution:**
Pastikan menggunakan field yang benar:
- ✅ `identifier` (bukan `email`)
- ✅ `password`

### Issue 5: Wrong Credentials

**Symptoms:**
```
{
  "success": false,
  "message": "NIK/Email atau password salah"
}
```

**Solution:**
Gunakan credentials yang benar:
- Email: `kepala.sppg@sppg.com`
- Password: `password123`

## Test Credentials

| Email | Password | Role |
|-------|----------|------|
| kepala.sppg@sppg.com | password123 | Kepala SPPG |
| kepala.yayasan@sppg.com | password123 | Kepala Yayasan |
| akuntan@sppg.com | password123 | Akuntan |
| ahli.gizi@sppg.com | password123 | Ahli Gizi |
| chef@sppg.com | password123 | Chef |
| driver1@sppg.com | password123 | Driver |

## Verification Steps

### Step 1: Verify Backend

```bash
# Test login dari terminal
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:5175" \
  -d '{"identifier":"kepala.sppg@sppg.com","password":"password123"}' \
  | jq '.'
```

Expected output:
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

### Step 2: Verify Frontend Config

```bash
# Check frontend .env
cat web/.env | grep VITE_API_BASE_URL
# Should show: VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### Step 3: Verify CORS

```bash
# Check backend .env
cat backend/.env | grep ALLOWED_ORIGINS
# Should include: http://localhost:5175
```

## Quick Fix Checklist

- [ ] Backend server running on port 8080
- [ ] Frontend running on port 5175
- [ ] CORS includes `http://localhost:5175` in backend/.env
- [ ] Backend server restarted after CORS change
- [ ] Frontend restarted
- [ ] Browser cache cleared
- [ ] Using correct credentials
- [ ] Using correct field names (`identifier`, not `email`)

## Still Not Working?

### Option 1: Check Backend Logs

Lihat terminal backend untuk error messages saat login attempt.

### Option 2: Check Frontend Logs

Lihat browser console untuk detailed error messages.

### Option 3: Test with Postman/Insomnia

1. Install Postman atau Insomnia
2. Create POST request to `http://localhost:8080/api/v1/auth/login`
3. Set Headers:
   - `Content-Type: application/json`
   - `Origin: http://localhost:5175`
4. Set Body (JSON):
   ```json
   {
     "identifier": "kepala.sppg@sppg.com",
     "password": "password123"
   }
   ```
5. Send request

### Option 4: Temporary CORS Bypass (Development Only)

Jika masih ada masalah CORS, bisa temporary allow all origins:

**backend/internal/middleware/cors.go:**
```go
// TEMPORARY - FOR DEVELOPMENT ONLY
c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
```

⚠️ **WARNING:** Jangan gunakan ini di production!

## Success Indicators

Setelah login berhasil, Anda harus melihat:

1. **Network Tab:**
   - Status: 200 OK
   - Response contains `token` and `user`

2. **Console Tab:**
   - No error messages
   - Possible success message from app

3. **Application Tab:**
   - Token saved in localStorage/sessionStorage
   - User data saved

4. **Browser:**
   - Redirect ke dashboard
   - Dashboard menampilkan data

## Next Steps After Successful Login

1. Dashboard akan fetch data dari `/api/v1/dashboard/kepala-sppg`
2. Data akan ditampilkan dengan real-time updates
3. Cek Firebase Console untuk melihat sync data

---

**Last Updated:** 2026-02-25 01:56

**Status:** CORS fixed, backend running, ready for login
