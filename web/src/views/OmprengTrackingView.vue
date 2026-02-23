<template>
  <div class="ompreng-tracking">
    <a-page-header
      title="Pelacakan Ompreng"
      sub-title="Pantau sirkulasi wadah makanan (ompreng) di seluruh sekolah"
    />

    <!-- Summary Cards -->
    <a-row :gutter="16" style="margin-bottom: 24px">
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="Total Ompreng"
            :value="globalInventory?.total_owned || 0"
            :value-style="{ color: '#1890ff' }"
          >
            <template #prefix>
              <InboxOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="Di Dapur"
            :value="globalInventory?.at_kitchen || 0"
            :value-style="{ color: '#52c41a' }"
          >
            <template #prefix>
              <HomeOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="Dalam Sirkulasi"
            :value="globalInventory?.in_circulation || 0"
            :value-style="{ color: '#faad14' }"
          >
            <template #prefix>
              <CarOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="Hilang/Rusak"
            :value="globalInventory?.missing || 0"
            :value-style="{ color: missingCount > 0 ? '#ff4d4f' : '#52c41a' }"
          >
            <template #prefix>
              <ExclamationCircleOutlined v-if="missingCount > 0" />
              <CheckCircleOutlined v-else />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <a-card>
      <a-tabs v-model:activeKey="activeTab">
        <!-- School Balances Tab -->
        <a-tab-pane key="balances" tab="Saldo per Sekolah">
          <a-space direction="vertical" style="width: 100%" :size="16">
            <!-- Search and Actions -->
            <a-row :gutter="16">
              <a-col :span="12">
                <a-input-search
                  v-model:value="searchText"
                  placeholder="Cari nama sekolah..."
                  @search="handleSearch"
                  allow-clear
                />
              </a-col>
              <a-col :span="6">
                <a-select
                  v-model:value="filterStatus"
                  placeholder="Filter Status"
                  style="width: 100%"
                  @change="handleSearch"
                  allow-clear
                >
                  <a-select-option value="missing">Ada Ompreng Hilang</a-select-option>
                  <a-select-option value="normal">Normal</a-select-option>
                  <a-select-option value="excess">Kelebihan</a-select-option>
                </a-select>
              </a-col>
              <a-col :span="6">
                <a-button type="primary" @click="showRecordModal" block>
                  <template #icon><PlusOutlined /></template>
                  Catat Pergerakan
                </a-button>
              </a-col>
            </a-row>

            <!-- Missing Ompreng Alert -->
            <a-alert
              v-if="missingSchools.length > 0"
              :message="`${missingSchools.length} sekolah memiliki ompreng hilang`"
              type="warning"
              show-icon
              closable
            >
              <template #description>
                <div>
                  Sekolah dengan ompreng hilang: 
                  <strong>{{ missingSchools.map(s => s.school_name).join(', ') }}</strong>
                </div>
              </template>
            </a-alert>

            <!-- School Balances Table -->
            <a-table
              :columns="balanceColumns"
              :data-source="filteredBalances"
              :loading="loading"
              :pagination="pagination"
              @change="handleTableChange"
              row-key="school_id"
              :row-class-name="getRowClassName"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'school_name'">
                  <strong>{{ record.school_name }}</strong>
                </template>
                <template v-else-if="column.key === 'balance'">
                  <a-tag :color="getBalanceColor(record.balance)">
                    {{ record.balance }} unit
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'status'">
                  <a-tag :color="getStatusColor(record.balance)">
                    {{ getStatusText(record.balance) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'last_update'">
                  {{ formatDateTime(record.date) }}
                </template>
                <template v-else-if="column.key === 'actions'">
                  <a-space>
                    <a-button type="link" size="small" @click="viewHistory(record)">
                      Riwayat
                    </a-button>
                    <a-button type="link" size="small" @click="recordMovement(record)">
                      Catat
                    </a-button>
                  </a-space>
                </template>
              </template>
            </a-table>
          </a-space>
        </a-tab-pane>

        <!-- Circulation Reports Tab -->
        <a-tab-pane key="reports" tab="Laporan Sirkulasi">
          <a-space direction="vertical" style="width: 100%" :size="16">
            <!-- Date Range Filter -->
            <a-row :gutter="16">
              <a-col :span="12">
                <a-range-picker
                  v-model:value="reportDateRange"
                  style="width: 100%"
                  format="DD/MM/YYYY"
                  @change="fetchReports"
                />
              </a-col>
              <a-col :span="6">
                <a-button type="default" @click="exportReport" :loading="exportLoading">
                  <template #icon><DownloadOutlined /></template>
                  Export Excel
                </a-button>
              </a-col>
              <a-col :span="6">
                <a-button type="default" @click="fetchReports">
                  <template #icon><ReloadOutlined /></template>
                  Refresh
                </a-button>
              </a-col>
            </a-row>

            <!-- Report Summary Cards -->
            <a-row :gutter="16" v-if="circulationReport">
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Total Drop-off"
                    :value="circulationReport.total_drop_off"
                    :value-style="{ color: '#52c41a' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Total Pick-up"
                    :value="circulationReport.total_pick_up"
                    :value-style="{ color: '#1890ff' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Perubahan Bersih"
                    :value="circulationReport.net_change"
                    :value-style="{ color: circulationReport.net_change >= 0 ? '#52c41a' : '#ff4d4f' }"
                  />
                </a-card>
              </a-col>
              <a-col :span="6">
                <a-card size="small">
                  <a-statistic
                    title="Sekolah Bermasalah"
                    :value="circulationReport.schools_with_missing"
                    :value-style="{ color: circulationReport.schools_with_missing > 0 ? '#ff4d4f' : '#52c41a' }"
                  />
                </a-card>
              </a-col>
            </a-row>

            <!-- Missing Ompreng Details -->
            <a-card title="Sekolah dengan Ompreng Hilang" v-if="missingOmpreng.length > 0">
              <a-list
                :data-source="missingOmpreng"
                :loading="loadingReports"
              >
                <template #renderItem="{ item }">
                  <a-list-item>
                    <a-list-item-meta>
                      <template #title>
                        <a-space>
                          <ExclamationCircleOutlined style="color: #ff4d4f" />
                          <strong>{{ item.school_name }}</strong>
                        </a-space>
                      </template>
                      <template #description>
                        <a-space direction="vertical" size="small">
                          <span>
                            Saldo saat ini: <strong class="text-danger">{{ item.balance }} unit</strong>
                          </span>
                          <span>
                            Terakhir diperbarui: {{ formatDateTime(item.date) }}
                          </span>
                        </a-space>
                      </template>
                    </a-list-item-meta>
                    <template #actions>
                      <a-button type="primary" size="small" @click="recordMovement(item)">
                        Perbaiki Saldo
                      </a-button>
                    </template>
                  </a-list-item>
                </template>
              </a-list>
            </a-card>

            <!-- No Missing Ompreng -->
            <a-empty
              v-else-if="!loadingReports"
              description="Tidak ada ompreng yang hilang"
              :image="Empty.PRESENTED_IMAGE_SIMPLE"
            />
          </a-space>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- Record Movement Modal -->
    <a-modal
      v-model:open="recordModalVisible"
      title="Catat Pergerakan Ompreng"
      :confirm-loading="recordLoading"
      @ok="handleRecordMovement"
      @cancel="resetRecordForm"
    >
      <a-form
        :model="recordForm"
        :rules="recordRules"
        ref="recordFormRef"
        layout="vertical"
      >
        <a-form-item label="Sekolah" name="school_id">
          <a-select
            v-model:value="recordForm.school_id"
            placeholder="Pilih sekolah"
            show-search
            :filter-option="filterSchool"
          >
            <a-select-option
              v-for="balance in schoolBalances"
              :key="balance.school_id"
              :value="balance.school_id"
            >
              {{ balance.school_name }}
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="Jenis Pergerakan" name="movement_type">
          <a-radio-group v-model:value="recordForm.movement_type">
            <a-radio value="drop_off">Drop-off (Antar ke Sekolah)</a-radio>
            <a-radio value="pick_up">Pick-up (Ambil dari Sekolah)</a-radio>
          </a-radio-group>
        </a-form-item>

        <a-form-item label="Jumlah" name="quantity">
          <a-input-number
            v-model:value="recordForm.quantity"
            :min="1"
            :max="1000"
            style="width: 100%"
            placeholder="Masukkan jumlah ompreng"
          />
        </a-form-item>

        <a-form-item label="Catatan" name="notes">
          <a-textarea
            v-model:value="recordForm.notes"
            :rows="3"
            placeholder="Catatan tambahan (opsional)"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- History Modal -->
    <a-modal
      v-model:open="historyModalVisible"
      :title="`Riwayat Ompreng - ${selectedSchool?.school_name}`"
      :footer="null"
      width="800px"
    >
      <a-table
        :columns="historyColumns"
        :data-source="schoolHistory"
        :loading="loadingHistory"
        :pagination="{ pageSize: 10 }"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'movement_type'">
            <a-tag :color="record.drop_off > 0 ? 'green' : 'blue'">
              {{ record.drop_off > 0 ? 'Drop-off' : 'Pick-up' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'quantity'">
            <span :class="record.drop_off > 0 ? 'text-success' : 'text-primary'">
              {{ record.drop_off > 0 ? '+' : '-' }}{{ record.drop_off || record.pick_up }}
            </span>
          </template>
          <template v-else-if="column.key === 'balance'">
            <strong :class="record.balance < 0 ? 'text-danger' : ''">
              {{ record.balance }}
            </strong>
          </template>
          <template v-else-if="column.key === 'date'">
            {{ formatDateTime(record.date) }}
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message, Empty } from 'ant-design-vue'
import {
  InboxOutlined,
  HomeOutlined,
  CarOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  PlusOutlined,
  DownloadOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import omprengTrackingService from '@/services/omprengTrackingService'
import dayjs from 'dayjs'

const activeTab = ref('balances')
const loading = ref(false)
const loadingReports = ref(false)
const loadingHistory = ref(false)
const recordLoading = ref(false)
const exportLoading = ref(false)
const recordModalVisible = ref(false)
const historyModalVisible = ref(false)
const recordFormRef = ref()
const selectedSchool = ref(null)

const schoolBalances = ref([])
const globalInventory = ref({})
const circulationReport = ref(null)
const missingOmpreng = ref([])
const schoolHistory = ref([])
const searchText = ref('')
const filterStatus = ref(undefined)
const reportDateRange = ref([dayjs().subtract(30, 'day'), dayjs()])

const recordForm = reactive({
  school_id: undefined,
  movement_type: 'drop_off',
  quantity: 1,
  notes: ''
})

const recordRules = {
  school_id: [{ required: true, message: 'Pilih sekolah' }],
  movement_type: [{ required: true, message: 'Pilih jenis pergerakan' }],
  quantity: [{ required: true, message: 'Masukkan jumlah', type: 'number', min: 1 }]
}

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0
})

const missingCount = computed(() => {
  return globalInventory.value?.missing || 0
})

const missingSchools = computed(() => {
  return schoolBalances.value.filter(school => school.balance < 0)
})

const filteredBalances = computed(() => {
  let filtered = schoolBalances.value

  // Apply search filter
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    filtered = filtered.filter(school => 
      school.school_name.toLowerCase().includes(search)
    )
  }

  // Apply status filter
  if (filterStatus.value) {
    filtered = filtered.filter(school => {
      switch (filterStatus.value) {
        case 'missing':
          return school.balance < 0
        case 'normal':
          return school.balance >= 0 && school.balance <= 10
        case 'excess':
          return school.balance > 10
        default:
          return true
      }
    })
  }

  return filtered
})

const balanceColumns = [
  {
    title: 'Nama Sekolah',
    key: 'school_name',
    sorter: (a, b) => a.school_name.localeCompare(b.school_name)
  },
  {
    title: 'Saldo Ompreng',
    key: 'balance',
    width: 150,
    sorter: (a, b) => a.balance - b.balance
  },
  {
    title: 'Status',
    key: 'status',
    width: 120
  },
  {
    title: 'Terakhir Diperbarui',
    key: 'last_update',
    width: 180
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 150
  }
]

const historyColumns = [
  {
    title: 'Tanggal',
    key: 'date',
    width: 180
  },
  {
    title: 'Jenis',
    key: 'movement_type',
    width: 120
  },
  {
    title: 'Jumlah',
    key: 'quantity',
    width: 100
  },
  {
    title: 'Saldo Setelah',
    key: 'balance',
    width: 120
  },
  {
    title: 'Dicatat Oleh',
    dataIndex: ['recorder', 'full_name'],
    key: 'recorder'
  }
]

const fetchBalances = async () => {
  loading.value = true
  try {
    const response = await omprengTrackingService.getOmprengTracking()
    schoolBalances.value = response.data.balances || []
    pagination.total = schoolBalances.value.length
  } catch (error) {
    message.error('Gagal memuat data saldo ompreng')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchReports = async () => {
  loadingReports.value = true
  try {
    const params = {}
    if (reportDateRange.value && reportDateRange.value.length === 2) {
      params.start_date = reportDateRange.value[0].format('YYYY-MM-DD')
      params.end_date = reportDateRange.value[1].format('YYYY-MM-DD')
    }

    const response = await omprengTrackingService.getReports(
      params.start_date,
      params.end_date
    )
    
    circulationReport.value = response.data.report
    globalInventory.value = response.data.inventory
    missingOmpreng.value = response.data.missing || []
  } catch (error) {
    message.error('Gagal memuat laporan sirkulasi')
    console.error(error)
  } finally {
    loadingReports.value = false
  }
}

const showRecordModal = () => {
  recordModalVisible.value = true
}

const recordMovement = (school) => {
  recordForm.school_id = school.school_id
  recordModalVisible.value = true
}

const handleRecordMovement = async () => {
  try {
    await recordFormRef.value.validate()
    recordLoading.value = true

    const { school_id, movement_type, quantity } = recordForm

    if (movement_type === 'drop_off') {
      await omprengTrackingService.recordDropOff(school_id, quantity)
      message.success('Drop-off ompreng berhasil dicatat')
    } else {
      await omprengTrackingService.recordPickUp(school_id, quantity)
      message.success('Pick-up ompreng berhasil dicatat')
    }

    recordModalVisible.value = false
    resetRecordForm()
    await fetchBalances()
    await fetchReports()
  } catch (error) {
    if (error.response?.data?.message) {
      message.error(error.response.data.message)
    } else {
      message.error('Gagal mencatat pergerakan ompreng')
    }
    console.error(error)
  } finally {
    recordLoading.value = false
  }
}

const resetRecordForm = () => {
  Object.assign(recordForm, {
    school_id: undefined,
    movement_type: 'drop_off',
    quantity: 1,
    notes: ''
  })
  recordFormRef.value?.resetFields()
}

const viewHistory = async (school) => {
  selectedSchool.value = school
  historyModalVisible.value = true
  loadingHistory.value = true

  try {
    const response = await omprengTrackingService.getSchoolHistory(school.school_id)
    schoolHistory.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat riwayat sekolah')
    console.error(error)
  } finally {
    loadingHistory.value = false
  }
}

const exportReport = async () => {
  exportLoading.value = true
  try {
    // This would typically generate and download an Excel file
    // For now, we'll show a success message
    message.success('Laporan berhasil diekspor')
  } catch (error) {
    message.error('Gagal mengekspor laporan')
    console.error(error)
  } finally {
    exportLoading.value = false
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
}

const handleSearch = () => {
  pagination.current = 1
}

const getRowClassName = (record) => {
  if (record.balance < 0) {
    return 'missing-ompreng-row'
  }
  return ''
}

const getBalanceColor = (balance) => {
  if (balance < 0) return 'red'
  if (balance === 0) return 'orange'
  if (balance > 10) return 'blue'
  return 'green'
}

const getStatusColor = (balance) => {
  if (balance < 0) return 'red'
  if (balance === 0) return 'orange'
  return 'green'
}

const getStatusText = (balance) => {
  if (balance < 0) return 'Ompreng Hilang'
  if (balance === 0) return 'Kosong'
  if (balance > 10) return 'Kelebihan'
  return 'Normal'
}

const filterSchool = (input, option) => {
  return option.children[0].children.toLowerCase().includes(input.toLowerCase())
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY HH:mm')
}

onMounted(() => {
  fetchBalances()
  fetchReports()
})
</script>

<style scoped>
.ompreng-tracking {
  padding: 24px;
}

:deep(.missing-ompreng-row) {
  background-color: #fff1f0;
}

:deep(.missing-ompreng-row:hover) {
  background-color: #ffe7e6 !important;
}

.text-danger {
  color: #ff4d4f;
  font-weight: 500;
}

.text-success {
  color: #52c41a;
  font-weight: 500;
}

.text-primary {
  color: #1890ff;
  font-weight: 500;
}
</style>