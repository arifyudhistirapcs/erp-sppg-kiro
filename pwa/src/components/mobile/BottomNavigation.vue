<template>
  <van-tabbar
    v-model="activeIndex"
    :active-color="activeColor"
    :inactive-color="inactiveColor"
    class="bottom-navigation"
    @change="onTabChange"
  >
    <van-tabbar-item
      v-for="item in navItems"
      :key="item.name"
      :icon="isActive(item) ? item.activeIcon : item.icon"
    >
      {{ item.label }}
    </van-tabbar-item>
  </van-tabbar>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const activeColor = '#5A4372'
const inactiveColor = '#ACA9B0'

const NAV_CONFIG = {
  dashboard: { name: 'dashboard', label: 'Dashboard', icon: 'bar-chart-o', activeIcon: 'bar-chart-o', route: '/dashboard' },
  monitoring: { name: 'monitoring', label: 'Monitoring', icon: 'eye-o', activeIcon: 'eye', route: '/monitoring' },
  menu: { name: 'menu', label: 'Menu', icon: 'notes-o', activeIcon: 'notes-o', route: '/menu-planning' },
  tasks: { name: 'tasks', label: 'Tugas', icon: 'todo-list-o', activeIcon: 'todo-list-o', route: '/tasks' },
  attendance: { name: 'attendance', label: 'Absensi', icon: 'clock-o', activeIcon: 'clock', route: '/attendance' },
  profile: { name: 'profile', label: 'Profil', icon: 'contact-o', activeIcon: 'contact', route: '/profile' },
  schoolMonitoring: { name: 'schoolMonitoring', label: 'Monitoring', icon: 'eye-o', activeIcon: 'eye', route: '/school-monitoring' }
}

const ROLE_NAV_MAP = {
  driver: ['tasks', 'attendance', 'profile'],
  asisten_lapangan: ['tasks', 'attendance', 'profile'],
  kepala_sppg: ['dashboard', 'monitoring', 'menu', 'attendance', 'profile'],
  ahli_gizi: ['menu', 'attendance', 'profile'],
  sekolah: ['schoolMonitoring', 'profile']
}

const DEFAULT_NAV = ['attendance', 'profile']

const navItems = computed(() => {
  const role = authStore.user?.role?.toLowerCase() || ''
  const keys = ROLE_NAV_MAP[role] || DEFAULT_NAV
  return keys.map(key => NAV_CONFIG[key])
})

const activeIndex = ref(0)

function findActiveIndex() {
  const currentPath = route.path
  const idx = navItems.value.findIndex(item => currentPath.startsWith(item.route))
  return idx >= 0 ? idx : 0
}

function isActive(item) {
  const idx = navItems.value.indexOf(item)
  return idx === activeIndex.value
}

function onTabChange(index) {
  const item = navItems.value[index]
  if (item) {
    router.push(item.route)
  }
}

// Sync active index with current route
watch(
  () => route.path,
  () => {
    activeIndex.value = findActiveIndex()
  },
  { immediate: true }
)

onMounted(() => {
  activeIndex.value = findActiveIndex()
})
</script>

<style scoped>
.bottom-navigation {
  height: 64px;
  background: #FFFFFF;
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.08);
  border-radius: 24px 24px 0 0 !important;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
}

.bottom-navigation :deep(.van-tabbar-item__icon) {
  font-size: 22px;
  margin-bottom: 2px;
}

.bottom-navigation :deep(.van-tabbar-item__text) {
  font-size: 11px;
  font-weight: 500;
}

.bottom-navigation :deep(.van-tabbar-item--active .van-tabbar-item__icon) {
  transform: scale(1.1);
  transition: transform 0.2s ease;
}
</style>
