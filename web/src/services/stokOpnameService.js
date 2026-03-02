import api from './api'

const stokOpnameService = {
  // Form operations
  createForm(data) {
    return api.post('/stok-opname/forms', data)
  },

  getForms(params = {}) {
    return api.get('/stok-opname/forms', { params })
  },

  getForm(id) {
    return api.get(`/stok-opname/forms/${id}`)
  },

  updateFormNotes(id, notes) {
    return api.put(`/stok-opname/forms/${id}/notes`, { notes })
  },

  deleteForm(id) {
    return api.delete(`/stok-opname/forms/${id}`)
  },

  // Item operations
  addItem(formId, data) {
    return api.post(`/stok-opname/forms/${formId}/items`, data)
  },

  updateItem(itemId, data) {
    return api.put(`/stok-opname/items/${itemId}`, data)
  },

  removeItem(itemId) {
    return api.delete(`/stok-opname/items/${itemId}`)
  },

  // Workflow operations
  submitForApproval(formId) {
    return api.post(`/stok-opname/forms/${formId}/submit`)
  },

  approveForm(formId) {
    return api.post(`/stok-opname/forms/${formId}/approve`)
  },

  rejectForm(formId, reason) {
    return api.post(`/stok-opname/forms/${formId}/reject`, { reason })
  },

  // Export operation
  exportForm(formId, format = 'excel') {
    return api.get(`/stok-opname/forms/${formId}/export`, {
      params: { format },
      responseType: 'blob'
    })
  }
}

export default stokOpnameService
