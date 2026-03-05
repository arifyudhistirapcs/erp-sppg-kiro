import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'

export const useMonitoringStore = defineStore('monitoring', () => {
  const activities = ref([])
  const selectedDate = ref(new Date())
  const filterType = ref('all') // 'all' | 'attendance' | 'delivery' | 'pickup'
  const loading = ref(false)
  const hasMore = ref(true)
  const page = ref(1)
  const error = ref(null)

  const PAGE_LIMIT = 20

  function _formatDate(date) {
    const d = date instanceof Date ? date : new Date(date)
    
    // Use local date components to avoid timezone issues
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    const formatted = `${year}-${month}-${day}`
    
    console.log('[Monitoring] Formatting date:', date, '→', formatted)
    return formatted
  }

  async function fetchActivities() {
    loading.value = true
    error.value = null
    page.value = 1
    hasMore.value = true

    try {
      const dateParam = _formatDate(selectedDate.value)
      console.log('[Monitoring] Fetching activities for date:', dateParam)
      
      const params = {
        date: dateParam
      }

      const res = await api.get('/activity-tracker/orders', { params })
      
      console.log('[Monitoring] Response:', res.data)

      if (res.data.success) {
        // Map orders to activities format
        const orders = res.data.data?.orders ?? []
        
        activities.value = orders.map(order => ({
          id: order.id,
          employeeName: order.school?.name || 'Sekolah',
          activityType: 'delivery',
          timestamp: formatTimestamp(order.updated_at || order.created_at),
          status: String(order.current_status || 'Pending'),
          portions: order.portions || 0,
          menuName: order.menu?.name || order.menu_name || null
        }))
        
        console.log('[Monitoring] Mapped activities count:', activities.value.length)
        hasMore.value = false
      }
    } catch (err) {
      console.error('Monitoring error:', err)
      error.value = err.response?.data?.message || 'Gagal memuat data monitoring. Silakan coba lagi.'
    } finally {
      loading.value = false
    }
  }

  function formatTimestamp(timestamp) {
    if (!timestamp) return '-'
    const date = new Date(timestamp)
    return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
  }

  async function loadMore() {
    // Disabled for now - no pagination
    return
  }

  function setFilter(type) {
    filterType.value = type
    return fetchActivities()
  }

  function setDate(date) {
    selectedDate.value = date
    return fetchActivities()
  }

  function retry() {
    return fetchActivities()
  }

  return {
    activities,
    selectedDate,
    filterType,
    loading,
    hasMore,
    page,
    error,
    fetchActivities,
    loadMore,
    setFilter,
    setDate,
    retry
  }
})
