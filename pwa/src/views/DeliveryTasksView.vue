<template>
  <div class="delivery-tasks-container">
    <!-- Navigation Bar -->
    <van-nav-bar title="Tugas Pengiriman Hari Ini" fixed>
      <template #right>
        <van-icon 
          name="refresh" 
          @click="refreshTasks" 
          :class="{ 'rotating': isRefreshing }"
        />
      </template>
    </van-nav-bar>

    <!-- Offline Indicator -->
    <van-notice-bar 
      v-if="!isOnline" 
      type="warning" 
      text="Mode offline - Data mungkin tidak terbaru"
      left-icon="warning-o"
    />

    <!-- Loading State -->
    <van-loading v-if="isLoading" type="spinner" vertical>
      Memuat tugas pengiriman...
    </van-loading>

    <!-- Empty State -->
    <van-empty 
      v-else-if="!isLoading && tasks.length === 0"
      image="search"
      description="Tidak ada tugas pengiriman hari ini"
    />

    <!-- Tasks List -->
    <div v-else class="tasks-list">
      <van-card
        v-for="task in sortedTasks"
        :key="task.id"
        :title="task.school?.name || 'Sekolah tidak diketahui'"
        :desc="formatAddress(task.school?.address)"
        class="task-card"
        @click="showTaskDetail(task)"
      >
        <template #tags>
          <van-tag 
            :type="getStatusType(task.status)" 
            size="medium"
          >
            {{ getStatusText(task.status) }}
          </van-tag>
          <van-tag 
            type="primary" 
            size="medium" 
            class="route-tag"
          >
            Urutan: {{ task.route_order }}
          </van-tag>
        </template>

        <template #footer>
          <div class="task-info">
            <div class="info-row">
              <van-icon name="location-o" />
              <span class="info-text">
                {{ task.school?.latitude?.toFixed(6) }}, {{ task.school?.longitude?.toFixed(6) }}
              </span>
              <van-button 
                size="mini" 
                type="primary" 
                @click.stop="openMaps(task.school)"
                icon="guide-o"
              >
                Navigasi
              </van-button>
            </div>
            
            <div class="info-row">
              <van-icon name="friends-o" />
              <span class="info-text">{{ task.portions }} porsi</span>
            </div>

            <div class="info-row" v-if="task.menu_items && task.menu_items.length > 0">
              <van-icon name="shop-o" />
              <span class="info-text">
                {{ task.menu_items.map(item => item.recipe?.name).join(', ') }}
              </span>
            </div>
          </div>
        </template>
      </van-card>
    </div>

    <!-- Bottom Navigation -->
    <van-tabbar v-model="active" route fixed>
      <van-tabbar-item to="/tasks" icon="orders-o">Tugas</van-tabbar-item>
      <van-tabbar-item to="/attendance" icon="clock-o">Absensi</van-tabbar-item>
      <van-tabbar-item to="/profile" icon="user-o">Profil</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import { showToast, showConfirmDialog } from 'vant'

const router = useRouter()

const authStore = useAuthStore()
const deliveryTasksStore = useDeliveryTasksStore()

// Reactive data
const active = ref(0)
const isLoading = ref(false)
const isRefreshing = ref(false)
const isOnline = ref(navigator.onLine)

// Computed properties
const tasks = computed(() => deliveryTasksStore.tasks)
const sortedTasks = computed(() => {
  return [...tasks.value].sort((a, b) => a.route_order - b.route_order)
})

// Methods
const loadTasks = async () => {
  if (!authStore.user?.id) return
  
  isLoading.value = true
  try {
    await deliveryTasksStore.fetchTodayTasks(authStore.user.id)
  } catch (error) {
    console.error('Error loading tasks:', error)
    showToast('Gagal memuat tugas pengiriman')
  } finally {
    isLoading.value = false
  }
}

const refreshTasks = async () => {
  if (!authStore.user?.id) return
  
  isRefreshing.value = true
  try {
    await deliveryTasksStore.fetchTodayTasks(authStore.user.id, true) // force refresh
    showToast('Data berhasil diperbarui')
  } catch (error) {
    console.error('Error refreshing tasks:', error)
    showToast('Gagal memperbarui data')
  } finally {
    isRefreshing.value = false
  }
}

const showTaskDetail = (task) => {
  router.push(`/tasks/${task.id}`)
}

const formatAddress = (address) => {
  if (!address) return 'Alamat tidak tersedia'
  return address.length > 50 ? address.substring(0, 50) + '...' : address
}

const getStatusType = (status) => {
  const statusTypes = {
    'pending': 'warning',
    'in_progress': 'primary',
    'completed': 'success',
    'cancelled': 'danger'
  }
  return statusTypes[status] || 'default'
}

const getStatusText = (status) => {
  const statusTexts = {
    'pending': 'Menunggu',
    'in_progress': 'Dalam Perjalanan',
    'completed': 'Selesai',
    'cancelled': 'Dibatalkan'
  }
  return statusTexts[status] || status
}

const openMaps = (school) => {
  if (!school?.latitude || !school?.longitude) {
    showToast('Koordinat GPS tidak tersedia')
    return
  }
  
  const url = `https://www.google.com/maps/dir/?api=1&destination=${school.latitude},${school.longitude}`
  window.open(url, '_blank')
}

// Network status handlers
const handleOnline = () => {
  isOnline.value = true
  showToast('Koneksi internet tersambung')
  // Sync offline data when back online
  deliveryTasksStore.syncAllOfflineData()
}

const handleOffline = () => {
  isOnline.value = false
  showToast('Mode offline - Data akan disinkronkan saat online')
}

// Lifecycle
onMounted(() => {
  loadTasks()
  
  // Listen for network status changes
  window.addEventListener('online', handleOnline)
  window.addEventListener('offline', handleOffline)
})

onUnmounted(() => {
  window.removeEventListener('online', handleOnline)
  window.removeEventListener('offline', handleOffline)
})
</script>

<style scoped>
.delivery-tasks-container {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-top: 46px; /* Nav bar height */
  padding-bottom: 50px; /* Tab bar height */
}

.tasks-list {
  padding: 8px;
}

.task-card {
  margin-bottom: 12px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.route-tag {
  margin-left: 8px;
}

.task-info {
  padding: 8px 0;
}

.info-row {
  display: flex;
  align-items: center;
  margin-bottom: 4px;
  font-size: 14px;
  color: #646566;
}

.info-row .van-icon {
  margin-right: 8px;
  color: #969799;
}

.info-text {
  flex: 1;
  margin-right: 8px;
}

.rotating {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Responsive adjustments */
@media (max-width: 375px) {
  .info-row {
    font-size: 12px;
  }
  
  .task-card {
    margin-bottom: 8px;
  }
}
</style>