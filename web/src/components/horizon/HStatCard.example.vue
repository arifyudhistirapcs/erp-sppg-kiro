<template>
  <div class="example-container">
    <h1>HStatCard Component Examples</h1>
    
    <div class="example-section">
      <h2>Basic Usage</h2>
      <div class="card-grid">
        <HStatCard
          :icon="DollarOutlined"
          icon-bg="linear-gradient(135deg, #05CD99 0%, #26d9a8 100%)"
          label="Total Pendapatan"
          value="Rp 45.2M"
          change="+12.5%"
          change-type="increase"
        />
        
        <HStatCard
          :icon="ShoppingOutlined"
          icon-bg="linear-gradient(135deg, #FFB547 0%, #ffc266 100%)"
          label="Total Pesanan"
          value="154"
          change="+8%"
          change-type="increase"
        />
        
        <HStatCard
          :icon="TeamOutlined"
          icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
          label="Karyawan Aktif"
          value="42"
          change="-2%"
          change-type="decrease"
        />
        
        <HStatCard
          :icon="CheckCircleOutlined"
          icon-bg="linear-gradient(135deg, #EE5D50 0%, #f17a6f 100%)"
          label="Tasks Selesai"
          value="89%"
          change="+5%"
          change-type="increase"
        />
      </div>
    </div>
    
    <div class="example-section">
      <h2>Without Change Indicator</h2>
      <div class="card-grid">
        <HStatCard
          :icon="HomeOutlined"
          icon-bg="linear-gradient(135deg, #5A4372 0%, #6a5382 100%)"
          label="Total Sekolah"
          value="25"
        />
      </div>
    </div>
    
    <div class="example-section">
      <h2>Loading State</h2>
      <div class="card-grid">
        <HStatCard
          :icon="DollarOutlined"
          label="Loading..."
          value="0"
          :loading="true"
        />
      </div>
    </div>
    
    <div class="example-section">
      <h2>Responsive Grid (Try resizing)</h2>
      <p>On mobile (< 768px), cards stack vertically. On desktop, they display in a grid.</p>
    </div>
  </div>
</template>

<script setup>
import HStatCard from './HStatCard.vue'
import { 
  DollarOutlined, 
  ShoppingOutlined, 
  TeamOutlined, 
  CheckCircleOutlined,
  HomeOutlined
} from '@ant-design/icons-vue'
</script>

<style scoped>
.example-container {
  padding: var(--h-spacing-8);
  background: var(--h-bg-primary);
  min-height: 100vh;
}

h1 {
  font-size: var(--h-text-3xl);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  margin-bottom: var(--h-spacing-8);
}

h2 {
  font-size: var(--h-text-xl);
  font-weight: var(--h-font-semibold);
  color: var(--h-text-primary);
  margin-bottom: var(--h-spacing-4);
}

.example-section {
  margin-bottom: var(--h-spacing-10);
}

.example-section p {
  color: var(--h-text-secondary);
  margin-bottom: var(--h-spacing-4);
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--h-spacing-5);
}

/* Mobile */
@media (max-width: 767px) {
  .card-grid {
    grid-template-columns: 1fr;
  }
}
</style>
