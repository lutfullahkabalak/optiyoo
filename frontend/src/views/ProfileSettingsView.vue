<script setup lang="ts">
import { defineAsyncComponent, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

const AvatarCropModal = defineAsyncComponent(() => import('../components/AvatarCropModal.vue'))
import { useAuthStore } from '../stores/auth'
import {
  absoluteApiUrl,
  DEFAULT_AVATAR_COLORS,
  hashAvatarColor,
  resolveAvatarBackground,
  usernameInitial
} from '../utils/avatar'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const router = useRouter()
const authStore = useAuthStore()

const loadError = ref('')
const displayNameMsg = ref('')
const usernameMsg = ref('')
const emailMsg = ref('')
const passwordMsg = ref('')
const avatarMsg = ref('')

const newDisplayName = ref('')

const newUsername = ref('')

const newEmail = ref('')

const passwordCurrentPw = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const busyDisplayName = ref(false)
const busyUser = ref(false)
const busyEmail = ref(false)
const busyPassword = ref(false)
const busyAvatar = ref(false)

const selectedAvatarColor = ref('#6366f1')
const cropOpen = ref(false)
const cropFile = ref<File | null>(null)
const avatarFileInput = ref<HTMLInputElement | null>(null)

const syncFormDefaults = () => {
  const u = authStore.user
  if (!u) return
  newDisplayName.value = u.name || ''
  newUsername.value = u.username || ''
  newEmail.value = u.email || ''
  if (u.avatar_color && /^#[0-9A-Fa-f]{6}$/i.test(u.avatar_color)) {
    selectedAvatarColor.value = u.avatar_color
  } else {
    selectedAvatarColor.value = hashAvatarColor(u.id)
  }
}

const refreshProfile = async () => {
  const uid = authStore.user?.id
  if (!uid) return false
  const res = await fetch(`${API_BASE}/api/users/${encodeURIComponent(uid)}`, {
    headers: authStore.authHeadersGet()
  })
  if (!res.ok) return false
  const data = await res.json()
  authStore.setUser({ ...authStore.user!, ...data })
  syncFormDefaults()
  return true
}

onMounted(async () => {
  if (!authStore.isAuthenticated || !authStore.user?.id) {
    router.push('/auth')
    return
  }
  syncFormDefaults()
  try {
    const res = await fetch(`${API_BASE}/api/users/${encodeURIComponent(authStore.user.id)}`, {
      headers: authStore.authHeadersGet()
    })
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
  body: Record<string, unknown>,
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
      headers: authStore.authHeadersJson(),
      body: JSON.stringify(body)
    })
    const text = await res.text()
    if (!res.ok) {
      setMsg(text || 'İşlem başarısız.')
      return
    }
    const data = JSON.parse(text)
    authStore.setUser(data)
    syncFormDefaults()
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
  patchUser({ name: n }, (v) => (busyDisplayName.value = v), (m) => (displayNameMsg.value = m))
}

const saveUsername = () => {
  const u = newUsername.value.trim()
  if (!u) {
    usernameMsg.value = 'Kullanıcı adı gerekli.'
    return
  }
  patchUser({ username: u }, (v) => (busyUser.value = v), (m) => (usernameMsg.value = m))
}

const saveEmail = () => {
  const e = newEmail.value.trim()
  if (!e) {
    emailMsg.value = 'E-posta gerekli.'
    return
  }
  patchUser({ email: e }, (v) => (busyEmail.value = v), (m) => (emailMsg.value = m))
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

const openAvatarPicker = () => {
  avatarFileInput.value?.click()
}

const onAvatarFileChange = (e: Event) => {
  const t = e.target as HTMLInputElement
  const f = t.files?.[0]
  t.value = ''
  if (!f) return
  cropFile.value = f
  cropOpen.value = true
}

const onAvatarCropped = async (blob: Blob) => {
  avatarMsg.value = ''
  busyAvatar.value = true
  try {
    const fd = new FormData()
    fd.append('file', blob, 'avatar.jpg')
    const c = selectedAvatarColor.value.trim()
    if (c && /^#[0-9A-Fa-f]{6}$/i.test(c)) {
      fd.append('avatar_color', c)
    }
    const res = await fetch(`${API_BASE}/api/user-media`, {
      method: 'POST',
      headers: authStore.authHeadersMultipart(),
      body: fd
    })
    const text = await res.text()
    if (!res.ok) {
      avatarMsg.value = text || 'Yükleme başarısız.'
      return
    }
    await refreshProfile()
    avatarMsg.value = 'Profil resmi güncellendi.'
  } catch (e) {
    avatarMsg.value = 'Bağlantı hatası.'
    console.error(e)
  } finally {
    busyAvatar.value = false
  }
}

const saveAvatarColor = () => {
  const c = selectedAvatarColor.value.trim()
  if (c && !/^#[0-9A-Fa-f]{6}$/i.test(c)) {
    avatarMsg.value = 'Geçerli bir renk seçin (#RRGGBB).'
    return
  }
  patchUser({ avatar_color: c }, (v) => (busyAvatar.value = v), (m) => (avatarMsg.value = m))
}

const clearStoredAvatarColor = () => {
  patchUser({ avatar_color: '' }, (v) => (busyAvatar.value = v), (m) => (avatarMsg.value = m))
}

const removeAvatarPhoto = () => {
  if (!authStore.user?.avatar_url) {
    avatarMsg.value = 'Yüklü bir fotoğraf yok.'
    return
  }
  patchUser({ remove_avatar: true }, (v) => (busyAvatar.value = v), (m) => (avatarMsg.value = m))
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

      <AvatarCropModal v-model="cropOpen" :file="cropFile" @apply="onAvatarCropped" />

      <div class="profile-head card">
        <div
          class="profile-avatar"
          :class="{ 'profile-avatar--photo': !!authStore.user.avatar_url }"
          :style="
            authStore.user.avatar_url
              ? undefined
              : { background: resolveAvatarBackground(authStore.user.avatar_color, authStore.user.id) }
          "
        >
          <img
            v-if="authStore.user.avatar_url"
            class="profile-avatar-img"
            :src="absoluteApiUrl(API_BASE, authStore.user.avatar_url)"
            alt=""
          />
          <template v-else>{{ usernameInitial(authStore.user.username) }}</template>
        </div>
        <div>
          <h1 class="profile-title">Profil ayarları</h1>
          <p class="profile-sub">{{ authStore.user.name }}</p>
        </div>
      </div>

      <section class="card settings-section">
        <h2>Profil fotoğrafı ve renk</h2>
        <p class="hint">
          Fotoğraf seçtiğinizde daire içinde kırpma penceresi açılır. Fotoğraf yoksa kullanıcı adınızın baş harfi
          gösterilir. Arka plan rengini aşağıdan seçebilir veya kayıtlı rengi sıfırlayıp otomatik palete dönebilirsiniz.
        </p>
        <div class="avatar-preview-row">
          <div
            class="avatar-preview-circle"
            :class="{ 'avatar-preview-circle--photo': !!authStore.user.avatar_url }"
            :style="
              authStore.user.avatar_url
                ? undefined
                : { background: resolveAvatarBackground(selectedAvatarColor, authStore.user.id) }
            "
          >
            <img
              v-if="authStore.user.avatar_url"
              class="avatar-preview-img"
              :src="absoluteApiUrl(API_BASE, authStore.user.avatar_url)"
              alt=""
            />
            <span v-else class="avatar-preview-letter">{{ usernameInitial(authStore.user.username) }}</span>
          </div>
          <div class="avatar-preview-actions">
            <input
              ref="avatarFileInput"
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              class="visually-hidden"
              @change="onAvatarFileChange"
            />
            <button type="button" class="btn btn-primary" :disabled="busyAvatar" @click="openAvatarPicker">
              {{ busyAvatar ? 'İşleniyor…' : 'Resim seç' }}
            </button>
            <button
              v-if="authStore.user.avatar_url"
              type="button"
              class="btn btn-outline"
              :disabled="busyAvatar"
              @click="removeAvatarPhoto"
            >
              Fotoğrafı kaldır
            </button>
          </div>
        </div>
        <p class="hint label-tight">Arka plan rengi (harf gösteriminde)</p>
        <div class="avatar-swatches" role="list">
          <button
            v-for="col in DEFAULT_AVATAR_COLORS"
            :key="col"
            type="button"
            class="swatch-btn"
            :style="{ background: col }"
            :title="col"
            :aria-pressed="selectedAvatarColor.toLowerCase() === col.toLowerCase()"
            @click="selectedAvatarColor = col"
          />
        </div>
        <div class="form-group">
          <label class="form-label">Özel renk</label>
          <input v-model="selectedAvatarColor" type="color" class="form-control color-native" />
        </div>
        <div class="avatar-save-row">
          <button type="button" class="btn btn-primary" :disabled="busyAvatar" @click="saveAvatarColor">
            {{ busyAvatar ? 'Kaydediliyor…' : 'Rengi kaydet' }}
          </button>
          <button
            v-if="authStore.user.avatar_color"
            type="button"
            class="btn btn-outline"
            :disabled="busyAvatar"
            @click="clearStoredAvatarColor"
          >
            Kayıtlı rengi sıfırla
          </button>
        </div>
        <p v-if="avatarMsg" class="msg" :class="{ ok: avatarMsg.includes('güncellendi') || avatarMsg === 'Kaydedildi.' }">
          {{ avatarMsg }}
        </p>
      </section>

      <section class="card settings-section">
        <h2>Görünen ad</h2>
        <p class="hint">Profil ve kartlarda görünen adınız (ad soyad).</p>
        <div class="form-group">
          <label class="form-label">Görünen ad</label>
          <input v-model="newDisplayName" type="text" class="form-control" autocomplete="name" maxlength="255" />
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
  overflow: hidden;
}
.profile-avatar--photo {
  background: var(--border-color, #e5e7eb);
}
.profile-avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
.avatar-preview-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-4);
  margin-bottom: var(--spacing-4);
}
.avatar-preview-circle {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 1.35rem;
  color: #fff;
  flex-shrink: 0;
  overflow: hidden;
  box-shadow: inset 0 0 0 2px rgba(255, 255, 255, 0.5);
}
.avatar-preview-circle--photo {
  background: var(--border-color, #e5e7eb);
}
.avatar-preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.avatar-preview-letter {
  line-height: 1;
}
.avatar-preview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
}
.label-tight {
  margin-bottom: var(--spacing-2);
}
.avatar-swatches {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-2);
  margin-bottom: var(--spacing-4);
}
.swatch-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid transparent;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
}
.swatch-btn:hover {
  transform: scale(1.06);
}
.swatch-btn[aria-pressed='true'] {
  border-color: var(--text-color);
  box-shadow: 0 0 0 2px var(--color-white), 0 0 0 4px var(--primary-color);
}
.color-native {
  height: 44px;
  padding: var(--spacing-1);
  cursor: pointer;
}
.avatar-save-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-3);
  margin-top: var(--spacing-2);
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
