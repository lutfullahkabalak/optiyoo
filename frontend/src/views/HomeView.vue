<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const isLoginMode = ref(true)
const email = ref('')
const password = ref('')
const name = ref('')
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
    payload.name = name.value
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
    router.push('/dashboard')
    
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
        <div class="logo">optiyoo</div>
        <nav class="nav">
          <button class="btn btn-outline" style="margin-right: 8px;" @click="isLoginMode = true">Giriş Yap</button>
          <button class="btn btn-primary" @click="isLoginMode = false">Kaydol</button>
        </nav>
      </div>
    </header>

    <main class="container main-content mt-8 flex justify-center items-center flex-col">
      <h1 class="text-center" style="font-size: var(--font-size-4xl); margin-bottom: var(--spacing-2);">Anket Oluştur & Çöz</h1>
      <p class="text-center text-muted mb-8" style="font-size: var(--font-size-lg); color: var(--text-color-muted); max-width: 600px;">
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
          <div class="form-group">
            <label class="form-label">E-posta</label>
            <input type="email" v-model="email" class="form-control" placeholder="ornek@mail.com" required />
          </div>
          <div class="form-group">
            <label class="form-label">Şifre</label>
            <input type="password" v-model="password" class="form-control" placeholder="••••••••" required />
          </div>
          <button class="btn btn-primary" :disabled="isLoading" style="width: 100%; margin-top: var(--spacing-2);">
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
}
.logo {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.05em;
}
.main-content {
  flex: 1;
}
</style>
