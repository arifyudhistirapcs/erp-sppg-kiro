<template>
  <div class="kds-packing-view">
    <div class="kds-header">
      <div class="header-content">
        <div class="header-left">
          <h2 class="header-title">Packing - Pengemasan</h2>
          <p class="header-subtitle">Alokasi porsi per sekolah</p>
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
            <a-badge :count="readyCount" :number-style="{ backgroundColor: '#52c41a' }">
              <a-button @click="refreshData" :loading="loading" type="default">
                <template #icon><reload-outlined /></template>
                Refresh
              </a-button>
            </a-badge>
          </a-space>
        </div>
      </div>
    </div>
    <!-- All Ready Notification -->
    <a-alert
      v-if="allSchoolsReady && schools.length > 0"
      message="Semua Sekolah Siap Kirim!"
      description="Semua sekolah telah selesai dikemas dan siap untuk pengiriman."
      type="success"
      show-icon
      closable
      style="margin-bottom: 16px"
    >
      <template #icon>
        <check-circle-outlined />
      </template>
    </a-alert>

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
        <a-empty v-if="!loading && schools.length === 0" :description="emptyMessage" />
        
        <a-row :gutter="[16, 16]" v-else>
          <a-col
            v-for="school in schools"
            :key="school.school_id"
            :xs="24"
            :sm="24"
            :md="12"
            :lg="8"
            :xl="6"
          >
            <a-card
              :class="['school-card', `status-${school.status}`]"
            >
              <div class="card-header">
                <div class="school-name">{{ school.school_name }}</div>
                <a-tag :color="getStatusColor(school.status)" class="status-tag">
                  {{ getStatusText(school.status) }}
                </a-tag>
              </div>

              <div class="school-info">
                <a-statistic
                  title="Total Porsi"
                  :value="school.total_portions"
                  suffix="porsi"
                  :value-style="{ color: '#1890ff', fontSize: '28px', fontWeight: 'bold' }"
                />

                <!-- Portion Size Breakdown -->
                <div v-if="school.portion_size_type === 'mixed'" class="portion-breakdown">
                  <div class="portion-breakdown-title">Rincian Ukuran Porsi</div>
                  <a-row :gutter="12">
                    <a-col :span="12">
                      <div class="portion-size-card small">
                        <div class="portion-icon">S</div>
                        <div class="portion-label">Kecil (Kelas 1-3)</div>
                        <div class="portion-value">{{ school.portions_small }} porsi</div>
                      </div>
                    </a-col>
                    <a-col :span="12">
                      <div class="portion-size-card large">
                        <div class="portion-icon">L</div>
                        <div class="portion-label">Besar (Kelas 4-6)</div>
                        <div class="portion-value">{{ school.portions_large }} porsi</div>
                      </div>
                    </a-col>
                  </a-row>
                </div>
                <div v-else class="portion-breakdown">
                  <div class="portion-breakdown-title">Rincian Ukuran Porsi</div>
                  <div class="portion-size-card large single">
                    <div class="portion-icon">L</div>
                    <div class="portion-label">Porsi Besar</div>
                    <div class="portion-value">{{ school.portions_large }} porsi</div>
                  </div>
                </div>

                <a-divider>Menu Items</a-divider>
                <a-list
                  size="small"
                  :data-source="school.menu_items"
                  :split="false"
                >
                  <template #renderItem="{ item }">
                    <a-list-item>
                      <a-list-item-meta>
                        <template #avatar>
                          <a-avatar
                            v-if="item.photo_url"
                            :src="item.photo_url"
                            shape="square"
                            :size="48"
                          />
                          <a-avatar
                            v-else
                            shape="square"
                            :size="48"
                            style="background-color: #f0f0f0; color: #999"
                          >
                            <template #icon><picture-outlined /></template>
                          </a-avatar>
                        </template>
                        <template #title>
                          {{ item.recipe_name }}
                        </template>
                        <template #description>
                          <div class="menu-item-portions">
                            <a-tag v-if="item.portions_small > 0" color="cyan" class="portion-tag">
                              Kecil: {{ item.portions_small }}
                            </a-tag>
                            <a-tag v-if="item.portions_large > 0" color="blue" class="portion-tag">
                              Besar: {{ item.portions_large }}
                            </a-tag>
                            <a-tag color="default" class="portion-tag">
                              Total: {{ item.total_portions }}
                            </a-tag>
                          </div>
                        </template>
                      </a-list-item-meta>
                    </a-list-item>
                  </template>
                </a-list>
              </div>

              <template #actions>
                <a-button
                  v-if="school.status === 'pending'"
                  type="primary"
                  block
                  @click="startPacking(school)"
                  :loading="updatingSchoolId === school.school_id"
                >
                  <template #icon><play-circle-outlined /></template>
                  Mulai Packing
                </a-button>
                <a-button
                  v-else-if="school.status === 'packing'"
                  type="primary"
                  block
                  @click="finishPacking(school)"
                  :loading="updatingSchoolId === school.school_id"
                  style="background-color: #52c41a; border-color: #52c41a"
                >
                  <template #icon><check-circle-outlined /></template>
                  Siap Kirim
                </a-button>
                <a-button
                  v-else
                  type="default"
                  block
                  disabled
                >
                  <template #icon><check-outlined /></template>
                  Sudah Siap
                </a-button>
              </template>
            </a-card>
          </a-col>
        </a-row>
      </a-spin>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { message, notification } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  PictureOutlined
} from '@ant-design/icons-vue'
import KDSDatePicker from '@/components/KDSDatePicker.vue'
import { getPackingToday, updatePackingStatus } from '@/services/kdsService'
import { database } from '@/services/firebase'
import { ref as dbRef, onValue, off } from 'firebase/database'

const schools = ref([])
const loading = ref(false)
const updatingSchoolId = ref(null)
const isConnected = ref(true)
const selectedDate = ref(new Date())
const error = ref(null)
let firebaseListener = null
let notificationListener = null

// Computed: Count of ready schools
const readyCount = computed(() => {
  return schools.value.filter(s => s.status === 'ready').length
})

// Computed: Check if all schools are ready
const allSchoolsReady = computed(() => {
  return schools.value.length > 0 && schools.value.every(s => s.status === 'ready')
})

// Compute empty message based on selected date
const emptyMessage = computed(() => {
  const today = new Date()
  const isToday = selectedDate.value.toDateString() === today.toDateString()
  return isToday ? 'Tidak ada alokasi packing untuk hari ini' : 'Tidak ada alokasi packing untuk tanggal ini'
})

// Get status color
const getStatusColor = (status) => {
  const colors = {
    pending: 'default',
    packing: 'processing',
    ready: 'success'
  }
  return colors[status] || 'default'
}

// Get status text in Indonesian
const getStatusText = (status) => {
  const texts = {
    pending: 'Belum Dimulai',
    packing: 'Sedang Packing',
    ready: 'Siap Kirim'
  }
  return texts[status] || status
}

// Load data from API
const loadData = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await getPackingToday(selectedDate.value)
    if (response.success) {
      schools.value = response.data || []
    } else {
      error.value = response.message || 'Gagal memuat data'
    }
  } catch (err) {
    console.error('Error loading packing data:', err)
    error.value = err.response?.data?.message || 'Gagal memuat data alokasi packing. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
}

// Retry loading data
const retryLoad = () => {
  loadData()
}

// Refresh data
const refreshData = () => {
  loadData()
}

// Start packing for a school
const startPacking = async (school) => {
  updatingSchoolId.value = school.school_id
  try {
    const response = await updatePackingStatus(school.school_id, 'packing')
    if (response.success) {
      message.success('Status berhasil diperbarui: Mulai Packing')
      // Reload data from API to get updated status
      await loadData()
    } else {
      message.error(response.message || 'Gagal memperbarui status')
    }
  } catch (error) {
    console.error('Error updating status:', error)
    message.error(error.response?.data?.message || 'Gagal memperbarui status')
  } finally {
    updatingSchoolId.value = null
  }
}

// Finish packing for a school
const finishPacking = async (school) => {
  updatingSchoolId.value = school.school_id
  try {
    const response = await updatePackingStatus(school.school_id, 'ready')
    if (response.success) {
      message.success(`${school.school_name} siap untuk pengiriman!`)
      // Reload data from API to get updated status
      await loadData()
    } else {
      message.error(response.message || 'Gagal memperbarui status')
    }
  } catch (error) {
    console.error('Error updating status:', error)
    message.error(error.response?.data?.message || 'Gagal memperbarui status')
  } finally {
    updatingSchoolId.value = null
  }
}

// Setup Firebase real-time listener for packing data
const setupFirebaseListener = () => {
  // Clean up existing listener first
  cleanupFirebaseListener()
  
  const dateStr = selectedDate.value.toISOString().split('T')[0]
  const packingRef = dbRef(database, `/kds/packing/${dateStr}`)
  
  firebaseListener = onValue(
    packingRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      
      if (data) {
        // Update schools with Firebase data
        const firebaseSchools = Object.values(data)
        
        // Merge with existing schools to preserve menu items and update portion size data
        schools.value = schools.value.map(school => {
          const firebaseSchool = firebaseSchools.find(fs => fs.school_id === school.school_id)
          if (firebaseSchool) {
            return {
              ...school,
              status: firebaseSchool.status,
              // Update portion size data if present in Firebase
              portion_size_type: firebaseSchool.portion_size_type || school.portion_size_type,
              portions_small: firebaseSchool.portions_small !== undefined ? firebaseSchool.portions_small : school.portions_small,
              portions_large: firebaseSchool.portions_large !== undefined ? firebaseSchool.portions_large : school.portions_large,
              total_portions: firebaseSchool.total_portions || school.total_portions
            }
          }
          return school
        })
      }
    },
    (error) => {
      console.error('Firebase listener error:', error)
      isConnected.value = false
    }
  )
}

// Setup Firebase listener for notifications
const setupNotificationListener = () => {
  const notificationRef = dbRef(database, '/notifications/logistics/packing_complete')
  
  notificationListener = onValue(
    notificationRef,
    (snapshot) => {
      const data = snapshot.val()
      
      if (data) {
        // Get the latest notification
        const notifications = Object.values(data)
        const latest = notifications[notifications.length - 1]
        
        if (latest && latest.message) {
          notification.success({
            message: 'Notifikasi',
            description: latest.message,
            duration: 10,
            placement: 'topRight'
          })
        }
      }
    },
    (error) => {
      console.error('Notification listener error:', error)
    }
  )
}

// Cleanup Firebase listener
const cleanupFirebaseListener = () => {
  if (firebaseListener) {
    const dateStr = selectedDate.value.toISOString().split('T')[0]
    const packingRef = dbRef(database, `/kds/packing/${dateStr}`)
    off(packingRef)
    firebaseListener = null
  }
}

// Cleanup notification listener
const cleanupNotificationListener = () => {
  if (notificationListener) {
    const notificationRef = dbRef(database, '/notifications/logistics/packing_complete')
    off(notificationRef)
    notificationListener = null
  }
}

// Handle date change from date picker
const handleDateChange = (date) => {
  selectedDate.value = date
  loadData()
  setupFirebaseListener()
}

// Watch for date changes
watch(selectedDate, () => {
  // This ensures Firebase listener is updated if date changes from other sources
  setupFirebaseListener()
})

onMounted(() => {
  loadData()
  setupFirebaseListener()
  setupNotificationListener()
})

onUnmounted(() => {
  cleanupFirebaseListener()
  cleanupNotificationListener()
})
</script>

<style scoped>
.kds-packing-view {
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

.school-card {
  height: 100%;
  transition: all 0.3s ease;
}

.school-card.status-pending {
  border-left: 4px solid #d9d9d9;
}

.school-card.status-packing {
  border-left: 4px solid #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.2);
}

.school-card.status-ready {
  border-left: 4px solid #52c41a;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.2);
}

.card-header {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.school-name {
  font-size: 18px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  line-height: 1.4;
}

.status-tag {
  align-self: flex-start;
}

.school-info {
  margin-top: 16px;
}

:deep(.ant-statistic-title) {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
}

:deep(.ant-list-item) {
  padding: 8px 0;
}

:deep(.ant-list-item-meta-title) {
  margin-bottom: 4px;
  font-weight: 500;
}

:deep(.ant-list-item-meta-description) {
  margin-top: 4px;
}

.portion-breakdown {
  margin: 16px 0;
}

.portion-breakdown-title {
  font-size: 14px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.75);
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.portion-size-card {
  padding: 16px 12px;
  border-radius: 8px;
  text-align: center;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.portion-size-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.portion-size-card.small {
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  border: 2px solid #40a9ff;
}

.portion-size-card.large {
  background: linear-gradient(135deg, #f0f5ff 0%, #d6e4ff 100%);
  border: 2px solid #597ef7;
}

.portion-size-card.single {
  background: linear-gradient(135deg, #e6f7ff 0%, #bae7ff 100%);
  border: 2px solid #40a9ff;
}

.portion-icon {
  display: inline-block;
  width: 32px;
  height: 32px;
  line-height: 32px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.9);
  color: #1890ff;
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.portion-size-card.large .portion-icon {
  color: #597ef7;
}

.portion-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.75);
  margin-bottom: 6px;
  font-weight: 600;
}

.portion-value {
  font-size: 24px;
  font-weight: 700;
  color: #1890ff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  line-height: 1.2;
}

.portion-size-card.large .portion-value {
  color: #597ef7;
}

.menu-item-portions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 4px;
}

.portion-tag {
  margin: 0;
  font-size: 12px;
}
</style>
