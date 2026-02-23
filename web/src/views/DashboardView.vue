<template>
  <div>
    <!-- Role-specific dashboard redirect for Kepala SPPG -->
    <div v-if="authStore.user?.role === 'kepala_sppg'">
      <a-spin :spinning="redirecting" tip="Mengarahkan ke dashboard Kepala SPPG...">
        <a-result
          status="info"
          title="Mengarahkan ke Dashboard Kepala SPPG"
          sub-title="Anda akan diarahkan ke dashboard khusus Kepala SPPG dengan monitoring real-time."
        >
          <template #extra>
            <a-button type="primary" @click="goToKepalaSSPGDashboard">
              Buka Dashboard Kepala SPPG
            </a-button>
          </template>
        </a-result>
      </a-spin>
    </div>

    <!-- General dashboard for other roles -->
    <div v-else>
      <a-row :gutter="[16, 16]">
        <a-col :span="24">
          <a-alert
            message="Selamat Datang di Sistem ERP SPPG"
            :description="`Halo, ${userName}! Anda login sebagai ${roleLabel}.`"
            type="info"
            show-icon
            closable
          />
        </a-col>
      </a-row>

      <a-row :gutter="[16, 16]" style="margin-top: 24px;">
        <a-col :xs="24" :sm="12" :lg="6">
          <a-card>
            <a-statistic
              title="Total Resep"
              :value="0"
              :prefix="h(BookOutlined)"
            />
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :lg="6">
          <a-card>
            <a-statistic
              title="Menu Aktif"
              :value="0"
              :prefix="h(CalendarOutlined)"
            />
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :lg="6">
          <a-card>
            <a-statistic
              title="Pengiriman Hari Ini"
              :value="0"
              :prefix="h(CarOutlined)"
            />
          </a-card>
        </a-col>
        <a-col :xs="24" :sm="12" :lg="6">
          <a-card>
            <a-statistic
              title="Stok Menipis"
              :value="0"
              :prefix="h(WarningOutlined)"
              :value-style="{ color: '#cf1322' }"
            />
          </a-card>
        </a-col>
      </a-row>

      <a-row :gutter="[16, 16]" style="margin-top: 24px;">
        <a-col :xs="24" :lg="16">
          <a-card title="Status Produksi Hari Ini">
            <a-empty description="Data akan ditampilkan setelah modul KDS diimplementasikan" />
          </a-card>
        </a-col>
        <a-col :xs="24" :lg="8">
          <a-card title="Aktivitas Terbaru">
            <a-empty description="Data akan ditampilkan setelah modul audit trail diimplementasikan" />
          </a-card>
        </a-col>
      </a-row>
    </div>
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

// Auto-redirect Kepala SPPG users to their specific dashboard
onMounted(() => {
  if (authStore.user?.role === 'kepala_sppg') {
    setTimeout(() => {
      goToKepalaSSPGDashboard()
    }, 2000) // Auto-redirect after 2 seconds
  }
})
</script>

<style scoped>
/* Dashboard specific styles */
</style>
