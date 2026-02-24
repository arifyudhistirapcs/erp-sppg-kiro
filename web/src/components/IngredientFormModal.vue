<template>
  <a-modal
    :visible="visible"
    title="Tambah Bahan Baru"
    width="600px"
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
          <a-form-item label="Nama Bahan" name="name">
            <a-input 
              v-model:value="formData.name" 
              placeholder="Masukkan nama bahan"
            />
          </a-form-item>
        </a-col>
        <a-col :span="8">
          <a-form-item label="Satuan" name="unit">
            <a-select v-model:value="formData.unit" placeholder="Pilih satuan">
              <a-select-option value="kg">kg (Kilogram)</a-select-option>
              <a-select-option value="gram">gram</a-select-option>
              <a-select-option value="liter">liter</a-select-option>
              <a-select-option value="ml">ml (Mililiter)</a-select-option>
              <a-select-option value="pcs">pcs (Pieces)</a-select-option>
              <a-select-option value="bungkus">bungkus</a-select-option>
              <a-select-option value="kaleng">kaleng</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>
    </a-form>
  </a-modal>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import recipeService from '@/services/recipeService'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  }
})

const emit = defineEmits(['update:visible', 'success'])

const formRef = ref()
const loading = ref(false)

const formData = reactive({
  name: '',
  unit: undefined
})

const rules = {
  name: [
    { required: true, message: 'Nama bahan harus diisi', trigger: 'blur' },
    { min: 2, message: 'Minimal 2 karakter', trigger: 'blur' }
  ],
  unit: [
    { required: true, message: 'Satuan harus dipilih', trigger: 'change' }
  ]
}

// Define resetForm before it's used
const resetForm = () => {
  formData.name = ''
  formData.unit = undefined
  formRef.value?.resetFields()
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    loading.value = true

    const payload = {
      name: formData.name,
      unit: formData.unit
    }

    const response = await recipeService.createIngredient(payload)
    message.success('Bahan berhasil ditambahkan')
    
    // Emit the created ingredient data from API response
    emit('success', response.data.data || payload)
    resetForm()
  } catch (error) {
    if (error.errorFields) {
      message.error('Mohon lengkapi semua field yang wajib diisi')
    } else {
      message.error('Gagal menambahkan bahan')
      console.error('Error creating ingredient:', error)
    }
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
  resetForm()
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    resetForm()
  }
})
</script>
