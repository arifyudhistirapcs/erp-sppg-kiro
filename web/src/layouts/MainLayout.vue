<template>
  <a-layout style="min-height: 100vh">
    <!-- Sidebar -->
    <a-layout-sider
      v-model:collapsed="collapsed"
      :trigger="null"
      collapsible
      :width="240"
      class="sidebar-dark"
    >
      <div class="logo">
        <h2 v-if="!collapsed">ERP SPPG</h2>
        <h2 v-else>ERP</h2>
      </div>
      
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
        :items="menuItems"
        @click="handleMenuClick"
      />
    </a-layout-sider>

    <!-- Main Content -->
    <a-layout>
      <!-- Header -->
      <a-layout-header style="background: #fff; padding: 0 24px; display: flex; justify-content: space-between; align-items: center;">
        <div style="display: flex; align-items: center;">
          <MenuUnfoldOutlined
            v-if="collapsed"
            class="trigger"
            @click="() => (collapsed = !collapsed)"
          />
          <MenuFoldOutlined
            v-else
            class="trigger"
            @click="() => (collapsed = !collapsed)"
          />
          <h3 style="margin: 0 0 0 16px;">{{ pageTitle }}</h3>
        </div>

        <div style="display: flex; align-items: center; gap: 16px;">
          <!-- Notifications -->
          <a-badge :count="unreadCount" :overflow-count="99">
            <a-button type="text" @click="showNotifications">
              <BellOutlined style="font-size: 18px;" />
            </a-button>
          </a-badge>

          <!-- User Menu -->
          <a-dropdown>
            <a-button type="text">
              <UserOutlined style="font-size: 18px; margin-right: 8px;" />
              <span>{{ userName }}</span>
              <DownOutlined style="margin-left: 8px;" />
            </a-button>
            <template #overlay>
              <a-menu>
                <a-menu-item key="profile">
                  <UserOutlined />
                  Profil Saya
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout" @click="handleLogout">
                  <LogoutOutlined />
                  Keluar
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- Content -->
      <a-layout-content style="margin: 24px; padding: 24px; background: #fff; min-height: 280px;">
        <router-view />
      </a-layout-content>

      <!-- Footer -->
      <a-layout-footer style="text-align: center; background: #f0f2f5;">
        Sistem ERP SPPG © {{ currentYear }} - Satuan Pelayanan Pemenuhan Gizi
      </a-layout-footer>
    </a-layout>

    <!-- Notifications Drawer -->
    <a-drawer
      v-model:open="notificationsVisible"
      title="Notifikasi"
      placement="right"
      :width="400"
    >
      <a-list
        :data-source="notifications"
        :loading="loadingNotifications"
      >
        <template #renderItem="{ item }">
          <a-list-item>
            <a-list-item-meta
              :description="item.message"
            >
              <template #title>
                <a @click="handleNotificationClick(item)">{{ item.title }}</a>
              </template>
              <template #avatar>
                <a-badge dot :status="item.isRead ? 'default' : 'processing'">
                  <BellOutlined style="font-size: 20px;" />
                </a-badge>
              </template>
            </a-list-item-meta>
            <template #extra>
              <span style="font-size: 12px; color: #999;">
                {{ formatTime(item.createdAt) }}
              </span>
            </template>
          </a-list-item>
        </template>
      </a-list>
    </a-drawer>
  </a-layout>
</template>

<script setup>
import { ref, computed, onMounted, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { message } from 'ant-design-vue'
import api from '@/services/api'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  DashboardOutlined,
  UserOutlined,
  BellOutlined,
  LogoutOutlined,
  DownOutlined,
  BookOutlined,
  CalendarOutlined,
  DesktopOutlined,
  ShoppingCartOutlined,
  InboxOutlined,
  CarOutlined,
  TeamOutlined,
  DollarOutlined,
  SettingOutlined,
  FileTextOutlined
} from '@ant-design/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const collapsed = ref(false)
const selectedKeys = ref([route.path])
const notificationsVisible = ref(false)
const notifications = ref([])
const loadingNotifications = ref(false)
const unreadCount = ref(0)

const currentYear = new Date().getFullYear()

const userName = computed(() => {
  return authStore.user?.fullName || authStore.user?.email || 'User'
})

const userRole = computed(() => {
  return authStore.user?.role || ''
})

const pageTitle = computed(() => {
  const titles = {
    '/dashboard/kepala-sppg': 'Dashboard Kepala SPPG',
    '/dashboard/kepala-yayasan': 'Dashboard Kepala Yayasan',
    '/recipes': 'Manajemen Resep',
    '/ingredients': 'Manajemen Bahan',
    '/semi-finished': 'Barang Setengah Jadi',
    '/menu-planning': 'Perencanaan Menu',
    '/kds': 'Kitchen Display System',
    '/suppliers': 'Manajemen Supplier',
    '/purchase-orders': 'Purchase Order',
    '/goods-receipts': 'Penerimaan Barang',
    '/inventory': 'Inventori',
    '/schools': 'Data Sekolah',
    '/delivery-tasks': 'Tugas Pengiriman',
    '/ompreng-tracking': 'Pelacakan Ompreng',
    '/employees': 'Data Karyawan',
    '/attendance-report': 'Laporan Absensi',
    '/wifi-config': 'Konfigurasi Wi-Fi',
    '/attendance': 'Absensi',
    '/assets': 'Aset Dapur',
    '/cash-flow': 'Arus Kas',
    '/financial-reports': 'Laporan Keuangan',
    '/audit-trail': 'Audit Trail',
    '/system-config': 'Konfigurasi Sistem'
  }
  return titles[route.path] || 'ERP SPPG'
})

// Define menu items based on role permissions
const getMenuItems = () => {
  const role = userRole.value
  
  const allMenuItems = [
    {
      key: 'dashboard',
      icon: () => h(DashboardOutlined),
      label: 'Dashboard',
      roles: ['kepala_sppg', 'kepala_yayasan', 'akuntan', 'ahli_gizi', 'pengadaan'],
      children: [
        {
          key: '/dashboard/kepala-sppg',
          label: 'Dashboard Kepala SPPG',
          roles: ['kepala_sppg']
        },
        {
          key: '/dashboard/kepala-yayasan',
          label: 'Dashboard Kepala Yayasan',
          roles: ['kepala_yayasan']
        }
      ]
    },
    {
      key: 'recipes',
      icon: () => h(BookOutlined),
      label: 'Resep & Menu',
      roles: ['kepala_sppg', 'ahli_gizi'],
      children: [
        {
          key: '/recipes',
          label: 'Manajemen Resep',
          roles: ['kepala_sppg', 'ahli_gizi']
        },
        {
          key: '/ingredients',
          label: 'Manajemen Bahan',
          roles: ['kepala_sppg', 'ahli_gizi']
        },
        {
          key: '/semi-finished',
          label: 'Barang Setengah Jadi',
          roles: ['kepala_sppg', 'ahli_gizi', 'chef']
        },
        {
          key: '/menu-planning',
          label: 'Perencanaan Menu',
          roles: ['kepala_sppg', 'ahli_gizi']
        }
      ]
    },
    {
      key: 'kds',
      icon: () => h(DesktopOutlined),
      label: 'Kitchen Display',
      roles: ['kepala_sppg', 'ahli_gizi', 'chef', 'packing'],
      children: [
        {
          key: '/kds/cooking',
          label: 'Dapur - Memasak',
          roles: ['kepala_sppg', 'ahli_gizi', 'chef']
        },
        {
          key: '/kds/packing',
          label: 'Packing',
          roles: ['kepala_sppg', 'ahli_gizi', 'chef', 'packing']
        }
      ]
    },
    {
      key: 'supply-chain',
      icon: () => h(ShoppingCartOutlined),
      label: 'Supply Chain',
      roles: ['kepala_sppg', 'pengadaan', 'akuntan'],
      children: [
        {
          key: '/suppliers',
          label: 'Supplier',
          roles: ['kepala_sppg', 'pengadaan']
        },
        {
          key: '/purchase-orders',
          label: 'Purchase Order',
          roles: ['kepala_sppg', 'pengadaan']
        },
        {
          key: '/goods-receipts',
          label: 'Penerimaan Barang',
          roles: ['kepala_sppg', 'pengadaan']
        },
        {
          key: '/inventory',
          label: 'Inventori',
          roles: ['kepala_sppg', 'pengadaan', 'akuntan']
        }
      ]
    },
    {
      key: 'logistics',
      icon: () => h(CarOutlined),
      label: 'Logistik',
      roles: ['kepala_sppg', 'pengadaan'],
      children: [
        {
          key: '/schools',
          label: 'Data Sekolah',
          roles: ['kepala_sppg', 'pengadaan']
        },
        {
          key: '/delivery-tasks',
          label: 'Tugas Pengiriman',
          roles: ['kepala_sppg', 'pengadaan']
        },
        {
          key: '/ompreng-tracking',
          label: 'Pelacakan Ompreng',
          roles: ['kepala_sppg', 'pengadaan']
        }
      ]
    },
    {
      key: 'hrm',
      icon: () => h(TeamOutlined),
      label: 'SDM',
      roles: ['kepala_sppg', 'akuntan'],
      children: [
        {
          key: '/employees',
          label: 'Data Karyawan',
          roles: ['kepala_sppg', 'akuntan']
        },
        {
          key: '/attendance-report',
          label: 'Laporan Absensi',
          roles: ['kepala_sppg', 'akuntan']
        },
        {
          key: '/wifi-config',
          label: 'Konfigurasi Wi-Fi',
          roles: ['kepala_sppg', 'akuntan']
        },
        {
          key: '/attendance',
          label: 'Absensi',
          roles: ['kepala_sppg', 'akuntan']
        }
      ]
    },
    {
      key: 'financial',
      icon: () => h(DollarOutlined),
      label: 'Keuangan',
      roles: ['kepala_sppg', 'kepala_yayasan', 'akuntan'],
      children: [
        {
          key: '/assets',
          label: 'Aset Dapur',
          roles: ['kepala_sppg', 'akuntan']
        },
        {
          key: '/cash-flow',
          label: 'Arus Kas',
          roles: ['kepala_sppg', 'akuntan']
        },
        {
          key: '/financial-reports',
          label: 'Laporan Keuangan',
          roles: ['kepala_sppg', 'kepala_yayasan', 'akuntan']
        }
      ]
    },
    {
      key: 'system',
      icon: () => h(SettingOutlined),
      label: 'Sistem',
      roles: ['kepala_sppg'],
      children: [
        {
          key: '/audit-trail',
          label: 'Audit Trail',
          roles: ['kepala_sppg']
        },
        {
          key: '/system-config',
          label: 'Konfigurasi',
          roles: ['kepala_sppg']
        }
      ]
    }
  ]

  // Filter menu items based on user role
  const filterByRole = (items) => {
    return items
      .filter(item => !item.roles || item.roles.includes(role))
      .map(item => {
        if (item.children) {
          const filteredChildren = filterByRole(item.children)
          if (filteredChildren.length > 0) {
            return { ...item, children: filteredChildren }
          }
          return null
        }
        return item
      })
      .filter(item => item !== null)
  }

  return filterByRole(allMenuItems)
}

const menuItems = computed(() => getMenuItems())

const handleMenuClick = ({ key }) => {
  router.push(key)
}

const handleLogout = async () => {
  try {
    await authStore.logout()
    message.success('Berhasil keluar dari sistem')
    router.push('/login')
  } catch (error) {
    console.error('Logout error:', error)
    message.error('Gagal keluar dari sistem')
  }
}

const showNotifications = () => {
  notificationsVisible.value = true
  loadNotifications()
}

const loadNotifications = async () => {
  loadingNotifications.value = true
  try {
    const response = await api.get('/notifications')
    notifications.value = response.data.data || []
    unreadCount.value = notifications.value.filter(n => !n.is_read).length
  } catch (error) {
    console.error('Failed to load notifications:', error)
  } finally {
    loadingNotifications.value = false
  }
}

const handleNotificationClick = async (notification) => {
  try {
    // Mark notification as read
    if (!notification.is_read) {
      await api.put(`/notifications/${notification.id}/read`)
      notification.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
    
    // Navigate to link if provided
    if (notification.link) {
      router.push(notification.link)
    }
  } catch (error) {
    console.error('Failed to mark notification as read:', error)
  }
  notificationsVisible.value = false
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now - date
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'Baru saja'
  if (minutes < 60) return `${minutes} menit yang lalu`
  if (hours < 24) return `${hours} jam yang lalu`
  if (days < 7) return `${days} hari yang lalu`
  return date.toLocaleDateString('id-ID')
}

onMounted(() => {
  // Update selected keys when route changes
  selectedKeys.value = [route.path]
  
  // Load initial notification count
  loadNotifications()
})
</script>

<style scoped>
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
  margin: 16px;
  border-radius: 4px;
}

.logo h2 {
  color: white;
  margin: 0;
  font-size: 20px;
  font-weight: 600;
}

.sidebar-dark {
  background: linear-gradient(180deg, #1a1a2e 0%, #16213e 100%) !important;
}

.trigger {
  font-size: 18px;
  cursor: pointer;
  transition: color 0.3s;
}

.trigger:hover {
  color: #f82c17;
}
</style>
