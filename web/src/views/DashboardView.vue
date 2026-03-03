<template>
  <div>
    <!-- Role-specific dashboard redirect for Kepala SPPG -->
    <div v-if="authStore.user?.role === 'kepala_sppg'" class="redirect-container">
      <a-spin :spinning="redirecting" tip="Mengarahkan ke dashboard Kepala SPPG...">
        <a-result
          status="info"
          title="Mengarahkan ke Dashboard Kepala SPPG"
          sub-title="Anda akan diarahkan ke dashboard khusus Kepala SPPG dengan monitoring real-time."
        >
          <template #extra>
            <a-button type="primary" @click="goToKepalaSSPGDashboard" class="h-button">
              Buka Dashboard Kepala SPPG
            </a-button>
          </template>
        </a-result>
      </a-spin>
    </div>

    <!-- General dashboard for other roles -->
    <template v-else>
      <div class="h-card welcome-card">
        <h2 class="welcome-title">Selamat Datang, {{ userName }}!</h2>
        <p class="welcome-subtitle">Anda login sebagai {{ roleLabel }}</p>
      </div>

      <div class="stats-row">
        <HStatCard
          :icon="BookOutlined"
          icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
          label="Total Resep"
          value="0"
          :loading="false"
        />
        <HStatCard
          :icon="CalendarOutlined"
          icon-bg="linear-gradient(135deg, #05CD99 0%, #04b587 100%)"
          label="Menu Aktif"
          value="0"
          :loading="false"
        />
        <HStatCard
          :icon="CarOutlined"
          icon-bg="linear-gradient(135deg, #FFB547 0%, #ff9f1a 100%)"
          label="Pengiriman Hari Ini"
          value="0"
          :loading="false"
        />
        <HStatCard
          :icon="WarningOutlined"
          icon-bg="linear-gradient(135deg, #EE5D50 0%, #e84438 100%)"
          label="Stok Menipis"
          value="0"
          :loading="false"
        />
      </div>

      <div class="content-row">
        <div class="h-card content-card">
          <h3 class="card-title">Status Produksi Hari Ini</h3>
          <a-empty description="Data akan ditampilkan setelah modul KDS diimplementasikan" />
        </div>
        <div class="h-card content-card">
          <h3 class="card-title">Aktivitas Terbaru</h3>
          <a-empty description="Data akan ditampilkan setelah modul audit trail diimplementasikan" />
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { h, computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  BookOutlined,
  CalendarOutlined,
  CarOutlined,
  WarningOutlined
} from '@ant-design/icons-vue'
import HStatCard from '@/components/horizon/HStatCard.vue'

const router = useRouter()
const authStore = useAuthStore()
const redirecting = ref(false)

const userName = computed(() => {
  return authStore.user?.fullName || authStore.user?.email || 'User'
})

const roleLabel = computed(() => {
  const roleLabels = {
    'kepala_sppg': 'Kepala SPPG',
    'kepala_yayasan': 'Kepala Yayasan',
    'akuntan': 'Akuntan',
    'ahli_gizi': 'Ahli Gizi',
    'pengadaan': 'Staff Pengadaan',
    'chef': 'Chef',
    'packing': 'Staff Packing',
    'driver': 'Driver',
    'asisten': 'Asisten Lapangan'
  }
  return roleLabels[authStore.user?.role] || 'User'
})

const goToKepalaSSPGDashboard = () => {
  redirecting.value = true
  router.push('/dashboard/kepala-sppg')
}

onMounted(() => {
  if (authStore.user?.role === 'kepala_sppg') {
    setTimeout(() => {
      goToKepalaSSPGDashboard()
    }, 2000)
  }
})
</script>

<style scoped>
.redirect-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.welcome-card {
  text-align: center;
  padding: var(--h-spacing-8, 32px);
}

.welcome-title {
  font-size: var(--h-text-2xl, 24px);
  font-weight: var(--h-font-bold, 700);
  color: var(--h-text-primary, #322837);
  margin: 0 0 var(--h-spacing-2, 8px) 0;
}

.welcome-subtitle {
  font-size: var(--h-text-sm, 14px);
  color: var(--h-text-secondary, #74788C);
  margin: 0;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
}

@media (max-width: 1024px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-row {
    grid-template-columns: 1fr;
    gap: 12px;
  }
}

.content-row {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 20px;
}

@media (max-width: 1024px) {
  .content-row {
    grid-template-columns: 1fr;
  }
}

.content-card {
  padding: var(--h-spacing-6, 24px);
}

.card-title {
  font-size: var(--h-text-lg, 18px);
  font-weight: var(--h-font-bold, 700);
  color: var(--h-text-primary, #322837);
  margin: 0 0 var(--h-spacing-4, 16px) 0;
}

.h-button {
  background: linear-gradient(135deg, #5A4372 0%, #3D2B53 100%);
  border: none;
  border-radius: var(--h-radius-md, 12px);
  height: 44px;
  font-weight: var(--h-font-semibold, 600);
}
</style>
