<template>
  <HorizonLayout
    :page-title="pageTitle"
    :breadcrumb="{ parent: 'Pages', current: pageTitle }"
    :notification-count="unreadCount"
    @notification-click="showNotifications"
  >
    <router-view />
  </HorizonLayout>

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
          <a-list-item-meta :description="item.message">
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
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { BellOutlined } from '@ant-design/icons-vue'
import api from '@/services/api'
import HorizonLayout from '@/layouts/HorizonLayout.vue'

const router = useRouter()
const route = useRoute()

const notificationsVisible = ref(false)
const notifications = ref([])
const loadingNotifications = ref(false)
const unreadCount = ref(0)

const pageTitle = computed(() => {
  return route.meta?.title || 'ERP SPPG'
})

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
    if (!notification.is_read) {
      await api.put(`/notifications/${notification.id}/read`)
      notification.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
    }
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
  loadNotifications()
})
</script>
