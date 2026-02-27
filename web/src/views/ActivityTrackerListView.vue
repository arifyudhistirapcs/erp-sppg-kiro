<template>
  <div class="activity-tracker-list">
    <div class="page-header">
      <h1>Aktivitas Pelacakan</h1>
      <p class="subtitle">Monitor proses order dari persiapan hingga selesai</p>
    </div>

    <div class="filters-section">
      <a-row :gutter="16">
        <a-col :span="8">
          <a-date-picker
            v-model:value="selectedDate"
            format="YYYY-MM-DD"
            placeholder="Pilih tanggal"
            style="width: 100%"
            @change="fetchOrders"
          />
        </a-col>
        <a-col :span="8">
          <a-select
            v-model:value="selectedSchoolId"
            placeholder="Semua Sekolah"
            style="width: 100%"
            allow-clear
            @change="fetchOrders"
          >
            <a-select-option :value="null">Semua Sekolah</a-select-option>
            <a-select-option
              v-for="school in schools"
              :key="school.id"
              :value="school.id"
            >
              {{ school.name }}
            </a-select-option>
          </a-select>
        </a-col>
        <a-col :span="8">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Cari menu atau sekolah..."
            @search="fetchOrders"
          />
        </a-col>
      </a-row>
    </div>

    <div v-if="loading" class="loading-container">
      <a-spin size="large" />
    </div>

    <div v-else-if="orders.length === 0" class="empty-state">
      <a-empty description="Tidak ada order untuk tanggal ini">
        <template #image>
          <inbox-outlined style="font-size: 64px; color: #d9d9d9" />
        </template>
      </a-empty>
    </div>

    <div v-else>
      <div class="summary-section">
        <a-row :gutter="16">
          <a-col :span="8">
            <a-statistic
              title="Total Order"
              :value="summary.total_orders"
              :prefix="h(ShoppingOutlined)"
            />
          </a-col>
          <a-col :span="16">
            <div class="status-distribution">
              <span class="label">Status:</span>
              <a-tag
                v-for="(count, status) in summary.status_distribution"
                :key="status"
                :color="getStatusColor(status)"
              >
                {{ getStatusLabel(status) }}: {{ count }}
              </a-tag>
            </div>
          </a-col>
        </a-row>
      </div>

      <div class="orders-grid">
        <a-row :gutter="[16, 16]">
          <a-col
            v-for="order in orders"
            :key="order.id"
            :xs="24"
            :sm="12"
            :md="8"
            :lg="6"
          >
            <a-card
              hoverable
              class="order-card"
              @click="navigateToDetail(order.id)"
            >
              <template #cover>
                <div class="order-image">
                  <img
                    v-if="order.menu.photo_url"
                    :src="order.menu.photo_url"
                    :alt="order.menu.name"
                  />
                  <div v-else class="no-image">
                    <picture-outlined style="font-size: 48px" />
                  </div>
                </div>
              </template>
              <a-card-meta>
                <template #title>
                  <div class="order-title">{{ order.menu.name }}</div>
                </template>
                <template #description>
                  <div class="order-info">
                    <div class="school-name">
                      <environment-outlined />
                      {{ order.school.name }}
                    </div>
                    <div class="portions">
                      <team-outlined />
                      {{ order.portions }} porsi
                    </div>
                    <div class="status-badge">
                      <a-tag :color="getStatusColor(order.current_status)">
                        Stage {{ order.current_stage }}: {{ getStatusLabel(order.current_status) }}
                      </a-tag>
                    </div>
                  </div>
                </template>
              </a-card-meta>
            </a-card>
          </a-col>
        </a-row>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue';
import { useRouter } from 'vue-router';
import dayjs from 'dayjs';
import api from '@/services/api';
import {
  InboxOutlined,
  ShoppingOutlined,
  EnvironmentOutlined,
  TeamOutlined,
  PictureOutlined,
} from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';

const router = useRouter();

const selectedDate = ref(dayjs());
const selectedSchoolId = ref(null);
const searchQuery = ref('');
const orders = ref([]);
const schools = ref([]);
const summary = ref({
  total_orders: 0,
  status_distribution: {},
});
const loading = ref(false);
const retryCount = ref(0);
const maxRetries = 3;

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

const fetchOrders = async (isRetry = false) => {
  if (!isRetry) {
    retryCount.value = 0;
  }
  
  loading.value = true;
  try {
    const params = {
      date: selectedDate.value.format('YYYY-MM-DD'),
    };
    
    if (selectedSchoolId.value) {
      params.school_id = selectedSchoolId.value;
    }
    
    if (searchQuery.value) {
      params.search = searchQuery.value;
    }
    
    const response = await api.get('/activity-tracker/orders', { params });
    
    if (response.data.success) {
      orders.value = response.data.data.orders;
      summary.value = response.data.data.summary;
      retryCount.value = 0;
    }
  } catch (error) {
    console.error('Error fetching orders:', error);
    
    // Retry with exponential backoff
    if (retryCount.value < maxRetries) {
      retryCount.value++;
      const backoffTime = Math.pow(2, retryCount.value) * 1000; // 2s, 4s, 8s
      message.warning(`Gagal memuat data. Mencoba lagi dalam ${backoffTime / 1000} detik... (${retryCount.value}/${maxRetries})`);
      await sleep(backoffTime);
      return fetchOrders(true);
    } else {
      message.error('Gagal memuat data order setelah beberapa percobaan. Silakan coba lagi nanti.');
    }
  } finally {
    loading.value = false;
  }
};

const fetchSchools = async () => {
  try {
    const response = await api.get('/schools');
    if (response.data.success) {
      schools.value = response.data.data;
    }
  } catch (error) {
    console.error('Error fetching schools:', error);
    message.warning('Gagal memuat daftar sekolah');
  }
};

const navigateToDetail = (orderId) => {
  router.push({ name: 'ActivityTrackerDetail', params: { id: orderId } });
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
  fetchSchools();
  fetchOrders();
});
</script>

<style scoped>
.activity-tracker-list {
  padding: 24px;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 28px;
  font-weight: 600;
  margin-bottom: 8px;
}

.subtitle {
  color: #8c8c8c;
  font-size: 14px;
}

.filters-section {
  margin-bottom: 24px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background: #fff;
  border-radius: 8px;
}

.summary-section {
  margin-bottom: 24px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.status-distribution {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.status-distribution .label {
  font-weight: 500;
  color: #595959;
}

.orders-grid {
  margin-top: 16px;
}

.order-card {
  cursor: pointer;
  transition: all 0.3s;
}

.order-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.order-image {
  height: 180px;
  overflow: hidden;
  background: #f5f5f5;
  display: flex;
  align-items: center;
  justify-content: center;
}

.order-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-image {
  color: #d9d9d9;
}

.order-title {
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.school-name,
.portions {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #595959;
}

.status-badge {
  margin-top: 4px;
}
</style>
