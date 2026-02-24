<template>
  <div class="cash-flow-list">
    <a-page-header
      title="Manajemen Arus Kas"
      sub-title="Kelola pencatatan arus kas operasional"
    >
      <template #extra>
        <a-space>
          <a-button @click="exportReport" :loading="exporting">
            <template #icon>
              <ExportOutlined />
            </template>
            Export Laporan
          </a-button>
          <a-button type="primary" @click="showCreateModal">
            <template #icon>
              <PlusOutlined />
            </template>
            Tambah Transaksi
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Filters -->
        <a-row :gutter="16">
          <a-col :span="6">
            <a-range-picker
              v-model:value="dateRange"
              style="width: 100%"
              placeholder="['Tanggal Mulai', 'Tanggal Akhir']"
              format="DD/MM/YYYY"
              @change="handleDateRangeChange"
            />
          </a-col>
          <a-col :span="5">
            <a-select
              v-model:value="filterCategory"
              placeholder="Kategori"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="bahan_baku">Bahan Baku</a-select-option>
              <a-select-option value="gaji">Gaji</a-select-option>
              <a-select-option value="utilitas">Utilitas</a-select-option>
              <a-select-option value="operasional">Operasional</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="4">
            <a-select
              v-model:value="filterType"
              placeholder="Tipe"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="income">Pemasukan</a-select-option>
              <a-select-option value="expense">Pengeluaran</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari deskripsi atau referensi..."
              @search="handleSearch"
              allow-clear
            />
          </a-col>
          <a-col :span="3">
            <a-button @click="resetFilters" style="width: 100%">
              Reset Filter
            </a-button>
          </a-col>
        </a-row>

        <!-- Summary Cards -->
        <a-row :gutter="16" v-if="summary">
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Pemasukan"
                :value="summary.total_income || 0"
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
                :value="summary.total_expense || 0"
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
                :value="summary.net_cash_flow || 0"
                :precision="0"
                :value-style="{ color: summary.net_cash_flow >= 0 ? '#52c41a' : '#ff4d4f' }"
                suffix="IDR"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Transaksi"
                :value="cashFlowEntries.length || 0"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Running Balance by Category -->
        <a-card size="small" title="Saldo Berjalan per Kategori" v-if="runningBalances.length > 0">
          <a-row :gutter="16">
            <a-col :span="4" v-for="balance in runningBalances" :key="balance.category">
              <a-statistic
                :title="getCategoryLabel(balance.category)"
                :value="balance.balance"
                :precision="0"
                :value-style="{ color: balance.balance >= 0 ? '#52c41a' : '#ff4d4f', fontSize: '14px' }"
                suffix="IDR"
              />
            </a-col>
          </a-row>
        </a-card>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="cashFlowEntries"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'amount'">
              <span :style="{ color: record.type === 'income' ? '#52c41a' : '#ff4d4f' }">
                {{ record.type === 'income' ? '+' : '-' }}{{ formatCurrency(record.amount) }}
              </span>
            </template>
            <template v-else-if="column.key === 'category'">
              <a-tag :color="getCategoryColor(record.category)">
                {{ getCategoryLabel(record.category) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'type'">
              <a-tag :color="record.type === 'income' ? 'green' : 'red'">
                {{ record.type === 'income' ? 'Pemasukan' : 'Pengeluaran' }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'date'">
              {{ formatDate(record.date) }}
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewCashFlow(record)">
                  Detail
                </a-button>
                <!-- Edit and Delete functionality disabled - backend endpoints not implemented yet -->
                <!-- <a-button type="link" size="small" @click="editCashFlow(record)">
                  Edit
                </a-button>
                <a-popconfirm
                  title="Yakin ingin menghapus transaksi ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="deleteCashFlow(record.id)"
                >
                  <a-button type="link" size="small" danger>
                    Hapus
                  </a-button>
                </a-popconfirm> -->
              </a-space>
            </template>
          </template>
        </a-table>
      </a-space>
    </a-card>

    <!-- Create/Edit Modal -->
    <CashFlowFormModal
      v-model:visible="modalVisible"
      :cash-flow="editingCashFlow"
      @success="handleFormSuccess"
    />

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Transaksi"
      :footer="null"
      width="600px"
    >
      <a-descriptions v-if="selectedCashFlow" bordered :column="2">
        <a-descriptions-item label="ID Transaksi">
          {{ selectedCashFlow.transaction_id }}
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal">
          {{ formatDate(selectedCashFlow.date) }}
        </a-descriptions-item>
        <a-descriptions-item label="Kategori">
          <a-tag :color="getCategoryColor(selectedCashFlow.category)">
            {{ getCategoryLabel(selectedCashFlow.category) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Tipe">
          <a-tag :color="selectedCashFlow.type === 'income' ? 'green' : 'red'">
            {{ selectedCashFlow.type === 'income' ? 'Pemasukan' : 'Pengeluaran' }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Jumlah" :span="2">
          <span :style="{ color: selectedCashFlow.type === 'income' ? '#52c41a' : '#ff4d4f', fontSize: '18px', fontWeight: 'bold' }">
            {{ selectedCashFlow.type === 'income' ? '+' : '-' }}{{ formatCurrency(selectedCashFlow.amount) }}
          </span>
        </a-descriptions-item>
        <a-descriptions-item label="Deskripsi" :span="2">
          {{ selectedCashFlow.description || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="Referensi" :span="2">
          {{ selectedCashFlow.reference || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="Dibuat Oleh">
          {{ selectedCashFlow.creator?.full_name || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal Dibuat">
          {{ formatDateTime(selectedCashFlow.created_at) }}
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, ExportOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import cashFlowService from '@/services/cashFlowService'
import CashFlowFormModal from '@/components/CashFlowFormModal.vue'

const loading = ref(false)
const exporting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const editingCashFlow = ref(null)
const selectedCashFlow = ref(null)
const cashFlowEntries = ref([])
const summary = ref(null)
const runningBalances = ref([])
const searchText = ref('')
const filterCategory = ref(undefined)
const filterType = ref(undefined)
const dateRange = ref([])

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const columns = [
  {
    title: 'ID Transaksi',
    dataIndex: 'transaction_id',
    key: 'transaction_id',
    width: 140
  },
  {
    title: 'Tanggal',
    key: 'date',
    width: 120,
    sorter: true
  },
  {
    title: 'Kategori',
    key: 'category',
    width: 120
  },
  {
    title: 'Tipe',
    key: 'type',
    width: 100
  },
  {
    title: 'Jumlah',
    key: 'amount',
    width: 150,
    sorter: true
  },
  {
    title: 'Deskripsi',
    dataIndex: 'description',
    key: 'description',
    ellipsis: true
  },
  {
    title: 'Referensi',
    dataIndex: 'reference',
    key: 'reference',
    width: 120
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 100
  }
]

const categories = ['bahan_baku', 'gaji', 'utilitas', 'operasional']

const fetchCashFlowEntries = async () => {
  loading.value = true
  try {
    const params = {
      category: filterCategory.value || undefined,
      type: filterType.value || undefined
    }

    // Add date range if selected
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0].format('YYYY-MM-DD')
      params.end_date = dateRange.value[1].format('YYYY-MM-DD')
    }

    const response = await cashFlowService.getCashFlowEntries(params)
    cashFlowEntries.value = response.cash_flows || []
    
    // Filter by search text locally if needed
    if (searchText.value) {
      cashFlowEntries.value = cashFlowEntries.value.filter(entry => 
        entry.description?.toLowerCase().includes(searchText.value.toLowerCase()) ||
        entry.reference?.toLowerCase().includes(searchText.value.toLowerCase())
      )
    }
  } catch (error) {
    message.error('Gagal memuat data arus kas')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchSummary = async () => {
  if (!dateRange.value || dateRange.value.length !== 2) return
  
  try {
    const startDate = dateRange.value[0].format('YYYY-MM-DD')
    const endDate = dateRange.value[1].format('YYYY-MM-DD')
    
    const response = await cashFlowService.getCashFlowSummary(startDate, endDate)
    summary.value = response.summary || {}
  } catch (error) {
    console.error('Gagal memuat ringkasan arus kas:', error)
  }
}

const fetchRunningBalances = async () => {
  const today = dayjs().format('YYYY-MM-DD')
  const balances = []
  
  try {
    for (const category of categories) {
      const response = await cashFlowService.getRunningBalance(category, today)
      balances.push({
        category: category,
        balance: response.balance || 0
      })
    }
    runningBalances.value = balances
  } catch (error) {
    console.error('Gagal memuat saldo berjalan:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchCashFlowEntries()
}

const handleSearch = () => {
  pagination.current = 1
  fetchCashFlowEntries()
}

const handleDateRangeChange = () => {
  fetchCashFlowEntries()
  fetchSummary()
}

const resetFilters = () => {
  searchText.value = ''
  filterCategory.value = undefined
  filterType.value = undefined
  dateRange.value = []
  summary.value = null
  pagination.current = 1
  fetchCashFlowEntries()
}

const showCreateModal = () => {
  editingCashFlow.value = null
  modalVisible.value = true
}

const editCashFlow = (cashFlow) => {
  // Edit functionality disabled - backend endpoint not implemented yet
  message.info('Fitur edit akan tersedia setelah backend API diimplementasikan')
  // editingCashFlow.value = cashFlow
  // modalVisible.value = true
}

const viewCashFlow = (cashFlow) => {
  selectedCashFlow.value = cashFlow
  detailModalVisible.value = true
}

const handleFormSuccess = () => {
  modalVisible.value = false
  fetchCashFlowEntries()
  fetchSummary()
  fetchRunningBalances()
}

const deleteCashFlow = async (id) => {
  // Delete functionality disabled - backend endpoint not implemented yet
  message.info('Fitur hapus akan tersedia setelah backend API diimplementasikan')
  // try {
  //   await cashFlowService.deleteCashFlow(id)
  //   message.success('Transaksi berhasil dihapus')
  //   fetchCashFlowEntries()
  //   fetchSummary()
  //   fetchRunningBalances()
  // } catch (error) {
  //   message.error('Gagal menghapus transaksi')
  //   console.error(error)
  // }
}

const exportReport = async () => {
  if (!dateRange.value || dateRange.value.length !== 2) {
    message.warning('Pilih rentang tanggal terlebih dahulu')
    return
  }

  exporting.value = true
  try {
    const startDate = dateRange.value[0].format('YYYY-MM-DD')
    const endDate = dateRange.value[1].format('YYYY-MM-DD')
    
    const response = await cashFlowService.exportCashFlowReport(startDate, endDate, 'excel')
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan-arus-kas-${startDate}-${endDate}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    
    message.success('Laporan berhasil diexport')
  } catch (error) {
    message.error('Gagal mengexport laporan')
    console.error(error)
  } finally {
    exporting.value = false
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
}

onMounted(() => {
  setDefaultDateRange()
  fetchCashFlowEntries()
  fetchSummary()
  fetchRunningBalances()
})
</script>

<style scoped>
.cash-flow-list {
  padding: 24px;
}
</style>