/**
 * Horizon UI ECharts Theme Configuration
 * 
 * Custom ECharts theme matching the Horizon UI design system
 * with purple color palette and support for light/dark modes.
 */

// Color palette based on Horizon UI design system
const HORIZON_COLORS = {
  primary: '#5A4372',      // Primary Purple
  accent: '#3D2B53',       // Accent Purple
  mediumGray: '#74788C',   // Medium Gray
  lightGray: '#ACA9B0',    // Light Gray
  darkText: '#322837',     // Dark text (light mode)
  lightText: '#F8FDEA',    // Light text (dark mode)
  success: '#05CD99',      // Success green
  warning: '#FFB547',      // Warning orange
  error: '#EE5D50',        // Error red
}

/**
 * Light theme configuration for ECharts
 */
export const horizonChartThemeLight = {
  // Color palette for series
  color: [
    HORIZON_COLORS.primary,
    HORIZON_COLORS.accent,
    HORIZON_COLORS.mediumGray,
    HORIZON_COLORS.lightGray,
  ],

  // Transparent background to inherit from card
  backgroundColor: 'transparent',

  // Global text style
  textStyle: {
    fontFamily: 'DM Sans, -apple-system, BlinkMacSystemFont, sans-serif',
    fontSize: 14,
    color: HORIZON_COLORS.darkText,
  },

  // Grid configuration
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },

  // Axis styling
  categoryAxis: {
    axisLine: {
      lineStyle: {
        color: '#E9EDF7',
      },
    },
    axisTick: {
      lineStyle: {
        color: '#E9EDF7',
      },
    },
    axisLabel: {
      color: HORIZON_COLORS.mediumGray,
      fontSize: 12,
    },
    splitLine: {
      lineStyle: {
        color: '#E9EDF7',
        type: 'dashed',
      },
    },
  },

  valueAxis: {
    axisLine: {
      show: false,
    },
    axisTick: {
      show: false,
    },
    axisLabel: {
      color: HORIZON_COLORS.mediumGray,
      fontSize: 12,
    },
    splitLine: {
      lineStyle: {
        color: '#E9EDF7',
        type: 'dashed',
      },
    },
  },

  // Tooltip styling
  tooltip: {
    backgroundColor: '#FFFFFF',
    borderColor: '#E9EDF7',
    borderWidth: 1,
    textStyle: {
      color: HORIZON_COLORS.darkText,
      fontSize: 14,
    },
    padding: [12, 16],
    extraCssText: 'box-shadow: 0px 4px 6px rgba(0, 0, 0, 0.07); border-radius: 8px;',
  },

  // Legend styling
  legend: {
    textStyle: {
      color: HORIZON_COLORS.darkText,
      fontSize: 14,
    },
    icon: 'circle',
    itemWidth: 10,
    itemHeight: 10,
    itemGap: 16,
  },

  // Title styling
  title: {
    textStyle: {
      color: HORIZON_COLORS.darkText,
      fontSize: 18,
      fontWeight: 700,
    },
    subtextStyle: {
      color: HORIZON_COLORS.mediumGray,
      fontSize: 14,
    },
  },

  // Line series styling
  line: {
    smooth: true,
    lineStyle: {
      width: 2,
    },
    symbolSize: 6,
    emphasis: {
      scale: true,
      scaleSize: 10,
    },
  },

  // Bar series styling
  bar: {
    barMaxWidth: 40,
    itemStyle: {
      borderRadius: [4, 4, 0, 0],
    },
  },

  // Pie series styling
  pie: {
    itemStyle: {
      borderRadius: 4,
      borderColor: '#FFFFFF',
      borderWidth: 2,
    },
    label: {
      color: HORIZON_COLORS.darkText,
      fontSize: 14,
    },
  },

  // Scatter series styling
  scatter: {
    symbolSize: 8,
  },
}

/**
 * Dark theme configuration for ECharts
 */
export const horizonChartThemeDark = {
  // Color palette for series (same as light)
  color: [
    HORIZON_COLORS.primary,
    HORIZON_COLORS.accent,
    HORIZON_COLORS.mediumGray,
    HORIZON_COLORS.lightGray,
  ],

  // Transparent background to inherit from card
  backgroundColor: 'transparent',

  // Global text style (light text for dark mode)
  textStyle: {
    fontFamily: 'DM Sans, -apple-system, BlinkMacSystemFont, sans-serif',
    fontSize: 14,
    color: HORIZON_COLORS.lightText,
  },

  // Grid configuration (same as light)
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true,
  },

  // Axis styling (dark mode colors)
  categoryAxis: {
    axisLine: {
      lineStyle: {
        color: HORIZON_COLORS.primary,
      },
    },
    axisTick: {
      lineStyle: {
        color: HORIZON_COLORS.primary,
      },
    },
    axisLabel: {
      color: HORIZON_COLORS.lightGray,
      fontSize: 12,
    },
    splitLine: {
      lineStyle: {
        color: HORIZON_COLORS.primary,
        type: 'dashed',
      },
    },
  },

  valueAxis: {
    axisLine: {
      show: false,
    },
    axisTick: {
      show: false,
    },
    axisLabel: {
      color: HORIZON_COLORS.lightGray,
      fontSize: 12,
    },
    splitLine: {
      lineStyle: {
        color: HORIZON_COLORS.primary,
        type: 'dashed',
      },
    },
  },

  // Tooltip styling (dark mode)
  tooltip: {
    backgroundColor: HORIZON_COLORS.accent,
    borderColor: HORIZON_COLORS.primary,
    borderWidth: 1,
    textStyle: {
      color: HORIZON_COLORS.lightText,
      fontSize: 14,
    },
    padding: [12, 16],
    extraCssText: 'box-shadow: 0px 4px 6px rgba(0, 0, 0, 0.3); border-radius: 8px;',
  },

  // Legend styling (dark mode)
  legend: {
    textStyle: {
      color: HORIZON_COLORS.lightText,
      fontSize: 14,
    },
    icon: 'circle',
    itemWidth: 10,
    itemHeight: 10,
    itemGap: 16,
  },

  // Title styling (dark mode)
  title: {
    textStyle: {
      color: HORIZON_COLORS.lightText,
      fontSize: 18,
      fontWeight: 700,
    },
    subtextStyle: {
      color: HORIZON_COLORS.lightGray,
      fontSize: 14,
    },
  },

  // Line series styling (same as light)
  line: {
    smooth: true,
    lineStyle: {
      width: 2,
    },
    symbolSize: 6,
    emphasis: {
      scale: true,
      scaleSize: 10,
    },
  },

  // Bar series styling (same as light)
  bar: {
    barMaxWidth: 40,
    itemStyle: {
      borderRadius: [4, 4, 0, 0],
    },
  },

  // Pie series styling (dark mode)
  pie: {
    itemStyle: {
      borderRadius: 4,
      borderColor: HORIZON_COLORS.accent,
      borderWidth: 2,
    },
    label: {
      color: HORIZON_COLORS.lightText,
      fontSize: 14,
    },
  },

  // Scatter series styling (same as light)
  scatter: {
    symbolSize: 8,
  },
}

/**
 * Get the appropriate theme based on dark mode state
 * @param {boolean} isDark - Whether dark mode is active
 * @returns {Object} ECharts theme configuration
 */
export function getHorizonChartTheme(isDark = false) {
  return isDark ? horizonChartThemeDark : horizonChartThemeLight
}

// Default export is light theme
export default horizonChartThemeLight
