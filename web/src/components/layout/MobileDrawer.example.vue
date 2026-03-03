<template>
  <div class="mobile-drawer-example">
    <h2>MobileDrawer Example</h2>
    
    <!-- Button to open drawer -->
    <button @click="isDrawerOpen = true" class="open-button">
      Open Drawer
    </button>
    
    <!-- MobileDrawer Component -->
    <MobileDrawer v-model="isDrawerOpen">
      <div class="drawer-content">
        <h3>Drawer Content</h3>
        <p>This is the sidebar content that slides in from the left.</p>
        
        <nav class="drawer-nav">
          <a href="#" class="nav-item">Dashboard</a>
          <a href="#" class="nav-item">Monitoring Aktivitas</a>
          <a href="#" class="nav-item">Display (KDS)</a>
          <a href="#" class="nav-item">Menu & Komponen</a>
          <a href="#" class="nav-item">Supply Chain</a>
          <a href="#" class="nav-item">Logistik</a>
          <a href="#" class="nav-item">SDM</a>
          <a href="#" class="nav-item">Keuangan</a>
          <a href="#" class="nav-item">Sistem</a>
        </nav>
        
        <button @click="isDrawerOpen = false" class="close-button">
          Close Drawer
        </button>
      </div>
    </MobileDrawer>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import MobileDrawer from './MobileDrawer.vue'

const isDrawerOpen = ref(false)
</script>

<style scoped>
.mobile-drawer-example {
  padding: 24px;
}

.open-button {
  padding: 12px 24px;
  background-color: var(--h-primary, #5A4372);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  font-weight: 500;
}

.open-button:hover {
  background-color: var(--h-primary-hover, #4a3562);
}

.drawer-content {
  padding: 24px;
}

.drawer-content h3 {
  margin-top: 0;
  margin-bottom: 16px;
  color: var(--h-text-primary, #322837);
}

.drawer-content p {
  color: var(--h-text-secondary, #74788C);
  margin-bottom: 24px;
}

.drawer-nav {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 24px;
}

.nav-item {
  padding: 12px 16px;
  color: var(--h-text-primary, #322837);
  text-decoration: none;
  border-radius: 8px;
  transition: background-color 0.2s ease;
}

.nav-item:hover {
  background-color: var(--h-bg-light, #F4F7FE);
}

.close-button {
  padding: 12px 24px;
  background-color: var(--h-bg-light, #ACA9B0);
  color: var(--h-text-primary, #322837);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  width: 100%;
}

.close-button:hover {
  opacity: 0.8;
}

/* Dark mode support */
.dark .nav-item {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.dark .nav-item:hover {
  background-color: var(--h-bg-secondary-dark, #3D2B53);
}
</style>
