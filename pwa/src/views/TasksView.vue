<template>
  <div class="tasks-view">
    <!-- Nav Bar -->
    <van-nav-bar :title="`Tugas (${allTasks.length})`" fixed>
      <template #right>
        <van-icon
          name="replay"
          size="20"
          color="#ffffff"
          :class="{ 'tasks-view__refresh--spinning': isRefreshing }"
          @click="refreshTasks"
        />
      </template>
    </van-nav-bar>

    <van-pull-refresh v-model="isPullRefreshing" @refresh="onPullRefresh">
      <div class="tasks-view__content">
        <!-- Offline Notice -->
        <van-notice-bar
          v-if="!isOnline"
          type="warning"
          text="Mode offline - Data mungkin tidak terbaru"
          left-icon="warning-o"
        />

        <!-- Tab Selection -->
        <div class="tasks-view__tabs">
          <button
            :class="['tasks-view__tab', { 'tasks-view__tab--active': activeTab === 'delivery' }]"
            @click="activeTab = 'delivery'"
          >
            Pengiriman
          </button>
          <button
            :class="['tasks-view__tab', { 'tasks-view__tab--active': activeTab === 'pickup' }]"
            @click="activeTab = 'pickup'"
          >
            Pengambilan
          </button>
        </div>

        <!-- Date Filter -->
        <div class="tasks-view__date-filter">
          <van-field
            v-model="selectedDateFormatted"
            placeholder="Pilih tanggal"
            readonly
            right-icon="calendar-o"
            @click="handleDatePickerClick"
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

        <!-- Summary Cards -->
        <div class="tasks-view__summary">
          <SummaryCard
            icon="clock-o"
            :icon-color="'var(--h-warning)'"
            :label="activeTab === 'delivery' ? 'Siap Dikirim' : 'Belum Dimulai'"
            :value="summaryCountsByTab.pending"
            :loading="isLoading"
          />
          <SummaryCard
            icon="logistics"
            :icon-color="'var(--h-primary)'"
            :label="activeTab === 'delivery' ? 'Dalam Perjalanan' : 'Dalam Proses'"
            :value="summaryCountsByTab.inProgress"
            :loading="isLoading"
          />
          <SummaryCard
            icon="passed"
            :icon-color="'var(--h-success)'"
            :label="activeTab === 'delivery' ? 'Sudah Diterima' : 'Tiba di SPPG'"
            :value="summaryCountsByTab.completed"
            :loading="isLoading"
          />
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="tasks-view__skeleton-list">
          <SkeletonCard v-for="n in 3" :key="n" :rows="3" />
        </div>

        <!-- Empty State -->
        <van-empty
          v-else-if="filteredTasks.length === 0"
          image="search"
          :description="`Tidak ada tugas ${activeTab === 'delivery' ? 'pengiriman' : 'pengambilan'} untuk tanggal yang dipilih`"
        />

        <!-- Task List -->
        <div v-else class="tasks-view__list">
          <h3 class="tasks-view__section-title">Daftar Tugas</h3>
          <TaskCard
            v-for="task in sortedFilteredTasks"
            :key="task.id"
            :school-name="task.school?.name || 'Sekolah tidak diketahui'"
            :address="formatAddress(task.school?.address)"
            :task-type="task.task_type === 'pickup' ? 'pickup' : 'delivery'"
            :status="mapStatus(task)"
            :route-order="task.route_order || 0"
            @click="showTaskDetail(task)"
          />
        </div>

        <!-- Performance Summary -->
        <div v-if="!isLoading && filteredTasks.length > 0" class="tasks-view__performance">
          <h3 class="tasks-view__section-title">Ringkasan Performa</h3>
          <div class="tasks-view__performance-card">
            <div class="tasks-view__perf-row">
              <span class="tasks-view__perf-label">Total tugas {{ activeTab === 'delivery' ? 'pengiriman' : 'pengambilan' }}</span>
              <span class="tasks-view__perf-value">{{ performanceSummary.total }}</span>
            </div>
            <div class="tasks-view__perf-row">
              <span class="tasks-view__perf-label">Selesai</span>
              <span class="tasks-view__perf-value tasks-view__perf-value--success">{{ performanceSummary.completed }}</span>
            </div>
            <div class="tasks-view__perf-row">
              <span class="tasks-view__perf-label">Pending</span>
              <span class="tasks-view__perf-value tasks-view__perf-value--warning">{{ performanceSummary.pending }}</span>
            </div>
            <div class="tasks-view__perf-row">
              <span class="tasks-view__perf-label">Persentase penyelesaian</span>
              <span class="tasks-view__perf-value tasks-view__perf-value--primary">{{ performanceSummary.percentage }}%</span>
            </div>
          </div>
        </div>

        <!-- History Section - Hidden for now as offline data not available -->
        <!-- 
        <div v-if="!isLoading" class="tasks-view__history">
          <h3 class="tasks-view__section-title">Riwayat Tugas</h3>
          <van-field
            v-model="historyDateFormatted"
            placeholder="Pilih tanggal"
            readonly
            right-icon="calendar-o"
            @click="showHistoryDatePicker = true"
          />
          <div v-if="isHistoryDateToday" class="tasks-view__history-note">
            Menampilkan tugas hari ini
          </div>
          <div v-else class="tasks-view__history-note">
            Riwayat tugas untuk tanggal yang dipilih belum tersedia secara offline.
          </div>
        </div>

        <van-popup v-model:show="showHistoryDatePicker" position="bottom" :style="{ zIndex: 9999 }">
          <van-date-picker
            v-model="historyDate"
            :min-date="new Date(2020, 0, 1)"
            :max-date="today"
            @confirm="onHistoryDateConfirm"
            @cancel="showHistoryDatePicker = false"
            title="Pilih Tanggal Riwayat"
          >
            <template #confirm>
              <span style="color: #5A4372;">Konfirmasi</span>
            </template>
          </van-date-picker>
        </van-popup>
        -->
      </div>
    </van-pull-refresh>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import { showToast } from 'vant'
import TaskCard from '@/components/mobile/TaskCard.vue'
import SummaryCard from '@/components/mobile/SummaryCard.vue'
import SkeletonCard from '@/components/mobile/SkeletonCard.vue'

const router = useRouter()
const authStore = useAuthStore()
const deliveryTasksStore = useDeliveryTasksStore()

// Reactive state
const activeTab = ref('delivery')
const isLoading = ref(false)
const isRefreshing = ref(false)
const isPullRefreshing = ref(false)
const isOnline = ref(navigator.onLine)
const today = new Date()
const selectedDate = ref([today.getFullYear(), today.getMonth() + 1, today.getDate()])
const showDatePicker = ref(false)
const historyDate = ref([today.getFullYear(), today.getMonth() + 1, today.getDate()])
const showHistoryDatePicker = ref(false)

// Computed: all tasks from store
const allTasks = computed(() => deliveryTasksStore.tasks)

// Computed: tasks filtered by active tab AND selected date
const filteredTasks = computed(() => {
  const [year, month, day] = selectedDate.value
  const selectedDateStr = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
  
  console.log('[TasksView] Filtering tasks:', {
    activeTab: activeTab.value,
    selectedDateStr,
    totalTasks: allTasks.value.length,
    taskDates: allTasks.value.map(t => ({ id: t.id, type: t.task_type, task_date: t.task_date }))
  })
  
  return allTasks.value.filter(task => {
    // Filter by task type (delivery or pickup)
    if (task.task_type !== activeTab.value) return false
    
    // Filter by selected date
    if (task.task_date) {
      // Handle both ISO format (2026-03-05T10:00:00+07:00) and simple format (2026-03-05)
      let taskDateStr = task.task_date
      if (taskDateStr.includes('T')) {
        taskDateStr = taskDateStr.split('T')[0]
      } else if (taskDateStr.includes(' ')) {
        taskDateStr = taskDateStr.split(' ')[0]
      }
      console.log('[TasksView] Comparing dates:', { taskDateStr, selectedDateStr, match: taskDateStr === selectedDateStr })
      return taskDateStr === selectedDateStr
    }
    
    return false
  })
})

// Computed: sorted filtered tasks by route_order
const sortedFilteredTasks = computed(() => {
  return [...filteredTasks.value].sort((a, b) => (a.route_order || 0) - (b.route_order || 0))
})

// Computed: summary counts for the active tab
const summaryCountsByTab = computed(() => {
  const tasks = filteredTasks.value
  const pending = tasks.filter(t => mapStatus(t) === 'pending').length
  // Include 'arrived' in inProgress count for summary purposes
  const inProgress = tasks.filter(t => ['in_progress', 'arrived'].includes(mapStatus(t))).length
  const completed = tasks.filter(t => mapStatus(t) === 'completed').length
  return { pending, inProgress, completed }
})

// Computed: overall performance summary (filtered by date and tab)
const performanceSummary = computed(() => {
  // Use filteredTasks to respect both date and tab filters
  const total = filteredTasks.value.length
  const completed = filteredTasks.value.filter(t => mapStatus(t) === 'completed').length
  const pending = total - completed
  const percentage = total > 0 ? Math.round((completed / total) * 100) : 0
  return { completed, pending, percentage, total }
})

// Computed: check if history date is today
const isHistoryDateToday = computed(() => {
  const [year, month, day] = historyDate.value
  return year === today.getFullYear() &&
    month === today.getMonth() + 1 &&
    day === today.getDate()
})

// Computed: check if selected date is today
const isSelectedDateToday = computed(() => {
  const [year, month, day] = selectedDate.value
  return year === today.getFullYear() &&
    month === today.getMonth() + 1 &&
    day === today.getDate()
})

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

// Format history date for display
const historyDateFormatted = computed(() => {
  const [year, month, day] = historyDate.value
  const date = new Date(year, month - 1, day)
  return date.toLocaleDateString('id-ID', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
})

// Map task status to simplified status for TaskCard
function mapStatus(task) {
  if (task.task_type === 'pickup') {
    // Pickup task stages:
    // Stage 10: Driver Menuju Lokasi Pengambilan
    // Stage 11: Driver Tiba di Lokasi
    // Stage 12: Driver Kembali ke SPPG
    // Stage 13: Driver Tiba di SPPG (completed)
    const stage = task.current_stage
    console.log('[TasksView] mapStatus pickup task:', { id: task.id, current_stage: stage, status: task.status })
    if (stage >= 13) return 'completed' // Driver Tiba di SPPG
    if (stage === 12) return 'in_progress' // Driver Kembali ke SPPG
    if (stage === 11) return 'arrived' // Driver Tiba di Lokasi
    if (stage === 10) return 'in_progress' // Driver Menuju Lokasi
    return 'pending'
  }
  
  // Delivery task statuses:
  // Stage 1: Siap Dikirim (pending)
  // Stage 2: Diperjalanan (in_progress)
  // Stage 3: Sudah Sampai Sekolah (arrived)
  // Stage 4: Sudah Diterima (completed)
  const status = task.status
  const stage = task.current_stage
  
  // Check by stage first (more accurate)
  if (stage) {
    if (stage >= 4) return 'completed' // Sudah Diterima
    if (stage === 3) return 'arrived' // Sudah Sampai Sekolah
    if (stage === 2) return 'in_progress' // Diperjalanan
    return 'pending' // Siap Dikirim
  }
  
  // Fallback to status field
  if (['received', 'completed'].includes(status)) return 'completed'
  if (status === 'arrived') return 'arrived'
  if (status === 'in_progress') return 'in_progress'
  return 'pending'
}

function formatAddress(address) {
  if (!address) return 'Alamat tidak tersedia'
  return address.length > 60 ? address.substring(0, 60) + '...' : address
}

function showTaskDetail(task) {
  if (task.task_type === 'pickup') {
    router.push(`/pickup-tasks/${task.pickup_task_id}?record=${task.delivery_record_id}`)
  } else {
    router.push(`/tasks/${task.id}`)
  }
}

function onDateConfirm() {
  showDatePicker.value = false
  console.log('Date selected:', selectedDate.value)
  // Filter will automatically update via computed property
  if (!isSelectedDateToday.value) {
    const [year, month, day] = selectedDate.value
    showToast(`Menampilkan tugas tanggal ${day}/${month}/${year}`)
  } else {
    showToast('Menampilkan tugas hari ini')
  }
}

function handleDatePickerClick() {
  console.log('Date picker clicked, current value:', showDatePicker.value)
  showDatePicker.value = true
  console.log('Date picker after set:', showDatePicker.value)
}

function onHistoryDateConfirm() {
  showHistoryDatePicker.value = false
  if (!isHistoryDateToday.value) {
    showToast('Riwayat untuk tanggal yang dipilih belum tersedia')
  }
}

async function loadTasks() {
  if (!authStore.user?.id) return
  isLoading.value = true
  try {
    await deliveryTasksStore.fetchTodayTasks(authStore.user.id)
  } catch (error) {
    console.error('Error loading tasks:', error)
    showToast('Gagal memuat tugas')
  } finally {
    isLoading.value = false
  }
}

async function refreshTasks() {
  if (!authStore.user?.id) return
  isRefreshing.value = true
  try {
    await deliveryTasksStore.fetchTodayTasks(authStore.user.id, true)
    showToast('Data berhasil diperbarui')
  } catch (error) {
    console.error('Error refreshing tasks:', error)
    showToast('Gagal memperbarui data')
  } finally {
    isRefreshing.value = false
  }
}

async function onPullRefresh() {
  try {
    if (authStore.user?.id) {
      await deliveryTasksStore.fetchTodayTasks(authStore.user.id, true)
      showToast('Data berhasil diperbarui')
    }
  } catch (error) {
    console.error('Error on pull refresh:', error)
    showToast('Gagal memperbarui data')
  } finally {
    isPullRefreshing.value = false
  }
}

// Network status handlers
function handleOnline() {
  isOnline.value = true
  showToast('Koneksi internet tersambung')
  deliveryTasksStore.syncAllOfflineData()
}

function handleOffline() {
  isOnline.value = false
  showToast('Mode offline - Data akan disinkronkan saat online')
}

// Lifecycle
onMounted(() => {
  loadTasks()
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
})

onUnmounted(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped>
.tasks-view {
  min-height: 100vh;
  background: var(--h-bg-primary);
  padding-top: 56px;
}

.tasks-view__content {
  padding: 0;
  padding-bottom: 80px;
}

.tasks-view__content > :not(.van-notice-bar) {
  padding-left: var(--h-spacing-lg);
  padding-right: var(--h-spacing-lg);
}

/* Refresh icon spin */
.tasks-view__refresh--spinning {
  animation: tasks-spin 1s linear infinite;
}

@keyframes tasks-spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Tab Selection */
.tasks-view__tabs {
  display: flex;
  gap: 12px;
  padding: var(--h-spacing-md) 0;
  justify-content: center;
}

.tasks-view__tab {
  flex: 1;
  max-width: 200px;
  padding: 10px 20px;
  border-radius: 20px;
  border: none;
  font-family: var(--h-font-family);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #FFFFFF;
  color: #74788C;
  box-shadow: 0px 2px 8px rgba(0, 0, 0, 0.08);
  min-height: 44px;
}

.tasks-view__tab--active {
  background: #5A4372;
  color: #FFFFFF;
  box-shadow: 0px 4px 12px rgba(90, 67, 114, 0.3);
}

/* Date Filter */
.tasks-view__date-filter {
  margin-bottom: var(--h-spacing-lg);
  background: #FFFFFF;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0px 2px 8px rgba(0, 0, 0, 0.08);
}

/* Summary Cards */
.tasks-view__summary {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--h-spacing-sm);
  margin-bottom: var(--h-spacing-lg);
}

/* Skeleton list */
.tasks-view__skeleton-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-md);
}

/* Task List */
.tasks-view__list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-md);
  margin-bottom: var(--h-spacing-xl);
}

.tasks-view__section-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--h-text-primary);
  margin: 0 0 var(--h-spacing-md) 0;
}

/* Performance Summary */
.tasks-view__performance {
  margin-bottom: var(--h-spacing-xl);
}

.tasks-view__performance-card {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-lg);
}

.tasks-view__perf-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--h-spacing-sm) 0;
}

.tasks-view__perf-row:not(:last-child) {
  border-bottom: 1px solid var(--h-border-light);
}

.tasks-view__perf-label {
  font-size: 14px;
  color: var(--h-text-secondary);
}

.tasks-view__perf-value {
  font-size: 16px;
  font-weight: 700;
  color: var(--h-text-primary);
}

.tasks-view__perf-value--success {
  color: var(--h-success);
}

.tasks-view__perf-value--warning {
  color: var(--h-warning);
}

.tasks-view__perf-value--primary {
  color: var(--h-primary);
}

/* History Section */
.tasks-view__history {
  margin-bottom: var(--h-spacing-xl);
}

.tasks-view__history .van-field {
  background: #FFFFFF;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0px 2px 8px rgba(0, 0, 0, 0.08);
  margin-bottom: var(--h-spacing-md);
}

.tasks-view__history-note {
  margin-top: var(--h-spacing-md);
  font-size: 13px;
  color: var(--h-text-secondary);
  padding: var(--h-spacing-md);
  background: var(--h-bg-card);
  border-radius: var(--h-radius-md);
  text-align: center;
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

/* Responsive */
@media (max-width: 375px) {
  .tasks-view__summary {
    grid-template-columns: 1fr;
    gap: var(--h-spacing-sm);
  }

  .tasks-view__tab {
    padding: 6px 14px;
    font-size: 13px;
  }
}
</style>
