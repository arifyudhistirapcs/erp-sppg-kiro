<template>
  <div class="stok-opname-form">
    <a-spin :spinning="loading">
      <!-- Form Header -->
      <a-card style="margin-bottom: 16px">
        <a-row :gutter="16">
          <a-col :span="6">
            <a-statistic title="Nomor Form" :value="formData.form_number || 'Auto-generate'" />
          </a-col>
          <a-col :span="6">
            <a-statistic title="Tanggal" :value="formatDate(formData.created_at)" />
          </a-col>
          <a-col :span="6">
            <a-statistic title="Pembuat" :value="formData.creator?.full_name || currentUser?.full_name || '-'" />
          </a-col>
          <a-col :span="6">
            <a-statistic title="Status">
              <template #formatter>
                <a-tag :color="getStatusColor(formData.status)">
                  {{ getStatusText(formData.status) }}
                </a-tag>
              </template>
            </a-statistic>
          </a-col>
        </a-row>
      </a-card>

      <!-- Form Notes -->
      <a-card title="Catatan Form" style="margin-bottom: 16px">
        <a-textarea
          v-model:value="formData.notes"
          placeholder="Masukkan catatan untuk form ini..."
          :rows="3"
          :disabled="!isEditable"
        />
      </a-card>

      <!-- Items List -->
      <a-card title="Daftar Item" style="margin-bottom: 16px">
        <a-table
          :columns="itemColumns"
          :data-source="formData.items"
          :pagination="false"
          row-key="id"
        >
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'ingredient'">
              <a-select
                v-if="isEditable && !record.ingredient"
                v-model:value="record.ingredient_id"
                placeholder="Pilih bahan"
                show-search
                option-filter-prop="label"
                style="width: 100%"
                @change="onIngredientSelect(record)"
              >
                <a-select-option
                  v-for="inv in availableIngredients"
                  :key="inv.ingredient_id"
                  :value="inv.ingredient_id"
                  :label="inv.ingredient?.name || 'Unknown'"
                >
                  {{ inv.ingredient?.name || 'Unknown' }}
                </a-select-option>
              </a-select>
              <span v-else>{{ record.ingredient?.name || 'Pilih bahan' }}</span>
            </template>
            <template v-else-if="column.key === 'system_stock'">
              {{ record.system_stock }} {{ record.ingredient?.unit }}
            </template>
            <template v-else-if="column.key === 'physical_count'">
              <a-input-number
                v-model:value="record.physical_count"
                :min="0"
                :precision="2"
                style="width: 100%"
                @change="calculateDifference(record)"
                :disabled="!isEditable"
              />
            </template>
            <template v-else-if="column.key === 'difference'">
              <span :class="getDifferenceClass(record.difference)">
                {{ formatDifference(record.difference) }} {{ record.ingredient?.unit }}
              </span>
            </template>
            <template v-else-if="column.key === 'item_notes'">
              <a-input
                v-model:value="record.item_notes"
                placeholder="Catatan item..."
                :disabled="!isEditable"
              />
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-button
                v-if="isEditable"
                type="link"
                size="small"
                danger
                @click="removeItem(index)"
              >
                Hapus
              </a-button>
            </template>
          </template>
        </a-table>

        <a-button
          v-if="isEditable"
          type="dashed"
          block
          style="margin-top: 16px"
          @click="addItem"
        >
          <template #icon><PlusOutlined /></template>
          Tambah Item
        </a-button>
      </a-card>

      <!-- Action Buttons -->
      <a-card>
        <a-space>
          <a-button @click="cancel">
            Batal
          </a-button>
          <a-button
            v-if="isEditable"
            type="default"
            @click="saveDraft"
            :loading="saving"
          >
            Simpan Draft
          </a-button>
          <a-button
            v-if="isEditable && formData.id"
            type="primary"
            @click="submitForApproval"
            :loading="submitting"
          >
            Ajukan Persetujuan
          </a-button>
        </a-space>
      </a-card>
    </a-spin>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { useRouter, useRoute } from 'vue-router'
import { PlusOutlined } from '@ant-design/icons-vue'
import stokOpnameService from '@/services/stokOpnameService'
import inventoryService from '@/services/inventoryService'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// State
const loading = ref(false)
const saving = ref(false)
const submitting = ref(false)
const ingredients = ref([])
const inventory = ref([])
const currentUser = computed(() => authStore.user)

const formData = reactive({
  id: null,
  form_number: '',
  created_at: new Date().toISOString(),
  creator: null,
  status: 'pending',
  notes: '',
  items: []
})

// Computed
const isEditable = computed(() => {
  return !formData.status || formData.status === 'pending'
})

const availableIngredients = computed(() => {
  console.log('Computing available ingredients. Total inventory:', inventory.value.length)
  const selectedIds = formData.items
    .filter(item => item.ingredient_id)
    .map(item => item.ingredient_id)
  console.log('Selected ingredient IDs:', selectedIds)
  
  // Filter out already selected ingredients
  const available = inventory.value.filter(inv => !selectedIds.includes(inv.ingredient_id))
  console.log('Available ingredients:', available.length)
  
  return available
})

// Table columns
const itemColumns = [
  {
    title: 'Bahan',
    key: 'ingredient',
    width: 200
  },
  {
    title: 'Stok Sistem',
    key: 'system_stock',
    width: 120
  },
  {
    title: 'Jumlah Fisik',
    key: 'physical_count',
    width: 150
  },
  {
    title: 'Selisih',
    key: 'difference',
    width: 120
  },
  {
    title: 'Catatan',
    key: 'item_notes',
    width: 200
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 100
  }
]

// Methods
const fetchInventory = async () => {
  try {
    const response = await inventoryService.getInventory()
    console.log('Inventory response:', response.data)
    inventory.value = response.data.inventory_items || []
    console.log('Inventory items loaded:', inventory.value.length)
    
    if (inventory.value.length === 0) {
      message.warning('Tidak ada data inventory. Silakan inisialisasi inventory terlebih dahulu.')
    }
  } catch (error) {
    message.error('Gagal memuat data inventory: ' + (error.response?.data?.message || error.message))
    console.error('Inventory fetch error:', error)
  }
}

const fetchForm = async (id) => {
  loading.value = true
  try {
    console.log('Fetching form for edit, ID:', id)
    const response = await stokOpnameService.getForm(id)
    console.log('Form response:', response.data)
    
    // Backend returns { success, data: form }
    const form = response.data.data || response.data.form || response.data
    console.log('Form data:', form)
    
    Object.assign(formData, form)
    console.log('Form data assigned:', formData)
  } catch (error) {
    console.error('Fetch form error:', error)
    message.error('Gagal memuat data form: ' + (error.response?.data?.message || error.message))
    router.push('/inventory')
  } finally {
    loading.value = false
  }
}

const addItem = () => {
  formData.items.push({
    ingredient_id: null,
    ingredient: null,
    system_stock: 0,
    physical_count: 0,
    difference: 0,
    item_notes: ''
  })
}

const removeItem = (index) => {
  formData.items.splice(index, 1)
}

const onIngredientSelect = (item) => {
  console.log('Ingredient selected, ID:', item.ingredient_id)
  const inv = inventory.value.find(i => i.ingredient_id === item.ingredient_id)
  console.log('Found inventory item:', inv)
  if (inv) {
    // Use Object.assign to ensure Vue reactivity
    Object.assign(item, {
      ingredient: inv.ingredient,
      system_stock: inv.quantity,
      difference: 0 - inv.quantity // Will be recalculated
    })
    calculateDifference(item)
    console.log('Item after selection:', item)
  }
}

const calculateDifference = (item) => {
  if (item.physical_count !== null && item.physical_count !== undefined) {
    item.difference = item.physical_count - item.system_stock
  }
}

const saveDraft = async () => {
  if (!validateForm()) return
  
  console.log('Saving draft with data:', formData)
  
  saving.value = true
  try {
    if (formData.id) {
      // Update existing form
      console.log('Updating form notes:', formData.notes)
      await stokOpnameService.updateFormNotes(formData.id, formData.notes)
      
      // Update items
      for (const item of formData.items) {
        console.log('Processing item:', item)
        if (item.id) {
          await stokOpnameService.updateItem(item.id, {
            physical_count: item.physical_count,
            notes: item.item_notes || ''
          })
        } else if (item.ingredient_id) {
          const itemData = {
            ingredient_id: item.ingredient_id,
            physical_count: item.physical_count,
            notes: item.item_notes || ''
          }
          console.log('Adding item:', itemData)
          const response = await stokOpnameService.addItem(formData.id, itemData)
          // Update item with ID from server
          if (response.data.item) {
            item.id = response.data.item.id
          }
        }
      }
      message.success('Form berhasil disimpan')
    } else {
      // Create new form
      console.log('Creating new form with notes:', formData.notes)
      const response = await stokOpnameService.createForm({ notes: formData.notes })
      console.log('Form created:', response.data)
      const newForm = response.data.data || response.data.form || response.data
      formData.id = newForm.id
      formData.form_number = newForm.form_number
      
      // Add items
      for (const item of formData.items) {
        if (item.ingredient_id) {
          const itemData = {
            ingredient_id: item.ingredient_id,
            physical_count: item.physical_count,
            notes: item.item_notes || ''
          }
          console.log('Adding item to new form:', itemData)
          const response = await stokOpnameService.addItem(formData.id, itemData)
          // Update item with ID from server
          if (response.data.item) {
            item.id = response.data.item.id
          }
        }
      }
      message.success('Form berhasil dibuat dan disimpan')
      
      // Update URL to edit mode without redirecting
      router.replace(`/inventory/stok-opname/${formData.id}/edit`)
    }
  } catch (error) {
    console.error('Save draft error:', error)
    console.error('Error response:', error.response?.data)
    const errorMessage = error.response?.data?.message || error.response?.data?.error?.message || error.message || 'Gagal menyimpan form'
    message.error(errorMessage)
  } finally {
    saving.value = false
  }
}

const submitForApproval = async () => {
  if (!validateForm(true)) return
  
  submitting.value = true
  try {
    await stokOpnameService.submitForApproval(formData.id)
    message.success('Form berhasil diajukan untuk persetujuan')
    router.push('/inventory')
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal mengajukan form')
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const validateForm = (forSubmission = false) => {
  if (formData.items.length === 0) {
    message.error('Form harus memiliki minimal satu item')
    return false
  }
  
  // Check for items without ingredient selected
  const hasItemWithoutIngredient = formData.items.some(item => !item.ingredient_id)
  if (hasItemWithoutIngredient) {
    message.error('Semua item harus memiliki bahan yang dipilih')
    return false
  }
  
  if (forSubmission) {
    const hasInvalidItem = formData.items.some(item => 
      !item.ingredient_id || 
      item.physical_count === null || 
      item.physical_count === undefined ||
      item.physical_count < 0
    )
    
    if (hasInvalidItem) {
      message.error('Semua item harus memiliki bahan dan jumlah fisik yang valid (>= 0)')
      return false
    }
  }
  
  return true
}

const cancel = () => {
  router.push('/inventory')
}

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    approved: 'green',
    rejected: 'red'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    pending: 'Pending',
    approved: 'Disetujui',
    rejected: 'Ditolak'
  }
  return texts[status] || status
}

const getDifferenceClass = (diff) => {
  if (diff > 0) return 'text-success'
  if (diff < 0) return 'text-danger'
  return ''
}

const formatDifference = (diff) => {
  if (diff > 0) return `+${diff}`
  return diff
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

// Lifecycle
onMounted(async () => {
  await fetchInventory()
  
  const formId = route.params.id
  if (formId) {
    await fetchForm(formId)
  } else {
    // Add one empty item for new form
    addItem()
  }
})
</script>

<style scoped>
.stok-opname-form {
  padding: 24px;
}

.text-success {
  color: #3f8600;
  font-weight: 500;
}

.text-danger {
  color: #cf1322;
  font-weight: 500;
}
</style>
