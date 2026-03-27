<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import SurveyUserHeader from '../components/survey/SurveyUserHeader.vue'
import SurveyQuestionBlock from '../components/survey/SurveyQuestionBlock.vue'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const survey = ref<any>(null)
const isLoading = ref(true)
const isSubmitting = ref(false)
const isCompleted = ref(false)
const givenAnswers = ref<Record<string, string>>({})
const lockedQuestionIds = ref<Set<string>>(new Set())
const copiedLink = ref(false)

const surveyId = route.params.id as string

const isSurveyFullyAnswered = (s: any, answers: Record<string, string>) => {
  const questions = s?.questions || []
  if (questions.length === 0) return false
  return questions.every((q: any) => q.id != null && Boolean(answers[String(q.id)]))
}

const syncCompletedLocalStorage = (full: boolean) => {
  const uid = authStore.user?.id
  if (!uid) return
  const completed = new Set<string>(
    JSON.parse(localStorage.getItem('completed_polls_' + uid) || '[]')
  )
  if (full) completed.add(surveyId)
  else completed.delete(surveyId)
  localStorage.setItem('completed_polls_' + uid, JSON.stringify(Array.from(completed)))
}

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/auth')
    return
  }

  try {
    const res = await fetch(
      `${API_BASE}/api/surveys/${surveyId}?user_id=${authStore.user?.id}`
    )
    if (res.ok) {
      survey.value = await res.json()
      const uaRaw = survey.value.user_answers as Record<string, string> | undefined
      const ua =
        uaRaw && typeof uaRaw === 'object'
          ? Object.fromEntries(
              Object.entries(uaRaw).map(([k, v]) => [String(k), String(v)])
            )
          : undefined
      if (ua && Object.keys(ua).length > 0) {
        givenAnswers.value = { ...givenAnswers.value, ...ua }
        lockedQuestionIds.value = new Set([...lockedQuestionIds.value, ...Object.keys(ua)])
        isCompleted.value = isSurveyFullyAnswered(survey.value, givenAnswers.value)
        syncCompletedLocalStorage(isCompleted.value)
      } else if (survey.value.user_answer) {
        const firstQuestionId = survey.value?.questions?.[0]?.id
        if (firstQuestionId) {
          const fid = String(firstQuestionId)
          givenAnswers.value = {
            ...givenAnswers.value,
            [fid]: String(survey.value.user_answer)
          }
          lockedQuestionIds.value = new Set(lockedQuestionIds.value).add(fid)
        }
        isCompleted.value = isSurveyFullyAnswered(survey.value, givenAnswers.value)
        syncCompletedLocalStorage(isCompleted.value)
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

const revertQuestionSelection = (questionId: string) => {
  const qid = String(questionId)
  const next = new Set(lockedQuestionIds.value)
  next.delete(qid)
  lockedQuestionIds.value = next
  const ga = { ...givenAnswers.value }
  delete ga[qid]
  givenAnswers.value = ga
}

const handleQuestionSelect = async (payload: { questionId: string; optionId: string }) => {
  if (isCompleted.value || !survey.value) return
  const qid = String(payload.questionId)
  if (lockedQuestionIds.value.has(qid)) return
  givenAnswers.value = { ...givenAnswers.value, [qid]: String(payload.optionId) }
  lockedQuestionIds.value = new Set(lockedQuestionIds.value).add(qid)
  await submitSingleAnswer(qid, String(payload.optionId))
}

const submitSingleAnswer = async (questionId: string, optionId: string) => {
  if (isSubmitting.value || !survey.value) return

  isSubmitting.value = true
  const payload = {
    user_id: authStore.user?.id,
    answers: [{ question_id: questionId, value: optionId }]
  }

  try {
    const res = await fetch(`${API_BASE}/api/surveys/${surveyId}/answers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })

    if (res.ok) {
      const q = survey.value.questions?.find((x: any) => String(x.id) === questionId)
      const opt = q?.options?.find((o: any) => String(o.id) === optionId)
      if (opt) opt.vote_count = (opt.vote_count || 0) + 1

      if (isSurveyFullyAnswered(survey.value, givenAnswers.value)) {
        isCompleted.value = true
        syncCompletedLocalStorage(true)
      }
    } else if (res.status === 403) {
      const r = await fetch(
        `${API_BASE}/api/surveys/${surveyId}?user_id=${authStore.user?.id}`
      )
      if (r.ok) {
        survey.value = await r.json()
        const uaRaw = survey.value.user_answers as Record<string, string> | undefined
        if (uaRaw && typeof uaRaw === 'object') {
          const ua = Object.fromEntries(
            Object.entries(uaRaw).map(([k, v]) => [String(k), String(v)])
          )
          givenAnswers.value = { ...givenAnswers.value, ...ua }
          lockedQuestionIds.value = new Set([...Object.keys(ua)])
        }
        isCompleted.value = isSurveyFullyAnswered(survey.value, givenAnswers.value)
        syncCompletedLocalStorage(isCompleted.value)
      }
    } else {
      const text = await res.text()
      revertQuestionSelection(questionId)
      alert('Oylama İptali: ' + text)
    }
  } catch (e) {
    console.error(e)
    revertQuestionSelection(questionId)
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
  const h = String(s?.creator_username || '').trim()
  if (h) return `@${h}`
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
        <router-link to="/" class="logo-link">
          <span class="logo">optiyoo</span>
        </router-link>
        <nav class="nav">
          <router-link to="/" class="back-btn-wrap">
            <button type="button" class="btn btn-outline back-btn">← Geri</button>
          </router-link>
        </nav>
      </div>
    </header>

    <main class="container main-content survey-main">
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
          <SurveyUserHeader
            :avatar-color="getAvatarColor(survey.id)"
            :avatar-text="getCreatorName(survey).charAt(0).toUpperCase()"
            :display-name="getCreatorName(survey)"
            :handle="getCreatorHandle(survey)"
            :time-ago="timeAgo(survey.created_at)"
            :is-new="!isCompleted"
          />
        </div>

        <div
          v-for="(q, qIdx) in survey.questions || []"
          :key="q.id"
          class="survey-detail-question"
        >
          <SurveyQuestionBlock
            :survey-id="survey.id"
            :question="q"
            :question-count="1"
            :question-index="Number(qIdx) + 1"
            :can-go-prev="false"
            :can-go-next="false"
            :selected-answer="(q.id != null ? givenAnswers[String(q.id)] : '') || ''"
            :is-completed="isCompleted"
            :is-answer-locked="Boolean(q.id != null && lockedQuestionIds.has(String(q.id)))"
            :get-percent="(opt: any) => getTotalVotes(q) > 0 ? Math.round((opt.vote_count / getTotalVotes(q)) * 100) : 0"
            :api-base="API_BASE"
            @select="handleQuestionSelect"
          />

          <div
            class="s-actions"
            v-if="q.type === 'single_choice' && (isCompleted || qIdx === (survey.questions?.length || 0) - 1)"
          >
            <p
              v-if="isCompleted || (q.id != null && lockedQuestionIds.has(String(q.id)))"
              class="s-total-votes"
            >
              Toplam {{ getTotalVotes(q) }} Oy
            </p>
            <button
              v-if="qIdx === (survey.questions?.length || 0) - 1"
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
  padding-top: max(var(--spacing-4), env(safe-area-inset-top, 0px));
}
.header-inner {
  gap: var(--spacing-3);
  flex-wrap: wrap;
}
.logo-link {
  text-decoration: none;
  min-width: 0;
}
.logo {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.05em;
}
.back-btn-wrap {
  text-decoration: none;
  flex-shrink: 0;
}
.back-btn {
  font-size: 13px;
  white-space: nowrap;
}

.survey-main {
  max-width: 600px;
  width: 100%;
  margin-left: auto;
  margin-right: auto;
  margin-top: var(--spacing-8);
  padding-bottom: max(var(--spacing-8), env(safe-area-inset-bottom, 0px));
}

.feed-card {
  border: 1px solid #f0f0f5;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03), 0 1px 3px rgba(0, 0, 0, 0.02);
  margin-bottom: var(--spacing-6);
  position: relative;
  overflow: hidden;
  padding: 24px;
}

@media (max-width: 480px) {
  .logo {
    font-size: var(--font-size-xl);
  }
  .feed-card {
    padding: 16px 14px;
    margin-bottom: var(--spacing-4);
  }
  .survey-main {
    margin-top: var(--spacing-4);
  }
}
.s-header {
  margin-bottom: 14px;
}
.survey-detail-question + .survey-detail-question {
  margin-top: 28px;
  padding-top: 24px;
  border-top: 1px solid var(--border-color);
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
