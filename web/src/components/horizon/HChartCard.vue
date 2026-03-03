<template>
  <div class="h-chart-card h-card">
    <!-- Loading State -->
    <a-skeleton v-if="loading" active :paragraph="{ rows: 4 }" />
    
    <!-- Content -->
    <div v-else class="h-chart-card__content">
      <!-- Header -->
      <div class="h-chart-card__header">
        <div class="h-chart-card__header-left">
          <!-- Title -->
          <h3 class="h-chart-card__title">
            {{ title }}
          </h3>
          
          <!-- Subtitle -->
          <p v-if="subtitle" class="h-chart-card__subtitle">
            {{ subtitle }}
          </p>
        </div>
        
        <!-- Header Right Slot for Custom Actions -->
        <div v-if="$slots['header-right']" class="h-chart-card__header-right">
          <slot name="header-right" />
        </div>
      </div>
      
      <!-- Chart Content Area -->
      <div 
        class="h-chart-card__chart" 
        :style="{ height: chartHeight }"
      >
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  /**
   * Chart card title
   * Example: "Weekly Revenue", "Daily Traffic"
   */
  title: {
    type: String,
    required: true
  },
  
  /**
   * Optional subtitle text
   * Example: "Last 7 days", "This month"
   */
  subtitle: {
    type: String,
    default: ''
  },
  
  /**
   * Chart height in pixels
   * Default: 320px (desktop), 240px (mobile)
   */
  height: {
    type: Number,
    default: 320
  },
  
  /**
   * Loading state - shows skeleton
   */
  loading: {
    type: Boolean,
    default: false
  }
})

// Compute chart height with px unit
const chartHeight = computed(() => `${props.height}px`)
</script>

<style scoped>
.h-chart-card {
  /* Use h-card utility class for base styling */
  display: flex;
  flex-direction: column;
  transition: all var(--h-transition-base);
}

.h-chart-card__content {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-5);
  height: 100%;
}

/* Header */
.h-chart-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--h-spacing-4);
}

.h-chart-card__header-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-1);
  min-width: 0; /* Allow text truncation */
}

.h-chart-card__header-right {
  flex-shrink: 0;
}

/* Title */
.h-chart-card__title {
  font-size: var(--h-text-lg);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  line-height: var(--h-leading-tight);
  margin: 0;
}

/* Subtitle */
.h-chart-card__subtitle {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-normal);
  color: var(--h-text-secondary);
  line-height: var(--h-leading-normal);
  margin: 0;
}

/* Chart Content Area */
.h-chart-card__chart {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
  position: relative;
  overflow: hidden;
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .h-chart-card__content {
    gap: var(--h-spacing-4);
  }
  
  .h-chart-card__header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .h-chart-card__header-right {
    width: 100%;
  }
  
  .h-chart-card__title {
    font-size: var(--h-text-base);
  }
  
  .h-chart-card__subtitle {
    font-size: var(--h-text-xs);
  }
  
  /* Reduced height on mobile */
  .h-chart-card__chart {
    height: 240px !important;
    min-height: 240px;
  }
}

/* Dark Mode Support */
.dark .h-chart-card__title {
  color: var(--h-text-primary);
}

.dark .h-chart-card__subtitle {
  color: var(--h-text-secondary);
}
</style>
