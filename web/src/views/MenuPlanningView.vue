<template>
  <div class="menu-planning">
    <a-page-header
      title="Perencanaan Menu Mingguan"
      sub-title="Susun menu mingguan dengan standar gizi terpenuhi"
    >
      <template #extra>
        <a-space>
          <a-button @click="duplicatePreviousWeek">
            <template #icon><CopyOutlined /></template>
            Duplikat Minggu Lalu
          </a-button>
          <a-button type="primary" @click="showCreateModal">
            <template #icon><PlusOutlined /></template>
            Buat Menu Baru
          </a-button>
        </a-space>
      </template>
    </a-page-header>

    <a-card>
      <!-- Week Selector -->
      <a-row :gutter="16" class="mb-4">
        <a-col :span="12">
          <a-space>
            <a-button @click="previousWeek">
              <template #icon><LeftOutlined /></template>
            </a-button>
            <a-date-picker
              v-model:value="selectedWeekStart"
              picker="week"
              format="[Minggu] w, YYYY"
              @change="onWeekChange"
              style="width: 250px"
            />
            <a-button @click="nextWeek">
              <template #icon><RightOutlined /></template>
            </a-button>
            <a-button type="link" @click="goToCurrentWeek">
              Minggu Ini
            </a-button>
          </a-space>
        </a-col>
        <a-col :span="12" style="text-align: right">
          <a-space v-if="currentMenuPlan">
            <a-tag :color="currentMenuPlan.status === 'approved' ? 'green' : 'orange'">
              {{ currentMenuPlan.status === 'approved' ? 'Disetujui' : 'Draft' }}
            </a-tag>
            <a-button
              v-if="currentMenuPlan.status === 'draft' && canApprove"
              type="primary"
              @click="approveMenu"
              :loading="approving"
            >
              <template #icon><CheckOutlined /></template>
              Setujui Menu
            </a-button>
          </a-space>
        </a-col>
      </a-row>

      <!-- Weekly Calendar -->
      <a-spin :spinning="loading">
        <div class="weekly-calendar">
          <a-row :gutter="16">
            <a-col
              v-for="day in weekDays"
              :key="day.date"
              :span="24 / 7"
            >
              <a-card
                size="small"
                :class="['day-card', { 'today': isToday(day.date) }]"
              >
                <template #title>
                  <div class="day-header">
                    <div>{{ day.dayName }}</div>
                    <div class="date-text">{{ formatDate(day.date) }}</div>
                  </div>
                </template>

                <!-- Menu Items for this day -->
                <div
                  class="menu-items-container"
                  @drop="onDrop($event, day.date)"
                  @dragover.prevent
                  @dragenter.prevent
                >
                  <div
                    v-for="item in getMenuItemsForDay(day.date)"
                    :key="item.id"
                    class="menu-item"
                    draggable="true"
                    @dragstart="onDragStart($event, item)"
                  >
                    <div class="menu-item-name">{{ item.recipe?.name }}</div>
                    <div class="menu-item-portions">{{ item.portions }} porsi</div>
                    <a-button
                      type="text"
                      size="small"
                      danger
                      @click="removeMenuItem(item)"
                    >
                      <template #icon><DeleteOutlined /></template>
                    </a-button>
                  </div>

                  <!-- Add Menu Button -->
                  <a-button
                    type="dashed"
                    size="small"
                    block
                    @click="showAddMenuModal(day.date)"
                    style="margin-top: 8px"
                  >
                    <template #icon><PlusOutlined /></template>
                    Tambah Menu
                  </a-button>
                </div>

                <!-- Daily Nutrition Summary -->
                <a-divider style="margin: 12px 0" />
                <div class="nutrition-summary">
                  <div class="nutrition-item">
                    <span class="label">Kalori:</span>
                    <span :class="['value', getDailyNutritionStatus(day.date, 'calories')]">
                      {{ getDailyNutrition(day.date, 'calories') }} kkal
                    </span>
                  </div>
                  <div class="nutrition-item">
                    <span class="label">Protein:</span>
                    <span :class="['value', getDailyNutritionStatus(day.date, 'protein')]">
                      {{ getDailyNutrition(day.date, 'protein') }} g
                    </span>
                  </div>
                  <div class="validation-status">
                    <a-tag
                      :color="isDailyNutritionValid(day.date) ? 'success' : 'warning'"
                      style="margin: 0"
                    >
                      {{ isDailyNutritionValid(day.date) ? '✓ Memenuhi Standar' : '⚠ Belum Memenuhi' }}
                    </a-tag>
                  </div>
                </div>
              </a-card>
            </a-col>
          </a-row>
        </div>
      </a-spin>
    </a-card>

    <!-- Add Menu Item Modal -->
    <a-modal
      v-model:visible="addMenuModalVisible"
      title="Tambah Menu"
      @ok="addMenuItem"
      ok-text="Tambah"
      cancel-text="Batal"
    >
      <a-form layout="vertical">
        <a-form-item label="Pilih Resep">
          <a-select
            v-model:value="selectedRecipeId"
            show-search
            placeholder="Cari dan pilih resep"
            :filter-option="filterRecipeOption"
            style="width: 100%"
          >
            <a-select-option
              v-for="recipe in availableRecipes"
              :key="recipe.id"
              :value="recipe.id"
            >
              {{ recipe.name }} ({{ recipe.category }})
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="Jumlah Porsi">
          <a-input-number
            v-model:value="selectedPortions"
            :min="1"
            style="width: 100%"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  PlusOutlined,
  CopyOutlined,
  LeftOutlined,
  RightOutlined,
  CheckOutlined,
  DeleteOutlined
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import weekOfYear from 'dayjs/plugin/weekOfYear'
import isoWeek from 'dayjs/plugin/isoWeek'
import menuPlanningService from '@/services/menuPlanningService'
import recipeService from '@/services/recipeService'
import { useAuthStore } from '@/stores/auth'

dayjs.extend(weekOfYear)
dayjs.extend(isoWeek)

const authStore = useAuthStore()
const canApprove = computed(() => {
  return authStore.user?.role === 'ahli_gizi' || authStore.user?.role === 'kepala_sppg'
})

const loading = ref(false)
const approving = ref(false)
const selectedWeekStart = ref(dayjs().startOf('isoWeek'))
const currentMenuPlan = ref(null)
const menuItems = ref([])
const availableRecipes = ref([])
const addMenuModalVisible = ref(false)
const selectedDate = ref(null)
const selectedRecipeId = ref(null)
const selectedPortions = ref(100)
const draggedItem = ref(null)

// Minimum nutrition standards per portion
const MIN_CALORIES_PER_PORTION = 600
const MIN_PROTEIN_PER_PORTION = 15

const weekDays = computed(() => {
  const days = []
  const dayNames = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu', 'Minggu']
  
  for (let i = 0; i < 7; i++) {
    const date = selectedWeekStart.value.add(i, 'day')
    days.push({
      date: date.format('YYYY-MM-DD'),
      dayName: dayNames[i],
      dayjs: date
    })
  }
  
  return days
})

const isToday = (date) => {
  return dayjs(date).isSame(dayjs(), 'day')
}

const formatDate = (date) => {
  return dayjs(date).format('DD/MM')
}

const getMenuItemsForDay = (date) => {
  return menuItems.value.filter(item => item.date === date)
}

const getDailyNutrition = (date, type) => {
  const items = getMenuItemsForDay(date)
  let total = 0
  
  items.forEach(item => {
    const recipe = item.recipe
    if (!recipe) return
    
    const portionFactor = item.portions / recipe.serving_size
    
    switch (type) {
      case 'calories':
        total += (recipe.total_calories || 0) * portionFactor
        break
      case 'protein':
        total += (recipe.total_protein || 0) * portionFactor
        break
      case 'carbs':
        total += (recipe.total_carbs || 0) * portionFactor
        break
      case 'fat':
        total += (recipe.total_fat || 0) * portionFactor
        break
    }
  })
  
  return total.toFixed(type === 'calories' ? 0 : 1)
}

const getDailyNutritionStatus = (date, type) => {
  const value = parseFloat(getDailyNutrition(date, type))
  
  if (type === 'calories') {
    return value >= MIN_CALORIES_PER_PORTION ? 'valid' : 'invalid'
  } else if (type === 'protein') {
    return value >= MIN_PROTEIN_PER_PORTION ? 'valid' : 'invalid'
  }
  
  return ''
}

const isDailyNutritionValid = (date) => {
  const calories = parseFloat(getDailyNutrition(date, 'calories'))
  const protein = parseFloat(getDailyNutrition(date, 'protein'))
  
  return calories >= MIN_CALORIES_PER_PORTION && protein >= MIN_PROTEIN_PER_PORTION
}

const loadMenuPlan = async () => {
  loading.value = true
  try {
    const weekStart = selectedWeekStart.value.format('YYYY-MM-DD')
    const weekEnd = selectedWeekStart.value.add(6, 'day').format('YYYY-MM-DD')
    
    const response = await menuPlanningService.getMenuPlans({
      week_start: weekStart,
      week_end: weekEnd
    })
    
    if (response.data.data && response.data.data.length > 0) {
      currentMenuPlan.value = response.data.data[0]
      menuItems.value = currentMenuPlan.value.menu_items || []
    } else {
      currentMenuPlan.value = null
      menuItems.value = []
    }
  } catch (error) {
    console.error('Error loading menu plan:', error)
    currentMenuPlan.value = null
    menuItems.value = []
  } finally {
    loading.value = false
  }
}

const loadRecipes = async () => {
  try {
    const response = await recipeService.getRecipes({ is_active: true })
    availableRecipes.value = response.data.data || []
  } catch (error) {
    message.error('Gagal memuat data resep')
    console.error('Error loading recipes:', error)
  }
}

const showCreateModal = async () => {
  if (currentMenuPlan.value) {
    message.warning('Menu untuk minggu ini sudah ada')
    return
  }
  
  try {
    const weekStart = selectedWeekStart.value.format('YYYY-MM-DD')
    const weekEnd = selectedWeekStart.value.add(6, 'day').format('YYYY-MM-DD')
    
    const response = await menuPlanningService.createMenuPlan({
      week_start: weekStart,
      week_end: weekEnd,
      status: 'draft'
    })
    
    currentMenuPlan.value = response.data.data
    menuItems.value = []
    message.success('Menu mingguan baru berhasil dibuat')
  } catch (error) {
    message.error('Gagal membuat menu mingguan')
    console.error('Error creating menu plan:', error)
  }
}

const showAddMenuModal = (date) => {
  if (!currentMenuPlan.value) {
    message.warning('Buat menu mingguan terlebih dahulu')
    return
  }
  
  if (currentMenuPlan.value.status === 'approved') {
    message.warning('Menu yang sudah disetujui tidak dapat diubah')
    return
  }
  
  selectedDate.value = date
  selectedRecipeId.value = null
  selectedPortions.value = 100
  addMenuModalVisible.value = true
}

const filterRecipeOption = (input, option) => {
  return option.children[0].children.toLowerCase().includes(input.toLowerCase())
}

const addMenuItem = async () => {
  if (!selectedRecipeId.value) {
    message.warning('Pilih resep terlebih dahulu')
    return
  }
  
  try {
    const recipe = availableRecipes.value.find(r => r.id === selectedRecipeId.value)
    
    const newItem = {
      menu_plan_id: currentMenuPlan.value.id,
      date: selectedDate.value,
      recipe_id: selectedRecipeId.value,
      portions: selectedPortions.value,
      recipe: recipe
    }
    
    // Update menu plan with new item
    const updatedItems = [...menuItems.value, newItem]
    await saveMenuPlan(updatedItems)
    
    addMenuModalVisible.value = false
    message.success('Menu berhasil ditambahkan')
  } catch (error) {
    message.error('Gagal menambahkan menu')
    console.error('Error adding menu item:', error)
  }
}

const removeMenuItem = async (item) => {
  if (currentMenuPlan.value.status === 'approved') {
    message.warning('Menu yang sudah disetujui tidak dapat diubah')
    return
  }
  
  try {
    const updatedItems = menuItems.value.filter(i => i !== item)
    await saveMenuPlan(updatedItems)
    message.success('Menu berhasil dihapus')
  } catch (error) {
    message.error('Gagal menghapus menu')
    console.error('Error removing menu item:', error)
  }
}

const saveMenuPlan = async (items) => {
  try {
    const payload = {
      week_start: currentMenuPlan.value.week_start,
      week_end: currentMenuPlan.value.week_end,
      status: currentMenuPlan.value.status,
      menu_items: items.map(item => ({
        date: item.date,
        recipe_id: item.recipe_id,
        portions: item.portions
      }))
    }
    
    const response = await menuPlanningService.updateMenuPlan(currentMenuPlan.value.id, payload)
    currentMenuPlan.value = response.data.data
    menuItems.value = response.data.data.menu_items || []
  } catch (error) {
    throw error
  }
}

const approveMenu = async () => {
  // Validate all days meet nutrition standards
  let allDaysValid = true
  weekDays.value.forEach(day => {
    if (!isDailyNutritionValid(day.date)) {
      allDaysValid = false
    }
  })
  
  if (!allDaysValid) {
    message.warning('Tidak semua hari memenuhi standar gizi minimum. Yakin ingin menyetujui?')
    // In production, you might want to show a confirmation modal here
  }
  
  approving.value = true
  try {
    await menuPlanningService.approveMenuPlan(currentMenuPlan.value.id)
    message.success('Menu berhasil disetujui')
    await loadMenuPlan()
  } catch (error) {
    message.error('Gagal menyetujui menu')
    console.error('Error approving menu:', error)
  } finally {
    approving.value = false
  }
}

const duplicatePreviousWeek = async () => {
  if (currentMenuPlan.value) {
    message.warning('Menu untuk minggu ini sudah ada')
    return
  }
  
  try {
    const previousWeekStart = selectedWeekStart.value.subtract(7, 'day').format('YYYY-MM-DD')
    const previousWeekEnd = selectedWeekStart.value.subtract(1, 'day').format('YYYY-MM-DD')
    
    const response = await menuPlanningService.getMenuPlans({
      week_start: previousWeekStart,
      week_end: previousWeekEnd
    })
    
    if (!response.data.data || response.data.data.length === 0) {
      message.warning('Tidak ada menu minggu lalu untuk diduplikat')
      return
    }
    
    const previousMenu = response.data.data[0]
    const weekStart = selectedWeekStart.value.format('YYYY-MM-DD')
    const weekEnd = selectedWeekStart.value.add(6, 'day').format('YYYY-MM-DD')
    
    // Create new menu with items from previous week
    const newMenuResponse = await menuPlanningService.createMenuPlan({
      week_start: weekStart,
      week_end: weekEnd,
      status: 'draft',
      menu_items: previousMenu.menu_items?.map(item => {
        // Calculate date offset
        const oldDate = dayjs(item.date)
        const newDate = oldDate.add(7, 'day').format('YYYY-MM-DD')
        
        return {
          date: newDate,
          recipe_id: item.recipe_id,
          portions: item.portions
        }
      }) || []
    })
    
    currentMenuPlan.value = newMenuResponse.data.data
    menuItems.value = newMenuResponse.data.data.menu_items || []
    message.success('Menu minggu lalu berhasil diduplikat')
  } catch (error) {
    message.error('Gagal menduplikat menu minggu lalu')
    console.error('Error duplicating previous week:', error)
  }
}

const onDragStart = (event, item) => {
  draggedItem.value = item
  event.dataTransfer.effectAllowed = 'move'
}

const onDrop = async (event, targetDate) => {
  if (!draggedItem.value) return
  
  if (currentMenuPlan.value.status === 'approved') {
    message.warning('Menu yang sudah disetujui tidak dapat diubah')
    return
  }
  
  try {
    // Update the date of the dragged item
    const updatedItems = menuItems.value.map(item => {
      if (item === draggedItem.value) {
        return { ...item, date: targetDate }
      }
      return item
    })
    
    await saveMenuPlan(updatedItems)
    message.success('Menu berhasil dipindahkan')
  } catch (error) {
    message.error('Gagal memindahkan menu')
    console.error('Error moving menu item:', error)
  } finally {
    draggedItem.value = null
  }
}

const onWeekChange = () => {
  loadMenuPlan()
}

const previousWeek = () => {
  selectedWeekStart.value = selectedWeekStart.value.subtract(7, 'day')
  loadMenuPlan()
}

const nextWeek = () => {
  selectedWeekStart.value = selectedWeekStart.value.add(7, 'day')
  loadMenuPlan()
}

const goToCurrentWeek = () => {
  selectedWeekStart.value = dayjs().startOf('isoWeek')
  loadMenuPlan()
}

onMounted(() => {
  loadMenuPlan()
  loadRecipes()
})
</script>

<style scoped>
.menu-planning {
  padding: 24px;
}

.mb-4 {
  margin-bottom: 16px;
}

.weekly-calendar {
  margin-top: 16px;
}

.day-card {
  height: 100%;
  min-height: 400px;
}

.day-card.today {
  border: 2px solid #1890ff;
}

.day-header {
  text-align: center;
}

.day-header .date-text {
  font-size: 12px;
  font-weight: normal;
  color: #8c8c8c;
}

.menu-items-container {
  min-height: 200px;
}

.menu-item {
  background: #f0f2f5;
  padding: 8px;
  margin-bottom: 8px;
  border-radius: 4px;
  cursor: move;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.menu-item:hover {
  background: #e6e9ed;
}

.menu-item-name {
  font-weight: 500;
  font-size: 13px;
  flex: 1;
}

.menu-item-portions {
  font-size: 11px;
  color: #8c8c8c;
  margin-right: 8px;
}

.nutrition-summary {
  font-size: 12px;
}

.nutrition-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.nutrition-item .label {
  color: #8c8c8c;
}

.nutrition-item .value {
  font-weight: 500;
}

.nutrition-item .value.valid {
  color: #52c41a;
}

.nutrition-item .value.invalid {
  color: #faad14;
}

.validation-status {
  margin-top: 8px;
  text-align: center;
}
</style>
