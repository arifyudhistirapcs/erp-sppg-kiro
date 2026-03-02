<template>
  <a-card title="Buat Tugas Pengambilan Ompreng">
    <a-space direction="vertical" style="width: 100%" :size="16">
      <!-- Eligible Orders Section -->
      <div>
        <h4>Pilih Order yang Siap Diambil (Stage 9: Sudah Diterima)</h4>
        <a-table
          :columns="eligibleOrderColumns"
          :data-source="eligibleOrders"
          :row-selection="rowSelection"
          :loading="loadingOrders"
          :pagination="{ pageSize: 10 }"
          row-key="delivery_record_id"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'school_info'">
              <div>
                <div><strong>{{ record.school_name }}</strong></div>
                <div class="text-gray">{{ record.school_address }}</div>
              </div>
            </template>
            <template v-else-if="column.key === 'gps'">
              <div class="text-small">
                {{ record.latitude?.toFixed(6) }}, {{ record.longitude?.toFixed(6) }}
                <a-button 
                  type="link" 
                  size="small" 
                  @click="openMaps(record.latitude, record.longitude)"
                >
                  <template #icon><EnvironmentOutlined /></template>
                </a-button>
              </div>
            </template>
            <template v-else-if="column.key === 'ompreng_count'">
              <a-tag color="blue">{{ record.ompreng_count }} wadah</a-tag>
            </template>
          </template>
        </a-table>
      </div>

      <!-- Selected Orders with Route Order (Drag and Drop) -->
      <div v-if="selectedOrders.length > 0">
        <a-divider>Urutan Rute Pengambilan ({{ selectedOrders.length }} sekolah)</a-divider>
        <a-alert
          message="Seret untuk mengatur urutan pengambilan"
          type="info"
          show-icon
          style="margin-bottom: 12px"
        />
        <draggable
          v-model="selectedOrders"
          item-key="delivery_record_id"
          handle=".drag-handle"
          @end="updateRouteOrder"
        >
          <template #item="{ element, index }">
            <a-card 
              size="small" 
              style="margin-bottom: 8px; cursor: move"
              :body-style="{ padding: '12px' }"
            >
              <div style="display: flex; align-items: center; gap: 12px">
                <div class="drag-handle" style="cursor: grab; font-size: 18px; color: #999">
                  <HolderOutlined />
                </div>
                <a-tag color="blue" style="margin: 0">Rute {{ index + 1 }}</a-tag>
                <div style="flex: 1">
                  <div><strong>{{ element.school_name }}</strong></div>
                  <div class="text-gray text-small">
                    {{ element.school_address }}
                  </div>
                  <div class="text-gray text-small">
                    GPS: {{ element.latitude?.toFixed(6) }}, {{ element.longitude?.toFixed(6) }}
                  </div>
                </div>
                <div>
                  <a-tag color="blue">{{ element.ompreng_count }} wadah</a-tag>
                </div>
                <a-button 
                  type="text" 
                  danger 
                  size="small"
                  @click="removeOrder(element.delivery_record_id)"
                >
                  <template #icon><DeleteOutlined /></template>
                </a-button>
              </div>
            </a-card>
          </template>
        </draggable>
      </div>

      <!-- Driver Selection -->
      <a-form-item label="Pilih Driver" :required="true">
        <a-select
          v-model:value="selectedDriver"
          placeholder="Pilih driver yang tersedia"
          show-search
          :filter-option="filterDriverOption"
          :loading="loadingDrivers"
          style="width: 100%"
        >
          <a-select-option 
            v-for="driver in availableDrivers" 
            :key="driver.driver_id" 
            :value="driver.driver_id"
          >
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>{{ driver.full_name }}</span>
              <span class="text-gray text-small">{{ driver.phone_number }}</span>
            </div>
          </a-select-option>
        </a-select>
        <div v-if="availableDrivers.length === 0 && !loadingDrivers" style="margin-top: 8px;">
          <a-alert
            message="Tidak ada driver yang tersedia"
            type="warning"
            show-icon
          />
        </div>
      </a-form-item>

      <!-- Submit Button -->
      <a-space>
        <a-button
          type="primary"
          size="large"
          @click="handleSubmit"
          :disabled="!canSubmit"
          :loading="submitting"
        >
          <template #icon><CheckOutlined /></template>
          Buat Tugas Pengambilan
        </a-button>
        <a-button
          size="large"
          @click="handleReset"
          :disabled="submitting"
        >
          <template #icon><ReloadOutlined /></template>
          Reset
        </a-button>
      </a-space>
    </a-space>
  </a-card>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import draggable from 'vuedraggable'
import { 
  EnvironmentOutlined,
  DeleteOutlined,
  CheckOutlined,
  ReloadOutlined,
  HolderOutlined
} from '@ant-design/icons-vue'
import pickupTaskService from '@/services/pickupTaskService'

// Emits
const emit = defineEmits(['task-created'])

// State
const loadingOrders = ref(false)
const loadingDrivers = ref(false)
const submitting = ref(false)
const eligibleOrders = ref([])
const availableDrivers = ref([])
const selectedOrderIds = ref([])
const selectedOrders = ref([])
const selectedDriver = ref(undefined)

// Table columns for eligible orders
const eligibleOrderColumns = [
  {
    title: 'Sekolah',
    key: 'school_info',
    width: 250
  },
  {
    title: 'Koordinat GPS',
    key: 'gps',
    width: 200
  },
  {
    title: 'Jumlah Ompreng',
    key: 'ompreng_count',
    width: 150,
    align: 'center'
  },
  {
    title: 'Tanggal Pengiriman',
    key: 'delivery_date',
    width: 150,
    customRender: ({ record }) => {
      return new Date(record.delivery_date).toLocaleDateString('id-ID')
    }
  }
]

// Row selection configuration
const rowSelection = reactive({
  selectedRowKeys: selectedOrderIds,
  onChange: (selectedRowKeys, selectedRows) => {
    selectedOrderIds.value = selectedRowKeys
    selectedOrders.value = selectedRows.map((row, index) => ({
      ...row,
      route_order: index + 1
    }))
  },
  getCheckboxProps: (record) => ({
    disabled: false,
    name: record.school_name
  })
})

// Computed
const canSubmit = computed(() => {
  return selectedOrders.value.length > 0 && 
         selectedDriver.value !== undefined && 
         !submitting.value
})

// Methods
const fetchEligibleOrders = async () => {
  loadingOrders.value = true
  try {
    const response = await pickupTaskService.getEligibleOrders()
    eligibleOrders.value = response.data.eligible_orders || []
  } catch (error) {
    message.error('Gagal memuat data order yang siap diambil')
    console.error('Error fetching eligible orders:', error)
  } finally {
    loadingOrders.value = false
  }
}

const fetchAvailableDrivers = async () => {
  loadingDrivers.value = true
  try {
    const response = await pickupTaskService.getAvailableDrivers()
    availableDrivers.value = response.data.available_drivers || []
  } catch (error) {
    message.error('Gagal memuat data driver yang tersedia')
    console.error('Error fetching available drivers:', error)
  } finally {
    loadingDrivers.value = false
  }
}

const updateRouteOrder = () => {
  // Update route_order after drag and drop
  selectedOrders.value.forEach((order, index) => {
    order.route_order = index + 1
  })
}

const removeOrder = (deliveryRecordId) => {
  const index = selectedOrders.value.findIndex(o => o.delivery_record_id === deliveryRecordId)
  if (index !== -1) {
    selectedOrders.value.splice(index, 1)
    selectedOrderIds.value = selectedOrders.value.map(o => o.delivery_record_id)
    updateRouteOrder()
  }
}

const handleSubmit = async () => {
  if (!canSubmit.value) {
    return
  }

  submitting.value = true
  try {
    const submitData = {
      task_date: new Date().toISOString(),
      driver_id: selectedDriver.value,
      delivery_records: selectedOrders.value.map(order => ({
        delivery_record_id: order.delivery_record_id,
        route_order: order.route_order
      }))
    }

    const response = await pickupTaskService.createPickupTask(submitData)
    message.success('Tugas pengambilan berhasil dibuat')
    
    // Emit event to parent
    emit('task-created', response.data.pickup_task)
    
    // Reset form
    handleReset()
    
    // Refresh eligible orders
    await fetchEligibleOrders()
  } catch (error) {
    const errorMsg = error.response?.data?.error?.message || 'Gagal membuat tugas pengambilan'
    message.error(errorMsg)
    console.error('Error creating pickup task:', error)
  } finally {
    submitting.value = false
  }
}

const handleReset = () => {
  selectedOrderIds.value = []
  selectedOrders.value = []
  selectedDriver.value = undefined
}

const filterDriverOption = (input, option) => {
  const driver = availableDrivers.value.find(d => d.driver_id === option.value)
  if (!driver) return false
  const searchText = `${driver.full_name} ${driver.phone_number}`.toLowerCase()
  return searchText.includes(input.toLowerCase())
}

const openMaps = (lat, lng) => {
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

// Lifecycle
onMounted(() => {
  fetchEligibleOrders()
  fetchAvailableDrivers()
})
</script>

<style scoped>
.text-gray {
  color: #666;
}

.text-small {
  font-size: 12px;
}

.drag-handle:active {
  cursor: grabbing;
}
</style>
