<template>
  <div class="delivery-task-list">
    <a-page-header
      title="Manajemen Tugas Pengiriman & Pengambilan"
      sub-title="Kelola tugas pengiriman makanan dan pengambilan ompreng"
    />

    <a-card 
      :tab-list="mainTabList" 
      :active-tab-key="mainTabKey" 
      @tab-change="onMainTabChange"
    >
      <!-- Tab 1: Manajemen Tugas Pengiriman -->
      <template v-if="mainTabKey === 'delivery'">
        <a-space direction="vertical" style="width: 100%" :size="16">
          <!-- Search and Filter with Create Button -->
          <a-row :gutter="16" align="middle">
            <a-col :span="5">
              <a-date-picker
                v-model:value="filterDate"
                placeholder="Pilih tanggal"
                style="width: 100%"
                @change="handleSearch"
                format="DD/MM/YYYY"
              />
            </a-col>
            <a-col :span="5">
              <a-select
                v-model:value="filterDriver"
                placeholder="Pilih driver"
                style="width: 100%"
                @change="handleSearch"
                allow-clear
                show-search
                :filter-option="filterDriverOption"
              >
                <a-select-option 
                  v-for="driver in drivers" 
                  :key="driver.id" 
                  :value="driver.id"
                >
                  {{ driver.full_name }}
                </a-select-option>
              </a-select>
            </a-col>
            <a-col :span="4">
              <a-select
                v-model:value="filterStatus"
                placeholder="Status"
                style="width: 100%"
                @change="handleSearch"
                allow-clear
              >
                <a-select-option value="pending">Menunggu</a-select-option>
                <a-select-option value="in_progress">Dalam Perjalanan</a-select-option>
                <a-select-option value="arrived">Sudah Tiba</a-select-option>
                <a-select-option value="received">Sudah Diterima</a-select-option>
                <a-select-option value="cancelled">Dibatalkan</a-select-option>
              </a-select>
            </a-col>
            <a-col :span="4">
              <a-button @click="resetFilters">
                <template #icon><ReloadOutlined /></template>
                Reset Filter
              </a-button>
            </a-col>
            <a-col :span="6" style="text-align: right;">
              <a-button type="primary" @click="showCreateModal">
                <template #icon><PlusOutlined /></template>
                Buat Tugas Pengiriman
              </a-button>
            </a-col>
          </a-row>

          <!-- Delivery Task List -->
          <DeliveryTaskList 
            :date="filterDate" 
            :driver-id="filterDriver"
            :status="filterStatus"
            :key="deliveryListKey" 
            ref="deliveryTaskListRef"
          />
        </a-space>
      </template>

      <!-- Tab 2: Manajemen Tugas Pengambilan -->
      <template v-else-if="mainTabKey === 'pickup'">
        <a-space direction="vertical" style="width: 100%" :size="16">
          <!-- Search and Filter with Create Button -->
          <a-row :gutter="16" align="middle">
            <a-col :span="5">
              <a-date-picker
                v-model:value="filterPickupDate"
                placeholder="Pilih tanggal"
                style="width: 100%"
                @change="handlePickupSearch"
                format="DD/MM/YYYY"
              />
            </a-col>
            <a-col :span="5">
              <a-select
                v-model:value="filterPickupDriver"
                placeholder="Pilih driver"
                style="width: 100%"
                @change="handlePickupSearch"
                allow-clear
                show-search
                :filter-option="filterDriverOption"
              >
                <a-select-option 
                  v-for="driver in drivers" 
                  :key="driver.id" 
                  :value="driver.id"
                >
                  {{ driver.full_name }}
                </a-select-option>
              </a-select>
            </a-col>
            <a-col :span="4">
              <a-select
                v-model:value="filterPickupStatus"
                placeholder="Status"
                style="width: 100%"
                @change="handlePickupSearch"
                allow-clear
              >
                <a-select-option value="active">Aktif</a-select-option>
                <a-select-option value="completed">Selesai</a-select-option>
                <a-select-option value="cancelled">Dibatalkan</a-select-option>
              </a-select>
            </a-col>
            <a-col :span="4">
              <a-button @click="resetPickupFilters">
                <template #icon><ReloadOutlined /></template>
                Reset Filter
              </a-button>
            </a-col>
            <a-col :span="6" style="text-align: right;">
              <a-button type="primary" @click="showCreatePickupModal">
                <template #icon><PlusOutlined /></template>
                Buat Tugas Pengambilan
              </a-button>
            </a-col>
          </a-row>

          <!-- Pickup Task List -->
          <PickupTaskList 
            :date="filterPickupDate" 
            :driver-id="filterPickupDriver"
            :status="filterPickupStatus"
            :key="pickupListKey" 
            ref="pickupTaskListRef"
          />
        </a-space>
      </template>
    </a-card>

    <!-- Create/Edit Modal -->
    <a-modal
      v-model:open="modalVisible"
      :title="editingTask ? 'Edit Tugas Pengiriman' : 'Buat Tugas Pengiriman'"
      :confirm-loading="submitting"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="900px"
    >
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Date and Time Selection -->
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Tanggal Pengiriman" :required="true">
              <a-date-picker
                v-model:value="formData.task_date"
                style="width: 100%"
                format="DD/MM/YYYY"
                placeholder="Pilih tanggal"
                @change="onDateChange"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Jam Pengiriman" :required="true">
              <a-time-picker
                v-model:value="formData.delivery_time"
                style="width: 100%"
                format="HH:mm"
                placeholder="Pilih jam"
                :minute-step="15"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <!-- Eligible Orders Section -->
        <div>
          <h4>Pilih Order yang Siap Kirim (Sudah Dikemas)</h4>
          <a-table
            :columns="eligibleDeliveryOrderColumns"
            :data-source="readyOrders"
            :row-selection="deliveryRowSelection"
            :loading="loadingOrders"
            :pagination="{ pageSize: 5 }"
            row-key="id"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'school_info'">
                <div>
                  <div><strong>{{ record.school_name }}</strong></div>
                  <div class="text-gray">{{ record.school_address }}</div>
                </div>
              </template>
              <template v-else-if="column.key === 'menu_info'">
                <div>
                  <div>{{ record.menu_item_name }}</div>
                  <div class="text-gray text-small" v-if="record.portions_small > 0 || record.portions_large > 0">
                    <span v-if="record.portions_small > 0">Kecil: {{ record.portions_small }}</span>
                    <span v-if="record.portions_small > 0 && record.portions_large > 0"> | </span>
                    <span v-if="record.portions_large > 0">Besar: {{ record.portions_large }}</span>
                  </div>
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
              <template v-else-if="column.key === 'portions'">
                <a-tag color="blue">{{ record.portions }} porsi</a-tag>
              </template>
            </template>
          </a-table>
          <div v-if="!formData.task_date" style="margin-top: 8px;">
            <a-alert
              message="Pilih tanggal pengiriman terlebih dahulu"
              type="info"
              show-icon
            />
          </div>
          <div v-else-if="readyOrders.length === 0 && !loadingOrders" style="margin-top: 8px;">
            <a-alert
              message="Tidak ada order yang siap kirim"
              description="Belum ada order yang selesai dikemas untuk tanggal ini. Pastikan proses packing sudah selesai di KDS Packing."
              type="warning"
              show-icon
            />
          </div>
        </div>

        <!-- Selected Orders with Route Order (Drag and Drop) -->
        <div v-if="formData.delivery_records.length > 0">
          <a-divider>Urutan Rute Pengiriman ({{ formData.delivery_records.length }} sekolah)</a-divider>
          <a-alert
            message="Seret untuk mengatur urutan pengiriman"
            type="info"
            show-icon
            style="margin-bottom: 12px"
          />
          <draggable
            v-model="formData.delivery_records"
            item-key="id"
            handle=".drag-handle"
            @end="updateDeliveryRouteOrder"
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
                      Menu: {{ element.menu_item_name }} ({{ element.portions }} porsi)
                    </div>
                    <div class="text-gray text-small">
                      GPS: {{ element.latitude?.toFixed(6) }}, {{ element.longitude?.toFixed(6) }}
                    </div>
                  </div>
                  <div>
                    <a-tag color="blue">{{ element.portions }} porsi</a-tag>
                  </div>
                  <a-button 
                    type="text" 
                    danger 
                    size="small"
                    @click="removeDeliveryOrder(element.id)"
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
            v-model:value="formData.driver_id"
            placeholder="Pilih driver yang tersedia"
            show-search
            :filter-option="filterDriverOption"
            :loading="loadingDrivers"
            style="width: 100%"
          >
            <a-select-option 
              v-for="driver in availableDrivers" 
              :key="driver.id" 
              :value="driver.id"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ driver.full_name }}</span>
                <span class="text-gray text-small">{{ driver.phone_number }}</span>
              </div>
            </a-select-option>
          </a-select>
          <div v-if="availableDrivers.length === 0 && !loadingDrivers && formData.task_date" style="margin-top: 8px;">
            <a-alert
              message="Tidak ada driver yang tersedia"
              type="warning"
              show-icon
            />
          </div>
        </a-form-item>
      </a-space>
    </a-modal>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Tugas Pengiriman"
      :footer="null"
      width="900px"
    >
      <div v-if="selectedTask">
        <a-descriptions bordered :column="2">
          <a-descriptions-item label="Tanggal Pengiriman" :span="2">
            {{ formatDate(selectedTask.task_date) }}
          </a-descriptions-item>
          <a-descriptions-item label="Driver">
            {{ selectedTask.driver?.full_name || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="getStatusColor(selectedTask.status)">
              {{ getStatusText(selectedTask.status) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Sekolah Tujuan">
            {{ selectedTask.school?.name }}
          </a-descriptions-item>
          <a-descriptions-item label="Total Porsi">
            {{ selectedTask.portions }} porsi
          </a-descriptions-item>
          <a-descriptions-item label="Urutan Rute">
            {{ selectedTask.route_order }}
          </a-descriptions-item>
          <a-descriptions-item label="Alamat Sekolah" :span="2">
            {{ selectedTask.school?.address }}
          </a-descriptions-item>
        </a-descriptions>

        <a-divider>Informasi Sekolah</a-divider>
        <a-descriptions bordered :column="2">
          <a-descriptions-item label="Kontak Person">
            {{ selectedTask.school?.contact_person || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Telepon">
            {{ selectedTask.school?.phone_number || '-' }}
          </a-descriptions-item>
          <a-descriptions-item label="Jumlah Siswa">
            {{ formatNumber(selectedTask.school?.student_count) }} siswa
          </a-descriptions-item>
          <a-descriptions-item label="Koordinat GPS">
            {{ selectedTask.school?.latitude?.toFixed(6) }}, {{ selectedTask.school?.longitude?.toFixed(6) }}
          </a-descriptions-item>
        </a-descriptions>

        <a-space style="margin-top: 16px">
          <a-button 
            type="primary" 
            @click="openMaps(selectedTask.school?.latitude, selectedTask.school?.longitude)"
          >
            <template #icon><EnvironmentOutlined /></template>
            Buka di Maps
          </a-button>
          <a-button @click="copyCoordinates(selectedTask.school?.latitude, selectedTask.school?.longitude)">
            <template #icon><CopyOutlined /></template>
            Salin Koordinat
          </a-button>
        </a-space>

        <a-divider>Menu Items</a-divider>
        <a-table
          :columns="menuItemColumns"
          :data-source="selectedTask.menu_items"
          :pagination="false"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'recipe_name'">
              {{ record.recipe?.name }}
            </template>
            <template v-else-if="column.key === 'portions'">
              {{ record.portions }} porsi
            </template>
          </template>
        </a-table>
      </div>
    </a-modal>

    <!-- Pickup Task Modal -->
    <a-modal
      v-model:open="pickupModalVisible"
      title="Buat Tugas Pengambilan"
      :confirm-loading="pickupSubmitting"
      @ok="handlePickupSubmit"
      @cancel="handlePickupCancel"
      width="900px"
    >
      <a-space direction="vertical" style="width: 100%" :size="16">
        <!-- Eligible Orders Section -->
        <div>
          <h4>Pilih Order yang Siap Diambil (Stage 9: Sudah Diterima)</h4>
          <a-table
            :columns="eligiblePickupOrderColumns"
            :data-source="eligiblePickupOrders"
            :row-selection="pickupRowSelection"
            :loading="loadingPickupOrders"
            :pagination="{ pageSize: 5 }"
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
              <template v-else-if="column.key === 'delivery_date'">
                {{ formatDate(record.delivery_date) }}
              </template>
            </template>
          </a-table>
        </div>

        <!-- Selected Orders with Route Order (Drag and Drop) -->
        <div v-if="selectedPickupOrders.length > 0">
          <a-divider>Urutan Rute Pengambilan ({{ selectedPickupOrders.length }} sekolah)</a-divider>
          <a-alert
            message="Seret untuk mengatur urutan pengambilan"
            type="info"
            show-icon
            style="margin-bottom: 12px"
          />
          <draggable
            v-model="selectedPickupOrders"
            item-key="delivery_record_id"
            handle=".drag-handle"
            @end="updatePickupRouteOrder"
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
                    @click="removePickupOrder(element.delivery_record_id)"
                  >
                    <template #icon><DeleteOutlined /></template>
                  </a-button>
                </div>
              </a-card>
            </template>
          </draggable>
        </div>

        <!-- Driver Selection -->
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="Tanggal Pengambilan" :required="true">
              <a-date-picker
                v-model:value="pickupTaskDate"
                placeholder="Pilih tanggal"
                style="width: 100%"
                format="DD/MM/YYYY"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="Jam Pengambilan" :required="true">
              <a-time-picker
                v-model:value="pickupTaskTime"
                placeholder="Pilih jam"
                style="width: 100%"
                format="HH:mm"
                :minute-step="15"
              />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item label="Pilih Driver" :required="true">
          <a-select
            v-model:value="selectedPickupDriver"
            placeholder="Pilih driver yang tersedia"
            show-search
            :filter-option="filterPickupDriverOption"
            :loading="loadingPickupDrivers"
            style="width: 100%"
          >
            <a-select-option 
              v-for="driver in availablePickupDrivers" 
              :key="driver.driver_id" 
              :value="driver.driver_id"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <span>{{ driver.full_name }}</span>
                <span class="text-gray text-small">{{ driver.phone_number }}</span>
              </div>
            </a-select-option>
          </a-select>
          <div v-if="availablePickupDrivers.length === 0 && !loadingPickupDrivers" style="margin-top: 8px;">
            <a-alert
              message="Tidak ada driver yang tersedia"
              type="warning"
              show-icon
            />
          </div>
        </a-form-item>
      </a-space>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import { 
  PlusOutlined, 
  ReloadOutlined, 
  DownOutlined, 
  UpOutlined,
  DeleteOutlined,
  EnvironmentOutlined,
  CopyOutlined,
  HolderOutlined
} from '@ant-design/icons-vue'
import deliveryTaskService from '@/services/deliveryTaskService'
import schoolService from '@/services/schoolService'
import pickupTaskService from '@/services/pickupTaskService'
import PickupTaskForm from '@/components/PickupTaskForm.vue'
import PickupTaskList from '@/components/PickupTaskList.vue'
import DeliveryTaskList from '@/components/DeliveryTaskList.vue'
import draggable from 'vuedraggable'

const loading = ref(false)
const submitting = ref(false)
const modalVisible = ref(false)
const detailModalVisible = ref(false)
const editingTask = ref(null)
const selectedTask = ref(null)
const deliveryTasks = ref([])
const drivers = ref([])
const availableDrivers = ref([])
const schools = ref([])
const recipes = ref([])
const readyOrders = ref([])
const selectedOrder = ref(null)
const selectedOrderId = ref(null)
const selectedDeliveryOrderIds = ref([])
const formRef = ref()
const loadingOrders = ref(false)
const loadingDrivers = ref(false)

// Main tab state
const mainTabKey = ref('delivery')
const mainTabList = [
  {
    key: 'delivery',
    tab: 'Pengiriman'
  },
  {
    key: 'pickup',
    tab: 'Pengambilan'
  }
]

// Pickup task state
const pickupTabKey = ref('create')
const pickupListKey = ref(0)
const pickupModalVisible = ref(false)
const pickupFormRef = ref()
const pickupSubmitting = ref(false)

// Pickup form state
const loadingPickupOrders = ref(false)
const loadingPickupDrivers = ref(false)
const eligiblePickupOrders = ref([])
const availablePickupDrivers = ref([])
const selectedPickupOrderIds = ref([])
const selectedPickupOrders = ref([])
const selectedPickupDriver = ref(undefined)
const pickupTaskDate = ref(null)
const pickupTaskTime = ref(null)

// Filters
const filterDate = ref(null)
const filterDriver = ref(undefined)
const filterStatus = ref(undefined)
const deliveryListKey = ref(0)
const deliveryTaskListRef = ref(null)

// Pickup filters
const filterPickupDate = ref(null)
const filterPickupDriver = ref(undefined)
const filterPickupStatus = ref(undefined)
const pickupTaskListRef = ref(null)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  task_date: null,
  delivery_time: null,
  delivery_records: [], // Array of {id, route_order, school_name, address, coordinates}
  driver_id: undefined
})

const rules = {
  task_date: [{ required: true, message: 'Tanggal pengiriman wajib diisi' }],
  delivery_time: [{ required: true, message: 'Jam pengiriman wajib diisi' }],
  delivery_records: [
    { required: true, message: 'Minimal 1 order harus dipilih' },
    { type: 'array', min: 1, message: 'Minimal 1 order harus dipilih' }
  ],
  driver_id: [{ required: true, message: 'Driver wajib dipilih' }]
}

const columns = [
  {
    title: 'Tanggal',
    key: 'task_date',
    width: 120,
    sorter: true
  },
  {
    title: 'Driver',
    key: 'driver',
    width: 150
  },
  {
    title: 'Sekolah & Porsi',
    key: 'school',
    width: 200
  },
  {
    title: 'Urutan',
    key: 'route_order',
    width: 80,
    align: 'center'
  },
  {
    title: 'Status',
    key: 'status',
    width: 120
  },
  {
    title: 'Menu Items',
    key: 'menu_items',
    width: 250
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 200
  }
]

const menuItemColumns = [
  {
    title: 'Menu',
    key: 'recipe_name'
  },
  {
    title: 'Porsi',
    key: 'portions',
    width: 100
  }
]

// Pickup task modal columns
const eligiblePickupOrderColumns = [
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
    width: 150
  }
]

// Delivery task modal columns
const eligibleDeliveryOrderColumns = [
  {
    title: 'Sekolah',
    key: 'school_info',
    width: 250
  },
  {
    title: 'Menu',
    key: 'menu_info',
    width: 200
  },
  {
    title: 'Koordinat GPS',
    key: 'gps',
    width: 180
  },
  {
    title: 'Jumlah Porsi',
    key: 'portions',
    width: 120,
    align: 'center'
  }
]

const fetchDeliveryTasks = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize
    }
    
    if (filterDate.value) {
      params.date = dayjs(filterDate.value).format('YYYY-MM-DD')
    }
    if (filterDriver.value) {
      params.driver_id = filterDriver.value
    }
    if (filterStatus.value) {
      params.status = filterStatus.value
    }

    const response = await deliveryTaskService.getDeliveryTasks(params)
    deliveryTasks.value = response.data.delivery_tasks || []
    pagination.total = response.data.total || 0
  } catch (error) {
    message.error('Gagal memuat data tugas pengiriman')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const fetchDrivers = async () => {
  try {
    const response = await deliveryTaskService.getDrivers()
    drivers.value = response.data.data || []
  } catch (error) {
    console.error('Gagal memuat data driver:', error)
  }
}

const fetchSchools = async () => {
  try {
    const response = await schoolService.getSchools({ is_active: true })
    schools.value = response.data.schools || []
  } catch (error) {
    console.error('Gagal memuat data sekolah:', error)
  }
}

const fetchRecipes = async () => {
  try {
    const response = await deliveryTaskService.getAvailableRecipes()
    recipes.value = response.data.recipes || []
  } catch (error) {
    console.error('Gagal memuat data resep:', error)
  }
}

const fetchReadyOrders = async (date) => {
  if (!date) return
  
  loadingOrders.value = true
  try {
    // Fetch delivery records yang sudah selesai dikemas (status = selesai_dipacking)
    const dateStr = dayjs(date).format('YYYY-MM-DD')
    console.log('[DeliveryTask] Fetching ready orders for date:', dateStr)
    const response = await deliveryTaskService.getReadyOrders(dateStr)
    console.log('[DeliveryTask] Ready orders response:', response)
    readyOrders.value = response.data.orders || []
    console.log('[DeliveryTask] Ready orders count:', readyOrders.value.length)
  } catch (error) {
    console.error('Gagal memuat data order:', error)
    message.error('Gagal memuat data order yang siap kirim')
  } finally {
    loadingOrders.value = false
  }
}

const fetchAvailableDrivers = async (date) => {
  if (!date) return
  
  loadingDrivers.value = true
  try {
    // Fetch drivers yang tidak sedang bertugas pada tanggal tersebut
    const dateStr = dayjs(date).format('YYYY-MM-DD')
    console.log('[DeliveryTask] Fetching available drivers for date:', dateStr)
    const response = await deliveryTaskService.getAvailableDrivers(dateStr)
    console.log('[DeliveryTask] Available drivers response:', response)
    availableDrivers.value = response.data.drivers || []
    console.log('[DeliveryTask] Available drivers count:', availableDrivers.value.length)
  } catch (error) {
    console.error('Gagal memuat data driver:', error)
    message.error('Gagal memuat data driver yang tersedia')
  } finally {
    loadingDrivers.value = false
  }
}

const onDateChange = (date) => {
  if (date) {
    fetchReadyOrders(date)
    fetchAvailableDrivers(date)
  } else {
    readyOrders.value = []
    availableDrivers.value = []
  }
  // Reset selections when date changes
  formData.delivery_records = []
  formData.driver_id = undefined
  selectedOrder.value = null
}

const fetchSchoolDetails = async (schoolId, recordIndex) => {
  try {
    const school = schools.value.find(s => s.id === schoolId)
    if (school) {
      formData.delivery_records[recordIndex].address = school.address || '-'
      formData.delivery_records[recordIndex].latitude = school.latitude
      formData.delivery_records[recordIndex].longitude = school.longitude
    }
  } catch (error) {
    console.error('Failed to fetch school details:', error)
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchDeliveryTasks()
}

const handleSearch = () => {
  // Trigger refresh on DeliveryTaskList component
  if (deliveryTaskListRef.value) {
    deliveryTaskListRef.value.refresh()
  }
}

const resetFilters = () => {
  filterDate.value = null
  filterDriver.value = undefined
  filterStatus.value = undefined
  handleSearch()
}

const handlePickupSearch = () => {
  // Trigger refresh on PickupTaskList component
  if (pickupTaskListRef.value) {
    pickupTaskListRef.value.refresh()
  }
}

const resetPickupFilters = () => {
  filterPickupDate.value = null
  filterPickupDriver.value = undefined
  filterPickupStatus.value = undefined
  handlePickupSearch()
}

const showCreateModal = () => {
  editingTask.value = null
  resetForm()
  modalVisible.value = true
}

const editTask = (task) => {
  editingTask.value = task
  Object.assign(formData, {
    task_date: dayjs(task.task_date),
    delivery_time: task.delivery_time ? dayjs(task.delivery_time, 'HH:mm') : null,
    delivery_record_id: task.delivery_record_id,
    driver_id: task.driver_id
  })
  
  // Load data for the selected date
  if (task.task_date) {
    fetchReadyOrders(dayjs(task.task_date))
    fetchAvailableDrivers(dayjs(task.task_date))
  }
  
  modalVisible.value = true
}

const viewTask = (task) => {
  selectedTask.value = task
  detailModalVisible.value = true
}

const handleSubmit = async () => {
  // Validation
  if (formData.delivery_records.length === 0) {
    message.warning('Pilih minimal 1 order untuk dikirim')
    return
  }
  
  if (!formData.task_date) {
    message.warning('Pilih tanggal pengiriman terlebih dahulu')
    return
  }
  
  if (!formData.delivery_time) {
    message.warning('Pilih jam pengiriman terlebih dahulu')
    return
  }
  
  if (!formData.driver_id) {
    message.warning('Pilih driver terlebih dahulu')
    return
  }

  submitting.value = true
  try {
    const submitData = {
      task_date: dayjs(formData.task_date).format('YYYY-MM-DD'),
      driver_id: formData.driver_id,
      delivery_records: formData.delivery_records.map(r => ({
        delivery_record_id: r.id,
        route_order: r.route_order
      }))
    }

    if (editingTask.value) {
      await deliveryTaskService.updateDeliveryTask(editingTask.value.id, submitData)
      message.success('Tugas pengiriman berhasil diperbarui')
    } else {
      await deliveryTaskService.createDeliveryTask(submitData)
      message.success('Tugas pengiriman berhasil dibuat')
    }

    modalVisible.value = false
    resetForm()
    deliveryListKey.value++
  } catch (error) {
    const errorMsg = error.response?.data?.message || error.response?.data?.error?.message || 'Gagal menyimpan tugas pengiriman'
    message.error(errorMsg)
    console.error('Error creating delivery task:', error)
    console.error('Error details:', error.response?.data)
  } finally {
    submitting.value = false
  }
}

const updateTaskStatus = async (taskId, status) => {
  try {
    await deliveryTaskService.updateDeliveryTaskStatus(taskId, status)
    message.success('Status tugas berhasil diperbarui')
    fetchDeliveryTasks()
  } catch (error) {
    message.error('Gagal memperbarui status tugas')
    console.error(error)
  }
}

const deleteTask = async (id) => {
  try {
    await deliveryTaskService.deleteDeliveryTask(id)
    message.success('Tugas pengiriman berhasil dihapus')
    fetchDeliveryTasks()
  } catch (error) {
    message.error('Gagal menghapus tugas pengiriman')
    console.error(error)
  }
}

const handleCancel = () => {
  modalVisible.value = false
  resetForm()
}

const resetForm = () => {
  Object.assign(formData, {
    task_date: null,
    delivery_time: null,
    delivery_records: [],
    driver_id: undefined
  })
  selectedOrder.value = null
  selectedDeliveryOrderIds.value = []
  readyOrders.value = []
  availableDrivers.value = []
}

const filterOrderOption = (input, option) => {
  const order = readyOrders.value.find(o => o.id === option.value)
  if (!order) return false
  const searchText = `${order.school_name} ${order.menu_item_name}`.toLowerCase()
  return searchText.includes(input.toLowerCase())
}

// Filter functions
const filterDriverOption = (input, option) => {
  const driver = availableDrivers.value.find(d => d.id === option.value)
  return driver?.full_name?.toLowerCase().includes(input.toLowerCase())
}

const filterSchoolOption = (input, option) => {
  const school = schools.value.find(s => s.id === option.value)
  return school?.name?.toLowerCase().includes(input.toLowerCase())
}

const filterRecipeOption = (input, option) => {
  const recipe = recipes.value.find(r => r.id === option.value)
  return recipe?.name?.toLowerCase().includes(input.toLowerCase())
}

// Utility functions
const formatDate = (date) => {
  return dayjs(date).format('DD/MM/YYYY')
}

const formatNumber = (value) => {
  return new Intl.NumberFormat('id-ID').format(value)
}

const getDriverInitials = (name) => {
  if (!name) return '?'
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
}

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    in_progress: 'blue',
    arrived: 'cyan',
    received: 'green',
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    pending: 'Menunggu',
    in_progress: 'Dalam Perjalanan',
    arrived: 'Sudah Tiba',
    received: 'Sudah Diterima',
    cancelled: 'Dibatalkan'
  }
  return texts[status] || status
}

const openMaps = (lat, lng) => {
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

const copyCoordinates = async (lat, lng) => {
  try {
    await navigator.clipboard.writeText(`${lat}, ${lng}`)
    message.success('Koordinat berhasil disalin')
  } catch (error) {
    message.error('Gagal menyalin koordinat')
  }
}

const handlePickupTaskCreated = () => {
  message.success('Tugas pengambilan berhasil dibuat')
  // Switch to list tab
  pickupTabKey.value = 'list'
  // Force refresh the pickup task list by updating the key
  pickupListKey.value++
}

const onMainTabChange = (key) => {
  mainTabKey.value = key
}

// Pickup task modal functions
const showCreatePickupModal = async () => {
  pickupModalVisible.value = true
  await fetchEligiblePickupOrders()
  await fetchAvailablePickupDrivers()
}

const fetchEligiblePickupOrders = async () => {
  loadingPickupOrders.value = true
  try {
    const response = await pickupTaskService.getEligibleOrders()
    eligiblePickupOrders.value = response.data.eligible_orders || []
  } catch (error) {
    message.error('Gagal memuat data order yang siap diambil')
    console.error('Error fetching eligible orders:', error)
  } finally {
    loadingPickupOrders.value = false
  }
}

const fetchAvailablePickupDrivers = async () => {
  loadingPickupDrivers.value = true
  try {
    const response = await pickupTaskService.getAvailableDrivers()
    availablePickupDrivers.value = response.data.available_drivers || []
  } catch (error) {
    message.error('Gagal memuat data driver yang tersedia')
    console.error('Error fetching available drivers:', error)
  } finally {
    loadingPickupDrivers.value = false
  }
}

const handlePickupOrderSelection = (selectedRowKeys, selectedRows) => {
  selectedPickupOrderIds.value = selectedRowKeys
  selectedPickupOrders.value = selectedRows.map((row, index) => ({
    ...row,
    route_order: index + 1
  }))
}

// Delivery order selection
const handleDeliveryOrderSelection = (selectedRowKeys, selectedRows) => {
  selectedDeliveryOrderIds.value = selectedRowKeys
  formData.delivery_records = selectedRows.map((row, index) => ({
    ...row,
    route_order: index + 1
  }))
}

// Delivery row selection configuration
const deliveryRowSelection = {
  selectedRowKeys: selectedDeliveryOrderIds,
  onChange: handleDeliveryOrderSelection,
  getCheckboxProps: (record) => ({
    disabled: false,
    name: record.school_name
  })
}

const updateDeliveryRouteOrder = () => {
  formData.delivery_records.forEach((order, index) => {
    order.route_order = index + 1
  })
}

const removeDeliveryOrder = (orderId) => {
  const index = formData.delivery_records.findIndex(o => o.id === orderId)
  if (index !== -1) {
    formData.delivery_records.splice(index, 1)
    selectedDeliveryOrderIds.value = formData.delivery_records.map(o => o.id)
    updateDeliveryRouteOrder()
  }
}

// Pickup row selection configuration
const pickupRowSelection = {
  selectedRowKeys: selectedPickupOrderIds,
  onChange: handlePickupOrderSelection,
  getCheckboxProps: (record) => ({
    disabled: false,
    name: record.school_name
  })
}

const updatePickupRouteOrder = () => {
  selectedPickupOrders.value.forEach((order, index) => {
    order.route_order = index + 1
  })
}

const removePickupOrder = (deliveryRecordId) => {
  const index = selectedPickupOrders.value.findIndex(o => o.delivery_record_id === deliveryRecordId)
  if (index !== -1) {
    selectedPickupOrders.value.splice(index, 1)
    selectedPickupOrderIds.value = selectedPickupOrders.value.map(o => o.delivery_record_id)
    updatePickupRouteOrder()
  }
}

const handlePickupSubmit = async () => {
  if (selectedPickupOrders.value.length === 0) {
    message.warning('Pilih minimal 1 order untuk diambil')
    return
  }
  
  if (!pickupTaskDate.value) {
    message.warning('Pilih tanggal pengambilan terlebih dahulu')
    return
  }
  
  if (!pickupTaskTime.value) {
    message.warning('Pilih jam pengambilan terlebih dahulu')
    return
  }
  
  if (!selectedPickupDriver.value) {
    message.warning('Pilih driver terlebih dahulu')
    return
  }

  pickupSubmitting.value = true
  try {
    // Combine date and time
    const taskDateTime = dayjs(pickupTaskDate.value)
      .hour(pickupTaskTime.value.hour())
      .minute(pickupTaskTime.value.minute())
      .second(0)
      .millisecond(0)
    
    const submitData = {
      task_date: taskDateTime.toISOString(),
      driver_id: selectedPickupDriver.value,
      delivery_records: selectedPickupOrders.value.map(order => ({
        delivery_record_id: order.delivery_record_id,
        route_order: order.route_order
      }))
    }

    await pickupTaskService.createPickupTask(submitData)
    message.success('Tugas pengambilan berhasil dibuat')
    
    // Close modal and reset
    pickupModalVisible.value = false
    resetPickupForm()
    
    // Refresh pickup task list
    pickupListKey.value++
  } catch (error) {
    const errorMsg = error.response?.data?.error?.message || 'Gagal membuat tugas pengambilan'
    message.error(errorMsg)
    console.error('Error creating pickup task:', error)
  } finally {
    pickupSubmitting.value = false
  }
}

const handlePickupCancel = () => {
  pickupModalVisible.value = false
  resetPickupForm()
}

const resetPickupForm = () => {
  selectedPickupOrderIds.value = []
  selectedPickupOrders.value = []
  selectedPickupDriver.value = undefined
  pickupTaskDate.value = null
  pickupTaskTime.value = null
}

const filterPickupDriverOption = (input, option) => {
  const driver = availablePickupDrivers.value.find(d => d.driver_id === option.value)
  if (!driver) return false
  const searchText = `${driver.full_name} ${driver.phone_number}`.toLowerCase()
  return searchText.includes(input.toLowerCase())
}

onMounted(() => {
  fetchDeliveryTasks()
  fetchDrivers()
  fetchSchools()
  fetchRecipes()
})
</script>

<style scoped>
.delivery-task-list {
  padding: 24px;
}

.menu-item {
  display: block;
  font-size: 12px;
  color: #666;
}

.menu-items-section {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  padding: 16px;
  background-color: #fafafa;
}

.menu-item-row {
  margin-bottom: 8px;
}

.menu-item-row:last-child {
  margin-bottom: 0;
}

.text-gray {
  color: #666;
  font-size: 12px;
}

.text-small {
  font-size: 12px;
}

.drag-handle:active {
  cursor: grabbing;
}
</style>