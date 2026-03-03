<template>
  <div>
    <div class="activity-tracker-detail">
      <!-- Back Button -->
      <div class="back-button">
        <a-button @click="goBack" class="h-button-secondary">
          <template #icon><arrow-left-outlined /></template>
          Kembali
        </a-button>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="loading-container h-card">
        <a-spin size="large" />
      </div>

      <!-- Order Details -->
      <div v-else-if="order" class="detail-content">
        <!-- Order Header Card -->
        <div class="order-header-card h-card">
          <div class="order-header">
            <div class="header-image">
              <img
                v-if="order.menu.photo_url"
                :src="order.menu.photo_url"
                :alt="order.menu.name"
              />
              <div v-else class="no-image">
                <picture-outlined style="font-size: 64px" />
              </div>
            </div>
            <div class="header-info">
              <div class="info-label">Aktivitas Pelacakan</div>
              <h2 class="menu-name">{{ order.menu.name }}</h2>
              
              <div class="info-grid">
                <div class="info-item">
                  <environment-outlined class="info-icon" />
                  <div class="info-content">
                    <div class="info-label-small">Sekolah</div>
                    <div class="info-value">{{ order.school.name }}</div>
                  </div>
                </div>
                <div class="info-item">
                  <calendar-outlined class="info-icon" />
                  <div class="info-content">
                    <div class="info-label-small">Tanggal</div>
                    <div class="info-value">{{ formatDate(order.order_date) }}</div>
                  </div>
                </div>
                <div class="info-item">
                  <team-outlined class="info-icon" />
                  <div class="info-content">
                    <div class="info-label-small">Porsi</div>
                    <div class="info-value">{{ order.portions }} porsi</div>
                  </div>
                </div>
                <div class="info-item">
                  <user-outlined class="info-icon" />
                  <div class="info-content">
                    <div class="info-label-small">Driver</div>
                    <div class="info-value">{{ order.driver.name }}</div>
                  </div>
                </div>
              </div>

              <div class="current-status">
                <div class="status-badge" :class="getStatusClass(order.current_status)">
                  <div class="status-dot"></div>
                  <span>Stage {{ order.current_stage }}: {{ getStatusLabel(order.current_status) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Timeline Card -->
        <div class="timeline-card h-card">
          <div class="card-header">
            <h3 class="card-title">Timeline Aktivitas</h3>
          </div>
          <div class="timeline-container">
            <VerticalTimeline :timeline="order.timeline" :current-stage="order.current_stage" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import api from '@/services/api';
import dayjs from 'dayjs';
import {
  ArrowLeftOutlined,
  PictureOutlined,
  EnvironmentOutlined,
  CalendarOutlined,
  TeamOutlined,
  UserOutlined,
} from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';
import VerticalTimeline from '../components/VerticalTimeline.vue';

const router = useRouter();
const route = useRoute();

const order = ref(null);
const loading = ref(false);
const retryCount = ref(0);
const maxRetries = 3;

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

const fetchOrderDetails = async (isRetry = false) => {
  if (!isRetry) {
    retryCount.value = 0;
  }
  
  loading.value = true;
  try {
    const orderId = route.params.id;
    const response = await api.get(`/activity-tracker/orders/${orderId}`);
    
    if (response.data.success) {
      order.value = response.data.data;
      retryCount.value = 0;
      
      // Debug: Log timeline data
      console.log('=== Order Timeline Data ===');
      console.log('Full order:', order.value);
      console.log('Timeline stages:', order.value.timeline);
      if (order.value.timeline && order.value.timeline.length > 0) {
        console.log('First stage with completed_at:', order.value.timeline.find(s => s.completed_at));
      }
      console.log('===========================');
    }
  } catch (error) {
    console.error('Error fetching order details:', error);
    
    // Handle 404 - order not found
    if (error.response?.status === 404) {
      message.error('Order tidak ditemukan');
      router.push({ name: 'activity-tracker' });
      return;
    }
    
    // Retry with exponential backoff for other errors
    if (retryCount.value < maxRetries) {
      retryCount.value++;
      const backoffTime = Math.pow(2, retryCount.value) * 1000; // 2s, 4s, 8s
      message.warning(`Gagal memuat data. Mencoba lagi dalam ${backoffTime / 1000} detik... (${retryCount.value}/${maxRetries})`);
      await sleep(backoffTime);
      return fetchOrderDetails(true);
    } else {
      message.error('Gagal memuat detail order setelah beberapa percobaan. Silakan coba lagi nanti.');
      router.push({ name: 'activity-tracker' });
    }
  } finally {
    loading.value = false;
  }
};

const goBack = () => {
  router.push({ name: 'ActivityTrackerList' });
};

const formatDate = (dateStr) => {
  return dayjs(dateStr).format('DD MMMM YYYY');
};

const getStatusClass = (status) => {
  const stageClasses = {
    order_disiapkan: 'status-pending',
    order_dimasak: 'status-processing',
    order_dikemas: 'status-processing',
    order_siap_diambil: 'status-success',
    pesanan_dalam_perjalanan: 'status-processing',
    pesanan_sudah_tiba: 'status-success',
    pesanan_sudah_diterima: 'status-success',
    driver_menuju_lokasi: 'status-processing',
    driver_tiba_di_lokasi: 'status-success',
    driver_kembali: 'status-processing',
    driver_tiba_di_sppg: 'status-success',
    ompreng_siap_dicuci: 'status-pending',
    ompreng_sedang_dicuci: 'status-processing',
    ompreng_selesai_dicuci: 'status-success',
    ompreng_siap_digunakan: 'status-success',
    order_selesai: 'status-success',
  };
  return stageClasses[status] || 'status-pending';
};

const getStatusLabel = (status) => {
  const labels = {
    order_disiapkan: 'Sedang Disiapkan',
    order_dimasak: 'Sedang Dimasak',
    order_dikemas: 'Sedang Dikemas',
    order_siap_diambil: 'Siap Diambil',
    pesanan_dalam_perjalanan: 'Dalam Perjalanan',
    pesanan_sudah_tiba: 'Sudah Tiba',
    pesanan_sudah_diterima: 'Sudah Diterima',
    driver_menuju_lokasi: 'Menuju Lokasi',
    driver_tiba_di_lokasi: 'Tiba di Lokasi',
    driver_kembali: 'Kembali',
    driver_tiba_di_sppg: 'Tiba di SPPG',
    ompreng_siap_dicuci: 'Siap Dicuci',
    ompreng_sedang_dicuci: 'Sedang Dicuci',
    ompreng_selesai_dicuci: 'Selesai Dicuci',
    ompreng_siap_digunakan: 'Siap Digunakan',
    order_selesai: 'Selesai',
  };
  return labels[status] || status;
};

onMounted(() => {
  fetchOrderDetails();
});
</script>

<style scoped>
.activity-tracker-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.back-button {
  margin-bottom: var(--h-spacing-6, 24px);
}

.h-button-secondary {
  border-radius: var(--h-radius-md, 12px);
  height: 40px;
  padding: 0 var(--h-spacing-4, 16px);
  font-weight: 500;
  transition: all var(--transition-base, 200ms);
}

.h-button-secondary:hover {
  transform: translateY(-2px);
  box-shadow: var(--h-shadow-md, 0px 4px 6px rgba(0, 0, 0, 0.07));
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-6, 24px);
}

/* Order Header Card */
.order-header-card {
  background: var(--h-bg-card, #FFFFFF);
  border-radius: var(--h-radius-lg, 16px);
  box-shadow: var(--h-shadow-card, 0px 18px 40px rgba(112, 144, 176, 0.12));
  padding: var(--h-spacing-6, 24px);
  transition: all var(--transition-base, 200ms);
}

.dark .order-header-card {
  background: var(--h-bg-card-dark, #3D2B53);
}

.order-header {
  display: flex;
  gap: var(--h-spacing-6, 24px);
}

.header-image {
  width: 240px;
  height: 240px;
  flex-shrink: 0;
  border-radius: var(--h-radius-lg, 16px);
  overflow: hidden;
  background: var(--h-bg-light, #ACA9B0);
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  color: var(--h-text-light, #ACA9B0);
}

.header-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4, 16px);
}

.info-label {
  font-size: var(--text-sm, 14px);
  color: var(--h-text-secondary, #74788C);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.dark .info-label {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.menu-name {
  font-size: var(--text-2xl, 24px);
  font-weight: 700;
  color: var(--h-text-primary, #322837);
  margin: 0;
  line-height: 1.3;
}

.dark .menu-name {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--h-spacing-4, 16px);
}

.info-item {
  display: flex;
  align-items: flex-start;
  gap: var(--h-spacing-3, 12px);
  padding: var(--h-spacing-3, 12px);
  background: var(--h-bg-primary, #F8FDEA);
  border-radius: var(--h-radius-md, 12px);
  transition: all var(--transition-base, 200ms);
}

.dark .info-item {
  background: rgba(90, 67, 114, 0.2);
}

.info-icon {
  font-size: 20px;
  color: var(--h-primary, #5A4372);
  margin-top: 2px;
}

.dark .info-icon {
  color: var(--h-primary-light, #6a5382);
}

.info-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label-small {
  font-size: var(--text-xs, 12px);
  color: var(--h-text-secondary, #74788C);
  font-weight: 500;
}

.dark .info-label-small {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.info-value {
  font-size: var(--text-base, 16px);
  color: var(--h-text-primary, #322837);
  font-weight: 600;
}

.dark .info-value {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.current-status {
  margin-top: var(--h-spacing-2, 8px);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: var(--h-spacing-2, 8px);
  padding: var(--h-spacing-2, 8px) var(--h-spacing-4, 16px);
  border-radius: var(--h-radius-full, 9999px);
  font-size: var(--text-sm, 14px);
  font-weight: 600;
  transition: all var(--transition-base, 200ms);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-pending {
  background: rgba(172, 169, 176, 0.15);
  color: var(--h-text-secondary, #74788C);
}

.status-pending .status-dot {
  background: var(--h-text-secondary, #74788C);
}

.status-processing {
  background: rgba(255, 181, 71, 0.15);
  color: var(--warning, #FFB547);
}

.status-processing .status-dot {
  background: var(--warning, #FFB547);
  animation: pulse 2s ease-in-out infinite;
}

.status-success {
  background: rgba(5, 205, 153, 0.15);
  color: var(--success, #05CD99);
}

.status-success .status-dot {
  background: var(--success, #05CD99);
}

.dark .status-pending {
  background: rgba(172, 169, 176, 0.25);
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.dark .status-processing {
  background: rgba(255, 181, 71, 0.25);
}

.dark .status-success {
  background: rgba(5, 205, 153, 0.25);
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

/* Timeline Card */
.timeline-card {
  background: var(--h-bg-card, #FFFFFF);
  border-radius: var(--h-radius-lg, 16px);
  box-shadow: var(--h-shadow-card, 0px 18px 40px rgba(112, 144, 176, 0.12));
  padding: var(--h-spacing-6, 24px);
  transition: all var(--transition-base, 200ms);
}

.dark .timeline-card {
  background: var(--h-bg-card-dark, #3D2B53);
}

.card-header {
  margin-bottom: var(--h-spacing-6, 24px);
  padding-bottom: var(--h-spacing-4, 16px);
  border-bottom: 1px solid var(--h-border-color, #E9EDF7);
}

.dark .card-header {
  border-bottom-color: var(--h-border-color-dark, #5A4372);
}

.card-title {
  font-size: var(--text-lg, 18px);
  font-weight: 700;
  color: var(--h-text-primary, #322837);
  margin: 0;
}

.dark .card-title {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.timeline-container {
  padding: var(--h-spacing-4, 16px) 0;
}

/* Responsive Design */
@media (max-width: 768px) {
  .order-header {
    flex-direction: column;
  }

  .header-image {
    width: 100%;
    height: 200px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .back-button {
    margin-bottom: var(--h-spacing-4, 16px);
  }

  .detail-content {
    gap: var(--h-spacing-4, 16px);
  }

  .order-header-card,
  .timeline-card {
    padding: var(--h-spacing-4, 16px);
  }
}

/* Touch targets for mobile */
@media (max-width: 768px) {
  .h-button-secondary {
    min-height: 44px;
    height: 44px;
  }
}
</style>
