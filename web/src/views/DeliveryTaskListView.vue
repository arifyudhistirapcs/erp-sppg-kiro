<template>
  <div class="delivery-task-list">
    <a-page-header
      title="Manajemen Tugas Pengiriman"
      sub-title="Kelola tugas pengiriman makanan ke sekolah"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Buat Tugas Pengiriman
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="6">
            <a-date-picker
              v-model:value="filterDate"
              placeholder="Pilih tanggal"
              style="width: 100%"
              @change="handleSearch"
              format="DD/MM/YYYY"
            />
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterDriver"
              placeholder="Pilih driver"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
              show-search
              :filter-option="filterDriverOption"
            >
              <a-select-option 
                v-for="driver in drivers" 
                :key="driver.id" 
                :value="driver.id"
              >
                {{ driver.full_name }}
              </a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterStatus"
              placeholder="Status"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="pending">Menunggu</a-select-option>
              <a-select-option value="in_progress">Dalam Perjalanan</a-select-option>
              <a-select-option value="completed">Selesai</a-select-option>
              <a-select-option value="cancelled">Dibatalkan</a-select-option>
            </a-select>
          </a-col>
          <a-col :span="6">
            <a-button @click="resetFilters">
              <template #icon><ReloadOutlined /></template>
              Reset Filter
            </a-button>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="deliveryTasks"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'task_date'">
              {{ formatDate(record.task_date) }}
            </template>
            <template v-else-if="column.key === 'driver'">
              <a-space>
                <a-avatar size="small">{{ getDriverInitials(record.driver?.full_name) }}</a-avatar>
                {{ record.driver?.full_name || '-' }}
              </a-space>
            </template>
            <template v-else-if="column.key === 'school'">
              <a-space direction="vertical" size="small">
                <span><strong>{{ record.school?.name }}</strong></span>
                <span class="text-gray">{{ record.portions }} porsi</span>
              </a-space>
            </template>
            <template v-else-if="column.key === 'route_order'">
              <a-tag color="blue">{{ record.route_order }}</a-tag>
            </template>
            <template v-else-if="column.key === 'status'">
              <a-tag :color="getStatusColor(record.status)">
                {{ getStatusText(record.status) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'menu_items'">
              <a-space direction="vertical" size="small">
                <span 
                  v-for="item in record.menu_items" 
                  :key="item.id"
                  class="menu-item"
                >
                  {{ item.recipe?.name }} ({{ item.portions }} porsi)
                </span>
              </a-space>
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewTask(record)">
                  Detail
                </a-button>
                <a-button 
                  type="link" 
                  size="small" 
                  @click="editTask(record)"
                  :disabled="record.status === 'completed'"
                >
                  Edit
                </a-button>
                <a-dropdown>
                  <template #overlay>
                    <a-menu @click="({ key }) => updateTaskStatus(record.id, key)">
                      <a-menu-item key="pending" :disabled="record.status === 'pending'">
                        Menunggu
                      </a-menu-item>
                      <a-menu-item key="in_progress" :disabled="record.status === 'in_progress'">
                        Dalam Perjalanan
                      </a-menu-item>
                      <a-menu-item key="completed" :disabled="record.status === 'completed'">
                        Selesai
                      </a-menu-item>
                      <a-menu-item key="cancelled" :disabled="record.status === 'cancelled'">
                        Batalkan
                      </a-menu-item>
                    </a-menu>
                  </template>
                  <a-button type="link" size="small">
                    Status <DownOutlined />
                  </a-button>
                </a-dropdown>
                <a-popconfirm
                  title="Yakin ingin menghapus tugas ini?"
                  ok-text="Ya"
                  cancel-text="Tidak"
                  @confirm="deleteTask(record.id)"
                  :disabled="record.status === 'completed'"
                >
                  <a-button 
                    type="link" 
                    size="small" 
                    danger
                    :disabled="record.status === 'completed'"
                  >
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
      :title="editingTask ? 'Edit Tugas Pengiriman' : 'Buat Tugas Pengiriman'"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="800px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Tanggal Pengiriman" name="task_date">
              <a-date-picker
                v-model:value="formData.task_date"
                style="width: 100%"
                format="DD/MM/YYYY"
                placeholder="Pilih tanggal"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Driver" name="driver_id">
              <a-select
                v-model:value="formData.driver_id"
                placeholder="Pilih driver"
                show-search
                :filter-option="filterDriverOption"
              >
                <a-select-option 
                  v-for="driver in drivers" 
                  :key="driver.id" 
                  :value="driver.id"
                >
                  {{ driver.full_name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Sekolah Tujuan" name="school_id">
              <a-select
                v-model:value="formData.school_id"
                placeholder="Pilih sekolah"
                show-search
                :filter-option="filterSchoolOption"
                @change="onSchoolChange"
              >
                <a-select-option 
                  v-for="school in schools" 
                  :key="school.id" 
                  :value="school.id"
                >
                  {{ school.name }} ({{ school.student_count }} siswa)
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="6">
            <a-form-item label="Total Porsi" name="portions">
              <a-input-number
                v-model:value="formData.portions"
                :min="1"
                style="width: 100%"
                placeholder="0"
              />
            </a-form-item>
          </a-col>
          <a-col :span="6">
            <a-form-item label="Urutan Rute" name="route_order">
              <a-input-number
                v-model:value="formData.route_order"
                :min="1"
                style="width: 100%"
                placeholder="1"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="Menu Items" name="menu_items">
          <div class="menu-items-section">
            <div 
              v-for="(item, index) in formData.menu_items" 
              :key="index"
              class="menu-item-row"
            >
              <a-row :gutter="8" align="middle">
                <a-col :span="12">
                  <a-select
                    v-model:value="item.recipe_id"
                    placeholder="Pilih menu"
                    show-search
                    :filter-option="filterRecipeOption"
                  >
                    <a-select-option 
                      v-for="recipe in recipes" 
                      :key="recipe.id" 
                      :value="recipe.id"
                    >
                      {{ recipe.name }}
                    </a-select-option>
                  </a-select>
                </a-col>
                <a-col :span="8">
                  <a-input-number
                    v-model:value="item.portions"
                    :min="1"
                    placeholder="Porsi"
                    style="width: 100%"
                  />
                </a-col>
                <a-col :span="4">
                  <a-button 
                    type="text" 
                    danger 
                    @click="removeMenuItem(index)"
                    :disabled="formData.menu_items.length === 1"
                  >
                    <template #icon><DeleteOutlined /></template>
                  </a-button>
                </a-col>
              </a-row>
            </div>
            <a-button 
              type="dashed" 
              @click="addMenuItem" 
              style="width: 100%; margin-top: 8px"
            >
              <template #icon><PlusOutlined /></template>
              Tambah Menu
            </a-button>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Tugas Pengiriman"
      :footer="null"
      width="900px"
    >
      <div v-if="selectedTask">
        <a-descriptions bordered :column="2">
          <a-descriptions-item label="Tanggal Pengiriman" :span="2">
            {{ formatDate(selectedTask.task_date) }}
          </a-descriptions-item>
          <a-descriptions-item label="Driver">
            {{ selectedTask.driver?.full_name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="getStatusColor(selectedTask.status)">
              {{ getStatusText(selectedTask.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Sekolah Tujuan">
            {{ selectedTask.school?.name }}
          </a-descriptions-item>
          <a-descriptions-item label="Total Porsi">
            {{ selectedTask.portions }} porsi
          </a-descriptions-item>
          <a-descriptions-item label="Urutan Rute">
            {{ selectedTask.route_order }}
          </a-descriptions-item>
          <a-descriptions-item label="Alamat Sekolah" :span="2">
            {{ selectedTask.school?.address }}
          </a-descriptions-item>
        </a-descriptions>

        <a-divider>Informasi Sekolah</a-divider>
        <a-descriptions bordered :column="2">
          <a-descriptions-item label="Kontak Person">
            {{ selectedTask.school?.contact_person || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Telepon">
            {{ selectedTask.school?.phone_number || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Jumlah Siswa">
            {{ formatNumber(selectedTask.school?.student_count) }} siswa
          </a-descriptions-item>
          <a-descriptions-item label="Koordinat GPS">
            {{ selectedTask.school?.latitude?.toFixed(6) }}, {{ selectedTask.school?.longitude?.toFixed(6) }}
          </a-descriptions-item>
        </a-descriptions>

        <a-space style="margin-top: 16px">
          <a-button 
            type="primary" 
            @click="openMaps(selectedTask.school?.latitude, selectedTask.school?.longitude)"
          >
            <template #icon><EnvironmentOutlined /></template>
            Buka di Maps
          </a-button>
          <a-button @click="copyCoordinates(selectedTask.school?.latitude, selectedTask.school?.longitude)">
            <template #icon><CopyOutlined /></template>
            Salin Koordinat
          </a-button>
        </a-space>

        <a-divider>Menu Items</a-divider>
        <a-table
          :columns="menuItemColumns"
          :data-source="selectedTask.menu_items"
          :pagination="false"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'recipe_name'">
              {{ record.recipe?.name }}
            </template>
            <template v-else-if="column.key === 'portions'">
              {{ record.portions }} porsi
            </template>
          </template>
        </a-table>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { 
  PlusOutlined, 
  ReloadOutlined, 
  DownOutlined, 
  DeleteOutlined,
  EnvironmentOutlined,
  CopyOutlined
} from '@ant-design/icons-vue'
import deliveryTaskService from '@/services/deliveryTaskService'
import schoolService from '@/services/schoolService'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const editingTask = ref(null)
const selectedTask = ref(null)
const deliveryTasks = ref([])
const drivers = ref([])
const schools = ref([])
const recipes = ref([])
const formRef = ref()

// Filters
const filterDate = ref(null)
const filterDriver = ref(undefined)
const filterStatus = ref(undefined)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  task_date: null,
  driver_id: undefined,
  school_id: undefined,
  portions: 0,
  route_order: 1,
  menu_items: [
    { recipe_id: undefined, portions: 0 }
  ]
})

const rules = {
  task_date: [{ required: true, message: 'Tanggal pengiriman wajib diisi' }],
  driver_id: [{ required: true, message: 'Driver wajib dipilih' }],
  school_id: [{ required: true, message: 'Sekolah tujuan wajib dipilih' }],
  portions: [
    { required: true, message: 'Total porsi wajib diisi' },
    { type: 'number', min: 1, message: 'Total porsi minimal 1' }
  ],
  route_order: [
    { required: true, message: 'Urutan rute wajib diisi' },
    { type: 'number', min: 1, message: 'Urutan rute minimal 1' }
  ],
  menu_items: [
    { required: true, message: 'Menu items wajib diisi' },
    { type: 'array', min: 1, message: 'Minimal 1 menu item' }
  ]
}

const columns = [
  {
    title: 'Tanggal',
    key: 'task_date',
    width: 120,
    sorter: true
  },
  {
    title: 'Driver',
    key: 'driver',
    width: 150
  },
  {
    title: 'Sekolah & Porsi',
    key: 'school',
    width: 200
  },
  {
    title: 'Urutan',
    key: 'route_order',
    width: 80,
    align: 'center'
  },
  {
    title: 'Status',
    key: 'status',
    width: 120
  },
  {
    title: 'Menu Items',
    key: 'menu_items',
    width: 250
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 200
  }
]

const menuItemColumns = [
  {
    title: 'Menu',
    key: 'recipe_name'
  },
  {
    title: 'Porsi',
    key: 'portions',
    width: 100
  }
]

const fetchDeliveryTasks = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize
    }
    
    if (filterDate.value) {
      params.date = dayjs(filterDate.value).format('YYYY-MM-DD')
    }
    if (filterDriver.value) {
      params.driver_id = filterDriver.value
    }
    if (filterStatus.value) {
      params.status = filterStatus.value
    }

    const response = await deliveryTaskService.getDeliveryTasks(params)
    deliveryTasks.value = response.data.delivery_tasks || []
    pagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat data tugas pengiriman')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchDrivers = async () => {
  try {
    const response = await deliveryTaskService.getDrivers()
    drivers.value = response.data.data || []
  } catch (error) {
    console.error('Gagal memuat data driver:', error)
  }
}

const fetchSchools = async () => {
  try {
    const response = await schoolService.getSchools({ is_active: true })
    schools.value = response.data.schools || []
  } catch (error) {
    console.error('Gagal memuat data sekolah:', error)
  }
}

const fetchRecipes = async () => {
  try {
    const response = await deliveryTaskService.getAvailableRecipes()
    recipes.value = response.data.recipes || []
  } catch (error) {
    console.error('Gagal memuat data resep:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchDeliveryTasks()
}

const handleSearch = () => {
  pagination.current = 1
  fetchDeliveryTasks()
}

const resetFilters = () => {
  filterDate.value = null
  filterDriver.value = undefined
  filterStatus.value = undefined
  handleSearch()
}

const showCreateModal = () => {
  editingTask.value = null
  resetForm()
  modalVisible.value = true
}

const editTask = (task) => {
  editingTask.value = task
  Object.assign(formData, {
    task_date: dayjs(task.task_date),
    driver_id: task.driver_id,
    school_id: task.school_id,
    portions: task.portions,
    route_order: task.route_order,
    menu_items: task.menu_items?.map(item => ({
      recipe_id: item.recipe_id,
      portions: item.portions
    })) || [{ recipe_id: undefined, portions: 0 }]
  })
  modalVisible.value = true
}

const viewTask = (task) => {
  selectedTask.value = task
  detailModalVisible.value = true
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    const submitData = {
      task_date: dayjs(formData.task_date).format('YYYY-MM-DD'),
      driver_id: formData.driver_id,
      school_id: formData.school_id,
      portions: formData.portions,
      route_order: formData.route_order,
      menu_items: formData.menu_items.filter(item => item.recipe_id && item.portions > 0)
    }

    if (editingTask.value) {
      await deliveryTaskService.updateDeliveryTask(editingTask.value.id, submitData)
      message.success('Tugas pengiriman berhasil diperbarui')
    } else {
      await deliveryTaskService.createDeliveryTask(submitData)
      message.success('Tugas pengiriman berhasil dibuat')
    }

    modalVisible.value = false
    fetchDeliveryTasks()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    message.error('Gagal menyimpan tugas pengiriman')
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const updateTaskStatus = async (taskId, status) => {
  try {
    await deliveryTaskService.updateDeliveryTaskStatus(taskId, status)
    message.success('Status tugas berhasil diperbarui')
    fetchDeliveryTasks()
  } catch (error) {
    message.error('Gagal memperbarui status tugas')
    console.error(error)
  }
}

const deleteTask = async (id) => {
  try {
    await deliveryTaskService.deleteDeliveryTask(id)
    message.success('Tugas pengiriman berhasil dihapus')
    fetchDeliveryTasks()
  } catch (error) {
    message.error('Gagal menghapus tugas pengiriman')
    console.error(error)
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    task_date: null,
    driver_id: undefined,
    school_id: undefined,
    portions: 0,
    route_order: 1,
    menu_items: [{ recipe_id: undefined, portions: 0 }]
  })
  formRef.value?.resetFields()
}

const addMenuItem = () => {
  formData.menu_items.push({ recipe_id: undefined, portions: 0 })
}

const removeMenuItem = (index) => {
  if (formData.menu_items.length > 1) {
    formData.menu_items.splice(index, 1)
  }
}

const onSchoolChange = (schoolId) => {
  const school = schools.value.find(s => s.id === schoolId)
  if (school) {
    // Auto-suggest portions based on student count
    formData.portions = school.student_count
  }
}

// Filter functions
const filterDriverOption = (input, option) => {
  const driver = drivers.value.find(d => d.id === option.value)
  return driver?.full_name?.toLowerCase().includes(input.toLowerCase())
}

const filterSchoolOption = (input, option) => {
  const school = schools.value.find(s => s.id === option.value)
  return school?.name?.toLowerCase().includes(input.toLowerCase())
}

const filterRecipeOption = (input, option) => {
  const recipe = recipes.value.find(r => r.id === option.value)
  return recipe?.name?.toLowerCase().includes(input.toLowerCase())
}

// Utility functions
const formatDate = (date) => {
  return dayjs(date).format('DD/MM/YYYY')
}

const formatNumber = (value) => {
  return new Intl.NumberFormat('id-ID').format(value)
}

const getDriverInitials = (name) => {
  if (!name) return '?'
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
}

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    in_progress: 'blue',
    completed: 'green',
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    pending: 'Menunggu',
    in_progress: 'Dalam Perjalanan',
    completed: 'Selesai',
    cancelled: 'Dibatalkan'
  }
  return texts[status] || status
}

const openMaps = (lat, lng) => {
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

const copyCoordinates = async (lat, lng) => {
  try {
    await navigator.clipboard.writeText(`${lat}, ${lng}`)
    message.success('Koordinat berhasil disalin')
  } catch (error) {
    message.error('Gagal menyalin koordinat')
  }
}

onMounted(() => {
  fetchDeliveryTasks()
  fetchDrivers()
  fetchSchools()
  fetchRecipes()
})
</script>

<style scoped>
.delivery-task-list {
  padding: 24px;
}

.menu-item {
  display: block;
  font-size: 12px;
  color: #666;
}

.menu-items-section {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 16px;
  background-color: #fafafa;
}

.menu-item-row {
  margin-bottom: 8px;
}

.menu-item-row:last-child {
  margin-bottom: 0;
}

.text-gray {
  color: #666;
  font-size: 12px;
}
</style>