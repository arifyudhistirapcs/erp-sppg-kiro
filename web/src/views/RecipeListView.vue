<template>
  <div>
    <!-- Header Actions -->
    <div class="recipe-list-header">
      <a-button type="primary" @click="showCreateModal" class="add-button">
        <template #icon><PlusOutlined /></template>
        Tambah Menu Baru
      </a-button>
    </div>

    <!-- Search and Filter Bar -->
    <div class="h-card search-filter-bar">
      <a-row :gutter="[16, 16]">
        <a-col :xs="24" :sm="24" :md="12" :lg="14">
          <a-input
            v-model:value="searchText"
            placeholder="Cari menu..."
            @change="handleSearch"
            allow-clear
            size="large"
          >
            <template #prefix>
              <SearchOutlined />
            </template>
          </a-input>
        </a-col>
        <a-col :xs="24" :sm="24" :md="12" :lg="10">
          <a-select
            v-model:value="filterCategory"
            placeholder="Semua Kategori"
            style="width: 100%"
            size="large"
            allow-clear
            @change="handleSearch"
          >
            <a-select-option value="masakan_indonesia">Masakan Indonesia</a-select-option>
            <a-select-option value="masakan_china">Masakan China</a-select-option>
            <a-select-option value="masakan_western">Masakan Western</a-select-option>
            <a-select-option value="masakan_india">Masakan India</a-select-option>
            <a-select-option value="masakan_gabungan">Masakan Gabungan</a-select-option>
            <a-select-option value="lainnya">Lainnya</a-select-option>
          </a-select>
        </a-col>
      </a-row>
    </div>

    <!-- Recipe Grid -->
    <a-spin :spinning="loading">
      <a-empty v-if="!loading && recipes.length === 0" description="Tidak ada menu" />
      
      <div v-else class="recipe-grid">
        <div
          v-for="recipe in recipes"
          :key="recipe.id"
          class="h-card recipe-card h-card-hover"
        >
          <!-- Card Header with Category and Actions -->
          <div class="recipe-card__header">
            <a-tag :color="getCategoryColor(recipe.category)" class="category-tag">
              {{ getCategoryLabel(recipe.category) }}
            </a-tag>
            <a-dropdown :trigger="['click']">
              <a class="action-menu-trigger" @click.prevent>
                <MoreOutlined />
              </a>
              <template #overlay>
                <a-menu>
                  <a-menu-item @click="viewRecipe(recipe)">
                    <EyeOutlined /> Lihat Detail
                  </a-menu-item>
                  <a-menu-item @click="editRecipe(recipe)">
                    <EditOutlined /> Edit
                  </a-menu-item>
                  <a-menu-item @click="viewHistory(recipe)">
                    <HistoryOutlined /> Riwayat
                  </a-menu-item>
                  <a-menu-divider />
                  <a-menu-item danger @click="confirmDelete(recipe)">
                    <DeleteOutlined /> Hapus
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>

          <!-- Recipe Title -->
          <div class="recipe-card__title">
            {{ recipe.name }}
          </div>

          <!-- Recipe Photo -->
          <div class="recipe-card__photo">
            <img 
              v-if="recipe.photo_url" 
              :src="recipe.photo_url" 
              :alt="recipe.name"
            />
            <div v-else class="no-photo">
              <PictureOutlined />
            </div>
          </div>

          <!-- Components Section -->
          <div class="recipe-card__section">
            <div class="section-title">Komponen Menu</div>
            <div class="component-list">
              <div
                v-for="item in recipe.recipe_items"
                :key="item.id"
                class="component-item"
              >
                <span class="component-dot" :style="{ backgroundColor: getComponentColor(item.semi_finished_goods?.category) }"></span>
                <span class="component-name">{{ item.semi_finished_goods?.name }}</span>
                <a-tag :color="getComponentTagColor(item.semi_finished_goods?.category)" size="small">
                  {{ getCategoryLabel(item.semi_finished_goods?.category) }}
                </a-tag>
              </div>
            </div>
          </div>

          <!-- Ingredients Section -->
          <div class="recipe-card__section">
            <div class="section-title">Bahan Baku</div>
            <div class="ingredient-list">
              <template v-for="item in recipe.recipe_items" :key="'ing-' + item.id">
                <div
                  v-for="ingredient in item.semi_finished_goods?.recipe?.ingredients || []"
                  :key="'raw-' + ingredient.id"
                  class="ingredient-item"
                >
                  <span class="ingredient-name">{{ ingredient.ingredient?.name || 'Unknown' }}</span>
                  <span class="ingredient-quantity">{{ ingredient.quantity }} {{ ingredient.ingredient?.unit || 'gr' }}</span>
                </div>
              </template>
              <div v-if="!hasIngredients(recipe)" class="ingredient-item">
                <span class="ingredient-name no-data">Tidak ada bahan baku</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </a-spin>

    <!-- Recipe Form Modal -->
    <RecipeFormModal
      v-model:visible="formModalVisible"
      :recipe="selectedRecipe"
      @success="handleFormSuccess"
    />

    <!-- Recipe History Modal -->
    <RecipeHistoryModal
      v-model:visible="historyModalVisible"
      :recipe-id="selectedRecipeId"
    />

    <!-- Recipe View Modal (View Only) -->
    <RecipeViewModal
      v-model:visible="viewModalVisible"
      :recipe="selectedRecipe"
      @edit="editRecipe"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { 
  PlusOutlined, 
  MoreOutlined, 
  EyeOutlined, 
  EditOutlined, 
  HistoryOutlined, 
  DeleteOutlined,
  PictureOutlined,
  SearchOutlined
} from '@ant-design/icons-vue'
import recipeService from '@/services/recipeService'
import RecipeFormModal from '@/components/RecipeFormModal.vue'
import RecipeHistoryModal from '@/components/RecipeHistoryModal.vue'
import RecipeViewModal from '@/components/RecipeViewModal.vue'

const loading = ref(false)
const recipes = ref([])
const searchText = ref('')
const filterCategory = ref(undefined)
const selectedRecipe = ref(null)
const selectedRecipeId = ref(null)
const formModalVisible = ref(false)
const historyModalVisible = ref(false)
const viewModalVisible = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => `Total ${total} menu`
})

const columns = [
  {
    title: 'Nama Menu',
    dataIndex: 'name',
    key: 'name',
    width: '25%'
  },
  {
    title: 'Kategori',
    dataIndex: 'category',
    key: 'category',
    width: '15%'
  },
  {
    title: 'Informasi Gizi (per menu)',
    key: 'nutrition',
    width: '30%'
  },
  {
    title: 'Status',
    key: 'status',
    width: '10%',
    align: 'center'
  },
  {
    title: 'Aksi',
    key: 'actions',
    width: '20%',
    align: 'center'
  }
]

const getComponentColor = (category) => {
  const colors = {
    nasi: '#FFA500',
    lauk: '#FF69B4',
    sambal: '#FF0000',
    sayur: '#00CED1',
    lauk_berkuah: '#32CD32'
  }
  return colors[category] || '#999'
}

const getComponentTagColor = (category) => {
  const colors = {
    nasi: 'orange',
    lauk: 'pink',
    sambal: 'red',
    sayur: 'cyan',
    lauk_berkuah: 'green',
    protein_hewani: 'volcano',
    protein_nabati: 'lime',
    karbohidrat: 'gold',
    susu: 'blue',
    buah: 'magenta',
    minuman: 'geekblue',
    snack: 'purple',
    pelengkap: 'default'
  }
  return colors[category] || 'default'
}

const hasIngredients = (recipe) => {
  if (!recipe.recipe_items) return false
  return recipe.recipe_items.some(item => 
    item.semi_finished_goods?.recipe?.ingredients && 
    item.semi_finished_goods.recipe.ingredients.length > 0
  )
}

const getCategoryColor = (category) => {
  const colors = {
    masakan_indonesia: 'red',
    masakan_china: 'orange',
    masakan_western: 'blue',
    masakan_india: 'purple',
    masakan_gabungan: 'green',
    lainnya: 'default'
  }
  return colors[category] || 'default'
}

const getCategoryLabel = (category) => {
  const labels = {
    masakan_indonesia: 'Masakan Indonesia',
    masakan_china: 'Masakan China',
    masakan_western: 'Masakan Western',
    masakan_india: 'Masakan India',
    masakan_gabungan: 'Masakan Gabungan',
    lainnya: 'Lainnya',
    nasi: 'Nasi',
    lauk: 'Lauk',
    sambal: 'Sambal',
    sayur: 'Sayur',
    lauk_berkuah: 'Lauk Berkuah',
    protein_hewani: 'Protein Hewani',
    protein_nabati: 'Protein Nabati',
    karbohidrat: 'Karbohidrat',
    susu: 'Susu',
    buah: 'Buah',
    minuman: 'Minuman',
    snack: 'Snack',
    pelengkap: 'Pelengkap'
  }
  if (labels[category]) return labels[category]
  // Fallback: snake_case → Title Case
  if (category && category.includes('_')) {
    return category.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ')
  }
  return category || '-'
}

const fetchRecipes = async () => {
  loading.value = true
  try {
    const params = {
      q: searchText.value || undefined,
      category: filterCategory.value || undefined,
      active_only: false
    }
    
    const response = await recipeService.getRecipes(params)
    recipes.value = response.data.recipes || []
    pagination.total = recipes.value.length
  } catch (error) {
    message.error('Gagal memuat data menu')
    console.error('Error fetching recipes:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  fetchRecipes()
}

const handleTableChange = (pag) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
  fetchRecipes()
}

const showCreateModal = () => {
  selectedRecipe.value = null
  formModalVisible.value = true
}

const viewRecipe = (recipe) => {
  selectedRecipe.value = { ...recipe }
  viewModalVisible.value = true
}

const editRecipe = (recipe) => {
  selectedRecipe.value = { ...recipe }
  formModalVisible.value = true
}

const viewHistory = (recipe) => {
  selectedRecipeId.value = recipe.id
  historyModalVisible.value = true
}

const confirmDelete = (recipe) => {
  Modal.confirm({
    title: 'Hapus Menu',
    content: `Yakin ingin menghapus menu "${recipe.name}"?`,
    okText: 'Ya, Hapus',
    okType: 'danger',
    cancelText: 'Batal',
    onOk: () => deleteRecipe(recipe)
  })
}

const deleteRecipe = async (recipe) => {
  try {
    await recipeService.deleteRecipe(recipe.id)
    message.success('Menu berhasil dihapus')
    fetchRecipes()
  } catch (error) {
    message.error('Gagal menghapus menu')
    console.error('Error deleting recipe:', error)
  }
}

const handleFormSuccess = () => {
  formModalVisible.value = false
  fetchRecipes()
}

onMounted(() => {
  fetchRecipes()
})
</script>

<style scoped>
/* Header Actions */
.recipe-list-header {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-bottom: 16px;
}

.add-button {
  background: var(--h-primary);
  border-color: var(--h-primary);
  height: var(--h-touch-target-min);
  font-weight: var(--h-font-semibold);
  border-radius: var(--h-radius-md);
  transition: all var(--h-transition-base);
}

.add-button:hover {
  background: var(--h-primary-dark);
  border-color: var(--h-primary-dark);
  transform: scale(1.02);
}

/* Search and Filter Bar */
.search-filter-bar {
  padding: var(--h-spacing-5);
  margin-bottom: 16px;
}

/* Recipe Grid */
.recipe-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: var(--h-spacing-5);
}

/* Recipe Card */
.recipe-card {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-4);
  transition: all var(--h-transition-base);
}

/* Card Header */
.recipe-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: var(--h-spacing-3);
  border-bottom: 1px solid var(--h-border-light);
}

.category-tag {
  margin: 0;
  font-size: var(--h-text-xs);
  padding: 4px 12px;
  border-radius: var(--h-radius-sm);
  font-weight: var(--h-font-semibold);
}

.action-menu-trigger {
  color: var(--h-text-secondary);
  font-size: 18px;
  padding: var(--h-spacing-2);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--h-radius-sm);
  transition: all var(--h-transition-fast);
}

.action-menu-trigger:hover {
  color: var(--h-text-primary);
  background: var(--h-bg-light);
}

/* Recipe Title */
.recipe-card__title {
  font-size: var(--h-text-base);
  font-weight: var(--h-font-bold);
  color: var(--h-text-primary);
  line-height: var(--h-leading-tight);
  min-height: 44px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* Recipe Photo */
.recipe-card__photo {
  width: 100%;
  height: 200px;
  background: var(--h-bg-light);
  border-radius: var(--h-radius-md);
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.recipe-card__photo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform var(--h-transition-base);
}

.recipe-card:hover .recipe-card__photo img {
  transform: scale(1.05);
}

.no-photo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: var(--h-bg-secondary);
  color: var(--h-text-light);
  font-size: 48px;
}

/* Card Section */
.recipe-card__section {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-3);
}

.section-title {
  font-size: var(--h-text-sm);
  font-weight: var(--h-font-bold);
  color: var(--h-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Component List */
.component-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-2);
}

.component-item {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2);
  font-size: var(--h-text-sm);
}

.component-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.component-name {
  flex: 1;
  color: var(--h-text-primary);
  font-weight: var(--h-font-medium);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Ingredient List */
.ingredient-list {
  display: flex;
  flex-direction: column;
  gap: var(--h-spacing-2);
  max-height: 140px;
  overflow-y: auto;
  padding-right: var(--h-spacing-2);
}

.ingredient-list::-webkit-scrollbar {
  width: 4px;
}

.ingredient-list::-webkit-scrollbar-track {
  background: var(--h-bg-light);
  border-radius: var(--h-radius-sm);
}

.ingredient-list::-webkit-scrollbar-thumb {
  background: var(--h-border-color);
  border-radius: var(--h-radius-sm);
}

.ingredient-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: var(--h-text-sm);
  padding: var(--h-spacing-1) 0;
}

.ingredient-name {
  flex: 1;
  color: var(--h-text-secondary);
  font-weight: var(--h-font-normal);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ingredient-name.no-data {
  color: var(--h-text-light);
  font-style: italic;
}

.ingredient-quantity {
  font-weight: var(--h-font-semibold);
  color: var(--h-text-primary);
  flex-shrink: 0;
  margin-left: var(--h-spacing-2);
}

/* Responsive - Tablet */
@media (max-width: 1024px) {
  .recipe-grid {
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  }
}

/* Responsive - Mobile */
@media (max-width: 767px) {
  .recipe-list-header {
    flex-direction: column;
    align-items: stretch;
  }
  
  .add-button {
    width: 100%;
  }
  
  .recipe-grid {
    grid-template-columns: 1fr;
  }
  
  .recipe-card__photo {
    height: 180px;
  }
}

/* Dark Mode Support */
.dark .recipe-card__header {
  border-bottom-color: var(--h-border-color);
}

.dark .action-menu-trigger {
  color: var(--h-text-secondary);
}

.dark .action-menu-trigger:hover {
  color: var(--h-text-primary);
  background: rgba(163, 174, 208, 0.1);
}

.dark .recipe-card__title {
  color: var(--h-text-primary);
}

.dark .recipe-card__photo {
  background: rgba(163, 174, 208, 0.05);
}

.dark .no-photo {
  background: rgba(163, 174, 208, 0.1);
  color: var(--h-text-light);
}

.dark .section-title {
  color: var(--h-text-secondary);
}

.dark .component-name {
  color: var(--h-text-primary);
}

.dark .ingredient-name {
  color: var(--h-text-secondary);
}

.dark .ingredient-quantity {
  color: var(--h-text-primary);
}

.dark .ingredient-list::-webkit-scrollbar-track {
  background: rgba(163, 174, 208, 0.05);
}

.dark .ingredient-list::-webkit-scrollbar-thumb {
  background: var(--h-border-color);
}
</style>
