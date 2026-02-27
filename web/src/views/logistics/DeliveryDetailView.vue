<template>
  <div class="delivery-detail-view">
    <a-page-header
      title="Detail Pengiriman"
      @back="goBack"
    >
      <template #extra>
        <a-button @click="refreshData" :loading="loading">
          <template #icon><reload-outlined /></template>
          Refresh
        </a-button>
      </template>
    </a-page-header>

    <div class="content-wrapper">
      <a-spin :spinning="loading">
        <!-- School Information Section -->
        <a-card title="Informasi Sekolah" style="margin-bottom: 16px">
          <a-descriptions bordered :column="{ xs: 1, sm: 2, md: 2 }">
            <a-descriptions-item label="Nama Sekolah">
              {{ deliveryDetail?.school?.name || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="Jumlah Porsi">
              {{ deliveryDetail?.portions || 0 }} porsi
            </a-descriptions-item>
            <a-descriptions-item label="Alamat" :span="2">
              {{ deliveryDetail?.school?.address || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="Kontak">
              {{ deliveryDetail?.school?.contact_person || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="Nomor Telepon">
              {{ deliveryDetail?.school?.phone || '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </a-card>

        <!-- Driver Information Section -->
        <a-card title="Informasi Driver" style="margin-bottom: 16px">
          <a-descriptions bordered :column="{ xs: 1, sm: 2, md: 2 }">
            <a-descriptions-item label="Nama Driver">
              {{ deliveryDetail?.driver?.full_name || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="Jenis Kendaraan">
              {{ deliveryDetail?.driver?.vehicle_type || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="Nomor Telepon">
              {{ deliveryDetail?.driver?.phone || '-' }}
            </a-descriptions-item>
            <a-descriptions-item label="Email">
              {{ deliveryDetail?.driver?.email || '-' }}
            </a-descriptions-item>
          </a-descriptions>
        </a-card>

        <!-- Timeline Section -->
        <a-card title="Timeline Pengiriman" style="margin-bottom: 16px">
          <DeliveryTimeline
            v-if="deliveryDetail"
            :current-status="deliveryDetail.current_status"
            :activity-log="activityLog"
          />
        </a-card>

        <!-- Activity Log Section -->
        <a-card title="Log Aktivitas">
          <ActivityLogTable
            v-if="activityLog.length > 0"
            :activity-log="activityLog"
          />
          <a-empty v-else description="Belum ada aktivitas" />
        </a-card>
      </a-spin>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { getDeliveryDetail, getActivityLog } from '@/services/monitoringService'
import DeliveryTimeline from '@/components/DeliveryTimeline.vue'
import ActivityLogTable from '@/components/ActivityLogTable.vue'

const router = useRouter()
const route = useRoute()

// State
const loading = ref(false)
const deliveryDetail = ref(null)
const activityLog = ref([])

// Methods
const goBack = () => {
  router.push('/logistics/monitoring')
}

const refreshData = () => {
  fetchData()
}

const fetchData = async () => {
  await Promise.all([
    fetchDeliveryDetail(),
    fetchActivityLog()
  ])
}

const fetchDeliveryDetail = async () => {
  loading.value = true
  try {
    const id = route.params.id
    const response = await getDeliveryDetail(id)
    
    if (response.success) {
      deliveryDetail.value = response.data
    } else {
      message.error(response.message || 'Gagal memuat detail pengiriman')
    }
  } catch (error) {
    console.error('Error fetching delivery detail:', error)
    if (error.response?.status === 404) {
      message.error('Data pengiriman tidak ditemukan')
      goBack()
    } else {
      message.error(error.response?.data?.message || 'Gagal memuat detail pengiriman')
    }
  } finally {
    loading.value = false
  }
}

const fetchActivityLog = async () => {
  try {
    const id = route.params.id
    const response = await getActivityLog(id)
    
    if (response.success) {
      activityLog.value = response.data || []
    } else {
      message.error(response.message || 'Gagal memuat log aktivitas')
    }
  } catch (error) {
    console.error('Error fetching activity log:', error)
    message.error(error.response?.data?.message || 'Gagal memuat log aktivitas')
  }
}

// Lifecycle
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.delivery-detail-view {
  padding: 24px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.content-wrapper {
  margin-top: 16px;
}

:deep(.ant-descriptions-item-label) {
  font-weight: 600;
  background-color: #fafafa;
}
</style>
