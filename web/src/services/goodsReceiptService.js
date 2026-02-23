import api from './api'

const goodsReceiptService = {
  // Get all goods receipts with optional filters
  getGoodsReceipts(params = {}) {
    return api.get('/goods-receipts', { params })
  },

  // Get single goods receipt by ID
  getGoodsReceipt(id) {
    return api.get(`/goods-receipts/${id}`)
  },

  // Create new goods receipt
  createGoodsReceipt(data) {
    return api.post('/goods-receipts', data)
  },

  // Upload invoice photo
  uploadInvoicePhoto(id, formData) {
    return api.post(`/goods-receipts/${id}/upload-invoice`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}

export default goodsReceiptService
