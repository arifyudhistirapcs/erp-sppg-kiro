<template>
  <a-table
    :columns="deliveryTaskColumns"
    :data-source="deliveryTasks"
    :loading="loading"
    :pagination="{
      current: pagination.current,
      pageSize: pagination.pageSize,
      total: pagination.total,
      showSizeChanger: true,
      showTotal: (total) => `Total ${total} tugas`
    }"
    @change="handleTableChange"
    :expandable="expandableConfig"
    v-model:expandedRowKeys="expandedRowKeys"
    row-key="id"
  >
    <template #bodyCell="{ column, record }">
      <template v-if="column.key === 'task_date'">
        {{ formatDate(record.task_date) }}
      </template>
      <template v-else-if="column.key === 'task_time'">
        {{ formatTime(record.task_date) }}
      </template>
      <template v-else-if="column.key === 'driver'">
        <a-space>
          <a-avatar size="small">{{ getDriverInitials(record.driver?.full_name) }}</a-avatar>
          {{ record.driver?.full_name || '-' }}
        </a-space>
      </template>
      <template v-else-if="column.key === 'school_count'">
        {{ record.delivery_records?.length || 0 }}
      </template>
      <template v-else-if="column.key === 'progress'">
        <div>
          <a-progress
            :percent="calculateProgress(record)"
            :status="getProgressStatus(record)"
            size="small"
          />
          <div class="text-small text-gray" style="margin-top: 4px;">
            {{ record.completed_count || 0 }} / {{ record.delivery_records?.length || 0 }} sekolah selesai
          </div>
        </div>
      </template>
      <template v-else-if="column.key === 'actions'">
        <a-button type="link" size="small" @click="viewTask(record)">
          Detail
        </a-button>
      </template>
    </template>

    <template #expandedRowRender="{ record }">
      <div style="padding: 12px; background-color: #fafafa;">
        <h4 style="margin-bottom: 12px;">Detail Rute Pengiriman</h4>
        <a-table
          :columns="deliveryRecordColumns"
          :data-source="record.delivery_records || []"
          :pagination="false"
          row-key="id"
          size="small"
        >
          <template #bodyCell="{ column, record: deliveryRecord }">
            <template v-if="column.key === 'route_order'">
              <a-tag color="blue">Rute {{ deliveryRecord.route_order }}</a-tag>
            </template>
            <template v-else-if="column.key === 'school_info'">
              <div>
                <div><strong>{{ deliveryRecord.school?.name || deliveryRecord.school_name }}</strong></div>
                <div class="text-gray text-small">{{ deliveryRecord.school?.address || deliveryRecord.school_address }}</div>
              </div>
            </template>
            <template v-else-if="column.key === 'gps'">
              <div class="text-small">
                {{ deliveryRecord.school?.latitude?.toFixed(6) || deliveryRecord.latitude?.toFixed(6) }}, 
                {{ deliveryRecord.school?.longitude?.toFixed(6) || deliveryRecord.longitude?.toFixed(6) }}
                <a-button 
                  type="link" 
                  size="small" 
                  @click="openMaps(
                    deliveryRecord.school?.latitude || deliveryRecord.latitude, 
                    deliveryRecord.school?.longitude || deliveryRecord.longitude
                  )"
                >
                  <template #icon><EnvironmentOutlined /></template>
                </a-button>
              </div>
            </template>
            <template v-else-if="column.key === 'stage'">
              <a-tag :color="getStageColor(deliveryRecord.current_stage)">
                {{ getStageText(deliveryRecord.current_stage) }}
              </a-tag>
            </template>
            <template v-else-if="column.key === 'portions'">
              <a-tag color="blue">{{ deliveryRecord.portions }} porsi</a-tag>
            </template>
            <template v-else-if="column.key === 'aksi'">
              <a-button
                v-if="deliveryRecord.current_stage !== 9"
                type="primary"
                size="small"
                @click="showStageConfirmation(record.id, deliveryRecord)"
              >
                {{ getNextStageButtonText(deliveryRecord.current_stage) }}
              </a-button>
              <span v-else class="text-gray">-</span>
            </template>
          </template>
        </a-table>
      </div>
    </template>
  </a-table>
</template>

<script setup>
import { ref, onMounted, watch, reactive } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { EnvironmentOutlined, DownOutlined } from '@ant-design/icons-vue'
import deliveryTaskService from '@/services/deliveryTaskService'

// Props
const props = defineProps({
  date: {
    type: String,
    default: null
  },
  driverId: {
    type: Number,
    default: null
  },
  status: {
    type: String,
    default: null
  }
})

// Emits
const emit = defineEmits(['task-selected', 'stage-updated'])

// State
const loading = ref(false)
const deliveryTasks = ref([])
const expandedRowKeys = ref([])

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

// Table columns for delivery tasks
const deliveryTaskColumns = [
  {
    title: 'Tanggal',
    dataIndex: 'task_date',
    key: 'task_date',
    width: 120,
    sorter: true
  },
  {
    title: 'Jam',
    dataIndex: 'task_date',
    key: 'task_time',
    width: 80
  },
  {
    title: 'Driver',
    key: 'driver',
    width: 150
  },
  {
    title: 'Jumlah Sekolah',
    key: 'school_count',
    width: 120,
    align: 'center'
  },
  {
    title: 'Progress',
    key: 'progress',
    width: 200
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 100
  }
]

// Table columns for delivery records (expanded view)
const deliveryRecordColumns = [
  {
    title: 'Urutan',
    key: 'route_order',
    width: 80,
    align: 'center'
  },
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
    title: 'Stage',
    key: 'stage',
    width: 200
  },
  {
    title: 'Jumlah Porsi',
    key: 'portions',
    width: 120,
    align: 'center'
  },
  {
    title: 'Aksi',
    key: 'aksi',
    width: 220,
    align: 'center'
  }
]

// Expandable configuration
const expandableConfig = {
  expandedRowRender: (record) => record,
  rowExpandable: (record) => (record.delivery_records?.length || 0) > 0
}

// Methods
const fetchDeliveryTasks = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.current,
      page_size: pagination.pageSize
    }
    
    if (props.status) {
      params.status = props.status
    }
    
    if (props.date) {
      // Convert dayjs object to YYYY-MM-DD format
      const dateStr = props.date.format ? props.date.format('YYYY-MM-DD') : props.date
      params.date = dateStr
    }
    
    if (props.driverId) {
      params.driver_id = props.driverId
    }
    
    // Fetch delivery records instead of delivery tasks
    // We need records that are assigned to drivers (stage >= 6)
    const response = await deliveryTaskService.getDeliveryTasks(params)
    
    console.log('Raw API response:', response.data)
    
    // Group delivery tasks by driver and date
    const tasksMap = new Map()
    
    if (response.data.delivery_tasks && response.data.delivery_tasks.length > 0) {
      response.data.delivery_tasks.forEach(task => {
        const key = `${task.driver_id}_${task.task_date}`
        
        if (!tasksMap.has(key)) {
          tasksMap.set(key, {
            id: task.id,
            task_date: task.task_date,
            driver: task.driver,
            driver_id: task.driver_id,
            status: 'active',
            delivery_records: [],
            completed_count: 0
          })
        }
        
        const taskGroup = tasksMap.get(key)
        
        // Add task as a delivery record with proper structure
        taskGroup.delivery_records.push({
          id: task.id,
          route_order: task.route_order,
          school: task.school,
          school_name: task.school?.name,
          school_address: task.school?.address,
          latitude: task.school?.latitude,
          longitude: task.school?.longitude,
          portions: task.portions,
          current_stage: task.current_stage || 1, // Default to stage 1 if not set
          menu_items: task.menu_items || []
        })
        
        // Count completed (stage 9)
        if (task.current_stage === 9) {
          taskGroup.completed_count++
        }
      })
    }
    
    deliveryTasks.value = Array.from(tasksMap.values())
    pagination.total = deliveryTasks.value.length
    
  } catch (error) {
    console.error('Error fetching delivery tasks:', error)
    if (error.response?.status === 403) {
      message.error('Anda tidak memiliki akses untuk melihat tugas pengiriman')
    } else {
      message.error('Gagal memuat data tugas pengiriman: ' + (error.response?.data?.error?.message || error.message))
    }
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
}

const calculateProgress = (task) => {
  const total = task.delivery_records?.length || 0
  if (total === 0) return 0
  const completed = task.completed_count || 0
  return Math.round((completed / total) * 100)
}

const getProgressStatus = (task) => {
  const progress = calculateProgress(task)
  if (progress === 100) return 'success'
  if (progress > 0) return 'active'
  return 'normal'
}

const getStatusColor = (status) => {
  const colors = {
    active: 'blue',
    completed: 'green',
    cancelled: 'red',
    pending: 'orange',
    in_progress: 'blue',
    arrived: 'purple',
    received: 'green'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    active: 'Aktif',
    completed: 'Selesai',
    cancelled: 'Dibatalkan',
    pending: 'Menunggu',
    in_progress: 'Dalam Perjalanan',
    arrived: 'Sudah Tiba',
    received: 'Sudah Diterima'
  }
  return texts[status] || status
}

const getStageColor = (stage) => {
  if (!stage && stage !== 0) return 'default'
  const colors = {
    1: 'orange',     // pending
    2: 'blue',       // in_progress
    3: 'purple',     // arrived
    9: 'green'       // received
  }
  return colors[stage] || 'default'
}

const getStageText = (stage) => {
  if (!stage && stage !== 0) return 'Belum Dimulai'
  const texts = {
    1: 'Menunggu',
    2: 'Dalam Perjalanan',
    3: 'Sudah Tiba',
    9: 'Sudah Diterima'
  }
  return texts[stage] || `Stage ${stage}`
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('id-ID', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

const formatTime = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getDriverInitials = (name) => {
  if (!name) return '?'
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
}

const viewTask = (task) => {
  const index = expandedRowKeys.value.indexOf(task.id)
  if (index > -1) {
    expandedRowKeys.value.splice(index, 1)
  } else {
    expandedRowKeys.value.push(task.id)
  }
}

const updateTaskStatus = async (taskId, status) => {
  try {
    // Update all delivery records in this task
    const task = deliveryTasks.value.find(t => t.id === taskId)
    if (task) {
      for (const record of task.delivery_records) {
        await deliveryTaskService.updateDeliveryTaskStatus(record.id, status)
      }
    }
    message.success('Status tugas berhasil diperbarui')
    fetchDeliveryTasks()
  } catch (error) {
    message.error('Gagal memperbarui status tugas')
    console.error(error)
  }
}

const openMaps = (lat, lng) => {
  if (!lat || !lng) return
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

const getNextStageButtonText = (currentStage) => {
  // Handle undefined or null stage - default to stage 1
  const stage = currentStage || 1
  const textMap = {
    1: 'Mulai Perjalanan',
    2: 'Sudah Tiba',
    3: 'Sudah Diterima'
  }
  return textMap[stage] || 'Update Status'
}

const getNextStageValue = (currentStage) => {
  // Handle undefined or null stage - default to stage 1
  const stage = currentStage || 1
  const stageMap = {
    1: 2,
    2: 3,
    3: 9
  }
  return stageMap[stage]
}

const getStageConfirmationMessage = (currentStage, deliveryRecord) => {
  // Handle undefined or null stage - default to stage 1
  const stage = currentStage || 1
  const schoolName = deliveryRecord.school?.name || deliveryRecord.school_name
  const messageMap = {
    1: `Konfirmasi bahwa driver mulai perjalanan ke ${schoolName}?`,
    2: `Konfirmasi bahwa driver sudah tiba di ${schoolName}?`,
    3: `Konfirmasi bahwa makanan sudah diterima oleh ${schoolName}?`
  }
  return messageMap[stage] || 'Konfirmasi perubahan status?'
}

const showStageConfirmation = (taskId, deliveryRecord) => {
  const nextStage = getNextStageValue(deliveryRecord.current_stage)
  
  Modal.confirm({
    title: 'Konfirmasi Perubahan Status',
    content: getStageConfirmationMessage(deliveryRecord.current_stage, deliveryRecord),
    okText: 'Ya, Konfirmasi',
    cancelText: 'Batal',
    onOk: async () => {
      await handleStageUpdate(deliveryRecord.id, nextStage)
    }
  })
}

const handleStageUpdate = async (deliveryRecordId, newStage) => {
  const stageStatusMap = {
    2: 'in_progress',
    3: 'arrived',
    9: 'received'
  }
  
  try {
    console.log('Updating delivery task:', deliveryRecordId, 'to stage:', newStage, 'status:', stageStatusMap[newStage])
    await deliveryTaskService.updateDeliveryTaskStatus(deliveryRecordId, stageStatusMap[newStage])
    
    message.success('Status berhasil diperbarui')
    emit('stage-updated')
    
    // Wait a bit for database to commit before refreshing
    setTimeout(async () => {
      console.log('Refreshing delivery tasks...')
      await fetchDeliveryTasks()
      console.log('Delivery tasks after refresh:', deliveryTasks.value)
    }, 500)
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal memperbarui status')
    console.error('Error updating stage:', error)
  }
}

// Lifecycle
onMounted(() => {
  fetchDeliveryTasks()
})

// Watch for prop changes
watch([() => props.date, () => props.driverId, () => props.status], () => {
  fetchDeliveryTasks()
})

// Expose refresh method for parent component
defineExpose({
  refresh: fetchDeliveryTasks
})
</script>

<style scoped>
.text-gray {
  color: #666;
}

.text-small {
  font-size: 12px;
}
</style>
