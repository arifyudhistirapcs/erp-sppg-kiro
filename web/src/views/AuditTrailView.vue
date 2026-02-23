<template>
  <div class="audit-trail">
    <a-page-header
      title="Audit Trail"
      sub-title="Riwayat aktivitas pengguna dalam sistem"
    >
      <template #extra>
        <a-button @click="refreshData" :loading="loading">
          <template #icon><ReloadOutlined /></template>
          Refresh
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Filters -->
        <a-card size="small" title="Filter & Pencarian">
          <a-row :gutter="16">
            <a-col :span="6">
              <a-form-item label="Rentang Tanggal">
                <a-range-picker
                  v-model:value="dateRange"
                  style="width: 100%"
                  placeholder="['Tanggal Mulai', 'Tanggal Akhir']"
                  format="DD/MM/YYYY"
                  @change="handleSearch"
                />
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item label="Pengguna">
                <a-select
                  v-model:value="filterUser"
                  placeholder="Pilih pengguna"
                  style="width: 100%"
                  @change="handleSearch"
                  allow-clear
                  show-search
                  :filter-option="filterUserOption"
                >
                  <a-select-option 
                    v-for="user in users" 
                    :key="user.id" 
                    :value="user.id"
                  >
                    {{ user.full_name }} ({{ user.nik }})
                  </a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item label="Jenis Aksi">
                <a-select
                  v-model:value="filterAction"
                  placeholder="Pilih aksi"
                  style="width: 100%"
                  @change="handleSearch"
                  allow-clear
                >
                  <a-select-option value="create">Membuat</a-select-option>
                  <a-select-option value="update">Mengubah</a-select-option>
                  <a-select-option value="delete">Menghapus</a-select-option>
                  <a-select-option value="login">Masuk</a-select-option>
                  <a-select-option value="logout">Keluar</a-select-option>
                  <a-select-option value="approve">Menyetujui</a-select-option>
                  <a-select-option value="reject">Menolak</a-select-option>
                  <a-select-option value="export">Mengekspor</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="4">
              <a-form-item label="Entitas">
                <a-select
                  v-model:value="filterEntity"
                  placeholder="Pilih entitas"
                  style="width: 100%"
                  @change="handleSearch"
                  allow-clear
                >
                  <a-select-option value="user">Pengguna</a-select-option>
                  <a-select-option value="recipe">Resep</a-select-option>
                  <a-select-option value="menu">Menu</a-select-option>
                  <a-select-option value="supplier">Supplier</a-select-option>
                  <a-select-option value="purchase_order">Purchase Order</a-select-option>
                  <a-select-option value="inventory">Inventori</a-select-option>
                  <a-select-option value="delivery_task">Tugas Pengiriman</a-select-option>
                  <a-select-option value="employee">Karyawan</a-select-option>
                  <a-select-option value="asset">Aset</a-select-option>
                  <a-select-option value="cash_flow">Arus Kas</a-select-option>
                </a-select>
              </a-form-item>
            </a-col>
            <a-col :span="6">
              <a-form-item label="Pencarian">
                <a-input-search
                  v-model:value="searchText"
                  placeholder="Cari dalam deskripsi..."
                  @search="handleSearch"
                  allow-clear
                />
              </a-form-item>
            </a-col>
          </a-row>
        </a-card>

        <!-- Statistics Cards -->
        <a-row :gutter="16">
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Total Aktivitas"
                :value="stats.total_entries || 0"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Login/Logout"
                :value="(stats.action_breakdown?.login || 0) + (stats.action_breakdown?.logout || 0)"
                :value-style="{ color: '#52c41a' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Perubahan Data"
                :value="(stats.action_breakdown?.create || 0) + (stats.action_breakdown?.update || 0) + (stats.action_breakdown?.delete || 0)"
                :value-style="{ color: '#722ed1' }"
              />
            </a-card>
          </a-col>
          <a-col :span="6">
            <a-card size="small">
              <a-statistic
                title="Approval/Export"
                :value="(stats.action_breakdown?.approve || 0) + (stats.action_breakdown?.reject || 0) + (stats.action_breakdown?.export || 0)"
                :value-style="{ color: '#fa8c16' }"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Audit Trail Table -->
        <a-table
          :columns="columns"
          :data-source="auditEntries"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
          :scroll="{ x: 1200 }"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'timestamp'">
              <a-tooltip :title="formatDateTime(record.timestamp)">
                {{ formatRelativeTime(record.timestamp) }}
              </a-tooltip>
            </template>
            <template v-else-if="column.key === 'user'">
              <a-space direction="vertical" :size="2">
                <span style="font-weight: 500;">{{ record.user.full_name }}</span>
                <span style="font-size: 12px; color: #666;">{{ record.user.nik }}</span>
              </a-space>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-tag :color="getActionColor(record.action)">
                {{ getActionLabel(record.action) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'entity'">
              <a-tag color="blue">
                {{ getEntityLabel(record.entity) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'description'">
              <span>{{ record.description }}</span>
            </template>
            <template v-else-if="column.key === 'ip_address'">
              <a-typography-text code>{{ record.ip_address }}</a-typography-text>
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-button type="link" size="small" @click="viewDetails(record)">
                Detail
              </a-button>
            </template>
          </template>
        </a-table>
      </a-space>
    </a-card>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Audit Trail"
      :footer="null"
      width="800px"
    >
      <a-descriptions v-if="selectedEntry" bordered :column="1">
        <a-descriptions-item label="ID">
          {{ selectedEntry.id }}
        </a-descriptions-item>
        <a-descriptions-item label="Waktu">
          {{ formatDateTime(selectedEntry.timestamp) }}
        </a-descriptions-item>
        <a-descriptions-item label="Pengguna">
          <a-space>
            <span>{{ selectedEntry.user.full_name }}</span>
            <a-tag>{{ selectedEntry.user.nik }}</a-tag>
            <a-tag color="blue">{{ getRoleLabel(selectedEntry.user.role) }}</a-tag>
          </a-space>
        </a-descriptions-item>
        <a-descriptions-item label="Aksi">
          <a-tag :color="getActionColor(selectedEntry.action)">
            {{ getActionLabel(selectedEntry.action) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Entitas">
          <a-tag color="blue">{{ getEntityLabel(selectedEntry.entity) }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="ID Entitas" v-if="selectedEntry.entity_id">
          {{ selectedEntry.entity_id }}
        </a-descriptions-item>
        <a-descriptions-item label="IP Address">
          <a-typography-text code>{{ selectedEntry.ip_address }}</a-typography-text>
        </a-descriptions-item>
        <a-descriptions-item label="Deskripsi">
          {{ selectedEntry.description }}
        </a-descriptions-item>
        <a-descriptions-item label="Nilai Lama" v-if="selectedEntry.old_value && selectedEntry.old_value !== '{}'">
          <pre style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-size: 12px; max-height: 200px; overflow-y: auto;">{{ formatJSON(selectedEntry.old_value) }}</pre>
        </a-descriptions-item>
        <a-descriptions-item label="Nilai Baru" v-if="selectedEntry.new_value && selectedEntry.new_value !== '{}'">
          <pre style="background: #f5f5f5; padding: 8px; border-radius: 4px; font-size: 12px; max-height: 200px; overflow-y: auto;">{{ formatJSON(selectedEntry.new_value) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import auditService from '@/services/auditService'
import employeeService from '@/services/employeeService'

// Add relative time plugin
dayjs.extend(relativeTime)

const loading = ref(false)
const auditEntries = ref([])
const stats = ref({})
const users = ref([])
const selectedEntry = ref(null)
const detailModalVisible = ref(false)

// Filters
const dateRange = ref([])
const filterUser = ref(undefined)
const filterAction = ref(undefined)
const filterEntity = ref(undefined)
const searchText = ref('')

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total, range) => `${range[0]}-${range[1]} dari ${total} entri`
})

const columns = [
  {
    title: 'Waktu',
    key: 'timestamp',
    width: 120,
    sorter: true
  },
  {
    title: 'Pengguna',
    key: 'user',
    width: 150
  },
  {
    title: 'Aksi',
    key: 'action',
    width: 100
  },
  {
    title: 'Entitas',
    key: 'entity',
    width: 120
  },
  {
    title: 'Deskripsi',
    key: 'description',
    ellipsis: true
  },
  {
    title: 'IP Address',
    key: 'ip_address',
    width: 120
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 80,
    fixed: 'right'
  }
]

const fetchAuditTrail = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      user_id: filterUser.value || undefined,
      action: filterAction.value || undefined,
      entity: filterEntity.value || undefined,
      start_date: dateRange.value?.[0]?.format('YYYY-MM-DD') || undefined,
      end_date: dateRange.value?.[1]?.format('YYYY-MM-DD') || undefined,
      search: searchText.value || undefined
    }

    const response = await auditService.getAuditTrail(params)
    auditEntries.value = response.data || []
    pagination.total = response.total || 0
  } catch (error) {
    message.error('Gagal memuat data audit trail')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const params = {
      start_date: dateRange.value?.[0]?.format('YYYY-MM-DD') || undefined,
      end_date: dateRange.value?.[1]?.format('YYYY-MM-DD') || undefined
    }

    const response = await auditService.getAuditStats(params)
    stats.value = response.data || {}
  } catch (error) {
    console.error('Gagal memuat statistik audit trail:', error)
  }
}

const fetchUsers = async () => {
  try {
    const response = await employeeService.getEmployees({ page_size: 1000 })
    users.value = response.data?.map(emp => ({
      id: emp.user_id,
      full_name: emp.full_name,
      nik: emp.nik,
      role: emp.user?.role
    })) || []
  } catch (error) {
    console.error('Gagal memuat data pengguna:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchAuditTrail()
}

const handleSearch = () => {
  pagination.current = 1
  fetchAuditTrail()
  fetchStats()
}

const refreshData = () => {
  fetchAuditTrail()
  fetchStats()
}

const viewDetails = (entry) => {
  selectedEntry.value = entry
  detailModalVisible.value = true
}

const filterUserOption = (input, option) => {
  const user = users.value.find(u => u.id === option.value)
  if (!user) return false
  
  const searchText = input.toLowerCase()
  return user.full_name.toLowerCase().includes(searchText) || 
         user.nik.toLowerCase().includes(searchText)
}

// Helper functions
const getActionColor = (action) => {
  const colors = {
    create: 'green',
    update: 'blue',
    delete: 'red',
    login: 'cyan',
    logout: 'orange',
    approve: 'purple',
    reject: 'magenta',
    export: 'gold'
  }
  return colors[action] || 'default'
}

const getActionLabel = (action) => {
  const labels = {
    create: 'Membuat',
    update: 'Mengubah',
    delete: 'Menghapus',
    login: 'Masuk',
    logout: 'Keluar',
    approve: 'Menyetujui',
    reject: 'Menolak',
    export: 'Mengekspor'
  }
  return labels[action] || action
}

const getEntityLabel = (entity) => {
  const labels = {
    user: 'Pengguna',
    recipe: 'Resep',
    menu: 'Menu',
    supplier: 'Supplier',
    purchase_order: 'Purchase Order',
    inventory: 'Inventori',
    delivery_task: 'Tugas Pengiriman',
    employee: 'Karyawan',
    asset: 'Aset',
    cash_flow: 'Arus Kas'
  }
  return labels[entity] || entity
}

const getRoleLabel = (role) => {
  const labels = {
    kepala_sppg: 'Kepala SPPG',
    kepala_yayasan: 'Kepala Yayasan',
    akuntan: 'Akuntan',
    ahli_gizi: 'Ahli Gizi',
    pengadaan: 'Pengadaan',
    chef: 'Chef',
    packing: 'Packing',
    driver: 'Driver',
    asisten_lapangan: 'Asisten Lapangan'
  }
  return labels[role] || role
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY HH:mm:ss')
}

const formatRelativeTime = (date) => {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

const formatJSON = (jsonString) => {
  try {
    const obj = JSON.parse(jsonString)
    return JSON.stringify(obj, null, 2)
  } catch (e) {
    return jsonString
  }
}

// Set default date range to last 7 days
const setDefaultDateRange = () => {
  const endDate = dayjs()
  const startDate = dayjs().subtract(7, 'day')
  dateRange.value = [startDate, endDate]
}

onMounted(() => {
  setDefaultDateRange()
  fetchUsers()
  fetchAuditTrail()
  fetchStats()
})
</script>

<style scoped>
.audit-trail {
  padding: 24px;
}

.ant-descriptions-item-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>