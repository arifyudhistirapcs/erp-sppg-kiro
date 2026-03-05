<template>
  <div class="mobile-layout">
    <div class="mobile-content">
      <router-view v-slot="{ Component }">
        <transition name="page-slide" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </div>
    <BottomNavigation />
  </div>
</template>

<script setup>
import BottomNavigation from '@/components/mobile/BottomNavigation.vue'
</script>

<style scoped>
.mobile-layout {
  min-height: 100vh;
  background: var(--h-bg-primary);
}

.mobile-content {
  padding-bottom: 80px;
}

/* Page slide transition - 200ms */
.page-slide-enter-active,
.page-slide-leave-active {
  transition: transform 200ms ease-in-out, opacity 200ms ease-in-out;
}

.page-slide-enter-from {
  transform: translateX(20px);
  opacity: 0;
}

.page-slide-leave-to {
  transform: translateX(-20px);
  opacity: 0;
}
</style>
