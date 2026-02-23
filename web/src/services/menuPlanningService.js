import api from './api'

const menuPlanningService = {
  // Get all menu plans
  getMenuPlans(params = {}) {
    return api.get('/menu-plans', { params })
  },

  // Get single menu plan by ID
  getMenuPlan(id) {
    return api.get(`/menu-plans/${id}`)
  },

  // Get current week menu plan
  getCurrentWeekMenu() {
    return api.get('/menu-plans/current-week')
  },

  // Create new menu plan
  createMenuPlan(data) {
    return api.post('/menu-plans', data)
  },

  // Update existing menu plan
  updateMenuPlan(id, data) {
    return api.put(`/menu-plans/${id}`, data)
  },

  // Approve menu plan
  approveMenuPlan(id) {
    return api.post(`/menu-plans/${id}/approve`)
  },

  // Delete menu plan
  deleteMenuPlan(id) {
    return api.delete(`/menu-plans/${id}`)
  }
}

export default menuPlanningService
