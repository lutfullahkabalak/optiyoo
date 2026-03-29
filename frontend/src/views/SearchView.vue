<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const CreateSurveyModal = defineAsyncComponent(() => import('../components/CreateSurveyModal.vue'))
import SurveyUserHeader from '../components/survey/SurveyUserHeader.vue'
import SurveyQuestionBlock from '../components/survey/SurveyQuestionBlock.vue'
import UserDropdownMenu from '../components/UserDropdownMenu.vue'
import { absoluteApiUrl, resolveAvatarBackground, usernameInitial } from '../utils/avatar'
import { surveyFeedApiBase, useSurveyFeed } from '../composables/useSurveyFeed'

const API_BASE = surveyFeedApiBase

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const searchQ = ref('')
const isLoading = ref(true)
const showCreateModal = ref(false)

const feed = useSurveyFeed({
  onConflictReload: async (apply) => {
    const q = String(route.query.q || '').trim()
    if (!q) return
    const uid = authStore.user?.id
    const params = new URLSearchParams({ q })
    if (uid) params.set('user_id', uid)
    const res = await fetch(`${surveyFeedApiBase}/api/search?${params.toString()}`, {
      headers: uid ? authStore.authHeadersGet() : {}
    })
    if (res.ok) apply(await res.json())
  }
})

const {
  surveys,
  givenAnswers,
  copiedSurveyId,
  openMenuSurveyId,
  activeQuestionIndexes,
  completedSurveys,
  applySurveyListPayload,
  copyLink,
  toggleSurveyMenu,
  closeSurveyMenu,
  openSurveyDetail,
  getTotalVotes,
  getCurrentQuestion,
  isCurrentQuestionAnswerLocked,
  canGoPrevQuestion,
  canGoNextQuestion,
  goToPrevQuestion,
  goToNextQuestion,
  getAnimatedPercent,
  handleQuestionSelect,
  timeAgo,
  getCreatorName,
  getCreatorHandle,
  creatorAvatarImageUrl,
  creatorAvatarBg,
  creatorAvatarLetter
} = feed

const fetchSearchResults = async (opts?: { silent?: boolean }) => {
  const q = String(route.query.q || '').trim()
  if (!q) {
    router.replace('/')
    return
  }
  if (!opts?.silent) isLoading.value = true
  try {
    const uid = authStore.user?.id
    const params = new URLSearchParams({ q })
    if (uid) params.set('user_id', uid)
    const res = await fetch(`${surveyFeedApiBase}/api/search?${params.toString()}`, {
      headers: uid ? authStore.authHeadersGet() : {}
    })
    if (res.ok) {
      applySurveyListPayload(await res.json())
    }
  } catch (e) {
    console.error(e)
  } finally {
    if (!opts?.silent) isLoading.value = false
  }
}

watch(
  () => route.fullPath,
  () => {
    if (!authStore.isAuthenticated) {
      router.push('/auth')
      return
    }
    if (route.path !== '/search') return
    const raw = route.query.q
    searchQ.value = raw != null && raw !== '' ? String(raw) : ''
    const q = searchQ.value.trim()
    if (!q) {
      router.replace('/')
      return
    }
    fetchSearchResults()
  },
  { immediate: true }
)

const submitSearch = () => {
  const q = searchQ.value.trim()
  if (!q) return
  router.push({ path: '/search', query: { q } })
}

const handleLogout = () => {
  authStore.logout()
  router.push('/auth')
}

const handleSurveyCreated = async () => {
  showCreateModal.value = false
  await fetchSearchResults({ silent: true })
}
</script>

<template>
  <div class="x-layout" v-if="authStore.user">
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

    <main class="x-feed">
      <div class="x-feed-header">
        <div class="x-feed-header-bar">
          <router-link to="/" class="x-feed-header-brand">optiyoo</router-link>
          <h2 class="x-feed-header-title">Arama</h2>
          <form class="x-header-search-form" @submit.prevent="submitSearch">
            <label class="x-search-field">
              <span class="x-search-icon" aria-hidden="true">
                <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="11" cy="11" r="8" />
                  <path d="m21 21-4.35-4.35" />
                </svg>
              </span>
              <input
                v-model="searchQ"
                class="x-search-input"
                type="search"
                name="q"
                placeholder=""
                enterkeyhint="search"
                autocomplete="off"
                aria-label="Anket ara"
              />
            </label>
          </form>
          <div class="x-feed-header-profile-wrap">
            <UserDropdownMenu />
          </div>
        </div>
      </div>

      <div v-if="isLoading" class="x-state-msg">Sonuçlar yükleniyor...</div>
      <div v-else-if="surveys.length === 0" class="x-state-msg">Aramanızla eşleşen anket yok.</div>

      <article v-for="s in surveys" :key="s.id" class="x-post" @click="closeSurveyMenu">
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
            :avatar-color="creatorAvatarBg(s)"
            :avatar-text="creatorAvatarLetter(s)"
            :avatar-image-url="creatorAvatarImageUrl(s)"
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
                width="17"
                height="17"
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

    <aside class="x-right-panel">
      <router-link to="/settings" class="x-user-card x-user-card-link" title="Profil ayarları">
        <div
          class="x-user-avatar"
          :class="{ 'x-user-avatar--photo': !!authStore.user.avatar_url }"
          :style="authStore.user.avatar_url ? undefined : { background: resolveAvatarBackground(authStore.user.avatar_color, authStore.user.id) }"
        >
          <img
            v-if="authStore.user.avatar_url"
            class="x-user-avatar-img"
            :src="absoluteApiUrl(API_BASE, authStore.user.avatar_url)"
            alt=""
          />
          <template v-else>{{ usernameInitial(authStore.user.username) }}</template>
        </div>
        <div class="x-user-info">
          <div class="x-user-name">{{ authStore.user.name }}</div>
          <div class="x-user-handle">@{{ authStore.user.username || authStore.user.email?.split('@')[0] || 'kullanici' }}</div>
        </div>
      </router-link>

      <div class="x-stats-card">
        <div class="x-stats-title">Arama özeti</div>
        <div class="x-stat-row">
          <span class="x-stat-label">Bulunan</span>
          <span class="x-stat-value">{{ surveys.length }}</span>
        </div>
      </div>
    </aside>

    <button
      type="button"
      class="x-fab-create"
      aria-label="Anket oluştur"
      title="Anket oluştur"
      @click="showCreateModal = true"
    >
      <svg viewBox="0 0 24 24" width="26" height="26" fill="none" stroke="currentColor" stroke-width="2.25" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="12" cy="12" r="10" />
        <line x1="12" y1="8" x2="12" y2="16" />
        <line x1="8" y1="12" x2="16" y2="12" />
      </svg>
    </button>

    <CreateSurveyModal
      v-if="showCreateModal"
      @close="showCreateModal = false"
      @created="handleSurveyCreated"
    />
  </div>
</template>

<style scoped src="../styles/survey-feed-layout.css"></style>
