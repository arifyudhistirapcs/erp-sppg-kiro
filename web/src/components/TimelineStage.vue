<template>
  <div class="timeline-stage" :class="{ 'is-last': isLast }">
    <div class="stage-line" v-if="!isLast"></div>
    <div class="stage-indicator" :class="getIndicatorClass()">
      <check-circle-filled v-if="isCompleted" />
      <loading-outlined v-else-if="isInProgress" spin />
      <div v-else class="empty-circle"></div>
    </div>
    <div class="stage-content">
      <div class="stage-header">
        <h3 class="stage-title">{{ stage.title }}</h3>
        <div v-if="stage.completed_at" class="stage-timestamp">
          {{ formatTimeRange(stage.started_at, stage.completed_at) }}
        </div>
      </div>
      <p class="stage-description">{{ stage.description }}</p>
      
      <div v-if="stage.media" class="stage-media">
        <div v-if="stage.media.type === 'photo'" class="media-photo" @click="openMedia">
          <img :src="stage.media.thumbnail_url || stage.media.url" :alt="stage.title" />
        </div>
        <div v-else-if="stage.media.type === 'video'" class="media-video" @click="openMedia">
          <div class="video-thumbnail">
            <img :src="stage.media.thumbnail_url" :alt="stage.title" />
            <div class="play-button">
              <play-circle-outlined style="font-size: 48px; color: #fff;" />
            </div>
          </div>
        </div>
      </div>

      <div v-if="stage.transitioned_by" class="stage-user">
        <user-outlined />
        <span>{{ stage.transitioned_by.name }}</span>
      </div>
    </div>

    <a-modal
      v-model:visible="showMediaModal"
      :footer="null"
      :width="800"
      centered
    >
      <div v-if="stage.media" class="media-modal-content">
        <img
          v-if="stage.media.type === 'photo'"
          :src="stage.media.url"
          :alt="stage.title"
          style="width: 100%;"
        />
        <video
          v-else-if="stage.media.type === 'video'"
          :src="stage.media.url"
          controls
          style="width: 100%;"
        ></video>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, defineProps } from 'vue';
import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import 'dayjs/locale/id';
import {
  CheckCircleFilled,
  LoadingOutlined,
  PlayCircleOutlined,
  UserOutlined,
} from '@ant-design/icons-vue';

dayjs.extend(utc);
dayjs.locale('id');

defineProps({
  stage: {
    type: Object,
    required: true,
  },
  isCompleted: {
    type: Boolean,
    default: false,
  },
  isInProgress: {
    type: Boolean,
    default: false,
  },
  isLast: {
    type: Boolean,
    default: false,
  },
});

const showMediaModal = ref(false);

const getIndicatorClass = () => {
  const props = defineProps({
    stage: Object,
    isCompleted: Boolean,
    isInProgress: Boolean,
    isLast: Boolean,
  });
  
  if (props.isCompleted) return 'completed';
  if (props.isInProgress) return 'in-progress';
  return 'pending';
};

const formatTimeRange = (startTime, endTime) => {
  if (!startTime || !endTime) return '';
  
  // Debug: log raw values
  console.log('=== formatTimeRange Debug ===');
  console.log('startTime:', startTime);
  console.log('endTime:', endTime);
  console.log('typeof startTime:', typeof startTime);
  
  // Extract date and time directly from ISO string
  // Format: "2026-03-02T21:26:04.483536+07:00"
  // We want: "Minggu, 02 Mar 2026, 21:26 - 21:32 WIB"
  
  // Extract date parts using regex
  const startMatch = startTime.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/);
  const endMatch = endTime.match(/^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/);
  
  console.log('startMatch:', startMatch);
  console.log('endMatch:', endMatch);
  
  if (!startMatch || !endMatch) return '';
  
  const [, startYear, startMonth, startDay, startHour, startMin] = startMatch;
  const [, endYear, endMonth, endDay, endHour, endMin] = endMatch;
  
  console.log('Extracted:', { startYear, startMonth, startDay, startHour, startMin });
  
  // Create date string for dayjs (without timezone)
  const startDateStr = `${startYear}-${startMonth}-${startDay} ${startHour}:${startMin}`;
  const endDateStr = `${endYear}-${endMonth}-${endDay} ${endHour}:${endMin}`;
  
  console.log('startDateStr:', startDateStr);
  console.log('endDateStr:', endDateStr);
  
  // Parse with dayjs
  const start = dayjs(startDateStr, 'YYYY-MM-DD HH:mm');
  const end = dayjs(endDateStr, 'YYYY-MM-DD HH:mm');
  
  console.log('Parsed start:', start.format('YYYY-MM-DD HH:mm'));
  console.log('Parsed end:', end.format('YYYY-MM-DD HH:mm'));
  
  // Format the values
  const startDayName = start.format('dddd');
  const startDate24 = start.format('DD MMM YYYY');
  const startTime24 = start.format('HH:mm');
  const endDayName = end.format('dddd');
  const endDate24 = end.format('DD MMM YYYY');
  const endTime24 = end.format('HH:mm');
  
  console.log('Final formatted:', startDate24);
  console.log('===========================');
  
  // Check if same day
  if (startDate24 === endDate24) {
    return `${startDayName}, ${startDate24}, ${startTime24} - ${endTime24} WIB`;
  }
  
  return `${startDayName}, ${startDate24}, ${startTime24} - ${endDayName}, ${endDate24}, ${endTime24} WIB`;
};

const openMedia = () => {
  showMediaModal.value = true;
};
</script>

<style scoped>
.timeline-stage {
  position: relative;
  display: flex;
  padding-bottom: var(--h-spacing-8, 32px);
}

.timeline-stage.is-last {
  padding-bottom: 0;
}

.stage-line {
  position: absolute;
  left: 15px;
  top: 32px;
  bottom: 0;
  width: 2px;
  background: var(--h-border-color, #E9EDF7);
  transition: background-color var(--transition-base, 200ms);
}

.dark .stage-line {
  background: var(--h-border-color-dark, #5A4372);
}

.stage-indicator {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  z-index: 1;
  background: var(--h-bg-card, #FFFFFF);
  transition: all var(--transition-base, 200ms);
}

.dark .stage-indicator {
  background: var(--h-bg-card-dark, #3D2B53);
}

.stage-indicator.completed {
  color: var(--success, #05CD99);
  font-size: 32px;
}

.stage-indicator.in-progress {
  color: var(--h-primary, #5A4372);
  font-size: 24px;
  border: 2px solid var(--h-primary, #5A4372);
  animation: pulse-border 2s ease-in-out infinite;
}

.stage-indicator.pending {
  border: 2px solid var(--h-border-color, #E9EDF7);
}

.dark .stage-indicator.pending {
  border-color: var(--h-border-color-dark, #5A4372);
}

@keyframes pulse-border {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(90, 67, 114, 0.4);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(90, 67, 114, 0);
  }
}

.empty-circle {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--h-bg-card, #FFFFFF);
}

.dark .empty-circle {
  background: var(--h-bg-card-dark, #3D2B53);
}

.stage-content {
  flex: 1;
  margin-left: var(--h-spacing-4, 16px);
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--h-spacing-2, 8px);
  gap: var(--h-spacing-4, 16px);
}

.stage-title {
  font-size: var(--text-base, 16px);
  font-weight: 600;
  margin: 0;
  color: var(--h-text-primary, #322837);
  transition: color var(--transition-base, 200ms);
}

.dark .stage-title {
  color: var(--h-text-primary-dark, #F8FDEA);
}

.stage-timestamp {
  font-size: var(--text-xs, 12px);
  color: var(--h-text-secondary, #74788C);
  white-space: nowrap;
  font-weight: 500;
  padding: 4px var(--h-spacing-2, 8px);
  background: var(--h-bg-primary, #F8FDEA);
  border-radius: var(--h-radius-sm, 8px);
  transition: all var(--transition-base, 200ms);
}

.dark .stage-timestamp {
  color: var(--h-text-secondary-dark, #ACA9B0);
  background: rgba(90, 67, 114, 0.2);
}

.stage-description {
  font-size: var(--text-sm, 14px);
  color: var(--h-text-secondary, #74788C);
  margin-bottom: var(--h-spacing-3, 12px);
  line-height: 1.6;
  transition: color var(--transition-base, 200ms);
}

.dark .stage-description {
  color: var(--h-text-secondary-dark, #ACA9B0);
}

.stage-media {
  margin-bottom: var(--h-spacing-3, 12px);
}

.media-photo,
.media-video {
  cursor: pointer;
  border-radius: var(--h-radius-md, 12px);
  overflow: hidden;
  max-width: 240px;
  transition: all var(--transition-base, 200ms);
  box-shadow: var(--h-shadow-md, 0px 4px 6px rgba(0, 0, 0, 0.07));
}

.media-photo:hover,
.media-video:hover {
  transform: translateY(-4px);
  box-shadow: var(--h-shadow-lg, 0px 10px 15px rgba(0, 0, 0, 0.1));
}

.media-photo img {
  width: 100%;
  height: auto;
  display: block;
}

.video-thumbnail {
  position: relative;
  width: 100%;
}

.video-thumbnail img {
  width: 100%;
  height: auto;
  display: block;
}

.play-button {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(90, 67, 114, 0.8);
  border-radius: 50%;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-base, 200ms);
}

.media-video:hover .play-button {
  background: rgba(90, 67, 114, 0.95);
  transform: translate(-50%, -50%) scale(1.1);
}

.stage-user {
  display: flex;
  align-items: center;
  gap: var(--h-spacing-2, 8px);
  font-size: var(--text-xs, 12px);
  color: var(--h-text-secondary, #74788C);
  padding: var(--h-spacing-2, 8px) var(--h-spacing-3, 12px);
  background: var(--h-bg-primary, #F8FDEA);
  border-radius: var(--h-radius-sm, 8px);
  display: inline-flex;
  font-weight: 500;
  transition: all var(--transition-base, 200ms);
}

.dark .stage-user {
  color: var(--h-text-secondary-dark, #ACA9B0);
  background: rgba(90, 67, 114, 0.2);
}

.media-modal-content {
  padding: var(--h-spacing-4, 16px) 0;
}

/* Responsive Design */
@media (max-width: 768px) {
  .stage-header {
    flex-direction: column;
    gap: var(--h-spacing-2, 8px);
  }

  .stage-timestamp {
    white-space: normal;
  }

  .media-photo,
  .media-video {
    max-width: 100%;
  }
}
</style>
