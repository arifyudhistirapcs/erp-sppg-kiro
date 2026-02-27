<template>
  <div class="activity-tracker-detail">
    <div class="back-button">
      <a-button @click="goBack">
        <template #icon><arrow-left-outlined /></template>
        Kembali
      </a-button>
    </div>

    <div v-if="loading" class="loading-container">
      <a-spin size="large" />
    </div>

    <div v-else-if="order" class="detail-content">
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
          <h1>Aktivitas Pelacakan</h1>
          <h2 class="menu-name">{{ order.menu.name }}</h2>
          <div class="info-row">
            <div class="info-item">
              <environment-outlined />
              <span>{{ order.school.name }}</span>
            </div>
            <div class="info-item">
              <calendar-outlined />
              <span>{{ formatDate(order.order_date) }}</span>
            </div>
          </div>
          <div class="info-row">
            <div class="info-item">
              <team-outlined />
              <span>{{ order.portions }} porsi</span>
            </div>
            <div class="info-item">
              <user-outlined />
              <span>{{ order.driver.name }}</span>
            </div>
          </div>
          <div class="current-status">
            <a-tag :color="getStatusColor(order.current_status)" style="font-size: 14px; padding: 4px 12px;">
              Stage {{ order.current_stage }}: {{ getStatusLabel(order.current_status) }}
            </a-tag>
          </div>
        </div>
      </div>

      <div class="timeline-container">
        <VerticalTimeline :timeline="order.timeline" :current-stage="order.current_stage" />
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

const getStatusColor = (status) => {
  const stageColors = {
    order_disiapkan: 'default',
    order_dimasak: 'processing',
    order_dikemas: 'processing',
    order_siap_diambil: 'success',
    pesanan_dalam_perjalanan: 'processing',
    pesanan_sudah_tiba: 'success',
    pesanan_sudah_diterima: 'success',
    driver_menuju_lokasi: 'processing',
    driver_tiba_di_lokasi: 'success',
    driver_kembali: 'processing',
    driver_tiba_di_sppg: 'success',
    ompreng_siap_dicuci: 'default',
    ompreng_sedang_dicuci: 'processing',
    ompreng_selesai_dicuci: 'success',
    ompreng_siap_digunakan: 'success',
    order_selesai: 'success',
  };
  return stageColors[status] || 'default';
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
  padding: 24px;
  max-width: 800px;
  margin: 0 auto;
}

.back-button {
  margin-bottom: 24px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.detail-content {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.order-header {
  display: flex;
  gap: 24px;
  padding: 24px;
  border-bottom: 1px solid #f0f0f0;
}

.header-image {
  width: 200px;
  height: 200px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background: #f5f5f5;
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
  color: #d9d9d9;
}

.header-info {
  flex: 1;
}

.header-info h1 {
  font-size: 16px;
  color: #8c8c8c;
  margin-bottom: 8px;
  font-weight: 400;
}

.menu-name {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 16px;
}

.info-row {
  display: flex;
  gap: 24px;
  margin-bottom: 12px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #595959;
  font-size: 14px;
}

.current-status {
  margin-top: 16px;
}

.timeline-container {
  padding: 24px;
}

@media (max-width: 768px) {
  .order-header {
    flex-direction: column;
  }

  .header-image {
    width: 100%;
    height: 250px;
  }
}
</style>
