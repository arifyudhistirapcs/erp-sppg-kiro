<template>
  <header class="h-header" :class="{ 'mobile': isMobile }">
    <!-- Mobile: Hamburger Menu Button -->
    <button
      v-if="isMobile"
      class="hamburger-button"
      @click="$emit('toggle-drawer')"
      aria-label="Toggle menu"
      type="button"
    >
      <MenuOutlined />
    </button>

    <!-- Breadcrumb & Title Section -->
    <div class="header-left">
      <!-- Breadcrumb (hidden on mobile) -->
      <div v-if="!isMobile && breadcrumb" class="breadcrumb">
        <span class="breadcrumb-parent">{{ breadcrumb.parent }}</span>
        <span class="breadcrumb-separator">/</span>
        <span class="breadcrumb-current">{{ breadcrumb.current }}</span>
      </div>
      
      <!-- Page Title -->
      <h1 class="page-title">{{ pageTitle }}</h1>
    </div>

    <!-- Right Section: Theme Toggle only -->
    <div class="header-right">
      <ThemeToggle />
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useBreakpoint } from '@/composables/useBreakpoint'
import { MenuOutlined } from '@ant-design/icons-vue'
import ThemeToggle from './ThemeToggle.vue'

const props = defineProps({
  pageTitle: {
    type: String,
    default: 'Dashboard'
  },
  breadcrumb: {
    type: Object,
    default: () => ({ parent: 'Pages', current: 'Dashboard' })
  }
})

defineEmits(['toggle-drawer'])

const { isMobile } = useBreakpoint()
</script>

<style scoped>
.h-header {
  display: flex;
  align-items: center;
  height: 88px;
  background-color: var(--h-bg-secondary, #FFFFFF);
  padding: 12px 24px;
  box-shadow: 0px 2px 4px rgba(0, 0, 0, 0.05);
  position: relative;
  transition: all 0.2s ease;
}

.h-header.mobile {
  height: 68px;
  padding: 8px 16px;
}

.dark .h-header {
  background-color: var(--h-bg-secondary-dark, #111C44);
  box-shadow: 0px 2px 4px rgba(0, 0, 0, 0.2);
}

.hamburger-button {
  width: 44px;
  height: 44px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  color: var(--h-text-primary, #322837);
  border-radius: 8px;
  transition: all 0.2s ease;
  margin-right: 12px;
}

.hamburger-button:hover {
  background-color: var(--h-bg-light, #F4F7FE);
}

.hamburger-button:active {
  transform: scale(0.95);
}

.dark .hamburger-button {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.dark .hamburger-button:hover {
  background-color: var(--h-bg-card-dark, #1B254B);
}

.header-left {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--h-text-secondary, #74788C);
}

.breadcrumb-parent {
  color: var(--h-text-secondary, #74788C);
}

.breadcrumb-separator {
  color: var(--h-text-secondary, #74788C);
}

.breadcrumb-current {
  color: var(--h-text-primary, #322837);
  font-weight: 500;
}

.dark .breadcrumb-current {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--h-text-primary, #322837);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mobile .page-title {
  font-size: 18px;
}

.dark .page-title {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

@media (max-width: 768px) {
  .header-right {
    gap: 8px;
  }
}
</style>