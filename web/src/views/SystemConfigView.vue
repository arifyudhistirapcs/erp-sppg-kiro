<template>
  <div class="system-config">
    <a-page-header
      title="Konfigurasi Sistem"
      sub-title="Kelola parameter operasional sistem SPPG"
    >
      <template #extra>
        <a-space>
          <a-button @click="refreshData" :loading="loading">
            <template #icon><ReloadOutlined /></template>
            Refresh
          </a-button>
          <a-button @click="initializeDefaults" :loading="initializing">
            <template #icon><SettingOutlined /></template>
            Inisialisasi Default
          </a-button>
          <a-button type="primary" @click="saveAllConfigs" :loading="saving">
            <template #icon><SaveOutlined /></template>
            Simpan Semua Perubahan
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-row :gutter="24">
      <!-- Inventory Settings -->
      <a-col :span="12">
        <a-card title="Pengaturan Inventori" class="config-card">
          <template #extra>
            <a-tag color="blue">{{ configCounts.inventory || 0 }} parameter</a-tag>
          </template>
          
          <a-space direction="vertical" style="width: 100%" :size="16">
            <a-form-item label="Ambang Batas Stok Minimum (hari)">
              <a-input-number
                v-model:value="configs.inventory_min_stock_days"
                :min="1"
                :max="365"
                style="width: 100%"
                placeholder="Masukkan jumlah hari"
              />
              <div class="config-description">
                Sistem akan memberikan peringatan jika stok bahan baku tersisa untuk kurang dari jumlah hari ini
              </div>
            </a-form-item>

            <a-form-item label="Persentase Alert Stok Rendah (%)">
              <a-input-number
                v-model:value="configs.inventory_low_stock_percentage"
                :min="1"
                :max="100"
                style="width: 100%"
                placeholder="Masukkan persentase"
              />
              <div class="config-description">
                Alert akan muncul ketika stok mencapai persentase ini dari ambang batas minimum
              </div>
            </a-form-item>

            <a-form-item label="Metode Pengelolaan Stok">
              <a-select
                v-model:value="configs.inventory_stock_method"
                style="width: 100%"
                placeholder="Pilih metode"
              >
                <a-select-option value="FIFO">FIFO (First In First Out)</a-select-option>
                <a-select-option value="FEFO">FEFO (First Expired First Out)</a-select-option>
              </a-select>
              <div class="config-description">
                Metode yang digunakan untuk menentukan urutan penggunaan bahan baku
              </div>
            </a-form-item>

            <a-form-item label="Auto Reorder">
              <a-switch 
                v-model:checked="configs.inventory_auto_reorder"
                checked-children="Aktif" 
                un-checked-children="Tidak Aktif"
              />
              <div class="config-description">
                Otomatis membuat draft Purchase Order ketika stok mencapai batas minimum
              </div>
            </a-form-item>
          </a-space>
        </a-card>
      </a-col>

      <!-- Nutrition Standards -->
      <a-col :span="12">
        <a-card title="Standar Nutrisi" class="config-card">
          <template #extra>
            <a-tag color="green">{{ configCounts.nutrition || 0 }} parameter</a-tag>
          </template>
          
          <a-space direction="vertical" style="width: 100%" :size="16">
            <a-form-item label="Minimum Kalori per Porsi">
              <a-input-number
                v-model:value="configs.nutrition_min_calories"
                :min="100"
                :max="2000"
                style="width: 100%"
                placeholder="Masukkan kalori"
              />
              <div class="config-description">
                Setiap menu harus memenuhi minimum kalori ini per porsi
              </div>
            </a-form-item>

            <a-form-item label="Minimum Protein (gram)">
              <a-input-number
                v-model:value="configs.nutrition_min_protein"
                :min="1"
                :max="100"
                :precision="1"
                style="width: 100%"
                placeholder="Masukkan protein"
              />
              <div class="config-description">
                Kandungan protein minimum yang harus dipenuhi setiap menu
              </div>
            </a-form-item>

            <a-form-item label="Minimum Karbohidrat (gram)">
              <a-input-number
                v-model:value="configs.nutrition_min_carbs"
                :min="1"
                :max="200"
                :precision="1"
                style="width: 100%"
                placeholder="Masukkan karbohidrat"
              />
              <div class="config-description">
                Kandungan karbohidrat minimum yang harus dipenuhi setiap menu
              </div>
            </a-form-item>

            <a-form-item label="Validasi Nutrisi Ketat">
              <a-switch 
                v-model:checked="configs.nutrition_strict_validation"
                checked-children="Aktif" 
                un-checked-children="Tidak Aktif"
              />
              <div class="config-description">
                Jika aktif, menu tidak dapat disetujui jika tidak memenuhi standar nutrisi
              </div>
            </a-form-item>
          </a-space>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="24" style="margin-top: 24px">
      <!-- Security Settings -->
      <a-col :span="12">
        <a-card title="Pengaturan Keamanan" class="config-card">
          <template #extra>
            <a-tag color="red">{{ configCounts.security || 0 }} parameter</a-tag>
          </template>
          
          <a-space direction="vertical" style="width: 100%" :size="16">
            <a-form-item label="Session Timeout (menit)">
              <a-input-number
                v-model:value="configs.security_session_timeout"
                :min="5"
                :max="480"
                style="width: 100%"
                placeholder="Masukkan menit"
              />
              <div class="config-description">
                Durasi maksimal user dapat tidak aktif sebelum otomatis logout
              </div>
            </a-form-item>

            <a-form-item label="Maksimal Percobaan Login">
              <a-input-number
                v-model:value="configs.security_max_login_attempts"
                :min="3"
                :max="10"
                style="width: 100%"
                placeholder="Masukkan jumlah"
              />
              <div class="config-description">
                Jumlah maksimal percobaan login yang gagal sebelum akun dikunci sementara
              </div>
            </a-form-item>

            <a-form-item label="Durasi Kunci Akun (menit)">
              <a-input-number
                v-model:value="configs.security_lockout_duration"
                :min="5"
                :max="60"
                style="width: 100%"
                placeholder="Masukkan menit"
              />
              <div class="config-description">
                Durasi akun dikunci setelah melebihi batas percobaan login
              </div>
            </a-form-item>

            <a-form-item label="Require Strong Password">
              <a-switch 
                v-model:checked="configs.security_strong_password"
                checked-children="Aktif" 
                un-checked-children="Tidak Aktif"
              />
              <div class="config-description">
                Memaksa pengguna menggunakan password yang kuat (minimal 8 karakter, huruf besar, kecil, angka)
              </div>
            </a-form-item>
          </a-space>
        </a-card>
      </a-col>

      <!-- System Operations -->
      <a-col :span="12">
        <a-card title="Operasional Sistem" class="config-card">
          <template #extra>
            <a-tag color="purple">{{ configCounts.system || 0 }} parameter</a-tag>
          </template>
          
          <a-space direction="vertical" style="width: 100%" :size="16">
            <a-form-item label="Jadwal Backup">
              <a-select
                v-model:value="configs.system_backup_schedule"
                style="width: 100%"
                placeholder="Pilih jadwal"
              >
                <a-select-option value="daily">Harian (00:00)</a-select-option>
                <a-select-option value="weekly">Mingguan (Minggu 00:00)</a-select-option>
                <a-select-option value="monthly">Bulanan (Tanggal 1, 00:00)</a-select-option>
              </a-select>
              <div class="config-description">
                Jadwal otomatis untuk backup database sistem
              </div>
            </a-form-item>

            <a-form-item label="Retensi Backup (hari)">
              <a-input-number
                v-model:value="configs.system_backup_retention"
                :min="7"
                :max="365"
                style="width: 100%"
                placeholder="Masukkan hari"
              />
              <div class="config-description">
                Berapa lama file backup disimpan sebelum dihapus otomatis
              </div>
            </a-form-item>

            <a-form-item label="Retensi Log Audit (hari)">
              <a-input-number
                v-model:value="configs.system_audit_retention"
                :min="30"
                :max="1095"
                style="width: 100%"
                placeholder="Masukkan hari"
              />
              <div class="config-description">
                Berapa lama log audit trail disimpan dalam sistem
              </div>
            </a-form-item>

            <a-form-item label="Notifikasi Email">
              <a-switch 
                v-model:checked="configs.system_email_notifications"
                checked-children="Aktif" 
                un-checked-children="Tidak Aktif"
              />
              <div class="config-description">
                Mengirim notifikasi penting melalui email ke administrator
              </div>
            </a-form-item>
          </a-space>
        </a-card>
      </a-col>
    </a-row>

    <!-- Advanced Configuration -->
    <a-card title="Konfigurasi Lanjutan" style="margin-top: 24px">
      <a-collapse>
        <a-collapse-panel key="1" header="Konfigurasi Manual">
          <a-space direction="vertical" style="width: 100%" :size="16">
            <a-alert
              message="Peringatan"
              description="Bagian ini untuk administrator sistem yang berpengalaman. Perubahan yang salah dapat mempengaruhi kinerja sistem."
              type="warning"
              show-icon
            />
            
            <a-table
              :columns="advancedColumns"
              :data-source="advancedConfigs"
              :loading="loading"
              :pagination="false"
              size="small"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'value'">
                  <a-input
                    v-if="record.data_type === 'string'"
                    v-model:value="record.value"
                    size="small"
                    @change="markConfigChanged(record)"
                  />
                  <a-input-number
                    v-else-if="record.data_type === 'int'"
                    v-model:value="record.value"
                    size="small"
                    style="width: 100%"
                    @change="markConfigChanged(record)"
                  />
                  <a-input-number
                    v-else-if="record.data_type === 'float'"
                    v-model:value="record.value"
                    :precision="2"
                    size="small"
                    style="width: 100%"
                    @change="markConfigChanged(record)"
                  />
                  <a-switch
                    v-else-if="record.data_type === 'bool'"
                    v-model:checked="record.value"
                    size="small"
                    @change="markConfigChanged(record)"
                  />
                  <a-input
                    v-else
                    v-model:value="record.value"
                    size="small"
                    @change="markConfigChanged(record)"
                  />
                </template>
                <template v-else-if="column.key === 'data_type'">
                  <a-tag :color="getDataTypeColor(record.data_type)">
                    {{ record.data_type }}
                  </a-tag>
                </template>
                <template v-else-if="column.key === 'actions'">
                  <a-popconfirm
                    title="Yakin ingin menghapus konfigurasi ini?"
                    ok-text="Ya"
                    cancel-text="Tidak"
                    @confirm="deleteConfig(record.key)"
                  >
                    <a-button type="link" size="small" danger>
                      Hapus
                    </a-button>
                  </a-popconfirm>
                </template>
              </template>
            </a-table>

            <a-button type="dashed" @click="showAddConfigModal" style="width: 100%">
              <template #icon><PlusOutlined /></template>
              Tambah Konfigurasi Baru
            </a-button>
          </a-space>
        </a-collapse-panel>
      </a-collapse>
    </a-card>

    <!-- Add Config Modal -->
    <a-modal
      v-model:open="addConfigModalVisible"
      title="Tambah Konfigurasi Baru"
      :confirm-loading="saving"
      @ok="addNewConfig"
      @cancel="cancelAddConfig"
    >
      <a-form
        ref="addConfigFormRef"
        :model="newConfigForm"
        :rules="addConfigRules"
        layout="vertical"
      >
        <a-form-item label="Key" name="key">
          <a-input 
            v-model:value="newConfigForm.key" 
            placeholder="Masukkan key konfigurasi"
          />
        </a-form-item>

        <a-form-item label="Nilai" name="value">
          <a-input 
            v-model:value="newConfigForm.value" 
            placeholder="Masukkan nilai konfigurasi"
          />
        </a-form-item>

        <a-form-item label="Tipe Data" name="data_type">
          <a-select
            v-model:value="newConfigForm.data_type"
            placeholder="Pilih tipe data"
          >
            <a-select-option value="string">String</a-select-option>
            <a-select-option value="int">Integer</a-select-option>
            <a-select-option value="float">Float</a-select-option>
            <a-select-option value="bool">Boolean</a-select-option>
            <a-select-option value="json">JSON</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="Kategori" name="category">
          <a-select
            v-model:value="newConfigForm.category"
            placeholder="Pilih kategori"
          >
            <a-select-option value="inventory">Inventory</a-select-option>
            <a-select-option value="nutrition">Nutrition</a-select-option>
            <a-select-option value="security">Security</a-select-option>
            <a-select-option value="system">System</a-select-option>
            <a-select-option value="other">Lainnya</a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined, SaveOutlined, PlusOutlined, SettingOutlined } from '@ant-design/icons-vue'
import systemConfigService from '@/services/systemConfigService'

const loading = ref(false)
const saving = ref(false)
const initializing = ref(false)
const addConfigModalVisible = ref(false)
const addConfigFormRef = ref()

// Main configuration object
const configs = reactive({
  // Inventory settings
  inventory_min_stock_days: 7,
  inventory_low_stock_percentage: 20,
  inventory_stock_method: 'FEFO',
  inventory_auto_reorder: false,
  
  // Nutrition standards
  nutrition_min_calories: 600,
  nutrition_min_protein: 15.0,
  nutrition_min_carbs: 80.0,
  nutrition_strict_validation: true,
  
  // Security settings
  security_session_timeout: 30,
  security_max_login_attempts: 5,
  security_lockout_duration: 15,
  security_strong_password: true,
  
  // System operations
  system_backup_schedule: 'daily',
  system_backup_retention: 30,
  system_audit_retention: 365,
  system_email_notifications: true
})

// Advanced configurations (for manual editing)
const advancedConfigs = ref([])

// New config form
const newConfigForm = reactive({
  key: '',
  value: '',
  data_type: 'string',
  category: 'other'
})

const addConfigRules = {
  key: [
    { required: true, message: 'Key wajib diisi' },
    { min: 3, message: 'Key minimal 3 karakter' }
  ],
  value: [
    { required: true, message: 'Nilai wajib diisi' }
  ],
  data_type: [
    { required: true, message: 'Tipe data wajib dipilih' }
  ],
  category: [
    { required: true, message: 'Kategori wajib dipilih' }
  ]
}

const advancedColumns = [
  {
    title: 'Key',
    dataIndex: 'key',
    key: 'key',
    width: 200
  },
  {
    title: 'Nilai',
    key: 'value',
    width: 200
  },
  {
    title: 'Tipe',
    key: 'data_type',
    width: 100
  },
  {
    title: 'Kategori',
    dataIndex: 'category',
    key: 'category',
    width: 120
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 80
  }
]

// Computed properties
const configCounts = computed(() => {
  const counts = { inventory: 0, nutrition: 0, security: 0, system: 0 }
  
  Object.keys(configs).forEach(key => {
    if (key.startsWith('inventory_')) counts.inventory++
    else if (key.startsWith('nutrition_')) counts.nutrition++
    else if (key.startsWith('security_')) counts.security++
    else if (key.startsWith('system_')) counts.system++
  })
  
  return counts
})

// Methods
const fetchConfigs = async () => {
  loading.value = true
  try {
    const response = await systemConfigService.getConfigs()
    const configData = response.data || []
    
    // Map configurations to reactive object
    configData.forEach(config => {
      const key = config.key
      let value = config.value
      
      // Parse value based on data type
      if (config.data_type === 'int') {
        value = parseInt(value)
      } else if (config.data_type === 'float') {
        value = parseFloat(value)
      } else if (config.data_type === 'bool') {
        value = value === 'true'
      }
      
      // Set in main configs if it's a predefined config
      if (configs.hasOwnProperty(key)) {
        configs[key] = value
      } else {
        // Add to advanced configs
        advancedConfigs.value.push({
          ...config,
          value: value
        })
      }
    })
  } catch (error) {
    message.error('Gagal memuat konfigurasi sistem')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const saveAllConfigs = async () => {
  saving.value = true
  try {
    const configsToSave = []
    
    // Prepare main configs
    Object.keys(configs).forEach(key => {
      let value = configs[key]
      let dataType = 'string'
      let category = 'other'
      
      // Determine data type and category
      if (typeof value === 'number') {
        dataType = Number.isInteger(value) ? 'int' : 'float'
      } else if (typeof value === 'boolean') {
        dataType = 'bool'
        value = value.toString()
      } else {
        value = value.toString()
      }
      
      // Determine category from key prefix
      if (key.startsWith('inventory_')) category = 'inventory'
      else if (key.startsWith('nutrition_')) category = 'nutrition'
      else if (key.startsWith('security_')) category = 'security'
      else if (key.startsWith('system_')) category = 'system'
      
      configsToSave.push({
        key,
        value,
        data_type: dataType,
        category
      })
    })
    
    // Prepare advanced configs
    advancedConfigs.value.forEach(config => {
      let value = config.value
      if (config.data_type === 'bool') {
        value = value.toString()
      } else if (config.data_type === 'int' || config.data_type === 'float') {
        value = value.toString()
      }
      
      configsToSave.push({
        key: config.key,
        value,
        data_type: config.data_type,
        category: config.category
      })
    })
    
    await systemConfigService.setMultipleConfigs(configsToSave)
    message.success('Semua konfigurasi berhasil disimpan')
  } catch (error) {
    message.error('Gagal menyimpan konfigurasi')
    console.error(error)
  } finally {
    saving.value = false
  }
}

const refreshData = () => {
  fetchConfigs()
}

const initializeDefaults = async () => {
  initializing.value = true
  try {
    await systemConfigService.initializeDefaults()
    message.success('Konfigurasi default berhasil diinisialisasi')
    fetchConfigs() // Refresh data after initialization
  } catch (error) {
    message.error('Gagal menginisialisasi konfigurasi default')
    console.error(error)
  } finally {
    initializing.value = false
  }
}

const markConfigChanged = (config) => {
  // Mark config as changed (could be used for highlighting)
  config.changed = true
}

const getDataTypeColor = (dataType) => {
  const colors = {
    string: 'blue',
    int: 'green',
    float: 'orange',
    bool: 'purple',
    json: 'red'
  }
  return colors[dataType] || 'default'
}

const showAddConfigModal = () => {
  addConfigModalVisible.value = true
}

const addNewConfig = async () => {
  try {
    await addConfigFormRef.value.validate()
    
    const configData = {
      key: newConfigForm.key,
      value: newConfigForm.value,
      data_type: newConfigForm.data_type,
      category: newConfigForm.category
    }
    
    await systemConfigService.setConfig(configData)
    
    // Add to advanced configs list
    advancedConfigs.value.push({
      ...configData,
      value: newConfigForm.data_type === 'bool' ? 
        newConfigForm.value === 'true' : 
        newConfigForm.data_type === 'int' ? 
          parseInt(newConfigForm.value) :
          newConfigForm.data_type === 'float' ?
            parseFloat(newConfigForm.value) :
            newConfigForm.value
    })
    
    message.success('Konfigurasi baru berhasil ditambahkan')
    addConfigModalVisible.value = false
    resetNewConfigForm()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    message.error('Gagal menambahkan konfigurasi')
    console.error(error)
  }
}

const cancelAddConfig = () => {
  addConfigModalVisible.value = false
  resetNewConfigForm()
}

const resetNewConfigForm = () => {
  Object.assign(newConfigForm, {
    key: '',
    value: '',
    data_type: 'string',
    category: 'other'
  })
  addConfigFormRef.value?.resetFields()
}

const deleteConfig = async (key) => {
  try {
    await systemConfigService.deleteConfig(key)
    
    // Remove from advanced configs
    const index = advancedConfigs.value.findIndex(config => config.key === key)
    if (index > -1) {
      advancedConfigs.value.splice(index, 1)
    }
    
    message.success('Konfigurasi berhasil dihapus')
  } catch (error) {
    message.error('Gagal menghapus konfigurasi')
    console.error(error)
  }
}

onMounted(() => {
  fetchConfigs()
})
</script>

<style scoped>
.system-config {
  padding: 24px;
}

.config-card {
  height: 100%;
}

.config-description {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
  line-height: 1.4;
}

.ant-form-item {
  margin-bottom: 16px;
}

.ant-card-head-title {
  font-weight: 600;
}
</style>