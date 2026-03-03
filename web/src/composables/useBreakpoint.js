import { ref, computed, onMounted, onUnmounted } from 'vue'

/**
 * Composable for tracking viewport breakpoints and responsive states
 * 
 * Breakpoints:
 * - sm: < 640px
 * - md: < 768px
 * - lg: < 1024px
 * - xl: < 1280px
 * - 2xl: >= 1280px
 * 
 * @returns {Object} Breakpoint state and computed properties
 */
export function useBreakpoint() {
  const breakpoint = ref('xl')
  
  /**
   * Update breakpoint based on current window width
   */
  const updateBreakpoint = () => {
    const width = window.innerWidth
    
    if (width < 640) {
      breakpoint.value = 'sm'
    } else if (width < 768) {
      breakpoint.value = 'md'
    } else if (width < 1024) {
      breakpoint.value = 'lg'
    } else if (width < 1280) {
      breakpoint.value = 'xl'
    } else {
      breakpoint.value = '2xl'
    }
  }
  
  // Computed properties for device type detection
  const isMobile = computed(() => {
    return ['sm', 'md'].includes(breakpoint.value)
  })
  
  const isTablet = computed(() => {
    return breakpoint.value === 'lg'
  })
  
  const isDesktop = computed(() => {
    return ['xl', '2xl'].includes(breakpoint.value)
  })
  
  // Setup resize listener on mount
  onMounted(() => {
    updateBreakpoint()
    window.addEventListener('resize', updateBreakpoint)
  })
  
  // Cleanup listener on unmount
  onUnmounted(() => {
    window.removeEventListener('resize', updateBreakpoint)
  })
  
  return {
    breakpoint,
    isMobile,
    isTablet,
    isDesktop,
    updateBreakpoint
  }
}
