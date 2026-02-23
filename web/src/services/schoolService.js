import api from './api'

const schoolService = {
  // Get all schools with optional filters
  getSchools(params = {}) {
    return api.get('/schools', { params })
  },

  // Get single school by ID
  getSchool(id) {
    return api.get(`/schools/${id}`)
  },

  // Create new school
  createSchool(data) {
    return api.post('/schools', data)
  },

  // Update existing school
  updateSchool(id, data) {
    return api.put(`/schools/${id}`, data)
  },

  // Delete school
  deleteSchool(id) {
    return api.delete(`/schools/${id}`)
  },

  // Search schools by name
  searchSchools(query, activeOnly = true) {
    return api.get('/schools', { 
      params: { 
        q: query,
        active_only: activeOnly 
      } 
    })
  }
}

export default schoolService