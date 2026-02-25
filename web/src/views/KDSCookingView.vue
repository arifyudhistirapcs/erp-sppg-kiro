<template>
  <div class="kds-cooking-view">
    <a-page-header
      title="Dapur - Memasak"
      sub-title="Tampilan menu masakan"
    >
      <template #extra>
        <a-space>
          <KDSDatePicker
            v-model="selectedDate"
            :loading="loading"
            @change="handleDateChange"
          />
          <a-tag :color="isConnected ? 'green' : 'red'">
            <template #icon>
              <wifi-outlined v-if="isConnected" />
              <disconnect-outlined v-else />
            </template>
            {{ isConnected ? 'Terhubung' : 'Terputus' }}
          </a-tag>
          <a-button @click="refreshData" :loading="loading">
            <template #icon><reload-outlined /></template>
            Refresh
          </a-button>
        </a-space>
      </template>
    </a-page-header>

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

              <div class="recipe-info">
                <a-descriptions :column="1" size="small" bordered>
                  <a-descriptions-item label="Jumlah Porsi">
                    <strong>{{ recipe.portions_required }} porsi</strong>
                  </a-descriptions-item>
                  <a-descriptions-item label="Waktu Mulai" v-if="recipe.start_time">
                    {{ formatTime(recipe.start_time) }}
                  </a-descriptions-item>
                </a-descriptions>

                <a-divider>Alokasi Sekolah</a-divider>
                <a-list
                  v-if="recipe.school_allocations && recipe.school_allocations.length > 0"
                  size="small"
                  :data-source="recipe.school_allocations"
                  :split="false"
                >
                  <template #renderItem="{ item }">
                    <a-list-item>
                      <a-list-item-meta>
                        <template #title>
                          {{ item.school_name }}
                        </template>
                        <template #description>
                          {{ item.portions }} porsi
                        </template>
                      </a-list-item-meta>
                    </a-list-item>
                  </template>
                </a-list>
                <div v-else class="no-allocations">
                  Tidak ada alokasi sekolah
                </div>

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
                          {{ item.name }}
                        </template>
                        <template #description>
                          {{ item.quantity }} {{ item.unit }}
                        </template>
                      </a-list-item-meta>
                    </a-list-item>
                  </template>
                </a-list>

                <a-divider>Instruksi Memasak</a-divider>
                <div class="instructions">
                  {{ recipe.instructions }}
                </div>
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

// Format timestamp to readable time
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit'
  })
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
      // Update will come from Firebase listener
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
      // Update will come from Firebase listener
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
  
  firebaseListener = onValue(
    cookingRef,
    (snapshot) => {
      isConnected.value = true
      const data = snapshot.val()
      
      if (data) {
        // Update recipes with Firebase data
        const firebaseRecipes = Object.values(data)
        
        // Merge with existing recipes to preserve ingredients and instructions
        recipes.value = recipes.value.map(recipe => {
          const firebaseRecipe = firebaseRecipes.find(fr => fr.recipe_id === recipe.recipe_id)
          if (firebaseRecipe) {
            return {
              ...recipe,
              status: firebaseRecipe.status,
              start_time: firebaseRecipe.start_time
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
  padding: 24px;
  background-color: #f0f2f5;
  min-height: 100vh;
}

.content-wrapper {
  margin-top: 16px;
}

.recipe-card {
  height: 100%;
  transition: all 0.3s ease;
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

.no-allocations {
  padding: 12px;
  text-align: center;
  color: rgba(0, 0, 0, 0.45);
  font-style: italic;
}
</style>
