<template>
  <a-space direction="vertical" style="width: 100%" :size="16">
    <!-- Search and Filter Controls -->
    <a-row :gutter="16">
      <a-col :span="8">
        <a-input-search
          v-model:value="searchText"
          placeholder="Cari nomor form atau pembuat..."
          @search="handleSearch"
          allow-clear
        />
      </a-col>
      <a-col :span="5">
        <a-select
          v-model:value="filterStatus"
          placeholder="Status"
          style="width: 100%"
          @change="handleSearch"
          allow-clear
        >
          <a-select-option value="pending">Pending</a-select-option>
          <a-select-option value="approved">Disetujui</a-select-option>
          <a-select-option value="rejected">Ditolak</a-select-option>
        </a-select>
      </a-col>
      <a-col :span="7">
        <a-range-picker
          v-model:value="dateRange"
          style="width: 100%"
          format="DD/MM/YYYY"
          @change="handleSearch"
        />
      </a-col>
      <a-col :span="4">
        <a-button type="primary" @click="createNewForm" block>
          <template #icon><PlusOutlined /></template>
          Buat Form Baru
        </a-button>
      </a-col>
    </a-row>

    <!-- Table -->
    <a-table
      :columns="columns"
      :data-source="forms"
      :loading="loading"
      :pagination="{
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: pagination.total,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} form`
      }"
      @change="handleTableChange"
      row-key="id"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'form_number'">
          <a-tag color="blue">{{ record.form_number }}</a-tag>
        </template>
        <template v-else-if="column.key === 'created_at'">
          {{ formatDate(record.created_at) }}
        </template>
        <template v-else-if="column.key === 'creator'">
          {{ record.creator?.full_name || '-' }}
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'approver'">
          {{ record.approver?.full_name || '-' }}
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-space>
            <a-button type="link" size="small" @click="viewForm(record)">
              Lihat
            </a-button>
            <a-button
              v-if="record.status === 'pending'"
              type="link"
              size="small"
              @click="editForm(record)"
            >
              Edit
            </a-button>
            <a-button
              v-if="record.status === 'pending'"
              type="link"
              size="small"
              danger
              @click="confirmDelete(record)"
            >
              Hapus
            </a-button>
            <a-button type="link" size="small" @click="exportForm(record)">
              Export
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>
  </a-space>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import { PlusOutlined } from '@ant-design/icons-vue'
import stokOpnameService from '@/services/stokOpnameService'
import { debounce } from 'lodash-es'

const router = useRouter()

// State
const loading = ref(false)
const forms = ref([])
const searchText = ref('')
const filterStatus = ref(undefined)
const dateRange = ref(null)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0
})

// Table columns
const columns = [
  {
    title: 'Nomor Form',
    key: 'form_number',
    width: 150
  },
  {
    title: 'Tanggal',
    key: 'created_at',
    width: 120
  },
  {
    title: 'Pembuat',
    key: 'creator',
    width: 150
  },
  {
    title: 'Status',
    key: 'status',
    width: 120
  },
  {
    title: 'Penyetuju',
    key: 'approver',
    width: 150
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 250
  }
]

// Methods
const fetchForms = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: searchText.value || undefined,
      status: filterStatus.value
    }
    
    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0].format('YYYY-MM-DD')
      params.end_date = dateRange.value[1].format('YYYY-MM-DD')
    }
    
    console.log('Fetching forms with params:', params)
    const response = await stokOpnameService.getForms(params)
    console.log('Forms response:', response.data)
    
    // Backend returns { success, data, pagination }
    forms.value = response.data.data || []
    pagination.total = response.data.pagination?.total_count || 0
    
    console.log('Forms loaded:', forms.value.length)
  } catch (error) {
    message.error('Gagal memuat data stok opname: ' + (error.response?.data?.message || error.message))
    console.error('Fetch forms error:', error)
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchForms()
}

const handleSearch = debounce(() => {
  pagination.current = 1
  fetchForms()
}, 300)

const createNewForm = () => {
  router.push('/inventory/stok-opname/create')
}

const viewForm = (record) => {
  router.push(`/inventory/stok-opname/${record.id}`)
}

const editForm = (record) => {
  router.push(`/inventory/stok-opname/${record.id}/edit`)
}

const confirmDelete = (record) => {
  Modal.confirm({
    title: 'Konfirmasi Hapus',
    content: `Apakah Anda yakin ingin menghapus form ${record.form_number}?`,
    okText: 'Ya, Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      await deleteForm(record)
    }
  })
}

const deleteForm = async (record) => {
  try {
    await stokOpnameService.deleteForm(record.id)
    message.success('Form berhasil dihapus')
    fetchForms()
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal menghapus form')
    console.error(error)
  }
}

const exportForm = async (record) => {
  try {
    const response = await stokOpnameService.exportForm(record.id, 'excel')
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `stok-opname-${record.form_number}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    message.success('Form berhasil diekspor')
  } catch (error) {
    message.error('Gagal mengekspor form')
    console.error(error)
  }
}

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    approved: 'green',
    rejected: 'red'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    pending: 'Pending',
    approved: 'Disetujui',
    rejected: 'Ditolak'
  }
  return texts[status] || status
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

// Lifecycle
onMounted(() => {
  fetchForms()
})
</script>

<style scoped>
/* Add any custom styles here */
</style>
