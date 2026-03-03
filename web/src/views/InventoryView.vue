<template>
  <div>
    <!-- Stats Cards Row -->
    <a-row :gutter="[20, 20]">
      <a-col :xs="24" :sm="12" :md="8">
        <HStatCard
          :icon="WarningOutlined"
          :icon-bg="lowStockCount > 0 ? 'linear-gradient(135deg, #EE5D50 0%, #D43F3A 100%)' : 'linear-gradient(135deg, #05CD99 0%, #01B574 100%)'"
          label="Item Stok Menipis"
          :value="lowStockCount"
          :loading="loading"
        />
      </a-col>
      <a-col :xs="24" :sm="12" :md="8">
        <HStatCard
          :icon="InboxOutlined"
          icon-bg="linear-gradient(135deg, #4481EB 0%, #04BEFE 100%)"
          label="Total Item"
          :value="totalItems"
          :loading="loading"
        />
      </a-col>
      <a-col :xs="24" :sm="12" :md="8">
        <HStatCard
          :icon="ClockCircleOutlined"
          icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
          label="Terakhir Diperbarui"
          :value="lastUpdate"
          :loading="loading"
        />
      </a-col>
    </a-row>

    <!-- Tabs Card -->
    <div class="h-card">
      <a-tabs v-model:activeKey="activeTab">
        <!-- Inventory List Tab -->
        <a-tab-pane key="inventory" tab="Daftar Bahan Baku">
          <div class="tab-content">
            <!-- Search and Filter -->
            <a-row :gutter="[16, 16]" style="margin-bottom: 20px">
              <a-col :xs="24" :sm="24" :md="12">
                <a-input-search
                  v-model:value="searchText"
                  placeholder="Cari nama bahan..."
                  @search="handleSearch"
                  allow-clear
                  size="large"
                >
                  <template #prefix>
                    <SearchOutlined />
                  </template>
                </a-input-search>
              </a-col>
              <a-col :xs="24" :sm="12" :md="6">
                <a-select
                  v-model:value="filterStockLevel"
                  placeholder="Level Stok"
                  style="width: 100%"
                  size="large"
                  @change="handleSearch"
                  allow-clear
                >
                  <a-select-option value="low">Stok Menipis</a-select-option>
                  <a-select-option value="normal">Stok Normal</a-select-option>
                  <a-select-option value="high">Stok Berlebih</a-select-option>
                </a-select>
              </a-col>
              <a-col :xs="24" :sm="12" :md="6">
                <a-space style="width: 100%">
                  <a-button type="default" @click="fetchInventory" size="large">
                    <template #icon><ReloadOutlined /></template>
                    Refresh
                  </a-button>
                  <a-button type="primary" @click="initializeInventory" size="large">
                    <template #icon><PlusOutlined /></template>
                    Inisialisasi Bahan
                  </a-button>
                </a-space>
              </a-col>
            </a-row>

            <!-- Table -->
            <HDataTable
              :columns="inventoryColumns"
              :data-source="inventory"
              :loading="loading"
              :pagination="pagination"
              @change="handleTableChange"
              :mobile-card-view="true"
            />
          </div>
        </a-tab-pane>

        <!-- Low Stock Alerts Tab -->
        <a-tab-pane key="alerts" tab="Alert Stok Menipis">
          <div class="tab-content">
            <a-alert
              v-if="lowStockAlerts.length === 0"
              message="Tidak ada item dengan stok menipis"
              type="success"
              show-icon
              style="margin-bottom: 16px"
            />
            <a-list
              v-else
              :data-source="lowStockAlerts"
              :loading="loadingAlerts"
            >
              <template #renderItem="{ item }">
                <a-list-item>
                  <a-list-item-meta>
                    <template #title>
                      <a-space>
                        <WarningOutlined style="color: #cf1322" />
                        <strong>{{ item.ingredient_name }}</strong>
                      </a-space>
                    </template>
                    <template #description>
                      <a-space direction="vertical" size="small">
                        <span>
                          Stok saat ini: <strong>{{ item.current_stock }} {{ item.unit }}</strong>
                        </span>
                        <span>
                          Batas minimum: <strong>{{ item.min_threshold }} {{ item.unit }}</strong>
                        </span>
                        <span>
                          Perkiraan habis dalam: <strong class="text-danger">{{ item.days_remaining }} hari</strong>
                        </span>
                      </a-space>
                    </template>
                  </a-list-item-meta>
                  <template #actions>
                    <a-button type="primary" size="small" @click="createPOForItem(item)">
                      Buat PO
                    </a-button>
                  </template>
                </a-list-item>
              </template>
            </a-list>
          </div>
        </a-tab-pane>

        <!-- Movement History Tab -->
        <a-tab-pane key="movements" tab="Riwayat Pergerakan">
          <div class="tab-content">
            <!-- Filters -->
            <a-row :gutter="[16, 16]" style="margin-bottom: 20px">
              <a-col :xs="24" :sm="24" :md="8">
                <a-select
                  v-model:value="movementFilters.ingredient_id"
                  placeholder="Pilih bahan"
                  style="width: 100%"
                  size="large"
                  show-search
                  :filter-option="filterIngredient"
                  allow-clear
                  @change="fetchMovements"
                >
                  <a-select-option
                    v-for="item in inventory"
                    :key="item.ingredient_id"
                    :value="item.ingredient_id"
                  >
                    {{ item.ingredient?.name }}
                  </a-select-option>
                </a-select>
              </a-col>
              <a-col :xs="24" :sm="12" :md="8">
                <a-range-picker
                  v-model:value="movementFilters.dateRange"
                  style="width: 100%"
                  size="large"
                  format="DD/MM/YYYY"
                  @change="fetchMovements"
                />
              </a-col>
              <a-col :xs="24" :sm="12" :md="8">
                <a-select
                  v-model:value="movementFilters.movement_type"
                  placeholder="Tipe Pergerakan"
                  style="width: 100%"
                  size="large"
                  allow-clear
                  @change="fetchMovements"
                >
                  <a-select-option value="in">Masuk</a-select-option>
                  <a-select-option value="out">Keluar</a-select-option>
                  <a-select-option value="adjustment">Penyesuaian</a-select-option>
                </a-select>
              </a-col>
            </a-row>

            <!-- Movements Table -->
            <HDataTable
              :columns="movementColumns"
              :data-source="movements"
              :loading="loadingMovements"
              :pagination="movementPagination"
              @change="handleMovementTableChange"
              :mobile-card-view="true"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'movement_type'">
                  <a-tag :color="getMovementTypeColor(record.movement_type)">
                    {{ getMovementTypeText(record.movement_type) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'quantity'">
                  <span :class="record.movement_type === 'out' ? 'text-danger' : 'text-success'">
                    {{ record.movement_type === 'out' ? '-' : '+' }}{{ record.quantity }}
                  </span>
                </template>
                <template v-else-if="column.key === 'movement_date'">
                  {{ formatDateTime(record.movement_date) }}
                </template>
              </template>
            </HDataTable>
          </div>
        </a-tab-pane>

        <!-- Stok Opname Tab -->
        <a-tab-pane key="stok-opname" tab="Stok Opname">
          <div class="tab-content">
            <StokOpnameList />
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>

    <!-- Movement Detail Modal -->
    <a-modal
      v-model:open="movementModalVisible"
      :title="`Riwayat Pergerakan - ${selectedItem?.ingredient?.name}`"
      :footer="null"
      width="800px"
    >
      <HDataTable
        :columns="movementColumns"
        :data-source="itemMovements"
        :loading="loadingItemMovements"
        :pagination="{ pageSize: 10 }"
        :mobile-card-view="true"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'movement_type'">
            <a-tag :color="getMovementTypeColor(record.movement_type)">
              {{ getMovementTypeText(record.movement_type) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'quantity'">
            <span :class="record.movement_type === 'out' ? 'text-danger' : 'text-success'">
              {{ record.movement_type === 'out' ? '-' : '+' }}{{ record.quantity }}
            </span>
          </template>
          <template v-else-if="column.key === 'movement_date'">
            {{ formatDateTime(record.movement_date) }}
          </template>
        </template>
      </HDataTable>
    </a-modal>

    <!-- Initialize Ingredient Modal (Create New Ingredient) -->
    <a-modal
      v-model:open="initModalVisible"
      title="Inisialisasi Bahan Baru"
      width="600px"
      @ok="handleInitialize"
      :confirm-loading="initLoading"
      ok-text="Simpan"
      cancel-text="Batal"
    >
      <a-spin :spinning="initLoading">
        <a-form
          ref="initFormRef"
          :model="initFormData"
          :rules="initFormRules"
          layout="vertical"
        >
          <a-row :gutter="16">
            <a-col :span="12">
              <a-form-item label="Kode Bahan Baku" name="code">
                <a-input
                  v-model:value="initFormData.code"
                  placeholder="Auto-generate"
                  disabled
                />
              </a-form-item>
            </a-col>
            <a-col :span="12">
              <a-form-item label="Satuan" name="unit">
                <a-select v-model:value="initFormData.unit" placeholder="Pilih satuan">
                  <a-select-option value="gram">gram</a-select-option>
                  <a-select-option value="ml">ml (Mililiter)</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
          
          <a-form-item label="Nama Bahan" name="name">
            <a-input 
              v-model:value="initFormData.name" 
              placeholder="Masukkan nama bahan"
            />
          </a-form-item>

          <a-form-item label="Kategori" name="category">
            <a-select 
              v-model:value="initFormData.category" 
              placeholder="Pilih kategori bahan"
              allow-clear
            >
              <a-select-option value="Buah">Buah</a-select-option>
              <a-select-option value="Bumbu">Bumbu</a-select-option>
              <a-select-option value="Daging, Unggas">Daging, Unggas</a-select-option>
              <a-select-option value="Gula, Sirup, Konfeksioneri">Gula, Sirup, Konfeksioneri</a-select-option>
              <a-select-option value="Ikan, Kerang, Udang">Ikan, Kerang, Udang</a-select-option>
              <a-select-option value="Kacang, Biji, Bean">Kacang, Biji, Bean</a-select-option>
              <a-select-option value="Lemak dan Minyak">Lemak dan Minyak</a-select-option>
              <a-select-option value="Minuman">Minuman</a-select-option>
              <a-select-option value="Sayuran">Sayuran</a-select-option>
              <a-select-option value="Serealia">Serealia</a-select-option>
              <a-select-option value="Susu">Susu</a-select-option>
              <a-select-option value="Telur">Telur</a-select-option>
              <a-select-option value="Umbi Berpati">Umbi Berpati</a-select-option>
            </a-select>
          </a-form-item>
        </a-form>
      </a-spin>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import {
  WarningOutlined,
  CheckCircleOutlined,
  InboxOutlined,
  ClockCircleOutlined,
  ReloadOutlined,
  PlusOutlined,
  SearchOutlined
} from '@ant-design/icons-vue'
import inventoryService from '@/services/inventoryService'
import recipeService from '@/services/recipeService'
import StokOpnameList from '@/components/StokOpnameList.vue'
import HStatCard from '@/components/horizon/HStatCard.vue'
import HDataTable from '@/components/horizon/HDataTable.vue'

const router = useRouter()
const activeTab = ref('inventory')
const loading = ref(false)
const loadingAlerts = ref(false)
const loadingMovements = ref(false)
const loadingItemMovements = ref(false)
const movementModalVisible = ref(false)
const selectedItem = ref(null)
const inventory = ref([])
const lowStockAlerts = ref([])
const movements = ref([])
const itemMovements = ref([])
const searchText = ref('')
const filterStockLevel = ref(undefined)

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0
})

const movementPagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0
})

const movementFilters = reactive({
  ingredient_id: undefined,
  dateRange: null,
  movement_type: undefined
})

// Initialize modal state
const initModalVisible = ref(false)
const initLoading = ref(false)
const initFormRef = ref()
const initFormData = reactive({
  code: '',
  name: '',
  category: undefined,
  unit: undefined
})

const initFormRules = {
  name: [
    { required: true, message: 'Nama bahan harus diisi', trigger: 'blur' },
    { min: 2, message: 'Minimal 2 karakter', trigger: 'blur' }
  ],
  unit: [
    { required: true, message: 'Satuan harus dipilih', trigger: 'change' }
  ]
}

const lowStockCount = computed(() => {
  return inventory.value.filter(item => item.stock_status === 'Stok Rendah').length
})

const totalItems = computed(() => {
  return inventory.value.length
})

const lastUpdate = computed(() => {
  if (inventory.value.length === 0) return '-'
  const latest = inventory.value.reduce((max, item) => {
    const itemDate = new Date(item.last_updated)
    return itemDate > max ? itemDate : max
  }, new Date(0))
  return formatDateTime(latest)
})

const inventoryColumns = [
  {
    title: 'Kode',
    key: 'ingredient_code',
    dataIndex: 'ingredient_code',
    width: 100
  },
  {
    title: 'Nama Bahan',
    key: 'ingredient_name',
    dataIndex: 'ingredient_name',
    sorter: true
  },
  {
    title: 'Stok Saat Ini',
    key: 'quantity',
    dataIndex: 'quantity_display',
    width: 150
  },
  {
    title: 'Batas Minimum',
    key: 'min_threshold',
    dataIndex: 'min_threshold_display',
    width: 150
  },
  {
    title: 'Level Stok',
    key: 'stock_level',
    dataIndex: 'stock_level',
    type: 'progress',
    width: 180
  },
  {
    title: 'Status',
    key: 'stock_status',
    dataIndex: 'stock_status',
    type: 'status',
    width: 120
  }
]

const movementColumns = [
  {
    title: 'Bahan',
    dataIndex: ['ingredient', 'name'],
    key: 'ingredient_name'
  },
  {
    title: 'Tipe',
    key: 'movement_type',
    width: 120
  },
  {
    title: 'Jumlah',
    key: 'quantity',
    width: 100
  },
  {
    title: 'Referensi',
    dataIndex: 'reference',
    key: 'reference'
  },
  {
    title: 'Tanggal',
    key: 'movement_date',
    width: 180
  },
  {
    title: 'Catatan',
    dataIndex: 'notes',
    key: 'notes'
  }
]

const fetchInventory = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize
    }
    
    // Add search parameter if exists
    if (searchText.value && searchText.value.trim() !== '') {
      params.search = searchText.value.trim()
    }
    
    // Add stock level filter if exists
    if (filterStockLevel.value) {
      params.stock_level = filterStockLevel.value
    }
    
    const response = await inventoryService.getInventory(params)
    
    // Backend sends inventory_items, not data
    let items = response.data.inventory_items || []
    
    // Client-side filtering if backend doesn't support it
    if (searchText.value && searchText.value.trim() !== '') {
      const searchLower = searchText.value.toLowerCase().trim()
      items = items.filter(item => 
        item.ingredient?.name?.toLowerCase().includes(searchLower) ||
        item.ingredient?.code?.toLowerCase().includes(searchLower)
      )
    }
    
    // Client-side stock level filtering if backend doesn't support it
    if (filterStockLevel.value) {
      items = items.filter(item => {
        const ratio = item.quantity / item.min_threshold
        if (filterStockLevel.value === 'low') {
          return ratio < 1
        } else if (filterStockLevel.value === 'normal') {
          return ratio >= 1 && ratio < 2
        } else if (filterStockLevel.value === 'high') {
          return ratio >= 2
        }
        return true
      })
    }
    
    inventory.value = items.map(item => ({
      ...item,
      key: item.id || item.ingredient_id,
      ingredient_code: item.ingredient?.code || '-',
      ingredient_name: item.ingredient?.name || '-',
      quantity_display: `${item.quantity} ${item.ingredient?.unit || ''}`,
      min_threshold_display: `${item.min_threshold} ${item.ingredient?.unit || ''}`,
      stock_level: item.min_threshold > 0 ? Math.round((item.quantity / item.min_threshold) * 100) : 100,
      stock_status: item.quantity < item.min_threshold ? 'Stok Rendah' : 'Normal'
    }))
    pagination.total = items.length
  } catch (error) {
    message.error('Gagal memuat data inventory')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchLowStockAlerts = async () => {
  loadingAlerts.value = true
  try {
    const response = await inventoryService.getLowStockAlerts()
    lowStockAlerts.value = response.data.alerts || []
  } catch (error) {
    message.error('Gagal memuat alert stok menipis')
    console.error(error)
  } finally {
    loadingAlerts.value = false
  }
}

const fetchMovements = async () => {
  loadingMovements.value = true
  try {
    const params = {
      page: movementPagination.current,
      page_size: movementPagination.pageSize,
      ingredient_id: movementFilters.ingredient_id,
      movement_type: movementFilters.movement_type
    }
    
    if (movementFilters.dateRange && movementFilters.dateRange.length === 2) {
      params.start_date = movementFilters.dateRange[0].format('YYYY-MM-DD')
      params.end_date = movementFilters.dateRange[1].format('YYYY-MM-DD')
    }
    
    const response = await inventoryService.getInventoryMovements(params)
    movements.value = response.data.movements || []
    movementPagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat riwayat pergerakan')
    console.error(error)
  } finally {
    loadingMovements.value = false
  }
}

const viewMovements = async (item) => {
  selectedItem.value = item
  movementModalVisible.value = true
  loadingItemMovements.value = true
  
  try {
    const response = await inventoryService.getInventoryMovements({
      ingredient_id: item.ingredient_id
    })
    itemMovements.value = response.data.movements || []
  } catch (error) {
    message.error('Gagal memuat riwayat pergerakan')
    console.error(error)
  } finally {
    loadingItemMovements.value = false
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchInventory()
}

const handleMovementTableChange = (pag, filters, sorter) => {
  movementPagination.current = pag.current
  movementPagination.pageSize = pag.pageSize
  fetchMovements()
}

const handleSearch = () => {
  pagination.current = 1
  fetchInventory()
}

const initializeInventory = async () => {
  // Open modal and generate code
  initModalVisible.value = true
  await generateIngredientCode()
}

const generateIngredientCode = async () => {
  try {
    const response = await recipeService.generateIngredientCode()
    initFormData.code = response.data.code || ''
  } catch (error) {
    console.error('Failed to generate code:', error)
    initFormData.code = ''
  }
}

const resetInitForm = () => {
  initFormData.code = ''
  initFormData.name = ''
  initFormData.category = undefined
  initFormData.unit = undefined
  initFormRef.value?.resetFields()
}

const handleInitialize = async () => {
  try {
    await initFormRef.value.validate()
  } catch (error) {
    return
  }
  
  initLoading.value = true
  try {
    // Create new ingredient
    const ingredientData = {
      name: initFormData.name,
      category: initFormData.category,
      unit: initFormData.unit,
      code: initFormData.code
    }
    
    const response = await recipeService.createIngredient(ingredientData)
    const newIngredient = response.data.data || response.data.ingredient
    
    // Initialize inventory for the new ingredient
    if (newIngredient && newIngredient.id) {
      await inventoryService.initializeInventoryItem(newIngredient.id)
    }
    
    message.success('Bahan baru berhasil ditambahkan ke inventory')
    initModalVisible.value = false
    resetInitForm()
    fetchInventory()
  } catch (error) {
    message.error('Gagal menambahkan bahan')
    console.error(error)
  } finally {
    initLoading.value = false
  }
}

const createPOForItem = (item) => {
  // Navigate to PO creation page with pre-filled item
  router.push({
    name: 'purchase-orders',
    query: { ingredient_id: item.ingredient_id }
  })
}

const getStockLevelPercent = (record) => {
  if (!record.min_threshold || record.min_threshold === 0) return 100
  const ratio = (record.quantity / record.min_threshold) * 100
  return Math.min(Math.round(ratio), 100)
}

const getStockLevelColor = (record) => {
  const ratio = record.quantity / record.min_threshold
  if (ratio < 1) return '#EE5D50' // Red - low stock
  if (ratio < 1.5) return '#FFB547' // Orange - warning
  return '#05CD99' // Green - good stock
}

const getRowClassName = (record) => {
  if (record.quantity < record.min_threshold) {
    return 'low-stock-row'
  }
  return ''
}

const getStockStatusColor = (record) => {
  if (record.quantity < record.min_threshold) {
    return 'red'
  } else if (record.quantity < record.min_threshold * 1.5) {
    return 'orange'
  }
  return 'green'
}

const getStockStatusText = (record) => {
  if (record.quantity < record.min_threshold) {
    return 'Stok Menipis'
  } else if (record.quantity < record.min_threshold * 1.5) {
    return 'Perlu Perhatian'
  }
  return 'Stok Aman'
}

const getDaysOfSupply = (record) => {
  // Simple calculation: assume average daily usage is 10% of min threshold
  const avgDailyUsage = record.min_threshold * 0.1
  if (avgDailyUsage === 0) return 999
  return Math.floor(record.quantity / avgDailyUsage)
}

const getMovementTypeColor = (type) => {
  const colors = {
    in: 'green',
    out: 'red',
    adjustment: 'blue'
  }
  return colors[type] || 'default'
}

const getMovementTypeText = (type) => {
  const texts = {
    in: 'Masuk',
    out: 'Keluar',
    adjustment: 'Penyesuaian'
  }
  return texts[type] || type
}

const filterIngredient = (input, option) => {
  return option.children[0].children.toLowerCase().includes(input.toLowerCase())
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

onMounted(() => {
  fetchInventory()
  fetchLowStockAlerts()
  fetchMovements()
})

// Watch for modal close to reset form
watch(() => initModalVisible.value, (newVal) => {
  if (!newVal) {
    resetInitForm()
  }
})
</script>

<style scoped>
/* Stats to Tabs spacing */
.h-card {
  margin-top: 16px;
}

/* Tab Content */
.tab-content {
  padding: var(--h-spacing-5) 0;
}

/* Text Utilities */
.text-danger {
  color: var(--h-error);
  font-weight: var(--h-font-semibold);
}

.text-success {
  color: var(--h-success);
  font-weight: var(--h-font-semibold);
}

/* Low Stock Row Highlight */
:deep(.low-stock-row) {
  background-color: rgba(238, 93, 80, 0.05);
}

:deep(.low-stock-row:hover) {
  background-color: rgba(238, 93, 80, 0.1) !important;
}

/* Dark Mode Support */
.dark :deep(.low-stock-row) {
  background-color: rgba(238, 93, 80, 0.1);
}

.dark :deep(.low-stock-row:hover) {
  background-color: rgba(238, 93, 80, 0.15) !important;
}

.dark .text-danger {
  color: #EE5D50;
}

.dark .text-success {
  color: #05CD99;
}
</style>
