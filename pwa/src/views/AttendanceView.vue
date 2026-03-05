<template>
  <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
    <div class="attendance-page">
      <!-- NavBar -->
      <van-nav-bar title="Absensi" />

      <!-- Header: Employee name + role -->
      <div class="attendance-header">
        <div class="attendance-header__info">
          <div class="attendance-header__name-row">
            <h2 class="attendance-header__name">{{ employeeName }}</h2>
            <span class="attendance-header__role">{{ employeeRole }}</span>
          </div>
          <p class="attendance-header__date">{{ todayFormatted }}</p>
        </div>
      </div>

      <!-- Loading skeleton for status cards -->
      <template v-if="loading">
        <div class="status-cards">
          <SkeletonCard :rows="2" />
          <SkeletonCard :rows="2" />
        </div>
        <div class="swipe-section">
          <SkeletonCard :rows="1" />
        </div>
        <div class="summary-scroll">
          <SkeletonCard :rows="2" />
          <SkeletonCard :rows="2" />
          <SkeletonCard :rows="2" />
        </div>
        <div class="section-block">
          <SkeletonCard :rows="3" />
        </div>
        <div class="section-block">
          <SkeletonCard :rows="4" />
        </div>
        <div class="section-block">
          <SkeletonCard :rows="2" />
          <SkeletonCard :rows="2" />
        </div>
      </template>

      <template v-else>
        <!-- Check-in / Check-out Status Cards (two-column) -->
        <div class="status-cards">
          <div class="status-card status-card--checkin">
            <div class="status-card__icon-wrap status-card__icon-wrap--green">
              <van-icon name="play-circle-o" color="#fff" size="22" />
            </div>
            <span class="status-card__label">Check In</span>
            <span class="status-card__time">{{ checkInTime }}</span>
          </div>
          <div class="status-card status-card--checkout">
            <div class="status-card__icon-wrap status-card__icon-wrap--red">
              <van-icon name="stop-circle-o" color="#fff" size="22" />
            </div>
            <span class="status-card__label">Check Out</span>
            <span class="status-card__time">{{ checkOutTime }}</span>
          </div>
        </div>

        <!-- Swipe Action Bar -->
        <div class="swipe-section">
          <SwipeAction
            :canCheckIn="canCheckIn"
            :canCheckOut="canCheckOut"
            :loading="actionLoading"
            @check-in="performAutoCheckIn"
            @check-out="performCheckOut"
          />
        </div>

        <!-- Summary Cards (horizontal scroll) -->
        <div class="summary-scroll">
          <SummaryCard
            icon="calendar-o"
            :iconColor="'var(--h-primary)'"
            label="Hari Kerja"
            :value="stats.totalHariKerja"
            :loading="false"
          />
          <SummaryCard
            icon="clock-o"
            :iconColor="'var(--h-info)'"
            label="Jam Kerja"
            :value="stats.rataJamKerja"
            :loading="false"
          />
          <SummaryCard
            icon="checked"
            :iconColor="'var(--h-success)'"
            label="Hadir"
            :value="stats.totalHadir"
            :loading="false"
          />
        </div>

        <!-- Attendance Statistics -->
        <div class="section-block">
          <h3 class="section-title">Statistik Kehadiran</h3>
          <div class="stats-grid">
            <div class="stat-item">
              <span class="stat-value stat-value--success">{{ stats.totalHadir }}</span>
              <span class="stat-label">Hadir</span>
            </div>
            <div class="stat-item">
              <span class="stat-value stat-value--error">{{ stats.totalTidakHadir }}</span>
              <span class="stat-label">Tidak Hadir</span>
            </div>
            <div class="stat-item">
              <span class="stat-value stat-value--warning">{{ stats.totalTerlambat }}</span>
              <span class="stat-label">Terlambat</span>
            </div>
            <div class="stat-item">
              <span class="stat-value stat-value--primary">{{ stats.rataJamKerja }}</span>
              <span class="stat-label">Rata-rata Jam</span>
            </div>
          </div>
        </div>

        <!-- Mini Calendar -->
        <div class="section-block">
          <h3 class="section-title">Kalender Kehadiran</h3>
          <MiniCalendar
            :attendanceData="calendarData"
            :selectedDate="selectedDateStr"
            @select-date="onCalendarDateSelect"
          />
        </div>

        <!-- Activity Log: 7-day history -->
        <div class="section-block">
          <h3 class="section-title">Riwayat 7 Hari</h3>
          <div v-if="attendanceHistory.length === 0" class="empty-state">
            <van-icon name="info-o" size="32" color="var(--h-text-light)" />
            <span>Belum ada riwayat absensi</span>
          </div>
          <div v-else class="history-list">
            <div
              v-for="record in attendanceHistory"
              :key="record.id || record.date"
              class="history-card"
            >
              <div class="history-card__date">{{ formatDate(record.date) }}</div>
              <div class="history-card__times">
                <span class="history-card__in">
                  <van-icon name="play-circle-o" color="var(--h-success)" size="14" />
                  {{ formatTime(record.check_in || record.check_in_time) }}
                </span>
                <span class="history-card__separator">—</span>
                <span class="history-card__out">
                  <van-icon name="stop-circle-o" color="var(--h-error)" size="14" />
                  {{ formatTime(record.check_out || record.check_out_time) }}
                </span>
                <span class="history-card__duration" v-if="record.work_hours">
                  ({{ formatWorkHours(record.work_hours) }})
                </span>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </van-pull-refresh>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { showToast, showDialog, showLoadingToast, closeToast, showNotify } from 'vant'
import attendanceService from '@/services/attendanceService.js'
import wifiService from '@/services/wifiService.js'
import { attendanceAPI } from '@/services/api.js'
import { useAuthStore } from '@/stores/auth.js'
import SwipeAction from '@/components/mobile/SwipeAction.vue'
import SummaryCard from '@/components/mobile/SummaryCard.vue'
import MiniCalendar from '@/components/mobile/MiniCalendar.vue'
import SkeletonCard from '@/components/mobile/SkeletonCard.vue'

const authStore = useAuthStore()

const loading = ref(false)
const actionLoading = ref(false)
const refreshing = ref(false)
const currentAttendance = ref(null)
const attendanceHistory = ref([])
const authorizedNetworks = ref([])
const selectedDate = ref(new Date())

// Employee name from auth store
const employeeName = computed(() => authStore.user?.name || 'Karyawan')

// Employee role from auth store
const employeeRole = computed(() => {
  const role = authStore.user?.role
  if (!role) return ''
  
  // Format role untuk display
  const roleMap = {
    'driver': 'Driver',
    'asisten_lapangan': 'Asisten Lapangan',
    'kepala_sppg': 'Kepala SPPG',
    'ahli_gizi': 'Ahli Gizi',
    'chef': 'Chef',
    'packing': 'Packing',
    'pengadaan': 'Pengadaan',
    'akuntan': 'Akuntan',
    'sekolah': 'Sekolah'
  }
  
  return roleMap[role.toLowerCase()] || role
})

// Today's formatted date
const todayFormatted = computed(() => {
  const now = new Date()
  return now.toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
})

// Selected date as string for MiniCalendar
const selectedDateStr = computed(() => {
  const d = selectedDate.value
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
})

// Check-in / Check-out display times
const checkInTime = computed(() => {
  const att = currentAttendance.value
  if (!att) return '--:--'
  return formatTime(att.check_in || att.check_in_time) || '--:--'
})

const checkOutTime = computed(() => {
  const att = currentAttendance.value
  if (!att) return '--:--'
  return formatTime(att.check_out || att.check_out_time) || '--:--'
})

// Can check in / check out (preserved logic)
const canCheckIn = computed(() => {
  const attendance = currentAttendance.value
  return !attendance || (attendance && (attendance.check_out || attendance.check_out_time))
})

const canCheckOut = computed(() => {
  const attendance = currentAttendance.value
  return attendance &&
    (attendance.check_in || attendance.check_in_time) &&
    !(attendance.check_out || attendance.check_out_time)
})

// Attendance statistics (computed from history)
const stats = computed(() => {
  const history = attendanceHistory.value
  let totalHadir = 0
  let totalTidakHadir = 0
  let totalTerlambat = 0
  let totalJamKerja = 0

  // Get current month records
  const now = new Date()
  const currentMonth = now.getMonth()
  const currentYear = now.getFullYear()

  const monthRecords = history.filter(r => {
    if (!r.date) return false
    const d = new Date(r.date)
    return d.getMonth() === currentMonth && d.getFullYear() === currentYear
  })

  for (const record of monthRecords) {
    const hasCheckIn = record.check_in || record.check_in_time
    if (hasCheckIn) {
      totalHadir++
      if (record.is_late || record.status === 'late') {
        totalTerlambat++
      }
      if (record.work_hours) {
        totalJamKerja += record.work_hours
      }
    } else {
      totalTidakHadir++
    }
  }

  const rataJam = totalHadir > 0 ? (totalJamKerja / totalHadir) : 0
  const totalHariKerja = totalHadir + totalTidakHadir

  return {
    totalHadir,
    totalTidakHadir,
    totalTerlambat,
    totalHariKerja,
    rataJamKerja: rataJam > 0 ? `${Math.floor(rataJam)}j ${Math.round((rataJam % 1) * 60)}m` : '0j'
  }
})

// Calendar attendance data
const calendarData = computed(() => {
  return attendanceHistory.value
    .filter(r => r.date)
    .map(record => {
      let status = 'absent'
      const hasCheckIn = record.check_in || record.check_in_time
      if (hasCheckIn) {
        status = (record.is_late || record.status === 'late') ? 'late' : 'present'
      }
      return {
        date: record.date,
        status
      }
    })
})

// Wi-Fi status
const wifiStatus = ref({
  isConnected: false,
  message: 'Memeriksa koneksi Wi-Fi...'
})

// GPS status
const gpsStatus = ref({
  available: false,
  message: 'GPS belum dicek'
})

// Initialize
onMounted(async () => {
  await initializeAttendance()
})

async function initializeAttendance() {
  loading.value = true
  try {
    await attendanceService.initialize()
    currentAttendance.value = await attendanceService.getCurrentAttendance()
    attendanceHistory.value = await attendanceService.getAttendanceHistory(30)
    authorizedNetworks.value = attendanceService.getAuthorizedNetworks()
    await checkWiFiStatus()
  } catch (error) {
    console.error('Failed to initialize attendance:', error)
    showToast('Gagal memuat data absensi')
  } finally {
    loading.value = false
  }
}

// Pull to refresh
async function onRefresh() {
  try {
    await initializeAttendance()
  } finally {
    refreshing.value = false
  }
}

// Calendar date selection
function onCalendarDateSelect(dateStr) {
  selectedDate.value = new Date(dateStr)
}

/**
 * Check Wi-Fi connection status
 */
async function checkWiFiStatus() {
  try {
    const isConnected = await wifiService.isConnectedToWiFi()
    const networkInfo = wifiService.getNetworkInfo()
    let message = ''
    if (isConnected) {
      if (networkInfo.type === 'wifi') {
        message = 'Terhubung ke Wi-Fi'
      } else if (networkInfo.type === 'unknown') {
        message = 'Terhubung (Desktop/WiFi terdeteksi)'
      } else {
        message = `Terhubung (${networkInfo.type})`
      }
    } else {
      message = 'Tidak terhubung ke jaringan'
    }
    wifiStatus.value = { isConnected, message, networkInfo }
  } catch (error) {
    console.error('WiFi check error:', error)
    wifiStatus.value = { isConnected: false, message: 'Gagal memeriksa status Wi-Fi' }
  }
}

/**
 * Check GPS availability
 */
async function checkGPSStatus() {
  gpsStatus.value = { available: false, message: 'Memeriksa GPS...' }
  try {
    const location = await getCurrentLocation()
    gpsStatus.value = {
      available: true,
      message: `GPS aktif (akurasi: ${Math.round(location.accuracy)}m)`,
      location
    }
    showToast('GPS tersedia dan aktif')
  } catch (error) {
    gpsStatus.value = { available: false, message: error.message }
    showToast(error.message)
  }
}

/**
 * Get current GPS location
 */
async function getCurrentLocation() {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error('Geolocation tidak didukung oleh browser ini'))
      return
    }
    navigator.geolocation.getCurrentPosition(
      (position) => {
        resolve({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          accuracy: position.coords.accuracy
        })
      },
      (error) => {
        let message = 'Gagal mendapatkan lokasi'
        switch (error.code) {
          case error.PERMISSION_DENIED:
            message = 'Akses lokasi ditolak. Mohon izinkan akses lokasi di pengaturan browser.'
            break
          case error.POSITION_UNAVAILABLE:
            message = 'Informasi lokasi tidak tersedia'
            break
          case error.TIMEOUT:
            message = 'Waktu permintaan lokasi habis'
            break
        }
        reject(new Error(message))
      },
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 60000 }
    )
  })
}

/**
 * Automatic check-in with GPS and IP-based validation
 */
async function performAutoCheckIn() {
  actionLoading.value = true
  try {
    showLoadingToast({ message: 'Mendapatkan lokasi GPS...', forbidClick: true, duration: 0 })
    let gpsLocation = null
    try {
      gpsLocation = await getCurrentLocation()
      console.log('GPS location obtained:', gpsLocation)
    } catch (gpsError) {
      console.warn('GPS location failed:', gpsError.message)
    }
    showLoadingToast({ message: 'Memvalidasi lokasi...', forbidClick: true, duration: 0 })
    const checkInData = { ssid: 'AUTO-DETECT', bssid: '00:00:00:00:00:00' }
    if (gpsLocation) {
      checkInData.latitude = gpsLocation.latitude
      checkInData.longitude = gpsLocation.longitude
    }
    console.log('Check-in request:', checkInData)
    const response = await attendanceAPI.checkIn(checkInData)
    if (response.data.success) {
      const newAttendance = response.data.data
      currentAttendance.value = newAttendance
      attendanceService.currentAttendance = newAttendance
      console.log('[CHECK-IN SUCCESS] Updated currentAttendance:', newAttendance)
      closeToast()
      const validatedBy = response.data.validated_by
      let message = 'Check-in berhasil!'
      if (validatedBy) {
        if (validatedBy.method === 'gps_validation') {
          message = `Check-in berhasil! Lokasi: ${validatedBy.details?.location_name || 'Kantor'}`
        } else if (validatedBy.method === 'ip_validation') {
          message = `Check-in berhasil! Tervalidasi melalui IP`
        }
      }
      showNotify({ type: 'success', message, duration: 3000 })
      attendanceHistory.value = await attendanceService.getAttendanceHistory(30)
      return
    }
  } catch (error) {
    console.error('Auto check-in error:', error)
    closeToast()
    let errorMessage = 'Terjadi kesalahan saat check-in'
    if (error.response) {
      if (error.response.status === 409) {
        errorMessage = 'Anda sudah melakukan check-in hari ini'
        await initializeAttendance()
      } else if (error.response.status === 403) {
        const details = error.response.data?.details
        if (details?.gps_provided) {
          errorMessage = 'Check-in gagal. Anda tidak berada di area kantor yang terdaftar.'
        } else {
          const clientIP = error.response.data?.client_ip
          errorMessage = `Check-in gagal. IP Anda (${clientIP || 'unknown'}) tidak terdaftar di jaringan kantor.`
        }
      } else if (error.response.data?.message) {
        errorMessage = error.response.data.message
      }
    } else if (error.message) {
      errorMessage = error.message
    }
    showNotify({ type: 'warning', message: errorMessage, duration: 4000 })
  } finally {
    actionLoading.value = false
  }
}

/**
 * Perform check-out
 */
async function performCheckOut() {
  showDialog({
    title: 'Konfirmasi Check-out',
    message: 'Apakah Anda yakin ingin check-out sekarang?',
    confirmButtonText: 'Ya, Check-out',
    cancelButtonText: 'Batal'
  }).then(async () => {
    actionLoading.value = true
    showLoadingToast({ message: 'Memproses check-out...', forbidClick: true })
    try {
      const result = await attendanceService.checkOut()
      if (result.success) {
        currentAttendance.value = result.attendance
        attendanceService.currentAttendance = result.attendance
        closeToast()
        showNotify({
          type: 'success',
          message: `${result.message}\nJam kerja: ${attendanceService.formatWorkHours(result.workHours)}`,
          duration: 3000
        })
        await initializeAttendance()
      } else {
        closeToast()
        showNotify({ type: 'warning', message: result.message, duration: 3000 })
      }
    } catch (error) {
      console.error('Check-out error:', error)
      closeToast()
      let errorMessage = 'Terjadi kesalahan saat check-out'
      if (error.response?.data?.message) {
        errorMessage = error.response.data.message
      }
      showNotify({ type: 'danger', message: errorMessage, duration: 3000 })
    } finally {
      actionLoading.value = false
    }
  }).catch(() => {
    // User cancelled
  })
}

/**
 * Format time for display
 */
function formatTime(timeString) {
  if (!timeString) return '--:--'
  return new Date(timeString).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

/**
 * Format date for display
 */
function formatDate(dateString) {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    weekday: 'short',
    day: 'numeric',
    month: 'short'
  })
}

/**
 * Format work hours for display
 */
function formatWorkHours(hours) {
  return attendanceService.formatWorkHours(hours)
}
</script>

<style scoped>
.attendance-page {
  min-height: 100vh;
  background: var(--h-bg-primary);
  padding-bottom: 24px;
}

/* Header */
.attendance-header {
  padding: var(--h-spacing-lg);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--h-spacing-md);
}

.attendance-header__name-row {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-sm);
  flex-wrap: wrap;
}

.attendance-header__name {
  font-size: 18px;
  font-weight: 700;
  color: var(--h-text-primary);
  margin: 0;
  line-height: 1.3;
}

.attendance-header__role {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  background: var(--h-primary);
  color: #ffffff;
  font-size: 11px;
  font-weight: 600;
  border-radius: 12px;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.attendance-header__date {
  font-size: 13px;
  color: var(--h-text-secondary);
  margin: 2px 0 0;
}

/* Status Cards (two-column) */
.status-cards {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--h-spacing-md);
  padding: 0 var(--h-spacing-lg);
  margin-bottom: var(--h-spacing-lg);
}

.status-card {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--h-spacing-sm);
}

.status-card__icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: var(--h-radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-card__icon-wrap--green {
  background: var(--h-success);
}

.status-card__icon-wrap--red {
  background: var(--h-error);
}

.status-card__label {
  font-size: 13px;
  font-weight: 600;
  color: var(--h-text-secondary);
}

.status-card__time {
  font-size: 22px;
  font-weight: 700;
  color: var(--h-text-primary);
}

/* Swipe Section */
.swipe-section {
  margin-bottom: var(--h-spacing-lg);
}

/* Summary Cards (horizontal scroll) */
.summary-scroll {
  display: flex;
  gap: var(--h-spacing-md);
  padding: 0 var(--h-spacing-lg);
  margin-bottom: var(--h-spacing-xl);
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.summary-scroll::-webkit-scrollbar {
  display: none;
}

.summary-scroll > * {
  min-width: 140px;
  flex-shrink: 0;
}

/* Section Block */
.section-block {
  padding: 0 var(--h-spacing-lg);
  margin-bottom: var(--h-spacing-xl);
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--h-text-primary);
  margin: 0 0 var(--h-spacing-md);
}

/* Statistics Grid */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--h-spacing-md);
}

.stat-item {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-lg);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--h-spacing-xs);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
}

.stat-value--success {
  color: var(--h-success);
}

.stat-value--error {
  color: var(--h-error);
}

.stat-value--warning {
  color: var(--h-warning);
}

.stat-value--primary {
  color: var(--h-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--h-text-secondary);
  font-weight: 500;
}

/* History List */
.history-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-md);
}

.history-card {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-lg);
}

.history-card__date {
  font-size: 14px;
  font-weight: 600;
  color: var(--h-text-primary);
  margin-bottom: var(--h-spacing-sm);
}

.history-card__times {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-sm);
  flex-wrap: wrap;
  font-size: 13px;
  color: var(--h-text-secondary);
}

.history-card__in,
.history-card__out {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.history-card__separator {
  color: var(--h-text-light);
}

.history-card__duration {
  color: var(--h-primary);
  font-weight: 600;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--h-spacing-sm);
  padding: var(--h-spacing-2xl);
  color: var(--h-text-light);
  font-size: 14px;
}

/* Responsive */
@media (max-width: 375px) {
  .attendance-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .summary-scroll > * {
    min-width: 120px;
  }
}
</style>
