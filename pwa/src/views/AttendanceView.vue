<template>
  <div class="attendance-container">
    <van-nav-bar title="Absensi" />
    
    <!-- Current Status Card -->
    <div class="status-card">
      <van-card>
        <template #title>
          <div class="status-header">
            <van-icon name="clock-o" />
            <span>Status Absensi Hari Ini</span>
          </div>
        </template>
        <template #desc>
          <div class="status-info">
            <div v-if="currentAttendance">
              <div class="time-info">
                <div v-if="currentAttendance.check_in_time" class="check-in">
                  <van-icon name="play-circle-o" color="#07c160" />
                  <span>Masuk: {{ formatTime(currentAttendance.check_in_time) }}</span>
                </div>
                <div v-if="currentAttendance.check_out_time" class="check-out">
                  <van-icon name="stop-circle-o" color="#ee0a24" />
                  <span>Keluar: {{ formatTime(currentAttendance.check_out_time) }}</span>
                </div>
                <div v-if="currentAttendance.work_hours" class="work-hours">
                  <van-icon name="clock" color="#1989fa" />
                  <span>Jam Kerja: {{ formatWorkHours(currentAttendance.work_hours) }}</span>
                </div>
              </div>
            </div>
            <div v-else class="no-attendance">
              <van-icon name="info-o" />
              <span>Belum ada absensi hari ini</span>
            </div>
          </div>
        </template>
      </van-card>
    </div>

    <!-- Wi-Fi Status -->
    <van-cell-group title="Status Koneksi">
      <van-cell>
        <template #title>
          <div class="wifi-status">
            <van-icon :name="wifiStatus.isConnected ? 'wifi' : 'wifi-off'" 
                     :color="wifiStatus.isConnected ? '#07c160' : '#ee0a24'" />
            <span>{{ wifiStatus.message }}</span>
          </div>
        </template>
        <template #right-icon>
          <van-button size="mini" @click="checkWiFiStatus">Refresh</van-button>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- Action Buttons -->
    <div class="action-buttons">
      <van-button 
        v-if="canCheckIn" 
        type="primary" 
        size="large" 
        block 
        :loading="loading"
        @click="showCheckInOptions"
        class="check-in-btn">
        <van-icon name="play-circle-o" />
        Check In
      </van-button>
      
      <van-button 
        v-if="canCheckOut" 
        type="danger" 
        size="large" 
        block 
        :loading="loading"
        @click="performCheckOut"
        class="check-out-btn">
        <van-icon name="stop-circle-o" />
        Check Out
      </van-button>
    </div>

    <!-- Check-in Options Popup -->
    <van-popup v-model:show="showCheckInPopup" position="bottom" :style="{ height: '60%' }">
      <div class="check-in-popup">
        <van-nav-bar title="Pilih Metode Check-in" @click-left="showCheckInPopup = false">
          <template #left>
            <van-icon name="cross" />
          </template>
        </van-nav-bar>
        
        <div class="popup-content">
          <van-cell-group title="Metode Validasi Wi-Fi">
            <van-cell 
              title="Deteksi Otomatis" 
              label="Gunakan deteksi jaringan otomatis"
              is-link
              @click="checkInWithAutoDetection">
              <template #icon>
                <van-icon name="wifi" color="#1989fa" />
              </template>
            </van-cell>
            
            <van-cell 
              title="Input SSID Manual" 
              label="Masukkan nama Wi-Fi secara manual"
              is-link
              @click="showManualSSIDInput = true">
              <template #icon>
                <van-icon name="edit" color="#ff976a" />
              </template>
            </van-cell>
            
            <van-cell 
              title="Validasi GPS" 
              label="Gunakan lokasi GPS untuk validasi"
              is-link
              @click="checkInWithGPS">
              <template #icon>
                <van-icon name="location-o" color="#07c160" />
              </template>
            </van-cell>
          </van-cell-group>

          <div class="authorized-networks">
            <van-cell-group title="Jaringan Wi-Fi yang Diotorisasi">
              <van-cell 
                v-for="network in authorizedNetworks" 
                :key="network.ssid"
                :title="network.ssid"
                :label="network.location">
                <template #icon>
                  <van-icon name="wifi" color="#07c160" />
                </template>
              </van-cell>
            </van-cell-group>
          </div>
        </div>
      </div>
    </van-popup>

    <!-- Manual SSID Input -->
    <van-popup v-model:show="showManualSSIDInput" position="center">
      <div class="manual-ssid-popup">
        <van-cell-group title="Masukkan SSID Wi-Fi">
          <van-field
            v-model="manualSSID"
            label="SSID"
            placeholder="Nama jaringan Wi-Fi"
            :rules="[{ required: true, message: 'SSID harus diisi' }]"
          />
        </van-cell-group>
        
        <div class="popup-buttons">
          <van-button @click="showManualSSIDInput = false">Batal</van-button>
          <van-button type="primary" @click="checkInWithManualSSID">Check In</van-button>
        </div>
      </div>
    </van-popup>

    <!-- Attendance History -->
    <van-cell-group title="Riwayat Absensi (7 hari terakhir)">
      <van-cell 
        v-for="record in attendanceHistory" 
        :key="record.id"
        :title="formatDate(record.date)"
        :label="getAttendanceLabel(record)">
        <template #right-icon>
          <div class="history-hours">
            {{ record.work_hours ? formatWorkHours(record.work_hours) : '-' }}
          </div>
        </template>
      </van-cell>
      
      <van-cell v-if="attendanceHistory.length === 0" title="Belum ada riwayat absensi" />
    </van-cell-group>

    <van-tabbar v-model="active" route>
      <van-tabbar-item to="/tasks" icon="orders-o">Tugas</van-tabbar-item>
      <van-tabbar-item to="/attendance" icon="clock-o">Absensi</van-tabbar-item>
      <van-tabbar-item to="/profile" icon="user-o">Profil</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { showToast, showDialog, showLoadingToast, closeToast } from 'vant'
import attendanceService from '@/services/attendanceService.js'
import wifiService from '@/services/wifiService.js'

const active = ref(1)
const loading = ref(false)
const currentAttendance = ref(null)
const attendanceHistory = ref([])
const authorizedNetworks = ref([])

// Wi-Fi status
const wifiStatus = ref({
  isConnected: false,
  message: 'Memeriksa koneksi Wi-Fi...'
})

// Popup states
const showCheckInPopup = ref(false)
const showManualSSIDInput = ref(false)
const manualSSID = ref('')

// Computed properties
const canCheckIn = computed(() => {
  return attendanceService.canCheckIn()
})

const canCheckOut = computed(() => {
  return attendanceService.canCheckOut()
})

// Initialize component
onMounted(async () => {
  await initializeAttendance()
})

/**
 * Initialize attendance data
 */
async function initializeAttendance() {
  loading.value = true
  
  try {
    // Initialize services
    await attendanceService.initialize()
    
    // Load current attendance
    currentAttendance.value = await attendanceService.getCurrentAttendance()
    
    // Load attendance history
    attendanceHistory.value = await attendanceService.getAttendanceHistory(7)
    
    // Load authorized networks
    authorizedNetworks.value = attendanceService.getAuthorizedNetworks()
    
    // Check Wi-Fi status
    await checkWiFiStatus()
    
  } catch (error) {
    console.error('Failed to initialize attendance:', error)
    showToast('Gagal memuat data absensi')
  } finally {
    loading.value = false
  }
}

/**
 * Check Wi-Fi connection status
 */
async function checkWiFiStatus() {
  try {
    const isConnected = await wifiService.isConnectedToWiFi()
    const networkInfo = wifiService.getNetworkInfo()
    
    wifiStatus.value = {
      isConnected,
      message: isConnected 
        ? `Terhubung ke Wi-Fi (${networkInfo.type})` 
        : 'Tidak terhubung ke Wi-Fi'
    }
  } catch (error) {
    wifiStatus.value = {
      isConnected: false,
      message: 'Gagal memeriksa status Wi-Fi'
    }
  }
}

/**
 * Show check-in options
 */
function showCheckInOptions() {
  if (!wifiStatus.value.isConnected) {
    showDialog({
      title: 'Wi-Fi Tidak Aktif',
      message: 'Pastikan Wi-Fi aktif dan terhubung ke jaringan kantor sebelum check-in.',
      confirmButtonText: 'Coba Lagi',
      cancelButtonText: 'Lanjut Manual'
    }).then(() => {
      checkWiFiStatus()
    }).catch(() => {
      showCheckInPopup.value = true
    })
    return
  }
  
  showCheckInPopup.value = true
}

/**
 * Check-in with automatic Wi-Fi detection
 */
async function checkInWithAutoDetection() {
  showCheckInPopup.value = false
  loading.value = true
  
  const loadingToast = showLoadingToast({
    message: 'Memvalidasi Wi-Fi...',
    forbidClick: true
  })
  
  try {
    const result = await attendanceService.checkIn()
    
    if (result.success) {
      currentAttendance.value = result.attendance
      showToast({
        message: result.message,
        type: 'success'
      })
    } else {
      showDialog({
        title: result.error,
        message: result.message + (result.details ? '\n\n' + result.details : ''),
        confirmButtonText: 'OK'
      })
    }
  } catch (error) {
    showToast({
      message: 'Terjadi kesalahan saat check-in',
      type: 'fail'
    })
  } finally {
    closeToast()
    loading.value = false
  }
}

/**
 * Check-in with manual SSID input
 */
async function checkInWithManualSSID() {
  if (!manualSSID.value.trim()) {
    showToast('SSID harus diisi')
    return
  }
  
  showManualSSIDInput.value = false
  showCheckInPopup.value = false
  loading.value = true
  
  const loadingToast = showLoadingToast({
    message: 'Memvalidasi SSID...',
    forbidClick: true
  })
  
  try {
    const result = await attendanceService.checkIn(manualSSID.value)
    
    if (result.success) {
      currentAttendance.value = result.attendance
      showToast({
        message: result.message,
        type: 'success'
      })
      manualSSID.value = ''
    } else {
      showDialog({
        title: result.error,
        message: result.message + (result.details ? '\n\n' + result.details : ''),
        confirmButtonText: 'OK'
      })
    }
  } catch (error) {
    showToast({
      message: 'Terjadi kesalahan saat check-in',
      type: 'fail'
    })
  } finally {
    closeToast()
    loading.value = false
  }
}

/**
 * Check-in with GPS validation
 */
async function checkInWithGPS() {
  showCheckInPopup.value = false
  loading.value = true
  
  const loadingToast = showLoadingToast({
    message: 'Mendapatkan lokasi GPS...',
    forbidClick: true
  })
  
  try {
    const result = await attendanceService.checkIn(null, true)
    
    if (result.success) {
      currentAttendance.value = result.attendance
      showToast({
        message: result.message,
        type: 'success'
      })
    } else {
      showDialog({
        title: result.error,
        message: result.message + (result.details ? '\n\n' + result.details : ''),
        confirmButtonText: 'OK'
      })
    }
  } catch (error) {
    showToast({
      message: 'Terjadi kesalahan saat check-in',
      type: 'fail'
    })
  } finally {
    closeToast()
    loading.value = false
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
    loading.value = true
    
    const loadingToast = showLoadingToast({
      message: 'Memproses check-out...',
      forbidClick: true
    })
    
    try {
      const result = await attendanceService.checkOut()
      
      if (result.success) {
        currentAttendance.value = result.attendance
        showToast({
          message: `${result.message}\nJam kerja: ${attendanceService.formatWorkHours(result.workHours)}`,
          type: 'success',
          duration: 3000
        })
      } else {
        showDialog({
          title: result.error,
          message: result.message,
          confirmButtonText: 'OK'
        })
      }
    } catch (error) {
      showToast({
        message: 'Terjadi kesalahan saat check-out',
        type: 'fail'
      })
    } finally {
      closeToast()
      loading.value = false
    }
  }).catch(() => {
    // User cancelled
  })
}

/**
 * Format time for display
 */
function formatTime(timeString) {
  if (!timeString) return '-'
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
    day: '2-digit',
    month: '2-digit'
  })
}

/**
 * Format work hours for display
 */
function formatWorkHours(hours) {
  return attendanceService.formatWorkHours(hours)
}

/**
 * Get attendance label for history
 */
function getAttendanceLabel(record) {
  const checkIn = record.check_in_time ? formatTime(record.check_in_time) : '-'
  const checkOut = record.check_out_time ? formatTime(record.check_out_time) : '-'
  return `Masuk: ${checkIn} | Keluar: ${checkOut}`
}
</script>

<style scoped>
.attendance-container {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-bottom: 50px;
}

.status-card {
  margin: 16px;
  margin-bottom: 20px;
}

.status-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #323233;
}

.status-info {
  margin-top: 12px;
}

.time-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.check-in, .check-out, .work-hours {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
}

.no-attendance {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #969799;
  font-size: 14px;
}

.wifi-status {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-buttons {
  margin: 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.check-in-btn, .check-out-btn {
  height: 50px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.check-in-popup {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.popup-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.authorized-networks {
  margin-top: 20px;
}

.manual-ssid-popup {
  width: 300px;
  padding: 20px;
  border-radius: 8px;
}

.popup-buttons {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 20px;
}

.popup-buttons .van-button {
  flex: 1;
}

.history-hours {
  font-size: 12px;
  color: #1989fa;
  font-weight: 600;
}

/* Responsive adjustments */
@media (max-width: 375px) {
  .status-card {
    margin: 12px;
  }
  
  .action-buttons {
    margin: 16px 12px;
  }
  
  .popup-content {
    padding: 12px;
  }
}

/* Loading states */
.van-button--loading {
  opacity: 0.7;
}

/* Custom van-cell styling */
:deep(.van-cell__title) {
  font-weight: 500;
}

:deep(.van-cell__label) {
  color: #969799;
  font-size: 12px;
}

/* Popup styling */
:deep(.van-popup) {
  border-radius: 16px 16px 0 0;
}

:deep(.van-nav-bar) {
  background-color: #fff;
  border-bottom: 1px solid #ebedf0;
}

/* Card styling */
:deep(.van-card) {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* Cell group styling */
:deep(.van-cell-group) {
  margin: 16px;
  border-radius: 8px;
  overflow: hidden;
}

:deep(.van-cell-group__title) {
  padding: 16px 16px 8px;
  color: #646566;
  font-size: 14px;
  font-weight: 600;
}
</style>
