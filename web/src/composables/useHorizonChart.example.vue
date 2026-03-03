<!--
  Example usage of useHorizonChart composable
  
  This example demonstrates how to create a simple line chart
  with automatic theme switching and resize handling.
-->
<template>
  <div class="chart-example">
    <h3>Example: Line Chart with Horizon Theme</h3>
    
    <!-- Chart container -->
    <div ref="chartRef" class="chart-container"></div>
    
    <!-- Controls -->
    <div class="controls">
      <button @click="updateData">Update Data</button>
      <button @click="toggleTheme">Toggle Theme</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useHorizonChart } from './useHorizonChart'
import { useDarkMode } from './useDarkMode'

// Chart ref
const chartRef = ref(null)

// Dark mode
const { isDark, toggle: toggleTheme } = useDarkMode()

// Chart options
const chartOptions = {
  title: {
    text: 'Weekly Revenue',
    left: 'center'
  },
  tooltip: {
    trigger: 'axis'
  },
  xAxis: {
    type: 'category',
    data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
  },
  yAxis: {
    type: 'value'
  },
  series: [
    {
      name: 'Revenue',
      type: 'line',
      data: [120, 200, 150, 80, 70, 110, 130],
      smooth: true,
      areaStyle: {
        opacity: 0.3
      }
    }
  ]
}

// Initialize chart
const { chart, setOption } = useHorizonChart(chartRef, chartOptions)

// Update data function
const updateData = () => {
  const newData = Array.from({ length: 7 }, () => Math.floor(Math.random() * 200) + 50)
  
  setOption({
    series: [
      {
        name: 'Revenue',
        type: 'line',
        data: newData,
        smooth: true,
        areaStyle: {
          opacity: 0.3
        }
      }
    ]
  })
}
</script>

<style scoped>
.chart-example {
  padding: 24px;
}

.chart-container {
  width: 100%;
  height: 400px;
  margin: 20px 0;
}

.controls {
  display: flex;
  gap: 12px;
}

button {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid var(--h-border-color, #E9EDF7);
  background: var(--h-bg-card, #FFFFFF);
  color: var(--h-text-primary, #322837);
  cursor: pointer;
  transition: all 0.2s ease;
}

button:hover {
  background: var(--h-primary, #5A4372);
  color: white;
  transform: scale(1.02);
}
</style>
