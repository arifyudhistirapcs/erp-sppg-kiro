# Implementation Tasks - UI Redesign Horizon

## Phase 1: Foundation Setup (Week 1-2)

- [x] 1. Setup CSS Variables & Theme System
  - [x] 1.1 Create `web/src/styles/horizon/variables.css` dengan semua CSS variables: colors (primary #5A4372, accent #3D2B53, neutrals, success/warning/error), spacing (4px-48px), typography (DM Sans, sizes 12-36px, weights 400-700), border-radius (8-20px, full), shadows (sm/md/lg/xl/card), transitions (fast 150ms, base 200ms, slow 300ms), breakpoints (375px-1536px)
  - [x] 1.2 Create `web/src/styles/horizon/dark-mode.css` dengan dark mode overrides menggunakan `.dark` class selector: bg-primary #322837, bg-card #3D2B53, text-primary #F8FDEA, border-color #5A4372
  - [x] 1.3 Create `web/src/styles/horizon/base.css` dengan reset styles, font imports (DM Sans dari Google Fonts), dan base element styling sesuai design system
  - [x] 1.4 Create `web/src/styles/horizon/index.css` sebagai entry point yang import semua horizon style files, lalu import di `web/src/main.js`

- [x] 2. Create Base Utility Classes
  - [x] 2.1 Create `web/src/styles/horizon/utilities.css` dengan utility classes: `.h-card` (bg, radius 16px, shadow card, padding 24px), `.h-card-hover` (translateY -4px + shadow increase on hover), `.h-gradient-primary` (linear-gradient 135deg #5A4372 to #3D2B53), `.h-text-primary`, `.h-text-secondary`, `.h-button` (transition + scale on hover)
  - [x] 2.2 Create `web/src/styles/horizon/responsive.css` dengan responsive utility classes: `.hidden-mobile` (hidden < 768px), `.show-mobile` (shown < 768px), `.hidden-tablet`, `.show-tablet`, `.touch-target` (min 44x44px), `.tap-highlight`, mobile spacing utilities, dan media queries per breakpoint dari requirements

- [x] 3. Implement Dark Mode Composable
  - [x] 3.1 Create `web/src/composables/useDarkMode.js`: reactive `isDark` ref, load dari localStorage key 'theme', fallback ke `prefers-color-scheme: dark` system preference, `toggle()` function, watch isDark → add/remove `.dark` class pada `document.documentElement` + persist ke localStorage, smooth transition support

- [x] 4. Create Theme Toggle Component
  - [x] 4.1 Create `web/src/components/layout/ThemeToggle.vue`: button 40x40px, border-radius 12px, bg #F4F7FE (light) / #1B254B (dark), Sun icon (light mode) / Moon icon (dark mode) dengan smooth icon transition/animation, integrates dengan `useDarkMode` composable, keyboard accessible (Enter/Space), aria-label "Toggle dark mode"

- [x] 5. Setup Breakpoint Composable
  - [x] 5.1 Create `web/src/composables/useBreakpoint.js`: reactive `breakpoint` ref, `updateBreakpoint()` berdasarkan window.innerWidth (< 640 = 'sm', < 768 = 'md', < 1024 = 'lg', < 1280 = 'xl', else '2xl'), computed properties `isMobile` (sm/md), `isTablet` (lg), `isDesktop` (xl/2xl), addEventListener resize + cleanup di onUnmounted

- [x] 6. Create Mobile Drawer Component
  - [x] 6.1 Create `web/src/components/layout/MobileDrawer.vue`: fixed position, width 280px, slide-in dari kiri (transform translateX), overlay backdrop rgba(0,0,0,0.5) dengan tap-to-close, transition 300ms ease, z-index 1000, slot untuk sidebar content, v-model untuk open/close state, swipe-to-close gesture support

## Phase 2: Core Components (Week 3-4)

- [x] 7. Create HStatCard Component
  - [x] 7.1 Create `web/src/components/horizon/HStatCard.vue`: props (icon, iconBg, label, value, change, changeType, loading), icon container 56x56px dengan gradient background + border-radius 12px, label 14px #74788C font-weight 500, value 24px #322837 font-weight 700, change indicator 12px dengan arrow icon (green increase / red decrease), loading skeleton state, responsive (full width mobile / auto desktop), dark mode support via CSS variables
  - [x] 7.2 Create `web/src/components/horizon/index.js` sebagai barrel export untuk semua Horizon components

- [x] 8. Create HChartCard Component
  - [x] 8.1 Create `web/src/components/horizon/HChartCard.vue`: props (title, subtitle, height default 320, actions, loading), h-card styling, header dengan title 18px bold + subtitle 14px light + slot `header-right` untuk custom actions, default slot untuk chart content, loading skeleton state, responsive (reduced height 240px on mobile), dark mode support

- [x] 9. Setup ECharts Horizon Theme
  - [x] 9.1 Create `web/src/utils/horizonChartTheme.js`: ECharts theme object dengan color palette [#5A4372, #3D2B53, #74788C, #ACA9B0], transparent background, fontFamily 'DM Sans', fontSize 14, textStyle color #322837, grid config (left 3%, right 4%, bottom 3%, containLabel true), axis/tooltip/legend styling, export light dan dark variants
  - [x] 9.2 Create `web/src/composables/useHorizonChart.js`: composable yang init ECharts instance pada ref element, apply horizon theme, handle resize, cleanup di onUnmounted, support dark mode switching

- [x] 10. Create HDataTable Component
  - [x] 10.1 Create `web/src/components/horizon/HDataTable.vue`: wrapper around Ant Design Table, props (columns, dataSource, loading, pagination, rowSelection, mobileCardView), h-card container, custom styling: row height 56px, header text 12px uppercase #74788C font-weight 700, cell text 14px #322837, hover bg #F8FDEA, status badge rendering (colored dot + text, radius 8px, padding 4px 12px), progress bar rendering, action buttons styling, mobile card view option (stack rows as cards when < 768px), horizontal scroll with indicator on mobile, larger action buttons on mobile (44px touch targets), dark mode support

- [x] 11. Create HSidebar Component
  - [x] 11.1 Create `web/src/components/layout/HSidebar.vue`: width 280px expanded / 80px collapsed, bg white (light) / #111C44 (dark), logo area 64px height, menu items 44px height each, icon 20px, font 14px medium, active state (purple bg + white text), hover state (#F4F7FE bg), collapsible toggle button, nested menu support dengan expand/collapse, role-based menu filtering via `usePermissions`, smooth width transition, mobile: hidden by default (opens via MobileDrawer), ERP SPPG menu structure sesuai requirements (Dashboard, Monitoring Aktivitas, Display/KDS, Menu & Komponen, Supply Chain, Logistik, SDM, Keuangan, Sistem)

- [x] 12. Create HHeader Component
  - [x] 12.1 Create `web/src/components/layout/HHeader.vue`: height 72px, bg white/dark, padding 0 24px, breadcrumb navigation (Pages / Current Page), page title 24px bold, search bar (320px, height 40px, radius 12px, bg #F4F7FE, search icon left, placeholder "Search"), notification icon dengan badge, ThemeToggle integration, user avatar dropdown menu, mobile: reduced height 56px, hamburger menu button 44x44px (left), hide breadcrumb show title only, collapsible search (icon → full width)
  - [x] 12.2 Create `web/src/components/layout/HBreadcrumb.vue`: breadcrumb component yang reads current route, renders "Pages / Page Name" format, styled sesuai Horizon UI (14px, #74788C), hidden on mobile

- [x] 13. Create HorizonLayout Component
  - [x] 13.1 Create `web/src/layouts/HorizonLayout.vue`: wrapper combining HSidebar + HHeader + content area, `.horizon-layout` class dengan dark mode binding, content area bg #F4F7FE (light) / #0B1437 (dark), padding 24px, gap 20px, mobile: full width content, reduced padding 16px, drawer overlay management, integrates useDarkMode + useBreakpoint composables, smooth transitions

- [x] 14. Create HKanbanCard Component
  - [x] 14.1 Create `web/src/components/horizon/HKanbanCard.vue`: props (title, description, image, status, assignees, dueDate), width 100%, min-height 120px, radius 16px, shadow md, padding 20px, title 16px bold #322837, description 14px #74788C, optional image (full width, radius 12px), assignee avatars 32px overlapping (-8px margin), status badge (colored: Backlog gray, In Progress orange, Done green, Urgent red), drag handle optional, dark mode support

- [ ] 15. Mobile Table Card View Component
  - [ ] 15.1 Create `web/src/components/horizon/HMobileTableCard.vue`: converts table row data to card format, props (fields, data, actions), stack fields vertically dengan label + value pairs, action buttons at bottom, h-card styling, works as alternative view inside HDataTable when mobile

- [ ] 16. Mobile Bottom Navigation (Optional)
  - [ ] 16.1 Create `web/src/components/layout/MobileBottomNav.vue`: fixed bottom position, 4-5 main navigation items (Dashboard, KDS, Menu, Inventory, More), active state highlighting, icons + labels, safe area insets (iOS padding-bottom), only visible on mobile (< 768px), z-index above content below drawer

## Phase 3: Dashboard Pages (Week 5-6)

- [x] 17. Redesign Dashboard Kepala SPPG
  - [x] 17.1 Refactor `web/src/views/DashboardKepalaSSPGView.vue` menggunakan HorizonLayout, stats cards row (4 HStatCard: Pendapatan, Pengeluaran, Pesanan, Tasks) dengan data dari existing dashboardService API, charts row (HChartCard: Weekly Revenue line chart + Daily Traffic bar chart) menggunakan useHorizonChart + horizonChartTheme, tables row (2 HDataTable: Recent Orders + Recent Activities), mobile: stats stack 1 column, charts full width reduced height, tables card view, semua interactive elements 44px+ touch targets, dark mode support

- [x] 18. Redesign Dashboard Kepala Yayasan
  - [x] 18.1 Refactor `web/src/views/DashboardKepalaYayasanView.vue` menggunakan HorizonLayout, financial overview HStatCards, multi-location comparison charts (HChartCard), summary tables (HDataTable), data dari existing API, responsive layout, dark mode support

- [x] 19. Redesign Monitoring Aktivitas Page
  - [x] 19.1 Refactor `web/src/views/ActivityTrackerListView.vue` menggunakan HorizonLayout, status HStatCards (Pending, In Progress, Completed counts), activity list dengan HDataTable, filter dan search functionality, real-time updates via Firebase, responsive, dark mode
  - [x] 19.2 Refactor `web/src/views/ActivityTrackerDetailView.vue` menggunakan HorizonLayout, activity timeline display, detail cards, responsive, dark mode

## Phase 4: Feature Pages (Week 7-8)

- [x] 20. Redesign KDS Cooking Page
  - [x] 20.1 Refactor `web/src/views/KDSCookingView.vue` menggunakan HorizonLayout, status HStatCards (Pending/In Progress/Done counts), kanban board layout dengan 3 columns (Pending, In Progress, Done), HKanbanCard untuk setiap cooking task, keep existing drag-drop functionality, timer components, real-time Firebase updates, responsive, dark mode

- [x] 21. Redesign KDS Packing Page
  - [x] 21.1 Refactor `web/src/views/KDSPackingView.vue` menggunakan HorizonLayout, similar kanban layout to KDS Cooking, packing-specific fields, school allocation display, real-time updates, responsive, dark mode

- [x] 22. Redesign KDS Cleaning Page
  - [x] 22.1 Refactor `web/src/views/KDSCleaningView.vue` menggunakan HorizonLayout, checklist-style cards, status indicators, real-time updates, responsive, dark mode

- [x] 23. Redesign Menu Planning Page
  - [x] 23.1 Refactor `web/src/views/MenuPlanningView.vue` menggunakan HorizonLayout, calendar view dengan Horizon styling (h-card), recipe cards dengan images, keep existing drag-drop to calendar, nutrition info display, responsive, dark mode

- [x] 24. Redesign Recipe List Page
  - [x] 24.1 Refactor `web/src/views/RecipeListView.vue` menggunakan HorizonLayout, grid layout dengan recipe cards (h-card + image thumbnails), quick actions (edit, delete, view), search dan filter bar, responsive grid (1 col mobile, 2 col tablet, 3-4 col desktop), dark mode

- [x] 25. Redesign Inventory Page
  - [x] 25.1 Refactor `web/src/views/InventoryView.vue` menggunakan HorizonLayout, stats HStatCards (Total Items, Low Stock, Total Value), HDataTable untuk inventory list, stock level indicators (progress bars colored by level), alert cards untuk low stock items, trend charts (HChartCard), responsive, dark mode

## Phase 5: Forms, Tables & Polish (Week 9-10)

- [x] 26. Create HFormCard Component
  - [x] 26.1 Create `web/src/components/horizon/HFormCard.vue`: form wrapper dengan h-card styling, props (title, subtitle), section headers styling, better spacing (16px between fields), input height 48px, validation state styling (error border red, success border green), dark mode support

- [x] 27. Update All Form Pages
  - [x] 27.1 Refactor form pages menggunakan HorizonLayout + HFormCard: `EmployeeFormView.vue`, `SchoolFormView.vue`, `DeliveryTaskFormView.vue`, dan form modals (AssetFormModal, CashFlowFormModal, IngredientFormModal, RecipeFormModal, SemiFinishedFormModal, PickupTaskForm, StokOpnameForm), better spacing dan typography, validation states styled, responsive (full width inputs, 48px height, sticky footer for actions), dark mode

- [x] 28. Update All List/Table Pages
  - [x] 28.1 Refactor list pages menggunakan HorizonLayout + HDataTable: `EmployeeListView.vue`, `SchoolListView.vue`, `SupplierListView.vue`, `PurchaseOrderListView.vue`, `AssetListView.vue`, `CashFlowListView.vue`, `DeliveryTaskListView.vue`, `GoodsReceiptView.vue`, `IngredientListView.vue`, `SemiFinishedGoodsView.vue`, `AuditTrailView.vue`, `AttendanceReportView.vue`, `FinancialReportView.vue`, action buttons styled, status badges, responsive (card view on mobile), dark mode

- [x] 29. Redesign Login Page ✅ IMPLEMENTED
  - [x] 29.1 Refactor `web/src/views/LoginView.vue`: split screen layout 50/50, left side (white bg, max-width 480px, centered, padding 48px): title "Sign In" 36px bold, subtitle 14px, input fields (48px height, radius 12px), "Keep me logged in" checkbox, submit button (full width, purple gradient #5A4372 to #3D2B53, 48px). Right side: gradient bg (#5A4372 to #3D2B53), ERP SPPG branding centered with decorative circles. Responsive: hide right side on tablet/mobile (< 1024px), reduced padding on mobile, dark mode support

- [x] 30. Create HProfileCard Component
  - [x] 30.1 Create `web/src/components/horizon/HProfileCard.vue`: gradient banner 200px height (purple gradient, radius 16px top), avatar 80px centered overlapping banner (4px white border), stats row (3 columns: value 24px bold + label 14px light), content area h-card styling, dark mode support

- [x] 31. Add Page Transitions
  - [x] 31.1 Create `web/src/styles/horizon/animations.css`: page transition classes (.page-enter-active/.page-leave-active transition 200ms opacity + 300ms transform, .page-enter-from opacity 0 translateX 20px, .page-leave-to opacity 0 translateX -20px), card hover animations, button hover scale
  - [x] 31.2 Update `web/src/App.vue`: wrap `<router-view>` dengan `<Transition name="page">`, import animations.css

- [x] 32. Optimize Performance
  - [x] 32.1 Implement code splitting: lazy load route components di `web/src/router/index.js` menggunakan `() => import()`, lazy load heavy Horizon components (HDataTable, HChartCard), optimize image loading (lazy load), verify no performance regressions

- [x] 33. Documentation
  - [x] 33.1 Create `docs/HORIZON_UI_GUIDE.md`: design system overview, color palette, typography, spacing, component usage guide dengan code examples untuk setiap Horizon component (HStatCard, HChartCard, HDataTable, HSidebar, HHeader, HKanbanCard, HFormCard, HProfileCard), dark mode usage, responsive breakpoints
  - [x] 33.2 Create `docs/MIGRATION_GUIDE.md`: step-by-step guide untuk migrate existing pages ke Horizon UI, before/after examples, common patterns, troubleshooting
