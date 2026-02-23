import api from './api'

const omprengTrackingService = {
  // Get ompreng tracking data for all schools
  getOmprengTracking() {
    return api.get('/ompreng/tracking')
  },

  // Record ompreng drop-off at a school
  recordDropOff(schoolId, quantity) {
    return api.post('/ompreng/drop-off', {
      school_id: schoolId,
      quantity: quantity
    })
  },

  // Record ompreng pick-up from a school
  recordPickUp(schoolId, quantity) {
    return api.post('/ompreng/pick-up', {
      school_id: schoolId,
      quantity: quantity
    })
  },

  // Get ompreng circulation reports
  getReports(startDate = null, endDate = null) {
    const params = {}
    if (startDate) params.start_date = startDate
    if (endDate) params.end_date = endDate
    return api.get('/ompreng/reports', { params })
  },

  // Get school tracking history
  getSchoolHistory(schoolId, startDate = null, endDate = null) {
    const params = { school_id: schoolId }
    if (startDate) params.start_date = startDate
    if (endDate) params.end_date = endDate
    return api.get('/ompreng/tracking/history', { params })
  }
}

export default omprengTrackingService