<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? 'Edit Menu' : 'Tambah Menu Baru'"
    width="900px"
    :confirm-loading="loading"
    @ok="handleSubmit"
    @cancel="handleCancel"
    ok-text="Simpan"
    cancel-text="Batal"
  >
    <a-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      layout="vertical"
    >
      <a-row :gutter="16">
        <a-col :span="16">
          <a-form-item label="Nama Menu" name="name">
            <a-input v-model:value="formData.name" placeholder="Masukkan nama menu (contoh: Paket Ayam Goreng)" />
          </a-form-item>
        </a-col>
        <a-col :span="8">
          <a-form-item label="Kategori" name="category">
            <a-select v-model:value="formData.category" placeholder="Pilih kategori">
              <a-select-option value="masakan_indonesia">Masakan Indonesia</a-select-option>
              <a-select-option value="masakan_china">Masakan China</a-select-option>
              <a-select-option value="masakan_western">Masakan Western</a-select-option>
              <a-select-option value="masakan_india">Masakan India</a-select-option>
              <a-select-option value="masakan_gabungan">Masakan Gabungan</a-select-option>
              <a-select-option value="lainnya">Lainnya</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <a-form-item label="Foto Menu">
        <a-upload
          v-model:file-list="fileList"
          list-type="picture-card"
          :before-upload="beforeUpload"
          :custom-request="handleUpload"
          :max-count="1"
          accept="image/*"
          @remove="handleRemovePhoto"
        >
          <div v-if="fileList.length < 1">
            <plus-outlined />
            <div style="margin-top: 8px">Upload</div>
          </div>
        </a-upload>
        <div style="color: #8c8c8c; font-size: 12px; margin-top: 8px">
          Format: JPG, PNG, JPEG. Maksimal 2MB
        </div>
      </a-form-item>

      <a-form-item label="Instruksi Penyajian" name="instructions">
        <a-textarea
          v-model:value="formData.instructions"
          :rows="4"
          placeholder="Masukkan cara penyajian menu"
        />
      </a-form-item>

      <a-divider>Komposisi Menu (Komponen)</a-divider>

      <a-alert
        message="Panduan"
        description="Menu terdiri dari komponen seperti Nasi, Ayam Goreng, Sambal, dll. Pilih dari daftar yang sudah tersedia."
        type="info"
        show-icon
        style="margin-bottom: 16px"
      />

      <!-- Item Selection -->
      <a-button type="dashed" block @click="showItemSelector" style="margin-bottom: 16px">
        <template #icon><PlusOutlined /></template>
        Tambah Komponen Menu
      </a-button>

      <!-- Items Table -->
      <a-table
        v-if="formData.items.length > 0"
        :columns="itemColumns"
        :data-source="formData.items"
        :pagination="false"
        size="small"
        row-key="semi_finished_goods_id"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'name'">
            <div>
              <strong>{{ record.name }}</strong>
              <br />
              <span class="text-muted">{{ record.category }}</span>
            </div>
          </template>
          <template v-else-if="column.key === 'quantity_small'">
            <a-input-number
              v-model:value="record.quantity_per_portion_small"
              :min="0"
              :step="10"
              style="width: 100%"
              placeholder="0"
              @change="calculateNutrition"
            />
          </template>
          <template v-else-if="column.key === 'quantity_large'">
            <a-input-number
              v-model:value="record.quantity_per_portion_large"
              :min="0"
              :step="10"
              style="width: 100%"
              placeholder="0"
              @change="calculateNutrition"
            />
          </template>
          <template v-else-if="column.key === 'nutrition'">
            <div style="font-size: 11px">
              {{ ((record.calories_per_100g || 0) * (record.quantity_per_portion_large || 0) / 100).toFixed(0) }} kkal
            </div>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button type="link" danger size="small" @click="removeItem(index)">
              Hapus
            </a-button>
          </template>
        </template>
      </a-table>

      <a-empty v-else description="Belum ada komponen menu" style="margin: 24px 0" />

      <a-divider>Informasi Gizi (Total)</a-divider>

      <!-- Nutrition Summary -->
      <a-row :gutter="16">
        <a-col :span="6">
          <a-statistic
            title="Kalori"
            :value="nutritionSummary.calories"
            suffix="kkal"
            :precision="0"
          />
        </a-col>
        <a-col :span="6">
          <a-statistic
            title="Protein"
            :value="nutritionSummary.protein"
            suffix="g"
            :precision="1"
          />
        </a-col>
        <a-col :span="6">
          <a-statistic
            title="Karbohidrat"
            :value="nutritionSummary.carbs"
            suffix="g"
            :precision="1"
          />
        </a-col>
        <a-col :span="6">
          <a-statistic
            title="Lemak"
            :value="nutritionSummary.fat"
            suffix="g"
            :precision="1"
          />
        </a-col>
      </a-row>

      <!-- Note about nutrition -->
      <a-alert
        message="Catatan"
        description="Nilai gizi di atas adalah total per menu/paket. Standar minimum: 600 kkal dan 15g protein per menu."
        type="info"
        show-icon
        style="margin-top: 16px"
      />

      <!-- Validation Alert -->
      <a-alert
        v-if="validationMessage"
        :type="validationMessage.type"
        :message="validationMessage.text"
        show-icon
        style="margin-top: 16px"
      />
    </a-form>

    <!-- Item Selector Modal -->
    <a-modal
      v-model:visible="itemSelectorVisible"
      title="Pilih Komponen"
      width="700px"
      @ok="addSelectedItems"
      ok-text="Tambah"
      cancel-text="Batal"
    >
      <a-row :gutter="16" style="margin-bottom: 16px">
        <a-col :span="12">
          <a-input-search
            v-model:value="itemSearch"
            placeholder="Cari komponen..."
            @search="searchItems"
          />
        </a-col>
        <a-col :span="12">
          <a-select
            v-model:value="categoryFilter"
            placeholder="Filter kategori"
            allow-clear
            style="width: 100%"
            @change="loadItems"
          >
            <a-select-option value="nasi">Nasi</a-select-option>
            <a-select-option value="lauk">Lauk</a-select-option>
            <a-select-option value="sambal">Sambal</a-select-option>
            <a-select-option value="sayur">Sayur</a-select-option>
            <a-select-option value="lauk_berkuah">Lauk Berkuah</a-select-option>
          </a-select>
        </a-col>
      </a-row>
      <a-table
        :columns="itemSelectorColumns"
        :data-source="availableItems"
        :loading="itemsLoading"
        :row-selection="{ selectedRowKeys: selectedItemIds, onChange: onItemSelectionChange }"
        :pagination="false"
        size="small"
        row-key="id"
        :scroll="{ y: 300 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'nutrition'">
            <div style="font-size: 11px">
              {{ record.calories_per_100g }}kkal | P:{{ record.protein_per_100g }}g
            </div>
          </template>
          <template v-else-if="column.key === 'stock'">
            <a-tag :color="record.stock_quantity > 0 ? 'green' : 'red'">
              {{ record.stock_quantity?.toFixed(2) }} {{ record.unit }}
            </a-tag>
          </template>
        </template>
      </a-table>
    </a-modal>
  </a-modal>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import recipeService from '@/services/recipeService'
import semiFinishedService from '@/services/semiFinishedService'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  recipe: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref()
const loading = ref(false)
const itemsLoading = ref(false)
const itemSelectorVisible = ref(false)
const itemSearch = ref('')
const categoryFilter = ref(undefined)
const availableItems = ref([])
const selectedItemIds = ref([])
const fileList = ref([])
const uploadedPhotoUrl = ref('')
const uploading = ref(false)

const isEdit = computed(() => !!props.recipe?.id)

const formData = reactive({
  name: '',
  category: undefined,
  instructions: '',
  items: []
})

const nutritionSummary = reactive({
  calories: 0,
  protein: 0,
  carbs: 0,
  fat: 0
})

const validationMessage = ref(null)

const rules = {
  name: [{ required: true, message: 'Nama menu harus diisi', trigger: 'blur' }],
  category: [{ required: true, message: 'Kategori harus dipilih', trigger: 'change' }],
  instructions: [{ required: true, message: 'Instruksi penyajian harus diisi', trigger: 'blur' }]
}

const itemColumns = [
  { title: 'Komponen', key: 'name', width: '25%' },
  { title: 'Porsi Kecil (gram)', key: 'quantity_small', width: '20%' },
  { title: 'Porsi Besar (gram)', key: 'quantity_large', width: '20%' },
  { title: 'Kontribusi Gizi', key: 'nutrition', width: '20%' },
  { title: 'Aksi', key: 'actions', width: '15%', align: 'center' }
]

const itemSelectorColumns = [
  { title: 'Nama', dataIndex: 'name', key: 'name', width: '30%' },
  { title: 'Kategori', dataIndex: 'category', key: 'category', width: '15%' },
  { title: 'Gizi per 100g', key: 'nutrition', width: '25%' },
  { title: 'Stok', key: 'stock', width: '20%' }
]

const calculateNutrition = () => {
  let totalCalories = 0
  let totalProtein = 0
  let totalCarbs = 0
  let totalFat = 0

  formData.items.forEach(item => {
    // Use large portion as base for nutrition calculation
    const quantity = item.quantity_per_portion_large || item.quantity || 0
    const factor = quantity / 100
    totalCalories += (item.calories_per_100g || 0) * factor
    totalProtein += (item.protein_per_100g || 0) * factor
    totalCarbs += (item.carbs_per_100g || 0) * factor
    totalFat += (item.fat_per_100g || 0) * factor
  })

  nutritionSummary.calories = totalCalories
  nutritionSummary.protein = totalProtein
  nutritionSummary.carbs = totalCarbs
  nutritionSummary.fat = totalFat

  validateNutrition()
}

const validateNutrition = () => {
  // Minimum standards per menu
  const minCalories = 600
  const minProtein = 15

  if (nutritionSummary.calories < minCalories || nutritionSummary.protein < minProtein) {
    validationMessage.value = {
      type: 'warning',
      text: `Peringatan: Gizi menu belum memenuhi standar minimum (${minCalories} kkal, ${minProtein}g protein)`
    }
  } else {
    validationMessage.value = {
      type: 'success',
      text: 'Gizi menu sudah memenuhi standar minimum'
    }
  }
}

const showItemSelector = async () => {
  itemSelectorVisible.value = true
  await loadItems()
}

const loadItems = async () => {
  itemsLoading.value = true
  try {
    const params = {
      search: itemSearch.value || undefined,
      category: categoryFilter.value || undefined
    }
    const response = await semiFinishedService.getAllSemiFinishedGoods(params)
    availableItems.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat data komponen')
    console.error('Error loading items:', error)
  } finally {
    itemsLoading.value = false
  }
}

const searchItems = () => {
  loadItems()
}

const onItemSelectionChange = (selectedKeys) => {
  selectedItemIds.value = selectedKeys
}

const addSelectedItems = () => {
  const newItems = availableItems.value
    .filter(item => selectedItemIds.value.includes(item.id))
    .filter(item => !formData.items.some(existing => existing.semi_finished_goods_id === item.id))
    .map(item => ({
      semi_finished_goods_id: item.id,
      name: item.name,
      category: item.category,
      quantity: 100, // deprecated, kept for backward compatibility
      quantity_per_portion_small: 0, // default 0, user will fill
      quantity_per_portion_large: 0, // default 0, user will fill
      unit: item.unit,
      calories_per_100g: item.calories_per_100g,
      protein_per_100g: item.protein_per_100g,
      carbs_per_100g: item.carbs_per_100g,
      fat_per_100g: item.fat_per_100g
    }))

  formData.items.push(...newItems)
  calculateNutrition()
  
  itemSelectorVisible.value = false
  selectedItemIds.value = []
}

const removeItem = (index) => {
  formData.items.splice(index, 1)
  calculateNutrition()
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    message.error('Hanya file gambar yang diperbolehkan!')
    return false
  }
  
  const isLt2M = file.size / 1024 / 1024 < 2
  if (!isLt2M) {
    message.error('Ukuran gambar harus kurang dari 2MB!')
    return false
  }
  
  return true
}

const handleUpload = async ({ file, onSuccess, onError }) => {
  uploading.value = true
  
  try {
    const formData = new FormData()
    formData.append('photo', file)
    
    const token = localStorage.getItem('token')
    const response = await fetch('http://localhost:8080/api/v1/recipes/upload-photo', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      },
      body: formData
    })
    
    const data = await response.json()
    
    if (data.success) {
      uploadedPhotoUrl.value = `http://localhost:8080${data.photo_url}`
      fileList.value = [{
        uid: file.uid,
        name: file.name,
        status: 'done',
        url: uploadedPhotoUrl.value
      }]
      message.success('Foto berhasil diupload')
      onSuccess(data, file)
    } else {
      message.error(data.message || 'Gagal upload foto')
      onError(new Error(data.message))
    }
  } catch (error) {
    console.error('Upload error:', error)
    message.error('Gagal upload foto')
    onError(error)
  } finally {
    uploading.value = false
  }
}

const handleRemovePhoto = async () => {
  if (uploadedPhotoUrl.value) {
    try {
      const token = localStorage.getItem('token')
      await fetch(`http://localhost:8080/api/v1/recipes/delete-photo?photo_url=${encodeURIComponent(uploadedPhotoUrl.value)}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
    } catch (error) {
      console.error('Delete photo error:', error)
    }
  }
  
  uploadedPhotoUrl.value = ''
  fileList.value = []
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    if (formData.items.length === 0) {
      message.warning('Minimal harus ada 1 komponen menu')
      return
    }

    loading.value = true

    const payload = {
      name: formData.name,
      category: formData.category,
      photo_url: uploadedPhotoUrl.value || undefined,
      instructions: formData.instructions,
      is_active: true,
      total_calories: nutritionSummary.calories,
      total_protein: nutritionSummary.protein,
      total_carbs: nutritionSummary.carbs,
      total_fat: nutritionSummary.fat,
      items: formData.items.map(item => ({
        semi_finished_goods_id: item.semi_finished_goods_id,
        quantity: item.quantity_per_portion_large || item.quantity || 0, // fallback for backward compatibility
        quantity_per_portion_small: item.quantity_per_portion_small || 0,
        quantity_per_portion_large: item.quantity_per_portion_large || 0
      }))
    }

    if (isEdit.value) {
      await recipeService.updateRecipe(props.recipe.id, payload)
      message.success('Menu berhasil diperbarui')
    } else {
      await recipeService.createRecipe(payload)
      message.success('Menu berhasil ditambahkan')
    }

    emit('success')
  } catch (error) {
    if (error.errorFields) {
      message.error('Mohon lengkapi semua field yang wajib diisi')
    } else if (error.response?.data?.error_code === 'INSUFFICIENT_NUTRITION') {
      message.error('Nilai gizi tidak memenuhi standar minimum (600 kkal, 15g protein per menu)')
    } else {
      message.error('Gagal menyimpan menu')
      console.error('Error saving recipe:', error)
    }
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
}

const resetForm = () => {
  formData.name = ''
  formData.category = undefined
  formData.instructions = ''
  formData.items = []
  nutritionSummary.calories = 0
  nutritionSummary.protein = 0
  nutritionSummary.carbs = 0
  nutritionSummary.fat = 0
  validationMessage.value = null
  fileList.value = []
  uploadedPhotoUrl.value = ''
  formRef.value?.resetFields()
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    if (props.recipe) {
      // Load recipe data for editing
      Object.assign(formData, {
        name: props.recipe.name,
        category: props.recipe.category,
        instructions: props.recipe.instructions,
        items: props.recipe.recipe_items?.map(ri => ({
          semi_finished_goods_id: ri.semi_finished_goods_id,
          name: ri.semi_finished_goods?.name,
          category: ri.semi_finished_goods?.category,
          unit: ri.semi_finished_goods?.unit,
          quantity: ri.quantity,
          quantity_per_portion_small: ri.quantity_per_portion_small || 0,
          quantity_per_portion_large: ri.quantity_per_portion_large || 0,
          calories_per_100g: ri.semi_finished_goods?.calories_per_100g,
          protein_per_100g: ri.semi_finished_goods?.protein_per_100g,
          carbs_per_100g: ri.semi_finished_goods?.carbs_per_100g,
          fat_per_100g: ri.semi_finished_goods?.fat_per_100g
        })) || []
      })
      
      // Load existing photo
      if (props.recipe.photo_url) {
        uploadedPhotoUrl.value = props.recipe.photo_url
        fileList.value = [{
          uid: '-1',
          name: 'photo.jpg',
          status: 'done',
          url: props.recipe.photo_url
        }]
      }
      
      calculateNutrition()
    } else {
      resetForm()
    }
  }
})
</script>

<style scoped>
.text-muted {
  color: #8c8c8c;
  font-size: 11px;
}

.per-portion-summary {
  padding: 12px;
  background: #f6ffed;
  border: 1px solid #b7eb8f;
  border-radius: 4px;
}

.per-portion-summary strong {
  margin-right: 8px;
}

.per-portion-summary .ant-tag {
  margin: 0 4px;
}

:deep(.ant-upload-picture-card-wrapper) {
  width: auto;
}

:deep(.ant-upload.ant-upload-select-picture-card) {
  width: 200px;
  height: 200px;
  margin: 0;
}

:deep(.ant-upload-list-picture-card-container) {
  width: 200px;
  height: 200px;
  margin: 0;
}

:deep(.ant-upload-list-picture-card .ant-upload-list-item) {
  padding: 0;
}
</style>
