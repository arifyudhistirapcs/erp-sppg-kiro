<template>
  <a-modal
    :open="visible"
    :title="'Tambah Transaksi'"
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
      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Tanggal" name="date">
            <a-date-picker 
              v-model:value="formData.date" 
              style="width: 100%" 
              placeholder="Pilih tanggal transaksi"
              format="DD/MM/YYYY"
            />
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Tipe Transaksi" name="type">
            <a-select v-model:value="formData.type" placeholder="Pilih tipe transaksi">
              <a-select-option value="income">
                <span style="color: #52c41a">📈 Pemasukan</span>
              </a-select-option>
              <a-select-option value="expense">
                <span style="color: #ff4d4f">📉 Pengeluaran</span>
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
      </a-row>

      <a-row :gutter="16">
        <a-col :span="12">
          <a-form-item label="Kategori" name="category">
            <a-select v-model:value="formData.category" placeholder="Pilih kategori">
              <a-select-option value="bahan_baku">🥬 Bahan Baku</a-select-option>
              <a-select-option value="gaji">💰 Gaji</a-select-option>
              <a-select-option value="utilitas">⚡ Utilitas</a-select-option>
              <a-select-option value="operasional">🏢 Operasional</a-select-option>
            </a-select>
          </a-form-item>
        </a-col>
        <a-col :span="12">
          <a-form-item label="Jumlah" name="amount">
            <a-input-number
              v-model:value="formData.amount"
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

      <a-form-item label="Deskripsi" name="description">
        <a-textarea 
          v-model:value="formData.description" 
          :rows="3" 
          placeholder="Deskripsi transaksi (opsional)"
          show-count
          :maxlength="500"
        />
      </a-form-item>

      <a-form-item label="Referensi" name="reference">
        <a-input 
          v-model:value="formData.reference" 
          placeholder="Nomor referensi (GRN, Invoice, dll) - opsional"
        />
      </a-form-item>

      <!-- Preview -->
      <a-card v-if="formData.amount && formData.type" size="small" title="Ringkasan Transaksi">
        <a-descriptions :column="1" size="small">
          <a-descriptions-item label="Tipe">
            <a-tag :color="formData.type === 'income' ? 'green' : 'red'">
              {{ formData.type === 'income' ? 'Pemasukan' : 'Pengeluaran' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Kategori">
            <a-tag :color="getCategoryColor(formData.category)">
              {{ getCategoryLabel(formData.category) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Jumlah">
            <span :style="{ 
              color: formData.type === 'income' ? '#52c41a' : '#ff4d4f', 
              fontSize: '16px', 
              fontWeight: 'bold' 
            }">
              {{ formData.type === 'income' ? '+' : '-' }}{{ formatCurrency(formData.amount) }}
            </span>
          </a-descriptions-item>
        </a-descriptions>
      </a-card>
    </a-form>
  </a-modal>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import cashFlowService from '@/services/cashFlowService'

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  cashFlow: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'success'])

const submitting = ref(false)
const formRef = ref()

const formData = reactive({
  date: dayjs(),
  category: '',
  type: '',
  amount: 0,
  description: '',
  reference: ''
})

const rules = {
  date: [{ required: true, message: 'Tanggal wajib diisi' }],
  category: [{ required: true, message: 'Kategori wajib dipilih' }],
  type: [{ required: true, message: 'Tipe transaksi wajib dipilih' }],
  amount: [
    { required: true, message: 'Jumlah wajib diisi' },
    { type: 'number', min: 1, message: 'Jumlah harus lebih dari 0' }
  ]
}

// Watch for cashFlow prop changes to populate form
watch(() => props.cashFlow, (newCashFlow) => {
  if (newCashFlow) {
    Object.assign(formData, {
      date: newCashFlow.date ? dayjs(newCashFlow.date) : dayjs(),
      category: newCashFlow.category,
      type: newCashFlow.type,
      amount: newCashFlow.amount,
      description: newCashFlow.description || '',
      reference: newCashFlow.reference || ''
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
      date: formData.date ? formData.date.format('YYYY-MM-DD') : null
    }

    if (props.cashFlow) {
      // Edit functionality disabled - backend endpoint not implemented yet
      message.info('Fitur edit belum tersedia')
      return
    } else {
      await cashFlowService.createCashFlow(submitData)
      message.success('Transaksi berhasil ditambahkan')
    }

    emit('success')
  } catch (error) {
    if (error.errorFields) {
      return
    }
    
    const errorMessage = error.response?.data?.message || 'Gagal menyimpan transaksi'
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
    date: dayjs(),
    category: '',
    type: '',
    amount: 0,
    description: '',
    reference: ''
  })
  formRef.value?.resetFields()
}

const getCategoryColor = (category) => {
  const colors = {
    bahan_baku: 'blue',
    gaji: 'green',
    utilitas: 'orange',
    operasional: 'purple'
  }
  return colors[category] || 'default'
}

const getCategoryLabel = (category) => {
  const labels = {
    bahan_baku: 'Bahan Baku',
    gaji: 'Gaji',
    utilitas: 'Utilitas',
    operasional: 'Operasional'
  }
  return labels[category] || category
}

const formatCurrency = (value) => {
  if (!value) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(value)
}
</script>