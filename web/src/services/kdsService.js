import api from './api'

/**
 * KDS Service - Kitchen Display System API calls
 */

/**
 * Format date to YYYY-MM-DD
 * @param {Date} date - Date object to format
 * @returns {string} Formatted date string
 */
const formatDate = (date) => {
  // Use local date to avoid timezone conversion issues
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Get today's cooking menu
export const getCookingToday = async (date = null) => {
  console.log('[kdsService] getCookingToday called with date:', date)
  const params = date ? { date: formatDate(date) } : {}
  console.log('[kdsService] Request params:', params)
  const response = await api.get('/kds/cooking/today', { params })
  console.log('[kdsService] Response:', response.data)
  return response.data
}

// Update cooking status for a recipe
export const updateCookingStatus = async (recipeId, status) => {
  const response = await api.put(`/kds/cooking/${recipeId}/status`, { status })
  return response.data
}

// Get today's packing allocations
export const getPackingToday = async (date = null) => {
  const params = date ? { date: formatDate(date) } : {}
  const response = await api.get('/kds/packing/today', { params })
  return response.data
}

// Update packing status for a school
export const updatePackingStatus = async (schoolId, status) => {
  const response = await api.put(`/kds/packing/${schoolId}/status`, { status })
  return response.data
}

// Sync cooking menu to Firebase
export const syncCookingToFirebase = async () => {
  const response = await api.post('/kds/cooking/sync')
  return response.data
}

// Sync packing allocations to Firebase
export const syncPackingToFirebase = async () => {
  const response = await api.post('/kds/packing/sync')
  return response.data
}
