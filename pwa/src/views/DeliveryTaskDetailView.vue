<template>
  <div class="delivery-task-detail-container">
    <!-- Navigation Bar -->
    <van-nav-bar 
      title="Detail Tugas Pengiriman" 
      left-arrow 
      fixed
      @click-left="goBack"
    >
      <template #right>
        <div class="nav-right-actions">
          <SyncStatusIndicator />
          <van-icon 
            name="refresh" 
            @click="refreshTask" 
            :class="{ 'rotating': isRefreshing }"
            class="refresh-icon"
          />
        </div>
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
      Memuat detail tugas...
    </van-loading>

    <!-- Error State -->
    <van-empty 
      v-else-if="!isLoading && !task"
      image="error"
      description="Tugas tidak ditemukan"
    >
      <van-button type="primary" @click="goBack">
        Kembali ke Daftar Tugas
      </van-button>
    </van-empty>

    <!-- Task Detail Content -->
    <div v-else-if="task" class="detail-content">
      <!-- Status Card -->
      <van-card class="status-card">
        <template #title>
          <div class="status-header">
            <span>Status Pengiriman</span>
            <van-tag 
              :type="getStatusType(task.status)" 
              size="large"
            >
              {{ getStatusText(task.status) }}
            </van-tag>
          </div>
        </template>
        
        <div class="route-info">
          <van-icon name="location" />
          <span>Urutan Rute: {{ task.route_order }}</span>
        </div>
      </van-card>

      <!-- School Information -->
      <van-cell-group title="Informasi Sekolah" class="info-group">
        <van-cell 
          title="Nama Sekolah" 
          :value="task.school?.name || 'Tidak tersedia'" 
          icon="shop-o"
        />
        <van-cell 
          title="Alamat" 
          :value="task.school?.address || 'Tidak tersedia'" 
          icon="location-o"
          is-link
          @click="showFullAddress"
        />
        <van-cell 
          title="Kontak Person" 
          :value="task.school?.contact_person || 'Tidak tersedia'" 
          icon="contact"
        />
        <van-cell 
          title="Nomor Telepon" 
          :value="task.school?.phone_number || 'Tidak tersedia'" 
          icon="phone-o"
          is-link
          @click="callSchool"
        />
        <van-cell 
          title="Jumlah Siswa" 
          :value="task.school?.student_count?.toString() || 'Tidak tersedia'" 
          icon="friends-o"
        />
      </van-cell-group>

      <!-- GPS Coordinates -->
      <van-cell-group title="Koordinat GPS" class="info-group">
        <van-cell 
          title="Latitude" 
          :value="task.school?.latitude?.toFixed(6) || 'Tidak tersedia'" 
          icon="aim"
        />
        <van-cell 
          title="Longitude" 
          :value="task.school?.longitude?.toFixed(6) || 'Tidak tersedia'" 
          icon="aim"
        />
        <van-cell 
          title="Akurasi GPS" 
          :value="getGPSAccuracy()"
          icon="location-o"
        />
      </van-cell-group>

      <!-- Delivery Information -->
      <van-cell-group title="Informasi Pengiriman" class="info-group">
        <van-cell 
          title="Jumlah Porsi" 
          :value="task.portions?.toString() || '0'" 
          icon="shopping-cart-o"
        />
        <van-cell 
          title="Tanggal Pengiriman" 
          :value="formatDate(task.task_date)" 
          icon="calendar-o"
        />
      </van-cell-group>

      <!-- Menu Items -->
      <van-cell-group 
        title="Menu yang Dikirim" 
        class="info-group"
        v-if="task.menu_items && task.menu_items.length > 0"
      >
        <van-cell 
          v-for="item in task.menu_items" 
          :key="item.id"
          :title="item.recipe?.name || 'Menu tidak diketahui'"
          :value="`${item.portions} porsi`"
          icon="shop-o"
        />
      </van-cell-group>

      <!-- Action Buttons -->
      <div class="action-buttons">
        <!-- GPS Navigation Button -->
        <van-button 
          type="primary" 
          size="large"
          block 
          @click="openGPSNavigation"
          icon="guide-o"
          :disabled="!hasValidGPS"
          class="nav-button"
        >
          <van-icon name="guide-o" />
          Buka Navigasi GPS
        </van-button>

        <!-- Status Update Buttons -->
        <div class="status-buttons">
          <van-button 
            v-if="task.status === 'pending'"
            type="success" 
            size="large"
            block 
            @click="startDelivery"
            icon="play-circle-o"
            :loading="isUpdatingStatus"
          >
            Mulai Pengiriman
          </van-button>

          <van-button 
            v-if="task.status === 'in_progress'"
            type="warning" 
            size="large"
            block 
            @click="openePODForm"
            icon="edit"
            :loading="isUpdatingStatus"
            class="epod-button"
          >
            Buat Bukti Pengiriman (e-POD)
          </van-button>

          <van-button 
            v-if="task.status === 'in_progress'"
            type="default" 
            size="large"
            block 
            @click="completeDelivery"
            icon="checked"
            :loading="isUpdatingStatus"
            class="complete-button"
          >
            Selesaikan Tanpa e-POD
          </van-button>

          <van-button 
            v-if="task.status === 'completed'"
            type="default" 
            size="large"
            block 
            disabled
            icon="success"
          >
            Pengiriman Selesai
          </van-button>
        </div>
      </div>
    </div>

    <!-- Full Address Dialog -->
    <van-dialog 
      v-model:show="showAddressDialog" 
      title="Alamat Lengkap"
      :message="task?.school?.address"
      show-cancel-button
      cancel-button-text="Tutup"
      confirm-button-text="Salin"
      @confirm="copyAddress"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import SyncStatusIndicator from '@/components/SyncStatusIndicator.vue'
import { showToast, showConfirmDialog, showSuccessToast } from 'vant'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const deliveryTasksStore = useDeliveryTasksStore()

// Reactive data
const isLoading = ref(false)
const isRefreshing = ref(false)
const isUpdatingStatus = ref(false)
const isOnline = ref(navigator.onLine)
const showAddressDialog = ref(false)
const task = ref(null)

// Computed properties
const hasValidGPS = computed(() => {
  return task.value?.school?.latitude && 
         task.value?.school?.longitude &&
         Math.abs(task.value.school.latitude) <= 90 &&
         Math.abs(task.value.school.longitude) <= 180
})

// Methods
const loadTask = async () => {
  const taskId = route.params.id
  if (!taskId) {
    showToast('ID tugas tidak valid')
    goBack()
    return
  }

  isLoading.value = true
  try {
    // First try to get from store
    let foundTask = deliveryTasksStore.getTaskById(parseInt(taskId))
    
    if (!foundTask) {
      // If not in store, fetch today's tasks first
      await deliveryTasksStore.fetchTodayTasks(authStore.user.id)
      foundTask = deliveryTasksStore.getTaskById(parseInt(taskId))
    }
    
    if (foundTask) {
      task.value = foundTask
    } else {
      showToast('Tugas tidak ditemukan')
      goBack()
    }
  } catch (error) {
    console.error('Error loading task:', error)
    showToast('Gagal memuat detail tugas')
  } finally {
    isLoading.value = false
  }
}

const refreshTask = async () => {
  if (!authStore.user?.id) return
  
  isRefreshing.value = true
  try {
    await deliveryTasksStore.fetchTodayTasks(authStore.user.id, true)
    const taskId = route.params.id
    const updatedTask = deliveryTasksStore.getTaskById(parseInt(taskId))
    
    if (updatedTask) {
      task.value = updatedTask
      showToast('Data berhasil diperbarui')
    }
  } catch (error) {
    console.error('Error refreshing task:', error)
    showToast('Gagal memperbarui data')
  } finally {
    isRefreshing.value = false
  }
}

const goBack = () => {
  router.push('/tasks')
}

const formatDate = (dateString) => {
  if (!dateString) return 'Tidak tersedia'
  
  try {
    const date = new Date(dateString)
    return date.toLocaleDateString('id-ID', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  } catch (error) {
    return 'Format tanggal tidak valid'
  }
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

const getGPSAccuracy = () => {
  if (!hasValidGPS.value) return 'GPS tidak tersedia'
  return 'Koordinat valid'
}

const showFullAddress = () => {
  if (task.value?.school?.address) {
    showAddressDialog.value = true
  } else {
    showToast('Alamat tidak tersedia')
  }
}

const copyAddress = async () => {
  if (task.value?.school?.address) {
    try {
      await navigator.clipboard.writeText(task.value.school.address)
      showSuccessToast('Alamat berhasil disalin')
    } catch (error) {
      showToast('Gagal menyalin alamat')
    }
  }
}

const callSchool = () => {
  const phoneNumber = task.value?.school?.phone_number
  if (phoneNumber) {
    window.location.href = `tel:${phoneNumber}`
  } else {
    showToast('Nomor telepon tidak tersedia')
  }
}

const openGPSNavigation = () => {
  if (!hasValidGPS.value) {
    showToast('Koordinat GPS tidak tersedia atau tidak valid')
    return
  }
  
  const { latitude, longitude } = task.value.school
  const schoolName = task.value.school.name || 'Sekolah Tujuan'
  
  // Create Google Maps navigation URL
  const mapsUrl = `https://www.google.com/maps/dir/?api=1&destination=${latitude},${longitude}&destination_place_id=${encodeURIComponent(schoolName)}`
  
  // Try to open in Google Maps app first, fallback to web
  const androidMapsUrl = `google.navigation:q=${latitude},${longitude}`
  const iosMapsUrl = `maps://maps.google.com/maps?daddr=${latitude},${longitude}&amp;ll=`
  
  // Detect platform and open appropriate navigation
  const userAgent = navigator.userAgent.toLowerCase()
  
  if (userAgent.includes('android')) {
    // Try Android Google Maps app first
    window.location.href = androidMapsUrl
    
    // Fallback to web after a short delay
    setTimeout(() => {
      window.open(mapsUrl, '_blank')
    }, 1000)
  } else if (userAgent.includes('iphone') || userAgent.includes('ipad')) {
    // Try iOS Maps app first
    window.location.href = iosMapsUrl
    
    // Fallback to web after a short delay
    setTimeout(() => {
      window.open(mapsUrl, '_blank')
    }, 1000)
  } else {
    // Desktop or other platforms - open web version
    window.open(mapsUrl, '_blank')
  }
  
  showToast('Membuka navigasi GPS...')
}

const startDelivery = async () => {
  try {
    const confirmed = await showConfirmDialog({
      title: 'Mulai Pengiriman',
      message: `Apakah Anda yakin ingin memulai pengiriman ke ${task.value.school?.name}?`,
      confirmButtonText: 'Ya, Mulai',
      cancelButtonText: 'Batal'
    })
    
    if (confirmed) {
      isUpdatingStatus.value = true
      await deliveryTasksStore.updateTaskStatus(task.value.id, 'in_progress')
      
      // Update local task status
      task.value.status = 'in_progress'
      
      showSuccessToast('Status pengiriman diperbarui')
    }
  } catch (error) {
    console.error('Error starting delivery:', error)
    showToast('Gagal memperbarui status')
  } finally {
    isUpdatingStatus.value = false
  }
}

const completeDelivery = async () => {
  try {
    const confirmed = await showConfirmDialog({
      title: 'Selesaikan Pengiriman',
      message: `Apakah Anda yakin pengiriman ke ${task.value.school?.name} sudah selesai tanpa e-POD?`,
      confirmButtonText: 'Ya, Selesai',
      cancelButtonText: 'Belum'
    })
    
    if (confirmed) {
      isUpdatingStatus.value = true
      await deliveryTasksStore.updateTaskStatus(task.value.id, 'completed')
      
      // Update local task status
      task.value.status = 'completed'
      
      showSuccessToast('Pengiriman berhasil diselesaikan')
    }
  } catch (error) {
    console.error('Error completing delivery:', error)
    showToast('Gagal menyelesaikan pengiriman')
  } finally {
    isUpdatingStatus.value = false
  }
}

const openePODForm = () => {
  router.push(`/tasks/${task.value.id}/epod`)
}

// Network status handlers
const handleOnline = () => {
  isOnline.value = true
  showToast('Koneksi internet tersambung')
  deliveryTasksStore.syncAllOfflineData()
}

const handleOffline = () => {
  isOnline.value = false
  showToast('Mode offline - Data akan disinkronkan saat online')
}

// Lifecycle
onMounted(() => {
  loadTask()
  
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
.delivery-task-detail-container {
  min-height: 100vh;
  background-color: #f7f8fa;
  padding-top: 46px; /* Nav bar height */
  padding-bottom: 16px;
}

.detail-content {
  padding: 16px;
}

.status-card {
  margin-bottom: 16px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.status-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.route-info {
  display: flex;
  align-items: center;
  margin-top: 8px;
  color: #646566;
  font-size: 14px;
}

.route-info .van-icon {
  margin-right: 8px;
  color: #1989fa;
}

.info-group {
  margin-bottom: 16px;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.action-buttons {
  margin-top: 24px;
  padding: 0 4px;
}

.nav-button {
  margin-bottom: 16px;
  height: 48px;
  font-size: 16px;
  font-weight: 600;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(25, 137, 250, 0.3);
}

.nav-button .van-icon {
  margin-right: 8px;
  font-size: 18px;
}

.status-buttons .van-button {
  height: 44px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 8px;
  margin-bottom: 8px;
}

.epod-button {
  background: linear-gradient(135deg, #ff976a, #ff6b35);
  border: none;
  color: white;
  box-shadow: 0 2px 8px rgba(255, 151, 106, 0.3);
}

.complete-button {
  margin-top: 8px;
}

.rotating {
  animation: rotate 1s linear infinite;
}

.nav-right-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refresh-icon {
  font-size: 18px;
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
  .detail-content {
    padding: 12px;
  }
  
  .status-header {
    font-size: 14px;
  }
  
  .nav-button {
    height: 44px;
    font-size: 15px;
  }
  
  .status-buttons .van-button {
    height: 40px;
    font-size: 14px;
  }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .delivery-task-detail-container {
    background-color: #1a1a1a;
  }
  
  .status-card,
  .info-group {
    background-color: #2a2a2a;
  }
}
</style>