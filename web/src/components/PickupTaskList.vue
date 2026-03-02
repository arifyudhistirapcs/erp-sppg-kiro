<template>
  <a-table
    :columns="pickupTaskColumns"
    :data-source="pickupTasks"
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
        <template v-if="column.key === 'created_at'">
          {{ formatDate(record.created_at) }}
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
        <template v-else-if="column.key === 'progress'">
          <div>
            <a-progress
              :percent="calculateProgress(record)"
              :status="getProgressStatus(record)"
              size="small"
            />
            <div class="text-small text-gray" style="margin-top: 4px;">
              {{ record.completed_count || 0 }} / {{ record.school_count || 0 }} sekolah selesai
            </div>
          </div>
        </template>
        <template v-else-if="column.key === 'status'">
          <a-tag :color="getStatusColor(record.status)">
            {{ getStatusText(record.status) }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'actions'">
          <a-button type="link" size="small" @click="viewTask(record)">
            Detail
          </a-button>
        </template>
      </template>

      <template #expandedRowRender="{ record }">
        <div style="padding: 12px; background-color: #fafafa;">
          <h4 style="margin-bottom: 12px;">Detail Rute Pengambilan</h4>
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
              <template v-else-if="column.key === 'ompreng_count'">
                <a-tag color="blue">{{ deliveryRecord.ompreng_count }} wadah</a-tag>
              </template>
              <template v-else-if="column.key === 'aksi'">
                <a-button
                  v-if="deliveryRecord.current_stage >= 10 && deliveryRecord.current_stage < 13"
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
import { ref, onMounted, computed, watch, reactive } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { createVNode } from 'vue'
import { ReloadOutlined, EnvironmentOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import pickupTaskService from '@/services/pickupTaskService'

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
const pickupTasks = ref([])
const expandedRowKeys = ref([])

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0
})

// Table columns for pickup tasks
const pickupTaskColumns = [
  {
    title: 'Tanggal',
    dataIndex: 'created_at',
    key: 'created_at',
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
    dataIndex: 'school_count',
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
    title: 'Status',
    key: 'status',
    width: 120,
    align: 'center'
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: 150
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
    title: 'Jumlah Ompreng',
    key: 'ompreng_count',
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
const fetchPickupTasks = async () => {
  loading.value = true
  try {
    const params = {}
    
    if (props.status) {
      params.status = props.status
    }
    // Don't set default status filter - show all tasks (active and completed)
    
    if (props.date) {
      // Convert dayjs object to YYYY-MM-DD format
      const dateStr = props.date.format ? props.date.format('YYYY-MM-DD') : props.date
      params.date = dateStr
    }
    
    if (props.driverId) {
      params.driver_id = props.driverId
    }
    
    const response = await pickupTaskService.getPickupTasks(params)
    pickupTasks.value = response.data.pickup_tasks || []
    pagination.total = pickupTasks.value.length
    
    // Fetch detailed information for each task to get delivery records
    await Promise.all(
      pickupTasks.value.map(async (task) => {
        try {
          const detailResponse = await pickupTaskService.getPickupTask(task.id)
          const detailData = detailResponse.data.pickup_task
          
          // Merge detailed data with summary data
          task.delivery_records = detailData.delivery_records || []
          task.driver = detailData.driver || task.driver
          
          // Calculate completed count (stage >= 13 means pickup is complete)
          task.completed_count = task.delivery_records.filter(
            dr => dr.current_stage >= 13
          ).length
        } catch (error) {
          console.error(`Error fetching details for task ${task.id}:`, error)
        }
      })
    )
  } catch (error) {
    console.error('Error fetching pickup tasks:', error)
    if (error.response?.status === 404) {
      // Endpoint not found - show empty state instead of error
      pickupTasks.value = []
      pagination.total = 0
      console.warn('Pickup tasks endpoint not available yet. Showing empty state.')
    } else if (error.response?.status === 403) {
      message.error('Anda tidak memiliki akses untuk melihat tugas pengambilan')
    } else {
      message.error('Gagal memuat data tugas pengambilan: ' + (error.response?.data?.error?.message || error.message))
    }
  } finally {
    loading.value = false
  }
}

const handleTableChange = (pag, filters, sorter) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  // Pagination handled by frontend, no need to fetch again
}

const calculateProgress = (task) => {
  if (!task.school_count || task.school_count === 0) return 0
  const completed = task.completed_count || 0
  return Math.round((completed / task.school_count) * 100)
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
    cancelled: 'red'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    active: 'Aktif',
    completed: 'Selesai',
    cancelled: 'Dibatalkan'
  }
  return texts[status] || status
}

const getStageColor = (stage) => {
  // For pickup tasks, cap the stage display at 13 (Tiba di SPPG)
  // Even if the actual stage is higher (14-16 for cleaning process)
  const displayStage = Math.min(stage, 13)
  
  const colors = {
    10: 'blue',      // driver_menuju_lokasi_pengambilan
    11: 'orange',    // driver_tiba_di_lokasi_pengambilan
    12: 'purple',    // driver_kembali_ke_sppg
    13: 'green'      // driver_tiba_di_sppg (final stage for pickup)
  }
  return colors[displayStage] || 'default'
}

const getStageText = (stage) => {
  // For pickup tasks, cap the stage display at 13 (Tiba di SPPG)
  // Even if the actual stage is higher (14-16 for cleaning process)
  const displayStage = Math.min(stage, 13)
  
  const texts = {
    10: 'Menuju Lokasi',
    11: 'Tiba di Sekolah',
    12: 'Kembali ke SPPG',
    13: 'Tiba di SPPG'
  }
  return texts[displayStage] || `Stage ${displayStage}`
}

const formatDateTime = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('id-ID', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
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
  // Toggle expand row
  const index = expandedRowKeys.value.indexOf(task.id)
  if (index > -1) {
    // Row is expanded, collapse it
    expandedRowKeys.value.splice(index, 1)
  } else {
    // Row is collapsed, expand it
    expandedRowKeys.value.push(task.id)
  }
}

const openMaps = (lat, lng) => {
  if (!lat || !lng) return
  const url = `https://www.google.com/maps?q=${lat},${lng}`
  window.open(url, '_blank')
}

const getAvailableStages = (currentStage) => {
  // Returns available next stages based on current stage
  // Stage 10 → Stage 11 (Sudah Tiba)
  // Stage 11 → Stage 12 (Dalam Perjalanan Kembali)
  // Stage 12 → Stage 13 (Selesai)
  // Stage 13 → [] (empty, dropdown disabled)
  const stageMap = {
    10: [{ value: 11, label: 'Sudah Tiba (Stage 11)' }],
    11: [{ value: 12, label: 'Dalam Perjalanan Kembali (Stage 12)' }],
    12: [{ value: 13, label: 'Selesai (Stage 13)' }]
  }
  return stageMap[currentStage] || []
}

const getNextStageButtonText = (currentStage) => {
  const textMap = {
    10: 'Sudah Tiba',
    11: 'Kembali ke SPPG',
    12: 'Selesai'
  }
  return textMap[currentStage] || 'Update Status'
}

const getNextStageValue = (currentStage) => {
  const stageMap = {
    10: 11,
    11: 12,
    12: 13
  }
  return stageMap[currentStage]
}

const getStageConfirmationMessage = (currentStage, deliveryRecord) => {
  const schoolName = deliveryRecord.school?.name || deliveryRecord.school_name
  const messageMap = {
    10: `Konfirmasi bahwa driver sudah tiba di ${schoolName}?`,
    11: `Konfirmasi bahwa driver sudah kembali ke SPPG dari ${schoolName}?`,
    12: `Konfirmasi bahwa driver sudah tiba di SPPG dengan ompreng dari ${schoolName}?`
  }
  return messageMap[currentStage] || 'Konfirmasi perubahan status?'
}

const showStageConfirmation = (pickupTaskId, deliveryRecord) => {
  const nextStage = getNextStageValue(deliveryRecord.current_stage)
  
  Modal.confirm({
    title: 'Konfirmasi Perubahan Status',
    icon: createVNode(ExclamationCircleOutlined),
    content: getStageConfirmationMessage(deliveryRecord.current_stage, deliveryRecord),
    okText: 'Ya, Konfirmasi',
    cancelText: 'Batal',
    onOk: async () => {
      await handleStageUpdate(pickupTaskId, deliveryRecord.id, nextStage)
    }
  })
}

const handleStageUpdate = async (pickupTaskId, deliveryRecordId, newStage) => {
  // Map stage to status
  const stageStatusMap = {
    11: 'driver_tiba_di_lokasi_pengambilan',
    12: 'driver_kembali_ke_sppg',
    13: 'driver_tiba_di_sppg'
  }
  
  try {
    await pickupTaskService.updateDeliveryRecordStage(
      pickupTaskId,
      deliveryRecordId,
      {
        stage: newStage,
        status: stageStatusMap[newStage]
      }
    )
    
    message.success('Status berhasil diperbarui')
    emit('stage-updated')
    fetchPickupTasks() // Refresh data
  } catch (error) {
    message.error(error.response?.data?.error?.message || 'Gagal memperbarui status')
    console.error('Error updating stage:', error)
  }
}

// Lifecycle
onMounted(() => {
  fetchPickupTasks()
})

// Watch for prop changes
watch([() => props.date, () => props.driverId, () => props.status], () => {
  fetchPickupTasks()
})

// Expose refresh method for parent component
defineExpose({
  refresh: fetchPickupTasks
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
