<template>
  <div>
    <!-- Header Controls -->
    <div class="cooking-header">
      <div class="header-controls">
        <KDSDatePicker
          v-model="selectedDate"
          :loading="loading"
          @change="handleDateChange"
        />
        <a-tag :color="isConnected ? 'green' : 'red'" class="connection-tag">
          <template #icon>
            <wifi-outlined v-if="isConnected" />
            <disconnect-outlined v-else />
          </template>
          {{ isConnected ? 'Terhubung' : 'Terputus' }}
        </a-tag>
        <a-button @click="refreshData" :loading="loading" type="default">
          <template #icon><reload-outlined /></template>
          Refresh
        </a-button>
      </div>
    </div>

    <!-- Alerts -->
    <a-alert
      v-if="error"
      type="error"
      :message="error"
      closable
      show-icon
      @close="error = null"
      class="cooking-alert"
    >
      <template #action>
        <a-button size="small" type="primary" @click="retryLoad">
          Coba Lagi
        </a-button>
      </template>
    </a-alert>





    <!-- Kanban Board -->
    <a-spin :spinning="loading" tip="Memuat data...">
      <a-empty v-if="!loading && recipes.length === 0" :description="emptyMessage" />

      <div v-else class="kanban-board">
        <!-- Pending Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <ClockCircleOutlined class="column-icon" />
              Belum Dimulai
            </h3>
            <span class="kanban-column-count">{{ pendingRecipes.length }}</span>
          </div>
          <div class="kanban-column-content">
            <div v-for="recipe in pendingRecipes" :key="recipe.recipe_id" class="h-card recipe-card status-pending">
              <div class="recipe-card__header">
                <div class="recipe-card__name">{{ recipe.name }}</div>
                <div class="recipe-card__status-badge status-pending">
                  <span class="status-dot"></span>
                  Belum Dimulai
                </div>
              </div>

              <div v-if="recipe.photo_url" class="recipe-card__photo">
                <img :src="recipe.photo_url" :alt="recipe.name" />
              </div>

              <div class="recipe-card__portions">
                <span class="portions-label">Jumlah Porsi</span>
                <span class="portions-value">{{ recipe.portions_required }}</span>
              </div>

              <!-- Ingredients -->
              <div class="recipe-card__section">
                <div class="section-title">Bahan-Bahan</div>
                <div class="ingredients-list">
                  <div v-for="item in recipe.items" :key="item.name" class="ingredient-item">
                    <span class="ingredient-name">{{ item.name }}</span>
                    <div class="ingredient-qty">
                      <span class="qty-value">{{ item.quantity }} {{ item.unit }}</span>
                      <span v-if="item.current_stock !== undefined" class="stock-indicator" :class="item.current_stock >= item.quantity ? 'stock-ok' : 'stock-low'">
                        Stok: {{ item.current_stock }} {{ item.unit }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Instructions -->
              <div v-if="recipe.instructions" class="recipe-card__section">
                <div class="section-title">Instruksi</div>
                <div class="instructions-text">{{ recipe.instructions }}</div>
              </div>

              <!-- School Allocations -->
              <div v-if="recipe.school_allocations && recipe.school_allocations.length > 0" class="recipe-card__section">
                <div class="section-title">Alokasi Sekolah</div>
                <div class="school-list">
                  <div v-for="alloc in recipe.school_allocations" :key="alloc.school_id" class="school-item">
                    <div class="school-item__header">
                      <span class="school-name">{{ alloc.school_name }}</span>
                      <a-tag :color="getSchoolCategoryColor(alloc.school_category)" size="small">{{ alloc.school_category }}</a-tag>
                    </div>
                    <div class="school-item__portions">
                      <span v-if="alloc.portions_small > 0" class="portion-tag small">S: {{ alloc.portions_small }}</span>
                      <span v-if="alloc.portions_large > 0" class="portion-tag large">L: {{ alloc.portions_large }}</span>
                      <span class="portion-total">Total: {{ alloc.total_portions }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <a-button type="primary" block @click="startCooking(recipe)" :loading="updatingRecipeId === recipe.recipe_id" class="action-button">
                <template #icon><play-circle-outlined /></template>
                Mulai Masak
              </a-button>
            </div>
          </div>
        </div>

        <!-- Cooking Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <FireOutlined class="column-icon" />
              Sedang Dimasak
            </h3>
            <span class="kanban-column-count">{{ cookingRecipes.length }}</span>
          </div>
          <div class="kanban-column-content">
            <div v-for="recipe in cookingRecipes" :key="recipe.recipe_id" class="h-card recipe-card status-cooking">
              <div class="recipe-card__header">
                <div class="recipe-card__name">{{ recipe.name }}</div>
                <div class="recipe-card__status-badge status-cooking">
                  <span class="status-dot"></span>
                  Sedang Dimasak
                </div>
              </div>

              <div v-if="recipe.photo_url" class="recipe-card__photo">
                <img :src="recipe.photo_url" :alt="recipe.name" />
              </div>

              <div class="recipe-card__portions">
                <span class="portions-label">Jumlah Porsi</span>
                <span class="portions-value">{{ recipe.portions_required }}</span>
              </div>

              <div v-if="recipe.start_time" class="recipe-card__time">
                <span class="time-label">Mulai</span>
                <span class="time-value">{{ formatTime(recipe.start_time) }}</span>
              </div>

              <!-- Ingredients -->
              <div class="recipe-card__section">
                <div class="section-title">Bahan-Bahan</div>
                <div class="ingredients-list">
                  <div v-for="item in recipe.items" :key="item.name" class="ingredient-item">
                    <span class="ingredient-name">{{ item.name }}</span>
                    <div class="ingredient-qty">
                      <span class="qty-value">{{ item.quantity }} {{ item.unit }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Instructions -->
              <div v-if="recipe.instructions" class="recipe-card__section">
                <div class="section-title">Instruksi</div>
                <div class="instructions-text">{{ recipe.instructions }}</div>
              </div>

              <a-button type="primary" block @click="finishCooking(recipe)" :loading="updatingRecipeId === recipe.recipe_id" class="action-button action-button--success">
                <template #icon><check-circle-outlined /></template>
                Selesai Masak
              </a-button>
            </div>
          </div>
        </div>

        <!-- Ready Column -->
        <div class="kanban-column">
          <div class="kanban-column-header">
            <h3 class="kanban-column-title">
              <CheckCircleOutlined class="column-icon" />
              Selesai
            </h3>
            <span class="kanban-column-count">{{ readyRecipes.length }}</span>
          </div>
          <div class="kanban-column-content">
            <div v-for="recipe in readyRecipes" :key="recipe.recipe_id" class="h-card recipe-card status-ready">
              <div class="recipe-card__header">
                <div class="recipe-card__name">{{ recipe.name }}</div>
                <div class="recipe-card__status-badge status-ready">
                  <span class="status-dot"></span>
                  Selesai
                </div>
              </div>

              <div v-if="recipe.photo_url" class="recipe-card__photo">
                <img :src="recipe.photo_url" :alt="recipe.name" />
              </div>

              <div class="recipe-card__portions">
                <span class="portions-label">Jumlah Porsi</span>
                <span class="portions-value">{{ recipe.portions_required }}</span>
              </div>

              <div v-if="recipe.start_time || recipe.end_time" class="recipe-card__times">
                <div v-if="recipe.start_time" class="time-row">
                  <span class="time-label">Mulai</span>
                  <span class="time-value">{{ formatTime(recipe.start_time) }}</span>
                </div>
                <div v-if="recipe.end_time" class="time-row">
                  <span class="time-label">Selesai</span>
                  <span class="time-value">{{ formatTime(recipe.end_time) }}</span>
                </div>
                <div v-if="recipe.duration_minutes" class="time-row">
                  <span class="time-label">Durasi</span>
                  <span class="time-value duration">{{ recipe.duration_minutes }} menit</span>
                </div>
              </div>

              <div class="completed-badge">
                <check-outlined />
                Sudah Selesai
              </div>
            </div>
          </div>
        </div>
      </div>
    </a-spin>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { message } from 'ant-design-vue'
import {
  WifiOutlined,
  DisconnectOutlined,
  ReloadOutlined,
  PlayCircleOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  ClockCircleOutlined,
  FireOutlined
} from '@ant-design/icons-vue'
import HStatCard from '@/components/horizon/HStatCard.vue'
import KDSDatePicker from '@/components/KDSDatePicker.vue'
import { getCookingToday, updateCookingStatus } from '@/services/kdsService'
import { database } from '@/services/firebase'
import { ref as dbRef, onValue, off } from 'firebase/database'

const recipes = ref([])
const loading = ref(false)
const updatingRecipeId = ref(null)
const isConnected = ref(true)
const selectedDate = ref(new Date())
const error = ref(null)
let firebaseListener = null

// Filter recipes by status
const pendingRecipes = computed(() => recipes.value.filter(r => r.status === 'pending'))
const cookingRecipes = computed(() => recipes.value.filter(r => r.status === 'cooking'))
const readyRecipes = computed(() => recipes.value.filter(r => r.status === 'ready'))

// Counts
const pendingCount = computed(() => pendingRecipes.value.length)
const cookingCount = computed(() => cookingRecipes.value.length)
const readyCount = computed(() => readyRecipes.value.length)

const allRecipesCompleted = computed(() => {
  if (recipes.value.length === 0) return false
  return recipes.value.every(recipe => recipe.status === 'ready')
})

const emptyMessage = computed(() => {
  const today = new Date()
  const isToday = selectedDate.value.toDateString() === today.toDateString()
  return isToday ? 'Tidak ada menu untuk hari ini' : 'Tidak ada menu untuk tanggal ini'
})

const getSchoolCategoryColor = (category) => {
  const colors = { SD: 'blue', SMP: 'green', SMA: 'purple' }
  return colors[category] || 'default'
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' })
}

const getTotalSmallPortions = (allocations) => {
  if (!allocations || allocations.length === 0) return 0
  return allocations.reduce((total, alloc) => total + (alloc.portions_small || 0), 0)
}

const getTotalLargePortions = (allocations) => {
  if (!allocations || allocations.length === 0) return 0
  return allocations.reduce((total, alloc) => total + (alloc.portions_large || 0), 0)
}

const loadData = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await getCookingToday(selectedDate.value)
    if (response.success) {
      recipes.value = response.data || []
    } else {
      error.value = response.message || 'Gagal memuat data'
    }
  } catch (err) {
    console.error('Error loading cooking data:', err)
    error.value = err.response?.data?.message || 'Gagal memuat data menu. Silakan coba lagi.'
  } finally {
    loading.value = false
  }
}

const retryLoad = () => loadData()
const refreshData = () => loadData()

const startCooking = async (recipe) => {
  updatingRecipeId.value = recipe.recipe_id
  try {
    const response = await updateCookingStatus(recipe.recipe_id, 'cooking')
    if (response.success) {
      message.success('Status berhasil diperbarui: Mulai Masak')
      await loadData()
    } else {
      message.error(response.message || 'Gagal memperbarui status')
    }
  } catch (error) {
    console.error('Error updating status:', error)
    message.error(error.response?.data?.message || 'Gagal memperbarui status')
  } finally {
    updatingRecipeId.value = null
  }
}

const finishCooking = async (recipe) => {
  updatingRecipeId.value = recipe.recipe_id
  try {
    const response = await updateCookingStatus(recipe.recipe_id, 'ready')
    if (response.success) {
      message.success('Status berhasil diperbarui: Selesai')
      await loadData()
    } else {
      message.error(response.message || 'Gagal memperbarui status')
    }
  } catch (error) {
    console.error('Error updating status:', error)
    message.error(error.response?.data?.message || 'Gagal memperbarui status')
  } finally {
    updatingRecipeId.value = null
  }
}

const setupFirebaseListener = () => {
  cleanupFirebaseListener()
  const dateStr = selectedDate.value.toISOString().split('T')[0]
  const cookingRef = dbRef(database, `/kds/cooking/${dateStr}`)

  firebaseListener = onValue(
    cookingRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      if (data) {
        const firebaseRecipes = Object.values(data)
        recipes.value = recipes.value.map(recipe => {
          const fr = firebaseRecipes.find(f => f.recipe_id === recipe.recipe_id)
          if (fr) {
            return {
              ...recipe,
              status: fr.status,
              start_time: fr.start_time,
              school_allocations: fr.school_allocations || recipe.school_allocations
            }
          }
          return recipe
        })
      }
    },
    (error) => {
      console.error('Firebase listener error:', error)
      isConnected.value = false
    }
  )
}

const cleanupFirebaseListener = () => {
  if (firebaseListener) {
    const dateStr = selectedDate.value.toISOString().split('T')[0]
    const cookingRef = dbRef(database, `/kds/cooking/${dateStr}`)
    off(cookingRef)
    firebaseListener = null
  }
}

const handleDateChange = (date) => {
  selectedDate.value = date
  loadData()
  setupFirebaseListener()
}

watch(selectedDate, () => {
  setupFirebaseListener()
})

onMounted(() => {
  loadData()
  setupFirebaseListener()
})

onUnmounted(() => {
  cleanupFirebaseListener()
})
</script>

<style scoped>
/* Header Controls */
.cooking-header {
  display: flex;
  justify-content: flex-end;
  margin-bottom: var(--h-spacing-4);
}

.header-controls {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-3);
}

.connection-tag {
  font-size: var(--h-text-sm);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
}

.cooking-alert {
  margin-bottom: var(--h-spacing-5);
}

/* Stat Cards Row */
.stat-cards-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--h-spacing-5);
  margin-bottom: var(--h-spacing-6);
}

/* Kanban Board */
.kanban-board {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--h-spacing-5);
  align-items: start;
}

.kanban-column {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
  min-height: 400px;
}

.kanban-column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--h-spacing-4);
  background: var(--h-bg-card);
  border-radius: var(--h-radius-lg);
  box-shadow: var(--h-shadow-sm);
}

.kanban-column-title {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2);
  margin: 0;
  font-size: var(--h-text-lg);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
}

.column-icon {
  font-size: var(--h-text-xl);
  color: var(--h-primary);
}

.kanban-column-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 var(--h-spacing-2);
  background: var(--h-primary);
  color: white;
  border-radius: var(--h-radius-full);
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-bold);
}

.kanban-column-content {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
}

/* Recipe Card */
.recipe-card {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
  transition: all var(--h-transition-base);
  border-left: 4px solid var(--h-border-color);
}

.recipe-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.recipe-card.status-pending { border-left-color: #FFB547; }
.recipe-card.status-cooking { border-left-color: #5A4372; }
.recipe-card.status-ready { border-left-color: #05CD99; }

.recipe-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--h-spacing-3);
}

.recipe-card__name {
  flex: 1;
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  line-height: var(--h-leading-tight);
}

.recipe-card__status-badge {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  flex-shrink: 0;
}

.recipe-card__status-badge.status-pending { background: rgba(255, 181, 71, 0.1); color: #FFB547; }
.recipe-card__status-badge.status-cooking { background: rgba(90, 67, 114, 0.1); color: #5A4372; }
.recipe-card__status-badge.status-ready { background: rgba(5, 205, 153, 0.1); color: #05CD99; }

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

/* Photo */
.recipe-card__photo {
  width: 100%;
  border-radius: var(--h-radius-md);
  overflow: hidden;
}

.recipe-card__photo img {
  width: 100%;
  height: 160px;
  object-fit: cover;
}

/* Portions */
.recipe-card__portions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--h-spacing-4);
  background: linear-gradient(135deg, #5A4372 0%, #3D2B53 100%);
  border-radius: var(--h-radius-md);
  color: white;
}

.portions-label {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-medium);
}

.portions-value {
  font-size: var(--h-text-2xl);
  font-weight: var(--h-font-bold);
}

/* Portion Breakdown */
.portion-breakdown {
  display: flex;
  gap: var(--h-spacing-3);
}

.portion-size-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--h-spacing-2);
  padding: var(--h-spacing-3);
  border-radius: var(--h-radius-md);
  transition: all var(--h-transition-base);
}

.portion-size-item:hover { transform: translateY(-2px); }

.portion-size-item.portion-small {
  background: linear-gradient(135deg, #FFF4E6 0%, #FFE7BA 100%);
  border: 2px solid #FFB547;
}

.portion-size-item.portion-large {
  background: linear-gradient(135deg, #E6F7FF 0%, #BAE7FF 100%);
  border: 2px solid #1890FF;
}

.size-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: white;
  border-radius: var(--h-radius-full);
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-primary);
  box-shadow: var(--h-shadow-sm);
}

.size-label {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  color: var(--h-text-secondary);
}

.size-value {
  font-size: var(--h-text-xl);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
}

/* Time */
.recipe-card__time,
.recipe-card__times {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-2);
  padding: var(--h-spacing-3);
  background: var(--h-bg-light);
  border-radius: var(--h-radius-md);
}

.time-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.time-label {
  font-size: var(--h-text-sm);
  color: var(--h-text-secondary);
  font-weight: var(--h-font-medium);
}

.time-value {
  font-size: var(--h-text-sm);
  color: var(--h-text-primary);
  font-weight: var(--h-font-semibold);
}

.time-value.duration {
  color: #05CD99;
}

/* Sections */
.recipe-card__section {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-3);
}

.section-title {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Ingredients */
.ingredients-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-2);
}

.ingredient-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--h-spacing-2) var(--h-spacing-3);
  background: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
}

.ingredient-name {
  font-size: var(--h-text-sm);
  color: var(--h-text-primary);
  font-weight: var(--h-font-medium);
}

.ingredient-qty {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.qty-value {
  font-size: var(--h-text-sm);
  color: var(--h-primary);
  font-weight: var(--h-font-semibold);
}

.stock-indicator {
  font-size: 11px;
  font-weight: var(--h-font-medium);
}

.stock-ok { color: #05CD99; }
.stock-low { color: #EE5D50; }

/* Instructions */
.instructions-text {
  white-space: pre-wrap;
  line-height: 1.6;
  color: var(--h-text-secondary);
  padding: var(--h-spacing-3);
  background: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
  font-size: var(--h-text-sm);
  max-height: 160px;
  overflow-y: auto;
}

/* School Allocations */
.school-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-2);
}

.school-item {
  padding: var(--h-spacing-3);
  background: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
}

.school-item__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--h-spacing-2);
}

.school-name {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-medium);
  color: var(--h-text-primary);
}

.school-item__portions {
  display: flex;
  gap: var(--h-spacing-2);
  align-items: center;
  flex-wrap: wrap;
}

.portion-tag {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-medium);
  padding: 2px 8px;
  border-radius: var(--h-radius-sm);
}

.portion-tag.small {
  background: #FFF4E6;
  color: #FFB547;
  border: 1px solid #FFD591;
}

.portion-tag.large {
  background: #E6F7FF;
  color: #1890FF;
  border: 1px solid #91D5FF;
}

.portion-total {
  font-size: var(--h-text-xs);
  font-weight: var(--h-font-semibold);
  color: var(--h-text-primary);
  margin-left: auto;
}

/* Action Button */
.action-button {
  margin-top: var(--h-spacing-2);
  height: var(--h-touch-target-min);
  font-weight: var(--h-font-semibold);
  border-radius: var(--h-radius-md);
}

.action-button--success {
  background-color: #05CD99 !important;
  border-color: #05CD99 !important;
}

.action-button--success:hover {
  background-color: #04b888 !important;
  border-color: #04b888 !important;
}

/* Completed Badge */
.completed-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--h-spacing-2);
  padding: var(--h-spacing-3);
  background: rgba(5, 205, 153, 0.1);
  border-radius: var(--h-radius-md);
  color: #05CD99;
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-semibold);
}

/* Dark Mode */
.dark .kanban-column-header { background: var(--h-bg-card); }
.dark .kanban-column-title { color: var(--h-text-primary); }
.dark .column-icon { color: var(--h-primary-light); }
.dark .recipe-card__name { color: var(--h-text-primary); }
.dark .recipe-card__portions { background: linear-gradient(135deg, #6a5382 0%, #4d3b63 100%); }
.dark .portion-size-item.portion-small { background: rgba(255, 181, 71, 0.15); border-color: #FFB547; }
.dark .portion-size-item.portion-large { background: rgba(24, 144, 255, 0.15); border-color: #1890FF; }
.dark .size-badge { background: var(--h-bg-card); }
.dark .ingredient-item { background: rgba(255, 255, 255, 0.05); }
.dark .ingredient-name { color: var(--h-text-primary); }
.dark .instructions-text { background: rgba(255, 255, 255, 0.05); color: var(--h-text-secondary); }
.dark .school-item { background: rgba(255, 255, 255, 0.05); }
.dark .school-name { color: var(--h-text-primary); }
.dark .portion-tag.small { background: rgba(255, 181, 71, 0.2); color: #FFB547; border-color: rgba(255, 181, 71, 0.3); }
.dark .portion-tag.large { background: rgba(24, 144, 255, 0.2); color: #1890FF; border-color: rgba(24, 144, 255, 0.3); }
.dark .completed-badge { background: rgba(5, 205, 153, 0.2); }
.dark .recipe-card__time,
.dark .recipe-card__times { background: rgba(255, 255, 255, 0.05); }
.dark .time-value { color: var(--h-text-primary); }

/* Responsive */
@media (max-width: 1024px) {
  .stat-cards-row { grid-template-columns: repeat(3, 1fr); gap: var(--h-spacing-3); }
  .kanban-board { grid-template-columns: 1fr; gap: var(--h-spacing-6); }
  .kanban-column { min-height: auto; }
}

@media (max-width: 768px) {
  .header-controls { flex-wrap: wrap; gap: var(--h-spacing-2); }
  .stat-cards-row { grid-template-columns: 1fr; gap: var(--h-spacing-3); }
  .portion-breakdown { flex-direction: column; }
  .portion-size-item { flex-direction: row; justify-content: space-between; align-items: center; }
  .size-badge { width: 28px; height: 28px; font-size: var(--h-text-sm); }
}
</style>
