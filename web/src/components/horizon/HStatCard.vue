<template>
  <div class="h-stat-card h-card">
    <!-- Loading State -->
    <a-skeleton v-if="loading" active :paragraph="{ rows: 2 }" />
    
    <!-- Content -->
    <div v-else class="h-stat-card__content">
      <!-- Icon Container -->
      <div 
        class="h-stat-card__icon" 
        :style="{ background: iconBg || 'var(--h-gradient-primary)' }"
      >
        <component 
          v-if="icon" 
          :is="icon" 
          class="h-stat-card__icon-svg"
        />
      </div>
      
      <!-- Stats Content -->
      <div class="h-stat-card__stats">
        <!-- Label -->
        <div class="h-stat-card__label">
          {{ label }}
        </div>
        
        <!-- Value -->
        <div class="h-stat-card__value">
          {{ value }}
        </div>
        
        <!-- Change Indicator -->
        <div 
          v-if="change" 
          class="h-stat-card__change"
          :class="`h-stat-card__change--${changeType}`"
        >
          <ArrowUpOutlined v-if="changeType === 'increase'" class="h-stat-card__change-icon" />
          <ArrowDownOutlined v-if="changeType === 'decrease'" class="h-stat-card__change-icon" />
          <span>{{ change }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ArrowUpOutlined, ArrowDownOutlined } from '@ant-design/icons-vue'

defineProps({
  /**
   * Icon component from @ant-design/icons-vue
   * Example: DollarOutlined, ShoppingOutlined, etc.
   */
  icon: {
    type: Object,
    default: null
  },
  
  /**
   * Background gradient for icon container
   * Example: 'linear-gradient(135deg, #f82c17 0%, #ff4d38 100%)'
   */
  iconBg: {
    type: String,
    default: ''
  },
  
  /**
   * Label text (e.g., "Total Pendapatan", "Earnings")
   */
  label: {
    type: String,
    required: true
  },
  
  /**
   * Value to display (e.g., "Rp 45.2M", "$350.4")
   */
  value: {
    type: [String, Number],
    required: true
  },
  
  /**
   * Change percentage (e.g., "+23%", "-5%")
   */
  change: {
    type: String,
    default: ''
  },
  
  /**
   * Type of change: 'increase' or 'decrease'
   */
  changeType: {
    type: String,
    default: 'increase',
    validator: (value) => ['increase', 'decrease'].includes(value)
  },
  
  /**
   * Loading state - shows skeleton
   */
  loading: {
    type: Boolean,
    default: false
  }
})
</script>

<style scoped>
.h-stat-card {
  /* Use h-card utility class for base styling */
  display: flex;
  flex-direction: column;
  min-height: 100px;
  transition: all var(--h-transition-base);
}

.h-stat-card__content {
  display: flex;
  gap: var(--h-spacing-4);
  align-items: flex-start;
}

/* Icon Container */
.h-stat-card__icon {
  width: 56px;
  height: 56px;
  border-radius: var(--h-radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: linear-gradient(135deg, #5A4372 0%, #3D2B53 100%);
}

.h-stat-card__icon-svg {
  font-size: 24px;
  color: #FFFFFF;
}

/* Stats Content */
.h-stat-card__stats {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-1);
  min-width: 0; /* Allow text truncation */
}

/* Label */
.h-stat-card__label {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  font-weight: var(--h-font-medium);
  line-height: var(--h-leading-tight);
}

/* Value */
.h-stat-card__value {
  font-size: var(--h-text-2xl);
  color: var(--h-text-primary);
  font-weight: var(--h-font-bold);
  line-height: var(--h-leading-tight);
  word-break: break-word;
}

/* Change Indicator */
.h-stat-card__change {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-1);
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  line-height: var(--h-leading-tight);
}

.h-stat-card__change-icon {
  font-size: var(--h-text-xs);
}

.h-stat-card__change--increase {
  color: var(--h-success);
}

.h-stat-card__change--decrease {
  color: var(--h-error);
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .h-stat-card {
    width: 100%;
  }
  
  .h-stat-card__content {
    gap: var(--h-spacing-3);
  }
  
  .h-stat-card__icon {
    width: 48px;
    height: 48px;
  }
  
  .h-stat-card__icon-svg {
    font-size: 20px;
  }
  
  .h-stat-card__value {
    font-size: var(--h-text-xl);
  }
}

/* Dark Mode Support */
.dark .h-stat-card__label {
  color: var(--h-text-secondary);
}

.dark .h-stat-card__value {
  color: var(--h-text-primary);
}
</style>
