<template>
  <div class="kds-packing-view">
    <a-page-header
      title="Packing - Pengemasan"
      sub-title="Alokasi porsi per sekolah"
    >
      <template #extra>
        <a-space>
          <KDSDatePicker
            v-model="selectedDate"
            :loading="loading"
            @change="handleDateChange"
          />
          <a-tag :color="isConnected ? 'green' : 'red'">
            <template #icon>
              <wifi-outlined v-if="isConnected" />
              <disconnect-outlined v-else />
            </template>
            {{ isConnected ? 'Terhubung' : 'Terputus' }}
          </a-tag>
          <a-badge :count="readyCount" :number-style="{ backgroundColor: '#52c41a' }">
            <a-button @click="refreshData" :loading="loading">
              <template #icon><reload-outlined /></template>
              Refresh
            </a-button>
          </a-badge>
        </a-space>
      </template>
    </a-page-header>

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
              :title="school.school_name"
            >
              <template #extra>
                <a-tag :color="getStatusColor(school.status)">
                  {{ getStatusText(school.status) }}
                </a-tag>
              </template>

              <div class="school-info">
                <a-statistic
                  title="Total Porsi"
                  :value="school.portions"
                  suffix="porsi"
                  :value-style="{ color: '#1890ff', fontSize: '24px', fontWeight: 'bold' }"
                />

                <a-divider>Menu Items</a-divider>
                <a-list
                  size="small"
                  :data-source="school.menu_items"
                  :split="false"
                >
                  <template #renderItem="{ item }">
                    <a-list-item>
                      <a-list-item-meta>
                        <template #title>
                          {{ item.recipe_name }}
                        </template>
                        <template #description>
                          <a-tag color="blue">{{ item.portions }} porsi</a-tag>
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
  CheckOutlined
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
      // Update will come from Firebase listener
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
      // Update will come from Firebase listener
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
        
        // Merge with existing schools to preserve menu items
        schools.value = schools.value.map(school => {
          const firebaseSchool = firebaseSchools.find(fs => fs.school_id === school.school_id)
          if (firebaseSchool) {
            return {
              ...school,
              status: firebaseSchool.status
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
  padding: 24px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.content-wrapper {
  margin-top: 16px;
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
</style>
