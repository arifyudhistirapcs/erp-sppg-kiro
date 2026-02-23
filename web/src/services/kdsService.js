import api from './api'

/**
 * KDS Service - Kitchen Display System API calls
 */

// Get today's cooking menu
export const getCookingToday = async () => {
  const response = await api.get('/kds/cooking/today')
  return response.data
}

// Update cooking status for a recipe
export const updateCookingStatus = async (recipeId, status) => {
  const response = await api.put(`/kds/cooking/${recipeId}/status`, { status })
  return response.data
}

// Get today's packing allocations
export const getPackingToday = async () => {
  const response = await api.get('/kds/packing/today')
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
