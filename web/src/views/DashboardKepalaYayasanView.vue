<template>
  <div>
    <!-- Time Period Selection -->
    <div class="h-card" style="margin-bottom: 20px;">
      <a-row :gutter="16" align="middle">
        <a-col :xs="24" :sm="12" :md="6">
          <a-form-item label="Periode Waktu" style="margin-bottom: 0;">
            <a-select
              v-model:value="selectedPeriod"
              @change="handlePeriodChange"
              style="width: 100%"
            >
              <a-select-option value="this_month">Bulan Ini</a-select-option>
              <a-select-option value="this_quarter">Kuartal Ini</a-select-option>
              <a-select-option value="this_year">Tahun Ini</a-select-option>
              <a-select-option value="last_month">Bulan Lalu</a-select-option>
              <a-select-option value="last_quarter">Kuartal Lalu</a-select-option>
              <a-select-option value="custom">Kustom</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :xs="24" :sm="12" :md="8" v-if="selectedPeriod === 'custom'">
          <a-form-item label="Rentang Tanggal" style="margin-bottom: 0;">
            <a-range-picker
              v-model:value="customDateRange"
              @change="handleCustomDateChange"
              style="width: 100%"
              format="DD/MM/YYYY"
            />
          </a-form-item>
        </a-col>
        <a-col :xs="12" :sm="6" :md="4">
          <a-form-item style="margin-bottom: 0;">
            <a-switch 
              v-model:checked="autoRefreshEnabled" 
              @change="toggleAutoRefresh"
              checked-children="Auto"
              un-checked-children="Manual"
            />
          </a-form-item>
        </a-col>
        <a-col :xs="12" :sm="6" :md="6">
          <div class="period-info">
            <small>{{ currentPeriodText }}</small>
          </div>
        </a-col>
      </a-row>
    </div>

    <!-- KPI Stats Cards -->
    <a-row :gutter="[20, 20]">
      <a-col :xs="24" :sm="12" :md="6">
        <HStatCard
          :icon="DollarOutlined"
          icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
          label="Penyerapan Anggaran"
          :value="`${(dashboard?.budget_absorption?.rate || 0).toFixed(1)}%`"
          :change="`${formatCurrency(dashboard?.budget_absorption?.used || 0)} dari ${formatCurrency(dashboard?.budget_absorption?.total || 0)}`"
          :loading="loading"
          @click="drillDown('budget')"
          class="clickable-card"
        />
      </a-col>
      <a-col :xs="24" :sm="12" :md="6">
        <HStatCard
          :icon="CoffeeOutlined"
          icon-bg="linear-gradient(135deg, #05CD99 0%, #04b587 100%)"
          label="Total Porsi Terdistribusi"
          :value="formatNumber(dashboard?.nutrition_distribution?.total_portions || 0)"
          :change="`${dashboard?.nutrition_distribution?.schools_served || 0} sekolah • ${dashboard?.nutrition_distribution?.students_reached || 0} siswa`"
          :loading="loading"
          @click="drillDown('distribution')"
          class="clickable-card"
        />
      </a-col>
      <a-col :xs="24" :sm="12" :md="6">
        <HStatCard
          :icon="StarOutlined"
          icon-bg="linear-gradient(135deg, #FFB547 0%, #ff9f1a 100%)"
          label="Performa Supplier"
          :value="`${(dashboard?.supplier_performance?.average_rating || 0).toFixed(1)}/5.0`"
          :change="`${dashboard?.supplier_performance?.on_time_delivery_rate || 0}% ketepatan waktu`"
          :loading="loading"
          @click="drillDown('suppliers')"
          class="clickable-card"
        />
      </a-col>
      <a-col :xs="24" :sm="12" :md="6">
        <HStatCard
          :icon="ThunderboltOutlined"
          icon-bg="linear-gradient(135deg, #722ed1 0%, #531dab 100%)"
          label="Efisiensi Operasional"
          :value="`${(dashboard?.operational_efficiency?.score || 0).toFixed(1)}%`"
          change="Berdasarkan waktu produksi & distribusi"
          :loading="loading"
          @click="drillDown('efficiency')"
          class="clickable-card"
        />
      </a-col>
    </a-row>

    <!-- Budget Analysis & Nutrition Distribution -->
    <a-row :gutter="[20, 20]">
      <a-col :xs="24" :lg="12">
        <HChartCard
          title="Analisis Anggaran"
          subtitle="Penyerapan dan breakdown per kategori"
          :height="400"
          :loading="loading"
        >
          <template #header-right>
            <a-button type="link" size="small" @click="drillDown('budget')">
              Lihat Detail
            </a-button>
          </template>
          
          <div class="budget-overview">
            <a-row :gutter="16" style="margin-bottom: 16px;">
              <a-col :span="8">
                <div class="budget-item">
                  <div class="budget-label">Total Anggaran</div>
                  <div class="budget-value">{{ formatCurrency(dashboard?.budget_absorption?.total || 0) }}</div>
                </div>
              </a-col>
              <a-col :span="8">
                <div class="budget-item">
                  <div class="budget-label">Terpakai</div>
                  <div class="budget-value used">{{ formatCurrency(dashboard?.budget_absorption?.used || 0) }}</div>
                </div>
              </a-col>
              <a-col :span="8">
                <div class="budget-item">
                  <div class="budget-label">Sisa</div>
                  <div class="budget-value remaining">{{ formatCurrency(dashboard?.budget_absorption?.remaining || 0) }}</div>
                </div>
              </a-col>
            </a-row>

            <div class="budget-progress">
              <div class="progress-label">
                Penyerapan: {{ (dashboard?.budget_absorption?.rate || 0).toFixed(1) }}%
              </div>
              <a-progress 
                :percent="dashboard?.budget_absorption?.rate || 0" 
                :stroke-color="getBudgetAbsorptionColor(dashboard?.budget_absorption?.rate)"
                stroke-width="8"
              />
            </div>

            <div ref="budgetChartRef" style="height: 200px; margin-top: 16px;"></div>
          </div>
        </HChartCard>
      </a-col>

      <a-col :xs="24" :lg="12">
        <HChartCard
          title="Distribusi Gizi Kumulatif"
          subtitle="Total porsi, sekolah, dan siswa terlayani"
          :height="400"
          :loading="loading"
        >
          <template #header-right>
            <a-button type="link" size="small" @click="drillDown('distribution')">
              Lihat Detail
            </a-button>
          </template>

          <div class="nutrition-overview">
            <a-row :gutter="16" style="margin-bottom: 16px;">
              <a-col :xs="24" :sm="8">
                <div class="nutrition-metric">
                  <div class="metric-icon portions">
                    <CoffeeOutlined />
                  </div>
                  <div class="metric-content">
                    <div class="metric-value">{{ formatNumber(dashboard?.nutrition_distribution?.total_portions || 0) }}</div>
                    <div class="metric-label">Total Porsi</div>
                  </div>
                </div>
              </a-col>
              <a-col :xs="24" :sm="8">
                <div class="nutrition-metric">
                  <div class="metric-icon schools">
                    <HomeOutlined />
                  </div>
                  <div class="metric-content">
                    <div class="metric-value">{{ dashboard?.nutrition_distribution?.schools_served || 0 }}</div>
                    <div class="metric-label">Sekolah Terlayani</div>
                  </div>
                </div>
              </a-col>
              <a-col :xs="24" :sm="8">
                <div class="nutrition-metric">
                  <div class="metric-icon students">
                    <UserOutlined />
                  </div>
                  <div class="metric-content">
                    <div class="metric-value">{{ formatNumber(dashboard?.nutrition_distribution?.students_reached || 0) }}</div>
                    <div class="metric-label">Siswa Terjangkau</div>
                  </div>
                </div>
              </a-col>
            </a-row>

            <div ref="nutritionChartRef" style="height: 200px; margin-top: 16px;"></div>
          </div>
        </HChartCard>
      </a-col>
    </a-row>

    <!-- Supplier Performance -->
    <HChartCard
      title="Performa Supplier"
      :subtitle="`Rating Rata-rata: ${(dashboard?.supplier_performance?.average_rating || 0).toFixed(1)}/5.0`"
      :height="350"
      :loading="loading"
    >
      <template #header-right>
        <a-button type="link" size="small" @click="drillDown('suppliers')">
          Lihat Semua
        </a-button>
      </template>

      <div class="supplier-overview">
        <a-row :gutter="16" style="margin-bottom: 16px;">
          <a-col :xs="12" :sm="6">
            <div class="supplier-metric">
              <div class="metric-number">{{ dashboard?.supplier_performance?.total_suppliers || 0 }}</div>
              <div class="metric-label">Total Supplier</div>
            </div>
          </a-col>
          <a-col :xs="12" :sm="6">
            <div class="supplier-metric">
              <div class="metric-number active">{{ dashboard?.supplier_performance?.active_suppliers || 0 }}</div>
              <div class="metric-label">Supplier Aktif</div>
            </div>
          </a-col>
          <a-col :xs="12" :sm="6">
            <div class="supplier-metric">
              <div class="metric-number on-time">{{ (dashboard?.supplier_performance?.on_time_delivery_rate || 0).toFixed(1) }}%</div>
              <div class="metric-label">Ketepatan Waktu</div>
            </div>
          </a-col>
          <a-col :xs="12" :sm="6">
            <div class="supplier-metric">
              <div class="metric-number quality">{{ (dashboard?.supplier_performance?.quality_score || 0).toFixed(1) }}%</div>
              <div class="metric-label">Skor Kualitas</div>
            </div>
          </a-col>
        </a-row>

        <div class="top-suppliers" v-if="dashboard?.supplier_performance?.top_suppliers">
          <div class="section-title">Top 5 Supplier Terbaik</div>
          <a-row :gutter="[16, 16]">
            <a-col 
              v-for="(supplier, index) in dashboard.supplier_performance.top_suppliers.slice(0, 5)" 
              :key="supplier.id"
              :xs="24" :sm="12" :md="8" :lg="4"
            >
              <div class="supplier-item-card h-card">
                <div class="supplier-rank">{{ index + 1 }}</div>
                <div class="supplier-name">{{ supplier.name }}</div>
                <div class="supplier-rating">
                  <a-rate 
                    :value="supplier.rating" 
                    disabled 
                    :style="{ fontSize: '12px' }"
                  />
                  <span class="rating-text">{{ supplier.rating.toFixed(1) }}</span>
                </div>
                <div class="supplier-stats">
                  <div class="stat-item">
                    <span class="stat-label">Ketepatan:</span>
                    <span class="stat-value">{{ supplier.on_time_rate.toFixed(1) }}%</span>
                  </div>
                  <div class="stat-item">
                    <span class="stat-label">Transaksi:</span>
                    <span class="stat-value">{{ supplier.total_orders }}</span>
                  </div>
                </div>
              </div>
            </a-col>
          </a-row>
        </div>
      </div>
    </HChartCard>

    <!-- Trend Charts -->
    <a-row :gutter="[20, 20]">
      <a-col :xs="24" :lg="12">
        <HChartCard
          title="Tren Pengeluaran Anggaran"
          subtitle="Perbandingan budget vs pengeluaran aktual"
          :height="320"
          :loading="loading"
        >
          <div ref="budgetTrendChartRef" style="height: 100%; width: 100%;"></div>
        </HChartCard>
      </a-col>
      <a-col :xs="24" :lg="12">
        <HChartCard
          title="Tren Volume Distribusi"
          subtitle="Porsi dan sekolah terlayani per periode"
          :height="320"
          :loading="loading"
        >
          <div ref="distributionTrendChartRef" style="height: 100%; width: 100%;"></div>
        </HChartCard>
      </a-col>
    </a-row>

    <!-- Last Updated Info -->
    <div class="h-card update-info">
      <a-space :size="16" wrap>
        <span>
          <WifiOutlined v-if="isConnected" style="color: var(--h-success);" />
          <DisconnectOutlined v-else style="color: var(--h-error);" />
          {{ isConnected ? 'Terhubung' : 'Terputus' }}
        </span>
        <a-divider type="vertical" />
        <span>Terakhir diperbarui: {{ formatDateTime(dashboard?.updated_at) }}</span>
        <a-divider type="vertical" />
        <span>Periode: {{ currentPeriodText }}</span>
        <a-divider type="vertical" />
        <span>Auto-refresh: {{ autoRefreshEnabled ? 'Aktif' : 'Nonaktif' }}</span>
        <a-divider type="vertical" />
        <a-button @click="refreshData" :loading="loading" size="small">
          <template #icon><ReloadOutlined /></template>
          Refresh
        </a-button>
        <a-button @click="exportDashboard" :loading="exporting" size="small">
          <template #icon><DownloadOutlined /></template>
          Export
        </a-button>
      </a-space>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  DownloadOutlined,
  CoffeeOutlined,
  HomeOutlined,
  UserOutlined,
  DollarOutlined,
  StarOutlined,
  ThunderboltOutlined
} from '@ant-design/icons-vue'
import HStatCard from '@/components/horizon/HStatCard.vue'
import HChartCard from '@/components/horizon/HChartCard.vue'
import { useHorizonChart } from '@/composables/useHorizonChart'
import { getKepalaYayasanDashboard, exportDashboardData, syncDashboardToFirebase } from '@/services/dashboardService'
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
const selectedPeriod = ref('this_month')
const customDateRange = ref([])

// Chart refs
const nutritionChartRef = ref()
const budgetChartRef = ref()
const budgetTrendChartRef = ref()
const distributionTrendChartRef = ref()

// Firebase listeners
let dashboardListener = null
let autoRefreshInterval = null

// Computed properties
const currentPeriodText = computed(() => {
  const periodLabels = {
    this_month: `Bulan Ini (${dayjs().format('MMMM YYYY')})`,
    this_quarter: `Kuartal Ini (Q${dayjs().quarter()} ${dayjs().year()})`,
    this_year: `Tahun Ini (${dayjs().year()})`,
    last_month: `Bulan Lalu (${dayjs().subtract(1, 'month').format('MMMM YYYY')})`,
    last_quarter: `Kuartal Lalu (Q${dayjs().subtract(1, 'quarter').quarter()} ${dayjs().subtract(1, 'quarter').year()})`,
    custom: customDateRange.value.length === 2 ? 
      `${customDateRange.value[0].format('DD/MM/YYYY')} - ${customDateRange.value[1].format('DD/MM/YYYY')}` : 
      'Pilih rentang tanggal'
  }
  return periodLabels[selectedPeriod.value] || selectedPeriod.value
})

// Get date range based on selected period
const getDateRange = () => {
  const today = dayjs()
  
  switch (selectedPeriod.value) {
    case 'this_month':
      return [today.startOf('month').format('YYYY-MM-DD'), today.endOf('month').format('YYYY-MM-DD')]
    case 'this_quarter':
      return [today.startOf('quarter').format('YYYY-MM-DD'), today.endOf('quarter').format('YYYY-MM-DD')]
    case 'this_year':
      return [today.startOf('year').format('YYYY-MM-DD'), today.endOf('year').format('YYYY-MM-DD')]
    case 'last_month':
      const lastMonth = today.subtract(1, 'month')
      return [lastMonth.startOf('month').format('YYYY-MM-DD'), lastMonth.endOf('month').format('YYYY-MM-DD')]
    case 'last_quarter':
      const lastQuarter = today.subtract(1, 'quarter')
      return [lastQuarter.startOf('quarter').format('YYYY-MM-DD'), lastQuarter.endOf('quarter').format('YYYY-MM-DD')]
    case 'custom':
      if (customDateRange.value && customDateRange.value.length === 2) {
        return [customDateRange.value[0].format('YYYY-MM-DD'), customDateRange.value[1].format('YYYY-MM-DD')]
      }
      return [today.startOf('month').format('YYYY-MM-DD'), today.endOf('month').format('YYYY-MM-DD')]
    default:
      return [today.startOf('month').format('YYYY-MM-DD'), today.endOf('month').format('YYYY-MM-DD')]
  }
}

// Load dashboard data from API
const loadDashboardData = async () => {
  loading.value = true
  try {
    const [startDate, endDate] = getDateRange()
    const response = await getKepalaYayasanDashboard(startDate, endDate)
    if (response.success) {
      dashboard.value = response.dashboard
      
      // Generate charts after data is loaded
      await nextTick()
      generateCharts()
      
      // Sync to Firebase for real-time updates
      await syncDashboardToFirebase('kepala_yayasan', startDate, endDate)
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

// Handle period change
const handlePeriodChange = (value) => {
  selectedPeriod.value = value
  if (value !== 'custom') {
    loadDashboardData()
  }
}

// Handle custom date range change
const handleCustomDateChange = (dates) => {
  customDateRange.value = dates
  if (dates && dates.length === 2) {
    loadDashboardData()
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
    const [startDate, endDate] = getDateRange()
    const response = await exportDashboardData('kepala_yayasan', 'json', startDate, endDate)
    if (response.success) {
      const dataStr = JSON.stringify(response.data, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `dashboard-kepala-yayasan-${startDate}-${endDate}.json`
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
    case 'budget':
      router.push('/financial-reports')
      break
    case 'distribution':
      router.push('/delivery-tasks')
      break
    case 'suppliers':
      router.push('/suppliers')
      break
    case 'efficiency':
      router.push('/kds/cooking')
      break
    default:
      console.log('Unknown drill down type:', type)
  }
}

// Generate charts using useHorizonChart
const generateCharts = () => {
  generateNutritionChart()
  generateBudgetChart()
  generateBudgetTrendChart()
  generateDistributionTrendChart()
}

// Generate nutrition distribution chart
const { setOption: setNutritionOption } = useHorizonChart(nutritionChartRef)
const generateNutritionChart = () => {
  if (nutritionChartRef.value && dashboard.value?.nutrition_distribution?.daily_distribution) {
    const dailyData = dashboard.value.nutrition_distribution.daily_distribution
    const dates = dailyData.map(item => dayjs(item.date).format('DD/MM'))
    const portions = dailyData.map(item => item.portions)

    setNutritionOption({
      tooltip: {
        trigger: 'axis',
        formatter: '{b}: {c} porsi'
      },
      xAxis: {
        type: 'category',
        data: dates
      },
      yAxis: {
        type: 'value'
      },
      series: [{
        data: portions,
        type: 'line',
        smooth: true,
        itemStyle: { color: '#05CD99' },
        areaStyle: { opacity: 0.3 }
      }]
    })
  }
}

// Generate budget breakdown chart
const { setOption: setBudgetOption } = useHorizonChart(budgetChartRef)
const generateBudgetChart = () => {
  if (budgetChartRef.value && dashboard.value?.budget_absorption?.categories) {
    const categories = dashboard.value.budget_absorption.categories
    const data = categories.map(cat => ({
      name: cat.name,
      value: cat.amount
    }))

    setBudgetOption({
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)'
      },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        data: data,
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }]
    })
  }
}

// Generate budget trend chart
const { setOption: setBudgetTrendOption } = useHorizonChart(budgetTrendChartRef)
const generateBudgetTrendChart = () => {
  if (budgetTrendChartRef.value && dashboard.value?.budget_trends) {
    const trendData = dashboard.value.budget_trends
    const periods = trendData.map(item => item.period)
    const spending = trendData.map(item => item.spending)
    const budget = trendData.map(item => item.budget)

    setBudgetTrendOption({
      tooltip: {
        trigger: 'axis',
        formatter: function(params) {
          let result = params[0].name + '<br/>'
          params.forEach(param => {
            result += param.seriesName + ': ' + formatCurrency(param.value) + '<br/>'
          })
          return result
        }
      },
      legend: {
        data: ['Budget', 'Pengeluaran']
      },
      xAxis: {
        type: 'category',
        data: periods
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: function (value) {
            return (value / 1000000).toFixed(1) + 'M'
          }
        }
      },
      series: [
        {
          name: 'Budget',
          type: 'line',
          data: budget,
          itemStyle: { color: '#05CD99' }
        },
        {
          name: 'Pengeluaran',
          type: 'line',
          data: spending,
          itemStyle: { color: '#EE5D50' }
        }
      ]
    })
  }
}

// Generate distribution trend chart
const { setOption: setDistributionTrendOption } = useHorizonChart(distributionTrendChartRef)
const generateDistributionTrendChart = () => {
  if (distributionTrendChartRef.value && dashboard.value?.distribution_trends) {
    const trendData = dashboard.value.distribution_trends
    const periods = trendData.map(item => item.period)
    const portions = trendData.map(item => item.portions)
    const schools = trendData.map(item => item.schools)

    setDistributionTrendOption({
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['Porsi', 'Sekolah']
      },
      xAxis: {
        type: 'category',
        data: periods
      },
      yAxis: [
        {
          type: 'value',
          name: 'Porsi',
          position: 'left'
        },
        {
          type: 'value',
          name: 'Sekolah',
          position: 'right'
        }
      ],
      series: [
        {
          name: 'Porsi',
          type: 'bar',
          data: portions,
          itemStyle: { color: '#5A4372' }
        },
        {
          name: 'Sekolah',
          type: 'line',
          yAxisIndex: 1,
          data: schools,
          itemStyle: { color: '#05CD99' }
        }
      ]
    })
  }
}

// Setup Firebase real-time listeners
const setupFirebaseListeners = () => {
  const dashboardRef = dbRef(database, '/dashboard/kepala_yayasan')
  dashboardListener = onValue(
    dashboardRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      if (data) {
        dashboard.value = data
        nextTick(() => {
          generateCharts()
        })
      }
    },
    (error) => {
      console.error('Firebase dashboard listener error:', error)
      isConnected.value = false
    }
  )
}

// Cleanup Firebase listeners
const cleanupFirebaseListeners = () => {
  if (dashboardListener) {
    const dashboardRef = dbRef(database, '/dashboard/kepala_yayasan')
    off(dashboardRef)
    dashboardListener = null
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
const getBudgetAbsorptionColor = (rate) => {
  if (!rate) return '#5A4372'
  if (rate >= 80) return '#EE5D50' // Red for high absorption (might be overspending)
  if (rate >= 60) return '#FFB547' // Orange for moderate absorption
  if (rate >= 40) return '#05CD99' // Green for good absorption
  return '#5A4372' // Purple for low absorption
}

const formatCurrency = (value) => {
  if (!value) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(value)
}

const formatNumber = (value) => {
  if (!value) return '0'
  return new Intl.NumberFormat('id-ID').format(value)
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
/* Period Info */
.period-info {
  color: var(--h-text-secondary);
  font-size: var(--h-text-sm);
  padding: var(--h-spacing-2) 0;
}

/* Clickable Cards */
.clickable-card {
  cursor: pointer;
  transition: all var(--h-transition-base);
}

.clickable-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--h-shadow-xl);
}

/* Budget Overview */
.budget-overview {
  width: 100%;
}

.budget-item {
  text-align: center;
  padding: var(--h-spacing-3);
  background-color: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
}

.budget-label {
  font-size: var(--h-text-xs);
  color: var(--h-text-secondary);
  margin-bottom: var(--h-spacing-1);
  font-weight: var(--h-font-medium);
}

.budget-value {
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-primary);
}

.budget-value.used {
  color: var(--h-error);
}

.budget-value.remaining {
  color: var(--h-success);
}

.budget-progress {
  margin: var(--h-spacing-4) 0;
}

.progress-label {
  margin-bottom: var(--h-spacing-2);
  font-weight: var(--h-font-medium);
  color: var(--h-text-primary);
  font-size: var(--h-text-sm);
}

/* Nutrition Overview */
.nutrition-overview {
  width: 100%;
}

.nutrition-metric {
  display: flex;
  align-items: center;
  padding: var(--h-spacing-3);
  background-color: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
  transition: all var(--h-transition-base);
  gap: var(--h-spacing-3);
}

.nutrition-metric:hover {
  background-color: #F8FDEA;
}

.metric-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.metric-icon.portions {
  background-color: rgba(5, 205, 153, 0.1);
  color: var(--h-success);
}

.metric-icon.schools {
  background-color: rgba(90, 67, 114, 0.1);
  color: var(--h-primary);
}

.metric-icon.students {
  background-color: rgba(255, 181, 71, 0.1);
  color: var(--h-warning);
}

.metric-content {
  flex: 1;
  min-width: 0;
}

.metric-value {
  font-size: var(--h-text-xl);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  line-height: 1;
}

.metric-label {
  font-size: var(--h-text-xs);
  color: var(--h-text-secondary);
  margin-top: var(--h-spacing-1);
  font-weight: var(--h-font-medium);
}

/* Supplier Overview */
.supplier-overview {
  width: 100%;
}

.supplier-metric {
  text-align: center;
  padding: var(--h-spacing-4);
  background-color: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
}

.metric-number {
  font-size: var(--h-text-2xl);
  font-weight: var(--h-font-bold);
  color: var(--h-primary);
  margin-bottom: var(--h-spacing-1);
}

.metric-number.active {
  color: var(--h-success);
}

.metric-number.on-time {
  color: var(--h-warning);
}

.metric-number.quality {
  color: #722ed1;
}

.top-suppliers {
  margin-top: var(--h-spacing-6);
}

.section-title {
  font-weight: var(--h-font-semibold);
  margin-bottom: var(--h-spacing-4);
  color: var(--h-text-primary);
  font-size: var(--h-text-base);
}

.supplier-item-card {
  text-align: center;
  transition: all var(--h-transition-base);
  position: relative;
  padding: var(--h-spacing-4);
  min-height: 140px;
}

.supplier-item-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--h-shadow-lg);
}

.supplier-rank {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  background-color: var(--h-primary);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-bold);
}

.supplier-name {
  font-weight: var(--h-font-semibold);
  margin-bottom: var(--h-spacing-2);
  color: var(--h-text-primary);
  font-size: var(--h-text-sm);
}

.supplier-rating {
  margin-bottom: var(--h-spacing-2);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--h-spacing-1);
}

.rating-text {
  font-size: var(--h-text-xs);
  color: var(--h-warning);
  font-weight: var(--h-font-medium);
}

.supplier-stats {
  font-size: var(--h-text-xs);
}

.stat-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--h-spacing-1);
}

.stat-label {
  color: var(--h-text-secondary);
}

.stat-value {
  font-weight: var(--h-font-medium);
  color: var(--h-text-primary);
}

/* Update Info */
.update-info {
  text-align: center;
  background-color: var(--h-bg-light);
  padding: var(--h-spacing-4);
}

.update-info :deep(.ant-space-item) {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .nutrition-metric {
    padding: var(--h-spacing-2);
  }
  
  .metric-icon {
    width: 32px;
    height: 32px;
    font-size: 14px;
  }
  
  .metric-value {
    font-size: var(--h-text-base);
  }
  
  .supplier-metric {
    padding: var(--h-spacing-3);
  }
  
  .metric-number {
    font-size: var(--h-text-xl);
  }
  
  .budget-item {
    margin-bottom: var(--h-spacing-2);
  }
  
  .supplier-item-card {
    min-height: 120px;
  }
}

/* Dark Mode Support */
.dark .period-info {
  color: var(--h-text-secondary);
}

.dark .budget-item {
  background-color: rgba(90, 67, 114, 0.1);
}

.dark .nutrition-metric {
  background-color: rgba(90, 67, 114, 0.1);
}

.dark .nutrition-metric:hover {
  background-color: rgba(90, 67, 114, 0.2);
}

.dark .supplier-metric {
  background-color: rgba(90, 67, 114, 0.1);
}

.dark .update-info {
  background-color: rgba(90, 67, 114, 0.1);
}

.dark .metric-icon.portions {
  background-color: rgba(5, 205, 153, 0.2);
}

.dark .metric-icon.schools {
  background-color: rgba(90, 67, 114, 0.2);
}

.dark .metric-icon.students {
  background-color: rgba(255, 181, 71, 0.2);
}
</style>
