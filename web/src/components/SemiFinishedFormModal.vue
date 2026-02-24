<template>
  <a-modal
    :visible="visible"
    :title="isEdit ? 'Edit Barang Setengah Jadi' : 'Tambah Barang Setengah Jadi Baru'"
    @ok="handleSubmit"
    @cancel="handleCancel"
    :confirm-loading="submitting"
    width="800px"
  >
    <a-form :model="form" layout="vertical" ref="formRef">
      <!-- Basic Info -->
      <a-divider orientation="left">Informasi Dasar</a-divider>
      
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Nama" required>
            <a-input
              v-model:value="form.name"
              placeholder="Contoh: Nasi Putih, Ayam Goreng"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Kategori" required>
            <a-select
              v-model:value="form.category"
              placeholder="Pilih kategori"
            >
              <a-select-option value="nasi">Nasi</a-select-option>
              <a-select-option value="lauk">Lauk</a-select-option>
              <a-select-option value="sambal">Sambal</a-select-option>
              <a-select-option value="sayur">Sayur</a-select-option>
              <a-select-option value="lauk_berkuah">Lauk Berkuah</a-select-option>
              <a-select-option value="lainnya">Lainnya</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Satuan" required>
            <a-select
              v-model:value="form.unit"
              placeholder="Pilih satuan"
            >
              <a-select-option value="kg">Kilogram (kg)</a-select-option>
              <a-select-option value="gram">Gram (g)</a-select-option>
              <a-select-option value="liter">Liter (L)</a-select-option>
              <a-select-option value="ml">Mililiter (ml)</a-select-option>
              <a-select-option value="pcs">Pieces (pcs)</a-select-option>
              <a-select-option value="porsi">Porsi</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Deskripsi">
            <a-textarea
              v-model:value="form.description"
              rows="2"
              placeholder="Deskripsi singkat (opsional)"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <!-- Recipe Info -->
      <a-divider orientation="left">Resep Produksi</a-divider>
      
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Nama Resep" required>
            <a-input
              v-model:value="form.recipe.name"
              placeholder="Contoh: Resep Nasi Putih"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Yield (Hasil)" required>
            <a-input-number
              v-model:value="form.recipe.yield_amount"
              :min="0.1"
              :step="0.1"
              style="width: 100%"
              :addon-after="form.unit || 'unit'"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-form-item label="Cara Pembuatan">
        <a-textarea
          v-model:value="form.recipe.instructions"
          rows="3"
          placeholder="Langkah-langkah pembuatan..."
        />
      </a-form-item>

      <!-- Ingredients -->
      <a-divider orientation="left">Bahan Baku</a-divider>
      
      <div
        v-for="(ing, index) in form.ingredients"
        :key="index"
        class="ingredient-row"
      >
        <a-row :gutter="8" align="middle">
          <a-col :span="12">
            <a-form-item :label="index === 0 ? 'Bahan' : ''" required>
              <a-select
                v-model:value="ing.ingredient_id"
                placeholder="Pilih bahan"
                show-search
                :filter-option="filterIngredient"
                :options="ingredientOptions"
                :field-names="{ label: 'name', value: 'id' }"
              />
            </a-form-item>
          </a-col>
          <a-col :span="8">
            <a-form-item :label="index === 0 ? 'Jumlah' : ''" required>
              <a-input-number
                v-model:value="ing.quantity"
                :min="0.1"
                :step="0.1"
                style="width: 100%"
              />
            </a-form-item>
          </a-col>
          <a-col :span="4">
            <a-form-item :label="index === 0 ? '' : ''">
              <a-button
                type="danger"
                @click="removeIngredient(index)"
                :disabled="form.ingredients.length <= 1"
              >
                <DeleteOutlined />
              </a-button>
            </a-form-item>
          </a-col>
        </a-row>
      </div>

      <a-button type="dashed" block @click="addIngredient" class="mt-2">
        <PlusOutlined />
        Tambah Bahan
      </a-button>

      <!-- Nutrition Summary -->
      <a-divider orientation="left">Informasi Gizi (per 100g)</a-divider>
      
      <a-row :gutter="16">
        <a-col :span="6">
          <a-form-item label="Kalori (kkal)" required>
            <a-input-number
              v-model:value="form.calories_per_100g"
              :min="0"
              :step="0.1"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="6">
          <a-form-item label="Protein (g)" required>
            <a-input-number
              v-model:value="form.protein_per_100g"
              :min="0"
              :step="0.1"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="6">
          <a-form-item label="Karbohidrat (g)" required>
            <a-input-number
              v-model:value="form.carbs_per_100g"
              :min="0"
              :step="0.1"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="6">
          <a-form-item label="Lemak (g)" required>
            <a-input-number
              v-model:value="form.fat_per_100g"
              :min="0"
              :step="0.1"
              style="width: 100%"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-alert
        message="Perhitungan Gizi"
        description="Informasi gizi di atas akan digunakan untuk menghitung total nutrisi menu makanan. Pastikan data sudah benar."
        type="info"
        show-icon
        class="mt-4"
      />
    </a-form>
  </a-modal>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import semiFinishedService from '@/services/semiFinishedService'
import recipeService from '@/services/recipeService'

const props = defineProps({
  visible: Boolean,
  editData: Object
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref(null)
const submitting = ref(false)
const ingredientsList = ref([])

const isEdit = ref(false)

const form = ref({
  name: '',
  unit: 'kg',
  category: 'nasi',
  description: '',
  calories_per_100g: 0,
  protein_per_100g: 0,
  carbs_per_100g: 0,
  fat_per_100g: 0,
  recipe: {
    name: '',
    instructions: '',
    yield_amount: 1
  },
  ingredients: [
    { ingredient_id: null, quantity: 1 }
  ]
})

const ingredientOptions = computed(() => {
  return ingredientsList.value.map(ing => ({
    id: ing.id,
    name: `${ing.name} (${ing.unit})`
  }))
})

const fetchIngredients = async () => {
  try {
    const response = await recipeService.getIngredients()
    ingredientsList.value = response.data.data || []
  } catch (error) {
    console.error('Error fetching ingredients:', error)
  }
}

const filterIngredient = (input, option) => {
  return option.name.toLowerCase().includes(input.toLowerCase())
}

const addIngredient = () => {
  form.value.ingredients.push({ ingredient_id: null, quantity: 1 })
}

const removeIngredient = (index) => {
  form.value.ingredients.splice(index, 1)
}

const resetForm = () => {
  form.value = {
    name: '',
    unit: 'kg',
    category: 'nasi',
    description: '',
    calories_per_100g: 0,
    protein_per_100g: 0,
    carbs_per_100g: 0,
    fat_per_100g: 0,
    recipe: {
      name: '',
      instructions: '',
      yield_amount: 1
    },
    ingredients: [
      { ingredient_id: null, quantity: 1 }
    ]
  }
  isEdit.value = false
}

const loadEditData = () => {
  if (!props.editData) return
  
  const data = props.editData
  form.value = {
    name: data.name,
    unit: data.unit,
    category: data.category,
    description: data.description,
    calories_per_100g: data.calories_per_100g,
    protein_per_100g: data.protein_per_100g,
    carbs_per_100g: data.carbs_per_100g,
    fat_per_100g: data.fat_per_100g,
    recipe: {
      name: data.recipe?.name || '',
      instructions: data.recipe?.instructions || '',
      yield_amount: data.recipe?.yield_amount || 1
    },
    ingredients: data.recipe?.ingredients?.map(ing => ({
      ingredient_id: ing.ingredient_id,
      quantity: ing.quantity
    })) || [{ ingredient_id: null, quantity: 1 }]
  }
  isEdit.value = true
}

const handleSubmit = async () => {
  // Validate
  if (!form.value.name || !form.value.unit || !form.value.category) {
    message.error('Mohon lengkapi informasi dasar')
    return
  }
  
  if (!form.value.recipe.name || !form.value.recipe.yield_amount) {
    message.error('Mohon lengkapi informasi resep')
    return
  }
  
  const validIngredients = form.value.ingredients.filter(
    ing => ing.ingredient_id && ing.quantity > 0
  )
  
  if (validIngredients.length === 0) {
    message.error('Minimal harus ada 1 bahan baku')
    return
  }

  submitting.value = true
  try {
    const payload = {
      ...form.value,
      ingredients: validIngredients
    }

    if (isEdit.value && props.editData) {
      await semiFinishedService.updateSemiFinishedGoods(props.editData.id, payload)
      message.success('Barang setengah jadi berhasil diperbarui')
    } else {
      await semiFinishedService.createSemiFinishedGoods(payload)
      message.success('Barang setengah jadi berhasil dibuat')
    }
    
    emit('success')
    handleCancel()
  } catch (error) {
    message.error('Gagal menyimpan data')
    console.error('Error submitting form:', error)
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
  resetForm()
}

watch(() => props.visible, (val) => {
  if (val) {
    fetchIngredients()
    if (props.editData) {
      loadEditData()
    } else {
      resetForm()
    }
  }
})

watch(() => props.editData, (val) => {
  if (val && props.visible) {
    loadEditData()
  }
})

onMounted(() => {
  fetchIngredients()
})
</script>

<style scoped>
.ingredient-row {
  margin-bottom: 8px;
}

.mt-2 {
  margin-top: 8px;
}

.mt-4 {
  margin-top: 16px;
}
</style>
