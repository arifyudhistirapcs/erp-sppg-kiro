<template>
  <div class="activity-log-item" @click="handleClick">
    <div class="activity-log-item__left">
      <div :class="['activity-log-item__type-icon', `activity-log-item__type-icon--${activityType}`]">
        <van-icon :name="typeIcon" size="20" color="#ffffff" />
      </div>
    </div>
    <div class="activity-log-item__content">
      <h4 class="activity-log-item__name">{{ employeeName }}</h4>
      <p v-if="menuName" class="activity-log-item__menu">{{ menuName }}</p>
      <span :class="['activity-log-item__status', statusClass]">
        {{ formattedStatus }}
      </span>
    </div>
    <div class="activity-log-item__right">
      <span class="activity-log-item__timestamp">{{ timestamp }}</span>
      <span class="activity-log-item__portions">{{ portions }} porsi</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  employeeName: {
    type: String,
    required: true
  },
  activityType: {
    type: String,
    required: true,
    validator: (val) => ['attendance', 'delivery', 'pickup'].includes(val)
  },
  timestamp: {
    type: String,
    required: true
  },
  status: {
    type: String,
    required: true
  },
  portions: {
    type: Number,
    default: 0
  },
  orderId: {
    type: [String, Number],
    default: null
  },
  menuName: {
    type: String,
    default: null
  }
})

const emit = defineEmits(['click'])

function handleClick() {
  if (props.orderId) {
    emit('click', props.orderId)
  }
}

const typeLabel = computed(() => {
  const labels = {
    attendance: 'Absensi',
    delivery: 'Pengiriman',
    pickup: 'Pengambilan'
  }
  return labels[props.activityType] || props.activityType
})

const typeIcon = computed(() => {
  const icons = {
    attendance: 'clock-o',
    delivery: 'logistics',
    pickup: 'revoke'
  }
  return icons[props.activityType] || 'info-o'
})

const formattedStatus = computed(() => {
  if (!props.status) return 'Menunggu'
  
  // Convert snake_case to Title Case and make it more readable
  const statusMap = {
    'ompreng_proses_pencucian': 'Proses Pencucian',
    'ompreng_selesai_dicuci': 'Selesai Dicuci',
    'sudah_sampai_sekolah': 'Sudah Sampai Sekolah',
    'dalam_perjalanan': 'Dalam Perjalanan',
    'siap_dikirim': 'Siap Dikirim',
    'sedang_dikemas': 'Sedang Dikemas',
    'sedang_dimasak': 'Sedang Dimasak',
    'menunggu_produksi': 'Menunggu Produksi',
    'pending': 'Menunggu',
    'completed': 'Selesai',
    'in_progress': 'Dalam Proses'
  }
  
  const lowerStatus = String(props.status).toLowerCase()
  
  // Check if exact match exists
  if (statusMap[lowerStatus]) {
    return statusMap[lowerStatus]
  }
  
  // Fallback: convert snake_case to Title Case
  return String(props.status)
    .replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
})

const statusClass = computed(() => {
  if (!props.status) return 'activity-log-item__status--default'
  
  const s = String(props.status).toLowerCase()
  
  // Success states
  if (s.includes('selesai') || s.includes('completed') || s.includes('sampai')) {
    return 'activity-log-item__status--success'
  }
  
  // In progress states
  if (s.includes('proses') || s.includes('perjalanan') || s.includes('progress') || 
      s.includes('dikemas') || s.includes('dimasak')) {
    return 'activity-log-item__status--primary'
  }
  
  // Warning states
  if (s.includes('menunggu') || s.includes('pending') || s.includes('siap')) {
    return 'activity-log-item__status--warning'
  }
  
  return 'activity-log-item__status--default'
})
</script>

<style scoped>
.activity-log-item {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-lg);
  display: flex;
  align-items: center;
  gap: var(--h-spacing-md);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.activity-log-item:active {
  transform: scale(0.98);
  box-shadow: 0px 2px 4px rgba(0, 0, 0, 0.08);
}

.activity-log-item__left {
  flex-shrink: 0;
}

.activity-log-item__type-icon {
  width: 44px;
  height: 44px;
  border-radius: var(--h-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}

.activity-log-item__type-icon--attendance {
  background: var(--h-primary);
}

.activity-log-item__type-icon--delivery {
  background: linear-gradient(135deg, #05CD99 0%, #04A777 100%);
}

.activity-log-item__type-icon--pickup {
  background: var(--h-warning);
}

.activity-log-item__content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.activity-log-item__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--h-text-primary);
  margin: 0 0 4px 0;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.activity-log-item__menu {
  font-size: 13px;
  font-weight: 400;
  color: var(--h-text-secondary);
  margin: 0 0 6px 0;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.activity-log-item__status {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 10px;
  border-radius: var(--h-radius-sm);
  display: inline-block;
  width: fit-content;
}

.activity-log-item__status--success {
  color: #05CD99;
  background: rgba(5, 205, 153, 0.12);
}

.activity-log-item__status--warning {
  color: #FFB547;
  background: rgba(255, 181, 71, 0.12);
}

.activity-log-item__status--error {
  color: #EE5D50;
  background: rgba(238, 93, 80, 0.12);
}

.activity-log-item__status--primary {
  color: #5A4372;
  background: rgba(90, 67, 114, 0.12);
}

.activity-log-item__status--default {
  color: var(--h-text-secondary);
  background: var(--h-bg-light);
}

.activity-log-item__right {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.activity-log-item__timestamp {
  font-size: 12px;
  font-weight: 500;
  color: var(--h-text-secondary);
}

.activity-log-item__portions {
  font-size: 20px;
  font-weight: 700;
  color: var(--h-text-primary);
  line-height: 1;
}
</style>
