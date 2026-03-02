<template>
  <div class="kds-cleaning-view">
    <div class="kds-header">
      <div class="header-content">
        <div class="header-left">
          <h2 class="header-title">Kebersihan - Pencucian Ompreng</h2>
          <p class="header-subtitle">Kelola pencucian ompreng yang kembali dari sekolah</p>
        </div>
        <div class="header-right">
          <a-space :size="12">
            <KDSDatePicker
              v-model="selectedDate"
              :loading="loading"
              @change="handleDateChange"
            />
            <a-tag :color="isConnected ? 'green' : 'red'" class="connection-tag">
              <template #icon>
                <wifi-outlined v-if="isConnected" />
                <disconnect-outlined v-else />
              </template>
              {{ isConnected ? 'Terhubung' : 'Terputus' }}
            </a-tag>
            <a-button @click="refreshData" :loading="loading" type="default">
              <template #icon><reload-outlined /></template>
              Refresh
            </a-button>
          </a-space>
        </div>
      </div>
    </div>

    <div class="content-wrapper">
      <a-alert
        v-if="error"
        type="error"
        :message="error"
        closable
        show-icon
        @close="error = null"
        style="margin-bottom: 16px"
      >
        <template #action>
          <a-button size="small" type="primary" @click="retryLoad">
            Coba Lagi
          </a-button>
        </template>
      </a-alert>

      <a-alert
        v-if="!loading && allOmprengCompleted"
        type="success"
        message="Semua Ompreng Selesai Dicuci!"
        description="Semua ompreng telah selesai dicuci dan siap untuk digunakan kembali."
        show-icon
        closable
        style="margin-bottom: 16px"
      />

      <a-spin :spinning="loading" tip="Memuat data...">
        <a-empty 
          v-if="!loading && pendingOmpreng.length === 0" 
          description="Tidak ada ompreng yang perlu dicuci"
        />
        
        <a-row :gutter="[16, 16]" v-else>
          <a-col
            v-for="record in pendingOmpreng"
            :key="record.id"
            :xs="24"
            :sm="24"
            :md="12"
            :lg="8"
            :xl="6"
          >
            <a-card
              :class="['cleaning-card', `status-${record.cleaning_status}`]"
            >
              <div class="card-header">
                <div class="school-name">{{ record.school_name || record.delivery_record?.school?.name || '-' }}</div>
                <a-tag :color="getStatusColor(record.cleaning_status)" class="status-tag">
                  {{ getStatusText(record.cleaning_status) }}
                </a-tag>
              </div>

              <div class="cleaning-info">
                <a-statistic
                  title="Jumlah Ompreng"
                  :value="record.ompreng_count"
                  suffix="unit"
                  :value-style="{ color: '#1890ff', fontSize: '28px', fontWeight: 'bold' }"
                />

                <a-divider>Informasi Pengiriman</a-divider>
                
                <div class="info-item">
                  <span class="info-label">Tanggal Pengiriman:</span>
                  <span class="info-value">{{ formatDate(record.delivery_date || record.delivery_record?.delivery_date) }}</span>
                </div>

                <div v-if="record.started_at" class="info-item">
                  <span class="info-label">Mulai Cuci:</span>
                  <span class="info-value">{{ formatDateTime(record.started_at) }}</span>
                </div>

                <div v-if="record.completed_at" class="info-item">
                  <span class="info-label">Selesai:</span>
                  <span class="info-value">{{ formatDateTime(record.completed_at) }}</span>
                </div>
              </div>

              <template #actions>
                <a-button
                  v-if="record.cleaning_status === 'pending'"
                  type="primary"
                  block
                  @click="handleStartCleaning(record)"
                  :loading="updatingId === record.id"
                >
                  <template #icon><play-circle-outlined /></template>
                  Mulai Cuci
                </a-button>
                
                <a-button
                  v-else-if="record.cleaning_status === 'in_progress'"
                  type="primary"
                  block
                  @click="handleCompleteCleaning(record)"
                  :loading="updatingId === record.id"
                  style="background-color: #52c41a; border-color: #52c41a"
                >
                  <template #icon><check-circle-outlined /></template>
                  Selesai
                </a-button>
                
                <a-tag v-else color="success" style="width: 100%; text-align: center; padding: 8px 0;">
                  <template #icon><check-outlined /></template>
                  Sudah Selesai
                </a-tag>
              </template>
            </a-card>
          </a-col>
        </a-row>
      </a-spin>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CheckOutlined
} from '@ant-design/icons-vue'
import { getPendingOmpreng, startCleaning, completeCleaning } from '@/services/cleaningService'
import KDSDatePicker from '@/components/KDSDatePicker.vue'

// Try to import Firebase, but make it optional
let database = null
let dbRef = null
let onValue = null
let off = null

// Dynamically import Firebase modules
const initFirebase = async () => {
  try {
    const firebaseModule = await import('@/services/firebase')
    const firebaseDatabase = await import('firebase/database')
    database = firebaseModule.database
    dbRef = firebaseDatabase.ref
    onValue = firebaseDatabase.onValue
    off = firebaseDatabase.off
    console.log('[KDS Cleaning] Firebase modules loaded successfully')
    return true
  } catch (error) {
    console.warn('[KDS Cleaning] Firebase not configured, real-time updates disabled:', error.message)
    return false
  }
}

// State
const pendingOmpreng = ref([])
const loading = ref(false)
const updatingId = ref(null)
const isConnected = ref(true)
const error = ref(null)
const selectedDate = ref(new Date().toISOString().split('T')[0]) // Format: YYYY-MM-DD
let firebaseListener = null

// Check if all ompreng are completed
const allOmprengCompleted = computed(() => {
  if (pendingOmpreng.value.length === 0) return false
  return pendingOmpreng.value.every(ompreng => ompreng.cleaning_status === 'completed')
})

// Get status color
const getStatusColor = (status) => {
  const colors = {
    pending: 'default',
    in_progress: 'processing',
    completed: 'success'
  }
  return colors[status] || 'default'
}

// Get status text in Indonesian
const getStatusText = (status) => {
  const texts = {
    pending: 'Menunggu',
    in_progress: 'Sedang Dicuci',
    completed: 'Selesai'
  }
  return texts[status] || status
}

// Format date to readable format
const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  
  // Parse the date string directly without timezone conversion
  let date
  if (typeof dateStr === 'number') {
    date = new Date(dateStr < 10000000000 ? dateStr * 1000 : dateStr)
  } else {
    // Parse as UTC and display as-is (no timezone conversion)
    date = new Date(dateStr)
  }
  
  if (isNaN(date.getTime())) return '-'
  
  // Format using UTC to avoid timezone conversion
  const year = date.getUTCFullYear()
  const month = date.getUTCMonth()
  const day = date.getUTCDate()
  const dayOfWeek = date.getUTCDay()
  
  const dayNames = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu']
  const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  
  return `${dayNames[dayOfWeek]}, ${day} ${monthNames[month]} ${year}`
}

// Format date time to readable format
const formatDateTime = (dateStr) => {
  if (!dateStr) return '-'
  
  // Parse the date string directly without timezone conversion
  let date
  if (typeof dateStr === 'number') {
    date = new Date(dateStr < 10000000000 ? dateStr * 1000 : dateStr)
  } else {
    // Parse as UTC and display as-is (no timezone conversion)
    date = new Date(dateStr)
  }
  
  if (isNaN(date.getTime())) return '-'
  
  // Format using UTC to avoid timezone conversion
  const year = date.getUTCFullYear()
  const month = date.getUTCMonth()
  const day = date.getUTCDate()
  const hours = date.getUTCHours()
  const minutes = date.getUTCMinutes()
  
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  
  return `${day} ${monthNames[month]} ${year}, ${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`
}

// Load data from API
const loadData = async (date = null) => {
  loading.value = true
  error.value = null
  const dateToLoad = date || selectedDate.value
  console.log('[KDS Cleaning] Loading pending ompreng data for date:', dateToLoad)
  try {
    const response = await getPendingOmpreng(dateToLoad)
    console.log('[KDS Cleaning] API Response:', response)
    if (response.success) {
      pendingOmpreng.value = response.data || []
      console.log('[KDS Cleaning] Loaded ompreng:', pendingOmpreng.value.length)
    } else {
      error.value = response.message || 'Gagal memuat data'
    }
  } catch (err) {
    console.error('Error loading cleaning data:', err)
    error.value = err.response?.data?.message || 'Gagal memuat data. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
}

// Handle date change
const handleDateChange = (newDate) => {
  selectedDate.value = newDate
  loadData(newDate)
}

// Retry loading data
const retryLoad = () => {
  loadData()
}

// Refresh data
const refreshData = () => {
  loadData()
}

// Handle start cleaning
const handleStartCleaning = async (record) => {
  updatingId.value = record.id
  try {
    const response = await startCleaning(record.id)
    if (response.success) {
      message.success('Pencucian dimulai')
      // Reload data from API to get updated status
      await loadData()
    } else {
      message.error(response.message || 'Gagal memulai pencucian')
    }
  } catch (error) {
    console.error('Error starting cleaning:', error)
    message.error(error.response?.data?.message || 'Gagal memulai pencucian')
  } finally {
    updatingId.value = null
  }
}

// Handle complete cleaning
const handleCompleteCleaning = async (record) => {
  updatingId.value = record.id
  try {
    const response = await completeCleaning(record.id)
    if (response.success) {
      message.success('Pencucian selesai')
      // Reload data from API to get updated status
      await loadData()
    } else {
      message.error(response.message || 'Gagal menyelesaikan pencucian')
    }
  } catch (error) {
    console.error('Error completing cleaning:', error)
    message.error(error.response?.data?.message || 'Gagal menyelesaikan pencucian')
  } finally {
    updatingId.value = null
  }
}

// Setup Firebase real-time listener
const setupFirebaseListener = () => {
  // Skip if Firebase is not available
  if (!database || !dbRef || !onValue) {
    console.warn('[KDS Cleaning] Firebase not available, skipping real-time listener setup')
    isConnected.value = false
    return
  }
  
  try {
    // Clean up existing listener first
    cleanupFirebaseListener()
    
    const cleaningRef = dbRef(database, '/cleaning/pending')
    
    console.log('[KDS Cleaning] Setting up Firebase listener for path: /cleaning/pending')
    
    firebaseListener = onValue(
      cleaningRef,
      (snapshot) => {
        isConnected.value = true
        const data = snapshot.val()
        
        console.log('[KDS Cleaning] Firebase data received:', data)
        
        if (data) {
          // Update ompreng list with Firebase data
          const firebaseOmpreng = Object.values(data)
          
          console.log('[KDS Cleaning] Firebase ompreng:', firebaseOmpreng)
          
          // Merge with existing data to preserve all fields
          pendingOmpreng.value = pendingOmpreng.value.map(ompreng => {
            const firebaseRecord = firebaseOmpreng.find(fo => fo.id === ompreng.id)
            if (firebaseRecord) {
              console.log('[KDS Cleaning] Updating ompreng', ompreng.id, 'with status:', firebaseRecord.status)
              return {
                ...ompreng,
                cleaning_status: firebaseRecord.status,
                started_at: firebaseRecord.started_at,
                completed_at: firebaseRecord.completed_at
              }
            }
            return ompreng
          })
          
          console.log('[KDS Cleaning] Updated ompreng:', pendingOmpreng.value)
        }
      },
      (error) => {
        console.warn('[KDS Cleaning] Firebase listener error (permission denied - this is expected):', error.code)
        // Don't show error to user, just disable real-time updates
        isConnected.value = false
        // Clean up the listener to prevent repeated errors
        cleanupFirebaseListener()
      }
    )
  } catch (error) {
    console.error('[KDS Cleaning] Failed to setup Firebase listener:', error)
    isConnected.value = false
  }
}

// Cleanup Firebase listener
const cleanupFirebaseListener = () => {
  if (firebaseListener && database && dbRef && off) {
    try {
      const cleaningRef = dbRef(database, '/cleaning/pending')
      off(cleaningRef)
      firebaseListener = null
    } catch (error) {
      console.error('[KDS Cleaning] Error cleaning up Firebase listener:', error)
    }
  }
}

onMounted(async () => {
  loadData()
  // Initialize Firebase and setup listener
  await initFirebase()
  setupFirebaseListener()
})

onUnmounted(() => {
  cleanupFirebaseListener()
})
</script>

<style scoped>
.kds-cleaning-view {
  background-color: #f0f2f5;
  min-height: 100vh;
}

.kds-header {
  background: white;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1600px;
  margin: 0 auto;
}

.header-left {
  flex: 1;
}

.header-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  line-height: 1.4;
}

.header-subtitle {
  margin: 4px 0 0 0;
  font-size: 14px;
  color: #8c8c8c;
}

.header-right {
  display: flex;
  align-items: center;
}

.connection-tag {
  font-size: 13px;
  padding: 4px 12px;
  border-radius: 4px;
}

.content-wrapper {
  max-width: 1600px;
  margin: 24px auto;
  padding: 0 24px;
}

/* Card Styles */
.cleaning-card {
  height: 100%;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  overflow: hidden;
}

.cleaning-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.cleaning-card.status-pending {
  border-left: 4px solid #d9d9d9;
}

.cleaning-card.status-in_progress {
  border-left: 4px solid #1890ff;
}

.cleaning-card.status-completed {
  border-left: 4px solid #52c41a;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.school-name {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
  flex: 1;
  margin-right: 12px;
  line-height: 1.4;
}

.status-tag {
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 4px;
  flex-shrink: 0;
}

.cleaning-info {
  margin-top: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f0f0f0;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 13px;
  color: #8c8c8c;
  font-weight: 500;
}

.info-value {
  font-size: 13px;
  color: #262626;
  font-weight: 600;
}

:deep(.ant-statistic-title) {
  font-size: 13px;
  color: #8c8c8c;
  margin-bottom: 4px;
}

:deep(.ant-statistic-content) {
  font-size: 28px;
  line-height: 1.2;
}

:deep(.ant-divider) {
  margin: 16px 0;
  font-size: 13px;
  color: #8c8c8c;
}

:deep(.ant-card-actions) {
  background-color: #fafafa;
  border-top: 1px solid #f0f0f0;
}

:deep(.ant-card-actions > li) {
  margin: 8px 0;
}

:deep(.ant-card-actions > li > span) {
  display: block;
  padding: 0 12px;
}
</style>
