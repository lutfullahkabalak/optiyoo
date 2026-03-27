<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CreateSurveyModal from '../components/CreateSurveyModal.vue'
import SurveyUserHeader from '../components/survey/SurveyUserHeader.vue'
import SurveyQuestionBlock from '../components/survey/SurveyQuestionBlock.vue'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const router = useRouter()
const authStore = useAuthStore()

const surveys = ref<any[]>([])
const isLoading = ref(true)
const showCreateModal = ref(false)

const givenAnswers = ref<Record<string, string>>({})
const copiedSurveyId = ref<string | null>(null)
const animatedPercents = ref<Record<string, number>>({})
const openMenuSurveyId = ref<string | null>(null)
const activeQuestionIndexes = ref<Record<string, number>>({})

const copyLink = async (surveyId: string) => {
  const link = `${window.location.origin}/s/${surveyId}`
  await navigator.clipboard.writeText(link)
  copiedSurveyId.value = surveyId
  setTimeout(() => (copiedSurveyId.value = null), 1500)
}

const toggleSurveyMenu = (surveyId: string) => {
  openMenuSurveyId.value = openMenuSurveyId.value === surveyId ? null : surveyId
}

const closeSurveyMenu = () => {
  openMenuSurveyId.value = null
}

const openSurveyDetail = (surveyId: string) => {
  closeSurveyMenu()
  router.push(`/s/${surveyId}`)
}
const submittingState = ref<Record<string, boolean>>({})
const completedSurveys = ref<Set<string>>(new Set(JSON.parse(localStorage.getItem('completed_polls_' + authStore.user?.id) || '[]')))
/** Soru seçildikten sonra (çoklu ankette) o soru tekrar değiştirilemez */
const lockedQuestionIdsBySurvey = ref<Record<string, Record<string, true>>>({})

const isSurveyFullyAnswered = (survey: any, answers: Record<string, string>) => {
  const questions = survey?.questions || []
  if (questions.length === 0) return false
  return questions.every((q: any) => q.id != null && Boolean(answers[String(q.id)]))
}

const persistCompletedSurveys = () => {
  const uid = authStore.user?.id
  if (!uid) return
  localStorage.setItem('completed_polls_' + uid, JSON.stringify(Array.from(completedSurveys.value)))
}

const fetchSurveys = async (opts?: { silent?: boolean }) => {
  if (!opts?.silent) isLoading.value = true
  try {
    const res = await fetch(`${API_BASE}/api/surveys?user_id=${authStore.user?.id}`)
    if (res.ok) {
      surveys.value = await res.json()
      surveys.value.forEach((s: any) => {
        if (activeQuestionIndexes.value[s.id] === undefined) {
          activeQuestionIndexes.value[s.id] = 0
        }
        const ua = s.user_answers as Record<string, string> | undefined
        if (ua && typeof ua === 'object' && Object.keys(ua).length > 0) {
          const prevLocks = lockedQuestionIdsBySurvey.value[s.id] || {}
          const merged: Record<string, true> = { ...prevLocks }
          const nextAnswers = { ...givenAnswers.value }
          for (const [qid, optId] of Object.entries(ua)) {
            const k = String(qid)
            nextAnswers[k] = String(optId)
            merged[k] = true
          }
          givenAnswers.value = nextAnswers
          lockedQuestionIdsBySurvey.value = {
            ...lockedQuestionIdsBySurvey.value,
            [s.id]: merged
          }
          if (isSurveyFullyAnswered(s, nextAnswers)) {
            completedSurveys.value.add(s.id)
          } else {
            completedSurveys.value.delete(s.id)
          }
        } else if (s.user_answer) {
          const firstQuestionId = s?.questions?.[0]?.id
          if (firstQuestionId) {
            const fid = String(firstQuestionId)
            const nextAnswers = { ...givenAnswers.value, [fid]: String(s.user_answer) }
            givenAnswers.value = nextAnswers
            const prevLocks = lockedQuestionIdsBySurvey.value[s.id] || {}
            lockedQuestionIdsBySurvey.value = {
              ...lockedQuestionIdsBySurvey.value,
              [s.id]: { ...prevLocks, [fid]: true }
            }
            if (isSurveyFullyAnswered(s, nextAnswers)) {
              completedSurveys.value.add(s.id)
            } else {
              completedSurveys.value.delete(s.id)
            }
          }
        }
        if (completedSurveys.value.has(s.id)) {
          setSurveyPercentsInstant(s, getCurrentQuestion(s))
        }
      })
      persistCompletedSurveys()
    }
  } catch (e) {
    console.error(e)
  } finally {
    if (!opts?.silent) isLoading.value = false
  }
}

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/auth')
    return
  }
  await fetchSurveys()
})

const handleSurveyCreated = async () => {
  showCreateModal.value = false
  await fetchSurveys()
}

const handleLogout = () => {
  authStore.logout()
  router.push('/auth')
}

const getTotalVotes = (q: any) => {
  if (!q || !q.options) return 0
  return q.options.reduce((sum: number, o: any) => sum + (o.vote_count || 0), 0)
}

const getOptionPercent = (q: any, opt: any) => {
  const totalVotes = getTotalVotes(q)
  if (totalVotes <= 0) return 0
  return Math.round(((opt.vote_count || 0) / totalVotes) * 100)
}

const getOptionAnimKey = (surveyId: string, optionId: string) => `${surveyId}:${optionId}`

const getCurrentQuestion = (survey: any) => {
  const questions = survey?.questions || []
  if (questions.length === 0) return null
  const activeIndex = activeQuestionIndexes.value[survey.id] ?? 0
  return questions[Math.min(Math.max(activeIndex, 0), questions.length - 1)]
}

const isCurrentQuestionAnswerLocked = (survey: any) => {
  const q = getCurrentQuestion(survey)
  if (q?.id == null || completedSurveys.value.has(survey.id)) return false
  return Boolean(lockedQuestionIdsBySurvey.value[survey.id]?.[String(q.id)])
}

const canGoPrevQuestion = (survey: any) => (activeQuestionIndexes.value[survey.id] ?? 0) > 0

const canGoNextQuestion = (survey: any) => {
  const questions = survey?.questions || []
  return (activeQuestionIndexes.value[survey.id] ?? 0) < questions.length - 1
}

const goToPrevQuestion = (survey: any) => {
  if (!canGoPrevQuestion(survey)) return
  activeQuestionIndexes.value[survey.id] = (activeQuestionIndexes.value[survey.id] ?? 0) - 1
}

const goToNextQuestion = (survey: any) => {
  if (!canGoNextQuestion(survey)) return
  activeQuestionIndexes.value[survey.id] = (activeQuestionIndexes.value[survey.id] ?? 0) + 1
}

const getAnimatedPercent = (survey: any, q: any, opt: any) => {
  const key = getOptionAnimKey(survey.id, opt.id)
  return animatedPercents.value[key] ?? getOptionPercent(q, opt)
}

const setSurveyPercentsInstant = (survey: any, q: any) => {
  if (!q?.options) return

  q.options.forEach((opt: any) => {
    const key = getOptionAnimKey(survey.id, opt.id)
    animatedPercents.value[key] = getOptionPercent(q, opt)
  })
}

const animateSurveyPercents = (survey: any, q: any) => {
  if (!q?.options) return

  const targets: Array<{ key: string; target: number }> = q.options.map((opt: any) => ({
    key: getOptionAnimKey(survey.id, opt.id),
    target: getOptionPercent(q, opt)
  }))

  targets.forEach(({ key }) => {
    animatedPercents.value[key] = 0
  })

  const duration = 275
  const start = performance.now()

  const tick = (now: number) => {
    const progress = Math.min((now - start) / duration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)

    targets.forEach(({ key, target }) => {
      animatedPercents.value[key] = Math.round(target * eased)
    })

    if (progress < 1) requestAnimationFrame(tick)
  }

  requestAnimationFrame(tick)
}

const revertQuestionSelection = (survey: any, questionId: string) => {
  const qid = String(questionId)
  const prevLocks = { ...(lockedQuestionIdsBySurvey.value[survey.id] || {}) }
  delete prevLocks[qid]
  lockedQuestionIdsBySurvey.value = {
    ...lockedQuestionIdsBySurvey.value,
    [survey.id]: prevLocks
  }
  const next = { ...givenAnswers.value }
  delete next[qid]
  givenAnswers.value = next
}

const handleQuestionSelect = async (survey: any, payload: { questionId: string; optionId: string }) => {
  if (completedSurveys.value.has(survey.id)) return
  const qid = String(payload.questionId)
  if (lockedQuestionIdsBySurvey.value[survey.id]?.[qid]) return
  givenAnswers.value = { ...givenAnswers.value, [qid]: String(payload.optionId) }
  const prev = lockedQuestionIdsBySurvey.value[survey.id] || {}
  lockedQuestionIdsBySurvey.value = {
    ...lockedQuestionIdsBySurvey.value,
    [survey.id]: { ...prev, [qid]: true }
  }
  await submitSingleAnswer(survey, qid, String(payload.optionId))
}

const submitSingleAnswer = async (survey: any, questionId: string, optionId: string) => {
  if (submittingState.value[survey.id]) return

  submittingState.value[survey.id] = true
  const payload = {
    user_id: authStore.user?.id,
    answers: [{ question_id: questionId, value: optionId }]
  }

  try {
    const res = await fetch(`${API_BASE}/api/surveys/${survey.id}/answers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })

    if (res.ok) {
      const answeredQ = survey.questions?.find((q: any) => String(q.id) === questionId)
      const opt = answeredQ?.options?.find((o: any) => String(o.id) === optionId)
      if (opt) opt.vote_count = (opt.vote_count || 0) + 1

      if (isSurveyFullyAnswered(survey, givenAnswers.value)) {
        completedSurveys.value.add(survey.id)
        persistCompletedSurveys()
      }

      await nextTick()
      if (answeredQ) animateSurveyPercents(survey, answeredQ)
    } else {
      const text = await res.text()
      if (res.status === 403) {
        await fetchSurveys({ silent: true })
      } else {
        revertQuestionSelection(survey, questionId)
        alert('Oylama İptali: ' + text)
      }
    }
  } catch (e) {
    console.error(e)
    revertQuestionSelection(survey, questionId)
  } finally {
    submittingState.value[survey.id] = false
  }
}

// Avatar rengi için survey id'den deterministik renk üret
const avatarColors = [
  '#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#ef4444', '#14b8a6'
]
const getAvatarColor = (id: string) => {
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

const getCreatorName = (survey: any) => survey.creator_name || 'Optiyoo'

const getCreatorHandle = (survey: any) => {
  const h = String(survey.creator_username || '').trim()
  if (h) return `@${h}`
  const name = getCreatorName(survey)
  return `@${name.toLowerCase().replace(/\s+/g, '')}`
}
</script>

<template>
  <div class="x-layout" v-if="authStore.user">

    <!-- Sol Kenar Çubuğu -->
    <aside class="x-sidebar">
      <div class="x-sidebar-inner">
        <router-link to="/" class="x-logo">optiyoo</router-link>

        <nav class="x-nav">
          <router-link to="/" class="x-nav-item">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
            <span>Ana Sayfa</span>
          </router-link>

          <button class="x-nav-item" @click="showCreateModal = true">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/>
            </svg>
            <span>Oluştur</span>
          </button>

          <router-link to="/settings" class="x-nav-item" title="Profil ayarları">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
            </svg>
            <span>Profil</span>
          </router-link>

          <button class="x-nav-item x-logout" @click="handleLogout">
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
            </svg>
            <span>Çıkış Yap</span>
          </button>
        </nav>
      </div>
    </aside>

    <!-- Ana Feed -->
    <main class="x-feed">
      <div class="x-feed-header">
        <h2>Ana Sayfa</h2>
      </div>

      <div v-if="isLoading" class="x-state-msg">Akış yükleniyor...</div>
      <div v-else-if="surveys.length === 0" class="x-state-msg">Şu an aktif bir oylama bulunmuyor.</div>

      <article v-for="s in surveys" :key="s.id" class="x-post" @click="closeSurveyMenu">
        <!-- Post İçeriği -->
        <div class="x-post-body">
          <div class="x-post-top-actions">
            <button
              class="x-more-btn"
              type="button"
              :aria-expanded="openMenuSurveyId === s.id"
              title="Seçenekler"
              @click.stop="toggleSurveyMenu(s.id)"
            >
              &#8942;
            </button>
            <div v-if="openMenuSurveyId === s.id" class="x-more-menu" @click.stop>
              <button class="x-more-item" type="button" @click="openSurveyDetail(s.id)">
                Anketi Aç
              </button>
            </div>
          </div>

          <SurveyUserHeader
            :avatar-color="getAvatarColor(s.id)"
            :avatar-text="(s.questions && s.questions.length > 0 && getCurrentQuestion(s)?.text) ? getCurrentQuestion(s).text.charAt(0).toUpperCase() : '?'"
            :display-name="getCreatorName(s)"
            :handle="getCreatorHandle(s)"
            :time-ago="timeAgo(s.created_at)"
            :is-new="!completedSurveys.has(s.id)"
          />

          <SurveyQuestionBlock
            :survey-id="s.id"
            :question="getCurrentQuestion(s)"
            :question-count="s?.questions?.length || 0"
            :question-index="(activeQuestionIndexes[s.id] ?? 0) + 1"
            :can-go-prev="canGoPrevQuestion(s)"
            :can-go-next="canGoNextQuestion(s)"
            :selected-answer="getCurrentQuestion(s)?.id != null ? (givenAnswers[String(getCurrentQuestion(s).id)] || '') : ''"
            :is-completed="completedSurveys.has(s.id)"
            :is-answer-locked="isCurrentQuestionAnswerLocked(s)"
            :get-percent="(opt: any) => getAnimatedPercent(s, getCurrentQuestion(s), opt)"
            :api-base="API_BASE"
            @prev="goToPrevQuestion(s)"
            @next="goToNextQuestion(s)"
            @select="handleQuestionSelect(s, $event)"
          />

          <!-- Alt Aksiyonlar -->
          <div class="x-actions">
            <p v-if="completedSurveys.has(s.id)" class="x-total-votes">
              {{ getTotalVotes(getCurrentQuestion(s)) }} oy
            </p>
            <button
              class="x-action-btn x-share-btn"
              @click="copyLink(s.id)"
              :class="{ copied: copiedSurveyId === s.id }"
              :title="copiedSurveyId === s.id ? 'Kopyalandı!' : 'Linki Kopyala'"
            >
              <svg
                v-if="copiedSurveyId !== s.id"
                xmlns="http://www.w3.org/2000/svg"
                width="17" height="17"
                viewBox="0 0 24 24"
                fill="none" stroke="currentColor"
                stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
              >
                <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/>
                <polyline points="16 6 12 2 8 6"/>
                <line x1="12" y1="2" x2="12" y2="15"/>
              </svg>
              <span v-else style="font-size:13px; font-weight:700;">✓</span>
            </button>
          </div>
        </div>
      </article>
    </main>

    <!-- Sağ Panel -->
    <aside class="x-right-panel">
      <router-link to="/settings" class="x-user-card x-user-card-link" title="Profil ayarları">
        <div class="x-user-avatar" :style="{ background: getAvatarColor(authStore.user.id) }">
          {{ authStore.user.name.charAt(0).toUpperCase() }}
        </div>
        <div class="x-user-info">
          <div class="x-user-name">{{ authStore.user.name }}</div>
          <div class="x-user-handle">@{{ authStore.user.username || authStore.user.email?.split('@')[0] || 'kullanici' }}</div>
        </div>
      </router-link>

      <div class="x-stats-card">
        <div class="x-stats-title">Özet</div>
        <div class="x-stat-row">
          <span class="x-stat-label">Cevaplanmış</span>
          <span class="x-stat-value">{{ completedSurveys.size }}</span>
        </div>
        <div class="x-stat-row">
          <span class="x-stat-label">Toplam Anket</span>
          <span class="x-stat-value">{{ surveys.length }}</span>
        </div>
      </div>
    </aside>

    <CreateSurveyModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="handleSurveyCreated"
    />
  </div>
</template>

<style scoped>
/* ─── Layout ──────────────────────────────────────────── */
.x-layout {
  display: flex;
  min-height: 100dvh;
  background: transparent;
  max-width: 1280px;
  width: 100%;
  margin: 0 auto;
  overflow-x: hidden;
  padding-bottom: env(safe-area-inset-bottom, 0);
}

/* ─── Sol Kenar Çubuğu ─────────────────────────────────── */
.x-sidebar {
  width: 275px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: 100vh;
  height: 100dvh;
  background: transparent;
  border-right: 1px solid var(--border-color);
}
.x-sidebar-inner {
  display: flex;
  flex-direction: column;
  padding: 12px 16px;
  height: 100%;
}
.x-logo {
  display: block;
  padding: 12px 16px;
  color: var(--primary-color);
  font-size: var(--font-size-2xl);
  font-weight: 700;
  letter-spacing: -0.05em;
  margin-bottom: 8px;
  text-decoration: none;
}
.x-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.x-nav-item {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 12px 16px;
  border-radius: 9999px;
  font-size: var(--font-size-lg);
  font-weight: 500;
  color: var(--text-color);
  text-decoration: none;
  transition: background 0.15s;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  width: 100%;
}
.x-nav-item:hover {
  background: var(--bg-color);
}
.x-nav-item.router-link-active {
  font-weight: 700;
}
.x-logout {
  margin-top: 8px;
  color: var(--text-color-muted);
}

/* ─── Ana Feed ─────────────────────────────────────────── */
.x-feed {
  flex: 1;
  min-width: 0;
  border-right: 1px solid var(--border-color);
  max-width: 600px;
}
.x-feed-header {
  position: sticky;
  top: 0;
  background: color-mix(in srgb, var(--color-white) 88%, transparent);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  padding: 14px 20px;
  padding-top: max(14px, env(safe-area-inset-top, 0px));
  border-bottom: 1px solid var(--border-color);
  z-index: 10;
}
.x-feed-header h2 {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--text-color);
}
.x-state-msg {
  padding: 32px;
  text-align: center;
  color: var(--text-color-muted);
}

/* ─── Post ─────────────────────────────────────────────── */
.x-post {
  display: flex;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.15s;
  cursor: default;
}
.x-post:hover {
  background: #f7f9f9;
}
.x-post-body {
  flex: 1;
  min-width: 0;
  position: relative;
}
.x-post-top-actions {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 20;
}
.x-more-btn {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 9999px;
  background: transparent;
  color: var(--text-color-muted);
  cursor: pointer;
  font-size: 20px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease, color 0.2s ease;
}
.x-more-btn:hover {
  background: var(--bg-color);
  color: var(--text-color);
}
.x-more-menu {
  position: absolute;
  top: 34px;
  right: 0;
  min-width: 140px;
  padding: 6px;
  border-radius: 10px;
  border: 1px solid var(--border-color);
  background: var(--color-white);
  box-shadow: var(--shadow-md);
}
.x-more-item {
  width: 100%;
  border: none;
  background: transparent;
  text-align: left;
  padding: 8px 10px;
  border-radius: 8px;
  color: var(--text-color);
  font-size: 14px;
  cursor: pointer;
}
.x-more-item:hover {
  background: var(--bg-color);
}
.x-total-votes {
  font-size: 13px;
  color: var(--text-color-muted);
  margin: 0;
}

/* ─── Aksiyonlar ────────────────────────────────────────── */
.x-actions {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  margin-top: 12px;
  padding-top: 8px;
  gap: 12px;
}
.x-share-btn {
  margin-left: auto;
}
.x-action-btn {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: none;
  background: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-color-muted);
  transition: all 0.2s;
}
.x-action-btn:hover,
.x-share-btn:hover {
  background: var(--bg-color);
  color: var(--primary-color);
}
.x-share-btn.copied {
  color: var(--secondary-color);
}

/* ─── Sağ Panel ─────────────────────────────────────────── */
.x-right-panel {
  width: 350px;
  flex-shrink: 0;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: transparent;
  position: sticky;
  top: 16px;
  align-self: flex-start;
}
.x-user-card {
  background: var(--color-white);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.x-user-card-link {
  text-decoration: none;
  color: inherit;
  cursor: pointer;
  transition: background 0.15s, border-color 0.15s;
}
.x-user-card-link:hover {
  background: var(--bg-color);
  border-color: var(--color-gray-300);
}
.x-user-avatar {
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
.x-user-info {
  display: flex;
  flex-direction: column;
}
.x-user-name {
  font-weight: 700;
  font-size: 15px;
  color: var(--text-color);
}
.x-user-handle {
  font-size: 13px;
  color: var(--text-color-muted);
}
.x-stats-card {
  background: var(--color-white);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 16px;
}
.x-stats-title {
  font-weight: 700;
  font-size: 17px;
  color: var(--text-color);
  margin-bottom: 12px;
}
.x-stat-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}
.x-stat-row:last-child {
  border-bottom: none;
}
.x-stat-label {
  font-size: 14px;
  color: var(--text-color-muted);
}
.x-stat-value {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-color);
}

/* ─── Responsive ────────────────────────────────────────── */
@media (max-width: 1100px) {
  .x-right-panel { display: none; }
}
@media (max-width: 700px) {
  .x-sidebar {
    width: 64px;
    min-width: 64px;
  }
  .x-sidebar-inner {
    padding: 8px 6px;
    padding-top: max(8px, env(safe-area-inset-top, 0px));
  }
  .x-nav-item span, .x-logo span { display: none; }
  .x-nav-item {
    justify-content: center;
    min-height: 48px;
    min-width: 48px;
    padding: 12px;
    border-radius: 12px;
  }
  .x-logo {
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 0.7rem;
    font-weight: 800;
    padding: 10px 4px;
    line-height: 1.15;
    text-align: center;
    word-break: break-word;
    hyphens: auto;
  }
  .x-feed-header {
    padding-left: 12px;
    padding-right: 12px;
  }
  .x-post {
    padding: 12px 12px;
  }
}
@media (max-width: 380px) {
  .x-sidebar {
    width: 56px;
    min-width: 56px;
  }
  .x-nav-item {
    min-height: 44px;
    min-width: 44px;
    padding: 10px 8px;
  }
}
</style>
