<template>
  <div class="supplier-list">
    <a-page-header
      title="Manajemen Supplier"
      sub-title="Kelola data supplier dan lihat performa"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Tambah Supplier
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="12">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari nama supplier..."
              @search="handleSearch"
              allow-clear
            />
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterStatus"
              placeholder="Status"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="active">Aktif</a-select-option>
              <a-select-option value="inactive">Tidak Aktif</a-select-option>
            </a-select>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="suppliers"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'is_active'">
              <a-tag :color="record.is_active ? 'green' : 'red'">
                {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'on_time_delivery'">
              <a-progress
                :percent="Math.round(record.on_time_delivery || 0)"
                :status="record.on_time_delivery >= 80 ? 'success' : 'normal'"
                size="small"
              />
            </template>
            <template v-else-if="column.key === 'quality_rating'">
              <a-rate :value="record.quality_rating || 0" disabled allow-half />
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewSupplier(record)">
                  Detail
                </a-button>
                <a-button type="link" size="small" @click="editSupplier(record)">
                  Edit
                </a-button>
                <a-popconfirm
                  title="Yakin ingin menghapus supplier ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="deleteSupplier(record.id)"
                >
                  <a-button type="link" size="small" danger>
                    Hapus
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-space>
    </a-card>

    <!-- Create/Edit Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="editingSupplier ? 'Edit Supplier' : 'Tambah Supplier'"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="600px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
      >
        <a-form-item label="Nama Supplier" name="name">
          <a-input v-model:value="formData.name" placeholder="Masukkan nama supplier" />
        </a-form-item>

        <a-form-item label="Kategori Produk" name="product_category">
          <a-input v-model:value="formData.product_category" placeholder="Contoh: Sayuran, Daging, Bumbu" />
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Nama Kontak" name="contact_person">
              <a-input v-model:value="formData.contact_person" placeholder="Nama kontak person" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Nomor Telepon" name="phone_number">
              <a-input v-model:value="formData.phone_number" placeholder="08xxxxxxxxxx" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="Email" name="email">
          <a-input v-model:value="formData.email" type="email" placeholder="email@supplier.com" />
        </a-form-item>

        <a-form-item label="Alamat" name="address">
          <a-textarea v-model:value="formData.address" :rows="3" placeholder="Alamat lengkap supplier" />
        </a-form-item>

        <a-form-item label="Status" name="is_active">
          <a-switch v-model:checked="formData.is_active" checked-children="Aktif" un-checked-children="Tidak Aktif" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Supplier"
      :footer="null"
      width="800px"
    >
      <a-descriptions v-if="selectedSupplier" bordered :column="2">
        <a-descriptions-item label="Nama Supplier" :span="2">
          {{ selectedSupplier.name }}
        </a-descriptions-item>
        <a-descriptions-item label="Kategori Produk" :span="2">
          {{ selectedSupplier.product_category }}
        </a-descriptions-item>
        <a-descriptions-item label="Kontak Person">
          {{ selectedSupplier.contact_person }}
        </a-descriptions-item>
        <a-descriptions-item label="Telepon">
          {{ selectedSupplier.phone_number }}
        </a-descriptions-item>
        <a-descriptions-item label="Email" :span="2">
          {{ selectedSupplier.email }}
        </a-descriptions-item>
        <a-descriptions-item label="Alamat" :span="2">
          {{ selectedSupplier.address }}
        </a-descriptions-item>
        <a-descriptions-item label="Status">
          <a-tag :color="selectedSupplier.is_active ? 'green' : 'red'">
            {{ selectedSupplier.is_active ? 'Aktif' : 'Tidak Aktif' }}
          </a-tag>
        </a-descriptions-item>
      </a-descriptions>

      <a-divider>Metrik Performa</a-divider>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-statistic title="Pengiriman Tepat Waktu" :value="selectedSupplier.on_time_delivery || 0" suffix="%" />
        </a-col>
        <a-col :span="12">
          <a-statistic title="Rating Kualitas">
            <template #formatter>
              <a-rate :value="selectedSupplier.quality_rating || 0" disabled allow-half />
            </template>
          </a-statistic>
        </a-col>
      </a-row>

      <a-divider>Riwayat Transaksi</a-divider>

      <a-table
        :columns="transactionColumns"
        :data-source="transactionHistory"
        :loading="loadingTransactions"
        :pagination="{ pageSize: 5 }"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'amount'">
            {{ formatCurrency(record.amount) }}
          </template>
          <template v-else-if="column.key === 'order_date'">
            {{ formatDate(record.order_date) }}
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import supplierService from '@/services/supplierService'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const editingSupplier = ref(null)
const selectedSupplier = ref(null)
const suppliers = ref([])
const transactionHistory = ref([])
const loadingTransactions = ref(false)
const searchText = ref('')
const filterStatus = ref(undefined)
const formRef = ref()

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  name: '',
  product_category: '',
  contact_person: '',
  phone_number: '',
  email: '',
  address: '',
  is_active: true
})

const rules = {
  name: [{ required: true, message: 'Nama supplier wajib diisi' }],
  product_category: [{ required: true, message: 'Kategori produk wajib diisi' }],
  contact_person: [{ required: true, message: 'Nama kontak wajib diisi' }],
  phone_number: [{ required: true, message: 'Nomor telepon wajib diisi' }],
  email: [
    { required: true, message: 'Email wajib diisi' },
    { type: 'email', message: 'Format email tidak valid' }
  ]
}

const columns = [
  {
    title: 'Nama Supplier',
    dataIndex: 'name',
    key: 'name',
    sorter: true
  },
  {
    title: 'Kategori Produk',
    dataIndex: 'product_category',
    key: 'product_category'
  },
  {
    title: 'Kontak',
    dataIndex: 'contact_person',
    key: 'contact_person'
  },
  {
    title: 'Telepon',
    dataIndex: 'phone_number',
    key: 'phone_number'
  },
  {
    title: 'Pengiriman Tepat Waktu',
    key: 'on_time_delivery',
    width: 180
  },
  {
    title: 'Rating Kualitas',
    key: 'quality_rating',
    width: 150
  },
  {
    title: 'Status',
    key: 'is_active',
    width: 100
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 200
  }
]

const transactionColumns = [
  {
    title: 'Nomor PO',
    dataIndex: 'po_number',
    key: 'po_number'
  },
  {
    title: 'Tanggal',
    key: 'order_date'
  },
  {
    title: 'Jumlah',
    key: 'amount'
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status'
  }
]

const fetchSuppliers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: searchText.value || undefined,
      is_active: filterStatus.value
    }
    const response = await supplierService.getSuppliers(params)
    suppliers.value = response.data.data || []
    pagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat data supplier')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchSuppliers()
}

const handleSearch = () => {
  pagination.current = 1
  fetchSuppliers()
}

const showCreateModal = () => {
  editingSupplier.value = null
  resetForm()
  modalVisible.value = true
}

const editSupplier = (supplier) => {
  editingSupplier.value = supplier
  Object.assign(formData, {
    name: supplier.name,
    product_category: supplier.product_category,
    contact_person: supplier.contact_person,
    phone_number: supplier.phone_number,
    email: supplier.email,
    address: supplier.address,
    is_active: supplier.is_active
  })
  modalVisible.value = true
}

const viewSupplier = async (supplier) => {
  selectedSupplier.value = supplier
  detailModalVisible.value = true
  
  // Fetch transaction history
  loadingTransactions.value = true
  try {
    const response = await supplierService.getSupplierPerformance(supplier.id)
    transactionHistory.value = response.data.transactions || []
  } catch (error) {
    console.error('Gagal memuat riwayat transaksi:', error)
  } finally {
    loadingTransactions.value = false
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    if (editingSupplier.value) {
      await supplierService.updateSupplier(editingSupplier.value.id, formData)
      message.success('Supplier berhasil diperbarui')
    } else {
      await supplierService.createSupplier(formData)
      message.success('Supplier berhasil ditambahkan')
    }

    modalVisible.value = false
    fetchSuppliers()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    message.error('Gagal menyimpan data supplier')
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const deleteSupplier = async (id) => {
  try {
    await supplierService.deleteSupplier(id)
    message.success('Supplier berhasil dihapus')
    fetchSuppliers()
  } catch (error) {
    message.error('Gagal menghapus supplier')
    console.error(error)
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    name: '',
    product_category: '',
    contact_person: '',
    phone_number: '',
    email: '',
    address: '',
    is_active: true
  })
  formRef.value?.resetFields()
}

const formatCurrency = (value) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR'
  }).format(value)
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

onMounted(() => {
  fetchSuppliers()
})
</script>

<style scoped>
.supplier-list {
  padding: 24px;
}
</style>
