import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { useBreakpoint } from './useBreakpoint'

describe('useBreakpoint', () => {
  let originalInnerWidth
  
  beforeEach(() => {
    originalInnerWidth = window.innerWidth
  })
  
  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: originalInnerWidth
    })
  })
  
  const setWindowWidth = (width) => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: width
    })
  }
  
  it('should detect sm breakpoint for width < 640px', () => {
    setWindowWidth(500)
    const { breakpoint, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(breakpoint.value).toBe('sm')
  })
  
  it('should detect md breakpoint for width >= 640px and < 768px', () => {
    setWindowWidth(700)
    const { breakpoint, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(breakpoint.value).toBe('md')
  })
  
  it('should detect lg breakpoint for width >= 768px and < 1024px', () => {
    setWindowWidth(900)
    const { breakpoint, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(breakpoint.value).toBe('lg')
  })
  
  it('should detect xl breakpoint for width >= 1024px and < 1280px', () => {
    setWindowWidth(1100)
    const { breakpoint, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(breakpoint.value).toBe('xl')
  })
  
  it('should detect 2xl breakpoint for width >= 1280px', () => {
    setWindowWidth(1400)
    const { breakpoint, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(breakpoint.value).toBe('2xl')
  })
  
  it('should return isMobile true for sm and md breakpoints', () => {
    setWindowWidth(500)
    const { isMobile, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(isMobile.value).toBe(true)
    
    setWindowWidth(700)
    updateBreakpoint()
    expect(isMobile.value).toBe(true)
  })
  
  it('should return isMobile false for lg, xl, and 2xl breakpoints', () => {
    setWindowWidth(900)
    const { isMobile, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(isMobile.value).toBe(false)
  })
  
  it('should return isTablet true for lg breakpoint', () => {
    setWindowWidth(900)
    const { isTablet, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(isTablet.value).toBe(true)
  })
  
  it('should return isTablet false for non-lg breakpoints', () => {
    setWindowWidth(500)
    const { isTablet, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(isTablet.value).toBe(false)
  })
  
  it('should return isDesktop true for xl and 2xl breakpoints', () => {
    setWindowWidth(1100)
    const { isDesktop, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(isDesktop.value).toBe(true)
    
    setWindowWidth(1400)
    updateBreakpoint()
    expect(isDesktop.value).toBe(true)
  })
  
  it('should return isDesktop false for sm, md, and lg breakpoints', () => {
    setWindowWidth(500)
    const { isDesktop, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(isDesktop.value).toBe(false)
  })
  
  it('should handle resize events', () => {
    setWindowWidth(500)
    const { breakpoint, updateBreakpoint } = useBreakpoint()
    updateBreakpoint()
    expect(breakpoint.value).toBe('sm')
    
    setWindowWidth(1400)
    updateBreakpoint()
    expect(breakpoint.value).toBe('2xl')
  })
})
