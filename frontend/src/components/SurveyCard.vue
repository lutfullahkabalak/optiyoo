<script setup lang="ts">
import { computed, ref } from 'vue'

const props = defineProps({
  survey: {
    type: Object,
    required: true
  },
  isCompleted: {
    type: Boolean,
    default: false
  },
  givenAnswer: {
    type: String,
    default: null
  },
  isSubmitting: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['vote', 'share'])

const copiedLink = ref(false)

const handleShare = async () => {
  emit('share', props.survey.id)
  copiedLink.value = true
  setTimeout(() => (copiedLink.value = false), 1500)
}

const handleVote = (optionId: string) => {
  if (props.isCompleted || props.isSubmitting) return
  emit('vote', props.survey, optionId)
}

const q = computed(() => {
  if (props.survey.questions && props.survey.questions.length > 0) {
    return props.survey.questions[0]
  }
  return null
})

const totalVotes = computed(() => {
  if (!q.value || !q.value.options) return 0
  return q.value.options.reduce((sum: number, o: any) => sum + (o.vote_count || 0), 0)
})

const getPercentage = (voteCount: number) => {
  if (totalVotes.value === 0) return 0
  return Math.round((voteCount / totalVotes.value) * 100)
}

// Avatar Colors (Sanzo Wada Inspired)
const avatarColors = [
  '#2a4b7c', // Deep classic blue
  '#e2725b', // Terracotta
  '#5b7c7c', // Muted teal
  '#d4a373', // Warm sand
  '#8c3a3a', // Deep rust
  '#4a6b5d', // Sage green
  '#c77d54', // Clay
  '#6b5b95'  // Muted purple
]

const getAvatarColor = (id: string) => {
  if (!id) return avatarColors[0]
  let hash = 0
  for (let i = 0; i < id.length; i++) hash = id.charCodeAt(i) + ((hash << 5) - hash)
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

const timeAgo = (dateStr: string) => {
  if (!dateStr) return ''
  const parseCreatedAtMs = (value: string) => {
    const now = Date.now()
    const primary = new Date(value).getTime()
    if (!Number.isNaN(primary)) {
      // If UTC parse points to the future, try interpreting same value as local time.
      // This handles backend timestamps that include "Z" while actually representing local server time.
      if (primary > now && value.endsWith('Z')) {
        const withoutZ = value.slice(0, -1)
        const localFallback = new Date(withoutZ).getTime()
        if (!Number.isNaN(localFallback) && localFallback <= now) return localFallback
      }
      return primary
    }

    // Fallback for non-RFC strings like "2026-03-24 17:22:48.989507"
    const match = value.match(
      /^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?$/
    )
    if (match) {
      const [, y, mo, d, h, mi, s, frac = '0'] = match
      const ms = Number(frac.padEnd(3, '0').slice(0, 3))
      return new Date(
        Number(y),
        Number(mo) - 1,
        Number(d),
        Number(h),
        Number(mi),
        Number(s),
        ms
      ).getTime()
    }

    return now
  }

  const createdAtMs = parseCreatedAtMs(dateStr)
  const rawDiff = Math.floor((Date.now() - createdAtMs) / 1000)
  const diff = Math.max(0, rawDiff)
  if (diff < 60) return 'şimdi'
  if (diff < 3600) return `${Math.floor(diff / 60)}dk`
  if (diff < 86400) return `${Math.floor(diff / 3600)}sa`
  return `${Math.floor(diff / 86400)}g`
}

// Determines which option dominates (for highlighting)
const maxVoteCount = computed(() => {
  if (!q.value || !q.value.options) return 0
  return Math.max(...q.value.options.map((o: any) => o.vote_count || 0))
})
</script>

<template>
  <article class="s-card">
    <div class="s-header">
      <div class="s-avatar" :style="{ background: getAvatarColor(survey.id) }">
        {{ q?.text?.charAt(0).toUpperCase() || '?' }}
      </div>
      <div class="s-meta">
        <div class="s-author-row">
          <span class="s-author-name">Optiyoo</span>
          <span class="s-badge" v-if="!isCompleted">Yeni</span>
          <span class="s-time">{{ timeAgo(survey.created_at) }}</span>
        </div>
        <div class="s-handle">@optiyoo</div>
      </div>
      <button class="s-share-btn" :class="{ 's-share-copied': copiedLink }" @click="handleShare" :title="copiedLink ? 'Kopyalandı!' : 'Paylaş'">
        <svg v-if="!copiedLink" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="18" cy="5" r="3"></circle>
          <circle cx="6" cy="12" r="3"></circle>
          <circle cx="18" cy="19" r="3"></circle>
          <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
          <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
        </svg>
        <span v-else class="s-icon-check">✓</span>
      </button>
    </div>

    <div class="s-body">
      <h3 class="s-question-text">{{ q?.text || 'Soru bulunamadı' }}</h3>

      <div v-if="q && q.type === 'single_choice'" class="s-poll-options" :class="{ 's-poll-submitting': isSubmitting }">
        <div v-if="isSubmitting" class="s-submit-overlay">
          <div class="s-spinner"></div>
          <span>Oyunuz iletiliyor...</span>
        </div>

        <button 
          v-for="opt in q.options" :key="opt.id"
          class="s-option-btn"
          :class="{ 
            'is-voted': isCompleted,
            'is-selected': givenAnswer === opt.id,
            'is-winner': isCompleted && opt.vote_count === maxVoteCount && opt.vote_count > 0
          }"
          @click="handleVote(opt.id)"
          :disabled="isCompleted || isSubmitting"
        >
          <!-- Background Fill for Results -->
          <div 
            v-if="isCompleted" 
            class="s-option-fill" 
            :style="{ width: `${getPercentage(opt.vote_count || 0)}%` }"
            :class="{ 'fill-selected': givenAnswer === opt.id, 'fill-winner': opt.vote_count === maxVoteCount && opt.vote_count > 0 }"
          ></div>

          <div class="s-option-content" :class="{ 'is-user-choice': givenAnswer === opt.id }">
            <span class="s-option-text">
              {{ opt.text }}
              <span v-if="givenAnswer === opt.id" class="s-user-badge">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                Senin Yanıtın
              </span>
            </span>
            <span v-if="isCompleted" class="s-option-percent">{{ getPercentage(opt.vote_count || 0) }}%</span>
          </div>
        </button>

        <div v-if="isCompleted" class="s-poll-footer">
          <span>{{ totalVotes }} oy</span>
        </div>
      </div>
    </div>
  </article>
</template>

<style scoped>
.s-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03), 0 1px 3px rgba(0, 0, 0, 0.02);
  border: 1px solid #f0f0f5;
  transition: transform 0.2s cubic-bezier(0.2, 0.8, 0.2, 1), box-shadow 0.2s;
  position: relative;
  overflow: hidden;
}

.s-card:hover {
  box-shadow: 0 12px 30px rgba(0, 0, 0, 0.06), 0 4px 10px rgba(0, 0, 0, 0.03);
  transform: translateY(-2px);
}

/* Header */
.s-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.s-avatar {
  width: 44px;
  height: 44px;
  border-radius: 12px; /* modern rounded rectangle */
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 18px;
  flex-shrink: 0;
  box-shadow: inset 0 -2px 0 rgba(0,0,0,0.1);
}

.s-meta {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.s-author-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.s-author-name {
  font-weight: 700;
  font-size: 16px;
  color: #1a1a24;
}

.s-badge {
  background: linear-gradient(135deg, var(--primary-color) 0%, #8b5cf6 100%);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.s-time {
  font-size: 13px;
  color: #8f93a3;
}

.s-handle {
  font-size: 14px;
  color: #8f93a3;
  margin-top: 2px;
}

.s-share-btn {
  background: #f7f9fa;
  border: none;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.s-share-btn:hover {
  background: #eef2f6;
  color: var(--primary-color);
  transform: scale(1.05);
}

.s-share-copied {
  color: #10b981;
  background: #d1fae5;
}
.s-icon-check {
  font-weight: 800;
  font-size: 14px;
}

/* Body */
.s-question-text {
  font-size: 19px;
  font-weight: 700;
  line-height: 1.4;
  color: #111827;
  margin: 0 0 20px 0;
  letter-spacing: -0.01em;
}

/* Options */
.s-poll-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
  position: relative;
}

.s-poll-submitting {
  opacity: 0.8;
  pointer-events: none;
}

.s-submit-overlay {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: rgba(255,255,255,0.6);
  backdrop-filter: blur(2px);
  border-radius: 12px;
  color: var(--primary-color);
  font-weight: 600;
  font-size: 14px;
  gap: 8px;
}

.s-spinner {
  width: 24px;
  height: 24px;
  border: 3px solid rgba(0,0,0,0.1);
  border-top-color: var(--primary-color);
  border-radius: 50%;
  animation: s-spin 1s linear infinite;
}

@keyframes s-spin {
  to { transform: rotate(360deg); }
}

.s-option-btn {
  position: relative;
  width: 100%;
  text-align: left;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 0;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.2s cubic-bezier(0.2, 0.8, 0.2, 1);
  display: block;
}

.s-option-btn:not(.is-voted):hover {
  background: #f3f4f6;
  border-color: #d1d5db;
  transform: translateY(-1px);
}

.s-option-btn:not(.is-voted):active {
  transform: translateY(1px);
}

.s-option-content {
  position: relative;
  z-index: 2;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
}

.s-option-text {
  font-size: 15px;
  font-weight: 600;
  color: #374151;
  transition: color 0.2s;
}

.s-option-percent {
  font-size: 16px;
  font-weight: 800;
  color: #374151;
}

/* Voted States */
.s-option-btn.is-voted {
  cursor: default;
  border-color: #f3f4f6;
  background: #fdfdfd;
}

.s-option-fill {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background: #e5e7eb;
  border-radius: 12px;
  transition: width 1s cubic-bezier(0.2, 0.8, 0.2, 1);
  z-index: 1;
}

/* Active/User selection */
.s-option-btn.is-selected {
  border: 2px solid var(--primary-color);
  background: #fff;
}

.s-option-fill.fill-selected {
  background: var(--primary-color);
  opacity: 0.12; /* Light prominent background so text remains perfectly readable */
}

/* Winner styling (the option with the most votes) */
.s-option-btn.is-winner .s-option-text,
.s-option-btn.is-winner .s-option-percent {
  color: #111827;
}

/* Selected styling (the option the user voted for) */
.s-option-btn.is-selected .s-option-text,
.s-option-btn.is-selected .s-option-percent {
  color: var(--primary-color);
  font-weight: 800;
}

.s-user-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 8px;
  font-size: 12px;
  background: var(--primary-color);
  color: #fff;
  padding: 3px 8px;
  border-radius: 9999px;
  font-weight: 700;
  vertical-align: middle;
  letter-spacing: 0.5px;
}

/* If the user's selected choice is also the winner, we can make it pop more */
.s-option-fill.fill-selected.fill-winner {
  background: var(--primary-color);
  opacity: 0.2;
}

.s-poll-footer {
  margin-top: 12px;
  font-size: 14px;
  color: #6b7280;
  text-align: right;
  font-weight: 500;
}
</style>
