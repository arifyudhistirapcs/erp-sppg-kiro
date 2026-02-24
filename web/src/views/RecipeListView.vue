<template>
  <div class="recipe-list">
    <a-page-header
      title="Manajemen Resep"
      sub-title="Kelola resep dan informasi gizi"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Tambah Resep
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <!-- Search and Filter -->
      <a-row :gutter="16" class="mb-4">
        <a-col :span="8">
          <a-input-search
            v-model:value="searchText"
            placeholder="Cari nama resep..."
            @search="handleSearch"
            allow-clear
          />
        </a-col>
        <a-col :span="6">
          <a-select
            v-model:value="filterCategory"
            placeholder="Kategori"
            style="width: 100%"
            allow-clear
            @change="handleSearch"
          >
            <a-select-option value="makanan_pokok">Makanan Pokok</a-select-option>
            <a-select-option value="lauk_pauk">Lauk Pauk</a-select-option>
            <a-select-option value="sayuran">Sayuran</a-select-option>
            <a-select-option value="buah">Buah</a-select-option>
            <a-select-option value="minuman">Minuman</a-select-option>
          </a-select>
        </a-col>
      </a-row>

      <!-- Table -->
      <a-table
        :columns="columns"
        :data-source="recipes"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <a @click="viewRecipe(record)">{{ record.name }}</a>
          </template>
          <template v-else-if="column.key === 'category'">
            <a-tag :color="getCategoryColor(record.category)">
              {{ getCategoryLabel(record.category) }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'nutrition'">
            <div class="nutrition-summary">
              <div><strong>{{ record.total_calories?.toFixed(0) }}</strong> kkal</div>
              <div class="text-muted">P: {{ record.total_protein?.toFixed(1) }}g | K: {{ record.total_carbs?.toFixed(1) }}g | L: {{ record.total_fat?.toFixed(1) }}g</div>
            </div>
          </template>
          <template v-else-if="column.key === 'status'">
            <a-tag :color="record.is_active ? 'green' : 'red'">
              {{ record.is_active ? 'Aktif' : 'Nonaktif' }}
            </a-tag>
          </template>
          <template v-else-if="column.key === 'actions'">
            <a-space>
              <a-button type="link" size="small" @click="viewRecipe(record)">
                Lihat
              </a-button>
              <a-button type="link" size="small" @click="editRecipe(record)">
                Edit
              </a-button>
              <a-button type="link" size="small" @click="viewHistory(record)">
                Riwayat
              </a-button>
              <a-popconfirm
                title="Yakin ingin menghapus resep ini?"
                ok-text="Ya"
                cancel-text="Tidak"
                @confirm="deleteRecipe(record)"
              >
                <a-button type="link" danger size="small">
                  Hapus
                </a-button>
              </a-popconfirm>
            </a-space>
          </template>
        </template>
      </a-table>
    </a-card>

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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import recipeService from '@/services/recipeService'
import RecipeFormModal from '@/components/RecipeFormModal.vue'
import RecipeHistoryModal from '@/components/RecipeHistoryModal.vue'

const loading = ref(false)
const recipes = ref([])
const searchText = ref('')
const filterCategory = ref(undefined)
const selectedRecipe = ref(null)
const selectedRecipeId = ref(null)
const formModalVisible = ref(false)
const historyModalVisible = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => `Total ${total} resep`
})

const columns = [
  {
    title: 'Nama Resep',
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
    title: 'Porsi',
    dataIndex: 'serving_size',
    key: 'serving_size',
    width: '10%',
    align: 'center'
  },
  {
    title: 'Informasi Gizi (per porsi)',
    key: 'nutrition',
    width: '25%'
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
    width: '15%',
    align: 'center'
  }
]

const getCategoryColor = (category) => {
  const colors = {
    makanan_pokok: 'blue',
    lauk_pauk: 'orange',
    sayuran: 'green',
    buah: 'purple',
    minuman: 'cyan'
  }
  return colors[category] || 'default'
}

const getCategoryLabel = (category) => {
  const labels = {
    makanan_pokok: 'Makanan Pokok',
    lauk_pauk: 'Lauk Pauk',
    sayuran: 'Sayuran',
    buah: 'Buah',
    minuman: 'Minuman'
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
    message.error('Gagal memuat data resep')
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
  formModalVisible.value = true
}

const editRecipe = (recipe) => {
  selectedRecipe.value = { ...recipe }
  formModalVisible.value = true
}

const viewHistory = (recipe) => {
  selectedRecipeId.value = recipe.id
  historyModalVisible.value = true
}

const deleteRecipe = async (recipe) => {
  try {
    await recipeService.deleteRecipe(recipe.id)
    message.success('Resep berhasil dihapus')
    fetchRecipes()
  } catch (error) {
    message.error('Gagal menghapus resep')
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
}

.mb-4 {
  margin-bottom: 16px;
}

.nutrition-summary {
  font-size: 12px;
}

.text-muted {
  color: #8c8c8c;
  font-size: 11px;
}
</style>
