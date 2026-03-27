<script setup lang="ts">
defineProps<{
  avatarColor: string
  avatarText: string
  displayName: string
  handle: string
  timeAgo: string
  isNew?: boolean
}>()
</script>

<template>
  <div class="survey-user-header">
    <div class="survey-avatar" :style="{ background: avatarColor }">
      {{ avatarText }}
    </div>
    <div class="survey-meta">
      <div class="survey-meta-row">
        <span class="survey-display-name">{{ displayName }}</span>
        <span v-if="isNew" class="survey-new-badge">Yeni</span>
        <span class="survey-time">{{ timeAgo }}</span>
      </div>
      <div class="survey-handle">{{ handle }}</div>
    </div>
  </div>
</template>

<style scoped>
.survey-user-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.survey-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 16px;
  color: #fff;
  flex-shrink: 0;
}
.survey-meta {
  flex: 1;
  min-width: 0;
}
.survey-meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  min-width: 0;
}
.survey-display-name {
  font-weight: 700;
  font-size: 15px;
  color: var(--text-color);
  overflow-wrap: anywhere;
  word-break: break-word;
}
.survey-handle,
.survey-time {
  font-size: 13px;
  color: var(--text-color-muted);
  overflow-wrap: anywhere;
  word-break: break-word;
}
.survey-new-badge {
  background: var(--primary-color);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 9999px;
}
</style>
