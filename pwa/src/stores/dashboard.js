import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'
import { useAuthStore } from './auth'

export const useDashboardStore = defineStore('dashboard', () => {
  const summary = ref({
    porsiDisiapkan: 0,
    porsiDisiapkanTrend: '',
    deliveryRate: 0,
    deliveryRateTrend: '',
    ketersediaanStok: 0,
    stokKritisTrend: '',
    onTimeDelivery: 0,
    onTimeDeliveryTrend: '',
    ratingKeseluruhan: 0,
    ratingKeseluruhanTrend: '',
    ratingMenu: 0,
    ratingMenuTrend: '',
    ratingLayanan: 0,
    ratingLayananTrend: ''
  })
  const detailProduksi = ref([])
  const detailPengiriman = ref([])
  const detailPencucian = ref([])
  const stokKritis = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchDashboardData() {
    loading.value = true
    error.value = null

    try {
      const authStore = useAuthStore()
      const userRole = authStore.user?.role?.toLowerCase()

      // For Kepala SPPG, use the existing backend endpoint
      if (userRole === 'kepala_sppg') {
        try {
          const response = await api.get('/dashboard/kepala-sppg')
          
          console.log('[Dashboard] Kepala SPPG response:', response.data)
          
          if (response.data.success) {
            // Backend returns data in "dashboard" key, not "data"
            const data = response.data.dashboard || response.data.data
            console.log('[Dashboard] Data structure:', data)
            console.log('[Dashboard] Full response:', JSON.stringify(response.data, null, 2))
            
            // Map summary metrics from TodayKPIs
            const kpis = data?.today_kpis || {}
            summary.value = {
              porsiDisiapkan: kpis?.portions_prepared || 0,
              porsiDisiapkanTrend: '↑ porsi hari ini',
              deliveryRate: kpis?.delivery_rate || 0,
              deliveryRateTrend: '↑ 2 pengiriman hari ini',
              ketersediaanStok: kpis?.stock_availability || 0,
              stokKritisTrend: `↓ ${data?.critical_stock?.length || 0} item kritis`,
              onTimeDelivery: kpis?.on_time_delivery_rate || 0,
              onTimeDeliveryTrend: '↑ 2 pengiriman tepat waktu',
              ratingKeseluruhan: 0, // Not available in current backend
              ratingKeseluruhanTrend: '↑ 2 ulasan',
              ratingMenu: 0, // Not available in current backend
              ratingMenuTrend: '↑ kualitas makanan',
              ratingLayanan: 0, // Not available in current backend
              ratingLayananTrend: '↑ kualitas pengiriman'
            }
            
            // Map detail tables
            detailProduksi.value = (data?.production_details || []).map(item => ({
              id: item.school_id,
              sekolah: item.school_name,
              porsi: item.portions,
              status: item.status
            }))
            
            detailPengiriman.value = (data?.delivery_details || []).map(item => ({
              id: item.school_id,
              sekolah: item.school_name,
              porsi: item.portions,
              status: item.status
            }))
            
            detailPencucian.value = (data?.cleaning_details || []).map(item => ({
              id: item.school_id,
              sekolah: item.school_name,
              porsi: item.portions,
              status: item.status
            }))
            
            stokKritis.value = (data?.critical_stock || []).map(item => ({
              id: item.ingredient_id,
              nama: item.ingredient_name,
              stok: item.current_stock,
              min: item.min_threshold,
              satuan: item.unit
            }))
            
            console.log('[Dashboard] Mapped summary:', summary.value)
            console.log('[Dashboard] Detail Produksi:', detailProduksi.value)
            console.log('[Dashboard] Detail Pengiriman:', detailPengiriman.value)
            console.log('[Dashboard] Detail Pencucian:', detailPencucian.value)
            console.log('[Dashboard] Stok Kritis:', stokKritis.value)
          } else {
            console.warn('[Dashboard] Response not successful')
            resetToEmpty()
          }
        } catch (apiError) {
          console.error('Kepala SPPG dashboard error:', apiError)
          console.error('Error response:', apiError.response?.data)
          resetToEmpty()
          
          // Only show error for non-404 errors
          if (apiError.response?.status !== 404) {
            error.value = apiError.response?.data?.message || 'Gagal memuat data dashboard'
          }
        }
        
        loading.value = false
        return
      }

      // For driver and assistant roles, use local data only (no API calls)
      if (userRole === 'driver' || userRole === 'asisten_lapangan') {
        resetToEmpty()
        loading.value = false
        return
      }

      // For other roles, try generic endpoints (may not exist)
      try {
        const response = await api.get('/dashboard/summary')
        if (response?.data?.success) {
          const data = response.data.data
          summary.value = {
            porsiDisiapkan: data?.porsiDisiapkan ?? 0,
            porsiDisiapkanTrend: data?.porsiDisiapkanTrend ?? '',
            deliveryRate: data?.deliveryRate ?? 0,
            deliveryRateTrend: data?.deliveryRateTrend ?? '',
            ketersediaanStok: data?.ketersediaanStok ?? 0,
            stokKritisTrend: data?.stokKritisTrend ?? '',
            onTimeDelivery: data?.onTimeDelivery ?? 0,
            onTimeDeliveryTrend: data?.onTimeDeliveryTrend ?? '',
            ratingKeseluruhan: data?.ratingKeseluruhan ?? 0,
            ratingKeseluruhanTrend: data?.ratingKeseluruhanTrend ?? '',
            ratingMenu: data?.ratingMenu ?? 0,
            ratingMenuTrend: data?.ratingMenuTrend ?? '',
            ratingLayanan: data?.ratingLayanan ?? 0,
            ratingLayananTrend: data?.ratingLayananTrend ?? ''
          }
          detailProduksi.value = data?.detailProduksi ?? []
          detailPengiriman.value = data?.detailPengiriman ?? []
          detailPencucian.value = data?.detailPencucian ?? []
          stokKritis.value = data?.stokKritis ?? []
        }
      } catch (apiError) {
        console.warn('Dashboard API error (non-critical):', apiError)
      }
    } catch (err) {
      console.error('Dashboard fetch error:', err)
      if (err.message && !err.message.includes('404')) {
        error.value = 'Gagal memuat data dashboard. Silakan coba lagi.'
      }
    } finally {
      loading.value = false
    }
  }

  function resetToEmpty() {
    summary.value = {
      porsiDisiapkan: 0,
      porsiDisiapkanTrend: '',
      deliveryRate: 0,
      deliveryRateTrend: '',
      ketersediaanStok: 0,
      stokKritisTrend: '',
      onTimeDelivery: 0,
      onTimeDeliveryTrend: '',
      ratingKeseluruhan: 0,
      ratingKeseluruhanTrend: '',
      ratingMenu: 0,
      ratingMenuTrend: '',
      ratingLayanan: 0,
      ratingLayananTrend: ''
    }
    detailProduksi.value = []
    detailPengiriman.value = []
    detailPencucian.value = []
    stokKritis.value = []
  }

  function retry() {
    return fetchDashboardData()
  }

  return {
    summary,
    detailProduksi,
    detailPengiriman,
    detailPencucian,
    stokKritis,
    loading,
    error,
    fetchDashboardData,
    retry
  }
})
