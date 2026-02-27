import api from './api'

/**
 * Cleaning Service - Ompreng Cleaning API calls
 */

/**
 * Get pending ompreng awaiting cleaning
 * @param {string} date - Date in YYYY-MM-DD format (optional)
 * @returns {Promise} API response with pending ompreng list
 */
export const getPendingOmpreng = async (date = null) => {
  const params = date ? { date } : {}
  const response = await api.get('/cleaning/pending', { params })
  return response.data
}

/**
 * Start cleaning process for ompreng
 * @param {number} cleaningId - Cleaning record ID
 * @returns {Promise} API response with updated cleaning record
 */
export const startCleaning = async (cleaningId) => {
  const response = await api.post(`/cleaning/${cleaningId}/start`)
  return response.data
}

/**
 * Complete cleaning process for ompreng
 * @param {number} cleaningId - Cleaning record ID
 * @returns {Promise} API response with updated cleaning record
 */
export const completeCleaning = async (cleaningId) => {
  const response = await api.post(`/cleaning/${cleaningId}/complete`)
  return response.data
}
