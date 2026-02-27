<template>
  <div class="vertical-timeline">
    <TimelineStage
      v-for="(stage, index) in timeline"
      :key="stage.stage"
      :stage="stage"
      :is-completed="stage.is_completed"
      :is-in-progress="stage.stage === currentStage"
      :is-last="index === timeline.length - 1"
    />
  </div>
</template>

<script setup>
import { defineProps } from 'vue';
import TimelineStage from './TimelineStage.vue';

defineProps({
  timeline: {
    type: Array,
    required: true,
  },
  currentStage: {
    type: Number,
    required: true,
  },
});
</script>

<style scoped>
.vertical-timeline {
  position: relative;
  padding: 0;
}
</style>
