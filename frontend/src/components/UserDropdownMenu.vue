<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { absoluteApiUrl, resolveAvatarBackground, usernameInitial } from '../utils/avatar'
import { surveyFeedApiBase } from '../composables/useSurveyFeed'

const API_BASE = surveyFeedApiBase

const router = useRouter()
const authStore = useAuthStore()
const open = ref(false)
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuStyle = ref<Record<string, string>>({ top: '0px', right: '0px' })

const syncMenuPosition = () => {
  const el = triggerRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  const gap = 8
  menuStyle.value = {
    top: `${Math.round(r.bottom + gap)}px`,
    right: `${Math.round(window.innerWidth - r.right)}px`
  }
}

const onViewportChange = () => {
  if (open.value) syncMenuPosition()
}

const toggle = () => {
  open.value = !open.value
}
const close = () => {
  open.value = false
}

const onLogout = () => {
  close()
  authStore.logout()
  router.push('/auth')
}

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') close()
}

watch(open, async (isOpen) => {
  if (!isOpen) return
  await nextTick()
  syncMenuPosition()
})

onMounted(() => {
  window.addEventListener('scroll', onViewportChange, true)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('keydown', onKeydown)
})

onUnmounted(() => {
  window.removeEventListener('scroll', onViewportChange, true)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div v-if="authStore.user" class="x-user-menu-root">
    <button
      ref="triggerRef"
      type="button"
      class="x-user-menu-trigger"
      aria-haspopup="menu"
      :aria-expanded="open"
      aria-label="Hesap menüsü"
      @click.stop="toggle"
    >
      <div
        class="x-user-avatar x-user-avatar--toolbar"
        :class="{ 'x-user-avatar--photo': !!authStore.user.avatar_url }"
        :style="
          authStore.user.avatar_url
            ? undefined
            : { background: resolveAvatarBackground(authStore.user.avatar_color, authStore.user.id) }
        "
      >
        <img
          v-if="authStore.user.avatar_url"
          class="x-user-avatar-img"
          :src="absoluteApiUrl(API_BASE, authStore.user.avatar_url)"
          alt=""
        />
        <template v-else>{{ usernameInitial(authStore.user.username) }}</template>
      </div>
    </button>

    <Teleport to="body">
      <div v-if="open" class="x-menu-backdrop" aria-hidden="true" @click="close" />
      <div
        v-if="open"
        class="x-user-menu-dropdown x-user-menu-dropdown--portal"
        role="menu"
        :style="menuStyle"
        @click.stop
      >
        <router-link to="/settings" class="x-user-menu-item" role="menuitem" @click="close">
          Profil
        </router-link>
        <button type="button" class="x-user-menu-item x-user-menu-item--danger" role="menuitem" @click="onLogout">
          Çıkış Yap
        </button>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
/* Bu bileşen ayrı data-v-* alır; üstteki scoped survey-feed-layout kuralları buraya uygulanmaz. */
.x-user-menu-root {
  position: relative;
  flex-shrink: 0;
  z-index: 20;
}

.x-user-menu-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 44px;
  padding: 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  border-radius: 50%;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.x-user-menu-trigger:focus-visible {
  outline: none;
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--primary-color) 35%, transparent);
}

.x-user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 16px;
  line-height: 1;
  color: #fff;
  flex-shrink: 0;
  overflow: hidden;
}

.x-user-avatar--toolbar {
  width: 40px;
  height: 40px;
  font-size: 16px;
}

.x-user-avatar--photo {
  background: var(--border-color, #e5e7eb);
}

.x-user-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.x-menu-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9998;
  background: rgba(1, 69, 94, 0.12);
}

.x-user-menu-dropdown {
  min-width: 176px;
  padding: 8px;
  border-radius: 12px;
  border: 1px solid var(--border-color);
  background: var(--color-white);
  box-shadow: var(--shadow-md);
}

.x-user-menu-dropdown--portal {
  position: fixed;
  z-index: 9999;
  max-width: min(240px, calc(100vw - 24px));
  font-family: var(--font-family-base);
  font-size: 16px;
  line-height: 1.4;
  -webkit-font-smoothing: antialiased;
  pointer-events: auto;
}

.x-user-menu-item {
  display: block;
  width: 100%;
  box-sizing: border-box;
  border: none;
  background: transparent;
  text-align: left;
  padding: 12px 14px;
  border-radius: 8px;
  color: var(--text-color);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  text-decoration: none;
  font-family: inherit;
  transition: background 0.15s;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.x-user-menu-item:hover {
  background: var(--bg-color);
}

.x-user-menu-item--danger {
  color: var(--text-color-muted);
  margin-top: 4px;
  border-top: 1px solid var(--border-color);
  border-radius: 0 0 8px 8px;
  padding-top: 14px;
}

.x-user-menu-item--danger:hover {
  color: var(--text-color);
}
</style>
