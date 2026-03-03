<template>
  <div class="dashboard-sspg">
    <!-- Stats Cards Row -->
    <div class="stats-row">
      <HStatCard
        :icon="AppstoreOutlined"
        icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
        label="Porsi Disiapkan"
        :value="String(kpis.portions_prepared)"
        change="porsi hari ini"
        change-type="increase"
        :loading="loading"
      />
      <HStatCard
        :icon="CarOutlined"
        icon-bg="linear-gradient(135deg, #74788C 0%, #5A4372 100%)"
        label="Delivery Rate"
        :value="`${kpis.delivery_rate}%`"
        :change="`${delivery.total_deliveries} pengiriman hari ini`"
        change-type="increase"
        :loading="loading"
      />
      <HStatCard
        :icon="InboxOutlined"
        icon-bg="linear-gradient(135deg, #3D2B53 0%, #5A4372 100%)"
        label="Ketersediaan Stok"
        :value="`${kpis.stock_availability}%`"
        :change="`${criticalStockItems.length} item kritis`"
        :change-type="criticalStockItems.length > 0 ? 'decrease' : 'increase'"
        :loading="loading"
      />
      <HStatCard
        :icon="CheckCircleOutlined"
        icon-bg="linear-gradient(135deg, #05CD99 0%, #52c41a 100%)"
        label="On-Time Delivery"
        :value="`${kpis.on_time_delivery_rate}%`"
        :change="`${delivery.total_deliveries} pengiriman tepat waktu`"
        change-type="increase"
        :loading="loading"
      />
    </div>

    <!-- Charts Row -->
    <div class="charts-row">
      <HChartCard
        title="Status Produksi"
        subtitle="Resep & Packing hari ini"
        :height="320"
        :loading="loading"
        class="chart-card"
      >
        <div ref="productionChartRef" class="chart-container"></div>
      </HChartCard>

      <HChartCard
        title="Status Pengiriman & Pengambilan"
        subtitle="Progress delivery hari ini"
        :height="320"
        :loading="loading"
        class="chart-card"
      >
        <div ref="deliveryChartRef" class="chart-container"></div>
      </HChartCard>

      <HChartCard
        title="Status Pencucian"
        subtitle="Progress kebersihan hari ini"
        :height="320"
        :loading="loading"
        class="chart-card"
      >
        <div ref="cleaningChartRef" class="chart-container"></div>
      </HChartCard>
    </div>

    <!-- Production, Delivery & Cleaning Detail Tables -->
    <div class="tables-row">
      <div class="table-section">
        <h3 class="section-title">Detail Produksi</h3>
        <HDataTable
          :columns="productionColumns"
          :data-source="productionTableData"
          :loading="loading"
          :pagination="false"
          :mobile-card-view="true"
          class="data-table"
        >
          <template #cell-status="{ text }">
            <span class="status-badge" :class="`status-badge--${getStatusType(text)}`">
              <span class="status-dot"></span>
              <span>{{ text }}</span>
            </span>
          </template>
        </HDataTable>
      </div>

      <div class="table-section">
        <h3 class="section-title">Detail Pengiriman & Pengambilan</h3>
        <HDataTable
          :columns="deliveryColumns"
          :data-source="deliveryTableData"
          :loading="loading"
          :pagination="false"
          :mobile-card-view="true"
          class="data-table"
        >
          <template #cell-status="{ text }">
            <span class="status-badge" :class="`status-badge--${getStatusType(text)}`">
              <span class="status-dot"></span>
              <span>{{ text }}</span>
            </span>
          </template>
        </HDataTable>
      </div>

      <div class="table-section">
        <h3 class="section-title">Detail Pencucian</h3>
        <HDataTable
          :columns="cleaningColumns"
          :data-source="cleaningTableData"
          :loading="loading"
          :pagination="false"
          :mobile-card-view="true"
          class="data-table"
        >
          <template #cell-status="{ text }">
            <span class="status-badge" :class="`status-badge--${getStatusType(text)}`">
              <span class="status-dot"></span>
              <span>{{ text }}</span>
            </span>
          </template>
        </HDataTable>
      </div>
    </div>

    <!-- Critical Stock Section -->
    <div v-if="criticalStockItems.length > 0" class="critical-stock-section">
      <div class="section-header">
        <h3 class="section-title">
          <WarningOutlined style="color: #EE5D50; margin-right: 8px;" />
          Stok Kritis ({{ criticalStockItems.length }} item)
        </h3>
        <a-button 
          v-if="criticalStockItems.length > 6"
          type="link" 
          @click="goToInventory"
          style="color: #5A4372;"
        >
          Lihat Lainnya
          <RightOutlined />
        </a-button>
      </div>
      <div class="critical-stock-grid">
        <div
          v-for="item in displayedCriticalStock"
          :key="item.ingredient_id"
          class="critical-stock-card h-card"
        >
          <div class="critical-stock-header">
            <span class="critical-stock-name">{{ item.ingredient_name }}</span>
            <span
              class="critical-stock-badge"
              :class="item.current_stock <= item.min_threshold * 0.5 ? 'critical-stock-badge--danger' : 'critical-stock-badge--warning'"
            >
              {{ item.current_stock <= item.min_threshold * 0.5 ? 'Kritis' : 'Rendah' }}
            </span>
          </div>
          <div class="critical-stock-info">
            <div class="critical-stock-value">
              <span class="critical-stock-current">{{ item.current_stock }}</span>
              <span class="critical-stock-unit">{{ item.unit }}</span>
            </div>
            <div class="critical-stock-minimum">
              Min: {{ item.min_threshold }} {{ item.unit }}
            </div>
          </div>
          <a-progress
            :percent="Math.min(100, Math.round((item.current_stock / item.min_threshold) * 100))"
            :stroke-color="item.current_stock <= item.min_threshold * 0.5 ? '#EE5D50' : '#FFB547'"
            :show-info="false"
            size="small"
          />
          <div v-if="item.days_remaining" class="critical-stock-days">
            ~{{ item.days_remaining.toFixed(1) }} hari tersisa
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state if no critical stock -->
    <div v-else-if="!loading" class="no-critical-stock">
      <CheckCircleOutlined style="font-size: 24px; color: #05CD99; margin-right: 8px;" />
      <span>Semua stok dalam kondisi aman</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  AppstoreOutlined,
  CarOutlined,
  InboxOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  RightOutlined
} from '@ant-design/icons-vue'
import HStatCard from '@/components/horizon/HStatCard.vue'
import HChartCard from '@/components/horizon/HChartCard.vue'
import HDataTable from '@/components/horizon/HDataTable.vue'
import { useHorizonChart } from '@/composables/useHorizonChart'
import { getKepalaSSPGDashboard } from '@/services/dashboardService'
import { database } from '@/services/firebase'
import { ref as dbRef, onValue, off } from 'firebase/database'

const router = useRouter()
const dashboard = ref(null)
const loading = ref(true)

const productionChartRef = ref(null)
const deliveryChartRef = ref(null)
const cleaningChartRef = ref(null)

let dashboardListener = null

// Computed helpers for safe access
const kpis = computed(() => dashboard.value?.today_kpis || {
  portions_prepared: 0, delivery_rate: 0, stock_availability: 0, on_time_delivery_rate: 0
})

const production = computed(() => dashboard.value?.production_status || {
  total_recipes: 0, recipes_pending: 0, recipes_cooking: 0, recipes_ready: 0,
  packing_pending: 0, packing_in_progress: 0, packing_ready: 0, completion_rate: 0
})

const delivery = computed(() => dashboard.value?.delivery_status || {
  total_deliveries: 0, status_breakdown: [], completion_rate: 0
})

const cleaning = computed(() => dashboard.value?.cleaning_status || {
  total_items: 0, items_pending: 0, items_in_progress: 0, items_completed: 0, completion_rate: 0
})

const criticalStockItems = computed(() => dashboard.value?.critical_stock || [])

// Display only first 6 critical stock items
const displayedCriticalStock = computed(() => criticalStockItems.value.slice(0, 6))

// Navigate to inventory with low stock filter
const goToInventory = () => {
  router.push('/inventory')
}

// Production table data from real API
const productionColumns = [
  { title: 'Sekolah', dataIndex: 'school', key: 'school' },
  { title: 'Porsi', dataIndex: 'portions', key: 'portions' },
  { title: 'Status', dataIndex: 'status', key: 'status', type: 'status' }
]

const productionTableData = computed(() => {
  const details = dashboard.value?.production_details || []
  return details.map((detail, index) => ({
    key: String(index + 1),
    school: detail.school_name,
    portions: detail.portions,
    status: detail.status
  }))
})

// Delivery table data from real API
const deliveryColumns = [
  { title: 'Sekolah', dataIndex: 'school', key: 'school' },
  { title: 'Porsi', dataIndex: 'portions', key: 'portions' },
  { title: 'Status', dataIndex: 'status', key: 'status', type: 'status' }
]

const deliveryTableData = computed(() => {
  const details = dashboard.value?.delivery_details || []
  return details.map((detail, index) => ({
    key: String(index + 1),
    school: detail.school_name,
    portions: detail.portions,
    status: detail.status
  }))
})

// Cleaning table data from real API
const cleaningColumns = [
  { title: 'Sekolah', dataIndex: 'school', key: 'school' },
  { title: 'Porsi', dataIndex: 'portions', key: 'portions' },
  { title: 'Status', dataIndex: 'status', key: 'status', type: 'status' }
]

const cleaningTableData = computed(() => {
  const details = dashboard.value?.cleaning_details || []
  return details.map((detail, index) => ({
    key: String(index + 1),
    school: detail.school_name,
    portions: detail.portions, // Backend already returns portions field
    status: detail.status
  }))
})

// Chart composables
const { setOption: setProductionOption } = useHorizonChart(productionChartRef, {})
const { setOption: setDeliveryOption } = useHorizonChart(deliveryChartRef, {})
const { setOption: setCleaningOption } = useHorizonChart(cleaningChartRef, {})

const updateCharts = () => {
  const p = production.value
  const d = delivery.value

  // Production donut chart
  const pieData = [
    { value: p.recipes_pending, name: 'Menunggu', itemStyle: { color: '#A3AED0' } },
    { value: p.recipes_cooking, name: 'Sedang Dimasak', itemStyle: { color: '#FFB547' } },
    { value: p.recipes_ready, name: 'Selesai Dimasak', itemStyle: { color: '#05CD99' } },
    { value: p.packing_pending, name: 'Siap Packing', itemStyle: { color: '#9F7AEA' } },
    { value: p.packing_in_progress, name: 'Sedang Packing', itemStyle: { color: '#5A4372' } },
    { value: p.packing_ready, name: 'Selesai Packing', itemStyle: { color: '#3D2B53' } },
  ].filter(item => item.value > 0)

  // If no data, show a "no data" placeholder
  const finalPieData = pieData.length > 0 ? pieData : [{ value: 1, name: 'Belum ada data', itemStyle: { color: '#E9EDF7' } }]

  setProductionOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, data: finalPieData.map(i => i.name) },
    series: [{
      name: 'Produksi',
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{c}' },
      data: finalPieData
    }]
  })

  // Delivery bar chart - use status_breakdown from API
  const deliveryBreakdown = d.status_breakdown || []
  const deliveryLabels = deliveryBreakdown.map(item => item.status_label)
  const deliveryData = deliveryBreakdown.map((item, index) => {
    // Assign colors based on status
    const colors = ['#A3AED0', '#FFB547', '#9F7AEA', '#05CD99', '#5A4372', '#3D2B53']
    return {
      value: item.count,
      itemStyle: { color: colors[index % colors.length], borderRadius: [0, 4, 4, 0] }
    }
  })

  // If no data, show placeholder
  const finalDeliveryLabels = deliveryLabels.length > 0 ? deliveryLabels : ['Belum ada data']
  const finalDeliveryData = deliveryData.length > 0 ? deliveryData : [{ value: 0, itemStyle: { color: '#E9EDF7', borderRadius: [0, 4, 4, 0] } }]

  setDeliveryOption({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    grid: { left: '3%', right: '8%', bottom: '3%', containLabel: true },
    xAxis: {
      type: 'value',
      axisLabel: { formatter: '{value}' }
    },
    yAxis: {
      type: 'category',
      data: finalDeliveryLabels
    },
    series: [{
      name: 'Pengiriman',
      type: 'bar',
      data: finalDeliveryData,
      barWidth: '40%',
      label: { show: true, position: 'right', formatter: '{c}' }
    }]
  })

  // Cleaning pie chart (same style as production)
  const c = cleaning.value
  const cleaningPieData = [
    { value: c.items_pending, name: 'Menunggu', itemStyle: { color: '#A3AED0' } },
    { value: c.items_in_progress, name: 'Sedang Dicuci', itemStyle: { color: '#FFB547' } },
    { value: c.items_completed, name: 'Selesai', itemStyle: { color: '#05CD99' } },
  ].filter(item => item.value > 0)

  // If no data, show a "no data" placeholder
  const finalCleaningData = cleaningPieData.length > 0 ? cleaningPieData : [{ value: 1, name: 'Belum ada data', itemStyle: { color: '#E9EDF7' } }]

  setCleaningOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    legend: { bottom: 0, data: finalCleaningData.map(i => i.name) },
    series: [{
      name: 'Pencucian',
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '45%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{c}' },
      data: finalCleaningData
    }]
  })
}

const loadDashboardData = async () => {
  loading.value = true
  try {
    const response = await getKepalaSSPGDashboard()
    if (response.success) {
      dashboard.value = response.dashboard
    }
  } catch (error) {
    console.error('Error loading dashboard:', error)
    message.error('Gagal memuat data dashboard')
  } finally {
    loading.value = false
    await nextTick()
    updateCharts()
  }
}

const setupFirebaseListeners = () => {
  try {
    const dashRef = dbRef(database, '/dashboard/kepala_sppg')
    dashboardListener = onValue(dashRef, (snapshot) => {
      const data = snapshot.val()
      if (data) {
        dashboard.value = data
        updateCharts()
      }
    }, (error) => {
      console.error('Firebase listener error:', error)
    })
  } catch (e) {
    console.warn('Firebase not available:', e)
  }
}

const cleanupFirebaseListeners = () => {
  if (dashboardListener) {
    try { off(dbRef(database, '/dashboard/kepala_sppg')) } catch (e) { /* ignore */ }
    dashboardListener = null
  }
}

const getStatusType = (status) => {
  const s = String(status).toLowerCase()
  if (s.includes('selesai') || s.includes('completed')) return 'success'
  if (s.includes('packing') || s.includes('memasak') || s.includes('%')) return 'warning'
  return 'default'
}

onMounted(() => {
  loadDashboardData()
  setupFirebaseListeners()
})

onUnmounted(() => {
  cleanupFirebaseListeners()
})
</script>

<style scoped>
.dashboard-sspg {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 24px;
}
@media (max-width: 1024px) { .stats-row { grid-template-columns: repeat(2, 1fr); gap: 16px; } }
@media (max-width: 768px) { .stats-row { grid-template-columns: 1fr; gap: 12px; } }

.charts-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}
@media (max-width: 1024px) { .charts-row { grid-template-columns: repeat(2, 1fr); gap: 16px; } }
@media (max-width: 768px) { .charts-row { grid-template-columns: 1fr; gap: 16px; } }

.chart-container {
  width: 100%;
  height: 100%;
  min-height: 280px;
}
@media (max-width: 768px) { .chart-container { min-height: 240px; } }

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--h-text-primary, #322837);
  margin: 0;
  display: flex;
  align-items: center;
}
.dark .section-title { color: var(--h-text-primary-dark, #F8FDEA); }

.tables-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}
@media (max-width: 1024px) { .tables-row { grid-template-columns: repeat(2, 1fr); gap: 16px; } }
@media (max-width: 768px) { .tables-row { grid-template-columns: 1fr; gap: 16px; } }

.table-section { display: flex; flex-direction: column; }

.critical-stock-section { margin-top: 4px; }

.critical-stock-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
@media (max-width: 1024px) { .critical-stock-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 768px) { .critical-stock-grid { grid-template-columns: 1fr; } }

.critical-stock-card { display: flex; flex-direction: column; gap: 10px; }

.critical-stock-header { display: flex; justify-content: space-between; align-items: center; }

.critical-stock-name { font-size: 14px; font-weight: 600; color: var(--h-text-primary, #322837); }
.dark .critical-stock-name { color: var(--h-text-primary-dark, #F8FDEA); }

.critical-stock-badge { font-size: 11px; font-weight: 600; padding: 2px 10px; border-radius: 12px; }
.critical-stock-badge--danger { background: rgba(238, 93, 80, 0.15); color: #EE5D50; }
.critical-stock-badge--warning { background: rgba(255, 181, 71, 0.15); color: #FFB547; }

.critical-stock-info { display: flex; justify-content: space-between; align-items: baseline; }
.critical-stock-value { display: flex; align-items: baseline; gap: 4px; }
.critical-stock-current { font-size: 24px; font-weight: 700; color: var(--h-text-primary, #322837); }
.dark .critical-stock-current { color: var(--h-text-primary-dark, #F8FDEA); }
.critical-stock-unit { font-size: 13px; color: var(--h-text-secondary, #74788C); }
.critical-stock-minimum { font-size: 12px; color: var(--h-text-secondary, #74788C); }
.critical-stock-days { font-size: 11px; color: var(--h-text-secondary, #74788C); font-style: italic; }

.no-critical-stock {
  display: flex;
  align-items: center;
  padding: 16px 20px;
  background: rgba(5, 205, 153, 0.08);
  border-radius: 12px;
  font-size: 14px;
  color: #05CD99;
  font-weight: 500;
}

.status-badge { display: inline-flex; align-items: center; gap: 6px; padding: 4px 12px; border-radius: 8px; font-size: 12px; font-weight: 500; }
.status-dot { width: 6px; height: 6px; border-radius: 50%; }
.status-badge--success { background: rgba(5, 205, 153, 0.1); color: #05CD99; }
.status-badge--success .status-dot { background: #05CD99; }
.status-badge--warning { background: rgba(255, 181, 71, 0.1); color: #FFB547; }
.status-badge--warning .status-dot { background: #FFB547; }
.status-badge--default { background: rgba(163, 174, 208, 0.1); color: #74788C; }
.status-badge--default .status-dot { background: #74788C; }
</style>
