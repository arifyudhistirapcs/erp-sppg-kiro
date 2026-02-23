<template>
  <div class="dashboard-kepala-yayasan">
    <a-page-header
      title="Dashboard Kepala Yayasan"
      sub-title="Monitoring penyerapan anggaran dan capaian gizi secara real-time"
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
        <!-- Time Period Selection -->
        <a-card size="small" style="margin-bottom: 16px;">
          <a-row :gutter="16" align="middle">
            <a-col :span="6">
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
            <a-col :span="8" v-if="selectedPeriod === 'custom'">
              <a-form-item label="Rentang Tanggal" style="margin-bottom: 0;">
                <a-range-picker
                  v-model:value="customDateRange"
                  @change="handleCustomDateChange"
                  style="width: 100%"
                  format="DD/MM/YYYY"
                />
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item style="margin-bottom: 0;">
                <a-switch 
                  v-model:checked="autoRefreshEnabled" 
                  @change="toggleAutoRefresh"
                  checked-children="Auto"
                  un-checked-children="Manual"
                />
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <div class="period-info">
                <small>{{ currentPeriodText }}</small>
              </div>
            </a-col>
          </a-row>
        </a-card>

        <!-- KPI Cards -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('budget')"
              hoverable
            >
              <a-statistic
                title="Penyerapan Anggaran"
                :value="dashboard?.budget_absorption?.rate || 0"
                :precision="1"
                :value-style="{ color: getBudgetAbsorptionColor(dashboard?.budget_absorption?.rate) }"
                suffix="%"
              />
              <div class="kpi-subtitle">
                {{ formatCurrency(dashboard?.budget_absorption?.used || 0) }} dari {{ formatCurrency(dashboard?.budget_absorption?.total || 0) }}
              </div>
              <div class="kpi-progress">
                <a-progress 
                  :percent="dashboard?.budget_absorption?.rate || 0" 
                  :stroke-color="getBudgetAbsorptionColor(dashboard?.budget_absorption?.rate)"
                  :show-info="false"
                  size="small"
                />
              </div>
            </a-card>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('distribution')"
              hoverable
            >
              <a-statistic
                title="Total Porsi Terdistribusi"
                :value="dashboard?.nutrition_distribution?.total_portions || 0"
                :value-style="{ color: '#1890ff' }"
                suffix="porsi"
              />
              <div class="kpi-subtitle">
                {{ dashboard?.nutrition_distribution?.schools_served || 0 }} sekolah • {{ dashboard?.nutrition_distribution?.students_reached || 0 }} siswa
              </div>
            </a-card>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('suppliers')"
              hoverable
            >
              <a-statistic
                title="Performa Supplier"
                :value="dashboard?.supplier_performance?.average_rating || 0"
                :precision="1"
                :value-style="{ color: getSupplierPerformanceColor(dashboard?.supplier_performance?.average_rating) }"
                suffix="/5.0"
              />
              <div class="kpi-subtitle">
                {{ dashboard?.supplier_performance?.on_time_delivery_rate || 0 }}% ketepatan waktu
              </div>
            </a-card>
          </a-col>
          <a-col :xs="24" :sm="12" :md="6">
            <a-card 
              class="kpi-card clickable" 
              @click="drillDown('efficiency')"
              hoverable
            >
              <a-statistic
                title="Efisiensi Operasional"
                :value="dashboard?.operational_efficiency?.score || 0"
                :precision="1"
                :value-style="{ color: '#52c41a' }"
                suffix="%"
              />
              <div class="kpi-subtitle">
                Berdasarkan waktu produksi & distribusi
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- Budget Analysis -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :xs="24" :lg="12">
            <a-card title="Analisis Anggaran" class="budget-card">
              <template #extra>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="drillDown('budget')"
                >
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

                <a-divider />

                <div class="budget-breakdown">
                  <div class="breakdown-title">Breakdown per Kategori</div>
                  <div class="breakdown-items">
                    <div 
                      v-for="category in dashboard?.budget_absorption?.categories || []" 
                      :key="category.name"
                      class="breakdown-item"
                    >
                      <div class="category-info">
                        <span class="category-name">{{ category.name }}</span>
                        <span class="category-amount">{{ formatCurrency(category.amount) }}</span>
                      </div>
                      <a-progress 
                        :percent="category.percentage" 
                        :stroke-color="getCategoryColor(category.name)"
                        :show-info="false"
                        size="small"
                      />
                      <div class="category-percentage">{{ category.percentage.toFixed(1) }}%</div>
                    </div>
                  </div>
                </div>
              </div>
            </a-card>
          </a-col>

          <a-col :xs="24" :lg="12">
            <a-card title="Distribusi Gizi Kumulatif" class="nutrition-card">
              <template #extra>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="drillDown('distribution')"
                >
                  Lihat Detail
                </a-button>
              </template>

              <div class="nutrition-overview">
                <a-row :gutter="16" style="margin-bottom: 16px;">
                  <a-col :span="8">
                    <div class="nutrition-metric">
                      <div class="metric-icon portions">
                        <coffee-outlined />
                      </div>
                      <div class="metric-content">
                        <div class="metric-value">{{ formatNumber(dashboard?.nutrition_distribution?.total_portions || 0) }}</div>
                        <div class="metric-label">Total Porsi</div>
                      </div>
                    </div>
                  </a-col>
                  <a-col :span="8">
                    <div class="nutrition-metric">
                      <div class="metric-icon schools">
                        <home-outlined />
                      </div>
                      <div class="metric-content">
                        <div class="metric-value">{{ dashboard?.nutrition_distribution?.schools_served || 0 }}</div>
                        <div class="metric-label">Sekolah Terlayani</div>
                      </div>
                    </div>
                  </a-col>
                  <a-col :span="8">
                    <div class="nutrition-metric">
                      <div class="metric-icon students">
                        <user-outlined />
                      </div>
                      <div class="metric-content">
                        <div class="metric-value">{{ formatNumber(dashboard?.nutrition_distribution?.students_reached || 0) }}</div>
                        <div class="metric-label">Siswa Terjangkau</div>
                      </div>
                    </div>
                  </a-col>
                </a-row>

                <a-divider />

                <div class="nutrition-details">
                  <div class="detail-row">
                    <span class="detail-label">Rata-rata Porsi per Sekolah:</span>
                    <span class="detail-value">{{ (dashboard?.nutrition_distribution?.avg_portions_per_school || 0).toFixed(1) }} porsi</span>
                  </div>
                  <div class="detail-row">
                    <span class="detail-label">Rata-rata Porsi per Siswa:</span>
                    <span class="detail-value">{{ (dashboard?.nutrition_distribution?.avg_portions_per_student || 0).toFixed(1) }} porsi</span>
                  </div>
                  <div class="detail-row">
                    <span class="detail-label">Tingkat Cakupan:</span>
                    <span class="detail-value">{{ (dashboard?.nutrition_distribution?.coverage_rate || 0).toFixed(1) }}%</span>
                  </div>
                </div>

                <div class="nutrition-chart" ref="nutritionChartRef" style="height: 200px; margin-top: 16px;"></div>
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- Supplier Performance -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :span="24">
            <a-card title="Performa Supplier" class="supplier-card">
              <template #extra>
                <a-space>
                  <a-tag :color="getSupplierPerformanceColor(dashboard?.supplier_performance?.average_rating) === '#52c41a' ? 'green' : dashboard?.supplier_performance?.average_rating >= 3 ? 'orange' : 'red'">
                    Rating Rata-rata: {{ (dashboard?.supplier_performance?.average_rating || 0).toFixed(1) }}/5.0
                  </a-tag>
                  <a-button 
                    type="link" 
                    size="small" 
                    @click="drillDown('suppliers')"
                  >
                    Lihat Semua
                  </a-button>
                </a-space>
              </template>

              <div class="supplier-overview">
                <a-row :gutter="16" style="margin-bottom: 16px;">
                  <a-col :span="6">
                    <div class="supplier-metric">
                      <div class="metric-number">{{ dashboard?.supplier_performance?.total_suppliers || 0 }}</div>
                      <div class="metric-label">Total Supplier</div>
                    </div>
                  </a-col>
                  <a-col :span="6">
                    <div class="supplier-metric">
                      <div class="metric-number active">{{ dashboard?.supplier_performance?.active_suppliers || 0 }}</div>
                      <div class="metric-label">Supplier Aktif</div>
                    </div>
                  </a-col>
                  <a-col :span="6">
                    <div class="supplier-metric">
                      <div class="metric-number on-time">{{ (dashboard?.supplier_performance?.on_time_delivery_rate || 0).toFixed(1) }}%</div>
                      <div class="metric-label">Ketepatan Waktu</div>
                    </div>
                  </a-col>
                  <a-col :span="6">
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
                      <a-card size="small" class="supplier-item-card">
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
                      </a-card>
                    </a-col>
                  </a-row>
                </div>
              </div>
            </a-card>
          </a-col>
        </a-row>

        <!-- Trend Charts -->
        <a-row :gutter="[16, 16]" style="margin-bottom: 24px;">
          <a-col :xs="24" :lg="12">
            <a-card title="Tren Pengeluaran Anggaran" class="trend-card">
              <div ref="budgetTrendChartRef" style="height: 300px;"></div>
            </a-card>
          </a-col>
          <a-col :xs="24" :lg="12">
            <a-card title="Tren Volume Distribusi" class="trend-card">
              <div ref="distributionTrendChartRef" style="height: 300px;"></div>
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
                <span>Periode: {{ currentPeriodText }}</span>
                <a-divider type="vertical" />
                <span>Auto-refresh: {{ autoRefreshEnabled ? 'Aktif' : 'Nonaktif' }}</span>
              </a-space>
            </a-card>
          </a-col>
        </a-row>
      </a-spin>
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
  UserOutlined
} from '@ant-design/icons-vue'
import { getKepalaYayasanDashboard, exportDashboardData, syncDashboardToFirebase } from '@/services/dashboardService'
import { database } from '@/services/firebase'
import { ref as dbRef, onValue, off } from 'firebase/database'
import dayjs from 'dayjs'
import * as echarts from 'echarts'

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
const budgetTrendChartRef = ref()
const distributionTrendChartRef = ref()
let nutritionChart = null
let budgetTrendChart = null
let distributionTrendChart = null

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

// Lifecycle hooks
onMounted(() => {
  loadDashboardData()
  setupFirebaseListeners()
  toggleAutoRefresh(autoRefreshEnabled.value)
})

onUnmounted(() => {
  cleanupFirebaseListeners()
  cleanup()
  if (autoRefreshInterval) {
    clearInterval(autoRefreshInterval)
  }
})
</script>ue) => {
  if (!value) return '0'
  return new Intl.NumberFormat('id-ID').format(value)
}

const formatDateTime = (datetime) => {
  if (!datetime) return '-'
  return dayjs(datetime).format('DD/MM/YYYY HH:mm')
}

// Cleanup charts on unmount
const cleanup = () => {
  if (nutritionChart) {
    nutritionChart.dispose()
    nutritionChart = null
  }
  if (budgetTrendChart) {
    budgetTrendChart.dispose()
    budgetTrendChart = null
  }
  if (distributionTrendChart) {
    distributionTrendChart.dispose()
    dency: 'IDR',
    minimumFractionDigits: 0
  }).format(value)
}

const formatNumber = (valSupplierPerformanceColor = (rating) => {
  if (rating >= 4) return '#52c41a'
  if (rating >= 3) return '#faad14'
  return '#ff4d4f'
}

const getCategoryColor = (category) => {
  const colors = {
    'Bahan Baku': '#1890ff',
    'Gaji': '#52c41a',
    'Utilitas': '#faad14',
    'Operasional': '#722ed1',
    'Lainnya': '#f5222d'
  }
  return colors[category] || '#d9d9d9'
}

const formatCurrency = (value) => {
  if (!value) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    curr >= 40) return '#52c41a' // Green for good absorption
  return '#1890ff' // Blue for low absorption
}

const get) => {
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
  if (rate >= 80) return '#ff4d4f' // Red for high absorption (might be overspending)
  if (rate >= 60) return '#faad14' // Orange for moderate absorption
  if (rateonst toggleAutoRefresh = (enablede = data
        // Regenerate charts with new data
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
c(snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      if (data) {
        dashboard.valur',
          data: portions,
          itemStyle: { color: '#1890ff' }
        },
        {
          name: 'Sekolah',
          type: 'line',
          yAxisIndex: 1,
          data: schools,
          itemStyle: { color: '#52c41a' }
        }
      ]
    }
    
    distributionTrendChart.setOption(option)
  }
}

// Setup Firebase real-time listeners
const setupFirebaseListeners = () => {
  const dashboardRef = dbRef(database, '/dashboard/kepala_yayasan')
  dashboardListener = onValue(
    dashboardRef,
    
      series: [
        {
          name: 'Porsi',
          type: 'ba   title: {
        text: 'Tren Volume Distribusi',
        textStyle: { fontSize: 14 }
      },
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['Porsi', 'Sekolah'],
        top: 25
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
      ],ata.map(item => item.portions)
    const schools = trendData.map(item => item.schools)

    const option = {
    
    budgetTrendChart.setOption(option)
  }
}

// Generate distribution trend chart
const generateDistributionTrendChart = () => {
  if (distributionTrendChartRef.value && dashboard.value?.distribution_trends) {
    if (distributionTrendChart) {
      distributionTrendChart.dispose()
    }
    
    distributionTrendChart = echarts.init(distributionTrendChartRef.value)
    
    const trendData = dashboard.value.distribution_trends
    const periods = trendData.map(item => item.period)
    const portions = trendDcolor: '#52c41a' }
        },
        {
          name: 'Pengeluaran',
          type: 'line',
          data: spending,
          itemStyle: { color: '#ff4d4f' }
        }
      ]
    }
   
          })
          return result
        }
      },
      legend: {
        data: ['Budget', 'Pengeluaran'],
        top: 25
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
          itemStyle: { map(item => item.period)
    const spending = trendData.map(item => item.spending)
    const budget = trendData.map(item => item.budget)

    const option = {
      title: {
        text: 'Tren Pengeluaran vs Budget',
        textStyle: { fontSize: 14 }
      },
      tooltip: {
        trigger: 'axis',
        formatter: function(params) {
          let result = params[0].name + '<br/>'
          params.forEach(param => {
            result += param.seriesName + ': ' + formatCurrency(param.value) + '<br/>' dashboard.value.budget_trends
    const periods = trendData.: [{
        data: portions,
        type: 'line',
        smooth: true,
        itemStyle: { color: '#1890ff' },
        areaStyle: { opacity: 0.3 }
      }]
    }
    
    nutritionChart.setOption(option)
  }
}

// Generate budget trend chart
const generateBudgetTrendChart = () => {
  if (budgetTrendChartRef.value && dashboard.value?.budget_trends) {
    if (budgetTrendChart) {
      budgetTrendChart.dispose()
    }
    
    budgetTrendChart = echarts.init(budgetTrendChartRef.value)
    
    const trendData =e',
        axisLabel: { fontSize: 10 }
      },
      seriesion_distribution.daily_distribution
    const dates = dailyData.map(item => dayjs(item.date).format('DD/MM'))
    const portions = dailyData.map(item => item.portions)

    const option = {
      title: {
        text: 'Distribusi Harian',
        textStyle: { fontSize: 12 }
      },
      tooltip: {
        trigger: 'axis',
        formatter: '{b}: {c} porsi'
      },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: { fontSize: 10 }
      },
      yAxis: {
        type: 'value)
    
    const dailyData = dashboard.value.nutriteak
    default:
      console.log('Unknown drill down type:', type)
  }
}

// Generate charts
const generateCharts = () => {
  generateNutritionChart()
  generateBudgetTrendChart()
  generateDistributionTrendChart()
}

// Generate nutrition distribution chart
const generateNutritionChart = () => {
  if (nutritionChartRef.value && dashboard.value?.nutrition_distribution?.daily_distribution) {
    if (nutritionChart) {
      nutritionChart.dispose()
    }
    
    nutritionChart = echarts.init(nutritionChartRef.valuse 'suppliers':
      router.push('/suppliers')
      break
    case 'efficiency':
      router.push('/kds/cooking')
      brashboard berhasil diexport')
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
    caclick()
      document.body.removeChild(link)
      URL.revokeObjectURL(url)
      
      message.success('Desponse = await exportDashboardData('kepala_yayasan', 'json', startDate, endDate)
    if (response.success) {
      // Create download link
      const dataStr = JSON.stringify(response.data, null, 2)
      const dataBlob = new Blob([dataStr], { type: 'application/json' })
      const url = URL.createObjectURL(dataBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `dashboard-kepala-yayasan-${startDate}-${endDate}.json`
      document.body.appendChild(link)
      link.rd
const exportDashboard = async () => {
  exporting.value = true
  try {
    const [startDate, endDate] = getDateRange()
    const ror)
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

// Export dashboae updates
      await syncDashboardToFirebase('kepala_yayasan', startDate, endDate)
    } else {
      message.error(response.message || 'Gagal memuat data dashboard')
    }
  } catch (error) {
    console.error('Error loading dashboard:', err').format('YYYY-MM-DD'), today.endOf('month').format('YYYY-MM-DD')]
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
      
      // Sync to Firebase for real-timlastQuarter = today.subtract(1, 'quarter')
      return [lastQuarter.startOf('quarter').format('YYYY-MM-DD'), lastQuarter.endOf('quarter').format('YYYY-MM-DD')]
    case 'custom':
      if (customDateRange.value && customDateRange.value.length === 2) {
        return [customDateRange.value[0].format('YYYY-MM-DD'), customDateRange.value[1].format('YYYY-MM-DD')]
      }
      return [today.startOf('month').format('YYYY-MM-DD'), today.endOf('month').format('YYYY-MM-DD')]
    default:
      return [today.startOf('month today.endOf('month').format('YYYY-MM-DD')]
    case 'this_quarter':
      return [today.startOf('quarter').format('YYYY-MM-DD'), today.endOf('quarter').format('YYYY-MM-DD')]
    case 'this_year':
      return [today.startOf('year').format('YYYY-MM-DD'), today.endOf('year').format('YYYY-MM-DD')]
    case 'last_month':
      const lastMonth = today.subtract(1, 'month')
      return [lastMonth.startOf('month').format('YYYY-MM-DD'), lastMonth.endOf('month').format('YYYY-MM-DD')]
    case 'last_quarter':
      const 
<style scoped>
.dashboard-kepala-yayasan {
  padding: 24px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.dashboard-content {
  margin-top: 16px;
}

.period-info {
  color: rgba(0, 0, 0, 0.65);
  font-size: 12px;
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

.kpi-progress {
  margin-top: 8px;
}

.budget-card {
  border-left: 4px solid #1890ff;
}

.budget-overview {
  margin-bottom: 16px;
}

.budget-item {
  text-align: center;
}

.budget-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
  margin-bottom: 4px;
}

.budget-value {
  font-size: 16px;
  font-weight: bold;
  color: #1890ff;
}

.budget-value.used {
  color: #ff4d4f;
}

.budget-value.remaining {
  color: #52c41a;
}

.budget-progress {
  margin: 16px 0;
}

.progress-label {
  margin-bottom: 8px;
  font-weight: 500;
}

.budget-breakdown {
  margin-top: 16px;
}

.breakdown-title {
  font-weight: 500;
  margin-bottom: 12px;
  color: #262626;
}

.breakdown-items {
  space-y: 8px;
}

.breakdown-item {
  margin-bottom: 12px;
}

.category-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.category-name {
  font-size: 12px;
  color: #262626;
}

.category-amount {
  font-size: 12px;
  font-weight: 500;
  color: #1890ff;
}

.category-percentage {
  text-align: right;
  font-size: 11px;
  color: rgba(0, 0, 0, 0.45);
  margin-top: 2px;
}

.nutrition-card {
  border-left: 4px solid #52c41a;
}

.nutrition-overview {
  margin-bottom: 16px;
}

.nutrition-metric {
  display: flex;
  align-items: center;
  padding: 12px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.nutrition-metric:hover {
  background-color: #f0f0f0;
}

.metric-icon {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 18px;
}

.metric-icon.portions {
  background-color: #e6f7ff;
  color: #1890ff;
}

.metric-icon.schools {
  background-color: #f6ffed;
  color: #52c41a;
}

.metric-icon.students {
  background-color: #fff2e8;
  color: #fa8c16;
}

.metric-content {
  flex: 1;
}

.metric-value {
  font-size: 20px;
  font-weight: bold;
  color: #262626;
  line-height: 1;
}

.metric-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
  margin-top: 2px;
}

.nutrition-details {
  margin-top: 16px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
  border-bottom: 1px solid #f0f0f0;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
}

.detail-value {
  font-size: 12px;
  font-weight: 500;
  color: #262626;
}

.supplier-card {
  border-left: 4px solid #722ed1;
}

.supplier-overview {
  margin-bottom: 16px;
}

.supplier-metric {
  text-align: center;
  padding: 16px;
  background-color: #fafafa;
  border-radius: 6px;
}

.metric-number {
  font-size: 24px;
  font-weight: bold;
  color: #1890ff;
  margin-bottom: 4px;
}

.metric-number.active {
  color: #52c41a;
}

.metric-number.on-time {
  color: #faad14;
}

.metric-number.quality {
  color: #722ed1;
}

.metric-label {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.65);
}

.top-suppliers {
  margin-top: 24px;
}

.section-title {
  font-weight: 500;
  margin-bottom: 16px;
  color: #262626;
}

.supplier-item-card {
  text-align: center;
  transition: all 0.3s ease;
  position: relative;
}

.supplier-item-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.supplier-rank {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  background-color: #1890ff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
}

.supplier-name {
  font-weight: 500;
  margin-bottom: 8px;
  color: #262626;
  font-size: 14px;
}

.supplier-rating {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.rating-text {
  font-size: 12px;
  color: #faad14;
  font-weight: 500;
}

.supplier-stats {
  font-size: 11px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2px;
}

.stat-label {
  color: rgba(0, 0, 0, 0.65);
}

.stat-value {
  font-weight: 500;
  color: #262626;
}

.trend-card {
  height: 100%;
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

:deep(.ant-rate) {
  font-size: 12px;
}

:deep(.ant-rate-star) {
  margin-right: 2px;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .dashboard-kepala-yayasan {
    padding: 16px;
  }
  
  .nutrition-metric {
    padding: 8px;
  }
  
  .metric-icon {
    width: 32px;
    height: 32px;
    font-size: 14px;
    margin-right: 8px;
  }
  
  .metric-value {
    font-size: 16px;
  }
  
  .supplier-metric {
    padding: 12px;
  }
  
  .metric-number {
    font-size: 18px;
  }
}
</style>