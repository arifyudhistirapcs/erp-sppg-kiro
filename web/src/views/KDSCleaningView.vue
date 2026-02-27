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

      <a-spin :spinning="loading" tip="Memuat data...">
        <a-empty 
          v-if="!loading && pendingOmpreng.length === 0" 
          description="Tidak ada ompreng yang perlu dicuci"
        />
        
        <a-table
          v-else
          :columns="columns"
          :data-source="pendingOmpreng"
          :pagination="false"
          :row-key="record => record.id"
          :loading="loading"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'school_name'">
              <strong>{{ record.school_name }}</strong>
            </template>
            
            <template v-else-if="column.key === 'delivery_date'">
              {{ formatDate(record.delivery_date) }}
            </template>
            
            <template v-else-if="column.key === 'ompreng_count'">
              <a-tag color="blue">{{ record.ompreng_count }} unit</a-tag>
            </template>
            
            <template v-else-if="column.key === 'status'">
              <a-tag :color="getStatusColor(record.cleaning_status)">
                {{ getStatusText(record.cleaning_status) }}
              </a-tag>
            </template>
            
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button
                  v-if="record.cleaning_status === 'pending'"
                  type="primary"
                  size="small"
                  @click="handleStartCleaning(record)"
                  :loading="updatingId === record.id"
                >
                  <template #icon><play-circle-outlined /></template>
                  Mulai Cuci
                </a-button>
                
                <a-button
                  v-else-if="record.cleaning_status === 'in_progress'"
                  type="primary"
                  size="small"
                  @click="handleCompleteCleaning(record)"
                  :loading="updatingId === record.id"
                  style="background-color: #52c41a; border-color: #52c41a"
                >
                  <template #icon><check-circle-outlined /></template>
                  Selesai
                </a-button>
                
                <a-tag v-else color="success">
                  <template #icon><check-outlined /></template>
                  Sudah Selesai
                </a-tag>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-spin>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
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

// Table columns
const columns = [
  {
    title: 'Sekolah',
    dataIndex: 'school_name',
    key: 'school_name',
    width: '25%'
  },
  {
    title: 'Tanggal Pengiriman',
    dataIndex: 'delivery_date',
    key: 'delivery_date',
    width: '20%'
  },
  {
    title: 'Jumlah Ompreng',
    dataIndex: 'ompreng_count',
    key: 'ompreng_count',
    width: '15%',
    align: 'center'
  },
  {
    title: 'Status',
    key: 'status',
    width: '20%',
    align: 'center'
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: '20%',
    align: 'center'
  }
]

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
  const date = new Date(dateStr)
  return date.toLocaleDateString('id-ID', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
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
  background-color: white;
  padding: 24px;
  border-radius: 8px;
}

:deep(.ant-table) {
  background-color: white;
}

:deep(.ant-table-thead > tr > th) {
  background-color: #fafafa;
  font-weight: 600;
}

:deep(.ant-table-tbody > tr:hover > td) {
  background-color: #f5f5f5;
}
</style>
