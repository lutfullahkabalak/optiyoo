<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const survey = ref<any>(null)
const isLoading = ref(true)
const isSubmitting = ref(false)
const isCompleted = ref(false)
const givenAnswer = ref<string | null>(null)
const copiedLink = ref(false)

const surveyId = route.params.id as string

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/auth')
    return
  }

  const completed = new Set<string>(
    JSON.parse(localStorage.getItem('completed_polls_' + authStore.user?.id) || '[]')
  )
  isCompleted.value = completed.has(surveyId)

  try {
    const res = await fetch(
      `http://localhost:8080/api/surveys/${surveyId}?user_id=${authStore.user?.id}`
    )
    if (res.ok) {
      survey.value = await res.json()
      if (survey.value.user_answer) {
        givenAnswer.value = survey.value.user_answer
        isCompleted.value = true
      }
    }
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
})

const getTotalVotes = (q: any) => {
  if (!q || !q.options) return 0
  return q.options.reduce((sum: number, o: any) => sum + (o.vote_count || 0), 0)
}

const submitVote = async (optionId: string) => {
  if (isCompleted.value || isSubmitting.value || !survey.value) return

  isSubmitting.value = true
  const payload = {
    user_id: authStore.user?.id,
    answers: [{ question_id: survey.value.questions[0].id, value: optionId }]
  }

  try {
    const res = await fetch(`http://localhost:8080/api/surveys/${surveyId}/answers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })

    if (res.ok) {
      givenAnswer.value = optionId
      isCompleted.value = true

      const completed = new Set<string>(
        JSON.parse(localStorage.getItem('completed_polls_' + authStore.user?.id) || '[]')
      )
      completed.add(surveyId)
      localStorage.setItem(
        'completed_polls_' + authStore.user?.id,
        JSON.stringify(Array.from(completed))
      )

      // Yerel oy sayısını güncelle
      const q = survey.value.questions[0]
      const opt = q.options.find((o: any) => o.id === optionId)
      if (opt) opt.vote_count = (opt.vote_count || 0) + 1
      survey.value.questions[0] = { ...q }
    } else if (res.status === 403) {
      isCompleted.value = true
      const completed = new Set<string>(
        JSON.parse(localStorage.getItem('completed_polls_' + authStore.user?.id) || '[]')
      )
      completed.add(surveyId)
      localStorage.setItem(
        'completed_polls_' + authStore.user?.id,
        JSON.stringify(Array.from(completed))
      )
    }
  } catch (e) {
    console.error(e)
  } finally {
    isSubmitting.value = false
  }
}

const copyLink = async () => {
  await navigator.clipboard.writeText(`${window.location.origin}/s/${surveyId}`)
  copiedLink.value = true
  setTimeout(() => (copiedLink.value = false), 1500)
}

const getCreatorName = (s: any) => s?.creator_name || 'Optiyoo'

const getCreatorHandle = (s: any) => {
  const name = getCreatorName(s)
  return `@${name.toLowerCase().replace(/\s+/g, '')}`
}

const timeAgo = (dateStr: string) => {
  if (!dateStr) return ''
  const now = Date.now()
  const primary = new Date(dateStr).getTime()
  const createdAt = Number.isNaN(primary) ? now : primary
  const diff = Math.max(0, Math.floor((now - createdAt) / 1000))
  if (diff < 60) return 'şimdi'
  if (diff < 3600) return `${Math.floor(diff / 60)}dk`
  if (diff < 86400) return `${Math.floor(diff / 3600)}sa`
  return `${Math.floor(diff / 86400)}g`
}

const avatarColors = [
  '#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#ef4444', '#14b8a6'
]

const getAvatarColor = (id: string) => {
  let hash = 0
  for (let i = 0; i < id.length; i++) hash = id.charCodeAt(i) + ((hash << 5) - hash)
  return avatarColors[Math.abs(hash) % avatarColors.length]
}
</script>

<template>
  <div class="app-wrapper animate-fade-in">
    <header class="header">
      <div class="container flex justify-between items-center header-inner">
        <div class="logo">optiyoo</div>
        <nav class="nav">
          <router-link to="/" style="text-decoration:none;">
            <button class="btn btn-outline" style="font-size:13px;">← Geri</button>
          </router-link>
        </nav>
      </div>
    </header>

    <main class="container main-content mt-8" style="max-width: 600px; width: 100%;">
      <div v-if="isLoading" class="card text-center text-muted">Yükleniyor...</div>

      <div v-else-if="!survey" class="card text-center text-muted">
        Aradığınız anket bulunamadı.
        <br /><br />
        <router-link to="/"><button class="btn btn-outline">Panele Dön</button></router-link>
      </div>

      <div v-else class="card feed-card">
        <!-- Gönderiliyor overlay -->
        <div v-if="isSubmitting" class="completed-overlay animate-fade-in">
          <p style="font-size: 18px; font-weight: 600; color: var(--primary-color);">
            Oyunuz İletiliyor...
          </p>
        </div>

        <div class="s-header">
          <div class="s-avatar" :style="{ background: getAvatarColor(survey.id) }">
            {{ getCreatorName(survey).charAt(0).toUpperCase() }}
          </div>
          <div class="s-meta">
            <div class="s-author-row">
              <span class="s-author-name">{{ getCreatorName(survey) }}</span>
              <span class="s-badge" v-if="!isCompleted">Yeni</span>
              <span class="s-time">{{ timeAgo(survey.created_at) }}</span>
            </div>
            <div class="s-handle">{{ getCreatorHandle(survey) }}</div>
          </div>
          <span></span>
        </div>

        <h3 class="s-question-text">
          {{ survey.questions && survey.questions.length > 0 ? survey.questions[0].text : 'Soru bulunamadı' }}
        </h3>

        <div v-if="survey.questions && survey.questions.length > 0 && survey.questions[0].type === 'single_choice'" class="s-poll">
          <label
            v-for="opt in survey.questions[0].options"
            :key="opt.id"
            class="s-option-btn"
            :class="{ 'is-selected': givenAnswer === opt.id, 'is-disabled': isCompleted }"
          >
            <div
              class="s-option-fill"
              :class="{ 'is-visible': isCompleted }"
              :style="`width: ${isCompleted && getTotalVotes(survey.questions[0]) > 0 ? Math.round((opt.vote_count / getTotalVotes(survey.questions[0])) * 100) : 0}%`"
            ></div>
            <div class="s-option-content">
              <input
                type="radio"
                :name="'survey_' + survey.id"
                :value="opt.id"
                v-model="givenAnswer"
                @change="submitVote(opt.id)"
                :disabled="isCompleted"
                style="display: none;"
              />
              <div class="s-radio" :class="{ checked: givenAnswer === opt.id }"></div>
              <span class="s-option-text">{{ opt.text }}</span>
              <span v-if="isCompleted" class="s-option-percent">
                {{ getTotalVotes(survey.questions[0]) > 0 ? Math.round((opt.vote_count / getTotalVotes(survey.questions[0])) * 100) : 0 }}%
              </span>
            </div>
          </label>

          <div class="s-actions">
            <p v-if="isCompleted" class="s-total-votes">Toplam {{ getTotalVotes(survey.questions[0]) }} Oy</p>
            <button
              class="s-share-btn"
              @click="copyLink"
              :title="copiedLink ? 'Kopyalandı!' : 'Linki Kopyala'"
              :class="{ copied: copiedLink }"
            >
              <svg
                v-if="!copiedLink"
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
                <polyline points="16 6 12 2 8 6" />
                <line x1="12" y1="2" x2="12" y2="15" />
              </svg>
              <span v-else style="font-size: 13px; font-weight: 700;">✓</span>
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.header {
  background-color: var(--color-white);
  border-bottom: 1px solid var(--border-color);
  padding: var(--spacing-4) 0;
}
.logo {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.05em;
}

.feed-card {
  border: 1px solid #f0f0f5;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03), 0 1px 3px rgba(0, 0, 0, 0.02);
  margin-bottom: var(--spacing-6);
  position: relative;
  overflow: hidden;
  padding: 24px;
}
.s-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 14px;
}
.s-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 700;
  font-size: 16px;
  flex-shrink: 0;
}
.s-meta {
  flex: 1;
}
.s-author-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.s-author-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-color);
}
.s-badge {
  background: var(--primary-color);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
}
.s-time,
.s-handle {
  font-size: 13px;
  color: var(--text-color-muted);
}
.s-share-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  border: none;
  background: #f7f9fa;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-color-muted);
  transition: all 0.2s;
}
.s-share-btn:hover {
  color: var(--primary-color);
}
.s-share-btn.copied {
  color: var(--secondary-color);
}
.s-question-text {
  margin: 0 0 16px 0;
  color: var(--text-color);
}
.s-poll {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.s-option-btn {
  position: relative;
  padding: 0;
  border: 2px solid var(--border-color);
  border-radius: 9999px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.2s ease;
}
.s-option-btn:not(.is-disabled):hover {
  border-color: var(--primary-color);
}
.s-option-btn.is-selected {
  border-color: var(--primary-color);
}
.s-option-btn.is-disabled {
  cursor: default;
}
.s-option-content {
  position: relative;
  z-index: 2;
  padding: 10px 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
}
.s-option-fill {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background: #c8c4bd;
  opacity: 0;
  z-index: 1;
  transition: width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.s-option-fill.is-visible {
  opacity: 0.92;
}
.s-radio {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid var(--border-color);
  margin-right: 12px;
  transition: all 0.2s;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.s-radio.checked {
  border-color: var(--primary-color);
}
.s-radio.checked::after {
  content: '';
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background-color: var(--primary-color);
}
.s-option-text {
  font-weight: 500;
  flex: 1;
}
.s-option-percent {
  font-weight: 700;
  font-size: 14px;
}
.s-total-votes {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-color-muted);
  margin: 0;
}
.s-actions {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.s-actions .s-share-btn {
  margin-left: auto;
}

.completed-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 10;
  backdrop-filter: blur(2px);
  background: rgba(255, 255, 255, 0.7);
}
</style>
