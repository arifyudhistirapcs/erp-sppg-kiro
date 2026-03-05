import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'

export const useMenuPlanningStore = defineStore('menuPlanning', () => {
  const weeklyPlans = ref([])       // Array of { id, weekPeriod, approvalStatus, menuCount, menus: [...] }
  const selectedWeek = ref(null)    // Currently selected week for filtering (e.g. '2025-W02')
  const loading = ref(false)
  const approvalLoading = ref(false) // Separate loading for approve/reject actions
  const error = ref(null)

  async function fetchMenuPlans() {
    loading.value = true
    error.value = null

    try {
      const params = {}
      if (selectedWeek.value) {
        params.week = selectedWeek.value
      }

      console.log('[MenuPlanning] Fetching with params:', params)
      const res = await api.get('/menu-plans', { params })
      console.log('[MenuPlanning] Response:', res.data)

      if (res.data.success) {
        // The response has menu_plans array, not data
        const plans = res.data.menu_plans ?? []
        console.log('[MenuPlanning] Plans data:', plans)
        
        weeklyPlans.value = plans.map(plan => {
          console.log('[MenuPlanning] Processing plan:', plan)
          
          // Get week start date to format period
          const weekStart = plan.week_start ? new Date(plan.week_start) : null
          const weekEnd = plan.week_end ? new Date(plan.week_end) : null
          
          // Calculate week number and year from week_start
          let weekNumber = 1
          let year = new Date().getFullYear()
          if (weekStart) {
            weekNumber = getWeekNumber(weekStart)
            year = weekStart.getFullYear()
          }
          
          let weekPeriod = `Minggu ${weekNumber}, ${year}`
          if (weekStart && weekEnd) {
            weekPeriod = `Minggu ${weekNumber}, ${year} (${formatDateShort(weekStart)} - ${formatDateShort(weekEnd)})`
          }
          
          return {
            id: plan.id,
            weekPeriod: weekPeriod,
            approvalStatus: plan.status || 'pending',
            menuCount: plan.menu_items?.length || 0,
            menus: plan.menu_items?.map(item => {
              console.log('[MenuPlanning] Processing menu item:', item)
              
              // Get day of week from date
              const itemDate = item.date ? new Date(item.date) : null
              let dayOfWeek = 0
              if (itemDate) {
                // JavaScript getDay() returns 0-6 (Sunday-Saturday)
                // We need 1-7 (Monday-Sunday) for formatDay
                dayOfWeek = itemDate.getDay()
                // Convert: Sunday(0)→7, Monday(1)→1, ..., Saturday(6)→6
                if (dayOfWeek === 0) dayOfWeek = 7
              }
              
              console.log('[MenuPlanning] Item date:', item.date, 'Day of week:', dayOfWeek, 'Formatted:', formatDay(dayOfWeek))
              
              // Get school allocations for portions detail
              const allocations = item.school_allocations || []
              
              // Group allocations by school
              const schoolMap = new Map()
              allocations.forEach(alloc => {
                const schoolId = alloc.school_id
                if (!schoolMap.has(schoolId)) {
                  schoolMap.set(schoolId, {
                    schoolId: schoolId,
                    schoolName: alloc.school?.name || 'Sekolah',
                    portionsSmall: 0,
                    portionsLarge: 0
                  })
                }
                
                const school = schoolMap.get(schoolId)
                if (alloc.portion_size === 'small') {
                  school.portionsSmall += alloc.portions
                } else if (alloc.portion_size === 'large') {
                  school.portionsLarge += alloc.portions
                }
              })
              
              // Convert map to array and sort by school name
              const schools = Array.from(schoolMap.values()).sort((a, b) => 
                a.schoolName.localeCompare(b.schoolName)
              )
              
              // Calculate total portions from allocations
              const totalPortions = allocations.reduce((sum, alloc) => sum + (alloc.portions || 0), 0)
              
              return {
                day: formatDay(dayOfWeek),
                menuName: item.recipe?.name || 'Menu',
                portions: totalPortions || item.portions || 0,
                schools: schools
              }
            }) || []
          }
        })
        
        console.log('[MenuPlanning] Transformed weeklyPlans:', weeklyPlans.value)
      }
    } catch (err) {
      console.error('[MenuPlanning] Error:', err)
      console.error('[MenuPlanning] Error response:', err.response)
      error.value = err.response?.data?.message || 'Gagal memuat data perencanaan menu. Silakan coba lagi.'
    } finally {
      loading.value = false
    }
  }

  function formatDay(dayNum) {
    // day_of_week: 1=Monday, 2=Tuesday, ..., 7=Sunday
    const days = {
      1: 'Senin',
      2: 'Selasa', 
      3: 'Rabu',
      4: 'Kamis',
      5: 'Jumat',
      6: 'Sabtu',
      7: 'Minggu'
    }
    return days[dayNum] || `Hari ${dayNum}`
  }

  function getWeekNumber(date) {
    const d = new Date(date)
    d.setHours(0, 0, 0, 0)
    d.setDate(d.getDate() + 4 - (d.getDay() || 7))
    const yearStart = new Date(d.getFullYear(), 0, 1)
    return Math.ceil((((d - yearStart) / 86400000) + 1) / 7)
  }

  function formatDateShort(date) {
    return new Date(date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
  }

  async function approveMenu(id) {
    approvalLoading.value = true
    error.value = null

    try {
      await api.post(`/menu-plans/${id}/approve`)
      // Re-fetch to get updated status from server
      await fetchMenuPlans()
      return { success: true }
    } catch (err) {
      console.error('[MenuPlanning] Approve error:', err)
      error.value = err.response?.data?.message || 'Gagal menyetujui menu. Silakan coba lagi.'
      return { success: false, error: error.value }
    } finally {
      approvalLoading.value = false
    }
  }

  async function rejectMenu(id, reason) {
    approvalLoading.value = true
    error.value = null

    try {
      await api.post(`/menu-plans/${id}/reject`, { reason })
      // Re-fetch to get updated status from server
      await fetchMenuPlans()
      return { success: true }
    } catch (err) {
      console.error('[MenuPlanning] Reject error:', err)
      error.value = err.response?.data?.message || 'Gagal menolak menu. Silakan coba lagi.'
      return { success: false, error: error.value }
    } finally {
      approvalLoading.value = false
    }
  }

  function setWeek(week) {
    selectedWeek.value = week
    return fetchMenuPlans()
  }

  function retry() {
    return fetchMenuPlans()
  }

  return {
    weeklyPlans,
    selectedWeek,
    loading,
    approvalLoading,
    error,
    fetchMenuPlans,
    approveMenu,
    rejectMenu,
    setWeek,
    retry
  }
})
