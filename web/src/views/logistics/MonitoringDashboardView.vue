<template>
  <div class="monitoring-dashboard-view">
    <a-page-header
      title="Monitoring Pengiriman"
      sub-title="Pantau status pengiriman menu dan ompreng"
    >
      <template #extra>
        <a-space>
          <a-date-picker
            v-model:value="selectedDate"
            format="DD/MM/YYYY"
            :disabled-date="disabledDate"
            @change="handleDateChange"
          />
          <a-button @click="refreshData" :loading="loading">
            <template #icon><reload-outlined /></template>
            Refresh
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <div class="content-wrapper">
      <!-- Summary Statistics Cards -->
      <a-row :gutter="[16, 16]" style="margin-bottom: 24px">
        <a-col :xs="24" :sm="12" :md="6">
          <a-card>
            <a-statistic
              title="Total Pengiriman"
              :value="summary.total_deliveries"
              :loading="loadingSummary"
            >
              <template #prefix>
                <car-outlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :md="6">
          <a-card>
            <a-statistic
              title="Selesai Diterima"
              :value="summary.completed_deliveries"
              :loading="loadingSummary"
              :value-style="{ color: '#52c41a' }"
            >
              <template #prefix>
                <check-circle-outlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :md="6">
          <a-card>
            <a-statistic
              title="Dalam Proses"
              :value="summary.in_progress_deliveries"
              :loading="loadingSummary"
              :value-style="{ color: '#1890ff' }"
            >
              <template #prefix>
                <sync-outlined :spin="true" />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :md="6">
          <a-card>
            <a-statistic
              title="Ompreng Dicuci"
              :value="summary.cleaned_ompreng"
              :loading="loadingSummary"
            >
              <template #prefix>
                <experiment-outlined />
              </template>
            </a-statistic>
          </a-card>
        </a-col>
      </a-row>

      <!-- Filters -->
      <a-card style="margin-bottom: 16px">
        <a-row :gutter="16">
          <a-col :xs="24" :sm="8">
            <a-select
              v-model:value="filters.school_id"
              placeholder="Filter Sekolah"
              style="width: 100%"
              allow-clear
              show-search
              :filter-option="filterOption"
              @change="applyFilters"
            >
              <a-select-option
                v-for="school in schools"
                :key="school.id"
                :value="school.id"
              >
                {{ school.name }}
              </a-select-option>
            </a-select>
          </a-col>
          <a-col :xs="24" :sm="8">
            <a-select
              v-model:value="filters.status"
              placeholder="Filter Status"
              style="width: 100%"
              allow-clear
              @change="applyFilters"
            >
              <a-select-option value="sedang_dimasak">Sedang Dimasak</a-select-option>
              <a-select-option value="selesai_dimasak">Selesai Dimasak</a-select-option>
              <a-select-option value="siap_dipacking">Siap Dipacking</a-select-option>
              <a-select-option value="selesai_dipacking">Selesai Dipacking</a-select-option>
              <a-select-option value="siap_dikirim">Siap Dikirim</a-select-option>
              <a-select-option value="diperjalanan">Diperjalanan</a-select-option>
              <a-select-option value="sudah_sampai_sekolah">Sudah Sampai Sekolah</a-select-option>
              <a-select-option value="sudah_diterima_pihak_sekolah">Sudah Diterima</a-select-option>
              <a-select-option value="ompreng_proses_pencucian">Ompreng Dicuci</a-select-option>
              <a-select-option value="ompreng_selesai_dicuci">Ompreng Selesai Dicuci</a-select-option>
            </a-select>
          </a-col>
          <a-col :xs="24" :sm="8">
            <a-select
              v-model:value="filters.driver_id"
              placeholder="Filter Driver"
              style="width: 100%"
              allow-clear
              show-search
              :filter-option="filterOption"
              @change="applyFilters"
            >
              <a-select-option
                v-for="driver in drivers"
                :key="driver.id"
                :value="driver.id"
              >
                {{ driver.full_name }}
              </a-select-option>
            </a-select>
          </a-col>
        </a-row>
      </a-card>

      <!-- Delivery Records Table -->
      <a-card title="Daftar Pengiriman">
        <a-table
          :columns="columns"
          :data-source="filteredRecords"
          :loading="loading"
          :pagination="pagination"
          :row-key="record => record.id"
          @change="handleTableChange"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'school'">
              {{ record.school?.name || '-' }}
            </template>
            <template v-else-if="column.key === 'driver'">
              {{ record.driver?.full_name || '-' }}
            </template>
            <template v-else-if="column.key === 'status'">
              <a-badge :status="getStatusBadgeType(record.current_status)" />
              <span>{{ getStatusText(record.current_status) }}</span>
            </template>
            <template v-else-if="column.key === 'portions'">
              {{ record.portions }} porsi
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-button type="link" @click="viewDetail(record.id)">
                Detail
              </a-button>
            </template>
          </template>
        </a-table>
      </a-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  ReloadOutlined,
  CarOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  ExperimentOutlined
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getDeliveryRecords, getDailySummary } from '@/services/monitoringService'

const router = useRouter()

// State
const selectedDate = ref(dayjs())
const loading = ref(false)
const loadingSummary = ref(false)
const deliveryRecords = ref([])
const summary = ref({
  total_deliveries: 0,
  completed_deliveries: 0,
  in_progress_deliveries: 0,
  cleaned_ompreng: 0
})
const schools = ref([])
const drivers = ref([])

const filters = reactive({
  school_id: undefined,
  status: undefined,
  driver_id: undefined
})

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => `Total ${total} pengiriman`
})

// Table columns
const columns = [
  {
    title: 'Sekolah',
    key: 'school',
    dataIndex: 'school'
  },
  {
    title: 'Driver',
    key: 'driver',
    dataIndex: 'driver'
  },
  {
    title: 'Status',
    key: 'status',
    dataIndex: 'current_status'
  },
  {
    title: 'Porsi',
    key: 'portions',
    dataIndex: 'portions',
    align: 'center'
  },
  {
    title: 'Aksi',
    key: 'actions',
    align: 'center',
    width: 100
  }
]

// Computed
const filteredRecords = computed(() => {
  let records = deliveryRecords.value

  if (filters.school_id) {
    records = records.filter(r => r.school_id === filters.school_id)
  }

  if (filters.status) {
    records = records.filter(r => r.current_status === filters.status)
  }

  if (filters.driver_id) {
    records = records.filter(r => r.driver_id === filters.driver_id)
  }

  pagination.total = records.length
  return records
})

// Methods
const disabledDate = (current) => {
  // Can select any date
  return false
}

const handleDateChange = () => {
  fetchData()
}

const refreshData = () => {
  fetchData()
}

const fetchData = async () => {
  await Promise.all([
    fetchDeliveryRecords(),
    fetchDailySummary()
  ])
}

const fetchDeliveryRecords = async () => {
  loading.value = true
  try {
    const dateStr = selectedDate.value.format('YYYY-MM-DD')
    const response = await getDeliveryRecords(dateStr)
    
    if (response.success) {
      deliveryRecords.value = response.data || []
      extractFiltersData()
    } else {
      message.error(response.message || 'Gagal memuat data pengiriman')
    }
  } catch (error) {
    console.error('Error fetching delivery records:', error)
    message.error(error.response?.data?.message || 'Gagal memuat data pengiriman')
  } finally {
    loading.value = false
  }
}

const fetchDailySummary = async () => {
  loadingSummary.value = true
  try {
    const dateStr = selectedDate.value.format('YYYY-MM-DD')
    const response = await getDailySummary(dateStr)
    
    if (response.success) {
      summary.value = response.data || {
        total_deliveries: 0,
        completed_deliveries: 0,
        in_progress_deliveries: 0,
        cleaned_ompreng: 0
      }
    } else {
      message.error(response.message || 'Gagal memuat ringkasan')
    }
  } catch (error) {
    console.error('Error fetching summary:', error)
    message.error(error.response?.data?.message || 'Gagal memuat ringkasan')
  } finally {
    loadingSummary.value = false
  }
}

const extractFiltersData = () => {
  // Extract unique schools
  const schoolsMap = new Map()
  deliveryRecords.value.forEach(record => {
    if (record.school && !schoolsMap.has(record.school.id)) {
      schoolsMap.set(record.school.id, record.school)
    }
  })
  schools.value = Array.from(schoolsMap.values())

  // Extract unique drivers
  const driversMap = new Map()
  deliveryRecords.value.forEach(record => {
    if (record.driver && !driversMap.has(record.driver.id)) {
      driversMap.set(record.driver.id, record.driver)
    }
  })
  drivers.value = Array.from(driversMap.values())
}

const applyFilters = () => {
  pagination.current = 1
}

const filterOption = (input, option) => {
  return option.children[0].children.toLowerCase().indexOf(input.toLowerCase()) >= 0
}

const handleTableChange = (pag) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
}

const getStatusBadgeType = (status) => {
  const statusMap = {
    'sedang_dimasak': 'processing',
    'selesai_dimasak': 'success',
    'siap_dipacking': 'processing',
    'selesai_dipacking': 'success',
    'siap_dikirim': 'processing',
    'diperjalanan': 'processing',
    'sudah_sampai_sekolah': 'success',
    'sudah_diterima_pihak_sekolah': 'success',
    'driver_ditugaskan_mengambil_ompreng': 'processing',
    'driver_menuju_sekolah': 'processing',
    'driver_sampai_di_sekolah': 'success',
    'ompreng_telah_diambil': 'success',
    'ompreng_sampai_di_sppg': 'success',
    'ompreng_proses_pencucian': 'processing',
    'ompreng_selesai_dicuci': 'success'
  }
  return statusMap[status] || 'default'
}

const getStatusText = (status) => {
  const statusTexts = {
    'sedang_dimasak': 'Sedang Dimasak',
    'selesai_dimasak': 'Selesai Dimasak',
    'siap_dipacking': 'Siap Dipacking',
    'selesai_dipacking': 'Selesai Dipacking',
    'siap_dikirim': 'Siap Dikirim',
    'diperjalanan': 'Diperjalanan',
    'sudah_sampai_sekolah': 'Sudah Sampai Sekolah',
    'sudah_diterima_pihak_sekolah': 'Sudah Diterima',
    'driver_ditugaskan_mengambil_ompreng': 'Driver Ditugaskan',
    'driver_menuju_sekolah': 'Driver Menuju Sekolah',
    'driver_sampai_di_sekolah': 'Driver Sampai',
    'ompreng_telah_diambil': 'Ompreng Diambil',
    'ompreng_sampai_di_sppg': 'Ompreng Sampai SPPG',
    'ompreng_proses_pencucian': 'Ompreng Dicuci',
    'ompreng_selesai_dicuci': 'Ompreng Selesai Dicuci'
  }
  return statusTexts[status] || status
}

const viewDetail = (id) => {
  router.push(`/logistics/monitoring/deliveries/${id}`)
}

// Lifecycle
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.monitoring-dashboard-view {
  padding: 24px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.content-wrapper {
  margin-top: 16px;
}

:deep(.ant-statistic-title) {
  font-size: 14px;
  margin-bottom: 8px;
}

:deep(.ant-statistic-content) {
  font-size: 24px;
  font-weight: 600;
}

:deep(.ant-badge-status-dot) {
  width: 8px;
  height: 8px;
}

:deep(.ant-badge-status-text) {
  margin-left: 8px;
}
</style>
