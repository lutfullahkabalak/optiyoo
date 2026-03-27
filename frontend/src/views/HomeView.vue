<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const isLoginMode = ref(true)
const email = ref('test3@test.com')
const password = ref('123')
const name = ref('')
const username = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

const handleAuth = async () => {
  errorMessage.value = ''
  isLoading.value = true
  
  const endpoint = isLoginMode.value ? '/api/login' : '/api/register'
  const payload: any = { email: email.value, password: password.value }
  
  if (!isLoginMode.value) {
    if(!name.value) {
        errorMessage.value = 'Lütfen adınızı girin.'
        isLoading.value = false
        return
    }
    if (!username.value.trim()) {
        errorMessage.value = 'Lütfen kullanıcı adı girin.'
        isLoading.value = false
        return
    }
    payload.name = name.value
    payload.username = username.value.trim()
  }

  try {
    const res = await fetch(`http://localhost:8080${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    
    if (!res.ok) {
       const text = await res.text()
       throw new Error(text || 'Kullanıcı adı veya şifre hatalı.')
    }
    
    const data = await res.json()
    authStore.setUser(data)
    router.push('/')
    
  } catch (e: any) {
    errorMessage.value = e.message
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="app-wrapper animate-fade-in">
    <header class="header">
      <div class="container flex justify-between items-center header-inner">
        <router-link to="/" class="logo-link">
          <span class="logo">optiyoo</span>
        </router-link>
        <nav class="auth-nav">
          <button type="button" class="btn btn-outline auth-nav-btn" @click="isLoginMode = true">Giriş Yap</button>
          <button type="button" class="btn btn-primary auth-nav-btn" @click="isLoginMode = false">Kaydol</button>
        </nav>
      </div>
    </header>

    <main class="container main-content mt-8 flex justify-center items-center flex-col">
      <h1 class="text-center hero-title">Anket Oluştur & Çöz</h1>
      <p class="text-center text-muted hero-lead mb-8">
        Fikirlerinizi paylaşın, başkalarının fikirlerini öğrenin. Hemen kayıt olarak anket oluşturmaya başlayın.
      </p>

      <div class="card" style="width: 100%; max-width: 400px;">
        <h2 class="text-center mb-4">{{ isLoginMode ? 'Giriş Yap' : 'Kayıt Ol' }}</h2>
        
        <div v-if="errorMessage" class="mb-4 text-center" style="color: var(--color-vermillion); font-weight: 500;">
          {{ errorMessage }}
        </div>

        <form @submit.prevent="handleAuth">
          <div class="form-group" v-if="!isLoginMode">
            <label class="form-label">Ad Soyad</label>
            <input type="text" v-model="name" class="form-control" placeholder="Adınız" />
          </div>
          <div class="form-group" v-if="!isLoginMode">
            <label class="form-label">Kullanıcı adı</label>
            <input type="text" v-model="username" class="form-control" placeholder="benzersiz_kullanici_adi" autocomplete="username" />
          </div>
          <div class="form-group">
            <label class="form-label">E-posta</label>
            <input type="email" v-model="email" class="form-control" placeholder="ornek@mail.com" required />
          </div>
          <div class="form-group">
            <label class="form-label">Şifre</label>
            <input type="password" v-model="password" class="form-control" placeholder="••••••••" required />
          </div>
          <button class="btn btn-primary" :disabled="isLoading" style="width: 100%; margin-top: var(--spacing-4);">
            {{ isLoading ? 'İşleniyor...' : (isLoginMode ? 'Giriş Yap' : 'Kayıt Ol') }}
          </button>
        </form>
        
        <p class="text-center mt-4" style="font-size: var(--font-size-sm); color: var(--text-color-muted); cursor: pointer; user-select: none;" @click="isLoginMode = !isLoginMode">
          <u>{{ isLoginMode ? 'Hesabınız yok mu? Kayıt Olun' : 'Zaten hesabınız var mı? Giriş Yapın' }}</u>
        </p>
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
  flex-wrap: wrap;
  gap: var(--spacing-3);
}
.logo-link {
  text-decoration: none;
}
.logo {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.05em;
}
.auth-nav {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-2);
  justify-content: flex-end;
}
.auth-nav-btn {
  font-size: var(--font-size-sm);
}
.main-content {
  flex: 1;
  padding-left: max(0px, env(safe-area-inset-left, 0px));
  padding-right: max(0px, env(safe-area-inset-right, 0px));
  padding-bottom: max(var(--spacing-8), env(safe-area-inset-bottom, 0px));
}
.hero-title {
  font-size: clamp(1.5rem, 6vw, var(--font-size-4xl));
  margin-bottom: var(--spacing-2);
  line-height: 1.15;
  padding: 0 var(--spacing-2);
}
.hero-lead {
  font-size: clamp(var(--font-size-base), 3.5vw, var(--font-size-lg));
  color: var(--text-color-muted);
  max-width: 600px;
  padding: 0 var(--spacing-2);
  text-wrap: balance;
}
@media (max-width: 480px) {
  .auth-nav {
    width: 100%;
    justify-content: stretch;
  }
  .auth-nav-btn {
    flex: 1;
    min-height: 44px;
  }
}
</style>
