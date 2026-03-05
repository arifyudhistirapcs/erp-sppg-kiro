<template>
  <van-pull-refresh v-model="refreshing" @refresh="onRefresh">
    <div class="menu-planning-page">
      <!-- NavBar -->
      <van-nav-bar title="Perencanaan Menu" />

      <!-- Week Filter -->
      <div class="menu-planning-week-filter">
        <van-field
          v-model="selectedWeekFormatted"
          placeholder="Pilih minggu"
          readonly
          right-icon="calendar-o"
          @click="showWeekPicker = true"
        />
      </div>

      <!-- Week Picker Popup -->
      <van-popup v-model:show="showWeekPicker" position="bottom" :style="{ zIndex: 9999 }">
        <van-picker
          :columns="weekColumns"
          @confirm="onWeekConfirm"
          @cancel="showWeekPicker = false"
          title="Pilih Minggu"
        >
          <template #confirm>
            <span style="color: #5A4372;">Konfirmasi</span>
          </template>
        </van-picker>
      </van-popup>

      <!-- Loading State -->
      <template v-if="menuStore.loading">
        <div class="plans-list">
          <SkeletonCard v-for="n in 3" :key="n" :rows="3" />
        </div>
      </template>

      <!-- Error State -->
      <div v-else-if="menuStore.error && !menuStore.weeklyPlans.length" class="error-state">
        <van-icon name="warning-o" size="48" color="var(--h-error)" />
        <p class="error-state__message">{{ menuStore.error }}</p>
        <van-button type="primary" size="normal" @click="menuStore.retry()">
          Coba Lagi
        </van-button>
      </div>

      <!-- Weekly Plans List -->
      <template v-else>
        <div v-if="menuStore.weeklyPlans.length" class="plans-list">
          <div v-for="plan in menuStore.weeklyPlans" :key="plan.id">
            <MenuWeekCard
              :weekPeriod="plan.weekPeriod"
              :approvalStatus="plan.approvalStatus"
              :menuCount="plan.menuCount"
              :canApprove="isKepalaSPPG && plan.approvalStatus === 'pending'"
              @approve="handleApprove(plan.id)"
              @reject="showRejectDialog(plan.id)"
              @click="toggleDetail(plan.id)"
            />

            <!-- Expanded Detail: Daily Menus -->
            <transition name="expand">
              <div v-if="expandedPlanId === plan.id && plan.menus && plan.menus.length" class="daily-menus">
                <div v-for="menu in plan.menus" :key="menu.day" class="menu-day-card h-card">
                  <div class="menu-day-card__header-section">
                    <h4 class="menu-day-card__day">{{ menu.day }}</h4>
                    <p class="menu-day-card__name">{{ menu.menuName }}</p>
                    <p class="menu-day-card__total-portions">
                      <van-icon name="friends-o" size="14" color="var(--h-text-secondary)" />
                      {{ menu.portions }} porsi
                    </p>
                  </div>
                  
                  <!-- School Allocations List -->
                  <div v-if="menu.schools && menu.schools.length" class="menu-day-card__schools">
                    <div v-for="school in menu.schools" :key="school.schoolId" class="school-allocation">
                      <span class="school-allocation__name">{{ school.schoolName }}</span>
                      <span class="school-allocation__portions">
                        <span v-if="school.portionsSmall > 0" class="portion-badge portion-badge--small">
                          K: {{ school.portionsSmall }}
                        </span>
                        <span v-if="school.portionsLarge > 0" class="portion-badge portion-badge--large">
                          B: {{ school.portionsLarge }}
                        </span>
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </transition>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="!menuStore.weeklyPlans.length" class="empty-state">
          <van-icon name="notes-o" size="48" color="var(--h-text-light)" />
          <p class="empty-state__text">Tidak ada rencana menu ditemukan</p>
        </div>
      </template>

      <!-- Approval Loading Overlay -->
      <div v-if="menuStore.approvalLoading" class="approval-overlay">
        <van-loading type="spinner" color="var(--h-primary)" />
      </div>
    </div>
  </van-pull-refresh>

  <!-- Reject Dialog -->
  <van-dialog
    v-model:show="showReject"
    title="Tolak Menu"
    show-cancel-button
    confirm-button-text="Tolak"
    cancel-button-text="Batal"
    @confirm="handleReject"
    @cancel="closeRejectDialog"
  >
    <div class="reject-dialog__content">
      <van-field
        v-model="rejectReason"
        placeholder="Alasan penolakan"
        type="textarea"
        rows="3"
        autosize
      />
    </div>
  </van-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useMenuPlanningStore } from '@/stores/menuPlanning'
import { useAuthStore } from '@/stores/auth'
import { showToast } from 'vant'
import MenuWeekCard from '@/components/mobile/MenuWeekCard.vue'
import SkeletonCard from '@/components/mobile/SkeletonCard.vue'

const menuStore = useMenuPlanningStore()
const authStore = useAuthStore()

const refreshing = ref(false)
const showWeekPicker = ref(false)
const expandedPlanId = ref(null)
const showReject = ref(false)
const rejectReason = ref('')
const rejectTargetId = ref(null)

// Generate week options (current week and next 8 weeks)
const weekColumns = ref([])

function generateWeekColumns() {
  const weeks = []
  const today = new Date()
  
  // Generate for current week and next 8 weeks
  for (let i = 0; i < 9; i++) {
    const date = new Date(today)
    date.setDate(date.getDate() + (i * 7))
    
    const weekStart = getWeekStart(date)
    const weekEnd = new Date(weekStart)
    weekEnd.setDate(weekEnd.getDate() + 6)
    
    const weekNum = getWeekNumber(date)
    const year = date.getFullYear()
    
    weeks.push({
      text: `Minggu ${weekNum}, ${year} (${formatDate(weekStart)} - ${formatDate(weekEnd)})`,
      value: `${year}-W${String(weekNum).padStart(2, '0')}`
    })
  }
  
  weekColumns.value = weeks
}

function getWeekStart(date) {
  const d = new Date(date)
  const day = d.getDay()
  const diff = d.getDate() - day + (day === 0 ? -6 : 1) // Adjust when day is Sunday
  return new Date(d.setDate(diff))
}

function getWeekNumber(date) {
  const d = new Date(date)
  d.setHours(0, 0, 0, 0)
  d.setDate(d.getDate() + 4 - (d.getDay() || 7))
  const yearStart = new Date(d.getFullYear(), 0, 1)
  return Math.ceil((((d - yearStart) / 86400000) + 1) / 7)
}

function formatDate(date) {
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
}

// Role check — only Kepala_SPPG can approve/reject
const isKepalaSPPG = computed(() => authStore.user?.role?.toLowerCase() === 'kepala_sppg')

// Format selected week for display
const selectedWeekFormatted = computed(() => {
  if (!menuStore.selectedWeek) {
    const today = new Date()
    const weekNum = getWeekNumber(today)
    const year = today.getFullYear()
    const weekStart = getWeekStart(today)
    const weekEnd = new Date(weekStart)
    weekEnd.setDate(weekEnd.getDate() + 6)
    return `Minggu ${weekNum}, ${year} (${formatDate(weekStart)} - ${formatDate(weekEnd)})`
  }
  
  const selected = weekColumns.value.find(w => w.value === menuStore.selectedWeek)
  return selected ? selected.text : menuStore.selectedWeek
})

function onWeekConfirm(value) {
  showWeekPicker.value = false
  menuStore.setWeek(value.value)
  showToast(`Menampilkan menu ${value.text}`)
}

function toggleDetail(planId) {
  expandedPlanId.value = expandedPlanId.value === planId ? null : planId
}

async function handleApprove(planId) {
  const result = await menuStore.approveMenu(planId)
  if (result.success) {
    showToast({
      message: 'Menu berhasil disetujui',
      type: 'success'
    })
  } else {
    showToast({
      message: result.error || 'Gagal menyetujui menu',
      type: 'fail'
    })
  }
}

function showRejectDialog(planId) {
  rejectTargetId.value = planId
  rejectReason.value = ''
  showReject.value = true
}

function closeRejectDialog() {
  showReject.value = false
  rejectReason.value = ''
  rejectTargetId.value = null
}

async function handleReject() {
  if (rejectTargetId.value) {
    const result = await menuStore.rejectMenu(rejectTargetId.value, rejectReason.value)
    if (result.success) {
      showToast({
        message: 'Menu berhasil ditolak',
        type: 'success'
      })
    } else {
      showToast({
        message: result.error || 'Gagal menolak menu',
        type: 'fail'
      })
    }
  }
  closeRejectDialog()
}

async function onRefresh() {
  await menuStore.fetchMenuPlans()
  refreshing.value = false
}

onMounted(() => {
  generateWeekColumns()
  menuStore.fetchMenuPlans()
})
</script>

<style scoped>
.menu-planning-page {
  padding: 0;
  padding-bottom: 80px;
  min-height: 100vh;
  position: relative;
}

.menu-planning-page > :not(.van-nav-bar) {
  padding-left: var(--h-spacing-lg);
  padding-right: var(--h-spacing-lg);
}

/* Week Filter */
.menu-planning-week-filter {
  margin-top: var(--h-spacing-lg);
  margin-bottom: var(--h-spacing-lg);
  background: #FFFFFF;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0px 2px 8px rgba(0, 0, 0, 0.08);
}

/* Plans List */
.plans-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-md);
}

/* Daily Menus (expanded detail) */
.daily-menus {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-sm);
  margin-top: var(--h-spacing-sm);
  padding-left: var(--h-spacing-md);
}

.menu-day-card {
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-card);
  padding: var(--h-spacing-md);
  border-left: 3px solid var(--h-primary);
}

.menu-day-card__header-section {
  margin-bottom: var(--h-spacing-md);
  padding-bottom: var(--h-spacing-sm);
  border-bottom: 1px solid var(--h-border-light);
}

.menu-day-card__day {
  font-size: 14px;
  font-weight: 600;
  color: var(--h-primary);
  margin: 0 0 4px 0;
}

.menu-day-card__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--h-text-primary);
  margin: 0 0 8px 0;
}

.menu-day-card__total-portions {
  font-size: 13px;
  color: var(--h-text-secondary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

/* School Allocations List */
.menu-day-card__schools {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-sm);
}

.school-allocation {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--h-spacing-sm);
  background: rgba(90, 67, 114, 0.05);
  border-radius: var(--h-radius-md);
}

.school-allocation__name {
  font-size: 13px;
  color: var(--h-text-primary);
  font-weight: 500;
  flex: 1;
}

.school-allocation__portions {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-xs);
}

.portion-badge {
  font-size: 12px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--h-radius-sm);
  white-space: nowrap;
}

.portion-badge--small {
  color: #1989FA;
  background: rgba(25, 137, 250, 0.1);
}

.portion-badge--large {
  color: #07C160;
  background: rgba(7, 193, 96, 0.1);
}

/* Expand Transition */
.expand-enter-active,
.expand-leave-active {
  transition: all var(--h-transition-base);
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
}

.expand-enter-to,
.expand-leave-from {
  opacity: 1;
  max-height: 1000px;
}

/* Approval Loading Overlay */
.approval-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.6);
  z-index: 10;
  border-radius: var(--h-radius-lg);
}

/* Error State */
.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px var(--h-spacing-xl);
  text-align: center;
}

.error-state__message {
  font-size: 14px;
  color: var(--h-text-secondary);
  margin: var(--h-spacing-lg) 0;
  line-height: 1.5;
}

/* Empty State */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px var(--h-spacing-xl);
  text-align: center;
}

.empty-state__text {
  font-size: 14px;
  color: var(--h-text-light);
  margin: var(--h-spacing-md) 0 0 0;
}

/* Reject Dialog */
.reject-dialog__content {
  padding: var(--h-spacing-md);
}

/* Date Picker Popup Fix */
:deep(.van-popup) {
  z-index: 9999 !important;
  background: rgba(0, 0, 0, 0.7) !important;
}

:deep(.van-popup--bottom) {
  background: #FFFFFF !important;
}

:deep(.van-date-picker) {
  background: #FFFFFF !important;
}

:deep(.van-picker__toolbar) {
  background: #FFFFFF !important;
  border-bottom: 1px solid #ebedf0;
}

:deep(.van-picker-column) {
  background: #FFFFFF !important;
}

:deep(.van-picker-column__item) {
  color: #323233 !important;
}

:deep(.van-overlay) {
  z-index: 9998 !important;
}
</style>
