<template>
  <div>
    <!-- Page Subtitle -->
    <div class="page-subtitle">
      Monitor proses order dari persiapan hingga selesai
    </div>

    <!-- Filters Section -->
    <div class="h-card filters-section">
      <a-row :gutter="[16, 16]" align="middle">
        <a-col :xs="24" :sm="12" :md="8">
          <a-date-picker
            v-model:value="selectedDate"
            format="YYYY-MM-DD"
            placeholder="Pilih tanggal"
            style="width: 100%"
            @change="fetchOrders"
            size="large"
          />
        </a-col>
        <a-col :xs="24" :sm="12" :md="12">
          <a-input
            v-model:value="searchQuery"
            placeholder="Cari menu atau sekolah..."
            @change="fetchOrders"
            allow-clear
            size="large"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-col>
        <a-col :xs="24" :sm="24" :md="4">
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

    <!-- Status Stats Cards -->
    <div class="stats-row">
      <HStatCard
        :icon="ClockCircleOutlined"
        icon-bg="linear-gradient(135deg, #FFB547 0%, #FF9A3D 100%)"
        label="Pending"
        :value="statusCounts.pending"
        :loading="loading"
      />
      <HStatCard
        :icon="SyncOutlined"
        icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
        label="In Progress"
        :value="statusCounts.inProgress"
        :loading="loading"
      />
      <HStatCard
        :icon="CheckCircleOutlined"
        icon-bg="linear-gradient(135deg, #05CD99 0%, #04B886 100%)"
        label="Completed"
        :value="statusCounts.completed"
        :loading="loading"
      />
    </div>

    <!-- Activity List Table -->
    <HDataTable
      :columns="tableColumns"
      :data-source="tableData"
      :loading="loading"
      :pagination="{
        current: 1,
        pageSize: 20,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} aktivitas`
      }"
      :mobile-card-view="true"
    >
      <!-- Custom Cell: Menu Name -->
      <template #cell-menuName="{ record }">
        <div class="menu-cell">
          <div class="menu-name">{{ record.menuName }}</div>
          <div class="menu-school">
            <EnvironmentOutlined style="margin-right: 4px" />
            {{ record.schoolName }}
          </div>
        </div>
      </template>

      <!-- Custom Cell: Portions -->
      <template #cell-portions="{ record }">
        <div class="portions-cell">
          <TeamOutlined style="margin-right: 4px" />
          {{ record.portions }}
        </div>
      </template>

      <!-- Custom Cell: Status -->
      <template #cell-status="{ text }">
        <a-tag :color="getStatusColor(text)">
          {{ text }}
        </a-tag>
      </template>

      <!-- Custom Cell: Actions -->
      <template #actions="{ record }">
        <a-button
          type="link"
          size="small"
          @click="showOrderDetail(record)"
        >
          Detail
        </a-button>
      </template>
    </HDataTable>

    <!-- Detail Modal -->
    <a-modal
      v-model:open="detailModalVisible"
      title="Detail Aktivitas Pelacakan"
      width="800px"
      :footer="null"
      class="detail-modal"
    >
      <div v-if="selectedOrder" class="order-detail-modal">
        <div class="order-header">
          <h3>{{ selectedOrder.menuName }}</h3>
          <div class="order-meta">
            <span>
              <EnvironmentOutlined style="margin-right: 4px" />
              {{ selectedOrder.schoolName }}
            </span>
            <span>
              <TeamOutlined style="margin-right: 4px" />
              {{ selectedOrder.portions }}
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
                <CheckCircleFilled v-if="stage.completed" style="font-size: 16px" />
                <ClockCircleOutlined v-else-if="stage.inProgress" style="font-size: 16px" />
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

// Horizon Components
import HStatCard from '@/components/horizon/HStatCard.vue';
import HDataTable from '@/components/horizon/HDataTable.vue';

// Icons
import {
  ClockCircleOutlined,
  SyncOutlined,
  CheckCircleOutlined,
  EnvironmentOutlined,
  TeamOutlined,
  CheckCircleFilled,
  ReloadOutlined,
} from '@ant-design/icons-vue';
import { message } from 'ant-design-vue';

const router = useRouter();

// State
const selectedDate = ref(dayjs());
const searchQuery = ref('');
const orders = ref([]);
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

// Computed: Status Counts
const statusCounts = computed(() => {
  const counts = {
    pending: 0,
    inProgress: 0,
    completed: 0,
  };

  orders.value.forEach(order => {
    const status = order.current_status;
    
    // Pending statuses
    if (status === 'order_disiapkan' || status === 'siap_dipacking' || status === 'siap_dikirim') {
      counts.pending++;
    }
    // In Progress statuses
    else if (
      status === 'sedang_dimasak' ||
      status === 'selesai_dipacking' ||
      status === 'diperjalanan' ||
      status === 'driver_menuju_lokasi_pengambilan' ||
      status === 'driver_kembali_ke_sppg' ||
      status === 'ompreng_proses_pencucian'
    ) {
      counts.inProgress++;
    }
    // Completed statuses
    else if (
      status === 'selesai_dimasak' ||
      status === 'sudah_sampai_sekolah' ||
      status === 'sudah_diterima_pihak_sekolah' ||
      status === 'driver_tiba_di_lokasi_pengambilan' ||
      status === 'driver_tiba_di_sppg' ||
      status === 'ompreng_selesai_dicuci'
    ) {
      counts.completed++;
    }
  });

  return counts;
});

// Computed: Table Columns
const tableColumns = computed(() => [
  {
    title: 'Menu & Sekolah',
    dataIndex: 'menuName',
    key: 'menuName',
    width: 300,
  },
  {
    title: 'Porsi',
    dataIndex: 'portions',
    key: 'portions',
    width: 150,
  },
  {
    title: 'Stage',
    dataIndex: 'stage',
    key: 'stage',
    width: 100,
  },
  {
    title: 'Status',
    dataIndex: 'status',
    key: 'status',
    type: 'status',
    width: 200,
  },
  {
    title: 'Aksi',
    key: 'actions',
    type: 'actions',
    width: 100,
  },
]);

// Computed: Table Data
const tableData = computed(() => {
  return orders.value.map((order, index) => ({
    key: order.id || index,
    id: order.id,
    menuName: order.menu?.name || '-',
    schoolName: order.school?.name || '-',
    portions: formatPortions(order),
    stage: `Stage ${order.current_stage}`,
    status: getStatusLabel(order.current_status),
    rawStatus: order.current_status,
    rawOrder: order,
  }));
});

// Helper: Format Portions
const formatPortions = (order) => {
  if (order.portions_small > 0 && order.portions_large > 0) {
    return `${order.portions_small} kecil + ${order.portions_large} besar`;
  } else if (order.portions_small > 0) {
    return `${order.portions_small} porsi kecil`;
  } else if (order.portions_large > 0) {
    return `${order.portions_large} porsi besar`;
  } else {
    return `${order.portions || 0} porsi`;
  }
};

// Helper: Sleep
const sleep = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Fetch Orders
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

// Get Status Color
const getStatusColor = (status) => {
  const statusLower = String(status).toLowerCase();
  
  if (statusLower.includes('selesai') || statusLower.includes('diterima') || statusLower.includes('tiba')) {
    return 'success';
  }
  if (statusLower.includes('sedang') || statusLower.includes('proses') || statusLower.includes('perjalanan') || statusLower.includes('menuju')) {
    return 'processing';
  }
  if (statusLower.includes('siap') || statusLower.includes('disiapkan')) {
    return 'default';
  }
  
  return 'default';
};

// Get Status Label
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

// Show Order Detail
const showOrderDetail = async (record) => {
  selectedOrder.value = record;
  detailModalVisible.value = true;
  
  // Fetch activity log
  try {
    const response = await api.get(`/activity-tracker/orders/${record.id}/activity`);
    if (response.data.success) {
      orderActivityLog.value = response.data.data;
    }
  } catch (error) {
    console.error('Error fetching activity log:', error);
    message.error('Gagal memuat log aktivitas');
  }
};

// Order Timeline
const orderTimeline = computed(() => {
  if (!selectedOrder.value || !orderActivityLog.value || !Array.isArray(orderActivityLog.value)) return [];
  
  // Backend returns full timeline with all stages
  return orderActivityLog.value.map(stage => ({
    stage: stage.stage,
    status: stage.status,
    label: stage.title,
    description: stage.description,
    completed: stage.is_completed,
    inProgress: !stage.is_completed && stage.stage === selectedOrder.value.rawOrder?.current_stage,
    timestamp: stage.completed_at || stage.started_at,
  }));
});

// Format Timestamp
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

// Lifecycle
onMounted(() => {
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
/* Page Subtitle */
.page-subtitle {
  color: var(--h-text-secondary);
  font-size: var(--h-text-sm);
  margin-top: calc(var(--h-spacing-4) * -1);
  margin-bottom: var(--h-spacing-4);
}

/* Filters Section */
.filters-section {
  padding: var(--h-spacing-4);
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--h-spacing-5);
}

/* Mobile: Stack stats 1 column */
@media (max-width: 767px) {
  .stats-row {
    grid-template-columns: 1fr;
  }
}

/* Menu Cell */
.menu-cell {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-1);
}

.menu-name {
  font-size: var(--h-text-base);
  font-weight: var(--h-font-semibold);
  color: var(--h-text-primary);
}

.menu-school {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  display: flex;
  align-items: center;
}

/* Portions Cell */
.portions-cell {
  display: flex;
  align-items: center;
  font-size: var(--h-text-sm);
  color: var(--h-text-primary);
}

/* Detail Modal */
.detail-modal :deep(.ant-modal-content) {
  border-radius: var(--h-radius-lg);
}

.detail-modal :deep(.ant-modal-header) {
  border-radius: var(--h-radius-lg) var(--h-radius-lg) 0 0;
  padding: var(--h-spacing-5);
}

.detail-modal :deep(.ant-modal-body) {
  padding: var(--h-spacing-5);
}

.order-detail-modal {
  max-height: 600px;
  overflow-y: auto;
}

.order-header h3 {
  font-size: var(--h-text-xl);
  font-weight: var(--h-font-semibold);
  margin-bottom: var(--h-spacing-2);
  color: var(--h-text-primary);
}

.order-meta {
  display: flex;
  gap: var(--h-spacing-4);
  color: var(--h-text-secondary);
  font-size: var(--h-text-sm);
}

.order-meta span {
  display: flex;
  align-items: center;
}

.timeline-container {
  padding: var(--h-spacing-4) 0;
}

.timeline-content {
  padding-bottom: var(--h-spacing-4);
}

.timeline-title {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2);
  margin-bottom: var(--h-spacing-1);
  font-size: var(--h-text-base);
  color: var(--h-text-primary);
}

.timeline-description {
  color: var(--h-text-secondary);
  font-size: var(--h-text-sm);
  margin-bottom: var(--h-spacing-1);
}

.timeline-timestamp {
  color: var(--h-text-light);
  font-size: var(--h-text-xs);
}

.timeline-dot-empty {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid var(--h-border-color);
  background: var(--h-bg-card);
}

/* Dark Mode Support */
.dark .page-subtitle {
  color: var(--h-text-secondary);
}

.dark .menu-name {
  color: var(--h-text-primary);
}

.dark .menu-school {
  color: var(--h-text-secondary);
}

.dark .portions-cell {
  color: var(--h-text-primary);
}

.dark .order-header h3 {
  color: var(--h-text-primary);
}

.dark .order-meta {
  color: var(--h-text-secondary);
}

.dark .timeline-title {
  color: var(--h-text-primary);
}

.dark .timeline-description {
  color: var(--h-text-secondary);
}

.dark .timeline-timestamp {
  color: var(--h-text-light);
}

.dark .timeline-dot-empty {
  border-color: var(--h-border-color);
  background: var(--h-bg-card);
}
</style>
