<template>
  <a-modal
    :visible="visible"
    title="Riwayat Versi Resep"
    width="800px"
    :footer="null"
    @cancel="handleCancel"
  >
    <a-spin :spinning="loading">
      <a-timeline v-if="history.length > 0">
        <a-timeline-item
          v-for="item in history"
          :key="item.version"
          :color="item.version === currentVersion ? 'green' : 'blue'"
        >
          <template #dot>
            <span v-if="item.version === currentVersion" class="current-badge">
              Versi Saat Ini
            </span>
          </template>
          
          <a-card size="small">
            <template #title>
              <a-space>
                <span>Versi {{ item.version }}</span>
                <a-tag v-if="item.version === currentVersion" color="green">Aktif</a-tag>
              </a-space>
            </template>
            
            <a-descriptions :column="2" size="small">
              <a-descriptions-item label="Tanggal">
                {{ formatDate(item.updated_at) }}
              </a-descriptions-item>
              <a-descriptions-item label="Diubah Oleh">
                {{ item.updated_by_name || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="Nama Resep" :span="2">
                {{ item.name }}
              </a-descriptions-item>
              <a-descriptions-item label="Kategori">
                {{ getCategoryLabel(item.category) }}
              </a-descriptions-item>
              <a-descriptions-item label="Porsi">
                {{ item.serving_size }}
              </a-descriptions-item>
              <a-descriptions-item label="Kalori Total">
                {{ item.total_calories?.toFixed(0) }} kkal
              </a-descriptions-item>
              <a-descriptions-item label="Protein Total">
                {{ item.total_protein?.toFixed(1) }} g
              </a-descriptions-item>
            </a-descriptions>

            <a-divider style="margin: 12px 0" />

            <div v-if="item.changes && item.changes.length > 0">
              <strong>Perubahan:</strong>
              <ul style="margin-top: 8px; padding-left: 20px">
                <li v-for="(change, idx) in item.changes" :key="idx">
                  {{ change }}
                </li>
              </ul>
            </div>
          </a-card>
        </a-timeline-item>
      </a-timeline>

      <a-empty v-else description="Tidak ada riwayat versi" />
    </a-spin>
  </a-modal>
</template>

<script setup>
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import recipeService from '@/services/recipeService'
import dayjs from 'dayjs'

const props = defineProps({
  visible: {
    type: Boolean,
    required: true
  },
  recipeId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['update:visible'])

const loading = ref(false)
const history = ref([])
const currentVersion = ref(null)

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

const formatDate = (date) => {
  return dayjs(date).format('DD/MM/YYYY HH:mm')
}

const loadHistory = async () => {
  if (!props.recipeId) return

  loading.value = true
  try {
    const response = await recipeService.getRecipeHistory(props.recipeId)
    history.value = response.data.data || []
    
    // Get current version
    const recipeResponse = await recipeService.getRecipe(props.recipeId)
    currentVersion.value = recipeResponse.data.data?.version
  } catch (error) {
    message.error('Gagal memuat riwayat resep')
    console.error('Error loading recipe history:', error)
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  emit('update:visible', false)
}

watch(() => props.visible, (newVal) => {
  if (newVal && props.recipeId) {
    loadHistory()
  }
})
</script>

<style scoped>
.current-badge {
  background: #52c41a;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}
</style>
