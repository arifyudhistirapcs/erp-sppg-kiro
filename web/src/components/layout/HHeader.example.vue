<template>
  <div class="example-container">
    <h2>HHeader Component Examples</h2>
    
    <!-- Example 1: Basic Usage -->
    <section class="example-section">
      <h3>1. Basic Usage</h3>
      <HHeader
        page-title="Dashboard"
        :breadcrumb="{ parent: 'Pages', current: 'Dashboard' }"
        :notification-count="5"
        @toggle-drawer="handleDrawerToggle"
        @search="handleSearch"
        @notification-click="handleNotificationClick"
      />
    </section>

    <!-- Example 2: Different Page -->
    <section class="example-section">
      <h3>2. Different Page Title</h3>
      <HHeader
        page-title="Menu Planning"
        :breadcrumb="{ parent: 'Menu & Komponen', current: 'Perencanaan Menu' }"
        :notification-count="12"
        @toggle-drawer="handleDrawerToggle"
        @search="handleSearch"
        @notification-click="handleNotificationClick"
      />
    </section>

    <!-- Example 3: No Notifications -->
    <section class="example-section">
      <h3>3. No Notifications</h3>
      <HHeader
        page-title="Settings"
        :breadcrumb="{ parent: 'Sistem', current: 'Konfigurasi' }"
        :notification-count="0"
        @toggle-drawer="handleDrawerToggle"
        @search="handleSearch"
        @notification-click="handleNotificationClick"
      />
    </section>

    <!-- Usage Instructions -->
    <section class="usage-section">
      <h3>Usage Instructions</h3>
      <pre><code>&lt;template&gt;
  &lt;HHeader
    page-title="Dashboard"
    :breadcrumb="{ parent: 'Pages', current: 'Dashboard' }"
    :notification-count="5"
    @toggle-drawer="handleDrawerToggle"
    @search="handleSearch"
    @notification-click="handleNotificationClick"
  /&gt;
&lt;/template&gt;

&lt;script setup&gt;
import HHeader from '@/components/layout/HHeader.vue'

const handleDrawerToggle = () => {
  console.log('Toggle mobile drawer')
}

const handleSearch = (query) => {
  console.log('Search query:', query)
}

const handleNotificationClick = () => {
  console.log('Notification clicked')
}
&lt;/script&gt;</code></pre>
    </section>

    <!-- Props Documentation -->
    <section class="props-section">
      <h3>Props</h3>
      <table>
        <thead>
          <tr>
            <th>Prop</th>
            <th>Type</th>
            <th>Default</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>pageTitle</code></td>
            <td>String</td>
            <td>'Dashboard'</td>
            <td>Page title to display (24px bold)</td>
          </tr>
          <tr>
            <td><code>breadcrumb</code></td>
            <td>Object</td>
            <td>{ parent: 'Pages', current: 'Dashboard' }</td>
            <td>Breadcrumb navigation (hidden on mobile)</td>
          </tr>
          <tr>
            <td><code>notificationCount</code></td>
            <td>Number</td>
            <td>0</td>
            <td>Number of unread notifications for badge</td>
          </tr>
        </tbody>
      </table>
    </section>

    <!-- Events Documentation -->
    <section class="events-section">
      <h3>Events</h3>
      <table>
        <thead>
          <tr>
            <th>Event</th>
            <th>Payload</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td><code>toggle-drawer</code></td>
            <td>-</td>
            <td>Emitted when hamburger menu is clicked (mobile only)</td>
          </tr>
          <tr>
            <td><code>search</code></td>
            <td>String (query)</td>
            <td>Emitted when user types in search bar</td>
          </tr>
          <tr>
            <td><code>notification-click</code></td>
            <td>-</td>
            <td>Emitted when notification icon is clicked</td>
          </tr>
        </tbody>
      </table>
    </section>

    <!-- Features -->
    <section class="features-section">
      <h3>Features</h3>
      <ul>
        <li>✅ Height: 72px (desktop), 56px (mobile)</li>
        <li>✅ Background: white (light mode), dark (dark mode)</li>
        <li>✅ Breadcrumb navigation (desktop only)</li>
        <li>✅ Page title: 24px bold (18px on mobile)</li>
        <li>✅ Search bar: 320px width, 40px height, 12px radius, #F4F7FE bg</li>
        <li>✅ Notification icon with badge</li>
        <li>✅ ThemeToggle integration</li>
        <li>✅ User avatar dropdown (profile, settings, logout)</li>
        <li>✅ Mobile: hamburger menu button 44x44px</li>
        <li>✅ Mobile: collapsible search (icon → full width)</li>
        <li>✅ Dark mode support</li>
        <li>✅ Responsive design</li>
      </ul>
    </section>
  </div>
</template>

<script setup>
import HHeader from './HHeader.vue'

const handleDrawerToggle = () => {
  console.log('Toggle mobile drawer')
}

const handleSearch = (query) => {
  console.log('Search query:', query)
}

const handleNotificationClick = () => {
  console.log('Notification clicked')
}
</script>

<style scoped>
.example-container {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.example-section {
  margin-bottom: 48px;
}

.example-section h3 {
  margin-bottom: 16px;
  color: var(--h-text-primary, #322837);
}

.usage-section,
.props-section,
.events-section,
.features-section {
  margin-top: 48px;
  padding: 24px;
  background: var(--h-bg-card, #FFFFFF);
  border-radius: 16px;
  box-shadow: var(--h-shadow-card, 0px 18px 40px rgba(112, 144, 176, 0.12));
}

.usage-section h3,
.props-section h3,
.events-section h3,
.features-section h3 {
  margin-bottom: 16px;
  color: var(--h-text-primary, #322837);
}

pre {
  background: #F4F7FE;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}

.dark pre {
  background: #1B254B;
}

code {
  font-family: 'Courier New', monospace;
  font-size: 14px;
  color: var(--h-text-primary, #322837);
}

.dark code {
  color: var(--h-text-primary-dark, #F8FDEA);
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 16px;
}

th, td {
  text-align: left;
  padding: 12px;
  border-bottom: 1px solid var(--h-border-color, #E9EDF7);
}

th {
  font-weight: 600;
  color: var(--h-text-primary, #322837);
  background: #F4F7FE;
}

.dark th {
  background: #1B254B;
  color: var(--h-text-primary-dark, #F8FDEA);
}

td {
  color: var(--h-text-secondary, #74788C);
}

.dark td {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

td code {
  background: #F4F7FE;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.dark td code {
  background: #1B254B;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  padding: 8px 0;
  color: var(--h-text-secondary, #74788C);
}

.dark li {
  color: var(--h-text-secondary-dark, #ACA9B0);
}
</style>
