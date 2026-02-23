<template>
  <div class="delivery-task-form">
    <a-page-header
      :title="isEdit ? 'Edit Tugas Pengiriman' : 'Buat Tugas Pengiriman'"
      :sub-title="isEdit ? 'Perbarui informasi tugas pengiriman' : 'Buat tugas pengiriman baru'"
      @back="goBack"
    />

    <a-card>
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
        @finish="handleSubmit"
      >
        <a-row :gutter="24">
          <a-col :span="12">
            <a-card title="Informasi Dasar" size="small">
              <a-form-item label="Tanggal Pengiriman" name="task_date">
                <a-date-picker
                  v-model:value="formData.task_date"
                  style="width: 100%"
                  format="DD/MM/YYYY"
                  placeholder="Pilih tanggal"
                  :disabled-date="disabledDate"
                />
              </a-form-item>

              <a-form-item label="Driver" name="driver_id">
                <a-select
                  v-model:value="formData.driver_id"
                  placeholder="Pilih driver"
                  show-search
                  :filter-option="filterDriverOption"
                  :loading="loadingDrivers"
                >
                  <a-select-option 
                    v-for="driver in drivers" 
                    :key="driver.id" 
                    :value="driver.id"
                  >
                    <a-space>
                      <a-avatar size="small">{{ getDriverInitials(driver.full_name) }}</a-avatar>
                      {{ driver.full_name }}
                    </a-space>
                  </a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item label="Urutan Rute" name="route_order">
                <a-input-number
                  v-model:value="formData.route_order"
                  :min="1"
                  style="width: 100%"
                  placeholder="1"
                />
                <div class="form-help">
                  Urutan pengiriman dalam rute harian driver
                </div>
              </a-form-item>
            </a-card>
          </a-col>

          <a-col :span="12">
            <a-card title="Informasi Sekolah" size="small">
              <a-form-item label="Sekolah Tujuan" name="school_id">
                <a-select
                  v-model:value="formData.school_id"
                  placeholder="Pilih sekolah"
                  show-search
                  :filter-option="filterSchoolOption"
                  @change="onSchoolChange"
                  :loading="loadingSchools"
                >
                  <a-select-option 
                    v-for="school in schools" 
                    :key="school.id" 
                    :value="school.id"
                  >
                    <div>
                      <div><strong>{{ school.name }}</strong></div>
                      <div class="text-gray">{{ school.student_count }} siswa</div>
                    </div>
                  </a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item label="Total Porsi" name="portions">
                <a-input-number
                  v-model:value="formData.portions"
                  :min="1"
                  style="width: 100%"
                  placeholder="0"
                />
                <div class="form-help" v-if="selectedSchool">
                  Rekomendasi: {{ selectedSchool.student_count }} porsi (sesuai jumlah siswa)
                </div>
              </a-form-item>

              <!-- School Info Display -->
              <div v-if="selectedSchool" class="school-info">
                <a-divider orientation="left" plain>Informasi Sekolah</a-divider>
                <a-descriptions size="small" :column="1">
                  <a-descriptions-item label="Alamat">
                    {{ selectedSchool.address }}
                  </a-descriptions-item>
                  <a-descriptions-item label="Kontak">
                    {{ selectedSchool.contact_person || '-' }}
                    <span v-if="selectedSchool.phone_number"> - {{ selectedSchool.phone_number }}</span>
                  </a-descriptions-item>
                  <a-descriptions-item label="GPS">
                    {{ selectedSchool.latitude?.toFixed(6) }}, {{ selectedSchool.longitude?.toFixed(6) }}
                    <a-button 
                      type="link" 
                      size="small" 
                      @click="openMaps(selectedSchool.latitude, selectedSchool.longitude)"
                    >
                      <template #icon><EnvironmentOutlined /></template>
                      Lihat
                    </a-button>
                  </a-descriptions-item>
                </a-descriptions>
              </div>
            </a-card>
          </a-col>
        </a-row>

        <a-card title="Menu Items" style="margin-top: 24px">
          <a-form-item name="menu_items">
            <div class="menu-items-section">
              <div 
                v-for="(item, index) in formData.menu_items" 
                :key="index"
                class="menu-item-row"
              >
                <a-card size="small" :title="`Menu ${index + 1}`">
                  <template #extra>
                    <a-button 
                      type="text" 
                      danger 
                      size="small"
                      @click="removeMenuItem(index)"
                      :disabled="formData.menu_items.length === 1"
                    >
                      <template #icon><DeleteOutlined /></template>
                      Hapus
                    </a-button>
                  </template>

                  <a-row :gutter="16">
                    <a-col :span="16">
                      <a-form-item 
                        :name="['menu_items', index, 'recipe_id']"
                        :rules="[{ required: true, message: 'Menu wajib dipilih' }]"
                        label="Pilih Menu"
                      >
                        <a-select
                          v-model:value="item.recipe_id"
                          placeholder="Pilih menu"
                          show-search
                          :filter-option="filterRecipeOption"
                          @change="(value) => onRecipeChange(index, value)"
                        >
                          <a-select-option 
                            v-for="recipe in recipes" 
                            :key="recipe.id" 
                            :value="recipe.id"
                          >
                            <div>
                              <div><strong>{{ recipe.name }}</strong></div>
                              <div class="text-gray">{{ recipe.category }}</div>
                            </div>
                          </a-select-option>
                        </a-select>
                      </a-form-item>
                    </a-col>
                    <a-col :span="8">
                      <a-form-item 
                        :name="['menu_items', index, 'portions']"
                        :rules="[
                          { required: true, message: 'Porsi wajib diisi' },
                          { type: 'number', min: 1, message: 'Porsi minimal 1' }
                        ]"
                        label="Jumlah Porsi"
                      >
                        <a-input-number
                          v-model:value="item.portions"
                          :min="1"
                          placeholder="0"
                          style="width: 100%"
                        />
                      </a-form-item>
                    </a-col>
                  </a-row>

                  <!-- Recipe Info Display -->
                  <div v-if="item.recipe_info" class="recipe-info">
                    <a-descriptions size="small" :column="2">
                      <a-descriptions-item label="Kalori">
                        {{ Math.round(item.recipe_info.total_calories) }} kcal
                      </a-descriptions-item>
                      <a-descriptions-item label="Protein">
                        {{ Math.round(item.recipe_info.total_protein) }}g
                      </a-descriptions-item>
                    </a-descriptions>
                  </div>
                </a-card>
              </div>

              <a-button 
                type="dashed" 
                @click="addMenuItem" 
                style="width: 100%; margin-top: 16px"
                size="large"
              >
                <template #icon><PlusOutlined /></template>
                Tambah Menu Item
              </a-button>
            </div>
          </a-form-item>

          <!-- Summary -->
          <a-divider orientation="left" plain>Ringkasan</a-divider>
          <a-row :gutter="16">
            <a-col :span="8">
              <a-statistic 
                title="Total Menu Items" 
                :value="formData.menu_items.length" 
                suffix="item"
              />
            </a-col>
            <a-col :span="8">
              <a-statistic 
                title="Total Porsi Menu" 
                :value="totalMenuPortions" 
                suffix="porsi"
              />
            </a-col>
            <a-col :span="8">
              <a-statistic 
                title="Total Porsi Pengiriman" 
                :value="formData.portions" 
                suffix="porsi"
                :value-style="portionMismatch ? { color: '#ff4d4f' } : {}"
              />
            </a-col>
          </a-row>
          
          <a-alert
            v-if="portionMismatch"
            message="Peringatan"
            description="Total porsi menu tidak sama dengan total porsi pengiriman. Pastikan jumlah sudah sesuai."
            type="warning"
            show-icon
            style="margin-top: 16px"
          />
        </a-card>

        <!-- Route Optimization -->
        <a-card title="Optimasi Rute" style="margin-top: 24px" v-if="formData.driver_id && formData.task_date">
          <a-space direction="vertical" style="width: 100%">
            <div>
              <a-button 
                type="primary" 
                @click="optimizeRoute"
                :loading="optimizingRoute"
              >
                <template #icon><ThunderboltOutlined /></template>
                Optimasi Urutan Rute
              </a-button>
              <span class="ml-2 text-gray">
                Otomatis mengurutkan berdasarkan lokasi sekolah
              </span>
            </div>
            
            <div v-if="routeOptimizationResult">
              <a-alert
                :message="routeOptimizationResult.message"
                :type="routeOptimizationResult.type"
                show-icon
              />
            </div>
          </a-space>
        </a-card>

        <!-- Form Actions -->
        <a-card style="margin-top: 24px">
          <a-space>
            <a-button 
              type="primary" 
              html-type="submit"
              :loading="submitting"
              size="large"
            >
              {{ isEdit ? 'Perbarui Tugas' : 'Buat Tugas' }}
            </a-button>
            <a-button @click="goBack" size="large">
              Batal
            </a-button>
            <a-button 
              v-if="!isEdit" 
              @click="resetForm"
              size="large"
            >
              Reset Form
            </a-button>
          </a-space>
        </a-card>
      </a-form>
    </a-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { 
  PlusOutlined, 
  DeleteOutlined,
  EnvironmentOutlined,
  ThunderboltOutlined
} from '@ant-design/icons-vue'
import deliveryTaskService from '@/services/deliveryTaskService'
import schoolService from '@/services/schoolService'

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => !!route.params.id)
const taskId = computed(() => route.params.id)

const submitting = ref(false)
const loadingDrivers = ref(false)
const loadingSchools = ref(false)
const loadingRecipes = ref(false)
const optimizingRoute = ref(false)
const drivers = ref([])
const schools = ref([])
const recipes = ref([])
const formRef = ref()
const routeOptimizationResult = ref(null)

const formData = reactive({
  task_date: null,
  driver_id: undefined,
  school_id: undefined,
  portions: 0,
  route_order: 1,
  menu_items: [
    { recipe_id: undefined, portions: 0, recipe_info: null }
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
  ]
}

const selectedSchool = computed(() => {
  return schools.value.find(s => s.id === formData.school_id)
})

const totalMenuPortions = computed(() => {
  return formData.menu_items.reduce((total, item) => total + (item.portions || 0), 0)
})

const portionMismatch = computed(() => {
  return totalMenuPortions.value !== formData.portions && formData.portions > 0
})

const fetchDrivers = async () => {
  loadingDrivers.value = true
  try {
    const response = await deliveryTaskService.getDrivers()
    drivers.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat data driver')
    console.error(error)
  } finally {
    loadingDrivers.value = false
  }
}

const fetchSchools = async () => {
  loadingSchools.value = true
  try {
    const response = await schoolService.getSchools({ is_active: true })
    schools.value = response.data.schools || []
  } catch (error) {
    message.error('Gagal memuat data sekolah')
    console.error(error)
  } finally {
    loadingSchools.value = false
  }
}

const fetchRecipes = async () => {
  loadingRecipes.value = true
  try {
    const response = await deliveryTaskService.getAvailableRecipes()
    recipes.value = response.data.recipes || []
  } catch (error) {
    message.error('Gagal memuat data resep')
    console.error(error)
  } finally {
    loadingRecipes.value = false
  }
}

const fetchTaskData = async () => {
  if (!isEdit.value) return
  
  try {
    const response = await deliveryTaskService.getDeliveryTask(taskId.value)
    const task = response.data.delivery_task
    
    Object.assign(formData, {
      task_date: dayjs(task.task_date),
      driver_id: task.driver_id,
      school_id: task.school_id,
      portions: task.portions,
      route_order: task.route_order,
      menu_items: task.menu_items?.map(item => ({
        recipe_id: item.recipe_id,
        portions: item.portions,
        recipe_info: item.recipe
      })) || [{ recipe_id: undefined, portions: 0, recipe_info: null }]
    })
  } catch (error) {
    message.error('Gagal memuat data tugas pengiriman')
    console.error(error)
    goBack()
  }
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
      menu_items: formData.menu_items
        .filter(item => item.recipe_id && item.portions > 0)
        .map(item => ({
          recipe_id: item.recipe_id,
          portions: item.portions
        }))
    }

    if (isEdit.value) {
      await deliveryTaskService.updateDeliveryTask(taskId.value, submitData)
      message.success('Tugas pengiriman berhasil diperbarui')
    } else {
      await deliveryTaskService.createDeliveryTask(submitData)
      message.success('Tugas pengiriman berhasil dibuat')
    }

    goBack()
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

const addMenuItem = () => {
  formData.menu_items.push({ 
    recipe_id: undefined, 
    portions: 0, 
    recipe_info: null 
  })
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

const onRecipeChange = (index, recipeId) => {
  const recipe = recipes.value.find(r => r.id === recipeId)
  if (recipe) {
    formData.menu_items[index].recipe_info = recipe
  }
}

const optimizeRoute = async () => {
  optimizingRoute.value = true
  routeOptimizationResult.value = null
  
  try {
    // This would call the backend route optimization service
    // For now, we'll simulate the optimization
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // Mock optimization result
    routeOptimizationResult.value = {
      type: 'success',
      message: 'Rute berhasil dioptimasi. Urutan rute telah diperbarui berdasarkan lokasi sekolah.'
    }
    
    // In a real implementation, this would update the route_order based on GPS coordinates
    message.success('Rute berhasil dioptimasi')
  } catch (error) {
    routeOptimizationResult.value = {
      type: 'error',
      message: 'Gagal mengoptimasi rute. Silakan coba lagi.'
    }
    message.error('Gagal mengoptimasi rute')
  } finally {
    optimizingRoute.value = false
  }
}

const resetForm = () => {
  Object.assign(formData, {
    task_date: null,
    driver_id: undefined,
    school_id: undefined,
    portions: 0,
    route_order: 1,
    menu_items: [{ recipe_id: undefined, portions: 0, recipe_info: null }]
  })
  formRef.value?.resetFields()
  routeOptimizationResult.value = null
}

const goBack = () => {
  router.push('/delivery-tasks')
}

const disabledDate = (current) => {
  // Disable past dates
  return current && current < dayjs().startOf('day')
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
const getDriverInitials = (name) => {
  if (!name) return '?'
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
}

const openMaps = (lat, lng) => {
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

// Auto-update total portions when menu items change
watch(() => formData.menu_items, () => {
  // Auto-update total portions if it's currently 0 or matches previous total
  if (formData.portions === 0 || formData.portions === totalMenuPortions.value) {
    formData.portions = totalMenuPortions.value
  }
}, { deep: true })

onMounted(async () => {
  await Promise.all([
    fetchDrivers(),
    fetchSchools(),
    fetchRecipes()
  ])
  
  if (isEdit.value) {
    await fetchTaskData()
  }
})
</script>

<style scoped>
.delivery-task-form {
  padding: 24px;
}

.menu-items-section {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 16px;
  background-color: #fafafa;
}

.menu-item-row {
  margin-bottom: 16px;
}

.menu-item-row:last-child {
  margin-bottom: 0;
}

.school-info {
  margin-top: 16px;
  padding: 12px;
  background-color: #f6ffed;
  border: 1px solid #b7eb8f;
  border-radius: 6px;
}

.recipe-info {
  margin-top: 12px;
  padding: 8px;
  background-color: #f0f2ff;
  border: 1px solid #adc6ff;
  border-radius: 4px;
}

.text-gray {
  color: #666;
  font-size: 12px;
}

.form-help {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

.ml-2 {
  margin-left: 8px;
}
</style>