<template>
  <div class="stok-opname-detail">
    <a-spin :spinning="loading">
      <!-- Form Header -->
      <a-card style="margin-bottom: 16px">
        <a-row :gutter="16">
          <a-col :span="6">
            <a-statistic title="Nomor Form" :value="formData.form_number || '-'" />
          </a-col>
          <a-col :span="6">
            <a-statistic title="Tanggal" :value="formatDate(formData.created_at)" />
          </a-col>
          <a-col :span="6">
            <a-statistic title="Pembuat" :value="formData.creator?.full_name || '-'" />
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
      <a-card v-if="formData.notes" title="Catatan Form" style="margin-bottom: 16px">
        <p>{{ formData.notes }}</p>
      </a-card>

      <!-- Approval Information -->
      <a-card
        v-if="formData.status !== 'pending'"
        title="Informasi Persetujuan"
        style="margin-bottom: 16px"
      >
        <a-descriptions :column="2">
          <a-descriptions-item label="Penyetuju">
            {{ formData.approver?.full_name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Tanggal Persetujuan">
            {{ formatDateTime(formData.approved_at) }}
          </a-descriptions-item>
        </a-descriptions>
        
        <a-alert
          v-if="formData.status === 'rejected' && formData.rejection_reason"
          type="error"
          :message="`Alasan Penolakan: ${formData.rejection_reason}`"
          style="margin-top: 16px"
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
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'ingredient'">
              {{ record.ingredient?.name }}
            </template>
            <template v-else-if="column.key === 'system_stock'">
              {{ record.system_stock }} {{ record.ingredient?.unit }}
            </template>
            <template v-else-if="column.key === 'physical_count'">
              {{ record.physical_count }} {{ record.ingredient?.unit }}
            </template>
            <template v-else-if="column.key === 'difference'">
              <span :class="getDifferenceClass(record.difference)">
                {{ formatDifference(record.difference) }} {{ record.ingredient?.unit }}
              </span>
            </template>
            <template v-else-if="column.key === 'item_notes'">
              {{ record.item_notes || '-' }}
            </template>
          </template>
        </a-table>
      </a-card>

      <!-- Action Buttons -->
      <a-card>
        <a-space>
          <a-button @click="goBack">
            Kembali
          </a-button>
          
          <!-- Creator actions for pending forms -->
          <template v-if="formData.status === 'pending' && isCreator">
            <a-button type="default" @click="editForm">
              Edit
            </a-button>
            <a-button type="primary" danger @click="confirmDelete">
              Hapus
            </a-button>
          </template>
          
          <!-- Kepala SPPG actions for pending forms -->
          <template v-if="formData.status === 'pending' && isKepalaSPPG">
            <a-button type="primary" @click="showApproveModal">
              Setujui
            </a-button>
            <a-button danger @click="showRejectModal">
              Tolak
            </a-button>
          </template>
          
          <!-- Export button for all statuses -->
          <a-button @click="exportForm">
            Export
          </a-button>
        </a-space>
      </a-card>
    </a-spin>

    <!-- Approve Modal -->
    <a-modal
      v-model:open="approveModalVisible"
      title="Konfirmasi Persetujuan"
      @ok="handleApprove"
      :confirm-loading="approving"
      ok-text="Ya, Setujui"
      cancel-text="Batal"
    >
      <p>Apakah Anda yakin ingin menyetujui stok opname ini?</p>
      <p>Penyesuaian stok akan diterapkan secara otomatis setelah persetujuan.</p>
    </a-modal>

    <!-- Reject Modal -->
    <a-modal
      v-model:open="rejectModalVisible"
      title="Tolak Stok Opname"
      @ok="handleReject"
      :confirm-loading="rejecting"
      ok-text="Ya, Tolak"
      ok-type="danger"
      cancel-text="Batal"
    >
      <a-form layout="vertical">
        <a-form-item label="Alasan Penolakan" required>
          <a-textarea
            v-model:value="rejectionReason"
            placeholder="Masukkan alasan penolakan..."
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { useRouter, useRoute } from 'vue-router'
import stokOpnameService from '@/services/stokOpnameService'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

// State
const loading = ref(false)
const approving = ref(false)
const rejecting = ref(false)
const approveModalVisible = ref(false)
const rejectModalVisible = ref(false)
const rejectionReason = ref('')

const formData = reactive({
  id: null,
  form_number: '',
  created_at: null,
  creator: null,
  status: 'pending',
  notes: '',
  approver: null,
  approved_at: null,
  rejection_reason: '',
  items: []
})

// Computed
const currentUser = computed(() => authStore.user)

const isCreator = computed(() => {
  return currentUser.value?.id === formData.creator?.id
})

const isKepalaSPPG = computed(() => {
  return currentUser.value?.role === 'kepala_sppg'
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
    width: 150
  },
  {
    title: 'Jumlah Fisik',
    key: 'physical_count',
    width: 150
  },
  {
    title: 'Selisih',
    key: 'difference',
    width: 150
  },
  {
    title: 'Catatan',
    key: 'item_notes'
  }
]

// Methods
const fetchForm = async () => {
  loading.value = true
  try {
    console.log('Fetching form ID:', route.params.id)
    const response = await stokOpnameService.getForm(route.params.id)
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

const goBack = () => {
  router.push('/inventory')
}

const editForm = () => {
  router.push(`/inventory/stok-opname/${formData.id}/edit`)
}

const confirmDelete = () => {
  Modal.confirm({
    title: 'Konfirmasi Hapus',
    content: `Apakah Anda yakin ingin menghapus form ${formData.form_number}?`,
    okText: 'Ya, Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: async () => {
      await deleteForm()
    }
  })
}

const deleteForm = async () => {
  try {
    await stokOpnameService.deleteForm(formData.id)
    message.success('Form berhasil dihapus')
    router.push('/inventory')
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal menghapus form')
    console.error(error)
  }
}

const showApproveModal = () => {
  approveModalVisible.value = true
}

const handleApprove = async () => {
  approving.value = true
  try {
    await stokOpnameService.approveForm(formData.id)
    message.success('Form berhasil disetujui')
    approveModalVisible.value = false
    await fetchForm()
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal menyetujui form')
    console.error(error)
  } finally {
    approving.value = false
  }
}

const showRejectModal = () => {
  rejectionReason.value = ''
  rejectModalVisible.value = true
}

const handleReject = async () => {
  if (!rejectionReason.value.trim()) {
    message.error('Alasan penolakan harus diisi')
    return
  }
  
  rejecting.value = true
  try {
    await stokOpnameService.rejectForm(formData.id, rejectionReason.value)
    message.success('Form berhasil ditolak')
    rejectModalVisible.value = false
    await fetchForm()
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal menolak form')
    console.error(error)
  } finally {
    rejecting.value = false
  }
}

const exportForm = async () => {
  try {
    const response = await stokOpnameService.exportForm(formData.id, 'excel')
    const url = window.URL.createObjectURL(new Blob([response.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', `stok-opname-${formData.form_number}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    message.success('Form berhasil diekspor')
  } catch (error) {
    message.error('Gagal mengekspor form')
    console.error(error)
  }
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

const formatDateTime = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Lifecycle
onMounted(() => {
  fetchForm()
})
</script>

<style scoped>
.stok-opname-detail {
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
