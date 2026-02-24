<template>
  <div class="purchase-order-list">
    <a-page-header
      title="Purchase Order (PO)"
      sub-title="Kelola pemesanan barang ke supplier"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Buat PO Baru
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Search and Filter -->
        <a-row :gutter="16">
          <a-col :span="8">
            <a-input-search
              v-model:value="searchText"
              placeholder="Cari nomor PO atau supplier..."
              @search="handleSearch"
              allow-clear
            />
          </a-col>
          <a-col :span="6">
            <a-select
              v-model:value="filterStatus"
              placeholder="Status"
              style="width: 100%"
              @change="handleSearch"
              allow-clear
            >
              <a-select-option value="pending">Pending</a-select-option>
              <a-select-option value="approved">Disetujui</a-select-option>
              <a-select-option value="received">Diterima</a-select-option>
              <a-select-option value="cancelled">Dibatalkan</a-select-option>
            </a-select>
          </a-col>
        </a-row>

        <!-- Table -->
        <a-table
          :columns="columns"
          :data-source="purchaseOrders"
          :loading="loading"
          :pagination="pagination"
          @change="handleTableChange"
          row-key="id"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'status'">
              <a-tag :color="getStatusColor(record.status)">
                {{ getStatusText(record.status) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'total_amount'">
              {{ formatCurrency(record.total_amount) }}
            </template>
            <template v-else-if="column.key === 'order_date'">
              {{ formatDate(record.order_date) }}
            </template>
            <template v-else-if="column.key === 'expected_delivery'">
              {{ formatDate(record.expected_delivery) }}
            </template>
            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-button type="link" size="small" @click="viewPO(record)">
                  Detail
                </a-button>
                <a-button
                  v-if="record.status === 'pending' && canApprove"
                  type="link"
                  size="small"
                  @click="approvePO(record)"
                >
                  Setujui
                </a-button>
                <a-button
                  v-if="record.status === 'pending'"
                  type="link"
                  size="small"
                  @click="editPO(record)"
                >
                  Edit
                </a-button>
              </a-space>
            </template>
          </template>
        </a-table>
      </a-space>
    </a-card>

    <!-- Create/Edit Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="editingPO ? 'Edit Purchase Order' : 'Buat Purchase Order Baru'"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="900px"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        layout="vertical"
      >
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Supplier" name="supplier_id">
              <a-select
                v-model:value="formData.value.supplier_id"
                placeholder="Pilih supplier"
                show-search
                :filter-option="filterSupplier"
                @change="handleSupplierChange"
              >
                <a-select-option
                  v-for="supplier in activeSuppliers"
                  :key="supplier.id"
                  :value="supplier.id"
                >
                  {{ supplier.name }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Tanggal Pengiriman Diharapkan" name="expected_delivery">
              <a-date-picker
                v-model:value="formData.value.expected_delivery"
                style="width: 100%"
                format="DD/MM/YYYY"
                :disabled-date="disabledDate"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider>Item Pesanan</a-divider>

        <a-form-item label="Item" name="items">
          <a-table
            :columns="itemColumns"
            :data-source="formData.value.items"
            :pagination="false"
            size="small"
          >
            <template #bodyCell="{ column, record, index }">
              <template v-if="column.key === 'ingredient_id'">
                <a-select
                  v-model:value="record.ingredient_id"
                  placeholder="Pilih bahan"
                  show-search
                  style="width: 100%"
                  :filter-option="filterIngredient"
                  @change="calculateSubtotal(index)"
                >
                  <a-select-option
                    v-for="ingredient in ingredients"
                    :key="ingredient.id"
                    :value="ingredient.id"
                  >
                    {{ ingredient.name }} ({{ ingredient.unit }})
                  </a-select-option>
                </a-select>
              </template>
              <template v-else-if="column.key === 'quantity'">
                <a-input-number
                  v-model:value="record.quantity"
                  :min="0.01"
                  :step="0.1"
                  style="width: 100%"
                  @change="calculateSubtotal(index)"
                />
              </template>
              <template v-else-if="column.key === 'unit_price'">
                <a-input-number
                  v-model:value="record.unit_price"
                  :min="0"
                  :step="1000"
                  style="width: 100%"
                  :formatter="value => `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')"
                  :parser="value => value.replace(/Rp\s?|(,*)/g, '')"
                  @change="calculateSubtotal(index)"
                />
              </template>
              <template v-else-if="column.key === 'subtotal'">
                {{ formatCurrency(record.subtotal || 0) }}
              </template>
              <template v-else-if="column.key === 'actions'">
                <a-button type="link" size="small" danger @click="removeItem(index)">
                  Hapus
                </a-button>
              </template>
            </template>
          </a-table>

          <a-button type="dashed" block @click="addItem" style="margin-top: 8px">
            <template #icon><PlusOutlined /></template>
            Tambah Item
          </a-button>
        </a-form-item>

        <a-row justify="end">
          <a-col>
            <a-statistic
              title="Total"
              :value="totalAmount"
            >
              <template #formatter>
                {{ formatCurrency(totalAmount) }}
              </template>
            </a-statistic>
          </a-col>
        </a-row>
      </a-form>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Purchase Order"
      :footer="null"
      width="900px"
    >
      <a-descriptions v-if="selectedPO" bordered :column="2">
        <a-descriptions-item label="Nomor PO" :span="2">
          <strong>{{ selectedPO.po_number }}</strong>
        </a-descriptions-item>
        <a-descriptions-item label="Supplier">
          {{ selectedPO.supplier?.name }}
        </a-descriptions-item>
        <a-descriptions-item label="Status">
          <a-tag :color="getStatusColor(selectedPO.status)">
            {{ getStatusText(selectedPO.status) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal Order">
          {{ formatDate(selectedPO.order_date) }}
        </a-descriptions-item>
        <a-descriptions-item label="Tanggal Pengiriman">
          {{ formatDate(selectedPO.expected_delivery) }}
        </a-descriptions-item>
        <a-descriptions-item label="Total" :span="2">
          <strong>{{ formatCurrency(selectedPO.total_amount) }}</strong>
        </a-descriptions-item>
        <a-descriptions-item v-if="selectedPO.approved_by" label="Disetujui Oleh" :span="2">
          {{ selectedPO.approver?.name }} pada {{ formatDate(selectedPO.approved_at) }}
        </a-descriptions-item>
      </a-descriptions>

      <a-divider>Item Pesanan</a-divider>

      <a-table
        :columns="detailItemColumns"
        :data-source="selectedPO?.po_items || []"
        :pagination="false"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'unit_price'">
            {{ formatCurrency(record.unit_price) }}
          </template>
          <template v-else-if="column.key === 'subtotal'">
            {{ formatCurrency(record.subtotal) }}
          </template>
        </template>
      </a-table>

      <template v-if="selectedPO?.status === 'pending' && canApprove">
        <a-divider />
        <a-row justify="end">
          <a-space>
            <a-button @click="detailModalVisible = false">Tutup</a-button>
            <a-button type="primary" @click="approvePO(selectedPO)">
              Setujui PO
            </a-button>
          </a-space>
        </a-row>
      </template>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import purchaseOrderService from '@/services/purchaseOrderService'
import supplierService from '@/services/supplierService'
import recipeService from '@/services/recipeService'
import { useAuthStore } from '@/stores/auth'
import dayjs from 'dayjs'

const authStore = useAuthStore()
const canApprove = computed(() => authStore.user?.role === 'kepala_sppg')

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const editingPO = ref(null)
const selectedPO = ref(null)
const purchaseOrders = ref([])
const activeSuppliers = ref([])
const ingredients = ref([])
const searchText = ref('')
const filterStatus = ref(undefined)
const formRef = ref()

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const formData = ref({
  supplier_id: undefined,
  expected_delivery: null,
  items: []
})

const rules = {
  supplier_id: [{ required: true, message: 'Supplier wajib dipilih' }],
  expected_delivery: [{ required: true, message: 'Tanggal pengiriman wajib diisi' }],
  items: [
    {
      required: true,
      validator: (rule, value) => {
        if (!value || value.length === 0) {
          return Promise.reject('Minimal satu item harus ditambahkan')
        }
        return Promise.resolve()
      }
    }
  ]
}

const columns = [
  {
    title: 'Nomor PO',
    dataIndex: 'po_number',
    key: 'po_number'
  },
  {
    title: 'Supplier',
    dataIndex: ['supplier', 'name'],
    key: 'supplier_name'
  },
  {
    title: 'Tanggal Order',
    key: 'order_date'
  },
  {
    title: 'Tanggal Pengiriman',
    key: 'expected_delivery'
  },
  {
    title: 'Total',
    key: 'total_amount'
  },
  {
    title: 'Status',
    key: 'status',
    width: 120
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 200
  }
]

const itemColumns = [
  {
    title: 'Bahan',
    key: 'ingredient_id',
    width: 250
  },
  {
    title: 'Jumlah',
    key: 'quantity',
    width: 120
  },
  {
    title: 'Harga Satuan',
    key: 'unit_price',
    width: 150
  },
  {
    title: 'Subtotal',
    key: 'subtotal',
    width: 150
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 80
  }
]

const detailItemColumns = [
  {
    title: 'Bahan',
    dataIndex: ['ingredient', 'name'],
    key: 'ingredient_name'
  },
  {
    title: 'Jumlah',
    dataIndex: 'quantity',
    key: 'quantity'
  },
  {
    title: 'Satuan',
    dataIndex: ['ingredient', 'unit'],
    key: 'unit'
  },
  {
    title: 'Harga Satuan',
    key: 'unit_price'
  },
  {
    title: 'Subtotal',
    key: 'subtotal'
  }
]

const totalAmount = computed(() => {
  return formData.value.items.reduce((sum, item) => {
    return sum + (parseFloat(item.subtotal) || 0)
  }, 0)
})

const fetchPurchaseOrders = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize,
      search: searchText.value || undefined,
      status: filterStatus.value
    }
    const response = await purchaseOrderService.getPurchaseOrders(params)
    purchaseOrders.value = response.data.purchase_orders || []
    pagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat data purchase order')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchSuppliers = async () => {
  try {
    const response = await supplierService.getSuppliers({ is_active: 'active' })
    activeSuppliers.value = response.data.suppliers || []
  } catch (error) {
    console.error('Gagal memuat data supplier:', error)
  }
}

const fetchIngredients = async () => {
  try {
    const response = await recipeService.getIngredients()
    ingredients.value = response.data.data || []
  } catch (error) {
    console.error('Gagal memuat data bahan:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchPurchaseOrders()
}

const handleSearch = () => {
  pagination.current = 1
  fetchPurchaseOrders()
}

const showCreateModal = () => {
  editingPO.value = null
  resetForm()
  modalVisible.value = true
}

const editPO = (po) => {
  editingPO.value = po
  formData.value = {
    supplier_id: po.supplier_id,
    expected_delivery: dayjs(po.expected_delivery),
    items: po.po_items.map(item => ({
      ingredient_id: item.ingredient_id,
      quantity: item.quantity,
      unit_price: item.unit_price,
      subtotal: item.subtotal
    }))
  }
  modalVisible.value = true
}

const viewPO = async (po) => {
  try {
    const response = await purchaseOrderService.getPurchaseOrder(po.id)
    selectedPO.value = response.data.purchase_order
    detailModalVisible.value = true
  } catch (error) {
    message.error('Gagal memuat detail PO')
    console.error(error)
  }
}

const approvePO = async (po) => {
  try {
    await purchaseOrderService.approvePurchaseOrder(po.id)
    message.success('Purchase Order berhasil disetujui')
    detailModalVisible.value = false
    fetchPurchaseOrders()
  } catch (error) {
    message.error('Gagal menyetujui PO')
    console.error(error)
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    
    if (formData.value.items.length === 0) {
      message.error('Minimal satu item harus ditambahkan')
      return
    }

    submitting.value = true

    const payload = {
      supplier_id: formData.value.supplier_id,
      expected_delivery: formData.value.expected_delivery.format('YYYY-MM-DD'),
      items: formData.value.items.map(item => ({
        ingredient_id: item.ingredient_id,
        quantity: item.quantity,
        unit_price: item.unit_price
      }))
    }

    if (editingPO.value) {
      await purchaseOrderService.updatePurchaseOrder(editingPO.value.id, payload)
      message.success('Purchase Order berhasil diperbarui')
    } else {
      await purchaseOrderService.createPurchaseOrder(payload)
      message.success('Purchase Order berhasil dibuat')
    }

    modalVisible.value = false
    fetchPurchaseOrders()
  } catch (error) {
    if (error.errorFields) {
      return
    }
    message.error('Gagal menyimpan purchase order')
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
  formData.value = {
    supplier_id: undefined,
    expected_delivery: null,
    items: []
  }
  formRef.value?.resetFields()
}

const addItem = () => {
  formData.value.items.push({
    ingredient_id: undefined,
    quantity: 1,
    unit_price: 0,
    subtotal: 0
  })
}

const removeItem = (index) => {
  formData.value.items.splice(index, 1)
}

const calculateSubtotal = (index) => {
  const items = [...formData.value.items]
  const item = items[index]
  const quantity = parseFloat(item.quantity) || 0
  const unitPrice = parseCurrency(item.unit_price)
  items[index] = {
    ...item,
    subtotal: quantity * unitPrice
  }
  formData.value.items = items
}

const handleSupplierChange = () => {
  // Could fetch supplier-specific pricing here
}

const filterSupplier = (input, option) => {
  return option.children[0].children.toLowerCase().includes(input.toLowerCase())
}

const filterIngredient = (input, option) => {
  return option.children[0].children.toLowerCase().includes(input.toLowerCase())
}

const disabledDate = (current) => {
  return current && current < dayjs().startOf('day')
}

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    approved: 'blue',
    received: 'green',
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    pending: 'Pending',
    approved: 'Disetujui',
    received: 'Diterima',
    cancelled: 'Dibatalkan'
  }
  return texts[status] || status
}

const formatCurrency = (value) => {
  if (value === undefined || value === null || isNaN(value)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(value)
}

const parseCurrency = (value) => {
  if (!value) return 0
  // Remove Rp, spaces, and dots (thousand separators), replace comma with dot
  const cleaned = value.toString().replace(/Rp\s?|[.]/g, '').replace(',', '.')
  return parseFloat(cleaned) || 0
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('id-ID', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

onMounted(() => {
  fetchPurchaseOrders()
  fetchSuppliers()
  fetchIngredients()
})
</script>

<style scoped>
.purchase-order-list {
  padding: 24px;
}
</style>
