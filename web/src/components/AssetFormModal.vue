<template>
  <a-modal
    :open="visible"
    :title="asset ? 'Edit Aset' : 'Tambah Aset'"
    :confirm-loading="submitting"
    @ok="handleSubmit"
    @cancel="handleCancel"
    width="700px"
  >
    <a-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      layout="vertical"
    >
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Kode Aset" name="asset_code">
            <a-input v-model:value="formData.asset_code" placeholder="Contoh: AST-001" />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Nama Aset" name="name">
            <a-input v-model:value="formData.name" placeholder="Nama aset" />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Kategori" name="category">
            <a-select v-model:value="formData.category" placeholder="Pilih kategori">
              <a-select-option value="Peralatan Masak">Peralatan Masak</a-select-option>
              <a-select-option value="Peralatan Packing">Peralatan Packing</a-select-option>
              <a-select-option value="Elektronik">Elektronik</a-select-option>
              <a-select-option value="Furniture">Furniture</a-select-option>
              <a-select-option value="Kendaraan">Kendaraan</a-select-option>
              <a-select-option value="Lainnya">Lainnya</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Kondisi" name="condition">
            <a-select v-model:value="formData.condition" placeholder="Pilih kondisi">
              <a-select-option value="good">Baik</a-select-option>
              <a-select-option value="fair">Cukup</a-select-option>
              <a-select-option value="poor">Buruk</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Tanggal Pembelian" name="purchase_date">
            <a-date-picker 
              v-model:value="formData.purchase_date" 
              style="width: 100%" 
              placeholder="Pilih tanggal pembelian"
              format="DD/MM/YYYY"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Harga Pembelian" name="purchase_price">
            <a-input-number
              v-model:value="formData.purchase_price"
              style="width: 100%"
              :min="0"
              :precision="0"
              placeholder="0"
              :formatter="value => `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
              :parser="value => value.replace(/Rp\s?|(,*)/g, '')"
            />
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Tingkat Depresiasi (%/tahun)" name="depreciation_rate">
            <a-input-number
              v-model:value="formData.depreciation_rate"
              style="width: 100%"
              :min="0"
              :max="100"
              :precision="2"
              placeholder="0"
              suffix="%"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Lokasi" name="location">
            <a-input v-model:value="formData.location" placeholder="Lokasi aset" />
          </a-form-item>
        </a-col>
      </a-row>

      <!-- Preview Calculations -->
      <a-card v-if="formData.purchase_price && formData.depreciation_rate" size="small" title="Perkiraan Nilai">
        <a-row :gutter="16">
          <a-col :span="8">
            <a-statistic
              title="Nilai Pembelian"
              :value="formData.purchase_price"
              :precision="0"
              suffix="IDR"
            />
          </a-col>
          <a-col :span="8">
            <a-statistic
              title="Depresiasi/Tahun"
              :value="calculateAnnualDepreciation()"
              :precision="0"
              suffix="IDR"
            />
          </a-col>
          <a-col :span="8">
            <a-statistic
              title="Nilai Setelah 5 Tahun"
              :value="calculateFutureValue(5)"
              :precision="0"
              suffix="IDR"
            />
          </a-col>
        </a-row>
      </a-card>
    </a-form>
  </a-modal>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import assetService from '@/services/assetService'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  asset: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'success'])

const submitting = ref(false)
const formRef = ref()

const formData = reactive({
  asset_code: '',
  name: '',
  category: '',
  condition: 'good',
  purchase_date: null,
  purchase_price: 0,
  depreciation_rate: 0,
  location: ''
})

const rules = {
  asset_code: [{ required: true, message: 'Kode aset wajib diisi' }],
  name: [{ required: true, message: 'Nama aset wajib diisi' }],
  category: [{ required: true, message: 'Kategori wajib dipilih' }],
  condition: [{ required: true, message: 'Kondisi wajib dipilih' }],
  purchase_date: [{ required: true, message: 'Tanggal pembelian wajib diisi' }],
  purchase_price: [{ required: true, message: 'Harga pembelian wajib diisi' }],
  depreciation_rate: [{ required: true, message: 'Tingkat depresiasi wajib diisi' }]
}

// Watch for asset prop changes to populate form
watch(() => props.asset, (newAsset) => {
  if (newAsset) {
    Object.assign(formData, {
      asset_code: newAsset.asset_code,
      name: newAsset.name,
      category: newAsset.category,
      condition: newAsset.condition,
      purchase_date: newAsset.purchase_date ? dayjs(newAsset.purchase_date) : null,
      purchase_price: newAsset.purchase_price,
      depreciation_rate: newAsset.depreciation_rate,
      location: newAsset.location
    })
  } else {
    resetForm()
  }
}, { immediate: true })

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    const submitData = {
      ...formData,
      purchase_date: formData.purchase_date ? formData.purchase_date.format('YYYY-MM-DD') : null
    }

    if (props.asset) {
      await assetService.updateAsset(props.asset.id, submitData)
      message.success('Aset berhasil diperbarui')
    } else {
      await assetService.createAsset(submitData)
      message.success('Aset berhasil ditambahkan')
    }

    emit('success')
  } catch (error) {
    if (error.errorFields) {
      return
    }
    
    const errorMessage = error.response?.data?.message || 'Gagal menyimpan data aset'
    message.error(errorMessage)
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    asset_code: '',
    name: '',
    category: '',
    condition: 'good',
    purchase_date: null,
    purchase_price: 0,
    depreciation_rate: 0,
    location: ''
  })
  formRef.value?.resetFields()
}

const calculateAnnualDepreciation = () => {
  if (!formData.purchase_price || !formData.depreciation_rate) return 0
  return formData.purchase_price * (formData.depreciation_rate / 100)
}

const calculateFutureValue = (years) => {
  if (!formData.purchase_price || !formData.depreciation_rate) return 0
  const totalDepreciation = formData.purchase_price * (formData.depreciation_rate / 100) * years
  const futureValue = formData.purchase_price - totalDepreciation
  return Math.max(0, futureValue)
}
</script>