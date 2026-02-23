<template>
  <div class="inventory">
    <a-page-header
      title="Manajemen Inventory"
      sub-title="Pantau stok bahan baku dan pergerakan inventory"
    />

    <a-row :gutter="16" style="margin-bottom: 16px">
      <!-- Low Stock Alert Card -->
      <a-col :span="8">
        <a-card>
          <a-statistic
            title="Item Stok Menipis"
            :value="lowStockCount"
            :value-style="{ color: lowStockCount > 0 ? '#cf1322' : '#3f8600' }"
          >
            <template #prefix>
              <WarningOutlined v-if="lowStockCount > 0" />
              <CheckCircleOutlined v-else />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <!-- Total Items Card -->
      <a-col :span="8">
        <a-card>
          <a-statistic
            title="Total Item"
            :value="totalItems"
          >
            <template #prefix>
              <InboxOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>

      <!-- Last Update Card -->
      <a-col :span="8">
        <a-card>
          <a-statistic
            title="Terakhir Diperbarui"
            :value="lastUpdate"
          >
            <template #prefix>
              <ClockCircleOutlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <a-card>
      <a-tabs v-model:activeKey="activeTab">
        <!-- Inventory List Tab -->
        <a-tab-pane key="inventory" tab="Daftar Inventory">
          <a-space direction="vertical" style="width: 100%" :size="16">
            <!-- Search and Filter -->
            <a-row :gutter="16">
              <a-col :span="12">
                <a-input-search
                  v-model:value="searchText"
                  placeholder="Cari nama bahan..."
                  @search="handleSearch"
                  allow-clear
                />
              </a-col>
              <a-col :span="6">
                <a-select
                  v-model:value="filterStockLevel"
                  placeholder="Level Stok"
                  style="width: 100%"
                  @change="handleSearch"
                  allow-clear
                >
                  <a-select-option value="low">Stok Menipis</a-select-option>
                  <a-select-option value="normal">Stok Normal</a-select-option>
                  <a-select-option value="high">Stok Berlebih</a-select-option>
                </a-select>
              </a-col>
              <a-col :span="6">
                <a-button type="default" @click="fetchInventory" block>
                  <template #icon><ReloadOutlined /></template>
                  Refresh
                </a-button>
              </a-col>
            </a-row>

            <!-- Table -->
            <a-table
              :columns="inventoryColumns"
              :data-source="inventory"
              :loading="loading"
              :pagination="pagination"
              @change="handleTableChange"
              row-key="id"
              :row-class-name="getRowClassName"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'ingredient_name'">
                  <strong>{{ record.ingredient?.name }}</strong>
                </template>
                <template v-else-if="column.key === 'quantity'">
                  {{ record.quantity }} {{ record.ingredient?.unit }}
                </template>
                <template v-else-if="column.key === 'min_threshold'">
                  {{ record.min_threshold }} {{ record.ingredient?.unit }}
                </template>
                <template v-else-if="column.key === 'stock_status'">
                  <a-tag :color="getStockStatusColor(record)">
                    {{ getStockStatusText(record) }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'days_of_supply'">
                  <span :class="{ 'text-danger': getDaysOfSupply(record) < 7 }">
                    {{ getDaysOfSupply(record) }} hari
                  </span>
                </template>
                <template v-else-if="column.key === 'last_updated'">
                  {{ formatDateTime(record.last_updated) }}
                </template>
                <template v-else-if="column.key === 'actions'">
                  <a-button type="link" size="small" @click="viewMovements(record)">
                    Riwayat
                  </a-button>
                </template>
              </template>
            </a-table>
          </a-space>
        </a-tab-pane>

        <!-- Low Stock Alerts Tab -->
        <a-tab-pane key="alerts" tab="Alert Stok Menipis">
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
                      <strong>{{ item.ingredient?.name }}</strong>
                    </a-space>
                  </template>
                  <template #description>
                    <a-space direction="vertical" size="small">
                      <span>
                        Stok saat ini: <strong>{{ item.quantity }} {{ item.ingredient?.unit }}</strong>
                      </span>
                      <span>
                        Batas minimum: <strong>{{ item.min_threshold }} {{ item.ingredient?.unit }}</strong>
                      </span>
                      <span>
                        Perkiraan habis dalam: <strong class="text-danger">{{ getDaysOfSupply(item) }} hari</strong>
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
        </a-tab-pane>

        <!-- Movement History Tab -->
        <a-tab-pane key="movements" tab="Riwayat Pergerakan">
          <a-space direction="vertical" style="width: 100%" :size="16">
            <!-- Filters -->
            <a-row :gutter="16">
              <a-col :span="8">
                <a-select
                  v-model:value="movementFilters.ingredient_id"
                  placeholder="Pilih bahan"
                  style="width: 100%"
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
              <a-col :span="8">
                <a-range-picker
                  v-model:value="movementFilters.dateRange"
                  style="width: 100%"
                  format="DD/MM/YYYY"
                  @change="fetchMovements"
                />
              </a-col>
              <a-col :span="8">
                <a-select
                  v-model:value="movementFilters.movement_type"
                  placeholder="Tipe Pergerakan"
                  style="width: 100%"
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
            <a-table
              :columns="movementColumns"
              :data-source="movements"
              :loading="loadingMovements"
              :pagination="movementPagination"
              @change="handleMovementTableChange"
              row-key="id"
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
            </a-table>
          </a-space>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- Movement Detail Modal -->
    <a-modal
      v-model:open="movementModalVisible"
      :title="`Riwayat Pergerakan - ${selectedItem?.ingredient?.name}`"
      :footer="null"
      width="800px"
    >
      <a-table
        :columns="movementColumns"
        :data-source="itemMovements"
        :loading="loadingItemMovements"
        :pagination="{ pageSize: 10 }"
        size="small"
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
      </a-table>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { useRouter } from 'vue-router'
import {
  WarningOutlined,
  CheckCircleOutlined,
  InboxOutlined,
  ClockCircleOutlined,
  ReloadOutlined
} from '@ant-design/icons-vue'
import inventoryService from '@/services/inventoryService'

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

const lowStockCount = computed(() => {
  return inventory.value.filter(item => item.quantity < item.min_threshold).length
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
    title: 'Nama Bahan',
    key: 'ingredient_name',
    sorter: true
  },
  {
    title: 'Stok Saat Ini',
    key: 'quantity',
    width: 150
  },
  {
    title: 'Batas Minimum',
    key: 'min_threshold',
    width: 150
  },
  {
    title: 'Status',
    key: 'stock_status',
    width: 120
  },
  {
    title: 'Perkiraan Habis',
    key: 'days_of_supply',
    width: 150
  },
  {
    title: 'Terakhir Diperbarui',
    key: 'last_updated',
    width: 180
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 100
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
      page_size: pagination.pageSize,
      search: searchText.value || undefined,
      stock_level: filterStockLevel.value
    }
    const response = await inventoryService.getInventory(params)
    inventory.value = response.data.data || []
    pagination.total = response.data.total || 0
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
    lowStockAlerts.value = response.data.data || []
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
    movements.value = response.data.data || []
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
    itemMovements.value = response.data.data || []
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

const createPOForItem = (item) => {
  // Navigate to PO creation page with pre-filled item
  router.push({
    name: 'purchase-orders',
    query: { ingredient_id: item.ingredient_id }
  })
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
</script>

<style scoped>
.inventory {
  padding: 24px;
}

:deep(.low-stock-row) {
  background-color: #fff1f0;
}

:deep(.low-stock-row:hover) {
  background-color: #ffe7e6 !important;
}

.text-danger {
  color: #cf1322;
  font-weight: 500;
}

.text-success {
  color: #3f8600;
  font-weight: 500;
}
</style>
