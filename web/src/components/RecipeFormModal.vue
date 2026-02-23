<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? 'Edit Resep' : 'Tambah Resep Baru'"
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
          <a-form-item label="Nama Resep" name="name">
            <a-input v-model:value="formData.name" placeholder="Masukkan nama resep" />
          </a-form-item>
        </a-col>
        <a-col :span="8">
          <a-form-item label="Kategori" name="category">
            <a-select v-model:value="formData.category" placeholder="Pilih kategori">
              <a-select-option value="makanan_pokok">Makanan Pokok</a-select-option>
              <a-select-option value="lauk_pauk">Lauk Pauk</a-select-option>
              <a-select-option value="sayuran">Sayuran</a-select-option>
              <a-select-option value="buah">Buah</a-select-option>
              <a-select-option value="minuman">Minuman</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Jumlah Porsi" name="serving_size">
            <a-input-number
              v-model:value="formData.serving_size"
              :min="1"
              style="width: 100%"
              placeholder="Jumlah porsi"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Status" name="is_active">
            <a-switch v-model:checked="formData.is_active" checked-children="Aktif" un-checked-children="Nonaktif" />
          </a-form-item>
        </a-col>
      </a-row>

      <a-form-item label="Instruksi Memasak" name="instructions">
        <a-textarea
          v-model:value="formData.instructions"
          :rows="4"
          placeholder="Masukkan langkah-langkah memasak"
        />
      </a-form-item>

      <a-divider>Bahan-Bahan</a-divider>

      <!-- Ingredient Selection -->
      <a-form-item>
        <a-button type="dashed" block @click="showIngredientSelector">
          <template #icon><PlusOutlined /></template>
          Tambah Bahan
        </a-button>
      </a-form-item>

      <!-- Ingredients Table -->
      <a-table
        v-if="formData.ingredients.length > 0"
        :columns="ingredientColumns"
        :data-source="formData.ingredients"
        :pagination="false"
        size="small"
        row-key="ingredient_id"
      >
        <template #bodyCell="{ column, record, index }">
          <template v-if="column.key === 'name'">
            {{ record.name }}
          </template>
          <template v-else-if="column.key === 'quantity'">
            <a-input-number
              v-model:value="record.quantity"
              :min="0.01"
              :step="0.1"
              style="width: 100%"
              @change="calculateNutrition"
            />
          </template>
          <template v-else-if="column.key === 'unit'">
            {{ record.unit }}
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-button type="link" danger size="small" @click="removeIngredient(index)">
              Hapus
            </a-button>
          </template>
        </template>
      </a-table>

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

      <!-- Validation Alert -->
      <a-alert
        v-if="validationMessage"
        :type="validationMessage.type"
        :message="validationMessage.text"
        show-icon
        style="margin-top: 16px"
      />
    </a-form>

    <!-- Ingredient Selector Modal -->
    <a-modal
      v-model:visible="ingredientSelectorVisible"
      title="Pilih Bahan"
      width="600px"
      @ok="addSelectedIngredients"
      ok-text="Tambah"
      cancel-text="Batal"
    >
      <a-input-search
        v-model:value="ingredientSearch"
        placeholder="Cari bahan..."
        style="margin-bottom: 16px"
        @search="searchIngredients"
      />
      <a-table
        :columns="ingredientSelectorColumns"
        :data-source="availableIngredients"
        :loading="ingredientsLoading"
        :row-selection="{ selectedRowKeys: selectedIngredientIds, onChange: onIngredientSelectionChange }"
        :pagination="false"
        size="small"
        row-key="id"
        :scroll="{ y: 300 }"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'nutrition'">
            <div style="font-size: 11px">
              {{ record.calories_per_100g }}kkal | P:{{ record.protein_per_100g }}g | K:{{ record.carbs_per_100g }}g | L:{{ record.fat_per_100g }}g
            </div>
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
const ingredientsLoading = ref(false)
const ingredientSelectorVisible = ref(false)
const ingredientSearch = ref('')
const availableIngredients = ref([])
const selectedIngredientIds = ref([])

const isEdit = computed(() => !!props.recipe?.id)

const formData = reactive({
  name: '',
  category: undefined,
  serving_size: 1,
  instructions: '',
  is_active: true,
  ingredients: []
})

const nutritionSummary = reactive({
  calories: 0,
  protein: 0,
  carbs: 0,
  fat: 0
})

const validationMessage = ref(null)

const rules = {
  name: [{ required: true, message: 'Nama resep harus diisi', trigger: 'blur' }],
  category: [{ required: true, message: 'Kategori harus dipilih', trigger: 'change' }],
  serving_size: [{ required: true, message: 'Jumlah porsi harus diisi', trigger: 'blur' }],
  instructions: [{ required: true, message: 'Instruksi memasak harus diisi', trigger: 'blur' }]
}

const ingredientColumns = [
  { title: 'Bahan', key: 'name', dataIndex: 'name', width: '40%' },
  { title: 'Jumlah', key: 'quantity', width: '25%' },
  { title: 'Satuan', key: 'unit', dataIndex: 'unit', width: '20%' },
  { title: 'Aksi', key: 'actions', width: '15%', align: 'center' }
]

const ingredientSelectorColumns = [
  { title: 'Nama Bahan', dataIndex: 'name', key: 'name', width: '40%' },
  { title: 'Satuan', dataIndex: 'unit', key: 'unit', width: '20%' },
  { title: 'Gizi per 100g', key: 'nutrition', width: '40%' }
]

const calculateNutrition = () => {
  let totalCalories = 0
  let totalProtein = 0
  let totalCarbs = 0
  let totalFat = 0

  formData.ingredients.forEach(ing => {
    const factor = ing.quantity / 100
    totalCalories += (ing.calories_per_100g || 0) * factor
    totalProtein += (ing.protein_per_100g || 0) * factor
    totalCarbs += (ing.carbs_per_100g || 0) * factor
    totalFat += (ing.fat_per_100g || 0) * factor
  })

  nutritionSummary.calories = totalCalories
  nutritionSummary.protein = totalProtein
  nutritionSummary.carbs = totalCarbs
  nutritionSummary.fat = totalFat

  validateNutrition()
}

const validateNutrition = () => {
  // Minimum standards per portion (example values)
  const minCaloriesPerPortion = 600
  const minProteinPerPortion = 15

  const caloriesPerPortion = nutritionSummary.calories / formData.serving_size
  const proteinPerPortion = nutritionSummary.protein / formData.serving_size

  if (caloriesPerPortion < minCaloriesPerPortion || proteinPerPortion < minProteinPerPortion) {
    validationMessage.value = {
      type: 'warning',
      text: `Peringatan: Gizi per porsi belum memenuhi standar minimum (${minCaloriesPerPortion} kkal, ${minProteinPerPortion}g protein)`
    }
  } else {
    validationMessage.value = {
      type: 'success',
      text: 'Gizi per porsi sudah memenuhi standar minimum'
    }
  }
}

const showIngredientSelector = async () => {
  ingredientSelectorVisible.value = true
  await loadIngredients()
}

const loadIngredients = async () => {
  ingredientsLoading.value = true
  try {
    const response = await recipeService.getIngredients({ search: ingredientSearch.value })
    availableIngredients.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat data bahan')
    console.error('Error loading ingredients:', error)
  } finally {
    ingredientsLoading.value = false
  }
}

const searchIngredients = () => {
  loadIngredients()
}

const onIngredientSelectionChange = (selectedKeys) => {
  selectedIngredientIds.value = selectedKeys
}

const addSelectedIngredients = () => {
  const newIngredients = availableIngredients.value
    .filter(ing => selectedIngredientIds.value.includes(ing.id))
    .filter(ing => !formData.ingredients.some(existing => existing.ingredient_id === ing.id))
    .map(ing => ({
      ingredient_id: ing.id,
      name: ing.name,
      unit: ing.unit,
      quantity: 100,
      calories_per_100g: ing.calories_per_100g,
      protein_per_100g: ing.protein_per_100g,
      carbs_per_100g: ing.carbs_per_100g,
      fat_per_100g: ing.fat_per_100g
    }))

  formData.ingredients.push(...newIngredients)
  calculateNutrition()
  
  ingredientSelectorVisible.value = false
  selectedIngredientIds.value = []
}

const removeIngredient = (index) => {
  formData.ingredients.splice(index, 1)
  calculateNutrition()
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    if (formData.ingredients.length === 0) {
      message.warning('Minimal harus ada 1 bahan')
      return
    }

    loading.value = true

    const payload = {
      name: formData.name,
      category: formData.category,
      serving_size: formData.serving_size,
      instructions: formData.instructions,
      is_active: formData.is_active,
      total_calories: nutritionSummary.calories,
      total_protein: nutritionSummary.protein,
      total_carbs: nutritionSummary.carbs,
      total_fat: nutritionSummary.fat,
      recipe_ingredients: formData.ingredients.map(ing => ({
        ingredient_id: ing.ingredient_id,
        quantity: ing.quantity
      }))
    }

    if (isEdit.value) {
      await recipeService.updateRecipe(props.recipe.id, payload)
      message.success('Resep berhasil diperbarui')
    } else {
      await recipeService.createRecipe(payload)
      message.success('Resep berhasil ditambahkan')
    }

    emit('success')
  } catch (error) {
    if (error.errorFields) {
      message.error('Mohon lengkapi semua field yang wajib diisi')
    } else {
      message.error('Gagal menyimpan resep')
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
  formData.serving_size = 1
  formData.instructions = ''
  formData.is_active = true
  formData.ingredients = []
  nutritionSummary.calories = 0
  nutritionSummary.protein = 0
  nutritionSummary.carbs = 0
  nutritionSummary.fat = 0
  validationMessage.value = null
  formRef.value?.resetFields()
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    if (props.recipe) {
      // Load recipe data for editing
      Object.assign(formData, {
        name: props.recipe.name,
        category: props.recipe.category,
        serving_size: props.recipe.serving_size,
        instructions: props.recipe.instructions,
        is_active: props.recipe.is_active,
        ingredients: props.recipe.recipe_ingredients?.map(ri => ({
          ingredient_id: ri.ingredient_id,
          name: ri.ingredient?.name,
          unit: ri.ingredient?.unit,
          quantity: ri.quantity,
          calories_per_100g: ri.ingredient?.calories_per_100g,
          protein_per_100g: ri.ingredient?.protein_per_100g,
          carbs_per_100g: ri.ingredient?.carbs_per_100g,
          fat_per_100g: ri.ingredient?.fat_per_100g
        })) || []
      })
      calculateNutrition()
    } else {
      resetForm()
    }
  }
})
</script>
