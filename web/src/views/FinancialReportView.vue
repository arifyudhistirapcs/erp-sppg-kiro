<template>
  <div class="financial-report">
    <a-page-header
      title="Laporan Keuangan"
      sub-title="Laporan arus kas dan analisis keuangan komprehensif"
    >
      <template #extra>
        <a-space>
          <a-button @click="exportPDF" :loading="exportingPDF">
            <template #icon><FilePdfOutlined /></template>
            Export PDF
          </a-button>
          <a-button @click="exportExcel" :loading="exportingExcel">
            <template #icon><FileExcelOutlined /></template>
            Export Excel
          </a-button>
          <a-button type="primary" @click="generateReport" :loading="loading">
            <template #icon><BarChartOutlined /></template>
            Generate Laporan
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="24">
        <!-- Report Configuration -->
        <a-card size="small" title="Konfigurasi Laporan">
          <a-row :gutter="16">
            <a-col :span="8">
              <a-form-item label="Rentang Tanggal">
                <a-range-picker
                  v-model:value="dateRange"
                  style="width: 100%"
                  placeholder="['Tanggal Mulai', 'Tanggal Akhir']"
                  format="DD/MM/YYYY"
                />
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item label="Periode Cepat">
                <a-select
                  v-model:value="quickPeriod"
                  placeholder="Pilih periode"
                  style="width: 100%"
                  @change="handleQuickPeriodChange"
                  allow-clear
                >
                  <a-select-option value="today">Hari Ini</a-select-option>
                  <a-select-option value="this_week">Minggu Ini</a-select-option>
                  <a-select-option value="this_month">Bulan Ini</a-select-option>
                  <a-select-option value="this_quarter">Kuartal Ini</a-select-option>
                  <a-select-option value="this_year">Tahun Ini</a-select-option>
                  <a-select-option value="last_month">Bulan Lalu</a-select-option>
                  <a-select-option value="last_quarter">Kuartal Lalu</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Opsi Laporan">
                <a-checkbox-group v-model:value="reportOptions">
                  <a-checkbox value="budget">Perbandingan Budget</a-checkbox>
                  <a-checkbox value="assets">Ringkasan Aset</a-checkbox>
                  <a-checkbox value="trend">Tren Bulanan</a-checkbox>
                  <a-checkbox value="charts">Grafik & Chart</a-checkbox>
                </a-checkbox-group>
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- Report Content -->
        <div v-if="report">
          <!-- Report Header -->
          <a-card size="small">
            <a-descriptions title="Informasi Laporan" :column="4" bordered>
              <a-descriptions-item label="Periode">
                {{ report.report_period }}
              </a-descriptions-item>
              <a-descriptions-item label="Tanggal Generate">
                {{ formatDateTime(report.generated_at) }}
              </a-descriptions-item>
              <a-descriptions-item label="Rentang Tanggal">
                {{ formatDate(report.start_date) }} - {{ formatDate(report.end_date) }}
              </a-descriptions-item>
              <a-descriptions-item label="Status">
                <a-tag color="green">Berhasil</a-tag>
              </a-descriptions-item>
            </a-descriptions>
          </a-card>

          <!-- Cash Flow Summary -->
          <a-card title="Ringkasan Arus Kas" size="small">
            <a-row :gutter="16">
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Total Pemasukan"
                    :value="report.cash_flow_summary?.total_income || 0"
                    :precision="0"
                    :value-style="{ color: '#52c41a' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Total Pengeluaran"
                    :value="report.cash_flow_summary?.total_expense || 0"
                    :precision="0"
                    :value-style="{ color: '#ff4d4f' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Arus Kas Bersih"
                    :value="report.cash_flow_summary?.net_cash_flow || 0"
                    :precision="0"
                    :value-style="{ 
                      color: (report.cash_flow_summary?.net_cash_flow || 0) >= 0 ? '#52c41a' : '#ff4d4f' 
                    }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Saldo Akhir"
                    :value="report.cash_flow_summary?.ending_balance || 0"
                    :precision="0"
                    :value-style="{ color: '#1890ff' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
            </a-row>
          </a-card>

          <!-- Category Breakdown -->
          <a-card title="Breakdown Pengeluaran per Kategori" size="small">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-table
                  :columns="categoryColumns"
                  :data-source="report.category_breakdown || []"
                  :pagination="false"
                  size="small"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'amount'">
                      {{ formatCurrency(record.amount) }}
                    </template>
                    <template v-else-if="column.key === 'percentage'">
                      {{ record.percentage.toFixed(1) }}%
                    </template>
                    <template v-else-if="column.key === 'category'">
                      <a-tag :color="getCategoryColor(record.category)">
                        {{ getCategoryLabel(record.category) }}
                      </a-tag>
                    </template>
                  </template>
                </a-table>
              </a-col>
              <a-col :span="12">
                <div ref="categoryChartRef" style="height: 300px;"></div>
              </a-col>
            </a-row>
          </a-card>

          <!-- Budget Comparison -->
          <a-card 
            v-if="report.budget_comparison && reportOptions.includes('budget')" 
            title="Perbandingan Budget vs Aktual" 
            size="small"
          >
            <a-row :gutter="16" style="margin-bottom: 16px;">
              <a-col :span="8">
                <a-card size="small">
                  <a-statistic
                    title="Total Budget"
                    :value="report.budget_comparison.total_budget || 0"
                    :precision="0"
                    :value-style="{ color: '#1890ff' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="8">
                <a-card size="small">
                  <a-statistic
                    title="Total Aktual"
                    :value="report.budget_comparison.total_actual || 0"
                    :precision="0"
                    :value-style="{ color: '#722ed1' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="8">
                <a-card size="small">
                  <a-statistic
                    title="Variance"
                    :value="report.budget_comparison.variance || 0"
                    :precision="0"
                    :value-style="{ 
                      color: (report.budget_comparison.variance || 0) >= 0 ? '#52c41a' : '#ff4d4f' 
                    }"
                    suffix="IDR"
                  />
                  <div style="margin-top: 8px; font-size: 12px; color: #666;">
                    {{ (report.budget_comparison.variance_percent || 0).toFixed(1) }}% dari budget
                  </div>
                </a-card>
              </a-col>
            </a-row>

            <a-table
              :columns="budgetColumns"
              :data-source="report.budget_comparison.categories || []"
              :pagination="false"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'budget' || column.key === 'actual' || column.key === 'variance'">
                  {{ formatCurrency(record[column.key]) }}
                </template>
                <template v-else-if="column.key === 'variance_percent'">
                  <span :style="{ color: record.variance >= 0 ? '#52c41a' : '#ff4d4f' }">
                    {{ record.variance_percent.toFixed(1) }}%
                  </span>
                </template>
                <template v-else-if="column.key === 'category'">
                  <a-tag :color="getCategoryColor(record.category)">
                    {{ getCategoryLabel(record.category) }}
                  </a-tag>
                </template>
              </template>
            </a-table>
          </a-card>

          <!-- Asset Summary -->
          <a-card 
            v-if="report.asset_summary && reportOptions.includes('assets')" 
            title="Ringkasan Aset" 
            size="small"
          >
            <a-row :gutter="16">
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Total Aset"
                    :value="report.asset_summary.total_assets || 0"
                    :value-style="{ color: '#1890ff' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Nilai Pembelian"
                    :value="report.asset_summary.total_purchase_value || 0"
                    :precision="0"
                    :value-style="{ color: '#52c41a' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Nilai Buku Saat Ini"
                    :value="report.asset_summary.total_current_value || 0"
                    :precision="0"
                    :value-style="{ color: '#722ed1' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Total Depresiasi"
                    :value="report.asset_summary.total_depreciation || 0"
                    :precision="0"
                    :value-style="{ color: '#ff4d4f' }"
                    suffix="IDR"
                  />
                </a-card>
              </a-col>
            </a-row>
          </a-card>

          <!-- Monthly Trend -->
          <a-card 
            v-if="report.monthly_trend && reportOptions.includes('trend')" 
            title="Tren Arus Kas Bulanan" 
            size="small"
          >
            <div ref="trendChartRef" style="height: 400px;"></div>
          </a-card>
        </div>

        <!-- Empty State -->
        <a-empty 
          v-else-if="!loading"
          description="Pilih rentang tanggal dan klik 'Generate Laporan' untuk melihat laporan keuangan"
        >
          <a-button type="primary" @click="generateReport" :disabled="!dateRange || dateRange.length !== 2">
            Generate Laporan
          </a-button>
        </a-empty>
      </a-space>
    </a-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { message } from 'ant-design-vue'
import { 
  BarChartOutlined, 
  FilePdfOutlined, 
  FileExcelOutlined 
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import * as echarts from 'echarts'
import financialReportService from '@/services/financialReportService'

const loading = ref(false)
const exportingPDF = ref(false)
const exportingExcel = ref(false)
const report = ref(null)
const dateRange = ref([])
const quickPeriod = ref(undefined)
const reportOptions = ref(['budget', 'assets', 'trend', 'charts'])

// Chart refs
const categoryChartRef = ref()
const trendChartRef = ref()
let categoryChart = null
let trendChart = null

const categoryColumns = [
  {
    title: 'Kategori',
    key: 'category',
    width: 120
  },
  {
    title: 'Jumlah',
    key: 'amount',
    width: 120
  },
  {
    title: 'Persentase',
    key: 'percentage',
    width: 100
  },
  {
    title: 'Transaksi',
    dataIndex: 'count',
    key: 'count',
    width: 80
  }
]

const budgetColumns = [
  {
    title: 'Kategori',
    key: 'category',
    width: 120
  },
  {
    title: 'Budget',
    key: 'budget',
    width: 120
  },
  {
    title: 'Aktual',
    key: 'actual',
    width: 120
  },
  {
    title: 'Variance',
    key: 'variance',
    width: 120
  },
  {
    title: 'Variance %',
    key: 'variance_percent',
    width: 100
  }
]

const generateReport = async () => {
  if (!dateRange.value || dateRange.value.length !== 2) {
    message.warning('Pilih rentang tanggal terlebih dahulu')
    return
  }

  loading.value = true
  try {
    const startDate = dateRange.value[0].format('YYYY-MM-DD')
    const endDate = dateRange.value[1].format('YYYY-MM-DD')
    
    const options = {
      includeBudget: reportOptions.value.includes('budget'),
      includeAssets: reportOptions.value.includes('assets'),
      includeTrend: reportOptions.value.includes('trend')
    }

    const response = await financialReportService.generateFinancialReport(
      startDate, 
      endDate, 
      options
    )
    
    report.value = response.report
    
    // Generate charts after report is loaded
    if (reportOptions.value.includes('charts')) {
      await nextTick()
      generateCharts()
    }
    
    message.success('Laporan berhasil dibuat')
  } catch (error) {
    message.error('Gagal membuat laporan')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const generateCharts = () => {
  // Category breakdown pie chart
  if (categoryChartRef.value && report.value?.category_breakdown) {
    if (categoryChart) {
      categoryChart.dispose()
    }
    
    categoryChart = echarts.init(categoryChartRef.value)
    
    const categoryData = report.value.category_breakdown.map(item => ({
      name: getCategoryLabel(item.category),
      value: item.amount
    }))

    const categoryOption = {
      title: {
        text: 'Breakdown Pengeluaran',
        left: 'center'
      },
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)'
      },
      legend: {
        orient: 'vertical',
        left: 'left'
      },
      series: [
        {
          name: 'Pengeluaran',
          type: 'pie',
          radius: '50%',
          data: categoryData,
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)'
            }
          }
        }
      ]
    }
    
    categoryChart.setOption(categoryOption)
  }

  // Monthly trend line chart
  if (trendChartRef.value && report.value?.monthly_trend) {
    if (trendChart) {
      trendChart.dispose()
    }
    
    trendChart = echarts.init(trendChartRef.value)
    
    const months = report.value.monthly_trend.map(item => `${item.month} ${item.year}`)
    const incomeData = report.value.monthly_trend.map(item => item.income)
    const expenseData = report.value.monthly_trend.map(item => item.expense)
    const netCashFlowData = report.value.monthly_trend.map(item => item.net_cash_flow)

    const trendOption = {
      title: {
        text: 'Tren Arus Kas Bulanan',
        left: 'center'
      },
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['Pemasukan', 'Pengeluaran', 'Arus Kas Bersih'],
        top: 30
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: months
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
          name: 'Pemasukan',
          type: 'line',
          data: incomeData,
          itemStyle: { color: '#52c41a' }
        },
        {
          name: 'Pengeluaran',
          type: 'line',
          data: expenseData,
          itemStyle: { color: '#ff4d4f' }
        },
        {
          name: 'Arus Kas Bersih',
          type: 'line',
          data: netCashFlowData,
          itemStyle: { color: '#1890ff' }
        }
      ]
    }
    
    trendChart.setOption(trendOption)
  }
}

const handleQuickPeriodChange = (value) => {
  const today = dayjs()
  
  switch (value) {
    case 'today':
      dateRange.value = [today, today]
      break
    case 'this_week':
      dateRange.value = [today.startOf('week'), today.endOf('week')]
      break
    case 'this_month':
      dateRange.value = [today.startOf('month'), today.endOf('month')]
      break
    case 'this_quarter':
      dateRange.value = [today.startOf('quarter'), today.endOf('quarter')]
      break
    case 'this_year':
      dateRange.value = [today.startOf('year'), today.endOf('year')]
      break
    case 'last_month':
      const lastMonth = today.subtract(1, 'month')
      dateRange.value = [lastMonth.startOf('month'), lastMonth.endOf('month')]
      break
    case 'last_quarter':
      const lastQuarter = today.subtract(1, 'quarter')
      dateRange.value = [lastQuarter.startOf('quarter'), lastQuarter.endOf('quarter')]
      break
  }
}

const exportPDF = async () => {
  if (!report.value) {
    message.warning('Generate laporan terlebih dahulu')
    return
  }

  exportingPDF.value = true
  try {
    const startDate = dateRange.value[0].format('YYYY-MM-DD')
    const endDate = dateRange.value[1].format('YYYY-MM-DD')
    
    const options = {
      includeBudget: reportOptions.value.includes('budget'),
      includeAssets: reportOptions.value.includes('assets'),
      includeTrend: reportOptions.value.includes('trend'),
      includeCharts: reportOptions.value.includes('charts')
    }

    const response = await financialReportService.exportFinancialReport(
      startDate, 
      endDate, 
      'pdf', 
      options
    )
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan-keuangan-${startDate}-${endDate}.pdf`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    
    message.success('Laporan PDF berhasil diexport')
  } catch (error) {
    message.error('Gagal mengexport laporan PDF')
    console.error(error)
  } finally {
    exportingPDF.value = false
  }
}

const exportExcel = async () => {
  if (!report.value) {
    message.warning('Generate laporan terlebih dahulu')
    return
  }

  exportingExcel.value = true
  try {
    const startDate = dateRange.value[0].format('YYYY-MM-DD')
    const endDate = dateRange.value[1].format('YYYY-MM-DD')
    
    const options = {
      includeBudget: reportOptions.value.includes('budget'),
      includeAssets: reportOptions.value.includes('assets'),
      includeTrend: reportOptions.value.includes('trend'),
      includeCharts: reportOptions.value.includes('charts')
    }

    const response = await financialReportService.exportFinancialReport(
      startDate, 
      endDate, 
      'excel', 
      options
    )
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan-keuangan-${startDate}-${endDate}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    
    message.success('Laporan Excel berhasil diexport')
  } catch (error) {
    message.error('Gagal mengexport laporan Excel')
    console.error(error)
  } finally {
    exportingExcel.value = false
  }
}

const getCategoryColor = (category) => {
  const colors = {
    bahan_baku: 'blue',
    gaji: 'green',
    utilitas: 'orange',
    operasional: 'purple'
  }
  return colors[category] || 'default'
}

const getCategoryLabel = (category) => {
  const labels = {
    bahan_baku: 'Bahan Baku',
    gaji: 'Gaji',
    utilitas: 'Utilitas',
    operasional: 'Operasional'
  }
  return labels[category] || category
}

const formatCurrency = (value) => {
  if (!value) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(value)
}

const formatDate = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY')
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY HH:mm')
}

// Set default date range to current month
const setDefaultDateRange = () => {
  const startOfMonth = dayjs().startOf('month')
  const endOfMonth = dayjs().endOf('month')
  dateRange.value = [startOfMonth, endOfMonth]
  quickPeriod.value = 'this_month'
}

// Cleanup charts on unmount
const cleanup = () => {
  if (categoryChart) {
    categoryChart.dispose()
    categoryChart = null
  }
  if (trendChart) {
    trendChart.dispose()
    trendChart = null
  }
}

onMounted(() => {
  setDefaultDateRange()
  
  // Cleanup on unmount
  return cleanup
})
</script>

<style scoped>
.financial-report {
  padding: 24px;
}

.ant-statistic-content {
  font-size: 16px;
}

.ant-card-small > .ant-card-head {
  min-height: 38px;
  padding: 0 12px;
  font-size: 14px;
}

.ant-card-small > .ant-card-body {
  padding: 12px;
}
</style>