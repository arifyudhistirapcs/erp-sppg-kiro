# UI Redesign - Horizon UI Template

## 📋 Overview

Merombak UI/UX aplikasi ERP SPPG mengikuti design template Horizon UI dengan tetap mempertahankan:
- ✅ Semua business logic yang sudah ada
- ✅ Semua fitur dan functionality existing
- ✅ User experience flow yang sudah familiar

## 🎯 Goals

1. **Visual Modernization**: Adopsi design language Horizon UI yang modern dan clean
2. **Consistency**: Unified design system across all pages
3. **User Experience**: Improve usability dengan better spacing, typography, dan visual hierarchy
4. **Dark Mode**: Implement theme toggle (light/dark)
5. **Performance**: Maintain atau improve performance saat ini

## 🎨 Design Specifications (Based on Screenshots)

### Color Palette

**Primary Colors** (Based on provided palette):
```css
/* Primary Purple - Main brand color */
--primary: #5A4372;
--primary-hover: #4a3562;
--primary-active: #3a2752;
--primary-light: #6a5382;

/* Accent Purple - For highlights */
--accent-purple: #3D2B53;
--accent-purple-light: #4d3b63;
```

**Neutral Colors**:
```css
/* Light Mode */
--bg-primary: #F8FDEA;        /* Cream/Off-white background */
--bg-secondary: #FFFFFF;       /* Pure white */
--bg-card: #FFFFFF;            /* Card background */
--bg-light: #ACA9B0;           /* Light gray */

--text-primary: #322837;       /* Dark purple-gray for text */
--text-secondary: #74788C;     /* Medium gray for secondary text */
--text-light: #ACA9B0;         /* Light gray for hints */

--border-color: #E9EDF7;       /* Light border */

/* Dark Mode */
--bg-primary-dark: #322837;    /* Dark purple-gray */
--bg-secondary-dark: #3D2B53;  /* Dark purple */
--bg-card-dark: #3D2B53;       /* Card background dark */

--text-primary-dark: #F8FDEA;  /* Cream text on dark */
--text-secondary-dark: #ACA9B0;/* Light gray text */
--border-color-dark: #5A4372;  /* Purple border */
```

**Accent Colors**:
```css
--success: #05CD99;            /* Keep green for success */
--warning: #FFB547;            /* Keep orange for warning */
--error: #EE5D50;              /* Keep red for error */
--info: #5A4372;               /* Use primary purple for info */
```

**Color Reference** (from image):
```css
--cream: #F8FDEA;              /* Top - Background */
--light-gray: #ACA9B0;         /* Second - Light elements */
--medium-gray: #74788C;        /* Third - Medium elements */
--dark-gray: #322837;          /* Fourth - Dark text/bg */
--purple: #5A4372;             /* Fifth - Primary purple */
--dark-purple: #3D2B53;        /* Sixth - Accent purple */
```

### Typography

**Font Family**:
```css
--font-primary: 'DM Sans', -apple-system, BlinkMacSystemFont, sans-serif;
/* Fallback: Keep Montserrat if DM Sans not available */
```

**Font Sizes**:
```css
--text-xs: 12px;
--text-sm: 14px;
--text-base: 16px;
--text-lg: 18px;
--text-xl: 20px;
--text-2xl: 24px;
--text-3xl: 30px;
--text-4xl: 36px;
```

**Font Weights**:
```css
--font-normal: 400;
--font-medium: 500;
--font-semibold: 600;
--font-bold: 700;
```

### Spacing System

```css
--spacing-1: 4px;
--spacing-2: 8px;
--spacing-3: 12px;
--spacing-4: 16px;
--spacing-5: 20px;
--spacing-6: 24px;
--spacing-8: 32px;
--spacing-10: 40px;
--spacing-12: 48px;
```

### Border Radius

```css
--radius-sm: 8px;
--radius-md: 12px;
--radius-lg: 16px;
--radius-xl: 20px;
--radius-full: 9999px;
```

### Shadows

```css
--shadow-sm: 0px 2px 4px rgba(0, 0, 0, 0.05);
--shadow-md: 0px 4px 6px rgba(0, 0, 0, 0.07);
--shadow-lg: 0px 10px 15px rgba(0, 0, 0, 0.1);
--shadow-xl: 0px 20px 25px rgba(0, 0, 0, 0.15);
--shadow-card: 0px 18px 40px rgba(112, 144, 176, 0.12);
```

## 📐 Layout Structure

### Sidebar (Based on Screenshot 1)

**Specifications**:
- Width: 280px (expanded), 80px (collapsed)
- Background: White (light mode), #111C44 (dark mode)
- Logo area: 64px height
- Menu items: 44px height each
- Icon size: 20px
- Font size: 14px (medium weight)
- Active state: Purple/Red background with white text
- Hover state: Light background (#F4F7FE)

**Menu Structure**:
```
- Dashboard (home icon)
- NFT Marketplace (cart icon)
- Tables (table icon)
- Kanban (kanban icon)
- Profile (user icon)
- Sign In (lock icon)
```

**ERP SPPG Adaptation**:
```
- Dashboard
- Monitoring Aktivitas
- Display (KDS)
  - Dapur
  - Pengemasan
  - Kebersihan
- Menu & Komponen
  - Perencanaan Menu
  - Manajemen Menu
  - Manajemen Komponen
- Supply Chain
  - Supplier
  - Purchase Order
  - Penerimaan Barang
  - Manajemen Bahan Baku
- Logistik
  - Data Sekolah
  - Tugas Pengiriman & Pengambilan
- SDM
  - Data Karyawan
  - Laporan Absensi
  - Konfigurasi Wi-Fi
  - Absensi
- Keuangan
  - Aset Dapur
  - Arus Kas
  - Laporan Keuangan
- Sistem
  - Audit Trail
  - Konfigurasi
```

### Top Header (Based on Screenshot 1)

**Specifications**:
- Height: 72px
- Background: White (light mode), #111C44 (dark mode)
- Padding: 0 24px
- Elements (left to right):
  1. Breadcrumb (Pages / Dashboard)
  2. Page Title (bold, 24px)
  3. Search bar (center-right, 320px width)
  4. Notification icon with badge
  5. Dark mode toggle
  6. Info icon
  7. User avatar with dropdown

**Search Bar**:
- Width: 320px
- Height: 40px
- Border radius: 12px
- Background: #F4F7FE (light), #1B254B (dark)
- Icon: Search icon (left)
- Placeholder: "Search"

### Content Area

**Specifications**:
- Background: #F4F7FE (light mode), #0B1437 (dark mode)
- Padding: 24px
- Max width: 100%
- Gap between cards: 20px

## 🎴 Component Specifications

### 1. Stat Card (Mini Statistics)

**Based on Screenshot 1 - Top Row Cards**:

**Dimensions**:
- Height: 100px
- Border radius: 16px
- Background: White
- Shadow: 0px 18px 40px rgba(112, 144, 176, 0.12)
- Padding: 20px

**Layout**:
```
[Icon]  Label
        Value
        Change indicator
```

**Elements**:
- Icon: 56x56px, gradient background (purple gradient), rounded 12px
- Label: 14px, color #74788C, font-weight 500
- Value: 24px, color #322837, font-weight 700
- Change: 12px, green/red with arrow icon

**Example**:
```
[📊] Earnings
     $350.4
     +23% since last month
```

### 2. Chart Card (Weekly Revenue)

**Based on Screenshot 1 - Center Card**:

**Dimensions**:
- Height: 320px
- Border radius: 16px
- Background: White
- Shadow: card shadow
- Padding: 24px

**Header**:
- Title: 18px, bold
- Subtitle: 14px, light
- Action button: "This month" dropdown

**Chart Area**:
- ECharts integration
- Smooth line charts
- Gradient fills
- Grid lines: subtle
- Colors: Purple/Blue (adapt to Red/Orange)

### 3. Data Table (Check Table, Complex Table)

**Based on Screenshot 1 & 3**:

**Specifications**:
- Border radius: 16px
- Background: White
- Shadow: card shadow
- Padding: 24px

**Header**:
- Title: 18px, bold, color #322837
- Actions: Three dots menu (right)

**Table**:
- Row height: 56px
- Header background: Transparent
- Header text: 12px, uppercase, color #74788C, font-weight 700
- Cell text: 14px, color #322837
- Border: None (use spacing)
- Hover: Background #F8FDEA

**Columns**:
- Checkbox: 24px
- Name: Left aligned
- Progress: Progress bar (colored)
- Quantity: Right aligned
- Date: Right aligned

**Status Badges**:
- Approved: Green dot + text
- Disable: Red dot + text
- Error: Yellow dot + text
- Border radius: 8px
- Padding: 4px 12px

### 4. Kanban Card

**Based on Screenshot 4**:

**Dimensions**:
- Width: 100% (column width)
- Min height: 120px
- Border radius: 16px
- Background: White
- Shadow: 0px 4px 6px rgba(0, 0, 0, 0.07)
- Padding: 20px

**Elements**:
- Title: 16px, bold, color #322837
- Description: 14px, color #74788C, line-height 1.6
- Image: Full width, border-radius 12px (if present)
- Avatars: 32px, overlapping (-8px margin)
- Status badge: Right aligned, colored

**Status Colors**:
- Backlog: Gray
- In Progress: Orange
- Done: Green
- Urgent: Red

### 5. Profile Card

**Based on Screenshot 5**:

**Banner**:
- Height: 200px
- Gradient background: Purple/Blue (adapt to Red)
- Border radius: 16px 16px 0 0

**Avatar**:
- Size: 80px
- Border: 4px white
- Position: Centered, overlapping banner

**Stats Row**:
- 3 columns
- Value: 24px, bold
- Label: 14px, light

**Content Cards**:
- Background: White
- Border radius: 16px
- Shadow: card shadow
- Padding: 24px

### 6. Login Page

**Based on Screenshot 6**:

**Layout**: Split screen (50/50)

**Left Side**:
- Background: White
- Max width: 480px
- Centered content
- Padding: 48px

**Elements**:
- Back link: Top left
- Title: 36px, bold, "Sign In"
- Subtitle: 14px, light
- Google button: Full width, white, border, icon + text
- Divider: "or"
- Input fields: 48px height, border-radius 12px
- Checkbox: "Keep me logged in"
- Forgot password: Right aligned, purple link
- Submit button: Full width, purple, 48px height
- Register link: Bottom, centered

**Right Side**:
- Gradient background: Purple gradient (#5A4372 to #3D2B53)
- Logo: Centered, large
- Brand name: "ERP SPPG"
- CTA button: Outlined, white text
- Footer links: Bottom

## 📱 Responsive Design Requirements

### Breakpoints

```css
--breakpoint-xs: 375px;   /* Small Mobile (iPhone SE) */
--breakpoint-sm: 640px;   /* Mobile */
--breakpoint-md: 768px;   /* Tablet */
--breakpoint-lg: 1024px;  /* Desktop */
--breakpoint-xl: 1280px;  /* Large Desktop */
--breakpoint-2xl: 1536px; /* Extra Large */
```

### Mobile-First Behavior

**Mobile (< 768px)**:
- Sidebar: Hidden by default, slide-in drawer dari kiri
- Hamburger menu: Top left corner (44x44px touch target)
- Header: Simplified, stack elements vertically if needed
- Search: Collapsible atau full-width modal
- Stats cards: Stack vertically (1 column)
- Charts: Full width, reduced height (240px)
- Tables: Horizontal scroll atau card view
- Forms: Full width inputs, larger touch targets (min 44px)
- Buttons: Full width on mobile
- Bottom navigation: Optional untuk quick access

**Tablet (768px - 1024px)**:
- Sidebar: Collapsed by default (icon only)
- Stats cards: 2 columns
- Charts: 2 columns
- Tables: Full table view dengan horizontal scroll
- Forms: 2 column layout where appropriate

**Desktop (> 1024px)**:
- Sidebar: Expanded by default
- Stats cards: 4 columns
- Charts: Flexible grid
- Tables: Full table view
- Forms: Multi-column layouts

### Mobile-Specific Features

**Touch Optimization**:
- Minimum touch target: 44x44px (Apple HIG standard)
- Spacing between interactive elements: min 8px
- Swipe gestures: Drawer open/close, card actions
- Pull-to-refresh: On list pages
- Long-press: Context menus

**Mobile Navigation**:
- Bottom tab bar (optional): Quick access to main sections
- Floating action button (FAB): Primary actions
- Back button: Browser back atau custom back
- Breadcrumb: Hidden on mobile, show page title only

**Mobile Components**:
- Drawer menu: Slide-in sidebar
- Action sheets: Bottom sheet untuk actions
- Modal: Full screen on mobile
- Toast notifications: Bottom positioned
- Loading states: Skeleton screens

**Performance**:
- Lazy load images
- Virtual scrolling untuk long lists
- Debounced search
- Optimized bundle size
- Service worker caching

### Responsive Component Specifications

#### Mobile Sidebar (Drawer)
```
Width: 280px
Position: Fixed, left: -280px (hidden)
Transition: transform 0.3s ease
Overlay: rgba(0,0,0,0.5)
Z-index: 1000
```

#### Mobile Header
```
Height: 56px (reduced from 72px)
Elements:
- Hamburger (left): 44x44px
- Page title: 16px, truncate if long
- Actions (right): Max 2-3 icons
```

#### Mobile Stats Cards
```
Width: 100%
Height: 100px
Margin: 12px 0
Stack vertically
```

#### Mobile Tables
```
Option 1: Horizontal scroll
- Min width: 100%
- Scroll indicator
- Sticky first column (optional)

Option 2: Card view
- Each row = card
- Stack fields vertically
- Actions at bottom
```

#### Mobile Forms
```
Input height: 48px (larger for touch)
Label: Above input (not floating)
Spacing: 16px between fields
Buttons: Full width, 48px height
Sticky footer: For form actions
```

### Mobile Testing Requirements

**Devices to Test**:
- iPhone SE (375x667) - Smallest modern iPhone
- iPhone 12/13/14 (390x844) - Standard iPhone
- iPhone 14 Pro Max (430x932) - Large iPhone
- Samsung Galaxy S21 (360x800) - Standard Android
- iPad Mini (768x1024) - Small tablet
- iPad Pro (1024x1366) - Large tablet

**Orientations**:
- Portrait (primary)
- Landscape (secondary)

**Browsers**:
- Safari iOS (primary)
- Chrome Android (primary)
- Chrome iOS
- Firefox Android

**Test Scenarios**:
- [ ] Navigation works smoothly
- [ ] All buttons are tappable (44x44px min)
- [ ] Forms are easy to fill
- [ ] Tables are readable
- [ ] Charts are interactive
- [ ] No horizontal scroll (except tables)
- [ ] Text is readable (min 14px)
- [ ] Images load properly
- [ ] Performance is good (< 3s load)
- [ ] Offline mode works (PWA)

## 🎭 Animations & Transitions

```css
--transition-fast: 150ms ease-in-out;
--transition-base: 200ms ease-in-out;
--transition-slow: 300ms ease-in-out;
```

**Hover Effects**:
- Cards: Lift (translateY(-4px)) + shadow increase
- Buttons: Background color change + scale(1.02)
- Menu items: Background color change
- Links: Color change

**Page Transitions**:
- Fade in: 200ms
- Slide in: 300ms (from right)

## 🌓 Dark Mode

**Toggle Button**:
- Position: Top right header
- Icon: Sun (light mode), Moon (dark mode)
- Size: 40px
- Background: #F4F7FE (light), #1B254B (dark)
- Border radius: 12px

**Color Switching**:
- Use CSS variables
- Smooth transition: 200ms
- Persist preference: localStorage

## 📊 Chart Specifications

**ECharts Theme** (Custom Purple Palette):
```javascript
{
  color: ['#5A4372', '#3D2B53', '#74788C', '#ACA9B0'],
  backgroundColor: 'transparent',
  textStyle: {
    fontFamily: 'DM Sans, sans-serif',
    fontSize: 14,
    color: '#322837'
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  // ... more theme config
}
```

**Chart Types Needed**:
1. Line Chart (Weekly Revenue)
2. Bar Chart (Daily Traffic)
3. Pie Chart (Your Pie Chart)
4. Area Chart (Total Spent)

## 🔧 Technical Requirements

### Dependencies to Add

```json
{
  "dependencies": {
    "@vueuse/core": "^10.7.0",  // For dark mode, etc
    "chart.js": "^4.4.0",        // Alternative to ECharts (lighter)
    "vue-chartjs": "^5.3.0"      // Vue wrapper for Chart.js
  }
}
```

### File Structure

```
web/src/
├── components/
│   ├── horizon/              # New Horizon UI components
│   │   ├── HStatCard.vue
│   │   ├── HChartCard.vue
│   │   ├── HDataTable.vue
│   │   ├── HKanbanCard.vue
│   │   ├── HProfileCard.vue
│   │   └── index.js
│   ├── layout/
│   │   ├── HSidebar.vue
│   │   ├── HHeader.vue
│   │   ├── HBreadcrumb.vue
│   │   └── ThemeToggle.vue
│   └── ... (existing components)
├── layouts/
│   └── HorizonLayout.vue     # New layout
├── styles/
│   ├── horizon-theme.css     # New theme file
│   ├── horizon-dark.css      # Dark mode styles
│   └── theme.css             # Keep for backward compatibility
└── composables/
    ├── useDarkMode.js
    └── useHorizonTheme.js
```

## 🎯 Success Criteria

1. ✅ Visual design matches Horizon UI screenshots
2. ✅ All existing features work without regression
3. ✅ Dark mode fully functional
4. ✅ **Mobile responsive - comfortable on phones (375px+)**
5. ✅ **Touch-friendly - min 44x44px touch targets**
6. ✅ **Tablet optimized - works well on iPad**
7. ✅ Performance: No degradation (Lighthouse score > 90)
8. ✅ Accessibility: WCAG 2.1 AA compliant
9. ✅ Browser support: Chrome, Firefox, Safari, Edge (latest 2 versions)
10. ✅ **Mobile browsers: Safari iOS, Chrome Android**

## 📝 Notes

- Prioritize dashboard pages first (high visibility)
- Maintain existing routing structure
- Keep all API integrations unchanged
- Gradual rollout: Feature flag for new UI
- User feedback loop after each phase

## 🚫 Out of Scope

- Backend changes
- API modifications
- Business logic changes
- New features (unless UI-related)
- Database schema changes

---

**Status**: Requirements Defined
**Next Step**: Design Phase
**Owner**: Frontend Team
**Timeline**: 10 weeks (as per analysis document)
