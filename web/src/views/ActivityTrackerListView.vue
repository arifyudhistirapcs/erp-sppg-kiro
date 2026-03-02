<template>
  <div class="activity-tracker-list">
    <div class="page-header">
      <h1>Aktivitas Pelacakan</h1>
      <p class="subtitle">Monitor proses order dari persiapan hingga selesai</p>
    </div>

    <div class="filters-section">
      <a-row :gutter="16" align="middle">
        <a-col :span="7">
          <a-date-picker
            v-model:value="selectedDate"
            format="YYYY-MM-DD"
            placeholder="Pilih tanggal"
            style="width: 100%"
            @change="fetchOrders"
          />
        </a-col>
        <a-col :span="7">
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
        <a-col :span="7">
          <a-input-search
            v-model:value="searchQuery"
            placeholder="Cari menu atau sekolah..."
            @search="fetchOrders"
            allow-clear
            @clear="fetchOrders"
          />
        </a-col>
        <a-col :span="3">
          <a-button 
            type="default" 
            :icon="h(ReloadOutlined)" 
            @click="fetchOrders"
            :loading="loading"
            style="width: 100%"
          >
            Refresh
          </a-button>
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

      <div class="orders-list">
        <a-list
          :data-source="orders"
          :bordered="false"
        >
          <template #renderItem="{ item: order }">
            <a-list-item
              class="order-list-item"
              @click="showOrderDetail(order)"
            >
              <a-list-item-meta>
                <template #title>
                  <div class="order-list-title">
                    {{ order.menu.name }}
                  </div>
                </template>
                <template #description>
                  <div class="order-list-info">
                    <span class="info-item">
                      <environment-outlined />
                      {{ order.school.name }}
                    </span>
                    <span class="info-item">
                      <team-outlined />
                      <template v-if="order.portions_small > 0 && order.portions_large > 0">
                        {{ order.portions_small }} porsi kecil + {{ order.portions_large }} porsi besar
                      </template>
                      <template v-else-if="order.portions_small > 0">
                        {{ order.portions_small }} porsi kecil
                      </template>
                      <template v-else-if="order.portions_large > 0">
                        {{ order.portions_large }} porsi besar
                      </template>
                      <template v-else>
                        {{ order.portions }} porsi
                      </template>
                    </span>
                  </div>
                </template>
              </a-list-item-meta>
              <template #actions>
                <a-tag :color="getStatusColor(order.current_status)">
                  Stage {{ order.current_stage }}: {{ getStatusLabel(order.current_status) }}
                </a-tag>
              </template>
            </a-list-item>
          </template>
        </a-list>
      </div>
    </div>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Aktivitas Pelacakan"
      width="800px"
      :footer="null"
    >
      <div v-if="selectedOrder" class="order-detail-modal">
        <div class="order-header">
          <h3>{{ selectedOrder.menu.name }}</h3>
          <div class="order-meta">
            <span><environment-outlined /> {{ selectedOrder.school.name }}</span>
            <span>
              <team-outlined />
              <template v-if="selectedOrder.portions_small > 0 && selectedOrder.portions_large > 0">
                {{ selectedOrder.portions_small }} porsi kecil + {{ selectedOrder.portions_large }} porsi besar
              </template>
              <template v-else-if="selectedOrder.portions_small > 0">
                {{ selectedOrder.portions_small }} porsi kecil
              </template>
              <template v-else-if="selectedOrder.portions_large > 0">
                {{ selectedOrder.portions_large }} porsi besar
              </template>
              <template v-else>
                {{ selectedOrder.portions }} porsi
              </template>
            </span>
          </div>
        </div>

        <a-divider />

        <div class="timeline-container">
          <a-timeline>
            <a-timeline-item
              v-for="stage in orderTimeline"
              :key="stage.stage"
              :color="stage.completed ? 'green' : stage.inProgress ? 'blue' : 'gray'"
            >
              <template #dot>
                <check-circle-filled v-if="stage.completed" style="font-size: 16px" />
                <clock-circle-outlined v-else-if="stage.inProgress" style="font-size: 16px" />
                <span v-else class="timeline-dot-empty"></span>
              </template>
              <div class="timeline-content">
                <div class="timeline-title">
                  <strong>Stage {{ stage.stage }}: {{ stage.label }}</strong>
                  <a-tag v-if="stage.inProgress" color="processing">Sedang Berlangsung</a-tag>
                </div>
                <div class="timeline-description">{{ stage.description }}</div>
                <div v-if="stage.timestamp" class="timeline-timestamp">
                  {{ formatTimestamp(stage.timestamp) }}
                </div>
              </div>
            </a-timeline-item>
          </a-timeline>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, h, computed } from 'vue';
import { useRouter } from 'vue-router';
import dayjs from 'dayjs';
import 'dayjs/locale/id';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';
import api from '@/services/api';

// Configure dayjs plugins
dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.locale('id');
import {
  InboxOutlined,
  ShoppingOutlined,
  EnvironmentOutlined,
  TeamOutlined,
  CheckCircleFilled,
  ClockCircleOutlined,
  ReloadOutlined,
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
const detailModalVisible = ref(false);
const selectedOrder = ref(null);
const orderActivityLog = ref([]);
let refreshInterval = null;

const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

const fetchOrders = async (isRetry = false) => {
  if (!isRetry) {
    retryCount.value = 0;
  }
  
  // Ensure selectedDate has a value
  if (!selectedDate.value) {
    selectedDate.value = dayjs();
  }
  
  loading.value = true;
  try {
    const params = {
      date: selectedDate.value.format('YYYY-MM-DD'),
    };
    
    if (selectedSchoolId.value) {
      params.school_id = selectedSchoolId.value;
    }
    
    if (searchQuery.value && searchQuery.value.trim() !== '') {
      params.search = searchQuery.value.trim();
    }
    
    console.log('Fetching orders with params:', params);
    
    const response = await api.get('/activity-tracker/orders', { params });
    
    console.log('Orders response:', response.data);
    
    if (response.data.success) {
      orders.value = response.data.data.orders || [];
      summary.value = response.data.data.summary || { total_orders: 0, status_distribution: {} };
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
    console.log('Fetching schools...');
    const response = await api.get('/schools');
    console.log('Schools response:', response.data);
    if (response.data.success) {
      schools.value = response.data.data;
      console.log('Schools loaded:', schools.value.length, 'schools');
    } else {
      console.error('Schools API returned success=false');
    }
  } catch (error) {
    console.error('Error fetching schools:', error);
    console.error('Error details:', error.response?.data);
    message.warning('Gagal memuat daftar sekolah');
  }
};

const navigateToDetail = (orderId) => {
  router.push({ name: 'ActivityTrackerDetail', params: { id: orderId } });
};

const getStatusColor = (status) => {
  const stageColors = {
    order_disiapkan: 'default',
    sedang_dimasak: 'processing',
    selesai_dimasak: 'success',
    siap_dipacking: 'default',
    selesai_dipacking: 'success',
    siap_dikirim: 'success',
    diperjalanan: 'processing',
    sudah_sampai_sekolah: 'success',
    sudah_diterima_pihak_sekolah: 'success',
    driver_menuju_lokasi_pengambilan: 'processing',
    driver_tiba_di_lokasi_pengambilan: 'success',
    driver_kembali_ke_sppg: 'processing',
    driver_tiba_di_sppg: 'success',
    ompreng_siap_dicuci: 'default',
    ompreng_proses_pencucian: 'processing',
    ompreng_selesai_dicuci: 'success',
  };
  return stageColors[status] || 'default';
};

const getStatusLabel = (status) => {
  const labels = {
    order_disiapkan: 'Sedang Disiapkan',
    sedang_dimasak: 'Sedang Dimasak',
    selesai_dimasak: 'Selesai Dimasak',
    siap_dipacking: 'Siap Dipacking',
    selesai_dipacking: 'Selesai Dipacking',
    siap_dikirim: 'Siap Dikirim',
    diperjalanan: 'Dalam Perjalanan',
    sudah_sampai_sekolah: 'Sudah Tiba',
    sudah_diterima_pihak_sekolah: 'Sudah Diterima',
    driver_menuju_lokasi_pengambilan: 'Menuju Lokasi',
    driver_tiba_di_lokasi_pengambilan: 'Tiba di Lokasi',
    driver_kembali_ke_sppg: 'Kembali',
    driver_tiba_di_sppg: 'Tiba di SPPG',
    ompreng_siap_dicuci: 'Siap Dicuci',
    ompreng_proses_pencucian: 'Sedang Dicuci',
    ompreng_selesai_dicuci: 'Selesai Dicuci',
  };
  return labels[status] || status;
};

const showOrderDetail = async (order) => {
  selectedOrder.value = order;
  detailModalVisible.value = true;
  
  // Fetch activity log
  try {
    const response = await api.get(`/activity-tracker/orders/${order.id}/activity`);
    if (response.data.success) {
      orderActivityLog.value = response.data.data;
    }
  } catch (error) {
    console.error('Error fetching activity log:', error);
    message.error('Gagal memuat log aktivitas');
  }
};

const orderTimeline = computed(() => {
  if (!selectedOrder.value || !orderActivityLog.value || !Array.isArray(orderActivityLog.value)) return [];
  
  // Backend returns full timeline with all stages
  return orderActivityLog.value.map(stage => ({
    stage: stage.stage,
    status: stage.status,
    label: stage.title,
    description: stage.description,
    completed: stage.is_completed,
    inProgress: !stage.is_completed && stage.stage === selectedOrder.value.current_stage,
    timestamp: stage.completed_at || stage.started_at,
  }));
});

const formatTimestamp = (timestamp) => {
  if (!timestamp) return '';
  // Backend sends timestamp in WIB: "2026-02-28T14:35:33.390103+07:00"
  // Extract time directly from the string to avoid any timezone conversion
  const timeStr = timestamp.toString();
  const dateMatch = timeStr.match(/(\d{4})-(\d{2})-(\d{2})/);
  const timeMatch = timeStr.match(/T(\d{2}):(\d{2})/);
  
  if (dateMatch && timeMatch) {
    const [, year, month, day] = dateMatch;
    const [, hour, minute] = timeMatch;
    
    // Map day of week in Indonesian
    const date = new Date(year, parseInt(month) - 1, day);
    const days = ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'];
    const months = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 
                    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
    
    return `${days[date.getDay()]}, ${parseInt(day)} ${months[parseInt(month) - 1]} ${year} ${hour}:${minute}`;
  }
  
  // Fallback
  return dayjs(timestamp).format('dddd, DD MMMM YYYY HH:mm');
};

onMounted(() => {
  fetchSchools();
  fetchOrders();
  
  // Auto-refresh every 10 seconds
  refreshInterval = setInterval(() => {
    fetchOrders();
  }, 10000);
});

onUnmounted(() => {
  // Clear interval when component is destroyed
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
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

.orders-list {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.order-list-item {
  cursor: pointer;
  padding: 16px 24px;
  transition: background-color 0.3s;
}

.order-list-item:hover {
  background-color: #f5f5f5;
}

.order-list-title {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.order-list-info {
  display: flex;
  gap: 24px;
  margin-top: 4px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #595959;
}

.order-detail-modal {
  max-height: 600px;
  overflow-y: auto;
}

.order-header h3 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 8px;
}

.order-meta {
  display: flex;
  gap: 16px;
  color: #595959;
}

.order-meta span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.timeline-container {
  padding: 16px 0;
}

.timeline-content {
  padding-bottom: 16px;
}

.timeline-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.timeline-description {
  color: #595959;
  font-size: 14px;
  margin-bottom: 4px;
}

.timeline-timestamp {
  color: #8c8c8c;
  font-size: 13px;
}

.timeline-dot-empty {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid #d9d9d9;
  background: #fff;
}
</style>
