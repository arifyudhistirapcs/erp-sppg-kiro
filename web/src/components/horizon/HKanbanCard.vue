<template>
  <div class="h-kanban-card">
    <!-- Drag Handle (Optional) -->
    <div v-if="showDragHandle" class="h-kanban-card__drag-handle">
      <HolderOutlined />
    </div>
    
    <!-- Image (Optional) -->
    <div v-if="image" class="h-kanban-card__image-container">
      <img :src="image" :alt="title" class="h-kanban-card__image" />
    </div>
    
    <!-- Content -->
    <div class="h-kanban-card__content">
      <!-- Header: Title + Status Badge -->
      <div class="h-kanban-card__header">
        <h3 class="h-kanban-card__title">{{ title }}</h3>
        <span 
          v-if="status" 
          class="h-kanban-card__status-badge"
          :class="`h-kanban-card__status-badge--${statusType}`"
        >
          {{ status }}
        </span>
      </div>
      
      <!-- Description -->
      <p v-if="description" class="h-kanban-card__description">
        {{ description }}
      </p>
      
      <!-- Default Slot for Custom Content -->
      <slot />
      
      <!-- Footer: Assignees + Due Date (only if provided) -->
      <div v-if="assignees?.length || dueDate" class="h-kanban-card__footer">
        <!-- Assignee Avatars -->
        <div v-if="assignees?.length" class="h-kanban-card__assignees">
          <a-avatar
            v-for="(assignee, index) in assignees"
            :key="index"
            :size="32"
            :src="assignee.avatar"
            class="h-kanban-card__avatar"
          >
            {{ assignee.name?.charAt(0) || '?' }}
          </a-avatar>
        </div>
        
        <!-- Due Date -->
        <div v-if="dueDate" class="h-kanban-card__due-date">
          <CalendarOutlined class="h-kanban-card__due-date-icon" />
          <span>{{ formattedDueDate }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { HolderOutlined, CalendarOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  /**
   * Card title
   */
  title: {
    type: String,
    required: true
  },
  
  /**
   * Card description
   */
  description: {
    type: String,
    default: ''
  },
  
  /**
   * Optional image URL
   */
  image: {
    type: String,
    default: ''
  },
  
  /**
   * Status text (e.g., "Backlog", "In Progress", "Done", "Urgent")
   */
  status: {
    type: String,
    default: ''
  },
  
  /**
   * Array of assignees
   * Each assignee: { name: String, avatar: String }
   */
  assignees: {
    type: Array,
    default: () => []
  },
  
  /**
   * Due date (Date object or string)
   */
  dueDate: {
    type: [Date, String],
    default: null
  },
  
  /**
   * Show drag handle
   */
  showDragHandle: {
    type: Boolean,
    default: false
  }
})

// Determine status type for styling
const statusType = computed(() => {
  const statusLower = props.status.toLowerCase()
  if (statusLower.includes('backlog')) return 'backlog'
  if (statusLower.includes('progress')) return 'in-progress'
  if (statusLower.includes('done') || statusLower.includes('complete')) return 'done'
  if (statusLower.includes('urgent')) return 'urgent'
  return 'default'
})

// Format due date
const formattedDueDate = computed(() => {
  if (!props.dueDate) return ''
  
  const date = props.dueDate instanceof Date ? props.dueDate : new Date(props.dueDate)
  
  // Format as "MMM DD" (e.g., "Jan 15")
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
})
</script>

<style scoped>
.h-kanban-card {
  width: 100%;
  min-height: 120px;
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-md);
  padding: var(--h-spacing-5);
  transition: all var(--h-transition-base);
  cursor: pointer;
  position: relative;
}

.h-kanban-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--h-shadow-lg);
}

/* Drag Handle */
.h-kanban-card__drag-handle {
  position: absolute;
  top: var(--h-spacing-2);
  right: var(--h-spacing-2);
  color: var(--h-text-light);
  cursor: grab;
  font-size: 16px;
  padding: var(--h-spacing-1);
}

.h-kanban-card__drag-handle:active {
  cursor: grabbing;
}

/* Image */
.h-kanban-card__image-container {
  width: calc(100% + var(--h-spacing-5) * 2);
  margin: calc(var(--h-spacing-5) * -1) calc(var(--h-spacing-5) * -1) var(--h-spacing-4);
  border-radius: var(--h-radius-lg) var(--h-radius-lg) 0 0;
  overflow: hidden;
}

.h-kanban-card__image {
  width: 100%;
  height: auto;
  display: block;
  object-fit: cover;
  border-radius: var(--h-radius-md);
}

/* Content */
.h-kanban-card__content {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-3);
}

/* Header */
.h-kanban-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--h-spacing-2);
}

/* Title */
.h-kanban-card__title {
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  line-height: var(--h-leading-tight);
  margin: 0;
  flex: 1;
}

/* Status Badge */
.h-kanban-card__status-badge {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
  white-space: nowrap;
  flex-shrink: 0;
}

.h-kanban-card__status-badge--backlog {
  background: #F4F7FE;
  color: #74788C;
}

.h-kanban-card__status-badge--in-progress {
  background: #FFF4E6;
  color: #FFB547;
}

.h-kanban-card__status-badge--done {
  background: #E6FAF5;
  color: #05CD99;
}

.h-kanban-card__status-badge--urgent {
  background: #FFEBE9;
  color: #EE5D50;
}

.h-kanban-card__status-badge--default {
  background: #F4F7FE;
  color: #74788C;
}

/* Description */
.h-kanban-card__description {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  line-height: 1.6;
  margin: 0;
}

/* Footer */
.h-kanban-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--h-spacing-3);
  margin-top: var(--h-spacing-2);
}

/* Assignees */
.h-kanban-card__assignees {
  display: flex;
  align-items: center;
}

.h-kanban-card__avatar {
  border: 2px solid var(--h-bg-card);
  margin-left: -8px;
  transition: transform var(--h-transition-fast);
}

.h-kanban-card__avatar:first-child {
  margin-left: 0;
}

.h-kanban-card__avatar:hover {
  transform: translateY(-2px);
  z-index: 1;
}

/* Due Date */
.h-kanban-card__due-date {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-1);
  font-size: var(--h-text-xs);
  color: var(--h-text-secondary);
}

.h-kanban-card__due-date-icon {
  font-size: var(--h-text-xs);
}

/* Dark Mode Support */
.dark .h-kanban-card {
  background: var(--h-bg-card);
}

.dark .h-kanban-card__title {
  color: var(--h-text-primary);
}

.dark .h-kanban-card__description {
  color: var(--h-text-secondary);
}

.dark .h-kanban-card__status-badge--backlog {
  background: rgba(116, 120, 140, 0.2);
  color: #ACA9B0;
}

.dark .h-kanban-card__status-badge--in-progress {
  background: rgba(255, 181, 71, 0.2);
  color: #FFB547;
}

.dark .h-kanban-card__status-badge--done {
  background: rgba(5, 205, 153, 0.2);
  color: #05CD99;
}

.dark .h-kanban-card__status-badge--urgent {
  background: rgba(238, 93, 80, 0.2);
  color: #EE5D50;
}

.dark .h-kanban-card__status-badge--default {
  background: rgba(116, 120, 140, 0.2);
  color: #ACA9B0;
}

.dark .h-kanban-card__due-date {
  color: var(--h-text-secondary);
}

.dark .h-kanban-card__avatar {
  border-color: var(--h-bg-card);
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .h-kanban-card {
    padding: var(--h-spacing-4);
  }
  
  .h-kanban-card__footer {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--h-spacing-2);
  }
  
  .h-kanban-card__avatar {
    width: 28px;
    height: 28px;
  }
}
</style>
