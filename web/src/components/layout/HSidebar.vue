<template>
  <aside
    class="h-sidebar"
    :class="{ 'collapsed': isCollapsed, 'mobile': isMobile }"
  >
    <!-- Logo Area -->
    <div class="sidebar-logo">
      <div class="logo-content">
        <img
          v-if="!isCollapsed"
          :src="isDark ? '/gizera-dark.png' : '/gizera-light.png'"
          alt="Gizera"
          class="logo-img"
        />
        <img
          v-else
          :src="isDark ? '/gizera-dark.png' : '/gizera-light.png'"
          alt="Gizera"
          class="logo-img-collapsed"
        />
      </div>
      <button
        v-if="!isMobile"
        class="collapse-btn"
        @click="toggleCollapse"
        :aria-label="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      >
        <MenuUnfoldOutlined v-if="isCollapsed" />
        <MenuFoldOutlined v-else />
      </button>
    </div>

    <!-- Menu Items -->
    <nav class="sidebar-nav">
      <template v-for="item in filteredMenuItems" :key="item.key">
        <!-- Menu item dengan submenu -->
        <div v-if="item.children && item.children.length > 0" class="menu-group">
          <div
            class="menu-item"
            :class="{ 'active': isGroupActive(item), 'expanded': expandedGroups.includes(item.key) }"
            @click="toggleGroup(item.key)"
          >
            <component :is="item.icon" class="menu-icon" />
            <span v-if="!isCollapsed" class="menu-label">{{ item.label }}</span>
            <DownOutlined
              v-if="!isCollapsed"
              class="menu-arrow"
              :class="{ 'rotated': expandedGroups.includes(item.key) }"
            />
          </div>
          
          <!-- Submenu -->
          <Transition name="submenu">
            <div
              v-if="!isCollapsed && expandedGroups.includes(item.key)"
              class="submenu"
            >
              <router-link
                v-for="child in item.children"
                :key="child.key"
                :to="child.route"
                class="submenu-item"
                :class="{ 'active': isActive(child.route) }"
              >
                <component :is="child.icon" class="submenu-icon" />
                <span class="submenu-label">{{ child.label }}</span>
              </router-link>
            </div>
          </Transition>
        </div>

        <!-- Menu item tanpa submenu -->
        <router-link
          v-else
          :to="item.route"
          class="menu-item"
          :class="{ 'active': isActive(item.route) }"
        >
          <component :is="item.icon" class="menu-icon" />
          <span v-if="!isCollapsed" class="menu-label">{{ item.label }}</span>
        </router-link>
      </template>
    </nav>

    <!-- Logout Button -->
    <button
      class="logout-button"
      @click="handleLogout"
      aria-label="Keluar"
    >
      <LogoutOutlined class="logout-icon" />
      <span v-if="!isCollapsed" class="logout-label">Keluar</span>
    </button>
  </aside>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePermissions } from '@/composables/usePermissions'
import { useBreakpoint } from '@/composables/useBreakpoint'
import { useDarkMode } from '@/composables/useDarkMode'
import { useAuthStore } from '@/stores/auth'
import { message } from 'ant-design-vue'
import {
  HomeOutlined,
  DashboardOutlined,
  MonitorOutlined,
  AppstoreOutlined,
  ShoppingOutlined,
  CarOutlined,
  TeamOutlined,
  DollarOutlined,
  SettingOutlined,
  CalendarOutlined,
  FileTextOutlined,
  InboxOutlined,
  ShopOutlined,
  ShoppingCartOutlined,
  DatabaseOutlined,
  EnvironmentOutlined,
  UserOutlined,
  ClockCircleOutlined,
  WifiOutlined,
  BankOutlined,
  LineChartOutlined,
  AuditOutlined,
  ControlOutlined,
  DownOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  FireOutlined,
  ClearOutlined
} from '@ant-design/icons-vue'

/**
 * HSidebar Component
 * 
 * Main navigation sidebar untuk ERP SPPG dengan:
 * - Width 280px (expanded) / 80px (collapsed)
 * - Background white (light) / #111C44 (dark)
 * - Logo area 64px height
 * - Menu items 44px height
 * - Icon 20px, font 14px medium
 * - Active state (purple bg + white text)
 * - Hover state (#F4F7FE bg)
 * - Collapsible toggle button
 * - Nested menu support
 * - Role-based menu filtering
 * - Smooth width transition (300ms)
 * - Mobile: hidden by default (used inside MobileDrawer)
 */

const props = defineProps({
  /**
   * Collapsed state (controlled)
   */
  collapsed: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:collapsed'])

const route = useRoute()
const router = useRouter()
const { can, isAnyRole } = usePermissions()
const { isMobile } = useBreakpoint()
const { isDark } = useDarkMode()
const authStore = useAuthStore()

// Local collapsed state
const isCollapsed = ref(props.collapsed)

// Expanded groups tracking
const expandedGroups = ref([])

/**
 * Menu structure sesuai requirements
 */
const menuItems = [
  {
    key: 'dashboard',
    label: 'Dashboard',
    icon: HomeOutlined,
    route: '/dashboard',
    roles: ['kepala_sppg', 'kepala_yayasan', 'akuntan', 'ahli_gizi', 'pengadaan']
  },
  {
    key: 'monitoring',
    label: 'Monitoring Aktivitas',
    icon: MonitorOutlined,
    route: '/logistics/monitoring',
    roles: ['kepala_sppg', 'kepala_yayasan', 'akuntan', 'ahli_gizi', 'pengadaan', 'chef', 'packing', 'driver', 'asisten_lapangan']
  },
  {
    key: 'display',
    label: 'Display / KDS',
    icon: DashboardOutlined,
    roles: ['kepala_sppg', 'ahli_gizi', 'chef', 'packing', 'kebersihan'],
    children: [
      {
        key: 'kds-cooking',
        label: 'Dapur',
        icon: FireOutlined,
        route: '/kds/cooking',
        roles: ['kepala_sppg', 'ahli_gizi', 'chef']
      },
      {
        key: 'kds-packing',
        label: 'Pengemasan',
        icon: InboxOutlined,
        route: '/kds/packing',
        roles: ['kepala_sppg', 'ahli_gizi', 'chef', 'packing']
      },
      {
        key: 'kds-cleaning',
        label: 'Kebersihan',
        icon: ClearOutlined,
        route: '/kds/cleaning',
        roles: ['kepala_sppg', 'kebersihan']
      }
    ]
  },
  {
    key: 'menu',
    label: 'Menu & Komponen',
    icon: FileTextOutlined,
    roles: ['kepala_sppg', 'ahli_gizi', 'chef'],
    children: [
      {
        key: 'menu-planning',
        label: 'Perencanaan Menu',
        icon: CalendarOutlined,
        route: '/menu-planning',
        roles: ['kepala_sppg', 'ahli_gizi']
      },
      {
        key: 'recipes',
        label: 'Manajemen Menu',
        icon: FileTextOutlined,
        route: '/recipes',
        roles: ['kepala_sppg', 'ahli_gizi']
      },
      {
        key: 'semi-finished',
        label: 'Manajemen Komponen',
        icon: InboxOutlined,
        route: '/semi-finished',
        roles: ['kepala_sppg', 'ahli_gizi', 'chef']
      }
    ]
  },
  {
    key: 'supply-chain',
    label: 'Supply Chain',
    icon: ShoppingOutlined,
    roles: ['kepala_sppg', 'pengadaan'],
    children: [
      {
        key: 'suppliers',
        label: 'Supplier',
        icon: ShopOutlined,
        route: '/suppliers',
        roles: ['kepala_sppg', 'pengadaan']
      },
      {
        key: 'purchase-orders',
        label: 'Purchase Order',
        icon: ShoppingCartOutlined,
        route: '/purchase-orders',
        roles: ['kepala_sppg', 'pengadaan']
      },
      {
        key: 'goods-receipts',
        label: 'Penerimaan Barang',
        icon: InboxOutlined,
        route: '/goods-receipts',
        roles: ['kepala_sppg', 'pengadaan']
      },
      {
        key: 'inventory',
        label: 'Manajemen Bahan Baku',
        icon: DatabaseOutlined,
        route: '/inventory',
        roles: ['kepala_sppg', 'pengadaan', 'akuntan']
      }
    ]
  },
  {
    key: 'logistics',
    label: 'Logistik',
    icon: CarOutlined,
    roles: ['kepala_sppg', 'driver', 'asisten'],
    children: [
      {
        key: 'schools',
        label: 'Data Sekolah',
        icon: EnvironmentOutlined,
        route: '/schools',
        roles: ['kepala_sppg', 'driver', 'asisten']
      },
      {
        key: 'delivery-tasks',
        label: 'Tugas Pengiriman & Pengambilan',
        icon: CarOutlined,
        route: '/delivery-tasks',
        roles: ['kepala_sppg', 'driver', 'asisten']
      }
    ]
  },
  {
    key: 'sdm',
    label: 'SDM',
    icon: TeamOutlined,
    roles: ['kepala_sppg', 'akuntan'],
    children: [
      {
        key: 'employees',
        label: 'Data Karyawan',
        icon: UserOutlined,
        route: '/employees',
        roles: ['kepala_sppg', 'akuntan']
      },
      {
        key: 'attendance-report',
        label: 'Laporan Absensi',
        icon: ClockCircleOutlined,
        route: '/attendance-report',
        roles: ['kepala_sppg', 'akuntan']
      },
      {
        key: 'wifi-config',
        label: 'Konfigurasi Wi-Fi',
        icon: WifiOutlined,
        route: '/wifi-config',
        roles: ['kepala_sppg', 'akuntan']
      }
    ]
  },
  {
    key: 'finance',
    label: 'Keuangan',
    icon: DollarOutlined,
    roles: ['kepala_sppg', 'akuntan'],
    children: [
      {
        key: 'assets',
        label: 'Aset Dapur',
        icon: BankOutlined,
        route: '/assets',
        roles: ['kepala_sppg', 'akuntan']
      },
      {
        key: 'cash-flow',
        label: 'Arus Kas',
        icon: LineChartOutlined,
        route: '/cash-flow',
        roles: ['kepala_sppg', 'akuntan']
      },
      {
        key: 'financial-reports',
        label: 'Laporan Keuangan',
        icon: FileTextOutlined,
        route: '/financial-reports',
        roles: ['kepala_sppg', 'akuntan']
      }
    ]
  },
  {
    key: 'system',
    label: 'Sistem',
    icon: SettingOutlined,
    roles: ['kepala_sppg'],
    children: [
      {
        key: 'audit-trail',
        label: 'Audit Trail',
        icon: AuditOutlined,
        route: '/audit-trail',
        roles: ['kepala_sppg']
      },
      {
        key: 'system-config',
        label: 'Konfigurasi',
        icon: ControlOutlined,
        route: '/system-config',
        roles: ['kepala_sppg']
      }
    ]
  }
]

/**
 * Filter menu items berdasarkan role user
 */
const filteredMenuItems = computed(() => {
  return menuItems.filter(item => {
    // Check if user has required role
    if (item.roles && !isAnyRole(item.roles)) {
      return false
    }

    // Filter children jika ada
    if (item.children) {
      item.children = item.children.filter(child => {
        return !child.roles || isAnyRole(child.roles)
      })
      
      // Hide parent jika tidak ada children yang visible
      return item.children.length > 0
    }

    return true
  })
})

/**
 * Check if route is active
 */
const isActive = (routePath) => {
  if (!routePath) return false
  return route.path === routePath || route.path.startsWith(routePath + '/')
}

/**
 * Check if any child in group is active
 */
const isGroupActive = (item) => {
  if (!item.children) return false
  return item.children.some(child => isActive(child.route))
}

/**
 * Toggle group expansion
 */
const toggleGroup = (key) => {
  if (isCollapsed.value) {
    // Jika collapsed, expand sidebar dulu
    toggleCollapse()
  }
  
  const index = expandedGroups.value.indexOf(key)
  if (index > -1) {
    expandedGroups.value.splice(index, 1)
  } else {
    expandedGroups.value.push(key)
  }
}

/**
 * Toggle sidebar collapse
 */
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
  emit('update:collapsed', isCollapsed.value)
  if (isCollapsed.value) {
    expandedGroups.value = []
  }
}

/**
 * Handle logout
 */
const handleLogout = async () => {
  try {
    await authStore.logout()
    message.success('Berhasil keluar')
    router.push('/login')
  } catch (error) {
    message.error('Gagal keluar')
  }
}

/**
 * Auto-expand active group
 */
watch(() => route.path, () => {
  // Find and expand group containing active route
  menuItems.forEach(item => {
    if (item.children && isGroupActive(item)) {
      if (!expandedGroups.value.includes(item.key)) {
        expandedGroups.value.push(item.key)
      }
    }
  })
}, { immediate: true })

/**
 * Watch props.collapsed changes
 */
watch(() => props.collapsed, (newValue) => {
  isCollapsed.value = newValue
})
</script>

<style scoped>
.h-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  background-color: var(--h-bg-secondary, #FFFFFF);
  border-right: 1px solid var(--h-border-color, #E9EDF7);
  display: flex;
  flex-direction: column;
  transition: width 300ms ease;
  z-index: 100;
  overflow: hidden;
}

/* Dark mode */
.dark .h-sidebar {
  background-color: #111C44;
  border-right-color: #1B254B;
}

/* Collapsed state */
.h-sidebar.collapsed {
  width: 80px;
}

/* Mobile: hidden by default (will be used inside MobileDrawer) */
.h-sidebar.mobile {
  position: static;
  width: 280px;
  border-right: none;
}

/* Logo Area - 64px height */
.sidebar-logo {
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px 10px 16px;
  border-bottom: 1px solid var(--h-border-color, #E9EDF7);
}

.dark .sidebar-logo {
  border-bottom-color: #1B254B;
}

.collapsed .sidebar-logo {
  justify-content: center;
  padding: 10px 8px;
  gap: 0;
}

.collapse-btn {
  width: 28px;
  height: 28px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  color: var(--h-text-secondary, #74788C);
  border-radius: 6px;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.collapse-btn:hover {
  background-color: #F4F7FE;
  color: var(--h-text-primary, #322837);
}

.dark .collapse-btn {
  color: #ACA9B0;
}

.dark .collapse-btn:hover {
  background-color: #1B254B;
  color: #F8FDEA;
}

.collapse-btn:active {
  transform: scale(0.9);
}

.logo-content {
  display: flex;
  align-items: center;
  min-width: 0;
  flex: 1;
  height: 100%;
}

.logo-img {
  height: 200px;
  max-height: 100px;
  width: 300px;
  object-fit: contain;
}

.logo-img-collapsed {
  height: 60px;
  width: auto;
  object-fit: contain;
}

/* Navigation */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 16px 0;
  
  /* Smooth scrolling */
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
  scrollbar-color: rgba(90, 67, 114, 0.3) transparent;
}

.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background-color: rgba(90, 67, 114, 0.3);
  border-radius: 2px;
}

/* Menu Group */
.menu-group {
  margin-bottom: 4px;
}

/* Menu Item - 44px height */
.menu-item {
  height: 44px;
  display: flex;
  align-items: center;
  padding: 0 20px;
  margin: 0 12px;
  border-radius: 12px;
  cursor: pointer;
  text-decoration: none;
  color: var(--h-text-secondary, #74788C);
  transition: all 0.2s ease;
  position: relative;
}

.menu-item:hover {
  background-color: #F4F7FE;
  color: var(--h-text-primary, #322837);
}

.dark .menu-item:hover {
  background-color: #1B254B;
  color: #F8FDEA;
}

/* Active state - purple bg + white text */
.menu-item.active {
  background-color: #5A4372;
  color: #FFFFFF;
}

.dark .menu-item.active {
  background-color: #5A4372;
  color: #FFFFFF;
}

/* Menu Icon - 20px */
.menu-icon {
  font-size: 20px;
  min-width: 20px;
  margin-right: 12px;
  transition: margin 300ms ease;
}

.collapsed .menu-icon {
  margin-right: 0;
}

/* Menu Label - 14px medium */
.menu-label {
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

/* Menu Arrow */
.menu-arrow {
  font-size: 12px;
  margin-left: auto;
  transition: transform 0.3s ease;
}

.menu-arrow.rotated {
  transform: rotate(180deg);
}

/* Submenu */
.submenu {
  padding-left: 8px;
  overflow: hidden;
}

.submenu-item {
  height: 40px;
  display: flex;
  align-items: center;
  padding: 0 16px 0 44px;
  margin: 2px 12px;
  border-radius: 12px;
  cursor: pointer;
  text-decoration: none;
  color: var(--h-text-secondary, #74788C);
  transition: all 0.2s ease;
}

.submenu-item:hover {
  background-color: #F4F7FE;
  color: var(--h-text-primary, #322837);
}

.dark .submenu-item:hover {
  background-color: #1B254B;
  color: #F8FDEA;
}

.submenu-item.active {
  background-color: #5A4372;
  color: #FFFFFF;
}

.dark .submenu-item.active {
  background-color: #5A4372;
  color: #FFFFFF;
}

.submenu-icon {
  font-size: 16px;
  min-width: 16px;
  margin-right: 12px;
}

.submenu-label {
  font-size: 14px;
  font-weight: 400;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Submenu transition */
.submenu-enter-active,
.submenu-leave-active {
  transition: all 0.3s ease;
}

.submenu-enter-from,
.submenu-leave-to {
  opacity: 0;
  max-height: 0;
}

.submenu-enter-to,
.submenu-leave-from {
  opacity: 1;
  max-height: 500px;
}

/* Logout Button */
.logout-button {
  height: 44px;
  margin: 12px;
  border: none;
  border-radius: 12px;
  background-color: transparent;
  color: var(--h-text-secondary, #74788C);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.logout-button:hover {
  background-color: rgba(238, 93, 80, 0.1);
  color: #EE5D50;
}

.dark .logout-button {
  color: #ACA9B0;
}

.dark .logout-button:hover {
  background-color: rgba(238, 93, 80, 0.15);
  color: #EE5D50;
}

.logout-button:active {
  transform: scale(0.95);
}

.logout-icon {
  font-size: 20px;
}

.logout-label {
  white-space: nowrap;
}
</style>
