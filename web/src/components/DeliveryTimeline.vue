<template>
  <div class="delivery-timeline">
    <a-timeline>
      <a-timeline-item
        v-for="stage in stages"
        :key="stage.status"
        :color="getStageColor(stage.status)"
      >
        <template #dot>
          <check-circle-outlined v-if="isCompleted(stage.status)" style="font-size: 16px" />
          <loading-outlined v-else-if="isInProgress(stage.status)" style="font-size: 16px" />
          <clock-circle-outlined v-else style="font-size: 16px" />
        </template>
        
        <div class="timeline-content">
          <div class="stage-title">{{ stage.title }}</div>
          <div class="stage-description">{{ stage.description }}</div>
          <div v-if="getStageTimestamp(stage.status)" class="stage-timestamp">
            {{ formatTimestamp(getStageTimestamp(stage.status)) }}
          </div>
        </div>
      </a-timeline-item>
    </a-timeline>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import {
  CheckCircleOutlined,
  LoadingOutlined,
  ClockCircleOutlined
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import 'dayjs/locale/id'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'

dayjs.extend(utc)
dayjs.extend(timezone)
dayjs.locale('id')

const props = defineProps({
  currentStatus: {
    type: String,
    required: true
  },
  activityLog: {
    type: Array,
    default: () => []
  }
})

// Define all 15 lifecycle stages
const stages = [
  {
    status: 'sedang_dimasak',
    title: 'Sedang Dimasak',
    description: 'Menu sedang dalam proses memasak'
  },
  {
    status: 'selesai_dimasak',
    title: 'Selesai Dimasak',
    description: 'Proses memasak telah selesai'
  },
  {
    status: 'siap_dipacking',
    title: 'Siap Dipacking',
    description: 'Menu siap untuk dikemas'
  },
  {
    status: 'selesai_dipacking',
    title: 'Selesai Dipacking',
    description: 'Menu telah dikemas'
  },
  {
    status: 'siap_dikirim',
    title: 'Siap Dikirim',
    description: 'Menu siap untuk dikirim'
  },
  {
    status: 'diperjalanan',
    title: 'Diperjalanan',
    description: 'Driver sedang dalam perjalanan ke sekolah'
  },
  {
    status: 'sudah_sampai_sekolah',
    title: 'Sudah Sampai Sekolah',
    description: 'Driver telah tiba di sekolah'
  },
  {
    status: 'sudah_diterima_pihak_sekolah',
    title: 'Sudah Diterima',
    description: 'Menu telah diterima oleh pihak sekolah'
  },
  {
    status: 'driver_ditugaskan_mengambil_ompreng',
    title: 'Driver Ditugaskan',
    description: 'Driver ditugaskan untuk mengambil ompreng'
  },
  {
    status: 'driver_menuju_sekolah',
    title: 'Driver Menuju Sekolah',
    description: 'Driver menuju sekolah untuk mengambil ompreng'
  },
  {
    status: 'driver_sampai_di_sekolah',
    title: 'Driver Sampai',
    description: 'Driver telah sampai di sekolah untuk pengambilan'
  },
  {
    status: 'ompreng_telah_diambil',
    title: 'Ompreng Diambil',
    description: 'Ompreng telah diambil dari sekolah'
  },
  {
    status: 'ompreng_sampai_di_sppg',
    title: 'Ompreng Sampai SPPG',
    description: 'Ompreng telah sampai di SPPG'
  },
  {
    status: 'ompreng_proses_pencucian',
    title: 'Proses Pencucian',
    description: 'Ompreng sedang dalam proses pencucian'
  },
  {
    status: 'ompreng_selesai_dicuci',
    title: 'Selesai Dicuci',
    description: 'Ompreng telah selesai dicuci'
  }
]

// Get the index of current status
const currentStatusIndex = computed(() => {
  return stages.findIndex(stage => stage.status === props.currentStatus)
})

// Check if a stage is completed
const isCompleted = (status) => {
  const stageIndex = stages.findIndex(stage => stage.status === status)
  return stageIndex < currentStatusIndex.value
}

// Check if a stage is in progress
const isInProgress = (status) => {
  return status === props.currentStatus
}

// Get stage color based on status
const getStageColor = (status) => {
  if (isCompleted(status)) {
    return 'green'
  } else if (isInProgress(status)) {
    return 'blue'
  } else {
    return 'gray'
  }
}

// Get timestamp for a stage from activity log
const getStageTimestamp = (status) => {
  const activity = props.activityLog.find(log => log.to_status === status)
  return activity?.transitioned_at
}

// Format timestamp to local timezone (Asia/Jakarta)
const formatTimestamp = (timestamp) => {
  if (!timestamp) return ''
  return dayjs(timestamp).tz('Asia/Jakarta').format('DD MMM YYYY, HH:mm') + ' WIB'
}
</script>

<style scoped>
.delivery-timeline {
  padding: 16px 0;
}

.timeline-content {
  padding-bottom: 16px;
}

.stage-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
  color: rgba(0, 0, 0, 0.85);
}

.stage-description {
  font-size: 14px;
  color: rgba(0, 0, 0, 0.65);
  margin-bottom: 4px;
}

.stage-timestamp {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
  font-style: italic;
}

:deep(.ant-timeline-item-tail) {
  border-left-width: 2px;
}

:deep(.ant-timeline-item-head) {
  background-color: transparent;
}
</style>
