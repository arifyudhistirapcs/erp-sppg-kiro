/**
 * Signature validation utilities for e-POD form
 */

/**
 * Validates signature quality based on stroke complexity
 * @param {Array} strokes - Array of stroke data
 * @returns {Object} Validation result with quality score and feedback
 */
export const validateSignatureQuality = (strokes) => {
  if (!strokes || strokes.length === 0) {
    return {
      isValid: false,
      quality: 0,
      feedback: 'Tanda tangan tidak ditemukan'
    }
  }

  let totalPoints = 0
  let totalDistance = 0
  const strokeCount = strokes.length

  strokes.forEach(stroke => {
    totalPoints += stroke.length
    
    // Calculate stroke distance
    for (let i = 1; i < stroke.length; i++) {
      const dx = stroke[i].x - stroke[i-1].x
      const dy = stroke[i].y - stroke[i-1].y
      totalDistance += Math.sqrt(dx * dx + dy * dy)
    }
  })

  // Quality calculation
  let quality = 0
  
  if (strokeCount >= 2) quality += 30
  if (strokeCount >= 3) quality += 20
  if (totalDistance > 100) quality += 25
  if (totalDistance > 200) quality += 15
  if (totalPoints > 20) quality += 10

  quality = Math.min(100, quality)

  // Validation rules
  const isValid = quality >= 40 && strokeCount >= 2
  
  let feedback = ''
  if (!isValid) {
    if (strokeCount < 2) {
      feedback = 'Tanda tangan harus terdiri dari minimal 2 goresan'
    } else if (quality < 40) {
      feedback = 'Tanda tangan terlalu sederhana. Silakan buat tanda tangan yang lebih kompleks'
    }
  } else {
    feedback = getQualityFeedback(quality)
  }

  return {
    isValid,
    quality,
    feedback,
    strokeCount,
    totalDistance: Math.round(totalDistance),
    totalPoints
  }
}

/**
 * Get quality feedback text
 * @param {number} quality - Quality score (0-100)
 * @returns {string} Quality feedback text
 */
export const getQualityFeedback = (quality) => {
  if (quality >= 70) return 'Tanda tangan sangat baik'
  if (quality >= 50) return 'Tanda tangan baik'
  if (quality >= 40) return 'Tanda tangan cukup'
  return 'Tanda tangan kurang kompleks'
}

/**
 * Compress signature canvas to optimized PNG
 * @param {HTMLCanvasElement} canvas - Signature canvas
 * @param {number} maxWidth - Maximum width for compression
 * @param {number} maxHeight - Maximum height for compression
 * @returns {string} Compressed signature data URL
 */
export const compressSignature = (canvas, maxWidth = 400, maxHeight = 200) => {
  if (!canvas) return null

  // Create compression canvas
  const compressCanvas = document.createElement('canvas')
  const compressCtx = compressCanvas.getContext('2d')
  
  compressCanvas.width = maxWidth
  compressCanvas.height = maxHeight
  
  // Fill with white background
  compressCtx.fillStyle = '#ffffff'
  compressCtx.fillRect(0, 0, maxWidth, maxHeight)
  
  // Draw original signature scaled to fit
  compressCtx.drawImage(canvas, 0, 0, maxWidth, maxHeight)
  
  // Convert to PNG with good compression
  return compressCanvas.toDataURL('image/png', 0.8)
}

/**
 * Estimate signature file size
 * @param {string} dataURL - Signature data URL
 * @returns {Object} Size information
 */
export const getSignatureSize = (dataURL) => {
  if (!dataURL) return { bytes: 0, formatted: '0 B' }
  
  const base64Length = dataURL.length
  const sizeInBytes = (base64Length * 3) / 4
  
  let formatted = ''
  if (sizeInBytes < 1024) {
    formatted = `${Math.round(sizeInBytes)} B`
  } else if (sizeInBytes < 1024 * 1024) {
    formatted = `${Math.round(sizeInBytes / 1024)} KB`
  } else {
    formatted = `${Math.round(sizeInBytes / (1024 * 1024))} MB`
  }
  
  return {
    bytes: Math.round(sizeInBytes),
    formatted
  }
}