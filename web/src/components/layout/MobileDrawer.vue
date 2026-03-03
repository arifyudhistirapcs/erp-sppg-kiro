<template>
  <Teleport to="body">
    <!-- Overlay Backdrop -->
    <Transition name="fade">
      <div
        v-if="isOpen"
        class="mobile-drawer-overlay"
        @click="close"
        @touchstart="handleOverlayTouchStart"
      />
    </Transition>

    <!-- Drawer -->
    <Transition name="slide">
      <div
        v-if="isOpen"
        ref="drawerRef"
        class="mobile-drawer"
        @touchstart="handleTouchStart"
        @touchmove="handleTouchMove"
        @touchend="handleTouchEnd"
      >
        <slot />
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'

/**
 * MobileDrawer Component
 * 
 * Mobile drawer yang slide-in dari kiri dengan:
 * - Fixed position, width 280px
 * - Overlay backdrop rgba(0,0,0,0.5) dengan tap-to-close
 * - Transition 300ms ease
 * - Z-index 1000
 * - Swipe-to-close gesture support
 * - Prevent body scroll saat drawer open
 * - v-model support untuk open/close state
 */

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const drawerRef = ref(null)
const isOpen = ref(props.modelValue)

// Touch gesture tracking
const touchStartX = ref(0)
const touchCurrentX = ref(0)
const isDragging = ref(false)

/**
 * Close drawer
 */
const close = () => {
  emit('update:modelValue', false)
}

/**
 * Handle touch start on drawer
 */
const handleTouchStart = (e) => {
  touchStartX.value = e.touches[0].clientX
  touchCurrentX.value = e.touches[0].clientX
  isDragging.value = true
}

/**
 * Handle touch move - track swipe gesture
 */
const handleTouchMove = (e) => {
  if (!isDragging.value) return
  
  touchCurrentX.value = e.touches[0].clientX
  const deltaX = touchCurrentX.value - touchStartX.value
  
  // Only allow swipe left (negative delta)
  if (deltaX < 0 && drawerRef.value) {
    // Apply transform untuk visual feedback
    const translateX = Math.max(deltaX, -280)
    drawerRef.value.style.transform = `translateX(${translateX}px)`
  }
}

/**
 * Handle touch end - determine if should close
 */
const handleTouchEnd = () => {
  if (!isDragging.value) return
  
  const deltaX = touchCurrentX.value - touchStartX.value
  
  // Close jika swipe lebih dari 100px ke kiri
  if (deltaX < -100) {
    close()
  }
  
  // Reset transform
  if (drawerRef.value) {
    drawerRef.value.style.transform = ''
  }
  
  isDragging.value = false
}

/**
 * Handle overlay touch start - prevent touch events dari bubbling
 */
const handleOverlayTouchStart = (e) => {
  // Prevent default untuk avoid scroll issues
  e.preventDefault()
}

/**
 * Prevent body scroll saat drawer open
 */
const preventBodyScroll = (prevent) => {
  if (prevent) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
}

/**
 * Handle escape key press
 */
const handleEscapeKey = (e) => {
  if (e.key === 'Escape' && isOpen.value) {
    close()
  }
}

// Watch modelValue changes
watch(() => props.modelValue, (newValue) => {
  isOpen.value = newValue
  preventBodyScroll(newValue)
})

// Watch isOpen changes
watch(isOpen, (newValue) => {
  preventBodyScroll(newValue)
})

// Setup keyboard listener
onMounted(() => {
  document.addEventListener('keydown', handleEscapeKey)
})

// Cleanup
onUnmounted(() => {
  document.removeEventListener('keydown', handleEscapeKey)
  preventBodyScroll(false)
})
</script>

<style scoped>
/* Overlay Backdrop */
.mobile-drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 1000;
  cursor: pointer;
}

/* Drawer */
.mobile-drawer {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  background-color: var(--h-bg-secondary, #FFFFFF);
  z-index: 1001;
  overflow-y: auto;
  overflow-x: hidden;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.15);
  
  /* Smooth scrolling */
  -webkit-overflow-scrolling: touch;
}

/* Dark mode support */
.dark .mobile-drawer {
  background-color: var(--h-bg-secondary-dark, #111C44);
}

/* Fade transition untuk overlay */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 300ms ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Slide transition untuk drawer */
.slide-enter-active,
.slide-leave-active {
  transition: transform 300ms ease;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(-280px);
}

/* Prevent text selection during drag */
.mobile-drawer.dragging {
  user-select: none;
  -webkit-user-select: none;
}
</style>
