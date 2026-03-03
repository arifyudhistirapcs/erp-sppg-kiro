<template>
  <nav v-if="!isMobile && breadcrumbs.length > 0" class="h-breadcrumb" aria-label="Breadcrumb">
    <ol class="breadcrumb-list">
      <li
        v-for="(crumb, index) in breadcrumbs"
        :key="index"
        class="breadcrumb-item"
      >
        <router-link
          v-if="crumb.to && index < breadcrumbs.length - 1"
          :to="crumb.to"
          class="breadcrumb-link"
        >
          {{ crumb.label }}
        </router-link>
        <span
          v-else
          class="breadcrumb-text"
          :class="{ 'breadcrumb-current': index === breadcrumbs.length - 1 }"
        >
          {{ crumb.label }}
        </span>
        <span
          v-if="index < breadcrumbs.length - 1"
          class="breadcrumb-separator"
          aria-hidden="true"
        >
          /
        </span>
      </li>
    </ol>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useBreakpoint } from '@/composables/useBreakpoint'

/**
 * HBreadcrumb Component
 * 
 * Breadcrumb navigation yang:
 * - Reads current route dari vue-router
 * - Renders "Pages / Page Name" format
 * - Auto-generates dari route meta atau accepts custom breadcrumb prop
 * - Styled sesuai Horizon UI (14px, #74788C)
 * - Hidden on mobile (< 768px)
 * - Dark mode support
 */

const props = defineProps({
  /**
   * Custom breadcrumb items
   * Format: [{ label: 'Home', to: '/' }, { label: 'Current Page' }]
   * If not provided, auto-generates from route
   */
  items: {
    type: Array,
    default: null
  },
  
  /**
   * Root breadcrumb label (default: 'Pages')
   */
  rootLabel: {
    type: String,
    default: 'Pages'
  },
  
  /**
   * Show root breadcrumb
   */
  showRoot: {
    type: Boolean,
    default: true
  }
})

const route = useRoute()
const { isMobile } = useBreakpoint()

/**
 * Generate breadcrumbs from route or custom items
 */
const breadcrumbs = computed(() => {
  // Use custom items if provided
  if (props.items && props.items.length > 0) {
    return props.items
  }
  
  // Auto-generate from route
  const crumbs = []
  
  // Add root breadcrumb
  if (props.showRoot) {
    crumbs.push({
      label: props.rootLabel,
      to: null
    })
  }
  
  // Get page title from route meta or name
  const pageTitle = route.meta?.title || formatRouteName(route.name)
  
  // Add current page
  if (pageTitle) {
    crumbs.push({
      label: pageTitle,
      to: null
    })
  }
  
  return crumbs
})

/**
 * Format route name to readable title
 * Example: 'dashboard-kepala-sppg' -> 'Dashboard Kepala SPPG'
 */
const formatRouteName = (name) => {
  if (!name) return ''
  
  return name
    .split('-')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}
</script>

<style scoped>
.h-breadcrumb {
  display: flex;
  align-items: center;
  font-size: 14px;
  color: var(--h-text-secondary, #74788C);
}

.breadcrumb-list {
  display: flex;
  align-items: center;
  list-style: none;
  margin: 0;
  padding: 0;
  gap: 8px;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.breadcrumb-link {
  color: var(--h-text-secondary, #74788C);
  text-decoration: none;
  transition: color 0.2s ease;
}

.breadcrumb-link:hover {
  color: var(--h-primary, #5A4372);
}

.breadcrumb-text {
  color: var(--h-text-secondary, #74788C);
}

.breadcrumb-current {
  color: var(--h-text-primary, #322837);
  font-weight: 500;
}

.breadcrumb-separator {
  color: var(--h-text-secondary, #74788C);
  user-select: none;
}

/* Dark mode */
.dark .breadcrumb-link {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.dark .breadcrumb-link:hover {
  color: var(--h-primary-light, #6a5382);
}

.dark .breadcrumb-text {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.dark .breadcrumb-current {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.dark .breadcrumb-separator {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

/* Hidden on mobile (< 768px) */
@media (max-width: 767px) {
  .h-breadcrumb {
    display: none;
  }
}
</style>
