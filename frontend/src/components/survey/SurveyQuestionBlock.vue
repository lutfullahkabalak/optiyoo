<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    surveyId: string
    question: any | null
    questionCount: number
    questionIndex: number
    canGoPrev: boolean
    canGoNext: boolean
    selectedAnswer: string
    /** Tüm anket gönderildi */
    isCompleted: boolean
    /** Bu soruda seçim yapıldı; gönderim beklerken değiştirilemez */
    isAnswerLocked?: boolean
    getPercent: (opt: any) => number
    /** Mutlak görsel URL’leri için API kökü */
    apiBase?: string
  }>(),
  { apiBase: 'http://localhost:8080' }
)

const mediaSrc = (path: string | undefined) => {
  if (!path) return ''
  if (path.startsWith('http://') || path.startsWith('https://')) return path
  const base = props.apiBase.replace(/\/$/, '')
  return base + (path.startsWith('/') ? path : '/' + path)
}

const questionImageSrc = computed(() => mediaSrc(props.question?.image_url))

/** Eski kayıtlar `choice`; yeni oluşturma `single_choice` — ana sayfa akışında ikisi de poll sayılır. */
const isSingleChoicePoll = computed(() => {
  const t = props.question?.type
  return t === 'single_choice' || t === 'choice'
})

const pollFrozen = () => props.isCompleted || Boolean(props.isAnswerLocked)

const emit = defineEmits<{
  prev: []
  next: []
  select: [payload: { questionId: string; optionId: string }]
}>()

const onSelect = (optionId: string) => {
  if (!props.question?.id || pollFrozen()) return
  emit('select', { questionId: props.question.id, optionId })
}

/** Tarayıcı radio grupları tüm dokümanda tek seçim zorlar; ayrıca id string/number farkı border'ı kırabilir. */
const isOptSelected = (optId: string | undefined) =>
  optId != null &&
  props.selectedAnswer !== '' &&
  String(props.selectedAnswer) === String(optId)

const onRowClick = (optId: string | undefined) => {
  if (optId == null || pollFrozen()) return
  onSelect(optId)
}
</script>

<template>
  <div class="question-shell">
    <button
      v-if="questionCount > 1"
      class="question-side-btn question-side-left"
      type="button"
      :disabled="!canGoPrev"
      @click="emit('prev')"
    >
      ←
    </button>
    <button
      v-if="questionCount > 1"
      class="question-side-btn question-side-right"
      type="button"
      :disabled="!canGoNext"
      @click="emit('next')"
    >
      →
    </button>

    <Transition name="question-slide" mode="out-in">
      <div :key="`${surveyId}:${question?.id || 'empty'}`" class="question-body">
        <img
          v-if="questionImageSrc"
          :src="questionImageSrc"
          alt=""
          class="question-banner-img"
          loading="lazy"
          decoding="async"
        />
        <p class="question-title">{{ question?.text || 'Soru bulunamadı' }}</p>

        <div
          v-if="question && isSingleChoicePoll"
          class="question-poll"
          role="radiogroup"
          :aria-label="question.text"
        >
          <div
            v-for="(opt, oIdx) in question.options"
            :key="opt.id ?? 'o' + oIdx"
            class="question-poll-opt"
            role="radio"
            :aria-checked="isOptSelected(opt.id)"
            :tabindex="pollFrozen() ? -1 : 0"
            :class="{
              'is-selected': isOptSelected(opt.id),
              'is-disabled': pollFrozen()
            }"
            @click="onRowClick(opt.id)"
            @keydown.enter.prevent="onRowClick(opt.id)"
            @keydown.space.prevent="onRowClick(opt.id)"
          >
            <div
              class="question-progress-fill"
              :class="{ 'is-visible': pollFrozen(), 'is-selected-fill': isOptSelected(opt.id) }"
              :style="{ width: `${pollFrozen() ? getPercent(opt) : 0}%` }"
            ></div>
            <div class="question-opt-content">
              <div class="question-radio" :class="{ 'is-checked': isOptSelected(opt.id) }"></div>
              <img
                v-if="opt.image_url"
                :src="mediaSrc(opt.image_url)"
                alt=""
                class="question-opt-thumb"
                loading="lazy"
                decoding="async"
              />
              <span class="question-opt-text">{{ opt.text }}</span>
              <span class="question-opt-pct" :class="{ 'is-visible': pollFrozen() }">
                {{ getPercent(opt) }}%
              </span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.question-shell {
  position: relative;
  margin-top: 10px;
  padding: 0;
}
.question-side-btn {
  position: absolute;
  top: 50%;
  width: 32px;
  height: 32px;
  border: 1px solid var(--border-color);
  border-radius: 9999px;
  background: var(--color-white);
  color: var(--text-color);
  cursor: pointer;
  z-index: 5;
  -webkit-tap-highlight-color: transparent;
}
.question-side-left {
  left: 0;
  transform: translate(-120%, -50%);
}
.question-side-right {
  right: 0;
  transform: translate(120%, -50%);
}
.question-side-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}
.question-slide-enter-active,
.question-slide-leave-active {
  transition: all 0.24s ease;
}
.question-slide-enter-from {
  opacity: 0;
  transform: translateX(14px);
}
.question-slide-leave-to {
  opacity: 0;
  transform: translateX(-14px);
}
.question-body {
  padding-left: 20px;
  padding-right: 4px;
  min-width: 0;
}
.question-banner-img {
  display: block;
  width: 100%;
  max-width: 100%;
  max-height: min(36vh, 220px);
  height: auto;
  object-fit: contain;
  object-position: center;
  border-radius: 14px;
  margin: 0 0 12px 0;
  border: 1px solid var(--border-color);
  box-sizing: border-box;
  background: var(--bg-color);
}
.question-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
  margin-left: 0;
  margin-bottom: 2px;
}
.question-opt-thumb {
  width: 40px;
  height: 40px;
  object-fit: contain;
  object-position: center;
  border-radius: 8px;
  flex-shrink: 0;
  border: 1px solid var(--border-color);
  background: var(--bg-color);
}
.question-poll {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 12px;
}
.question-poll-opt {
  position: relative;
  border: 2px solid var(--border-color);
  border-radius: 9999px;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s;
}
.question-poll-opt:not(.is-disabled):hover {
  border-color: var(--primary-color);
}
.question-poll-opt.is-selected {
  border-color: var(--primary-color);
}
.question-poll-opt.is-disabled {
  cursor: default;
}
.question-poll-opt:focus {
  outline: none;
}
.question-poll-opt:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}
.question-progress-fill {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background: #c8c4bd;
  opacity: 0;
  border-radius: 9999px;
  transition: width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
  z-index: 0;
}
.question-progress-fill.is-visible {
  opacity: 0.92;
}
.question-opt-content {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  padding: 10px 18px;
  gap: 12px;
}
.question-radio {
  width: 18px;
  height: 18px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  background: #fff;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.question-radio.is-checked {
  border-color: var(--primary-color);
}
.question-radio.is-checked::after {
  content: '';
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--primary-color);
}
.question-opt-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  flex: 1;
}
.question-opt-pct {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-color);
  min-width: 44px;
  text-align: right;
  opacity: 0;
  transform: translateX(4px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.question-opt-pct.is-visible {
  opacity: 1;
  transform: translateX(0);
}

/* Ok düğmeleri dar ekranda taşmayı önle: içeride hizala */
@media (max-width: 640px) {
  .question-shell {
    margin-top: 8px;
    padding-left: 40px;
    padding-right: 40px;
  }
  .question-side-left {
    left: 2px;
    transform: translateY(-50%);
  }
  .question-side-right {
    right: 2px;
    transform: translateY(-50%);
  }
  .question-body {
    padding-left: 4px;
    padding-right: 4px;
  }
  .question-title {
    font-size: 0.9375rem;
  }
  .question-opt-content {
    padding: 10px 12px;
    gap: 8px;
  }
  .question-opt-text {
    font-size: 13px;
  }
  .question-opt-pct {
    min-width: 36px;
    font-size: 13px;
  }
  .question-poll-opt {
    border-radius: 14px;
  }
  .question-progress-fill {
    border-radius: 14px;
  }
}
</style>
