/**
 * Simple tests for signature functionality
 */
import { describe, it, expect } from 'vitest'
import { validateSignatureQuality, getSignatureSize } from '../src/utils/signatureValidator.js'

describe('Signature Validation', () => {
  // Test data
  const mockStrokes = [
    // Simple signature with 2 strokes
    [
      { x: 10, y: 50 },
      { x: 20, y: 45 },
      { x: 30, y: 40 },
      { x: 40, y: 35 }
    ],
    [
      { x: 15, y: 60 },
      { x: 25, y: 65 },
      { x: 35, y: 70 }
    ]
  ]

  const complexStrokes = [
    // Complex signature with 3 strokes
    [
      { x: 10, y: 50 },
      { x: 20, y: 45 },
      { x: 30, y: 40 },
      { x: 40, y: 35 },
      { x: 50, y: 30 },
      { x: 60, y: 25 }
    ],
    [
      { x: 15, y: 60 },
      { x: 25, y: 65 },
      { x: 35, y: 70 },
      { x: 45, y: 75 }
    ],
    [
      { x: 20, y: 80 },
      { x: 30, y: 85 },
      { x: 40, y: 90 },
      { x: 50, y: 95 },
      { x: 60, y: 100 }
    ]
  ]

  it('should reject empty signature', () => {
    const emptyResult = validateSignatureQuality([])
    expect(emptyResult.isValid).toBe(false)
    expect(emptyResult.quality).toBe(0)
    expect(emptyResult.feedback).toBe('Tanda tangan tidak ditemukan')
  })

  it('should validate simple signature', () => {
    const simpleResult = validateSignatureQuality(mockStrokes)
    expect(simpleResult.strokeCount).toBe(2)
    expect(simpleResult.quality).toBeGreaterThan(0)
  })

  it('should validate complex signature with higher quality', () => {
    const complexResult = validateSignatureQuality(complexStrokes)
    const simpleResult = validateSignatureQuality(mockStrokes)
    
    expect(complexResult.strokeCount).toBe(3)
    expect(complexResult.quality).toBeGreaterThan(simpleResult.quality)
    expect(complexResult.isValid).toBe(true)
  })

  it('should calculate signature size correctly', () => {
    const mockDataURL = 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=='
    const sizeResult = getSignatureSize(mockDataURL)
    
    expect(sizeResult.bytes).toBeGreaterThan(0)
    expect(sizeResult.formatted).toContain('B')
  })
})