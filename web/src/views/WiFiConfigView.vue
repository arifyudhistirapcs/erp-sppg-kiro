<template>
  <div class="wifi-config">
    <a-page-header
      title="Konfigurasi Wi-Fi"
      sub-title="Kelola jaringan Wi-Fi yang diotorisasi untuk absensi karyawan"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Tambah Jaringan Wi-Fi
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="8">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari SSID atau lokasi..."
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

        <!-- Statistics Cards -->
        <a-row :gutter="16">
          <a-col :span="8">
            <a-card size="small">
              <a-statistic
                title="Total Jaringan"
                :value="stats.total || 0"
                :value-style="{ color: '#1890ff' }"
              />
            </a-card>
          </a-col>
          <a-col :span="8">
            <a-card size="small">
              <a-statistic
                title="Jaringan Aktif"
                :value="stats.active || 0"
                :value-style="{ color: '#52c41a' }"
              />
            </a-card>
          </a-col>
          <a-col :span="8">
            <a-card size="small">
              <a-statistic
                title="Tidak Aktif"
                :value="stats.inactive || 0"
                :value-style="{ color: '#ff4d4f' }"
              />
            </a-card>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="wifiConfigs"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'ssid'">
              <a-typography-text copyable>{{ record.ssid }}</a-typography-text>
            </template>
            <template v-else-if="column.key === 'bssid'">
              <a-typography-text copyable code>{{ record.bssid }}</a-typography-text>
            </template>
            <template v-else-if="column.key === 'is_active'">
              <a-tag :color="record.is_active ? 'green' : 'red'">
                {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'created_at'">
              {{ formatDateTime(record.created_at) }}
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="editWiFiConfig(record)">
                  Edit
                </a-button>
                <a-popconfirm
                  :title="record.is_active ? 'Yakin ingin menonaktifkan jaringan Wi-Fi ini?' : 'Yakin ingin mengaktifkan jaringan Wi-Fi ini?'"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="toggleWiFiConfigStatus(record)"
                >
                  <a-button type="link" size="small" :danger="record.is_active">
                    {{ record.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
                  </a-button>
                </a-popconfirm>
                <a-popconfirm
                  title="Yakin ingin menghapus jaringan Wi-Fi ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="deleteWiFiConfig(record)"
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
      :title="editingConfig ? 'Edit Jaringan Wi-Fi' : 'Tambah Jaringan Wi-Fi'"
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
        <a-form-item label="SSID (Nama Jaringan)" name="ssid">
          <a-input 
            v-model:value="formData.ssid" 
            placeholder="Masukkan nama jaringan Wi-Fi (SSID)"
            :maxlength="100"
          />
          <div style="font-size: 12px; color: #666; margin-top: 4px;">
            Contoh: SPPG-Office, Kantor-WiFi
          </div>
        </a-form-item>

        <a-form-item label="BSSID (MAC Address)" name="bssid">
          <a-input 
            v-model:value="formData.bssid" 
            placeholder="Masukkan BSSID (MAC Address)"
            :maxlength="17"
            @input="formatBSSID"
          />
          <div style="font-size: 12px; color: #666; margin-top: 4px;">
            Format: XX:XX:XX:XX:XX:XX (contoh: 00:1A:2B:3C:4D:5E)
          </div>
        </a-form-item>

        <a-form-item label="Lokasi" name="location">
          <a-input 
            v-model:value="formData.location" 
            placeholder="Masukkan lokasi jaringan Wi-Fi"
            :maxlength="200"
          />
          <div style="font-size: 12px; color: #666; margin-top: 4px;">
            Contoh: Kantor Pusat SPPG, Ruang Meeting Lt. 2
          </div>
        </a-form-item>

        <a-form-item label="Status" name="is_active">
          <a-switch 
            v-model:checked="formData.is_active" 
            checked-children="Aktif" 
            un-checked-children="Tidak Aktif" 
          />
          <div style="font-size: 12px; color: #666; margin-top: 4px;">
            Hanya jaringan aktif yang dapat digunakan untuk absensi
          </div>
        </a-form-item>
      </a-form>

      <a-alert
        message="Informasi Penting"
        description="BSSID adalah alamat MAC unik dari access point Wi-Fi. Pastikan BSSID yang dimasukkan benar untuk menghindari masalah validasi absensi."
        type="info"
        show-icon
        style="margin-top: 16px"
      />
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import wifiConfigService from '@/services/wifiConfigService'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const editingConfig = ref(null)
const wifiConfigs = ref([])
const searchText = ref('')
const filterStatus = ref(undefined)
const formRef = ref()

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  ssid: '',
  bssid: '',
  location: '',
  is_active: true
})

const rules = {
  ssid: [
    { required: true, message: 'SSID wajib diisi' },
    { min: 1, max: 100, message: 'SSID harus antara 1-100 karakter' }
  ],
  bssid: [
    { required: true, message: 'BSSID wajib diisi' },
    { 
      pattern: /^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$/, 
      message: 'Format BSSID tidak valid (contoh: 00:1A:2B:3C:4D:5E)' 
    }
  ],
  location: [
    { required: true, message: 'Lokasi wajib diisi' },
    { max: 200, message: 'Lokasi maksimal 200 karakter' }
  ]
}

const columns = [
  {
    title: 'SSID',
    key: 'ssid',
    dataIndex: 'ssid',
    sorter: true,
    width: 200
  },
  {
    title: 'BSSID',
    key: 'bssid',
    dataIndex: 'bssid',
    width: 180
  },
  {
    title: 'Lokasi',
    dataIndex: 'location',
    key: 'location'
  },
  {
    title: 'Status',
    key: 'is_active',
    width: 100
  },
  {
    title: 'Dibuat Pada',
    key: 'created_at',
    width: 160
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 200
  }
]

const stats = computed(() => {
  const total = wifiConfigs.value.length
  const active = wifiConfigs.value.filter(config => config.is_active).length
  const inactive = total - active
  
  return { total, active, inactive }
})

const fetchWiFiConfigs = async () => {
  loading.value = true
  try {
    const response = await wifiConfigService.getWiFiConfigs()
    wifiConfigs.value = response.data || []
    pagination.total = wifiConfigs.value.length
    
    // Apply client-side filtering
    applyFilters()
  } catch (error) {
    message.error('Gagal memuat konfigurasi Wi-Fi')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const applyFilters = () => {
  let filtered = [...wifiConfigs.value]
  
  // Apply search filter
  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    filtered = filtered.filter(config => 
      config.ssid.toLowerCase().includes(search) ||
      config.location.toLowerCase().includes(search)
    )
  }
  
  // Apply status filter
  if (filterStatus.value === 'active') {
    filtered = filtered.filter(config => config.is_active)
  } else if (filterStatus.value === 'inactive') {
    filtered = filtered.filter(config => !config.is_active)
  }
  
  // Update pagination
  pagination.total = filtered.length
  
  // Apply pagination
  const start = (pagination.current - 1) * pagination.pageSize
  const end = start + pagination.pageSize
  wifiConfigs.value = filtered.slice(start, end)
}

const handleTableChange = (pag) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchWiFiConfigs()
}

const handleSearch = () => {
  pagination.current = 1
  fetchWiFiConfigs()
}

const showCreateModal = () => {
  editingConfig.value = null
  resetForm()
  modalVisible.value = true
}

const editWiFiConfig = (config) => {
  editingConfig.value = config
  Object.assign(formData, {
    ssid: config.ssid,
    bssid: config.bssid,
    location: config.location,
    is_active: config.is_active
  })
  modalVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    if (editingConfig.value) {
      await wifiConfigService.updateWiFiConfig(editingConfig.value.id, formData)
      message.success('Konfigurasi Wi-Fi berhasil diperbarui')
    } else {
      await wifiConfigService.createWiFiConfig(formData)
      message.success('Konfigurasi Wi-Fi berhasil ditambahkan')
    }

    modalVisible.value = false
    fetchWiFiConfigs()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    
    const errorMessage = error.response?.data?.message || 'Gagal menyimpan konfigurasi Wi-Fi'
    message.error(errorMessage)
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const toggleWiFiConfigStatus = async (config) => {
  try {
    await wifiConfigService.updateWiFiConfig(config.id, { 
      is_active: !config.is_active 
    })
    message.success(`Jaringan Wi-Fi berhasil ${config.is_active ? 'dinonaktifkan' : 'diaktifkan'}`)
    fetchWiFiConfigs()
  } catch (error) {
    message.error('Gagal mengubah status jaringan Wi-Fi')
    console.error(error)
  }
}

const deleteWiFiConfig = async (config) => {
  try {
    await wifiConfigService.deleteWiFiConfig(config.id)
    message.success('Jaringan Wi-Fi berhasil dihapus')
    fetchWiFiConfigs()
  } catch (error) {
    message.error('Gagal menghapus jaringan Wi-Fi')
    console.error(error)
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    ssid: '',
    bssid: '',
    location: '',
    is_active: true
  })
  formRef.value?.resetFields()
}

const formatBSSID = (e) => {
  let value = e.target.value.replace(/[^0-9A-Fa-f]/g, '')
  
  // Add colons every 2 characters
  if (value.length > 0) {
    value = value.match(/.{1,2}/g).join(':')
    if (value.length > 17) {
      value = value.substring(0, 17)
    }
  }
  
  formData.bssid = value.toUpperCase()
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY HH:mm')
}

onMounted(() => {
  fetchWiFiConfigs()
})
</script>

<style scoped>
.wifi-config {
  padding: 24px;
}
</style>