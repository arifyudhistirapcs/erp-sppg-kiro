<template>
  <button
    class="theme-toggle"
    :class="{ 'theme-toggle--dark': isDark }"
    @click="toggle"
    :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
    :title="isDark ? 'Light Mode' : 'Dark Mode'"
    type="button"
  >
    <!-- Sun icon (dark mode - click to go light) -->
    <span v-if="isDark" class="theme-icon-wrap">
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <circle cx="12" cy="12" r="5" stroke="#FFD700" stroke-width="2.5"/>
        <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42" stroke="#FFD700" stroke-width="2.5" stroke-linecap="round"/>
      </svg>
    </span>
    <!-- Moon icon (light mode - click to go dark) -->
    <span v-else class="theme-icon-wrap">
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke="#5A4372" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
    </span>
  </button>
</template>

<script setup>
import { useDarkMode } from '@/composables/useDarkMode'

const { isDark, toggle } = useDarkMode()
</script>

<style scoped>
.theme-toggle {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  border: 1px solid rgba(90, 67, 114, 0.25);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  background-color: #F4F7FE;
  padding: 0;
  flex-shrink: 0;
}

.theme-toggle--dark {
  background-color: #1B254B;
  border-color: rgba(255, 215, 0, 0.35);
}

.theme-toggle:hover {
  transform: scale(1.08);
  box-shadow: 0px 4px 8px rgba(0, 0, 0, 0.15);
}

.theme-toggle:active {
  transform: scale(0.95);
}

.theme-toggle:focus-visible {
  outline: 2px solid #5A4372;
  outline-offset: 2px;
}

.theme-icon-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  line-height: 0;
}

.theme-icon-wrap svg {
  display: block;
}
</style>
