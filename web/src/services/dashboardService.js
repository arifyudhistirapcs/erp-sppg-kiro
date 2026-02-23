import api from './api'

/**
 * Dashboard Service - Executive Dashboard API calls
 */

// Get Kepala SPPG Dashboard data
export const getKepalaSSPGDashboard = async () => {
  const response = await api.get('/dashboard/kepala-sppg')
  return response.data
}

// Get Kepala Yayasan Dashboard data
export const getKepalaYayasanDashboard = async (startDate, endDate) => {
  const params = {}
  if (startDate) params.start_date = startDate
  if (endDate) params.end_date = endDate
  
  const response = await api.get('/dashboard/kepala-yayasan', { params })
  return response.data
}

// Get KPIs data
export const getKPIs = async () => {
  const response = await api.get('/dashboard/kpi')
  return response.data
}

// Sync dashboard to Firebase
export const syncDashboardToFirebase = async (type, startDate, endDate) => {
  const params = { type }
  if (startDate) params.start_date = startDate
  if (endDate) params.end_date = endDate
  
  const response = await api.post('/dashboard/sync', null, { params })
  return response.data
}

// Export dashboard data
export const exportDashboardData = async (type, format, startDate, endDate) => {
  const data = {
    type,
    format,
    start_date: startDate,
    end_date: endDate
  }
  
  const response = await api.post('/dashboard/export', data)
  return response.data
}