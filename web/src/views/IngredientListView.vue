<template>
  <div class="ingredient-list">
    <a-page-header
      title="Manajemen Bahan"
      sub-title="Kelola bahan baku untuk resep"
    >
      <template #extra>
        <a-button type="primary" @click="showCreateModal">
          <template #icon><PlusOutlined /></template>
          Tambah Bahan Baru
        </a-button>
      </template>
    </a-page-header>

    <a-card>
      <!-- Search -->
      <a-row :gutter="16" class="mb-4">
        <a-col :span="8">
          <a-input-search
            v-model:value="searchText"
            placeholder="Cari nama bahan..."
            @search="handleSearch"
            allow-clear
          />
        </a-col>
      </a-row>

      <!-- Table -->
      <a-table
        :columns="columns"
        :data-source="ingredients"
        :loading="loading"
        :pagination="pagination"
        @change="handleTableChange"
        row-key="id"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'name'">
            <strong>{{ record.name }}</strong>
          </template>
          <template v-else-if="column.key === 'nutrition'">
            <div class="nutrition-info">
              <a-tag color="red">{{ record.calories_per100g }} kkal</a-tag>
              <a-tag color="blue">P: {{ record.protein_per100g }}g</a-tag>
              <a-tag color="green">K: {{ record.carbs_per100g }}g</a-tag>
              <a-tag color="orange">L: {{ record.fat_per100g }}g</a-tag>
            </div>
          </template>
          <template v-else-if="column.key === 'unit'">
            <a-tag>{{ record.unit }}</a-tag>
          </template>
        </template>
      </a-table>
    </a-card>

    <!-- Ingredient Form Modal -->
    <IngredientFormModal
      v-model:visible="formModalVisible"
      @success="handleFormSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined } from '@ant-design/icons-vue'
import recipeService from '@/services/recipeService'
import IngredientFormModal from '@/components/IngredientFormModal.vue'

const loading = ref(false)
const ingredients = ref([])
const searchText = ref('')
const formModalVisible = ref(false)

const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
  showSizeChanger: true,
  showTotal: (total) => `Total ${total} bahan`
})

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: '8%'
  },
  {
    title: 'Nama Bahan',
    dataIndex: 'name',
    key: 'name',
    width: '25%'
  },
  {
    title: 'Satuan',
    dataIndex: 'unit',
    key: 'unit',
    width: '12%',
    align: 'center'
  },
  {
    title: 'Informasi Gizi (per 100g)',
    key: 'nutrition',
    width: '40%'
  },
  {
    title: 'Dibuat',
    dataIndex: 'created_at',
    key: 'created_at',
    width: '15%',
    customRender: ({ text }) => {
      return new Date(text).toLocaleDateString('id-ID')
    }
  }
]

const fetchIngredients = async () => {
  loading.value = true
  try {
    const params = {
      search: searchText.value || undefined
    }
    
    const response = await recipeService.getIngredients(params)
    ingredients.value = response.data.data || []
    pagination.total = ingredients.value.length
  } catch (error) {
    message.error('Gagal memuat data bahan')
    console.error('Error fetching ingredients:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.current = 1
  fetchIngredients()
}

const handleTableChange = (pag) => {
  pagination.current = pag.current
  pagination.pageSize = pag.pageSize
}

const showCreateModal = () => {
  formModalVisible.value = true
}

const handleFormSuccess = () => {
  formModalVisible.value = false
  fetchIngredients()
}

onMounted(() => {
  fetchIngredients()
})
</script>

<style scoped>
.ingredient-list {
  padding: 24px;
}

.mb-4 {
  margin-bottom: 16px;
}

.nutrition-info {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}
</style>
