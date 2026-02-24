<template>
  <div class="asset-list">
    <a-page-header
      title="Manajemen Aset Dapur"
      sub-title="Kelola inventaris alat masak dan aset dapur"
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
            Tambah Aset
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="8">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari nama aset atau kode aset..."
              @search="handleSearch"
              allow-clear
            />
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterCategory"
              placeholder="Kategori"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="Peralatan Masak">Peralatan Masak</a-select-option>
              <a-select-option value="Peralatan Packing">Peralatan Packing</a-select-option>
              <a-select-option value="Elektronik">Elektronik</a-select-option>
              <a-select-option value="Furniture">Furniture</a-select-option>
              <a-select-option value="Kendaraan">Kendaraan</a-select-option>
              <a-select-option value="Lainnya">Lainnya</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterCondition"
              placeholder="Kondisi"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="good">Baik</a-select-option>
              <a-select-option value="fair">Cukup</a-select-option>
              <a-select-option value="poor">Buruk</a-select-option>
            </a-select>
          </a-col>
        </a-row>

        <!-- Statistics Cards -->
        <a-row :gutter="16" v-if="report">
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Aset"
                :value="report.total_assets || 0"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Nilai Pembelian"
                :value="report.total_purchase_value || 0"
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
                :value="report.total_current_value || 0"
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
                :value="report.total_depreciation || 0"
                :precision="0"
                :value-style="{ color: '#ff4d4f' }"
                suffix="IDR"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="assets"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'purchase_price'">
              {{ formatCurrency(record.purchase_price) }}
            </template>
            <template v-else-if="column.key === 'current_value'">
              {{ formatCurrency(record.current_value) }}
            </template>
            <template v-else-if="column.key === 'depreciation'">
              {{ formatCurrency(record.purchase_price - record.current_value) }}
            </template>
            <template v-else-if="column.key === 'condition'">
              <a-tag :color="getConditionColor(record.condition)">
                {{ getConditionLabel(record.condition) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'purchase_date'">
              {{ formatDate(record.purchase_date) }}
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewAsset(record)">
                  Detail
                </a-button>
                <a-button type="link" size="small" @click="editAsset(record)">
                  Edit
                </a-button>
                <a-button type="link" size="small" @click="showMaintenanceModal(record)">
                  Maintenance
                </a-button>
                <a-popconfirm
                  title="Yakin ingin menghapus aset ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="deleteAsset(record.id)"
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
    <AssetFormModal
      v-model:visible="modalVisible"
      :asset="editingAsset"
      @success="handleFormSuccess"
    />

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Aset"
      :footer="null"
      width="900px"
    >
      <a-descriptions v-if="selectedAsset" bordered :column="2">
        <a-descriptions-item label="Kode Aset">
          {{ selectedAsset.asset_code }}
        </a-descriptions-item>
        <a-descriptions-item label="Nama Aset">
          {{ selectedAsset.name }}
        </a-descriptions-item>
        <a-descriptions-item label="Kategori">
          {{ selectedAsset.category }}
        </a-descriptions-item>
        <a-descriptions-item label="Kondisi">
          <a-tag :color="getConditionColor(selectedAsset.condition)">
            {{ getConditionLabel(selectedAsset.condition) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal Pembelian">
          {{ formatDate(selectedAsset.purchase_date) }}
        </a-descriptions-item>
        <a-descriptions-item label="Harga Pembelian">
          {{ formatCurrency(selectedAsset.purchase_price) }}
        </a-descriptions-item>
        <a-descriptions-item label="Nilai Buku Saat Ini">
          {{ formatCurrency(selectedAsset.current_value) }}
        </a-descriptions-item>
        <a-descriptions-item label="Akumulasi Depresiasi">
          {{ formatCurrency(selectedAsset.purchase_price - selectedAsset.current_value) }}
        </a-descriptions-item>
        <a-descriptions-item label="Tingkat Depresiasi">
          {{ selectedAsset.depreciation_rate }}% per tahun
        </a-descriptions-item>
        <a-descriptions-item label="Lokasi">
          {{ selectedAsset.location || '-' }}
        </a-descriptions-item>
      </a-descriptions>

      <a-divider>Riwayat Maintenance</a-divider>

      <a-table
        :columns="maintenanceColumns"
        :data-source="selectedAsset.maintenance_records || []"
        :pagination="{ pageSize: 5 }"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'cost'">
            {{ formatCurrency(record.cost) }}
          </template>
          <template v-else-if="column.key === 'maintenance_date'">
            {{ formatDate(record.maintenance_date) }}
          </template>
        </template>
      </a-table>
    </a-modal>

    <!-- Maintenance Modal -->
    <a-modal
      v-model:open="maintenanceModalVisible"
      title="Tambah Catatan Maintenance"
      :confirm-loading="submittingMaintenance"
      @ok="handleMaintenanceSubmit"
      @cancel="handleMaintenanceCancel"
      width="600px"
    >
      <a-form
        ref="maintenanceFormRef"
        :model="maintenanceFormData"
        :rules="maintenanceRules"
        layout="vertical"
      >
        <a-form-item label="Tanggal Maintenance" name="maintenance_date">
          <a-date-picker 
            v-model:value="maintenanceFormData.maintenance_date" 
            style="width: 100%" 
            placeholder="Pilih tanggal maintenance"
            format="DD/MM/YYYY"
          />
        </a-form-item>

        <a-form-item label="Deskripsi" name="description">
          <a-textarea 
            v-model:value="maintenanceFormData.description" 
            :rows="3" 
            placeholder="Deskripsi kegiatan maintenance"
          />
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Biaya" name="cost">
              <a-input-number
                v-model:value="maintenanceFormData.cost"
                style="width: 100%"
                :min="0"
                :precision="0"
                placeholder="0"
                :formatter="value => `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
                :parser="value => value.replace(/Rp\s?|(,*)/g, '')"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Dilakukan Oleh" name="performed_by">
              <a-input v-model:value="maintenanceFormData.performed_by" placeholder="Nama teknisi/vendor" />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, ExportOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import assetService from '@/services/assetService'
import AssetFormModal from '@/components/AssetFormModal.vue'

const loading = ref(false)
const exporting = ref(false)
const submittingMaintenance = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const maintenanceModalVisible = ref(false)
const editingAsset = ref(null)
const selectedAsset = ref(null)
const maintenanceAsset = ref(null)
const assets = ref([])
const report = ref(null)
const searchText = ref('')
const filterCategory = ref(undefined)
const filterCondition = ref(undefined)
const maintenanceFormRef = ref()

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const maintenanceFormData = reactive({
  maintenance_date: null,
  description: '',
  cost: 0,
  performed_by: ''
})

const maintenanceRules = {
  maintenance_date: [{ required: true, message: 'Tanggal maintenance wajib diisi' }],
  description: [{ required: true, message: 'Deskripsi wajib diisi' }],
  cost: [{ required: true, message: 'Biaya wajib diisi' }],
  performed_by: [{ required: true, message: 'Pelaksana wajib diisi' }]
}

const columns = [
  {
    title: 'Kode Aset',
    dataIndex: 'asset_code',
    key: 'asset_code',
    width: 120
  },
  {
    title: 'Nama Aset',
    dataIndex: 'name',
    key: 'name',
    sorter: true
  },
  {
    title: 'Kategori',
    dataIndex: 'category',
    key: 'category'
  },
  {
    title: 'Harga Pembelian',
    key: 'purchase_price',
    width: 140,
    sorter: true
  },
  {
    title: 'Nilai Buku',
    key: 'current_value',
    width: 120,
    sorter: true
  },
  {
    title: 'Depresiasi',
    key: 'depreciation',
    width: 120
  },
  {
    title: 'Kondisi',
    key: 'condition',
    width: 100
  },
  {
    title: 'Tanggal Beli',
    key: 'purchase_date',
    width: 120
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 250
  }
]

const maintenanceColumns = [
  {
    title: 'Tanggal',
    key: 'maintenance_date',
    width: 120
  },
  {
    title: 'Deskripsi',
    dataIndex: 'description',
    key: 'description'
  },
  {
    title: 'Biaya',
    key: 'cost',
    width: 120
  },
  {
    title: 'Dilakukan Oleh',
    dataIndex: 'performed_by',
    key: 'performed_by'
  }
]

const fetchAssets = async () => {
  loading.value = true
  try {
    const params = {
      q: searchText.value || undefined,
      category: filterCategory.value || undefined,
      condition: filterCondition.value || undefined
    }
    const response = await assetService.getAssets(params)
    assets.value = response.assets || []
  } catch (error) {
    message.error('Gagal memuat data aset')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchReport = async () => {
  try {
    const response = await assetService.getAssetReport()
    report.value = response.report || {}
  } catch (error) {
    console.error('Gagal memuat laporan aset:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchAssets()
}

const handleSearch = () => {
  pagination.current = 1
  fetchAssets()
}

const showCreateModal = () => {
  editingAsset.value = null
  modalVisible.value = true
}

const editAsset = (asset) => {
  editingAsset.value = asset
  modalVisible.value = true
}

const viewAsset = async (asset) => {
  try {
    const response = await assetService.getAssetById(asset.id)
    selectedAsset.value = response.asset
    detailModalVisible.value = true
  } catch (error) {
    message.error('Gagal memuat detail aset')
    console.error(error)
  }
}

const showMaintenanceModal = (asset) => {
  maintenanceAsset.value = asset
  resetMaintenanceForm()
  maintenanceModalVisible.value = true
}

const handleFormSuccess = () => {
  modalVisible.value = false
  fetchAssets()
  fetchReport()
}

const deleteAsset = async (id) => {
  try {
    await assetService.deleteAsset(id)
    message.success('Aset berhasil dihapus')
    fetchAssets()
    fetchReport()
  } catch (error) {
    message.error('Gagal menghapus aset')
    console.error(error)
  }
}

const handleMaintenanceSubmit = async () => {
  try {
    await maintenanceFormRef.value.validate()
    submittingMaintenance.value = true

    const submitData = {
      ...maintenanceFormData,
      maintenance_date: maintenanceFormData.maintenance_date ? 
        maintenanceFormData.maintenance_date.format('YYYY-MM-DD') : null
    }

    await assetService.addMaintenanceRecord(maintenanceAsset.value.id, submitData)
    message.success('Catatan maintenance berhasil ditambahkan')
    
    maintenanceModalVisible.value = false
    fetchAssets()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    
    const errorMessage = error.response?.data?.message || 'Gagal menyimpan catatan maintenance'
    message.error(errorMessage)
    console.error(error)
  } finally {
    submittingMaintenance.value = false
  }
}

const handleMaintenanceCancel = () => {
  maintenanceModalVisible.value = false
  resetMaintenanceForm()
}

const resetMaintenanceForm = () => {
  Object.assign(maintenanceFormData, {
    maintenance_date: null,
    description: '',
    cost: 0,
    performed_by: ''
  })
  maintenanceFormRef.value?.resetFields()
}

const exportReport = async () => {
  exporting.value = true
  try {
    const response = await assetService.exportAssetReport('excel')
    
    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `laporan-aset-${dayjs().format('YYYY-MM-DD')}.xlsx`)
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

const getConditionColor = (condition) => {
  const colors = {
    good: 'green',
    fair: 'orange',
    poor: 'red'
  }
  return colors[condition] || 'default'
}

const getConditionLabel = (condition) => {
  const labels = {
    good: 'Baik',
    fair: 'Cukup',
    poor: 'Buruk'
  }
  return labels[condition] || condition
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

onMounted(() => {
  fetchAssets()
  fetchReport()
})
</script>

<style scoped>
.asset-list {
  padding: 24px;
}
</style>