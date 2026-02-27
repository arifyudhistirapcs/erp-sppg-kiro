<template>
  <div class="kds-cooking-view">
    <div class="kds-header">
      <div class="header-content">
        <div class="header-left">
          <h2 class="header-title">Dapur - Memasak</h2>
          <p class="header-subtitle">Tampilan menu masakan</p>
        </div>
        <div class="header-right">
          <a-space :size="12">
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
          </a-space>
        </div>
      </div>
    </div>

    <div class="content-wrapper">
      <a-alert
        v-if="error"
        type="error"
        :message="error"
        closable
        show-icon
        @close="error = null"
        style="margin-bottom: 16px"
      >
        <template #action>
          <a-button size="small" type="primary" @click="retryLoad">
            Coba Lagi
          </a-button>
        </template>
      </a-alert>

      <a-spin :spinning="loading" tip="Memuat data...">
        <a-empty v-if="!loading && recipes.length === 0" :description="emptyMessage" />
        
        <a-row :gutter="[16, 16]" v-else>
          <a-col
            v-for="recipe in recipes"
            :key="recipe.recipe_id"
            :xs="24"
            :sm="24"
            :md="12"
            :lg="8"
            :xl="8"
          >
            <a-card
              :class="['recipe-card', `status-${recipe.status}`]"
              :title="recipe.name"
            >
              <template #extra>
                <a-tag :color="getStatusColor(recipe.status)">
                  {{ getStatusText(recipe.status) }}
                </a-tag>
              </template>

              <!-- Photo -->
              <div v-if="recipe.photo_url" class="recipe-photo">
                <img :src="recipe.photo_url" :alt="recipe.name" />
              </div>

              <div class="recipe-info">
                <a-descriptions :column="1" size="small" bordered>
                  <a-descriptions-item label="Jumlah Porsi">
                    <div class="portions-summary">
                      <div class="portions-total">
                        <strong>{{ recipe.portions_required }} porsi</strong>
                      </div>
                      <div v-if="getTotalSmallPortions(recipe.school_allocations) > 0 || getTotalLargePortions(recipe.school_allocations) > 0" class="portions-breakdown">
                        <span v-if="getTotalSmallPortions(recipe.school_allocations) > 0" class="portion-badge portion-badge-small">
                          K: {{ getTotalSmallPortions(recipe.school_allocations) }}
                        </span>
                        <span v-if="getTotalLargePortions(recipe.school_allocations) > 0" class="portion-badge portion-badge-large">
                          B: {{ getTotalLargePortions(recipe.school_allocations) }}
                        </span>
                      </div>
                    </div>
                  </a-descriptions-item>
                  <a-descriptions-item label="Waktu Mulai" v-if="recipe.start_time">
                    {{ formatTime(recipe.start_time) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="Waktu Selesai" v-if="recipe.end_time">
                    {{ formatTime(recipe.end_time) }}
                  </a-descriptions-item>
                  <a-descriptions-item label="Durasi Memasak" v-if="recipe.duration_minutes">
                    <a-tag color="green">{{ recipe.duration_minutes }} menit</a-tag>
                  </a-descriptions-item>
                </a-descriptions>

                <a-divider>Bahan-Bahan</a-divider>
                <a-list
                  size="small"
                  :data-source="recipe.items"
                  :split="false"
                >
                  <template #renderItem="{ item }">
                    <a-list-item>
                      <a-list-item-meta>
                        <template #title>
                          <div style="display: flex; justify-content: space-between; align-items: center;">
                            <span>{{ item.name }}</span>
                            <span style="color: #1890ff; font-weight: 600;">{{ item.quantity }} {{ item.unit }}</span>
                          </div>
                        </template>
                        <template #description v-if="item.raw_materials && item.raw_materials.length > 0">
                          <a-collapse :bordered="false" size="small" style="background: transparent; margin-top: 8px;">
                            <a-collapse-panel key="1" header="Bahan Baku">
                              <div class="raw-materials-list">
                                <div v-for="(raw, idx) in item.raw_materials" :key="idx" class="raw-material-item">
                                  <span class="raw-material-name">{{ raw.name }}</span>
                                  <span class="raw-material-quantity">{{ raw.quantity.toFixed(2) }} {{ raw.unit }}</span>
                                </div>
                              </div>
                            </a-collapse-panel>
                          </a-collapse>
                        </template>
                      </a-list-item-meta>
                    </a-list-item>
                  </template>
                </a-list>

                <a-divider>Instruksi Memasak</a-divider>
                <div class="instructions">
                  {{ recipe.instructions }}
                </div>

                <a-collapse v-if="recipe.school_allocations && recipe.school_allocations.length > 0" :bordered="false" style="margin-top: 16px; background: transparent;">
                  <a-collapse-panel key="1" header="Alokasi Sekolah">
                    <a-list
                      size="small"
                      :data-source="recipe.school_allocations"
                      :split="false"
                    >
                      <template #renderItem="{ item }">
                        <a-list-item>
                          <a-list-item-meta>
                            <template #title>
                              <div class="school-allocation-title">
                                <span class="school-name-text">{{ formatSchoolAllocation(item) }}</span>
                                <a-tag :color="getSchoolCategoryColor(item.school_category)" size="small">
                                  {{ item.school_category }}
                                </a-tag>
                              </div>
                            </template>
                            <template #description>
                              <div class="portion-breakdown">
                                <div v-if="item.portion_size_type === 'mixed'">
                                  <div v-if="item.portions_small > 0" class="portion-item portion-small">
                                    <a-badge :count="item.portions_small" :number-style="{ backgroundColor: '#faad14', fontWeight: 'bold' }">
                                      <a-tag color="orange" class="portion-tag portion-tag-small">
                                        <template #icon>
                                          <span class="portion-icon portion-icon-small">S</span>
                                        </template>
                                        Kecil (Kelas 1-3)
                                      </a-tag>
                                    </a-badge>
                                  </div>
                                  <div v-if="item.portions_large > 0" class="portion-item portion-large">
                                    <a-badge :count="item.portions_large" :number-style="{ backgroundColor: '#1890ff', fontWeight: 'bold' }">
                                      <a-tag color="blue" class="portion-tag portion-tag-large">
                                        <template #icon>
                                          <span class="portion-icon portion-icon-large">L</span>
                                        </template>
                                        Besar (Kelas 4-6)
                                      </a-tag>
                                    </a-badge>
                                  </div>
                                </div>
                                <div v-else>
                                  <div class="portion-item portion-large">
                                    <a-badge :count="item.portions_large" :number-style="{ backgroundColor: '#1890ff', fontWeight: 'bold' }">
                                      <a-tag color="blue" class="portion-tag portion-tag-large">
                                        <template #icon>
                                          <span class="portion-icon portion-icon-large">L</span>
                                        </template>
                                        Besar
                                      </a-tag>
                                    </a-badge>
                                  </div>
                                </div>
                                <div class="portion-total">
                                  <strong>Total: {{ item.total_portions }} porsi</strong>
                                </div>
                              </div>
                            </template>
                          </a-list-item-meta>
                        </a-list-item>
                      </template>
                    </a-list>
                  </a-collapse-panel>
                </a-collapse>
              </div>

              <template #actions>
                <a-button
                  v-if="recipe.status === 'pending'"
                  type="primary"
                  block
                  @click="startCooking(recipe)"
                  :loading="updatingRecipeId === recipe.recipe_id"
                >
                  <template #icon><play-circle-outlined /></template>
                  Mulai Masak
                </a-button>
                <a-button
                  v-else-if="recipe.status === 'cooking'"
                  type="primary"
                  block
                  @click="finishCooking(recipe)"
                  :loading="updatingRecipeId === recipe.recipe_id"
                  style="background-color: #52c41a; border-color: #52c41a"
                >
                  <template #icon><check-circle-outlined /></template>
                  Selesai
                </a-button>
                <a-button
                  v-else
                  type="default"
                  block
                  disabled
                >
                  <template #icon><check-outlined /></template>
                  Sudah Selesai
                </a-button>
              </template>
            </a-card>
          </a-col>
        </a-row>
      </a-spin>
    </div>
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
  CheckOutlined
} from '@ant-design/icons-vue'
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

// Compute empty message based on selected date
const emptyMessage = computed(() => {
  const today = new Date()
  const isToday = selectedDate.value.toDateString() === today.toDateString()
  return isToday ? 'Tidak ada menu untuk hari ini' : 'Tidak ada menu untuk tanggal ini'
})

// Get status color
const getStatusColor = (status) => {
  const colors = {
    pending: 'default',
    cooking: 'processing',
    ready: 'success'
  }
  return colors[status] || 'default'
}

// Get status text in Indonesian
const getStatusText = (status) => {
  const texts = {
    pending: 'Belum Dimulai',
    cooking: 'Sedang Dimasak',
    ready: 'Selesai'
  }
  return texts[status] || status
}

// Get school category color
const getSchoolCategoryColor = (category) => {
  const colors = {
    SD: 'blue',
    SMP: 'green',
    SMA: 'purple'
  }
  return colors[category] || 'default'
}

// Format timestamp to readable time
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Format school allocation display
const formatSchoolAllocation = (item) => {
  const parts = [item.school_name + ':']
  
  if (item.portion_size_type === 'mixed') {
    // SD schools with both portion sizes
    const portionParts = []
    if (item.portions_small > 0) {
      portionParts.push(`Kecil (${item.portions_small})`)
    }
    if (item.portions_large > 0) {
      portionParts.push(`Besar (${item.portions_large})`)
    }
    return parts[0] + ' ' + portionParts.join(', ')
  } else {
    // SMP/SMA schools with only large portions
    return `${parts[0]} Besar (${item.portions_large})`
  }
}

// Calculate total small portions across all schools
const getTotalSmallPortions = (allocations) => {
  if (!allocations || allocations.length === 0) return 0
  return allocations.reduce((total, alloc) => {
    return total + (alloc.portions_small || 0)
  }, 0)
}

// Calculate total large portions across all schools
const getTotalLargePortions = (allocations) => {
  if (!allocations || allocations.length === 0) return 0
  return allocations.reduce((total, alloc) => {
    return total + (alloc.portions_large || 0)
  }, 0)
}

// Load data from API
const loadData = async () => {
  loading.value = true
  error.value = null
  console.log('[KDS Cooking] Loading data for date:', selectedDate.value)
  try {
    const response = await getCookingToday(selectedDate.value)
    console.log('[KDS Cooking] API Response:', response)
    if (response.success) {
      recipes.value = response.data || []
      console.log('[KDS Cooking] Loaded recipes:', recipes.value.length)
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

// Retry loading data
const retryLoad = () => {
  loadData()
}

// Refresh data
const refreshData = () => {
  loadData()
}

// Start cooking a recipe
const startCooking = async (recipe) => {
  updatingRecipeId.value = recipe.recipe_id
  try {
    const response = await updateCookingStatus(recipe.recipe_id, 'cooking')
    if (response.success) {
      message.success('Status berhasil diperbarui: Mulai Masak')
      // Reload data from API to get updated status
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

// Finish cooking a recipe
const finishCooking = async (recipe) => {
  updatingRecipeId.value = recipe.recipe_id
  try {
    const response = await updateCookingStatus(recipe.recipe_id, 'ready')
    if (response.success) {
      message.success('Status berhasil diperbarui: Selesai')
      // Reload data from API to get updated status
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

// Setup Firebase real-time listener
const setupFirebaseListener = () => {
  // Clean up existing listener first
  cleanupFirebaseListener()
  
  const dateStr = selectedDate.value.toISOString().split('T')[0]
  const cookingRef = dbRef(database, `/kds/cooking/${dateStr}`)
  
  console.log('[KDS Cooking] Setting up Firebase listener for path:', `/kds/cooking/${dateStr}`)
  
  firebaseListener = onValue(
    cookingRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      
      console.log('[KDS Cooking] Firebase data received:', data)
      
      if (data) {
        // Update recipes with Firebase data
        const firebaseRecipes = Object.values(data)
        
        console.log('[KDS Cooking] Firebase recipes:', firebaseRecipes)
        
        // Merge with existing recipes to preserve ingredients and instructions
        recipes.value = recipes.value.map(recipe => {
          const firebaseRecipe = firebaseRecipes.find(fr => fr.recipe_id === recipe.recipe_id)
          if (firebaseRecipe) {
            console.log('[KDS Cooking] Updating recipe', recipe.recipe_id, 'with status:', firebaseRecipe.status)
            return {
              ...recipe,
              status: firebaseRecipe.status,
              start_time: firebaseRecipe.start_time,
              // Update school allocations with portion size data if present
              school_allocations: firebaseRecipe.school_allocations || recipe.school_allocations
            }
          }
          return recipe
        })
        
        console.log('[KDS Cooking] Updated recipes:', recipes.value)
      }
    },
    (error) => {
      console.error('Firebase listener error:', error)
      isConnected.value = false
    }
  )
}

// Cleanup Firebase listener
const cleanupFirebaseListener = () => {
  if (firebaseListener) {
    const dateStr = selectedDate.value.toISOString().split('T')[0]
    const cookingRef = dbRef(database, `/kds/cooking/${dateStr}`)
    off(cookingRef)
    firebaseListener = null
  }
}

// Handle date change from date picker
const handleDateChange = (date) => {
  selectedDate.value = date
  loadData()
  setupFirebaseListener()
}

// Watch for date changes
watch(selectedDate, () => {
  // This ensures Firebase listener is updated if date changes from other sources
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
.kds-cooking-view {
  background-color: #f0f2f5;
  min-height: 100vh;
}

.kds-header {
  background: white;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  max-width: 1600px;
  margin: 0 auto;
}

.header-left {
  flex: 1;
}

.header-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #262626;
  line-height: 1.4;
}

.header-subtitle {
  margin: 4px 0 0 0;
  font-size: 14px;
  color: #8c8c8c;
}

.header-right {
  display: flex;
  align-items: center;
}

.connection-tag {
  font-size: 13px;
  padding: 4px 12px;
  border-radius: 4px;
}

.content-wrapper {
  max-width: 1600px;
  margin: 24px auto;
  padding: 0 24px;
}

.recipe-card {
  height: 100%;
  transition: all 0.3s ease;
}

:deep(.ant-card-head) {
  padding: 16px 24px;
}

:deep(.ant-card-head-title) {
  white-space: normal;
  word-wrap: break-word;
  line-height: 1.5;
  padding-right: 8px;
}

:deep(.ant-card-body) {
  padding-top: 16px;
  padding-bottom: 16px;
}

.recipe-card.status-pending {
  border-left: 4px solid #d9d9d9;
}

.recipe-card.status-cooking {
  border-left: 4px solid #1890ff;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.2);
}

.recipe-card.status-ready {
  border-left: 4px solid #52c41a;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.2);
}

.recipe-info {
  margin-top: 16px;
}

.instructions {
  white-space: pre-wrap;
  line-height: 1.6;
  color: rgba(0, 0, 0, 0.65);
  padding: 12px;
  background-color: #fafafa;
  border-radius: 4px;
  max-height: 200px;
  overflow-y: auto;
}

:deep(.ant-list-item) {
  padding: 8px 0;
}

:deep(.ant-list-item-meta-title) {
  margin-bottom: 2px;
  font-weight: 500;
}

:deep(.ant-list-item-meta-description) {
  color: #1890ff;
  font-weight: 600;
}

.portions-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.portions-total {
  font-size: 14px;
}

.portions-breakdown {
  display: flex;
  gap: 8px;
  align-items: center;
}

.portion-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.portion-badge-small {
  background-color: #fff7e6;
  color: #fa8c16;
  border: 1px solid #ffd591;
}

.portion-badge-large {
  background-color: #e6f7ff;
  color: #1890ff;
  border: 1px solid #91d5ff;
}

:deep(.ant-collapse) {
  background: transparent;
  border: none;
}

:deep(.ant-collapse-item) {
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  margin-bottom: 0;
}

:deep(.ant-collapse-header) {
  padding: 12px 16px;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
}

:deep(.ant-collapse-content) {
  border-top: 1px solid #f0f0f0;
}

:deep(.ant-collapse-content-box) {
  padding: 12px 16px;
}

.raw-materials-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.raw-material-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #fafafa;
  border-radius: 4px;
  border: 1px solid #f0f0f0;
}

.raw-material-name {
  font-size: 13px;
  color: rgba(0, 0, 0, 0.65);
}

.raw-material-quantity {
  font-size: 13px;
  color: #52c41a;
  font-weight: 600;
}

.school-allocation-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.school-name-text {
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
}

.portion-breakdown {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.portion-item {
  display: flex;
  align-items: center;
  padding: 6px 8px;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.portion-item.portion-small {
  background-color: #fff7e6;
  border: 1px solid #ffd591;
}

.portion-item.portion-large {
  background-color: #e6f7ff;
  border: 1px solid #91d5ff;
}

.portion-tag {
  font-size: 13px;
  font-weight: 500;
  padding: 4px 12px;
  border-radius: 4px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border-width: 2px;
}

.portion-tag-small {
  border-color: #fa8c16;
  box-shadow: 0 2px 4px rgba(250, 140, 22, 0.2);
}

.portion-tag-large {
  border-color: #1890ff;
  box-shadow: 0 2px 4px rgba(24, 144, 255, 0.2);
}

.portion-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background-color: rgba(255, 255, 255, 0.4);
  border-radius: 50%;
  font-weight: 700;
  font-size: 12px;
  border: 2px solid rgba(255, 255, 255, 0.6);
}

.portion-icon-small {
  background-color: rgba(255, 255, 255, 0.5);
  border-color: #fa8c16;
}

.portion-icon-large {
  background-color: rgba(255, 255, 255, 0.5);
  border-color: #1890ff;
}

:deep(.ant-badge) {
  margin-right: 8px;
}

:deep(.ant-badge-count) {
  font-weight: 600;
  font-size: 14px;
  min-width: 28px;
  height: 22px;
  line-height: 22px;
  padding: 0 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.15);
}

.portion-label {
  color: rgba(0, 0, 0, 0.65);
  font-size: 13px;
}

.portion-value {
  color: #1890ff;
  font-weight: 600;
  font-size: 13px;
}

.portion-total {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  color: rgba(0, 0, 0, 0.85);
  font-size: 14px;
}

.no-allocations {
  padding: 12px;
  text-align: center;
  color: rgba(0, 0, 0, 0.45);
  font-style: italic;
}

.recipe-photo {
  width: 100%;
  margin-bottom: 16px;
  border-radius: 4px;
  overflow: hidden;
}

.recipe-photo img {
  width: 100%;
  height: 200px;
  object-fit: cover;
}
</style>
