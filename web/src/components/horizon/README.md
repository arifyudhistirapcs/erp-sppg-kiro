# Horizon UI Components

Modern UI components following the Horizon UI design system with purple palette.

## Components

### HStatCard

Mini statistics card with icon, value, and change indicator.

**Usage:**
```vue
<HStatCard
  :icon="DollarOutlined"
  icon-bg="linear-gradient(135deg, #5A4372 0%, #3D2B53 100%)"
  label="Total Pendapatan"
  value="Rp 45.2M"
  change="+12.5%"
  change-type="increase"
  :loading="false"
/>
```

**Props:**
- `icon` (Object): Icon component from @ant-design/icons-vue
- `iconBg` (String): Custom gradient background for icon
- `label` (String, required): Label text
- `value` (String|Number, required): Value to display
- `change` (String): Change percentage (e.g., "+23%")
- `changeType` (String): "increase" or "decrease"
- `loading` (Boolean): Show skeleton loader

### HChartCard

Card wrapper for charts with header and actions.

**Usage:**
```vue
<HChartCard
  title="Weekly Revenue"
  subtitle="Last 7 days"
  :height="320"
  :loading="false"
>
  <template #header-right>
    <a-dropdown>
      <a-button>This month</a-button>
      <!-- dropdown menu -->
    </a-dropdown>
  </template>
  
  <!-- Chart content -->
  <div ref="chartRef" style="width: 100%; height: 100%;"></div>
</HChartCard>
```

**Props:**
- `title` (String, required): Chart card title
- `subtitle` (String): Optional subtitle text
- `height` (Number): Chart height in pixels (default: 320, mobile: 240)
- `loading` (Boolean): Show skeleton loader

**Slots:**
- `default`: Chart content area
- `header-right`: Custom actions (e.g., dropdown, buttons)

## Styling

All components use the Horizon UI design system variables defined in `styles/horizon/variables.css`:

- Colors: Purple palette (#5A4372, #3D2B53)
- Typography: DM Sans font family
- Spacing: 4px increments
- Border radius: 8-20px
- Shadows: Card shadow system

## Dark Mode

All components support dark mode via CSS variables. The `.dark` class on the root element switches the theme.

## Responsive

Components are mobile-responsive:
- Mobile (< 768px): Reduced sizes, stacked layouts
- Tablet (768px - 1024px): Optimized layouts
- Desktop (> 1024px): Full layouts

## Import

```javascript
import { HStatCard, HChartCard } from '@/components/horizon'
```
