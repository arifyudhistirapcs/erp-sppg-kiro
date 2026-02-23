<template>
  <div class="goods-receipt">
    <a-page-header
      title="Penerimaan Barang (GRN)"
      sub-title="Catat penerimaan barang dari supplier"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Buat GRN Baru
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search -->
        <a-input-search
          v-model:value="searchText"
          placeholder="Cari nomor GRN atau PO..."
          @search="handleSearch"
          allow-clear
          style="width: 400px"
        />

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="goodsReceipts"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'receipt_date'">
              {{ formatDate(record.receipt_date) }}
            </template>
            <template v-else-if="column.key === 'invoice_photo'">
              <a-button
                v-if="record.invoice_photo"
                type="link"
                size="small"
                @click="viewInvoice(record.invoice_photo)"
              >
                Lihat Foto
              </a-button>
              <span v-else class="text-muted">-</span>
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-button type="link" size="small" @click="viewGRN(record)">
                Detail
              </a-button>
            </template>
          </template>
        </a-table>
      </a-space>
    </a-card>

    <!-- Create Modal -->
    <a-modal
      v-model:open="modalVisible"
      title="Buat Penerimaan Barang Baru"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="1000px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Purchase Order" name="po_id">
              <a-select
                v-model:value="formData.po_id"
                placeholder="Pilih PO yang akan diterima"
                show-search
                :filter-option="filterPO"
                @change="handlePOChange"
              >
                <a-select-option
                  v-for="po in approvedPOs"
                  :key="po.id"
                  :value="po.id"
                >
                  {{ po.po_number }} - {{ po.supplier?.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Tanggal Penerimaan" name="receipt_date">
              <a-date-picker
                v-model:value="formData.receipt_date"
                style="width: 100%"
                format="DD/MM/YYYY"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="Foto Invoice/Nota" name="invoice_photo">
          <a-upload
            v-model:file-list="fileList"
            list-type="picture-card"
            :before-upload="beforeUpload"
            :max-count="1"
            accept="image/*"
            @preview="handlePreview"
          >
            <div v-if="fileList.length < 1">
              <PlusOutlined />
              <div style="margin-top: 8px">Upload</div>
            </div>
          </a-upload>
          <div class="upload-hint">Upload foto invoice/nota dari supplier (wajib)</div>
        </a-form-item>

        <a-divider>Item yang Diterima</a-divider>

        <a-alert
          v-if="hasDiscrepancy"
          message="Perhatian: Ada perbedaan antara jumlah yang dipesan dan diterima"
          type="warning"
          show-icon
          style="margin-bottom: 16px"
        />

        <a-table
          :columns="itemColumns"
          :data-source="formData.items"
          :pagination="false"
          size="small"
        >
          <template #bodyCell="{ column, record, index }">
            <template v-if="column.key === 'ordered_quantity'">
              {{ record.ordered_quantity }}
            </template>
            <template v-else-if="column.key === 'received_quantity'">
              <a-input-number
                v-model:value="record.received_quantity"
                :min="0"
                :step="0.1"
                style="width: 100%"
                @change="checkDiscrepancy(index)"
              />
            </template>
            <template v-else-if="column.key === 'expiry_date'">
              <a-date-picker
                v-model:value="record.expiry_date"
                style="width: 100%"
                format="DD/MM/YYYY"
                placeholder="Opsional"
              />
            </template>
            <template v-else-if="column.key === 'discrepancy'">
              <a-tag v-if="record.discrepancy !== 0" :color="record.discrepancy > 0 ? 'green' : 'red'">
                {{ record.discrepancy > 0 ? '+' : '' }}{{ record.discrepancy }}
              </a-tag>
              <span v-else>-</span>
            </template>
          </template>
        </a-table>

        <a-form-item label="Catatan" name="notes" style="margin-top: 16px">
          <a-textarea
            v-model:value="formData.notes"
            :rows="3"
            placeholder="Catatan tambahan (opsional)"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Penerimaan Barang"
      :footer="null"
      width="900px"
    >
      <a-descriptions v-if="selectedGRN" bordered :column="2">
        <a-descriptions-item label="Nomor GRN" :span="2">
          <strong>{{ selectedGRN.grn_number }}</strong>
        </a-descriptions-item>
        <a-descriptions-item label="Nomor PO">
          {{ selectedGRN.purchase_order?.po_number }}
        </a-descriptions-item>
        <a-descriptions-item label="Supplier">
          {{ selectedGRN.purchase_order?.supplier?.name }}
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal Penerimaan">
          {{ formatDate(selectedGRN.receipt_date) }}
        </a-descriptions-item>
        <a-descriptions-item label="Diterima Oleh">
          {{ selectedGRN.received_by_name }}
        </a-descriptions-item>
        <a-descriptions-item label="Foto Invoice" :span="2">
          <a-button
            v-if="selectedGRN.invoice_photo"
            type="link"
            @click="viewInvoice(selectedGRN.invoice_photo)"
          >
            Lihat Foto Invoice
          </a-button>
          <span v-else>-</span>
        </a-descriptions-item>
        <a-descriptions-item v-if="selectedGRN.notes" label="Catatan" :span="2">
          {{ selectedGRN.notes }}
        </a-descriptions-item>
      </a-descriptions>

      <a-divider>Item yang Diterima</a-divider>

      <a-table
        :columns="detailItemColumns"
        :data-source="selectedGRN?.grn_items || []"
        :pagination="false"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'expiry_date'">
            {{ record.expiry_date ? formatDate(record.expiry_date) : '-' }}
          </template>
          <template v-else-if="column.key === 'discrepancy'">
            <a-tag
              v-if="record.received_quantity !== record.ordered_quantity"
              :color="record.received_quantity > record.ordered_quantity ? 'green' : 'red'"
            >
              {{ record.received_quantity > record.ordered_quantity ? '+' : '' }}
              {{ record.received_quantity - record.ordered_quantity }}
            </a-tag>
            <span v-else>-</span>
          </template>
        </template>
      </a-table>
    </a-modal>

    <!-- Image Preview Modal -->
    <a-modal
      v-model:open="previewVisible"
      title="Foto Invoice"
      :footer="null"
    >
      <img :src="previewImage" style="width: 100%" />
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import goodsReceiptService from '@/services/goodsReceiptService'
import purchaseOrderService from '@/services/purchaseOrderService'
import dayjs from 'dayjs'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const previewVisible = ref(false)
const previewImage = ref('')
const selectedGRN = ref(null)
const goodsReceipts = ref([])
const approvedPOs = ref([])
const searchText = ref('')
const formRef = ref()
const fileList = ref([])

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  po_id: undefined,
  receipt_date: dayjs(),
  invoice_photo: null,
  items: [],
  notes: ''
})

const rules = {
  po_id: [{ required: true, message: 'Purchase Order wajib dipilih' }],
  receipt_date: [{ required: true, message: 'Tanggal penerimaan wajib diisi' }],
  invoice_photo: [{ required: true, message: 'Foto invoice wajib diupload' }]
}

const hasDiscrepancy = computed(() => {
  return formData.items.some(item => item.discrepancy !== 0)
})

const columns = [
  {
    title: 'Nomor GRN',
    dataIndex: 'grn_number',
    key: 'grn_number'
  },
  {
    title: 'Nomor PO',
    dataIndex: ['purchase_order', 'po_number'],
    key: 'po_number'
  },
  {
    title: 'Supplier',
    dataIndex: ['purchase_order', 'supplier', 'name'],
    key: 'supplier_name'
  },
  {
    title: 'Tanggal Penerimaan',
    key: 'receipt_date'
  },
  {
    title: 'Diterima Oleh',
    dataIndex: 'received_by_name',
    key: 'received_by_name'
  },
  {
    title: 'Foto Invoice',
    key: 'invoice_photo'
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 100
  }
]

const itemColumns = [
  {
    title: 'Bahan',
    dataIndex: 'ingredient_name',
    key: 'ingredient_name'
  },
  {
    title: 'Satuan',
    dataIndex: 'unit',
    key: 'unit'
  },
  {
    title: 'Jumlah Dipesan',
    key: 'ordered_quantity',
    width: 120
  },
  {
    title: 'Jumlah Diterima',
    key: 'received_quantity',
    width: 150
  },
  {
    title: 'Tanggal Kadaluarsa',
    key: 'expiry_date',
    width: 180
  },
  {
    title: 'Selisih',
    key: 'discrepancy',
    width: 100
  }
]

const detailItemColumns = [
  {
    title: 'Bahan',
    dataIndex: ['ingredient', 'name'],
    key: 'ingredient_name'
  },
  {
    title: 'Satuan',
    dataIndex: ['ingredient', 'unit'],
    key: 'unit'
  },
  {
    title: 'Jumlah Dipesan',
    dataIndex: 'ordered_quantity',
    key: 'ordered_quantity'
  },
  {
    title: 'Jumlah Diterima',
    dataIndex: 'received_quantity',
    key: 'received_quantity'
  },
  {
    title: 'Tanggal Kadaluarsa',
    key: 'expiry_date'
  },
  {
    title: 'Selisih',
    key: 'discrepancy'
  }
]

const fetchGoodsReceipts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: searchText.value || undefined
    }
    const response = await goodsReceiptService.getGoodsReceipts(params)
    goodsReceipts.value = response.data.data || []
    pagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat data penerimaan barang')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchApprovedPOs = async () => {
  try {
    const response = await purchaseOrderService.getPurchaseOrders({ status: 'approved' })
    approvedPOs.value = response.data.data || []
  } catch (error) {
    console.error('Gagal memuat data PO:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchGoodsReceipts()
}

const handleSearch = () => {
  pagination.current = 1
  fetchGoodsReceipts()
}

const showCreateModal = () => {
  resetForm()
  modalVisible.value = true
}

const handlePOChange = async (poId) => {
  try {
    const response = await purchaseOrderService.getPurchaseOrder(poId)
    const po = response.data
    
    formData.items = po.po_items.map(item => ({
      ingredient_id: item.ingredient_id,
      ingredient_name: item.ingredient?.name,
      unit: item.ingredient?.unit,
      ordered_quantity: item.quantity,
      received_quantity: item.quantity,
      expiry_date: null,
      discrepancy: 0
    }))
  } catch (error) {
    message.error('Gagal memuat detail PO')
    console.error(error)
  }
}

const checkDiscrepancy = (index) => {
  const item = formData.items[index]
  item.discrepancy = item.received_quantity - item.ordered_quantity
}

const beforeUpload = (file) => {
  const isImage = file.type.startsWith('image/')
  if (!isImage) {
    message.error('Hanya file gambar yang diperbolehkan!')
    return false
  }
  
  const isLt5M = file.size / 1024 / 1024 < 5
  if (!isLt5M) {
    message.error('Ukuran gambar harus kurang dari 5MB!')
    return false
  }
  
  formData.invoice_photo = file
  return false
}

const handlePreview = async (file) => {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
}

const getBase64 = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}

const viewInvoice = (url) => {
  previewImage.value = url
  previewVisible.value = true
}

const viewGRN = async (grn) => {
  try {
    const response = await goodsReceiptService.getGoodsReceipt(grn.id)
    selectedGRN.value = response.data
    detailModalVisible.value = true
  } catch (error) {
    message.error('Gagal memuat detail GRN')
    console.error(error)
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    if (!formData.invoice_photo) {
      message.error('Foto invoice wajib diupload')
      return
    }

    submitting.value = true

    // Create GRN
    const payload = {
      po_id: formData.po_id,
      receipt_date: formData.receipt_date.format('YYYY-MM-DD'),
      items: formData.items.map(item => ({
        ingredient_id: item.ingredient_id,
        ordered_quantity: item.ordered_quantity,
        received_quantity: item.received_quantity,
        expiry_date: item.expiry_date ? item.expiry_date.format('YYYY-MM-DD') : null
      })),
      notes: formData.notes
    }

    const response = await goodsReceiptService.createGoodsReceipt(payload)
    const grnId = response.data.id

    // Upload invoice photo
    const formDataUpload = new FormData()
    formDataUpload.append('invoice_photo', formData.invoice_photo)
    await goodsReceiptService.uploadInvoicePhoto(grnId, formDataUpload)

    message.success('Penerimaan barang berhasil dicatat. Stok inventory telah diperbarui.')
    modalVisible.value = false
    fetchGoodsReceipts()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    message.error('Gagal menyimpan penerimaan barang')
    console.error(error)
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  formData.po_id = undefined
  formData.receipt_date = dayjs()
  formData.invoice_photo = null
  formData.items = []
  formData.notes = ''
  fileList.value = []
  formRef.value?.resetFields()
}

const filterPO = (input, option) => {
  return option.children[0].children.toLowerCase().includes(input.toLowerCase())
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

onMounted(() => {
  fetchGoodsReceipts()
  fetchApprovedPOs()
})
</script>

<style scoped>
.goods-receipt {
  padding: 24px;
}

.upload-hint {
  color: #999;
  font-size: 12px;
  margin-top: 8px;
}

.text-muted {
  color: #999;
}
</style>
