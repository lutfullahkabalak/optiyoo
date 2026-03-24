<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import CreateSurveyModal from '../components/CreateSurveyModal.vue'

const router = useRouter()
const authStore = useAuthStore()

const surveys = ref<any[]>([])
const isLoading = ref(true)
const showCreateModal = ref(false)

const givenAnswers = ref<Record<string, string>>({})
const copiedSurveyId = ref<string | null>(null)
const animatedPercents = ref<Record<string, number>>({})
const openMenuSurveyId = ref<string | null>(null)

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

const fetchSurveys = async () => {
  isLoading.value = true
  try {
    const res = await fetch(`http://localhost:8080/api/surveys?user_id=${authStore.user?.id}`)
    if (res.ok) {
      surveys.value = await res.json()
      surveys.value.forEach((s: any) => {
        if (s.user_answer) {
          givenAnswers.value[s.questions[0].id] = s.user_answer
          completedSurveys.value.add(s.id)
        }
        if (completedSurveys.value.has(s.id)) {
          setSurveyPercentsInstant(s)
        }
      })
    }
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
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

const getAnimatedPercent = (survey: any, q: any, opt: any) => {
  const key = getOptionAnimKey(survey.id, opt.id)
  return animatedPercents.value[key] ?? getOptionPercent(q, opt)
}

const setSurveyPercentsInstant = (survey: any) => {
  const q = survey?.questions?.[0]
  if (!q?.options) return

  q.options.forEach((opt: any) => {
    const key = getOptionAnimKey(survey.id, opt.id)
    animatedPercents.value[key] = getOptionPercent(q, opt)
  })
}

const animateSurveyPercents = (survey: any) => {
  const q = survey?.questions?.[0]
  if (!q?.options) return

  const targets = q.options.map((opt: any) => ({
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

const submitPoll = async (survey: any, val: string) => {
  if (!val || completedSurveys.value.has(survey.id)) return

  submittingState.value[survey.id] = true
  const payload = {
    user_id: authStore.user?.id,
    answers: [{ question_id: survey.questions[0].id, value: val }]
  }

  try {
    const res = await fetch(`http://localhost:8080/api/surveys/${survey.id}/answers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })

    if (res.ok) {
      const q = survey.questions[0]
      const opt = q.options.find((o: any) => o.id === val)
      if (opt) opt.vote_count = (opt.vote_count || 0) + 1

      completedSurveys.value.add(survey.id)
      localStorage.setItem('completed_polls_' + authStore.user?.id, JSON.stringify(Array.from(completedSurveys.value)))

      await nextTick()
      animateSurveyPercents(survey)
    } else {
      const text = await res.text()
      if (res.status === 403) {
        completedSurveys.value.add(survey.id)
        localStorage.setItem('completed_polls_' + authStore.user?.id, JSON.stringify(Array.from(completedSurveys.value)))
        setSurveyPercentsInstant(survey)
      } else {
        alert('Oylama İptali: ' + text)
      }
    }
  } catch (e) {
    console.error(e)
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
  const name = getCreatorName(survey)
  return `@${name.toLowerCase().replace(/\s+/g, '')}`
}
</script>

<template>
  <div class="x-layout" v-if="authStore.user">

    <!-- Sol Kenar Çubuğu -->
    <aside class="x-sidebar">
      <div class="x-sidebar-inner">
        <div class="x-logo">optiyoo</div>

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
        <!-- Avatar -->
        <div class="x-avatar" :style="{ background: getAvatarColor(s.id) }">
          {{ (s.questions && s.questions.length > 0 && s.questions[0].text) ? s.questions[0].text.charAt(0).toUpperCase() : '?' }}
        </div>

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

          <!-- Post Başlığı -->
          <div class="x-post-meta">
            <span class="x-display-name">{{ getCreatorName(s) }}</span>
            <span class="x-handle">{{ getCreatorHandle(s) }}</span>
            <span class="x-dot">·</span>
            <span class="x-time">{{ timeAgo(s.created_at) }}</span>
            <span v-if="!completedSurveys.has(s.id)" class="x-new-badge">Yeni</span>
          </div>

          <!-- Soru Metni -->
          <p class="x-post-title">{{ s.questions && s.questions.length > 0 ? s.questions[0].text : 'Soru bulunamadı' }}</p>

          <!-- Poll Seçenekleri -->
          <div v-if="s.questions && s.questions.length > 0 && s.questions[0].type === 'single_choice'" class="x-poll">
            <label
              v-for="opt in s.questions[0].options"
              :key="opt.id"
              class="x-poll-opt"
              :class="{
                'is-voted': completedSurveys.has(s.id),
                'is-selected': givenAnswers[s.questions[0].id] === opt.id,
                'is-disabled': completedSurveys.has(s.id)
              }"
            >
              <!-- Progress bar -->
              <div
                class="x-progress-fill"
                :class="{
                  'is-selected-fill': givenAnswers[s.questions[0].id] === opt.id,
                  'is-visible': completedSurveys.has(s.id)
                }"
                :style="{ width: `${completedSurveys.has(s.id) ? getAnimatedPercent(s, s.questions[0], opt) : 0}%` }"
              ></div>

              <div class="x-poll-opt-content">
                <input
                  type="radio"
                  :name="'survey_' + s.id"
                  :value="opt.id"
                  v-model="givenAnswers[s.questions[0].id]"
                  @change="submitPoll(s, opt.id)"
                  :disabled="completedSurveys.has(s.id)"
                  style="display:none"
                />
                <div class="x-radio" :class="{ 'is-checked': givenAnswers[s.questions[0].id] === opt.id }"></div>
                <span class="x-opt-text">{{ opt.text }}</span>
                <span class="x-opt-pct" :class="{ 'is-visible': completedSurveys.has(s.id) }">
                  {{ getAnimatedPercent(s, s.questions[0], opt) }}%
                </span>
              </div>
            </label>

          </div>

          <!-- Alt Aksiyonlar -->
          <div class="x-actions">
            <p v-if="completedSurveys.has(s.id)" class="x-total-votes">
              {{ getTotalVotes(s.questions[0]) }} oy
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
      <div class="x-user-card">
        <div class="x-user-avatar" :style="{ background: getAvatarColor(authStore.user.id) }">
          {{ authStore.user.name.charAt(0).toUpperCase() }}
        </div>
        <div class="x-user-info">
          <div class="x-user-name">{{ authStore.user.name }}</div>
          <div class="x-user-handle">@{{ authStore.user.email?.split('@')[0] || 'kullanici' }}</div>
        </div>
      </div>

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
  min-height: 100vh;
  background: transparent;
  max-width: 1280px;
  margin: 0 auto;
}

/* ─── Sol Kenar Çubuğu ─────────────────────────────────── */
.x-sidebar {
  width: 275px;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  height: 100vh;
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
  padding: 12px 16px;
  color: var(--primary-color);
  font-size: var(--font-size-2xl);
  font-weight: 700;
  letter-spacing: -0.05em;
  margin-bottom: 8px;
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
  background: rgba(255,255,255,0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  padding: 14px 20px;
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
.x-avatar {
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
.x-post-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-wrap: wrap;
  margin-bottom: 4px;
}
.x-display-name {
  font-weight: 700;
  font-size: 15px;
  color: var(--text-color);
}
.x-handle, .x-time {
  font-size: 14px;
  color: var(--text-color-muted);
}
.x-dot {
  color: var(--text-color-muted);
  font-size: 14px;
}
.x-new-badge {
  margin-left: 6px;
  background: var(--primary-color);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 9999px;
}
.x-post-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: 2px;
}
.x-post-desc {
  font-size: 14px;
  color: var(--text-color-muted);
  margin-bottom: 12px;
}
.x-submitting {
  font-size: 14px;
  color: var(--primary-color);
  padding: 8px 0;
  font-weight: 600;
}

/* ─── Poll ─────────────────────────────────────────────── */
.x-poll {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 12px;
}
.x-poll-opt {
  position: relative;
  border: 2px solid var(--border-color);
  border-radius: 9999px;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s;
}
.x-poll-opt:not(.is-disabled):hover {
  border-color: var(--primary-color);
}
.x-poll-opt.is-selected {
  border-color: var(--primary-color);
}
.x-poll-opt.is-disabled {
  cursor: default;
}
.x-progress-fill {
  position: absolute;
  top: 0; left: 0; bottom: 0;
  background: #c8c4bd;
  opacity: 0;
  border-radius: 9999px;
  transition: width 0.4s cubic-bezier(0.2, 0.8, 0.2, 1);
  z-index: 0;
}
.x-progress-fill.is-visible {
  opacity: 0.92;
}
.x-progress-fill.is-selected-fill {
  background: #c8c4bd;
}
.x-poll-opt-content {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  padding: 10px 18px;
  gap: 12px;
}
.x-radio {
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
.x-radio.is-checked {
  border-color: var(--primary-color);
}
.x-radio.is-checked::after {
  content: '';
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: var(--primary-color);
}
.x-opt-text {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-color);
  flex: 1;
}
.x-opt-pct {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-color);
  min-width: 44px;
  text-align: right;
  opacity: 0;
  transform: translateX(4px);
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.x-opt-pct.is-visible {
  opacity: 1;
  transform: translateX(0);
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
  .x-sidebar { width: 70px; }
  .x-nav-item span, .x-logo span { display: none; }
  .x-nav-item { justify-content: center; }
  .x-logo { display: flex; justify-content: center; }
}
</style>
