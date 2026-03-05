<template>
  <div class="profile-view">
    <!-- Header with gradient background -->
    <div class="profile-header">
      <div class="header-bg">
        <div class="header-pattern"></div>
      </div>
      
      <!-- Avatar Section -->
      <div class="avatar-section">
        <div class="avatar-container">
          <div class="avatar-ring">
            <div class="avatar">
              <span class="avatar-text">{{ userInitials }}</span>
            </div>
          </div>
          <div class="online-indicator"></div>
        </div>
        
        <h1 class="user-name">{{ user?.name || 'User' }}</h1>
        <div class="user-role-badge">
          <van-icon name="shield-o" size="12" />
          <span>{{ formatRole(user?.role) }}</span>
        </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="stats-section">
      <div class="stat-card">
        <div class="stat-icon stat-icon--primary">
          <van-icon name="todo-list-o" size="20" />
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ todayTasksCount }}</span>
          <span class="stat-label">Tugas Hari Ini</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon--success">
          <van-icon name="passed" size="20" />
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ completedTasksCount }}</span>
          <span class="stat-label">Selesai</span>
        </div>
      </div>
    </div>

    <!-- Profile Content -->
    <div class="profile-content">
      <!-- Info Section -->
      <div class="section-card">
        <div class="section-header">
          <van-icon name="contact" class="section-icon" />
          <span class="section-title">Informasi Pribadi</span>
        </div>
        
        <div class="info-list">
          <div class="info-item">
            <div class="info-icon">
              <van-icon name="idcard" />
            </div>
            <div class="info-content">
              <span class="info-label">NIK</span>
              <span class="info-value">{{ user?.nik || '-' }}</span>
            </div>
          </div>
          
          <div class="info-item">
            <div class="info-icon">
              <van-icon name="user-o" />
            </div>
            <div class="info-content">
              <span class="info-label">Nama Lengkap</span>
              <span class="info-value">{{ user?.name || '-' }}</span>
            </div>
          </div>
          
          <div class="info-item">
            <div class="info-icon">
              <van-icon name="envelop-o" />
            </div>
            <div class="info-content">
              <span class="info-label">Email</span>
              <span class="info-value">{{ user?.email || '-' }}</span>
            </div>
          </div>
          
          <div class="info-item">
            <div class="info-icon">
              <van-icon name="shield-o" />
            </div>
            <div class="info-content">
              <span class="info-label">Role</span>
              <span class="info-value">{{ formatRole(user?.role) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Settings Section -->
      <div class="section-card">
        <div class="section-header">
          <van-icon name="setting-o" class="section-icon" />
          <span class="section-title">Pengaturan</span>
        </div>
        
        <div class="settings-list">
          <div class="setting-item" @click="showOfflineInfo">
            <div class="setting-left">
              <div class="setting-icon setting-icon--blue">
                <van-icon name="wifi-o" />
              </div>
              <div class="setting-content">
                <span class="setting-label">Mode Offline</span>
                <span class="setting-desc">Data tersimpan lokal</span>
              </div>
            </div>
            <van-icon name="arrow" class="setting-arrow" />
          </div>
          
          <div class="setting-item">
            <div class="setting-left">
              <div class="setting-icon setting-icon--purple">
                <van-icon name="info-o" />
              </div>
              <div class="setting-content">
                <span class="setting-label">Versi Aplikasi</span>
                <span class="setting-desc">v1.0.0</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Logout Button -->
      <button 
        class="logout-button" 
        :disabled="loading"
        @click="handleLogout"
      >
        <van-icon v-if="!loading" name="cross" size="18" />
        <van-loading v-else size="18" color="#fff" />
        <span>{{ loading ? 'Logging out...' : 'Logout' }}</span>
      </button>

      <!-- App Info -->
      <div class="app-info">
        <span>ERP SPPG Driver App</span>
        <span class="app-version">© 2026 GIZERA</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useDeliveryTasksStore } from '@/stores/deliveryTasks'
import { showDialog, showNotify } from 'vant'

const router = useRouter()
const authStore = useAuthStore()
const deliveryTasksStore = useDeliveryTasksStore()

const loading = ref(false)

const user = computed(() => authStore.user)

const userInitials = computed(() => {
  const name = user.value?.name || 'U'
  const parts = name.split(' ')
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }
  return name.substring(0, 2).toUpperCase()
})

const todayTasksCount = computed(() => {
  return deliveryTasksStore.tasks?.length || 0
})

const completedTasksCount = computed(() => {
  return deliveryTasksStore.tasks?.filter(t => 
    t.status === 'completed' || t.status === 'received' || t.current_stage >= 4
  ).length || 0
})

const formatRole = (role) => {
  if (!role) return '-'
  const roleMap = {
    'driver': 'Driver',
    'asisten_lapangan': 'Asisten Lapangan',
    'kepala_sppg': 'Kepala SPPG',
    'admin': 'Administrator'
  }
  return roleMap[role.toLowerCase()] || role
}

const handleLogout = async () => {
  showDialog({
    title: 'Konfirmasi Logout',
    message: 'Apakah Anda yakin ingin keluar dari aplikasi?',
    showCancelButton: true,
    confirmButtonText: 'Ya, Keluar',
    cancelButtonText: 'Batal',
    confirmButtonColor: '#EE5D50'
  }).then(async () => {
    loading.value = true
    
    try {
      await authStore.logout()
      
      showNotify({
        type: 'success',
        message: 'Logout berhasil'
      })
      
      router.push('/login')
    } catch (error) {
      console.error('Logout error:', error)
      showNotify({
        type: 'warning',
        message: 'Logout berhasil (offline mode)'
      })
      router.push('/login')
    } finally {
      loading.value = false
    }
  }).catch(() => {
    // User cancelled
  })
}

const showOfflineInfo = () => {
  showDialog({
    title: 'Mode Offline',
    message: 'Aplikasi ini dapat bekerja secara offline. Data akan disinkronkan otomatis saat koneksi internet tersedia.',
    confirmButtonText: 'Mengerti',
    confirmButtonColor: '#5A4372'
  })
}
</script>

<style scoped>
.profile-view {
  min-height: 100vh;
  background: var(--h-bg-primary);
  padding-bottom: 100px;
}

/* Header */
.profile-header {
  position: relative;
  padding-bottom: 20px;
}

.header-bg {
  height: 160px;
  background: linear-gradient(135deg, #5A4372 0%, #7B5E99 50%, #9D7EC2 100%);
  position: relative;
  overflow: hidden;
  border-radius: 0 0 32px 32px;
}

.header-pattern {
  position: absolute;
  inset: 0;
  background-image: 
    radial-gradient(circle at 20% 80%, rgba(255,255,255,0.1) 0%, transparent 50%),
    radial-gradient(circle at 80% 20%, rgba(255,255,255,0.08) 0%, transparent 40%),
    radial-gradient(circle at 40% 40%, rgba(255,255,255,0.05) 0%, transparent 30%);
}

/* Avatar Section */
.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: -50px;
  position: relative;
  z-index: 10;
}

.avatar-container {
  position: relative;
}

.avatar-ring {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  padding: 4px;
  background: linear-gradient(135deg, #5A4372, #9D7EC2);
  box-shadow: 0 8px 24px rgba(90, 67, 114, 0.3);
}

.avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: linear-gradient(135deg, #F8F6FA 0%, #FFFFFF 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 3px solid #FFFFFF;
}

.avatar-text {
  font-size: 32px;
  font-weight: 700;
  color: #5A4372;
  font-family: var(--h-font-family);
}

.online-indicator {
  position: absolute;
  bottom: 8px;
  right: 8px;
  width: 20px;
  height: 20px;
  background: #05CD99;
  border-radius: 50%;
  border: 3px solid #FFFFFF;
  box-shadow: 0 2px 8px rgba(5, 205, 153, 0.4);
}

.user-name {
  margin: 16px 0 8px;
  font-size: 24px;
  font-weight: 700;
  color: var(--h-text-primary);
  font-family: var(--h-font-family);
}

.user-role-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  background: linear-gradient(135deg, rgba(90, 67, 114, 0.1), rgba(157, 126, 194, 0.1));
  border-radius: 20px;
  color: #5A4372;
  font-size: 13px;
  font-weight: 600;
}

/* Stats Section */
.stats-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  padding: 20px 16px 0;
}

.stat-card {
  background: #FFFFFF;
  border-radius: 16px;
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon--primary {
  background: linear-gradient(135deg, rgba(90, 67, 114, 0.15), rgba(90, 67, 114, 0.05));
  color: #5A4372;
}

.stat-icon--success {
  background: linear-gradient(135deg, rgba(5, 205, 153, 0.15), rgba(5, 205, 153, 0.05));
  color: #05CD99;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--h-text-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 12px;
  color: var(--h-text-secondary);
  font-weight: 500;
}

/* Profile Content */
.profile-content {
  padding: 20px 16px;
}

/* Section Card */
.section-card {
  background: #FFFFFF;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #F0EDF3;
}

.section-icon {
  font-size: 20px;
  color: #5A4372;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--h-text-primary);
}

/* Info List */
.info-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 14px;
}

.info-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #F8F6FA;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #5A4372;
  flex-shrink: 0;
}

.info-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-label {
  font-size: 12px;
  color: var(--h-text-secondary);
  font-weight: 500;
}

.info-value {
  font-size: 15px;
  color: var(--h-text-primary);
  font-weight: 600;
}

/* Settings List */
.settings-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  background: #F8F6FA;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.setting-item:active {
  transform: scale(0.98);
  background: #F0EDF3;
}

.setting-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.setting-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.setting-icon--blue {
  background: linear-gradient(135deg, #4481EB, #04BEFE);
  color: #FFFFFF;
}

.setting-icon--purple {
  background: linear-gradient(135deg, #5A4372, #9D7EC2);
  color: #FFFFFF;
}

.setting-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-label {
  font-size: 14px;
  color: var(--h-text-primary);
  font-weight: 600;
}

.setting-desc {
  font-size: 12px;
  color: var(--h-text-secondary);
}

.setting-arrow {
  color: var(--h-text-light);
  font-size: 14px;
}

/* Logout Button */
.logout-button {
  width: 100%;
  height: 52px;
  border: none;
  border-radius: 14px;
  background: linear-gradient(135deg, #EE5D50 0%, #F87171 100%);
  color: #FFFFFF;
  font-size: 16px;
  font-weight: 600;
  font-family: var(--h-font-family);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(238, 93, 80, 0.3);
  transition: all 0.2s ease;
  margin-top: 8px;
}

.logout-button:active:not(:disabled) {
  transform: scale(0.98);
  box-shadow: 0 2px 8px rgba(238, 93, 80, 0.3);
}

.logout-button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* App Info */
.app-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  margin-top: 24px;
  padding-top: 16px;
}

.app-info span {
  font-size: 12px;
  color: var(--h-text-light);
}

.app-version {
  font-size: 11px !important;
  opacity: 0.7;
}

/* Responsive */
@media (max-width: 360px) {
  .avatar-ring {
    width: 88px;
    height: 88px;
  }
  
  .avatar-text {
    font-size: 28px;
  }
  
  .user-name {
    font-size: 20px;
  }
  
  .stat-card {
    padding: 12px;
  }
  
  .stat-value {
    font-size: 18px;
  }
}
</style>
