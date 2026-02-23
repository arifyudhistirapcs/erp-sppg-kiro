<template>
  <div class="attendance-report">
    <a-page-header
      title="Laporan Absensi"
      sub-title="Laporan kehadiran karyawan dengan filter tanggal dan karyawan"
    >
      <template #extra>
        <a-space>
          <a-button 
            type="primary" 
            :loading="exportingExcel"
            @click="exportToExcel"
            :disabled="!hasData"
          >
            <template #icon><FileExcelOutlined /></template>
            Export Excel
          </a-button>
          <a-button 
            :loading="exportingPDF"
            @click="exportToPDF"
            :disabled="!hasData"
          >
            <template #icon><FilePdfOutlined /></template>
            Export PDF
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Filter Section -->
        <a-card size="small" title="Filter Laporan">
          <a-form layout="inline" :model="filters" @finish="handleSearch">
            <a-form-item label="Periode">
              <a-range-picker
                v-model:value="dateRange"
                format="DD/MM/YYYY"
                placeholder="['Tanggal Mulai', 'Tanggal Akhir']"
                style="width: 280px"
              />
            </a-form-item>
            
            <a-form-item label="Karyawan">
              <a-select
                v-model:value="filters.employeeId"
                placeholder="Semua Karyawan"
                style="width: 200px"
                show-search
                option-filter-prop="children"
                allow-clear
                :filter-option="filterOption"
              >
                <a-select-option 
                  v-for="employee in employees" 
                  :key="employee.id" 
                  :value="employee.id"
                >
                  {{ employee.full_name }} ({{ employee.nik }})
                </a-select-option>
              </a-select>
            </a-form-item>

            <a-form-item>
              <a-button type="primary" html-type="submit" :loading="loading">
                <template #icon><SearchOutlined /></template>
                Cari
              </a-button>
            </a-form-item>

            <a-form-item>
              <a-button @click="resetFilters">
                <template #icon><ClearOutlined /></template>
                Reset
              </a-button>
            </a-form-item>
          </a-form>
        </a-card>

        <!-- Summary Statistics -->
        <a-row :gutter="16" v-if="hasData">
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Karyawan"
                :value="reportData.length"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Hari Kerja"
                :value="totalWorkDays"
                :value-style="{ color: '#52c41a' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Jam Kerja"
                :value="totalWorkHours"
                :precision="1"
                :value-style="{ color: '#722ed1' }"
                suffix="jam"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Rata-rata Jam/Hari"
                :value="averageHoursPerDay"
                :precision="1"
                :value-style="{ color: '#fa8c16' }"
                suffix="jam"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Report Table -->
        <a-table
          :columns="columns"
          :data-source="reportData"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="employee_id"
          :scroll="{ x: 800 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'full_name'">
              <a-button type="link" @click="viewDetailedReport(record)" style="padding: 0;">
                {{ record.full_name }}
              </a-button>
            </template>
            <template v-else-if="column.key === 'total_hours'">
              {{ formatHours(record.total_hours) }}
            </template>
            <template v-else-if="column.key === 'average_hours'">
              {{ formatHours(record.average_hours) }}
            </template>
            <template v-else-if="column.key === 'attendance_rate'">
              <a-progress 
                :percent="calculateAttendanceRate(record.total_days)" 
                size="small"
                :status="getAttendanceStatus(record.total_days)"
              />
            </template>
          </template>
        </a-table>

        <!-- Empty State -->
        <a-empty 
          v-if="!loading && !hasData" 
          description="Tidak ada data absensi untuk periode yang dipilih"
        >
          <a-button type="primary" @click="resetFilters">Reset Filter</a-button>
        </a-empty>
      </a-space>
    </a-card>

    <!-- Detailed Report Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      :title="`Detail Absensi - ${selectedEmployee?.full_name}`"
      width="900px"
      :footer="null"
    >
      <a-spin :spinning="loadingDetail">
        <a-table
          :columns="detailColumns"
          :data-source="detailData"
          :pagination="false"
          size="small"
          :scroll="{ y: 400 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'date'">
              {{ formatDate(record.date) }}
            </template>
            <template v-else-if="column.key === 'check_in'">
              {{ formatTime(record.check_in) }}
            </template>
            <template v-else-if="column.key === 'check_out'">
              {{ record.check_out ? formatTime(record.check_out) : '-' }}
            </template>
            <template v-else-if="column.key === 'work_hours'">
              {{ formatHours(record.work_hours) }}
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="getStatusColor(record)">
                {{ getStatusText(record) }}
              </a-tag>
            </template>
          </template>
        </a-table>
      </a-spin>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { 
  FileExcelOutlined, 
  FilePdfOutlined, 
  SearchOutlined, 
  ClearOutlined 
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import attendanceService from '@/services/attendanceService'
import employeeService from '@/services/employeeService'

const loading = ref(false)
const loadingDetail = ref(false)
const exportingExcel = ref(false)
const exportingPDF = ref(false)
const detailModalVisible = ref(false)
const reportData = ref([])
const detailData = ref([])
const employees = ref([])
const selectedEmployee = ref(null)
const dateRange = ref([])

const filters = reactive({
  employeeId: undefined
})

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total, range) => `${range[0]}-${range[1]} dari ${total} data`
})

const columns = [
  {
    title: 'Nama Karyawan',
    key: 'full_name',
    dataIndex: 'full_name',
    sorter: true,
    width: 200
  },
  {
    title: 'Posisi',
    dataIndex: 'position',
    key: 'position',
    width: 150
  },
  {
    title: 'Total Hari',
    dataIndex: 'total_days',
    key: 'total_days',
    sorter: true,
    width: 100,
    align: 'center'
  },
  {
    title: 'Total Jam',
    key: 'total_hours',
    sorter: true,
    width: 120,
    align: 'right'
  },
  {
    title: 'Rata-rata Jam/Hari',
    key: 'average_hours',
    sorter: true,
    width: 140,
    align: 'right'
  },
  {
    title: 'Tingkat Kehadiran',
    key: 'attendance_rate',
    width: 150,
    align: 'center'
  }
]

const detailColumns = [
  {
    title: 'Tanggal',
    key: 'date',
    width: 100
  },
  {
    title: 'Check In',
    key: 'check_in',
    width: 100
  },
  {
    title: 'Check Out',
    key: 'check_out',
    width: 100
  },
  {
    title: 'Jam Kerja',
    key: 'work_hours',
    width: 100,
    align: 'right'
  },
  {
    title: 'Status',
    key: 'status',
    width: 100,
    align: 'center'
  }
]

const hasData = computed(() => reportData.value.length > 0)

const totalWorkDays = computed(() => {
  return reportData.value.reduce((sum, item) => sum + item.total_days, 0)
})

const totalWorkHours = computed(() => {
  return reportData.value.reduce((sum, item) => sum + item.total_hours, 0)
})

const averageHoursPerDay = computed(() => {
  if (totalWorkDays.value === 0) return 0
  return totalWorkHours.value / totalWorkDays.value
})

const fetchEmployees = async () => {
  try {
    const response = await employeeService.getEmployees({ is_active: true })
    employees.value = response.data || []
  } catch (error) {
    console.error('Gagal memuat data karyawan:', error)
  }
}

const fetchReport = async () => {
  if (!dateRange.value || dateRange.value.length !== 2) {
    message.warning('Silakan pilih periode tanggal terlebih dahulu')
    return
  }

  loading.value = true
  try {
    const params = {
      start_date: dateRange.value[0].format('YYYY-MM-DD'),
      end_date: dateRange.value[1].format('YYYY-MM-DD')
    }

    if (filters.employeeId) {
      params.employee_id = filters.employeeId
    }

    const response = await attendanceService.getAttendanceReport(params)
    reportData.value = response.data || []
    pagination.total = reportData.value.length
  } catch (error) {
    message.error('Gagal memuat laporan absensi')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  fetchReport()
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
}

const resetFilters = () => {
  dateRange.value = []
  filters.employeeId = undefined
  reportData.value = []
  pagination.current = 1
}

const viewDetailedReport = async (record) => {
  if (!dateRange.value || dateRange.value.length !== 2) {
    message.warning('Periode tanggal tidak valid')
    return
  }

  selectedEmployee.value = record
  detailModalVisible.value = true
  loadingDetail.value = true

  try {
    const response = await attendanceService.getAttendanceByDateRange(
      record.employee_id,
      dateRange.value[0].format('YYYY-MM-DD'),
      dateRange.value[1].format('YYYY-MM-DD')
    )
    detailData.value = response.data || []
  } catch (error) {
    message.error('Gagal memuat detail absensi')
    console.error(error)
  } finally {
    loadingDetail.value = false
  }
}

const exportToExcel = async () => {
  if (!hasData.value) {
    message.warning('Tidak ada data untuk diekspor')
    return
  }

  exportingExcel.value = true
  try {
    const params = {
      start_date: dateRange.value[0].format('YYYY-MM-DD'),
      end_date: dateRange.value[1].format('YYYY-MM-DD'),
      format: 'excel'
    }

    if (filters.employeeId) {
      params.employee_id = filters.employeeId
    }

    const response = await attendanceService.exportToExcel(params)
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan-absensi-${params.start_date}-${params.end_date}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    message.success('Laporan Excel berhasil diunduh')
  } catch (error) {
    message.error('Gagal mengekspor ke Excel')
    console.error(error)
  } finally {
    exportingExcel.value = false
  }
}

const exportToPDF = async () => {
  if (!hasData.value) {
    message.warning('Tidak ada data untuk diekspor')
    return
  }

  exportingPDF.value = true
  try {
    const params = {
      start_date: dateRange.value[0].format('YYYY-MM-DD'),
      end_date: dateRange.value[1].format('YYYY-MM-DD'),
      format: 'pdf'
    }

    if (filters.employeeId) {
      params.employee_id = filters.employeeId
    }

    const response = await attendanceService.exportToPDF(params)
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan-absensi-${params.start_date}-${params.end_date}.pdf`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)

    message.success('Laporan PDF berhasil diunduh')
  } catch (error) {
    message.error('Gagal mengekspor ke PDF')
    console.error(error)
  } finally {
    exportingPDF.value = false
  }
}

const filterOption = (input, option) => {
  return option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

const formatHours = (hours) => {
  if (!hours) return '0.0 jam'
  return `${parseFloat(hours).toFixed(1)} jam`
}

const formatDate = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY')
}

const formatTime = (time) => {
  if (!time) return '-'
  return dayjs(time).format('HH:mm')
}

const calculateAttendanceRate = (totalDays) => {
  if (!dateRange.value || dateRange.value.length !== 2) return 0
  
  const workingDays = getWorkingDays(dateRange.value[0], dateRange.value[1])
  if (workingDays === 0) return 0
  
  return Math.round((totalDays / workingDays) * 100)
}

const getWorkingDays = (startDate, endDate) => {
  let count = 0
  let current = startDate.clone()
  
  while (current.isSameOrBefore(endDate)) {
    // Skip weekends (Saturday = 6, Sunday = 0)
    if (current.day() !== 0 && current.day() !== 6) {
      count++
    }
    current = current.add(1, 'day')
  }
  
  return count
}

const getAttendanceStatus = (totalDays) => {
  const rate = calculateAttendanceRate(totalDays)
  if (rate >= 90) return 'success'
  if (rate >= 75) return 'normal'
  return 'exception'
}

const getStatusColor = (record) => {
  if (!record.check_out) return 'orange'
  if (record.work_hours >= 8) return 'green'
  if (record.work_hours >= 6) return 'blue'
  return 'red'
}

const getStatusText = (record) => {
  if (!record.check_out) return 'Belum Check Out'
  if (record.work_hours >= 8) return 'Lengkap'
  if (record.work_hours >= 6) return 'Cukup'
  return 'Kurang'
}

onMounted(() => {
  fetchEmployees()
  
  // Set default date range to current month
  const now = dayjs()
  dateRange.value = [
    now.startOf('month'),
    now.endOf('month')
  ]
})
</script>

<style scoped>
.attendance-report {
  padding: 24px;
}

:deep(.ant-table-thead > tr > th) {
  background-color: #fafafa;
  font-weight: 600;
}

:deep(.ant-statistic-content) {
  font-size: 20px;
}

:deep(.ant-progress-line) {
  margin-right: 8px;
}
</style>