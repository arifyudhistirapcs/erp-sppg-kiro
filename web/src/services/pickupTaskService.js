import api from './api'

/**
 * Pickup Task Service
 * API client for managing pickup tasks (ompreng collection)
 * Validates Requirements: 1.5, 2.1, 5.1
 */
const pickupTaskService = {
  /**
   * Get delivery records that are eligible for pickup (Stage 9)
   * @param {string} date - Optional date filter in YYYY-MM-DD format
   * @returns {Promise} API response with eligible orders
   */
  getEligibleOrders(date) {
    const params = date ? { date } : {}
    return api.get('/pickup-tasks/eligible-orders', { params })
  },

  /**
   * Get drivers available for pickup task assignment
   * @param {string} date - Optional date filter in YYYY-MM-DD format
   * @returns {Promise} API response with available drivers
   */
  getAvailableDrivers(date) {
    const params = date ? { date } : {}
    return api.get('/pickup-tasks/available-drivers', { params })
  },

  /**
   * Create a new pickup task
   * @param {object} data - Pickup task data
   * @param {string} data.task_date - Task date in ISO format
   * @param {number} data.driver_id - Driver ID
   * @param {Array} data.delivery_records - Array of {delivery_record_id, route_order}
   * @returns {Promise} API response with created pickup task
   */
  createPickupTask(data) {
    return api.post('/pickup-tasks', data)
  },

  /**
   * Get all pickup tasks with optional filters
   * @param {object} params - Query parameters
   * @param {string} params.date - Filter by task date (YYYY-MM-DD)
   * @param {number} params.driver_id - Filter by driver ID
   * @param {string} params.status - Filter by status (active, completed, cancelled)
   * @returns {Promise} API response with pickup tasks list
   */
  getPickupTasks(params = {}) {
    return api.get('/pickup-tasks', { params })
  },

  /**
   * Get detailed information about a specific pickup task
   * @param {number} id - Pickup task ID
   * @returns {Promise} API response with pickup task details
   */
  getPickupTask(id) {
    return api.get(`/pickup-tasks/${id}`)
  },

  /**
   * Update the status of a pickup task
   * @param {number} id - Pickup task ID
   * @param {string} status - New status (active, completed, cancelled)
   * @returns {Promise} API response
   */
  updatePickupTaskStatus(id, status) {
    return api.put(`/pickup-tasks/${id}/status`, { status })
  },

  /**
   * Cancel a pickup task (soft delete)
   * @param {number} id - Pickup task ID
   * @returns {Promise} API response
   */
  cancelPickupTask(id) {
    return api.delete(`/pickup-tasks/${id}`)
  },

  /**
   * Update the stage of an individual delivery record within a pickup task
   * @param {number} pickupTaskId - Pickup task ID
   * @param {number} deliveryRecordId - Delivery record ID
   * @param {object} data - Stage update data
   * @param {number} data.stage - New stage (11, 12, or 13)
   * @param {string} data.status - New status corresponding to the stage
   * @returns {Promise} API response with updated delivery record
   * Validates Requirements: 11.1, 11.2
   */
  updateDeliveryRecordStage(pickupTaskId, deliveryRecordId, data) {
    return api.put(`/pickup-tasks/${pickupTaskId}/delivery-records/${deliveryRecordId}/stage`, data)
  }
}

export default pickupTaskService
