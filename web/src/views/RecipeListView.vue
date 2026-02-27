<template>
  <div class="recipe-list">
    <div class="page-header">
      <h2 class="page-title">Daftar Menu</h2>
      <a-button type="primary" @click="showCreateModal" class="add-button">
        <template #icon><PlusOutlined /></template>
        Tambah Menu Baru
      </a-button>
    </div>

    <!-- Search -->
    <div class="search-section">
      <a-input
        v-model:value="searchText"
        placeholder="Cari menu"
        @change="handleSearch"
        allow-clear
        class="search-input"
      />
      <a-select
        v-model:value="filterCategory"
        placeholder="Semua Kategori"
        style="width: 200px"
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
    </div>

    <!-- Grid Cards -->
    <a-spin :spinning="loading">
      <a-empty v-if="!loading && recipes.length === 0" description="Tidak ada menu" />
      
      <div v-else class="recipe-grid">
        <a-card
          v-for="recipe in recipes"
          :key="recipe.id"
          class="recipe-card"
          :body-style="{ padding: 0 }"
        >
          <!-- Header with Category and Menu -->
          <div class="card-header">
            <a-tag :color="getCategoryColor(recipe.category)" class="category-tag">
              {{ getCategoryLabel(recipe.category) }}
            </a-tag>
            <a-dropdown :trigger="['click']">
              <a class="ant-dropdown-link" @click.prevent>
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

          <div class="card-content">
            <!-- Title -->
            <div class="card-title">
              {{ recipe.name }}
            </div>

            <!-- Photo -->
            <div class="card-photo">
              <img 
                v-if="recipe.photo_url" 
                :src="recipe.photo_url" 
                :alt="recipe.name"
              />
              <div v-else class="no-photo">
                <PictureOutlined style="font-size: 48px; color: #d9d9d9" />
              </div>
            </div>

            <!-- Components -->
            <div class="card-section">
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

            <!-- Ingredients -->
            <div class="card-section">
              <div class="section-title">Bahan Baku</div>
              <div class="ingredient-list">
                <div
                  v-for="item in recipe.recipe_items"
                  :key="'ing-' + item.id"
                  class="ingredient-item"
                >
                  <span class="ingredient-name">{{ item.semi_finished_goods?.name }}</span>
                  <span class="ingredient-quantity">{{ item.quantity }} gr</span>
                </div>
              </div>
            </div>
          </div>
        </a-card>
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
  PictureOutlined
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
    lauk_berkuah: 'green'
  }
  return colors[category] || 'default'
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
    lauk_berkuah: 'Lauk Berkuah'
  }
  return labels[category] || category
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
.recipe-list {
  padding: 24px;
  background-color: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
  color: #262626;
}

.add-button {
  background-color: #d4af37;
  border-color: #d4af37;
  color: white;
  font-weight: 500;
  height: 40px;
  padding: 0 24px;
  border-radius: 4px;
}

.add-button:hover {
  background-color: #c19b2b;
  border-color: #c19b2b;
}

.search-section {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.search-input {
  flex: 1;
  max-width: 400px;
}

.recipe-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 20px;
}

.recipe-card {
  background: white;
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.3s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.recipe-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f0f0f0;
}

.category-tag {
  margin: 0;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
}

.card-content {
  padding: 16px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: #262626;
  margin-bottom: 12px;
  line-height: 1.4;
  min-height: 42px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-photo {
  width: 100%;
  height: 180px;
  background-color: #f5f5f5;
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-photo img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-photo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background-color: #fafafa;
}

.card-section {
  margin-bottom: 16px;
}

.card-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #595959;
  margin-bottom: 8px;
}

.component-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.component-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.component-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.component-name {
  flex: 1;
  color: #262626;
  font-weight: 500;
}

.ingredient-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 120px;
  overflow-y: auto;
}

.ingredient-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 12px;
  color: #595959;
  padding: 2px 0;
}

.ingredient-name {
  flex: 1;
}

.ingredient-quantity {
  font-weight: 500;
  color: #262626;
}

.ant-dropdown-link {
  color: #595959;
  font-size: 18px;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ant-dropdown-link:hover {
  color: #262626;
}
</style>
