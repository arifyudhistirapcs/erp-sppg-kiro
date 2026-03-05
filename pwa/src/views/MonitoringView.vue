<template>
  <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
    <div class="monitoring-page">
      <!-- NavBar -->
      <van-nav-bar title="Monitoring" />

      <!-- Date Filter -->
      <div class="monitoring-date-filter">
        <van-field
          v-model="selectedDateFormatted"
          placeholder="Pilih tanggal"
          readonly
          right-icon="calendar-o"
          @click="showDatePicker = true"
        />
      </div>

      <!-- Date Picker Popup -->
      <van-popup v-model:show="showDatePicker" position="bottom" :style="{ zIndex: 9999 }">
        <van-date-picker
          v-model="selectedDate"
          :min-date="new Date(2020, 0, 1)"
          :max-date="today"
          @confirm="onDateConfirm"
          @cancel="showDatePicker = false"
          title="Pilih Tanggal"
        >
          <template #confirm>
            <span style="color: #5A4372;">Konfirmasi</span>
          </template>
        </van-date-picker>
      </van-popup>

      <!-- Loading State (initial load) -->
      <template v-if="monitoringStore.loading && !monitoringStore.activities.length">
        <div class="activity-list">
          <SkeletonCard v-for="n in 4" :key="n" :rows="2" :avatar="true" />
        </div>
      </template>

      <!-- Error State -->
      <div v-else-if="monitoringStore.error && !monitoringStore.activities.length" class="error-state">
        <van-icon name="warning-o" size="48" color="var(--h-error)" />
        <p class="error-state__message">{{ monitoringStore.error }}</p>
        <van-button type="primary" size="normal" @click="monitoringStore.retry()">
          Coba Lagi
        </van-button>
      </div>

      <!-- Activity List with Infinite Scroll -->
      <template v-else>
        <div class="activity-list">
          <ActivityLogItem
            v-for="activity in monitoringStore.activities"
            :key="activity.id"
            :employeeName="activity.employeeName"
            :activityType="activity.activityType"
            :timestamp="activity.timestamp"
            :status="activity.status"
            :portions="activity.portions"
            :orderId="activity.id"
            :menuName="activity.menuName"
            @click="handleActivityClick"
          />
        </div>

        <!-- Empty State -->
        <div v-if="!monitoringStore.loading && !monitoringStore.activities.length" class="empty-state">
          <van-icon name="search" size="48" color="var(--h-text-light)" />
          <p class="empty-state__text">Tidak ada aktivitas ditemukan</p>
          <p class="empty-state__subtext">Menampilkan aktivitas tanggal {{ selectedDateFormatted }}</p>
        </div>
      </template>
    </div>
  </van-pull-refresh>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useMonitoringStore } from '@/stores/monitoring'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'
import ActivityLogItem from '@/components/mobile/ActivityLogItem.vue'
import SkeletonCard from '@/components/mobile/SkeletonCard.vue'

const monitoringStore = useMonitoringStore()
const router = useRouter()
const refreshing = ref(false)
const showDatePicker = ref(false)
const today = new Date()
const selectedDate = ref([today.getFullYear(), today.getMonth() + 1, today.getDate()])

// Format selected date for display
const selectedDateFormatted = computed(() => {
  const [year, month, day] = selectedDate.value
  const date = new Date(year, month - 1, day)
  return date.toLocaleDateString('id-ID', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

function handleActivityClick(orderId) {
  router.push({ name: 'monitoring-detail', params: { id: orderId } })
}

function onDateConfirm() {
  showDatePicker.value = false
  const [year, month, day] = selectedDate.value
  
  // Create date in local timezone (no UTC conversion)
  const date = new Date(year, month - 1, day, 0, 0, 0, 0)
  
  console.log('[MonitoringView] Selected date array:', selectedDate.value)
  console.log('[MonitoringView] Created date object:', date)
  console.log('[MonitoringView] Date ISO string:', date.toISOString())
  
  monitoringStore.setDate(date)
  showToast(`Menampilkan aktivitas tanggal ${day}/${month}/${year}`)
}

async function onRefresh() {
  await monitoringStore.fetchActivities()
  refreshing.value = false
}

onMounted(() => {
  monitoringStore.fetchActivities()
})
</script>

<style scoped>
.monitoring-page {
  padding: 0;
  padding-bottom: 80px;
  min-height: 100vh;
}

.monitoring-page > :not(.van-nav-bar) {
  padding-left: var(--h-spacing-lg);
  padding-right: var(--h-spacing-lg);
}

/* Date Filter */
.monitoring-date-filter {
  margin-top: var(--h-spacing-lg);
  margin-bottom: var(--h-spacing-lg);
  background: #FFFFFF;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0px 2px 8px rgba(0, 0, 0, 0.08);
}

/* Activity List */
.activity-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-md);
}

/* Error State */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px var(--h-spacing-xl);
  text-align: center;
}

.error-state__message {
  font-size: 14px;
  color: var(--h-text-secondary);
  margin: var(--h-spacing-lg) 0;
  line-height: 1.5;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px var(--h-spacing-xl);
  text-align: center;
}

.empty-state__text {
  font-size: 14px;
  color: var(--h-text-light);
  margin: var(--h-spacing-md) 0 0 0;
}

.empty-state__subtext {
  font-size: 12px;
  color: var(--h-text-secondary);
  margin: var(--h-spacing-xs) 0 0 0;
}

/* Date Picker Popup Fix */
:deep(.van-popup) {
  z-index: 9999 !important;
  background: rgba(0, 0, 0, 0.7) !important;
}

:deep(.van-popup--bottom) {
  background: #FFFFFF !important;
}

:deep(.van-date-picker) {
  background: #FFFFFF !important;
}

:deep(.van-picker__toolbar) {
  background: #FFFFFF !important;
  border-bottom: 1px solid #ebedf0;
}

:deep(.van-picker-column) {
  background: #FFFFFF !important;
}

:deep(.van-picker-column__item) {
  color: #323233 !important;
}

:deep(.van-overlay) {
  z-index: 9998 !important;
}
</style>
