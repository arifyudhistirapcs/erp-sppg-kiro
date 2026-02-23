# Kitchen Display System (KDS) - Dokumentasi

## Overview

Kitchen Display System (KDS) adalah modul untuk menampilkan informasi real-time kepada tim dapur dan packing. Sistem ini terdiri dari dua tampilan utama:

1. **KDS Cooking Display** - Untuk tim dapur/chef
2. **KDS Packing Display** - Untuk tim packing

## Fitur Utama

### KDS Cooking Display (`/kds/cooking`)

**Fitur:**
- Menampilkan daftar resep yang harus dimasak hari ini
- Menampilkan jumlah porsi yang diperlukan per resep
- Menampilkan daftar bahan-bahan dengan takaran
- Menampilkan instruksi memasak lengkap
- Status resep: Belum Dimulai, Sedang Dimasak, Selesai
- Tombol aksi: "Mulai Masak" dan "Selesai"
- Real-time update via Firebase
- Indikator koneksi Firebase

**Status Resep:**
- `pending` (Belum Dimulai) - Resep belum dimulai
- `cooking` (Sedang Dimasak) - Chef sedang memasak
- `ready` (Selesai) - Resep sudah selesai dimasak

**Alur Kerja:**
1. Chef melihat daftar resep hari ini
2. Chef klik "Mulai Masak" untuk memulai memasak
3. Sistem otomatis mengurangi stok bahan baku dari inventory
4. Chef klik "Selesai" setelah masakan selesai
5. Status diupdate ke Firebase secara real-time

### KDS Packing Display (`/kds/packing`)

**Fitur:**
- Menampilkan daftar sekolah dengan alokasi porsi
- Menampilkan total porsi per sekolah
- Menampilkan menu items yang harus dikemas per sekolah
- Status packing: Belum Dimulai, Sedang Packing, Siap Kirim
- Tombol aksi: "Mulai Packing" dan "Siap Kirim"
- Notifikasi ketika semua sekolah siap kirim
- Real-time update via Firebase
- Badge counter untuk sekolah yang sudah siap
- Indikator koneksi Firebase

**Status Packing:**
- `pending` (Belum Dimulai) - Belum mulai packing
- `packing` (Sedang Packing) - Sedang proses packing
- `ready` (Siap Kirim) - Sudah selesai dikemas dan siap dikirim

**Alur Kerja:**
1. Tim packing melihat daftar sekolah dengan alokasi porsi
2. Tim packing klik "Mulai Packing" untuk sekolah tertentu
3. Setelah selesai, klik "Siap Kirim"
4. Ketika semua sekolah sudah "Siap Kirim", sistem mengirim notifikasi ke tim logistik
5. Status diupdate ke Firebase secara real-time

## Teknologi

### Frontend
- **Vue 3** dengan Composition API
- **Ant Design Vue** untuk UI components
- **Firebase Realtime Database** untuk real-time updates
- **Axios** untuk API calls

### Backend API Endpoints
- `GET /api/v1/kds/cooking/today` - Get today's cooking menu
- `PUT /api/v1/kds/cooking/:recipe_id/status` - Update cooking status
- `GET /api/v1/kds/packing/today` - Get today's packing allocations
- `PUT /api/v1/kds/packing/:school_id/status` - Update packing status
- `POST /api/v1/kds/cooking/sync` - Manual sync to Firebase
- `POST /api/v1/kds/packing/sync` - Manual sync to Firebase

### Firebase Structure

**Cooking Data:**
```
/kds/cooking/{date}/{recipe_id}
  - recipe_id: number
  - name: string
  - status: "pending" | "cooking" | "ready"
  - start_time: timestamp (optional)
  - portions_required: number
  - instructions: string
  - ingredients: array
```

**Packing Data:**
```
/kds/packing/{date}/{school_id}
  - school_id: number
  - school_name: string
  - portions: number
  - menu_items: array
  - status: "pending" | "packing" | "ready"
  - updated_at: timestamp
```

**Notifications:**
```
/notifications/logistics/packing_complete
  - message: string
  - date: string
  - timestamp: number
```

## Permissions (RBAC)

### KDS Cooking Display
- `kepala_sppg` - Full access
- `ahli_gizi` - Full access
- `chef` - Full access

### KDS Packing Display
- `kepala_sppg` - Full access
- `ahli_gizi` - Full access
- `chef` - Full access
- `packing` - Full access

## Komponen

### Services
- `web/src/services/kdsService.js` - API service untuk KDS

### Views
- `web/src/views/KDSCookingView.vue` - Cooking display view
- `web/src/views/KDSPackingView.vue` - Packing display view

### Routes
- `/kds/cooking` - KDS Cooking Display
- `/kds/packing` - KDS Packing Display

## Real-time Updates

Kedua tampilan KDS menggunakan Firebase Realtime Database untuk mendapatkan update secara real-time:

1. **Auto-refresh**: Data diupdate otomatis tanpa perlu refresh halaman
2. **Connection Status**: Indikator menunjukkan status koneksi Firebase
3. **Optimistic Updates**: UI diupdate langsung, kemudian disinkronkan dengan Firebase
4. **Conflict Resolution**: Firebase menjadi single source of truth

## UI/UX Features

### Visual Indicators
- **Color-coded cards**: 
  - Grey border: Belum dimulai
  - Blue border + shadow: Sedang proses
  - Green border + shadow: Selesai
- **Status tags**: Warna berbeda untuk setiap status
- **Connection indicator**: Badge hijau/merah untuk status koneksi
- **Ready counter**: Badge menunjukkan jumlah sekolah yang siap

### Responsive Design
- Grid layout yang responsive (xs, sm, md, lg, xl)
- Mobile-friendly card design
- Scrollable content untuk instruksi panjang

### User Feedback
- Loading states pada tombol aksi
- Success/error messages menggunakan Ant Design message
- Notifications untuk event penting
- Real-time status updates

## Testing

### Manual Testing Checklist

**KDS Cooking:**
- [ ] Tampilan daftar resep hari ini
- [ ] Tampilan bahan-bahan dan takaran
- [ ] Tampilan instruksi memasak
- [ ] Tombol "Mulai Masak" mengubah status ke "Sedang Dimasak"
- [ ] Tombol "Selesai" mengubah status ke "Selesai"
- [ ] Real-time update dari Firebase
- [ ] Indikator koneksi Firebase
- [ ] Inventory deduction saat mulai masak

**KDS Packing:**
- [ ] Tampilan daftar sekolah dengan alokasi
- [ ] Tampilan menu items per sekolah
- [ ] Tombol "Mulai Packing" mengubah status
- [ ] Tombol "Siap Kirim" mengubah status
- [ ] Notifikasi ketika semua sekolah siap
- [ ] Real-time update dari Firebase
- [ ] Badge counter untuk sekolah siap
- [ ] Alert ketika semua sekolah siap kirim

## Troubleshooting

### Firebase Connection Issues
- Pastikan Firebase config sudah benar di `.env`
- Cek Firebase console untuk database rules
- Cek browser console untuk error messages

### Data Tidak Muncul
- Pastikan ada menu plan yang approved untuk hari ini
- Pastikan ada delivery tasks untuk hari ini (untuk packing)
- Cek API response di Network tab browser

### Status Tidak Update
- Cek koneksi Firebase (indikator di header)
- Cek Firebase console untuk data structure
- Refresh halaman untuk force reload

## Future Enhancements

1. **Audio Notifications**: Suara notifikasi untuk status changes
2. **Timer Display**: Countdown timer untuk cooking time
3. **Photo Upload**: Upload foto hasil masakan
4. **Print Function**: Print cooking instructions
5. **Multi-language**: Support bahasa lain selain Indonesia
6. **Dark Mode**: Mode gelap untuk mengurangi eye strain
7. **Fullscreen Mode**: Mode fullscreen untuk display di dapur
8. **Voice Commands**: Kontrol dengan suara untuk hands-free operation

## Maintenance

### Regular Tasks
- Monitor Firebase usage dan quota
- Review dan optimize Firebase rules
- Update dependencies secara berkala
- Backup Firebase data secara rutin

### Performance Optimization
- Implement pagination untuk data besar
- Cache frequently accessed data
- Optimize Firebase queries
- Minimize re-renders dengan proper Vue reactivity

## Support

Untuk pertanyaan atau issues terkait KDS module, hubungi:
- Development Team
- System Administrator
- Technical Support

---

**Last Updated**: 2024
**Version**: 1.0.0
**Module**: Kitchen Display System (KDS)
