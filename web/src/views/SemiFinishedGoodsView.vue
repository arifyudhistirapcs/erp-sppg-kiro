<template>
  <div class="semi-finished-goods">
    <a-page-header
      title="Manajemen Komponen"
      sub-title="Kelola komponen menu (nasi, lauk, sambal)"
    >
      <template #extra>
        <a-space>
          <a-button @click="showInventoryModal">
            <template #icon><StockOutlined /></template>
            Lihat Stok
          </a-button>
          <a-button type="primary" @click="showCreateModal">
            <template #icon><PlusOutlined /></template>
            Tambah Barang
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-card>
      <!-- Search & Filter -->
      <a-row :gutter="16" class="mb-4">
        <a-col :span="8">
          <a-input-search
            v-model:value="searchText"
            placeholder="Cari nama barang..."
            @search="handleSearch"
            allow-clear
          />
        </a-col>
        <a-col :span="8">
          <a-select
            v-model:value="categoryFilter"
            placeholder="Filter kategori"
            allow-clear
            style="width: 100%"
            @change="handleSearch"
          >
            <a-select-option value="protein_hewani">Protein Hewani</a-select-option>
            <a-select-option value="sumber_lemak">Sumber Lemak</a-select-option>
            <a-select-option value="lainnya">Lainnya</a-select-option>
            <a-select-option value="sayur">Sayur</a-select-option>
            <a-select-option value="karbohidrat">Karbohidrat</a-select-option>
            <a-select-option value="buah">Buah</a-select-option>
            <a-select-option value="susu">Susu</a-select-option>
            <a-select-option value="protein_nabati">Protein Nabati</a-select-option>
          </a-select>
        </a-col>
      </a-row>

      <!-- Table -->
      <a-table
        :columns="columns"
        :data-source="filteredGoods"
        :loading="loading"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <div>
              <strong>{{ record.name }}</strong>
              <br />
              <span class="text-muted">{{ record.description || '-' }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'category'">
            {{ getCategoryLabel(record.category) }}
          </template>
          <template v-else-if="column.key === 'nutrition'">
            <div class="nutrition-info">
              <a-tag color="red">{{ record.calories_per_100g?.toFixed(1) }} kkal</a-tag>
              <a-tag color="blue">P: {{ record.protein_per_100g?.toFixed(1) }}g</a-tag>
              <a-tag color="green">K: {{ record.carbs_per_100g?.toFixed(1) }}g</a-tag>
              <a-tag color="orange">L: {{ record.fat_per_100g?.toFixed(1) }}g</a-tag>
            </div>
          </template>
          <template v-else-if="column.key === 'portion_requirements'">
            <div class="portion-info">
              <div v-if="record.quantity_per_portion_small > 0" class="portion-item">
                <span class="portion-label">Kecil:</span>
                <span class="portion-value">{{ record.quantity_per_portion_small }} {{ record.unit }}</span>
              </div>
              <div v-if="record.quantity_per_portion_large > 0" class="portion-item">
                <span class="portion-label">Besar:</span>
                <span class="portion-value">{{ record.quantity_per_portion_large }} {{ record.unit }}</span>
              </div>
              <span v-if="!record.quantity_per_portion_small && !record.quantity_per_portion_large" class="text-muted">
                Belum diatur
              </span>
            </div>
          </template>
          <template v-else-if="column.key === 'stock'">
            <a-space>
              <span :class="{ 'text-danger': record.stock_quantity < record.min_threshold }">
                {{ record.stock_quantity?.toFixed(2) }} {{ record.unit }}
              </span>
              <a-tag v-if="record.stock_quantity < record.min_threshold" color="red">Stok Rendah</a-tag>
            </a-space>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="primary" size="small" @click="showProduceModal(record)">
                <template #icon><PlayCircleOutlined /></template>
                Produksi
              </a-button>
              <a-button size="small" @click="showEditModal(record)">
                <template #icon><EditOutlined /></template>
              </a-button>
              <a-popconfirm
                title="Yakin ingin menghapus?"
                @confirm="handleDelete(record.id)"
              >
                <a-button type="danger" size="small">
                  <template #icon><DeleteOutlined /></template>
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- Create/Edit Modal -->
    <SemiFinishedFormModal
      v-model:visible="formModalVisible"
      :edit-data="editingRecord"
      @success="handleFormSuccess"
    />

    <!-- Production Modal -->
    <a-modal
      v-model:visible="produceModalVisible"
      title="Produksi Komponen"
      @ok="handleProduce"
      :confirm-loading="produceLoading"
    >
      <template v-if="producingRecord">
        <a-descriptions :column="1" bordered size="small" class="mb-4">
          <a-descriptions-item label="Nama">{{ producingRecord.name }}</a-descriptions-item>
          <a-descriptions-item label="Resep">{{ producingRecord.recipe?.name }}</a-descriptions-item>
          <a-descriptions-item label="Yield">{{ producingRecord.recipe?.yield_amount }} {{ producingRecord.unit }}</a-descriptions-item>
        </a-descriptions>

        <a-alert
          message="Bahan Baku yang Diperlukan"
          type="info"
          class="mb-4"
        >
          <template #description>
            <ul class="ingredient-list">
              <li v-for="ing in producingRecord.recipe?.ingredients" :key="ing.id">
                {{ ing.ingredient?.name }}: {{ (ing.quantity * produceQuantity / producingRecord.recipe?.yield_amount).toFixed(2) }} {{ ing.ingredient?.unit }}
              </li>
            </ul>
          </template>
        </a-alert>

        <a-form :model="produceForm" layout="vertical">
          <a-form-item label="Jumlah yang Diproduksi" required>
            <a-input-number
              v-model:value="produceForm.quantity"
              :min="1"
              :step="0.1"
              style="width: 100%"
              addon-after="batch"
            />
            <div class="help-text">
              Hasil: {{ (produceForm.quantity * producingRecord.recipe?.yield_amount).toFixed(2) }} {{ producingRecord.unit }}
            </div>
          </a-form-item>
          <a-form-item label="Catatan">
            <a-textarea
              v-model:value="produceForm.notes"
              rows="2"
              placeholder="Catatan produksi (opsional)"
            />
          </a-form-item>
        </a-form>
      </template>
    </a-modal>

    <!-- Inventory Modal -->
    <a-modal
      v-model:visible="inventoryModalVisible"
      title="Stok Komponen"
      width="800px"
      :footer="null"
    >
      <a-table
        :columns="inventoryColumns"
        :data-source="inventoryData"
        :loading="inventoryLoading"
        size="small"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <a-tag :color="getStockStatusColor(record)">
              {{ getStockStatus(record) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'last_updated'">
            {{ formatDate(record.last_updated) }}
          </template>
        </template>
      </a-table>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  PlayCircleOutlined,
  StockOutlined
} from '@ant-design/icons-vue'
import semiFinishedService from '@/services/semiFinishedService'
import SemiFinishedFormModal from '@/components/SemiFinishedFormModal.vue'

const loading = ref(false)
const produceLoading = ref(false)
const inventoryLoading = ref(false)
const goods = ref([])
const inventoryData = ref([])
const searchText = ref('')
const categoryFilter = ref(undefined)

const formModalVisible = ref(false)
const produceModalVisible = ref(false)
const inventoryModalVisible = ref(false)
const editingRecord = ref(null)
const producingRecord = ref(null)

const produceForm = ref({
  quantity: 1,
  notes: ''
})

const columns = [
  { title: 'Nama', key: 'name', width: '20%' },
  { title: 'Kategori', dataIndex: 'category', key: 'category', width: '10%' },
  { title: 'Informasi Gizi (per 100g)', key: 'nutrition', width: '23%' },
  { title: 'Kebutuhan Per Porsi', key: 'portion_requirements', width: '17%' },
  { title: 'Stok', key: 'stock', width: '12%' },
  { title: 'Aksi', key: 'actions', width: '18%', align: 'center' }
]

const inventoryColumns = [
  { title: 'ID', dataIndex: 'semi_finished_goods_id', width: '10%' },
  { title: 'Nama Komponen', dataIndex: ['semi_finished_goods', 'name'], width: '25%' },
  { title: 'Stok', dataIndex: 'quantity', width: '15%' },
  { title: 'Min Threshold', dataIndex: 'min_threshold', width: '15%' },
  { title: 'Status', key: 'status', width: '15%' },
  { title: 'Terakhir Update', key: 'last_updated', width: '20%' }
]

const filteredGoods = computed(() => {
  let result = goods.value
  
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter(item => item.name.toLowerCase().includes(search))
  }
  
  if (categoryFilter.value) {
    result = result.filter(item => item.category === categoryFilter.value)
  }
  
  return result
})

const fetchGoods = async () => {
  loading.value = true
  try {
    const response = await semiFinishedService.getAllSemiFinishedGoods()
    goods.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat data komponen')
    console.error('Error fetching semi-finished goods:', error)
  } finally {
    loading.value = false
  }
}

const fetchInventory = async () => {
  inventoryLoading.value = true
  try {
    const response = await semiFinishedService.getSemiFinishedInventory()
    inventoryData.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat data stok')
    console.error('Error fetching inventory:', error)
  } finally {
    inventoryLoading.value = false
  }
}

const handleSearch = () => {
  // Filter is computed property, no need to fetch
}

const showCreateModal = () => {
  editingRecord.value = null
  formModalVisible.value = true
}

const showEditModal = (record) => {
  editingRecord.value = record
  formModalVisible.value = true
}

const showProduceModal = async (record) => {
  try {
    // Get full details including recipe
    const response = await semiFinishedService.getSemiFinishedGoods(record.id)
    producingRecord.value = response.data.data
    produceForm.value = { quantity: 1, notes: '' }
    produceModalVisible.value = true
  } catch (error) {
    message.error('Gagal memuat detail barang')
  }
}

const showInventoryModal = () => {
  inventoryModalVisible.value = true
  fetchInventory()
}

const handleProduce = async () => {
  if (!producingRecord.value) return
  
  produceLoading.value = true
  try {
    await semiFinishedService.produceSemiFinishedGoods(
      producingRecord.value.id,
      produceForm.value.quantity,
      produceForm.value.notes
    )
    message.success('Produksi berhasil! Stok telah diperbarui.')
    produceModalVisible.value = false
    fetchGoods()
  } catch (error) {
    if (error.response?.data?.error_code === 'INSUFFICIENT_STOCK') {
      message.error('Stok bahan baku tidak mencukupi!')
    } else {
      message.error('Gagal melakukan produksi')
    }
    console.error('Error producing:', error)
  } finally {
    produceLoading.value = false
  }
}

const handleDelete = async (id) => {
  try {
    await semiFinishedService.deleteSemiFinishedGoods(id)
    message.success('Komponen berhasil dihapus')
    fetchGoods()
  } catch (error) {
    message.error('Gagal menghapus barang')
    console.error('Error deleting:', error)
  }
}

const handleFormSuccess = () => {
  formModalVisible.value = false
  fetchGoods()
}

const getStockStatus = (record) => {
  if (record.quantity <= 0) return 'Habis'
  if (record.quantity < record.min_threshold) return 'Rendah'
  return 'Tersedia'
}

const getStockStatusColor = (record) => {
  if (record.quantity <= 0) return 'red'
  if (record.quantity < record.min_threshold) return 'orange'
  return 'green'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('id-ID')
}

const getCategoryLabel = (category) => {
  const labels = {
    protein_hewani: 'Protein Hewani',
    sumber_lemak: 'Sumber Lemak',
    lainnya: 'Lainnya',
    sayur: 'Sayur',
    karbohidrat: 'Karbohidrat',
    buah: 'Buah',
    susu: 'Susu',
    protein_nabati: 'Protein Nabati'
  }
  return labels[category] || category || '-'
}

onMounted(() => {
  fetchGoods()
})
</script>

<style scoped>
.semi-finished-goods {
  padding: 24px;
}

.mb-4 {
  margin-bottom: 16px;
}

.text-muted {
  color: #8c8c8c;
  font-size: 12px;
}

.text-danger {
  color: #ff4d4f;
  font-weight: bold;
}

.nutrition-info {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.portion-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.portion-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.portion-label {
  color: #8c8c8c;
  font-weight: 500;
  min-width: 45px;
}

.portion-value {
  color: #1890ff;
  font-weight: 600;
}

.help-text {
  color: #8c8c8c;
  font-size: 12px;
  margin-top: 4px;
}

.ingredient-list {
  margin: 0;
  padding-left: 16px;
}
</style>
