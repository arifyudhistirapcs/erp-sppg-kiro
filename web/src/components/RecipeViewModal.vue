<template>
  <a-modal
    :open="visible"
    title="Detail Resep"
    width="700px"
    :footer="null"
    @cancel="handleCancel"
  >
    <a-spin :spinning="loading">
      <template v-if="recipe">
        <!-- Basic Info -->
        <a-descriptions bordered :column="2" size="small" class="mb-4">
          <a-descriptions-item label="Nama Resep" :span="2">
            <strong>{{ recipe.name }}</strong>
          </a-descriptions-item>
          <a-descriptions-item label="Kategori">
            <a-tag :color="getCategoryColor(recipe.category)">
              {{ getCategoryLabel(recipe.category) }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Ukuran Porsi">
            {{ recipe.serving_size }} porsi
          </a-descriptions-item>
          <a-descriptions-item label="Versi">
            v{{ recipe.version }}
          </a-descriptions-item>
          <a-descriptions-item label="Status">
            <a-tag :color="recipe.is_active ? 'green' : 'red'">
              {{ recipe.is_active ? 'Aktif' : 'Nonaktif' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="Dibuat Oleh" v-if="recipe.creator">
            {{ recipe.creator.full_name }}
          </a-descriptions-item>
          <a-descriptions-item label="Tanggal Dibuat">
            {{ formatDate(recipe.created_at) }}
          </a-descriptions-item>
          <a-descriptions-item label="Terakhir Diupdate">
            {{ formatDate(recipe.updated_at) }}
          </a-descriptions-item>
        </a-descriptions>

        <!-- Nutrition Info -->
        <a-card size="small" title="Informasi Gizi Total" class="mb-4">
          <a-row :gutter="16">
            <a-col :span="6">
              <div class="nutrition-item">
                <div class="nutrition-value">{{ recipe.total_calories?.toFixed(0) }}</div>
                <div class="nutrition-label">kkal</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="nutrition-item">
                <div class="nutrition-value">{{ recipe.total_protein?.toFixed(1) }}g</div>
                <div class="nutrition-label">Protein</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="nutrition-item">
                <div class="nutrition-value">{{ recipe.total_carbs?.toFixed(1) }}g</div>
                <div class="nutrition-label">Karbohidrat</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="nutrition-item">
                <div class="nutrition-value">{{ recipe.total_fat?.toFixed(1) }}g</div>
                <div class="nutrition-label">Lemak</div>
              </div>
            </a-col>
          </a-row>
          <a-divider style="margin: 12px 0" />
          <a-row :gutter="16">
            <a-col :span="6">
              <div class="nutrition-item per-portion">
                <div class="nutrition-value">{{ caloriesPerPortion }}</div>
                <div class="nutrition-label">kkal/porsi</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="nutrition-item per-portion">
                <div class="nutrition-value">{{ proteinPerPortion }}g</div>
                <div class="nutrition-label">Protein/porsi</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="nutrition-item per-portion">
                <div class="nutrition-value">{{ carbsPerPortion }}g</div>
                <div class="nutrition-label">Karbo/porsi</div>
              </div>
            </a-col>
            <a-col :span="6">
              <div class="nutrition-item per-portion">
                <div class="nutrition-value">{{ fatPerPortion }}g</div>
                <div class="nutrition-label">Lemak/porsi</div>
              </div>
            </a-col>
          </a-row>
        </a-card>

        <!-- Instructions -->
        <a-card size="small" title="Instruksi" class="mb-4" v-if="recipe.instructions">
          <pre style="white-space: pre-wrap; font-family: inherit; margin: 0;">{{ recipe.instructions }}</pre>
        </a-card>

        <!-- Recipe Items -->
        <a-card size="small" title="Komponen Resep">
          <a-table
            :data-source="recipeItems"
            :pagination="false"
            size="small"
            row-key="id"
          >
            <a-table-column title="No" width="50" align="center">
              <template #default="{ index }">{{ index + 1 }}</template>
            </a-table-column>
            <a-table-column title="Nama" data-index="semi_finished_goods_name" />
            <a-table-column title="Kuantitas" width="120" align="right">
              <template #default="{ record }">
                {{ record.quantity }} {{ record.unit }}
              </template>
            </a-table-column>
          </a-table>
        </a-card>

        <!-- Actions -->
        <div style="margin-top: 24px; text-align: right;">
          <a-space>
            <a-button @click="handleCancel">Tutup</a-button>
            <a-button type="primary" @click="handleEdit">
              <template #icon>
                <EditOutlined />
              </template>
              Edit Resep
            </a-button>
          </a-space>
        </div>
      </template>
    </a-spin>
  </a-modal>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { EditOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  recipe: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['update:visible', 'edit'])

const loading = ref(false)

// Computed values for per-portion nutrition
const caloriesPerPortion = computed(() => {
  if (!props.recipe || !props.recipe.serving_size) return 0
  return (props.recipe.total_calories / props.recipe.serving_size).toFixed(0)
})

const proteinPerPortion = computed(() => {
  if (!props.recipe || !props.recipe.serving_size) return 0
  return (props.recipe.total_protein / props.recipe.serving_size).toFixed(1)
})

const carbsPerPortion = computed(() => {
  if (!props.recipe || !props.recipe.serving_size) return 0
  return (props.recipe.total_carbs / props.recipe.serving_size).toFixed(1)
})

const fatPerPortion = computed(() => {
  if (!props.recipe || !props.recipe.serving_size) return 0
  return (props.recipe.total_fat / props.recipe.serving_size).toFixed(1)
})

// Recipe items with processed data
const recipeItems = computed(() => {
  if (!props.recipe || !props.recipe.recipe_items) return []
  return props.recipe.recipe_items.map(item => ({
    ...item,
    semi_finished_goods_name: item.semi_finished_goods?.name || '-',
    unit: item.semi_finished_goods?.unit || ''
  }))
})

const getCategoryLabel = (category) => {
  const labels = {
    masakan_indonesia: 'Masakan Indonesia',
    masakan_china: 'Masakan China',
    masakan_western: 'Masakan Western',
    masakan_india: 'Masakan India',
    masakan_gabungan: 'Masakan Gabungan',
    lainnya: 'Lainnya'
  }
  return labels[category] || category
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

const formatDate = (date) => {
  if (!date) return '-'
  return dayjs(date).format('DD/MM/YYYY HH:mm')
}

const handleCancel = () => {
  emit('update:visible', false)
}

const handleEdit = () => {
  emit('edit', props.recipe)
  emit('update:visible', false)
}
</script>

<style scoped>
.mb-4 {
  margin-bottom: 16px;
}

.nutrition-item {
  text-align: center;
  padding: 8px;
  background: #f5f5f5;
  border-radius: 4px;
}

.nutrition-item.per-portion {
  background: #e6f7ff;
}

.nutrition-value {
  font-size: 18px;
  font-weight: bold;
  color: #1890ff;
}

.nutrition-label {
  font-size: 12px;
  color: #666;
  margin-top: 4px;
}

pre {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 4px;
}
</style>
