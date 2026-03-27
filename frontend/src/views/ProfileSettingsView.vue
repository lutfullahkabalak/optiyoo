<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const router = useRouter()
const authStore = useAuthStore()

const loadError = ref('')
const displayNameMsg = ref('')
const usernameMsg = ref('')
const emailMsg = ref('')
const passwordMsg = ref('')

const displayNameCurrentPw = ref('')
const newDisplayName = ref('')

const usernameCurrentPw = ref('')
const newUsername = ref('')

const emailCurrentPw = ref('')
const newEmail = ref('')

const passwordCurrentPw = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const busyDisplayName = ref(false)
const busyUser = ref(false)
const busyEmail = ref(false)
const busyPassword = ref(false)

const syncFormDefaults = () => {
  const u = authStore.user
  if (!u) return
  newDisplayName.value = u.name || ''
  newUsername.value = u.username || ''
  newEmail.value = u.email || ''
}

onMounted(async () => {
  if (!authStore.isAuthenticated || !authStore.user?.id) {
    router.push('/auth')
    return
  }
  syncFormDefaults()
  try {
    const res = await fetch(
      `${API_BASE}/api/users/${encodeURIComponent(authStore.user.id)}?user_id=${encodeURIComponent(authStore.user.id)}`
    )
    if (!res.ok) {
      loadError.value = await res.text()
      return
    }
    const data = await res.json()
    authStore.setUser({ ...authStore.user, ...data })
    syncFormDefaults()
  } catch (e) {
    loadError.value = 'Profil bilgisi alınamadı.'
    console.error(e)
  }
})

const patchUser = async (
  body: Record<string, string>,
  setBusy: (v: boolean) => void,
  setMsg: (v: string) => void
) => {
  const uid = authStore.user?.id
  if (!uid) return
  setMsg('')
  setBusy(true)
  try {
    const res = await fetch(`${API_BASE}/api/users/${encodeURIComponent(uid)}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: uid, ...body })
    })
    const text = await res.text()
    if (!res.ok) {
      setMsg(text || 'İşlem başarısız.')
      return
    }
    const data = JSON.parse(text)
    authStore.setUser(data)
    syncFormDefaults()
    displayNameCurrentPw.value = ''
    usernameCurrentPw.value = ''
    emailCurrentPw.value = ''
    passwordCurrentPw.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    setMsg('Kaydedildi.')
  } catch (e) {
    setMsg('Bağlantı hatası.')
    console.error(e)
  } finally {
    setBusy(false)
  }
}

const saveDisplayName = () => {
  const n = newDisplayName.value.trim()
  if (!n) {
    displayNameMsg.value = 'Görünen ad gerekli.'
    return
  }
  if (!displayNameCurrentPw.value) {
    displayNameMsg.value = 'Mevcut şifrenizi girin.'
    return
  }
  patchUser(
    { current_password: displayNameCurrentPw.value, name: n },
    (v) => (busyDisplayName.value = v),
    (m) => (displayNameMsg.value = m)
  )
}

const saveUsername = () => {
  const u = newUsername.value.trim()
  if (!u) {
    usernameMsg.value = 'Kullanıcı adı gerekli.'
    return
  }
  if (!usernameCurrentPw.value) {
    usernameMsg.value = 'Mevcut şifrenizi girin.'
    return
  }
  patchUser(
    { current_password: usernameCurrentPw.value, username: u },
    (v) => (busyUser.value = v),
    (m) => (usernameMsg.value = m)
  )
}

const saveEmail = () => {
  const e = newEmail.value.trim()
  if (!e) {
    emailMsg.value = 'E-posta gerekli.'
    return
  }
  if (!emailCurrentPw.value) {
    emailMsg.value = 'Mevcut şifrenizi girin.'
    return
  }
  patchUser(
    { current_password: emailCurrentPw.value, email: e },
    (v) => (busyEmail.value = v),
    (m) => (emailMsg.value = m)
  )
}

const savePassword = () => {
  if (!passwordCurrentPw.value) {
    passwordMsg.value = 'Mevcut şifrenizi girin.'
    return
  }
  if (newPassword.value.length < 6) {
    passwordMsg.value = 'Yeni şifre en az 6 karakter olmalı.'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    passwordMsg.value = 'Yeni şifreler eşleşmiyor.'
    return
  }
  patchUser(
    { current_password: passwordCurrentPw.value, new_password: newPassword.value },
    (v) => (busyPassword.value = v),
    (m) => (passwordMsg.value = m)
  )
}

const avatarColor = (id: string) => {
  const colors = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#ef4444', '#14b8a6']
  let hash = 0
  for (let i = 0; i < id.length; i++) hash = id.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}
</script>

<template>
  <div class="settings-page" v-if="authStore.user">
    <header class="settings-header">
      <div class="container settings-header-inner">
        <router-link to="/" class="back-link">← Ana Sayfa</router-link>
        <router-link to="/" class="logo-link">
          <span class="logo">optiyoo</span>
        </router-link>
      </div>
    </header>

    <main class="container settings-main">
      <div v-if="loadError" class="banner-error">{{ loadError }}</div>

      <div class="profile-head card">
        <div
          class="profile-avatar"
          :style="{ background: avatarColor(authStore.user.id) }"
        >
          {{ authStore.user.name.charAt(0).toUpperCase() }}
        </div>
        <div>
          <h1 class="profile-title">Profil ayarları</h1>
          <p class="profile-sub">{{ authStore.user.name }}</p>
        </div>
      </div>

      <section class="card settings-section">
        <h2>Görünen ad</h2>
        <p class="hint">Profil ve kartlarda görünen adınız (ad soyad).</p>
        <div class="form-group">
          <label class="form-label">Görünen ad</label>
          <input v-model="newDisplayName" type="text" class="form-control" autocomplete="name" maxlength="255" />
        </div>
        <div class="form-group">
          <label class="form-label">Mevcut şifre</label>
          <input v-model="displayNameCurrentPw" type="password" class="form-control" autocomplete="current-password" />
        </div>
        <p v-if="displayNameMsg" class="msg" :class="{ ok: displayNameMsg === 'Kaydedildi.' }">{{ displayNameMsg }}</p>
        <button type="button" class="btn btn-primary" :disabled="busyDisplayName" @click="saveDisplayName">
          {{ busyDisplayName ? 'Kaydediliyor…' : 'Görünen adı kaydet' }}
        </button>
      </section>

      <section class="card settings-section">
        <h2>Kullanıcı adı</h2>
        <p class="hint">Görünen @kullanıcı_adı bilgisini günceller.</p>
        <div class="form-group">
          <label class="form-label">Yeni kullanıcı adı</label>
          <input v-model="newUsername" type="text" class="form-control" autocomplete="username" />
        </div>
        <div class="form-group">
          <label class="form-label">Mevcut şifre</label>
          <input v-model="usernameCurrentPw" type="password" class="form-control" autocomplete="current-password" />
        </div>
        <p v-if="usernameMsg" class="msg" :class="{ ok: usernameMsg === 'Kaydedildi.' }">{{ usernameMsg }}</p>
        <button type="button" class="btn btn-primary" :disabled="busyUser" @click="saveUsername">
          {{ busyUser ? 'Kaydediliyor…' : 'Kullanıcı adını kaydet' }}
        </button>
      </section>

      <section class="card settings-section">
        <h2>E-posta</h2>
        <p class="hint">Giriş için kullandığınız adresi değiştirir.</p>
        <div class="form-group">
          <label class="form-label">Yeni e-posta</label>
          <input v-model="newEmail" type="email" class="form-control" autocomplete="email" />
        </div>
        <div class="form-group">
          <label class="form-label">Mevcut şifre</label>
          <input v-model="emailCurrentPw" type="password" class="form-control" autocomplete="current-password" />
        </div>
        <p v-if="emailMsg" class="msg" :class="{ ok: emailMsg === 'Kaydedildi.' }">{{ emailMsg }}</p>
        <button type="button" class="btn btn-primary" :disabled="busyEmail" @click="saveEmail">
          {{ busyEmail ? 'Kaydediliyor…' : 'E-postayı kaydet' }}
        </button>
      </section>

      <section class="card settings-section">
        <h2>Şifre</h2>
        <p class="hint">Hesap şifrenizi yenileyin (en az 6 karakter).</p>
        <div class="form-group">
          <label class="form-label">Mevcut şifre</label>
          <input v-model="passwordCurrentPw" type="password" class="form-control" autocomplete="current-password" />
        </div>
        <div class="form-group">
          <label class="form-label">Yeni şifre</label>
          <input v-model="newPassword" type="password" class="form-control" autocomplete="new-password" />
        </div>
        <div class="form-group">
          <label class="form-label">Yeni şifre (tekrar)</label>
          <input v-model="confirmPassword" type="password" class="form-control" autocomplete="new-password" />
        </div>
        <p v-if="passwordMsg" class="msg" :class="{ ok: passwordMsg === 'Kaydedildi.' }">{{ passwordMsg }}</p>
        <button type="button" class="btn btn-primary" :disabled="busyPassword" @click="savePassword">
          {{ busyPassword ? 'Kaydediliyor…' : 'Şifreyi güncelle' }}
        </button>
      </section>
    </main>
  </div>
</template>

<style scoped>
.settings-page {
  min-height: 100vh;
  background: var(--bg-color);
}
.settings-header {
  background: var(--color-white);
  border-bottom: 1px solid var(--border-color);
  padding: var(--spacing-4) 0;
  padding-top: max(var(--spacing-4), env(safe-area-inset-top, 0px));
}
.settings-header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-4);
  flex-wrap: wrap;
}
.back-link {
  color: var(--text-color-muted);
  text-decoration: none;
  font-weight: 500;
}
.back-link:hover {
  color: var(--primary-color);
}
.logo-link {
  text-decoration: none;
}
.logo {
  font-size: var(--font-size-xl);
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.05em;
}
.settings-main {
  max-width: 520px;
  margin: 0 auto;
  padding: var(--spacing-8) var(--spacing-4) max(var(--spacing-12), env(safe-area-inset-bottom, 0px));
  padding-left: max(var(--spacing-4), env(safe-area-inset-left, 0px));
  padding-right: max(var(--spacing-4), env(safe-area-inset-right, 0px));
}
.banner-error {
  padding: var(--spacing-3) var(--spacing-4);
  border-radius: var(--radius-md);
  background: rgba(206, 57, 39, 0.1);
  color: var(--color-vermillion);
  margin-bottom: var(--spacing-4);
  font-size: var(--font-size-sm);
}
.profile-head {
  display: flex;
  align-items: center;
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-6);
  padding: var(--spacing-6);
}
.profile-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.25rem;
  color: #fff;
  flex-shrink: 0;
}
.profile-title {
  margin: 0;
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--text-color);
}
.profile-sub {
  margin: var(--spacing-1) 0 0;
  font-size: var(--font-size-sm);
  color: var(--text-color-muted);
}
.settings-section {
  padding: var(--spacing-6);
  margin-bottom: var(--spacing-6);
}
.settings-section h2 {
  margin: 0 0 var(--spacing-2);
  font-size: var(--font-size-lg);
  font-weight: 700;
  color: var(--text-color);
}
.hint {
  margin: 0 0 var(--spacing-4);
  font-size: var(--font-size-sm);
  color: var(--text-color-muted);
}
.msg {
  margin: 0 0 var(--spacing-3);
  font-size: var(--font-size-sm);
  color: var(--color-vermillion);
  font-weight: 500;
}
.msg.ok {
  color: var(--secondary-color);
}

@media (max-width: 480px) {
  .profile-title {
    font-size: var(--font-size-xl);
  }
  .settings-section {
    padding: var(--spacing-4);
  }
  .profile-head {
    padding: var(--spacing-4);
  }
}
</style>
