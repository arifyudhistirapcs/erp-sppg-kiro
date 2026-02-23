<template>
  <div class="dashboard-kepala-sppg">
    <a-page-header
      title="Dashboard Kepala SPPG"
      sub-title="Monitoring operasional harian secara real-time"
    >
      <template #extra>
        <a-space>
          <a-tag :color="isConnected ? 'green' : 'red'">
            <template #icon>
              <wifi-outlined v-if="isConnected" />
              <disconnect-outlined v-else />
            </template>
            {{ isConnected ? 'Terhubung' : 'Terputus' }}
          </a-tag>
          <a-button @click="refreshData" :loading="loading">
            <template #icon><reload-outlined /></template>
            Refresh
          </a-button>
          <a-button @click="exportDashboard" :loading="exporting">
            <template #icon><download-outlined /></template>
            Export
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <div class="dashboard-content">
      <a-spin :spinning="loading" tip="Memuat data dashboard...">
        <!-- KPI Cards -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('production')"
              hoverable
            >
              <a-statistic
                title="Porsi Disiapkan Hari Ini"
                :value="dashboard?.today_kpis?.portions_prepared || 0"
                :value-style="{ color: '#1890ff' }"
                suffix="porsi"
              />
              <div class="kpi-subtitle">
                Target: {{ productionTarget }} porsi
              </div>
            </a-card>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('delivery')"
              hoverable
            >
              <a-statistic
                title="Tingkat Pengiriman"
                :value="dashboard?.today_kpis?.delivery_rate || 0"
                :precision="1"
                :value-style="{ color: getDeliveryRateColor(dashboard?.today_kpis?.delivery_rate) }"
                suffix="%"
              />
              <div class="kpi-subtitle">
                {{ dashboard?.delivery_status?.deliveries_completed || 0 }} dari {{ dashboard?.delivery_status?.total_deliveries || 0 }} sekolah
              </div>
            </a-card>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('inventory')"
              hoverable
            >
              <a-statistic
                title="Ketersediaan Stok"
                :value="dashboard?.today_kpis?.stock_availability || 0"
                :precision="1"
                :value-style="{ color: getStockAvailabilityColor(dashboard?.today_kpis?.stock_availability) }"
                suffix="%"
              />
              <div class="kpi-subtitle">
                {{ criticalStockCount }} item stok kritis
              </div>
            </a-card>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('on-time')"
              hoverable
            >
              <a-statistic
                title="Ketepatan Waktu"
                :value="dashboard?.today_kpis?.on_time_delivery_rate || 0"
                :precision="1"
                :value-style="{ color: '#52c41a' }"
                suffix="%"
              />
              <div class="kpi-subtitle">
                Pengiriman tepat waktu
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- Production Status -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :xs="24" :lg="12">
            <a-card title="Status Produksi" class="status-card">
              <template #extra>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="drillDown('production')"
                >
                  Lihat Detail
                </a-button>
              </template>
              
              <div class="production-overview">
                <a-row :gutter="16">
                  <a-col :span="8">
                    <div class="status-item">
                      <div class="status-number">{{ dashboard?.production_status?.total_recipes || 0 }}</div>
                      <div class="status-label">Total Menu</div>
                    </div>
                  </a-col>
                  <a-col :span="8">
                    <div class="status-item">
                      <div class="status-number cooking">{{ dashboard?.production_status?.recipes_cooking || 0 }}</div>
                      <div class="status-label">Sedang Dimasak</div>
                    </div>
                  </a-col>
                  <a-col :span="8">
                    <div class="status-item">
                      <div class="status-number ready">{{ dashboard?.production_status?.recipes_ready || 0 }}</div>
                      <div class="status-label">Siap Packing</div>
                    </div>
                  </a-col>
                </a-row>
              </div>

              <a-divider />

              <div class="progress-section">
                <div class="progress-label">
                  Progress Memasak: {{ (dashboard?.production_status?.completion_rate || 0).toFixed(1) }}%
                </div>
                <a-progress 
                  :percent="dashboard?.production_status?.completion_rate || 0" 
                  :stroke-color="getProgressColor(dashboard?.production_status?.completion_rate)"
                />
              </div>

              <div class="packing-status">
                <a-row :gutter="16">
                  <a-col :span="8">
                    <a-tag color="default">Pending: {{ dashboard?.production_status?.packing_pending || 0 }}</a-tag>
                  </a-col>
                  <a-col :span="8">
                    <a-tag color="processing">Packing: {{ dashboard?.production_status?.packing_in_progress || 0 }}</a-tag>
                  </a-col>
                  <a-col :span="8">
                    <a-tag color="success">Siap Kirim: {{ dashboard?.production_status?.packing_ready || 0 }}</a-tag>
                  </a-col>
                </a-row>
              </div>
            </a-card>
          </a-col>

          <a-col :xs="24" :lg="12">
            <a-card title="Status Pengiriman" class="status-card">
              <template #extra>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="drillDown('delivery')"
                >
                  Lihat Detail
                </a-button>
              </template>

              <div class="delivery-overview">
                <a-row :gutter="16">
                  <a-col :span="8">
                    <div class="status-item">
                      <div class="status-number">{{ dashboard?.delivery_status?.total_deliveries || 0 }}</div>
                      <div class="status-label">Total Sekolah</div>
                    </div>
                  </a-col>
                  <a-col :span="8">
                    <div class="status-item">
                      <div class="status-number in-progress">{{ dashboard?.delivery_status?.deliveries_in_progress || 0 }}</div>
                      <div class="status-label">Dalam Perjalanan</div>
                    </div>
                  </a-col>
                  <a-col :span="8">
                    <div class="status-item">
                      <div class="status-number completed">{{ dashboard?.delivery_status?.deliveries_completed || 0 }}</div>
                      <div class="status-label">Selesai</div>
                    </div>
                  </a-col>
                </a-row>
              </div>

              <a-divider />

              <div class="progress-section">
                <div class="progress-label">
                  Progress Pengiriman: {{ (dashboard?.delivery_status?.completion_rate || 0).toFixed(1) }}%
                </div>
                <a-progress 
                  :percent="dashboard?.delivery_status?.completion_rate || 0" 
                  :stroke-color="getProgressColor(dashboard?.delivery_status?.completion_rate)"
                />
              </div>

              <div class="delivery-real-time" v-if="realtimeDeliveryUpdates.length > 0">
                <a-divider>Update Terbaru</a-divider>
                <a-timeline size="small">
                  <a-timeline-item 
                    v-for="update in realtimeDeliveryUpdates.slice(0, 3)" 
                    :key="update.id"
                    :color="getUpdateColor(update.status)"
                  >
                    <div class="update-content">
                      <div class="update-school">{{ update.school_name }}</div>
                      <div class="update-status">{{ getUpdateStatusText(update.status) }}</div>
                      <div class="update-time">{{ formatTime(update.timestamp) }}</div>
                    </div>
                  </a-timeline-item>
                </a-timeline>
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- Critical Stock Items -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :span="24">
            <a-card title="Stok Kritis" class="critical-stock-card">
              <template #extra>
                <a-space>
                  <a-tag :color="criticalStockCount > 0 ? 'red' : 'green'">
                    {{ criticalStockCount }} Item Kritis
                  </a-tag>
                  <a-button 
                    type="link" 
                    size="small" 
                    @click="drillDown('inventory')"
                  >
                    Lihat Semua
                  </a-button>
                </a-space>
              </template>

              <div v-if="dashboard?.critical_stock && dashboard.critical_stock.length > 0">
                <a-row :gutter="[16, 16]">
                  <a-col 
                    v-for="item in dashboard.critical_stock.slice(0, 6)" 
                    :key="item.ingredient_id"
                    :xs="24" :sm="12" :md="8" :lg="6"
                  >
                    <a-card 
                      size="small" 
                      class="critical-item-card"
                      :class="{ 'very-critical': item.days_remaining <= 1 }"
                    >
                      <div class="critical-item">
                        <div class="item-name">{{ item.ingredient_name }}</div>
                        <div class="item-stock">
                          <span class="current-stock">{{ item.current_stock }}</span>
                          <span class="unit">{{ item.unit }}</span>
                        </div>
                        <div class="item-threshold">
                          Min: {{ item.min_threshold }} {{ item.unit }}
                        </div>
                        <div class="item-days" :class="getDaysRemainingClass(item.days_remaining)">
                          {{ item.days_remaining.toFixed(1) }} hari tersisa
                        </div>
                      </div>
                    </a-card>
                  </a-col>
                </a-row>
              </div>
              <a-empty 
                v-else 
                description="Tidak ada stok kritis" 
                :image="false"
                style="margin: 20px 0;"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Last Updated Info -->
        <a-row>
          <a-col :span="24">
            <a-card size="small" class="update-info">
              <a-space>
                <span>Terakhir diperbarui: {{ formatDateTime(dashboard?.updated_at) }}</span>
                <a-divider type="vertical" />
                <span>Auto-refresh: {{ autoRefreshEnabled ? 'Aktif' : 'Nonaktif' }}</span>
                <a-switch 
                  v-model:checked="autoRefreshEnabled" 
                  size="small"
                  @change="toggleAutoRefresh"
                />
              </a-space>
            </a-card>
          </a-col>
        </a-row>
      </a-spin>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  DownloadOutlined
} from '@ant-design/icons-vue'
import { getKepalaSSPGDashboard, exportDashboardData } from '@/services/dashboardService'
import { database } from '@/services/firebase'
import { ref as dbRef, onValue, off } from 'firebase/database'
import dayjs from 'dayjs'

const router = useRouter()

// Reactive data
const dashboard = ref(null)
const loading = ref(false)
const exporting = ref(false)
const isConnected = ref(true)
const autoRefreshEnabled = ref(true)
const realtimeDeliveryUpdates = ref([])

// Firebase listeners
let dashboardListener = null
let deliveryListener = null
let autoRefreshInterval = null

// Computed properties
const criticalStockCount = computed(() => {
  return dashboard.value?.critical_stock?.length || 0
})

const productionTarget = computed(() => {
  // Calculate target based on total recipes * average portions
  const totalRecipes = dashboard.value?.production_status?.total_recipes || 0
  return totalRecipes * 50 // Assuming 50 portions per recipe on average
})

// Load dashboard data from API
const loadDashboardData = async () => {
  loading.value = true
  try {
    const response = await getKepalaSSPGDashboard()
    if (response.success) {
      dashboard.value = response.dashboard
    } else {
      message.error(response.message || 'Gagal memuat data dashboard')
    }
  } catch (error) {
    console.error('Error loading dashboard:', error)
    message.error('Gagal memuat data dashboard')
  } finally {
    loading.value = false
  }
}

// Refresh data
const refreshData = () => {
  loadDashboardData()
}

// Export dashboard
const exportDashboard = async () => {
  exporting.value = true
  try {
    const response = await exportDashboardData('kepala_sppg', 'json')
    if (response.success) {
      // Create download link
      const dataStr = JSON.stringify(response.data, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `dashboard-kepala-sppg-${dayjs().format('YYYY-MM-DD-HH-mm')}.json`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      
      message.success('Dashboard berhasil diexport')
    } else {
      message.error(response.message || 'Gagal mengexport dashboard')
    }
  } catch (error) {
    console.error('Error exporting dashboard:', error)
    message.error('Gagal mengexport dashboard')
  } finally {
    exporting.value = false
  }
}

// Drill down to detail pages
const drillDown = (type) => {
  switch (type) {
    case 'production':
      router.push('/kds/cooking')
      break
    case 'delivery':
      router.push('/delivery-tasks')
      break
    case 'inventory':
      router.push('/inventory')
      break
    case 'on-time':
      router.push('/delivery-tasks')
      break
    default:
      console.log('Unknown drill down type:', type)
  }
}

// Setup Firebase real-time listeners
const setupFirebaseListeners = () => {
  // Dashboard data listener
  const dashboardRef = dbRef(database, '/dashboard/kepala_sppg')
  dashboardListener = onValue(
    dashboardRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      if (data) {
        dashboard.value = data
      }
    },
    (error) => {
      console.error('Firebase dashboard listener error:', error)
      isConnected.value = false
    }
  )

  // Delivery updates listener for real-time notifications
  const today = dayjs().format('YYYY-MM-DD')
  const deliveryRef = dbRef(database, `/delivery_updates/${today}`)
  deliveryListener = onValue(
    deliveryRef,
    (snapshot) => {
      const data = snapshot.val()
      if (data) {
        const updates = Object.values(data)
          .sort((a, b) => b.timestamp - a.timestamp)
        realtimeDeliveryUpdates.value = updates
      }
    },
    (error) => {
      console.error('Firebase delivery listener error:', error)
    }
  )
}

// Cleanup Firebase listeners
const cleanupFirebaseListeners = () => {
  if (dashboardListener) {
    const dashboardRef = dbRef(database, '/dashboard/kepala_sppg')
    off(dashboardRef)
    dashboardListener = null
  }
  
  if (deliveryListener) {
    const today = dayjs().format('YYYY-MM-DD')
    const deliveryRef = dbRef(database, `/delivery_updates/${today}`)
    off(deliveryRef)
    deliveryListener = null
  }
}

// Auto refresh functionality
const toggleAutoRefresh = (enabled) => {
  if (enabled) {
    autoRefreshInterval = setInterval(() => {
      loadDashboardData()
    }, 5 * 60 * 1000) // Refresh every 5 minutes
  } else {
    if (autoRefreshInterval) {
      clearInterval(autoRefreshInterval)
      autoRefreshInterval = null
    }
  }
}

// Helper functions for styling
const getDeliveryRateColor = (rate) => {
  if (rate >= 90) return '#52c41a'
  if (rate >= 70) return '#faad14'
  return '#ff4d4f'
}

const getStockAvailabilityColor = (availability) => {
  if (availability >= 80) return '#52c41a'
  if (availability >= 60) return '#faad14'
  return '#ff4d4f'
}

const getProgressColor = (percent) => {
  if (percent >= 80) return '#52c41a'
  if (percent >= 50) return '#faad14'
  return '#ff4d4f'
}

const getDaysRemainingClass = (days) => {
  if (days <= 1) return 'critical'
  if (days <= 3) return 'warning'
  return 'normal'
}

const getUpdateColor = (status) => {
  const colors = {
    completed: 'green',
    in_progress: 'blue',
    pending: 'gray'
  }
  return colors[status] || 'gray'
}

const getUpdateStatusText = (status) => {
  const texts = {
    completed: 'Pengiriman selesai',
    in_progress: 'Dalam perjalanan',
    pending: 'Menunggu pengiriman'
  }
  return texts[status] || status
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return dayjs(timestamp * 1000).format('HH:mm')
}

const formatDateTime = (datetime) => {
  if (!datetime) return '-'
  return dayjs(datetime).format('DD/MM/YYYY HH:mm')
}

// Lifecycle hooks
onMounted(() => {
  loadDashboardData()
  setupFirebaseListeners()
  toggleAutoRefresh(autoRefreshEnabled.value)
})

onUnmounted(() => {
  cleanupFirebaseListeners()
  if (autoRefreshInterval) {
    clearInterval(autoRefreshInterval)
  }
})
</script>

<style scoped>
.dashboard-kepala-sppg {
  padding: 24px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.dashboard-content {
  margin-top: 16px;
}

.kpi-card {
  text-align: center;
  transition: all 0.3s ease;
}

.kpi-card.clickable {
  cursor: pointer;
}

.kpi-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.kpi-subtitle {
  margin-top: 8px;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.status-card {
  height: 100%;
}

.production-overview,
.delivery-overview {
  margin-bottom: 16px;
}

.status-item {
  text-align: center;
}

.status-number {
  font-size: 24px;
  font-weight: bold;
  color: #1890ff;
}

.status-number.cooking {
  color: #faad14;
}

.status-number.ready {
  color: #52c41a;
}

.status-number.in-progress {
  color: #1890ff;
}

.status-number.completed {
  color: #52c41a;
}

.status-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
  margin-top: 4px;
}

.progress-section {
  margin-bottom: 16px;
}

.progress-label {
  margin-bottom: 8px;
  font-weight: 500;
}

.packing-status {
  margin-top: 16px;
}

.critical-stock-card {
  border-left: 4px solid #ff4d4f;
}

.critical-item-card {
  height: 100%;
  transition: all 0.3s ease;
}

.critical-item-card.very-critical {
  border-color: #ff4d4f;
  background-color: #fff2f0;
}

.critical-item-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.critical-item {
  text-align: center;
}

.item-name {
  font-weight: 500;
  margin-bottom: 8px;
  color: #262626;
}

.item-stock {
  margin-bottom: 4px;
}

.current-stock {
  font-size: 18px;
  font-weight: bold;
  color: #ff4d4f;
}

.unit {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
  margin-left: 4px;
}

.item-threshold {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
  margin-bottom: 4px;
}

.item-days {
  font-size: 12px;
  font-weight: 500;
}

.item-days.critical {
  color: #ff4d4f;
}

.item-days.warning {
  color: #faad14;
}

.item-days.normal {
  color: #52c41a;
}

.delivery-real-time {
  margin-top: 16px;
}

.update-content {
  font-size: 12px;
}

.update-school {
  font-weight: 500;
  color: #262626;
}

.update-status {
  color: rgba(0, 0, 0, 0.65);
  margin: 2px 0;
}

.update-time {
  color: rgba(0, 0, 0, 0.45);
}

.update-info {
  text-align: center;
  background-color: #fafafa;
}

:deep(.ant-statistic-title) {
  font-size: 14px;
  margin-bottom: 8px;
}

:deep(.ant-statistic-content) {
  font-size: 20px;
}

:deep(.ant-progress-text) {
  font-size: 12px;
}

:deep(.ant-timeline-item-content) {
  margin-left: 20px;
}
</style>