<template>
  <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
    <div class="monitoring-detail-page">
      <!-- NavBar -->
      <van-nav-bar
        title="Detail Aktivitas"
        left-arrow
        @click-left="onBack"
      />

      <!-- Loading State -->
      <template v-if="loading">
        <div class="detail-header">
          <van-skeleton title :row="2" />
        </div>
        <div class="timeline-section">
          <van-skeleton title :row="5" />
        </div>
      </template>

      <!-- Error State -->
      <div v-else-if="error" class="error-state">
        <van-icon name="warning-o" size="48" color="var(--h-error)" />
        <p class="error-state__message">{{ error }}</p>
        <van-button type="primary" size="normal" @click="fetchDetail">
          Coba Lagi
        </van-button>
      </div>

      <!-- Content -->
      <template v-else-if="orderDetail">
        <!-- Header Info -->
        <div class="detail-header h-card">
          <h2 class="school-name">{{ orderDetail.school?.name || 'Sekolah' }}</h2>
          
          <!-- Menu Info with Image -->
          <div v-if="orderDetail.menu?.name || orderDetail.menu_name" class="menu-info">
            <div v-if="orderDetail.menu?.photo_url || orderDetail.menu?.image_url" class="menu-image">
              <img :src="orderDetail.menu.photo_url || orderDetail.menu.image_url" :alt="orderDetail.menu.name" />
            </div>
            <div class="menu-details">
              <p class="menu-name">{{ orderDetail.menu?.name || orderDetail.menu_name }}</p>
              
              <!-- Nutrition Info -->
              <div v-if="hasNutritionInfo" class="nutrition-info">
                <div class="nutrition-grid">
                  <div v-if="orderDetail.menu?.calories || orderDetail.menu?.kalori" class="nutrition-item">
                    <span class="nutrition-label">Kalori</span>
                    <span class="nutrition-value">{{ formatNumber(orderDetail.menu.calories || orderDetail.menu.kalori) }} kkal</span>
                  </div>
                  <div v-if="orderDetail.menu?.protein" class="nutrition-item">
                    <span class="nutrition-label">Protein</span>
                    <span class="nutrition-value">{{ formatNumber(orderDetail.menu.protein) }} g</span>
                  </div>
                  <div v-if="orderDetail.menu?.carbs || orderDetail.menu?.carbohydrates || orderDetail.menu?.karbohidrat" class="nutrition-item">
                    <span class="nutrition-label">Karbohidrat</span>
                    <span class="nutrition-value">{{ formatNumber(orderDetail.menu.carbs || orderDetail.menu.carbohydrates || orderDetail.menu.karbohidrat) }} g</span>
                  </div>
                  <div v-if="orderDetail.menu?.fat || orderDetail.menu?.fats || orderDetail.menu?.lemak" class="nutrition-item">
                    <span class="nutrition-label">Lemak</span>
                    <span class="nutrition-value">{{ formatNumber(orderDetail.menu.fat || orderDetail.menu.fats || orderDetail.menu.lemak) }} g</span>
                  </div>
                </div>
                <p class="nutrition-note">Per porsi</p>
              </div>
            </div>
          </div>

          <div class="header-info">
            <div class="info-item">
              <span class="info-label">Porsi</span>
              <span class="info-value">{{ orderDetail.portions || 0 }} porsi</span>
            </div>
            <div v-if="orderDetail.small_portions || orderDetail.large_portions" class="info-item">
              <span class="info-label">Detail Porsi</span>
              <span class="info-value">
                <span v-if="orderDetail.small_portions">Kecil: {{ orderDetail.small_portions }}</span>
                <span v-if="orderDetail.small_portions && orderDetail.large_portions"> | </span>
                <span v-if="orderDetail.large_portions">Besar: {{ orderDetail.large_portions }}</span>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">Driver</span>
              <span class="info-value">{{ orderDetail.driver?.name || '-' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Status</span>
              <span :class="['info-badge', getStatusClass(orderDetail.current_status)]">
                {{ formatStatus(orderDetail.current_status) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Timeline -->
        <div class="timeline-section h-card">
          <h3 class="section-title">Timeline Aktivitas</h3>
          <div class="timeline">
            <div
              v-for="(activity, index) in timeline"
              :key="index"
              :class="['timeline-item', { 'timeline-item--active': activity.is_completed || activity.completed_at }]"
            >
              <div class="timeline-marker">
                <div :class="['timeline-dot', { 'timeline-dot--active': activity.is_completed || activity.completed_at }]">
                  <van-icon v-if="activity.is_completed || activity.completed_at" name="success" size="10" color="#ffffff" />
                </div>
                <div v-if="index < timeline.length - 1" class="timeline-line"></div>
              </div>
              <div class="timeline-content">
                <div class="timeline-header">
                  <div class="timeline-header-left">
                    <h4 class="timeline-title">{{ formatStageLabel(activity.title || activity.stage_label || activity.stage || activity.status) }}</h4>
                    <p class="timeline-description">{{ activity.description || activity.message || '-' }}</p>
                  </div>
                  <span v-if="activity.timestamp || activity.completed_at || activity.created_at" class="timeline-time">
                    {{ formatTimestamp(activity.timestamp || activity.completed_at || activity.created_at) }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </van-pull-refresh>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import { showToast } from 'vant'

const route = useRoute()
const router = useRouter()

const orderId = ref(route.params.id)
const orderDetail = ref(null)
const timeline = ref([])
const loading = ref(false)
const refreshing = ref(false)
const error = ref(null)

const hasNutritionInfo = computed(() => {
  if (!orderDetail.value?.menu) return false
  const menu = orderDetail.value.menu
  return menu.calories || menu.kalori || menu.protein || menu.carbs || menu.carbohydrates || menu.karbohidrat || menu.fat || menu.fats || menu.lemak
})

function onBack() {
  router.back()
}

async function fetchDetail() {
  loading.value = true
  error.value = null

  try {
    // Fetch order detail
    const detailRes = await api.get(`/activity-tracker/orders/${orderId.value}`)
    
    if (detailRes.data.success) {
      orderDetail.value = detailRes.data.data
      console.log('[MonitoringDetail] Order detail:', orderDetail.value)
      console.log('[MonitoringDetail] Menu data:', orderDetail.value.menu)
      console.log('[MonitoringDetail] Menu name:', orderDetail.value.menu?.name || orderDetail.value.menu_name)
      console.log('[MonitoringDetail] Menu image:', orderDetail.value.menu?.image_url)
      console.log('[MonitoringDetail] Has nutrition:', hasNutritionInfo.value)
    }

    // Fetch activity log (timeline)
    const activityRes = await api.get(`/activity-tracker/orders/${orderId.value}/activity`)
    
    if (activityRes.data.success) {
      timeline.value = activityRes.data.data || []
      console.log('[MonitoringDetail] Timeline data:', timeline.value)
      console.log('[MonitoringDetail] First activity:', timeline.value[0])
    }
  } catch (err) {
    console.error('Error fetching detail:', err)
    error.value = err.response?.data?.message || 'Gagal memuat detail aktivitas'
  } finally {
    loading.value = false
  }
}

async function onRefresh() {
  await fetchDetail()
  refreshing.value = false
}

function formatStatus(status) {
  if (!status) return 'Menunggu'
  
  const statusMap = {
    'ompreng_proses_pencucian': 'Proses Pencucian',
    'ompreng_selesai_dicuci': 'Selesai Dicuci',
    'sudah_sampai_sekolah': 'Sudah Sampai Sekolah',
    'dalam_perjalanan': 'Dalam Perjalanan',
    'siap_dikirim': 'Siap Dikirim',
    'sedang_dikemas': 'Sedang Dikemas',
    'sedang_dimasak': 'Sedang Dimasak',
    'menunggu_produksi': 'Menunggu Produksi'
  }
  
  const lowerStatus = String(status).toLowerCase()
  if (statusMap[lowerStatus]) {
    return statusMap[lowerStatus]
  }
  
  return String(status)
    .replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

function formatStageLabel(label) {
  if (!label) return 'Status Tidak Diketahui'
  
  // Map common stage labels to readable format
  const stageMap = {
    'menunggu_produksi': 'Menunggu Produksi',
    'sedang_dimasak': 'Sedang Dimasak',
    'selesai_dimasak': 'Selesai Dimasak',
    'sedang_dikemas': 'Sedang Dikemas',
    'selesai_dikemas': 'Selesai Dikemas',
    'siap_dikirim': 'Siap Dikirim',
    'dalam_perjalanan': 'Dalam Perjalanan',
    'sudah_sampai_sekolah': 'Sudah Sampai Sekolah',
    'ompreng_proses_pencucian': 'Proses Pencucian',
    'ompreng_selesai_dicuci': 'Selesai Dicuci',
    'pending': 'Menunggu',
    'in_progress': 'Dalam Proses',
    'completed': 'Selesai'
  }
  
  const lowerLabel = String(label).toLowerCase().trim()
  if (stageMap[lowerLabel]) {
    return stageMap[lowerLabel]
  }
  
  // Fallback: convert snake_case to Title Case
  return String(label)
    .replace(/_/g, ' ')
    .split(' ')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')
}

function getStatusClass(status) {
  if (!status) return 'status-default'
  
  const s = String(status).toLowerCase()
  
  if (s.includes('selesai') || s.includes('completed') || s.includes('sampai')) {
    return 'status-success'
  }
  
  if (s.includes('proses') || s.includes('perjalanan') || s.includes('dikemas') || s.includes('dimasak')) {
    return 'status-primary'
  }
  
  if (s.includes('menunggu') || s.includes('pending') || s.includes('siap')) {
    return 'status-warning'
  }
  
  return 'status-default'
}

function formatTimestamp(timestamp) {
  if (!timestamp) return ''
  
  try {
    const date = new Date(timestamp)
    
    // Check if date is valid
    if (isNaN(date.getTime())) {
      console.warn('[MonitoringDetail] Invalid timestamp:', timestamp)
      return ''
    }
    
    // Format: "14:30"
    const time = date.toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit'
    })
    
    // Format: "5 Mar"
    const dateStr = date.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short'
    })
    
    return `${time}\n${dateStr}`
  } catch (err) {
    console.error('[MonitoringDetail] Error formatting timestamp:', err)
    return ''
  }
}

function formatNumber(value) {
  if (!value) return '0'
  const num = parseFloat(value)
  return num.toFixed(2)
}

onMounted(() => {
  fetchDetail()
})
</script>

<style scoped>
.monitoring-detail-page {
  padding: 0;
  padding-bottom: 20px;
  min-height: 100vh;
  background: var(--h-bg-page);
}

.monitoring-detail-page > :not(.van-nav-bar) {
  padding-left: var(--h-spacing-lg);
  padding-right: var(--h-spacing-lg);
}

/* Detail Header */
.detail-header {
  margin-top: var(--h-spacing-lg);
  margin-bottom: var(--h-spacing-lg);
}

.school-name {
  font-size: 20px;
  font-weight: 700;
  color: var(--h-text-primary);
  margin: 0 0 var(--h-spacing-lg) 0;
}

/* Menu Info Section */
.menu-info {
  display: flex;
  gap: var(--h-spacing-md);
  margin-bottom: var(--h-spacing-lg);
  padding-bottom: var(--h-spacing-lg);
  border-bottom: 1px solid var(--h-border-light);
}

.menu-image {
  width: 80px;
  height: 80px;
  border-radius: var(--h-radius-md);
  overflow: hidden;
  flex-shrink: 0;
  background: var(--h-bg-light);
}

.menu-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.menu-details {
  flex: 1;
  min-width: 0;
}

.menu-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--h-text-primary);
  margin: 0 0 var(--h-spacing-sm) 0;
  line-height: 1.4;
}

/* Nutrition Info */
.nutrition-info {
  margin-top: var(--h-spacing-sm);
}

.nutrition-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--h-spacing-sm);
  margin-bottom: 6px;
}

.nutrition-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nutrition-label {
  font-size: 11px;
  font-weight: 500;
  color: var(--h-text-light);
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.nutrition-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--h-text-primary);
}

.nutrition-note {
  font-size: 11px;
  color: var(--h-text-light);
  margin: 0;
  font-style: italic;
}

.header-info {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-md);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-label {
  font-size: 14px;
  color: var(--h-text-secondary);
}

.info-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--h-text-primary);
}

.info-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
}

.status-success {
  color: #05CD99;
  background: rgba(5, 205, 153, 0.12);
}

.status-warning {
  color: #FFB547;
  background: rgba(255, 181, 71, 0.12);
}

.status-primary {
  color: #5A4372;
  background: rgba(90, 67, 114, 0.12);
}

.status-default {
  color: var(--h-text-secondary);
  background: var(--h-bg-light);
}

/* Timeline Section */
.timeline-section {
  margin-bottom: var(--h-spacing-lg);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--h-text-primary);
  margin: 0 0 var(--h-spacing-xl) 0;
}

.timeline {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.timeline-item {
  display: flex;
  gap: var(--h-spacing-md);
  position: relative;
}

.timeline-marker {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  padding-top: 2px;
}

.timeline-dot {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #E8E8E8;
  border: 3px solid #E8E8E8;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.08);
}

.timeline-dot--active {
  background: #05CD99;
  border-color: #05CD99;
  box-shadow: 0 0 0 4px rgba(5, 205, 153, 0.15);
}

.timeline-line {
  width: 3px;
  flex: 1;
  background: #E8E8E8;
  margin: 6px 0;
  border-radius: 2px;
  min-height: 40px;
}

.timeline-item--active .timeline-line {
  background: linear-gradient(180deg, #05CD99 0%, #E8E8E8 100%);
}

.timeline-content {
  flex: 1;
  padding-bottom: var(--h-spacing-lg);
  background: var(--h-bg-light);
  padding: var(--h-spacing-md);
  border-radius: var(--h-radius-md);
  border-left: 3px solid transparent;
  transition: all 0.3s ease;
}

.timeline-item--active .timeline-content {
  background: rgba(5, 205, 153, 0.05);
  border-left-color: #05CD99;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--h-spacing-md);
}

.timeline-header-left {
  flex: 1;
  min-width: 0;
}

.timeline-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--h-text-primary);
  margin: 0 0 6px 0;
  line-height: 1.4;
}

.timeline-item--active .timeline-title {
  color: #05CD99;
}

.timeline-description {
  font-size: 13px;
  color: var(--h-text-secondary);
  margin: 0;
  line-height: 1.5;
}

.timeline-time {
  font-size: 11px;
  font-weight: 600;
  color: var(--h-text-secondary);
  text-align: right;
  white-space: pre-line;
  line-height: 1.4;
  flex-shrink: 0;
  min-width: 50px;
}

.timeline-item--active .timeline-time {
  color: #05CD99;
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
</style>
