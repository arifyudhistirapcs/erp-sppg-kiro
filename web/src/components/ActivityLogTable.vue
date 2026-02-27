<template>
  <div class="activity-log-table">
    <a-table
      :columns="columns"
      :data-source="sortedActivityLog"
      :pagination="false"
      :row-key="record => record.id"
    >
      <template #bodyCell="{ column, record, index }">
        <template v-if="column.key === 'timestamp'">
          {{ formatTimestamp(record.transitioned_at) }}
        </template>
        <template v-else-if="column.key === 'from_status'">
          {{ getStatusText(record.from_status) }}
        </template>
        <template v-else-if="column.key === 'to_status'">
          <a-tag :color="getStatusColor(record.to_status)">
            {{ getStatusText(record.to_status) }}
          </a-tag>
        </template>
        <template v-else-if="column.key === 'user'">
          <div>
            <div style="font-weight: 500">{{ record.user?.full_name || '-' }}</div>
            <div style="font-size: 12px; color: rgba(0, 0, 0, 0.45)">
              {{ getRoleText(record.user?.role) }}
            </div>
          </div>
        </template>
        <template v-else-if="column.key === 'elapsed_time'">
          <span v-if="index > 0" style="color: rgba(0, 0, 0, 0.65)">
            {{ calculateElapsedTime(sortedActivityLog[index - 1].transitioned_at, record.transitioned_at) }}
          </span>
          <span v-else style="color: rgba(0, 0, 0, 0.45)">-</span>
        </template>
        <template v-else-if="column.key === 'notes'">
          {{ record.notes || '-' }}
        </template>
      </template>
    </a-table>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import dayjs from 'dayjs'
import 'dayjs/locale/id'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'
import duration from 'dayjs/plugin/duration'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(utc)
dayjs.extend(timezone)
dayjs.extend(duration)
dayjs.extend(relativeTime)
dayjs.locale('id')

const props = defineProps({
  activityLog: {
    type: Array,
    required: true,
    default: () => []
  }
})

// Table columns
const columns = [
  {
    title: 'Waktu',
    key: 'timestamp',
    dataIndex: 'transitioned_at',
    width: 180
  },
  {
    title: 'Status Awal',
    key: 'from_status',
    dataIndex: 'from_status',
    width: 150
  },
  {
    title: 'Status Baru',
    key: 'to_status',
    dataIndex: 'to_status',
    width: 180
  },
  {
    title: 'Pengguna',
    key: 'user',
    dataIndex: 'user',
    width: 150
  },
  {
    title: 'Durasi',
    key: 'elapsed_time',
    width: 120
  },
  {
    title: 'Catatan',
    key: 'notes',
    dataIndex: 'notes'
  }
]

// Sort activity log chronologically (oldest first)
const sortedActivityLog = computed(() => {
  return [...props.activityLog].sort((a, b) => {
    return dayjs(a.transitioned_at).valueOf() - dayjs(b.transitioned_at).valueOf()
  })
})

// Format timestamp to local timezone (Asia/Jakarta)
const formatTimestamp = (timestamp) => {
  if (!timestamp) return '-'
  return dayjs(timestamp).tz('Asia/Jakarta').format('DD MMM YYYY, HH:mm') + ' WIB'
}

// Get status text in Indonesian
const getStatusText = (status) => {
  if (!status) return '-'
  
  const statusTexts = {
    'sedang_dimasak': 'Sedang Dimasak',
    'selesai_dimasak': 'Selesai Dimasak',
    'siap_dipacking': 'Siap Dipacking',
    'selesai_dipacking': 'Selesai Dipacking',
    'siap_dikirim': 'Siap Dikirim',
    'diperjalanan': 'Diperjalanan',
    'sudah_sampai_sekolah': 'Sudah Sampai Sekolah',
    'sudah_diterima_pihak_sekolah': 'Sudah Diterima',
    'driver_ditugaskan_mengambil_ompreng': 'Driver Ditugaskan',
    'driver_menuju_sekolah': 'Driver Menuju Sekolah',
    'driver_sampai_di_sekolah': 'Driver Sampai',
    'ompreng_telah_diambil': 'Ompreng Diambil',
    'ompreng_sampai_di_sppg': 'Ompreng Sampai SPPG',
    'ompreng_proses_pencucian': 'Proses Pencucian',
    'ompreng_selesai_dicuci': 'Selesai Dicuci'
  }
  
  return statusTexts[status] || status
}

// Get status color
const getStatusColor = (status) => {
  const completedStatuses = [
    'selesai_dimasak',
    'selesai_dipacking',
    'sudah_sampai_sekolah',
    'sudah_diterima_pihak_sekolah',
    'driver_sampai_di_sekolah',
    'ompreng_telah_diambil',
    'ompreng_sampai_di_sppg',
    'ompreng_selesai_dicuci'
  ]
  
  if (completedStatuses.includes(status)) {
    return 'success'
  }
  
  return 'processing'
}

// Get role text in Indonesian
const getRoleText = (role) => {
  if (!role) return '-'
  
  const roleTexts = {
    'kepala_sppg': 'Kepala SPPG',
    'kepala_yayasan': 'Kepala Yayasan',
    'akuntan': 'Akuntan',
    'ahli_gizi': 'Ahli Gizi',
    'pengadaan': 'Pengadaan',
    'chef': 'Chef',
    'packing': 'Packing',
    'driver': 'Driver',
    'asisten_lapangan': 'Asisten Lapangan',
    'kebersihan': 'Kebersihan'
  }
  
  return roleTexts[role] || role
}

// Calculate elapsed time between two timestamps
const calculateElapsedTime = (fromTime, toTime) => {
  if (!fromTime || !toTime) return '-'
  
  const from = dayjs(fromTime)
  const to = dayjs(toTime)
  const diff = to.diff(from)
  
  const duration = dayjs.duration(diff)
  
  const days = Math.floor(duration.asDays())
  const hours = duration.hours()
  const minutes = duration.minutes()
  
  const parts = []
  
  if (days > 0) {
    parts.push(`${days}h`)
  }
  if (hours > 0) {
    parts.push(`${hours}j`)
  }
  if (minutes > 0 || parts.length === 0) {
    parts.push(`${minutes}m`)
  }
  
  return parts.join(' ')
}
</script>

<style scoped>
.activity-log-table {
  width: 100%;
}

:deep(.ant-table) {
  font-size: 14px;
}

:deep(.ant-table-thead > tr > th) {
  background-color: #fafafa;
  font-weight: 600;
}

:deep(.ant-table-tbody > tr > td) {
  padding: 12px 16px;
}
</style>
