import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { useDarkMode } from './useDarkMode'

/**
 * Composable untuk mengelola ECharts instance dengan Horizon UI theme
 */
export function useHorizonChart(chartRef, options = {}) {
  const chart = ref(null)
  const { isDark } = useDarkMode()
  let lastOptions = null

  const getThemeColors = (dark) => {
    if (dark) {
      return {
        textColor: '#ACA9B0',
        axisLineColor: '#5A4372',
        splitLineColor: 'rgba(90, 67, 114, 0.3)',
        tooltipBg: '#3D2B53',
        tooltipBorder: '#5A4372',
        tooltipTextColor: '#F8FDEA',
        legendTextColor: '#F8FDEA',
      }
    }
    return {
      textColor: '#74788C',
      axisLineColor: '#E9EDF7',
      splitLineColor: '#E9EDF7',
      tooltipBg: '#FFFFFF',
      tooltipBorder: '#E9EDF7',
      tooltipTextColor: '#322837',
      legendTextColor: '#322837',
    }
  }

  const applyTheme = (chartOptions) => {
    const t = getThemeColors(isDark.value)

    // Only apply theme to standard echarts option keys
    const themed = { ...chartOptions }

    // Apply tooltip theme
    if (themed.tooltip) {
      themed.tooltip = {
        ...themed.tooltip,
        backgroundColor: t.tooltipBg,
        borderColor: t.tooltipBorder,
        textStyle: { color: t.tooltipTextColor, fontSize: 13, ...(themed.tooltip.textStyle || {}) },
      }
    }

    // Apply legend theme
    if (themed.legend) {
      themed.legend = {
        ...themed.legend,
        textStyle: { color: t.legendTextColor, ...(themed.legend.textStyle || {}) },
      }
    }

    // Apply xAxis theme
    if (themed.xAxis) {
      const ax = Array.isArray(themed.xAxis) ? themed.xAxis : [themed.xAxis]
      themed.xAxis = ax.map(a => ({
        ...a,
        axisLine: { lineStyle: { color: t.axisLineColor }, ...(a.axisLine || {}) },
        axisLabel: { color: t.textColor, fontSize: 12, ...(a.axisLabel || {}) },
        splitLine: { lineStyle: { color: t.splitLineColor, type: 'dashed' }, ...(a.splitLine || {}) },
      }))
      if (themed.xAxis.length === 1) themed.xAxis = themed.xAxis[0]
    }

    // Apply yAxis theme
    if (themed.yAxis) {
      const ay = Array.isArray(themed.yAxis) ? themed.yAxis : [themed.yAxis]
      themed.yAxis = ay.map(a => ({
        ...a,
        axisLine: { show: false, ...(a.axisLine || {}) },
        axisTick: { show: false, ...(a.axisTick || {}) },
        axisLabel: { color: t.textColor, fontSize: 12, ...(a.axisLabel || {}) },
        splitLine: { lineStyle: { color: t.splitLineColor, type: 'dashed' }, ...(a.splitLine || {}) },
      }))
      if (themed.yAxis.length === 1) themed.yAxis = themed.yAxis[0]
    }

    // Global text style
    themed.textStyle = {
      fontFamily: 'DM Sans, -apple-system, BlinkMacSystemFont, sans-serif',
      color: t.textColor,
      ...(chartOptions.textStyle || {}),
    }

    themed.backgroundColor = 'transparent'

    return themed
  }

  const initChart = () => {
    if (!chartRef.value) return
    if (chart.value) chart.value.dispose()
    chart.value = echarts.init(chartRef.value)
    if (options && Object.keys(options).length > 0) {
      setOption(options)
    }
  }

  const setOption = (chartOptions, notMerge = false) => {
    if (!chart.value) {
      // Try to init if ref is available
      if (chartRef.value) {
        initChart()
      }
      if (!chart.value) return
    }
    lastOptions = chartOptions
    const themed = applyTheme(chartOptions)
    chart.value.setOption(themed, notMerge)
  }

  const updateTheme = () => {
    if (!chart.value || !lastOptions) return
    const themed = applyTheme(lastOptions)
    chart.value.setOption(themed, true)
  }

  const resize = () => {
    if (chart.value) chart.value.resize()
  }

  onMounted(() => {
    initChart()
    window.addEventListener('resize', resize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', resize)
    if (chart.value) {
      chart.value.dispose()
      chart.value = null
    }
  })

  watch(isDark, () => {
    updateTheme()
  })

  return { chart, setOption, updateTheme, resize, initChart }
}
