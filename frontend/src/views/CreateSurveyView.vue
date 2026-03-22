<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const survey = ref({
  title: '',
  description: '',
  creator_id: '',
  questions: [
    { type: 'single_choice', text: '', order: 1, options: [{ text: '' }, { text: '' }] }
  ] as any[]
})

const isSubmitting = ref(false)

onMounted(() => {
  if (!authStore.isAuthenticated) {
    router.push('/')
    return
  }
  survey.value.creator_id = authStore.user?.id || ''
})

const addOption = (qIndex: number) => {
  survey.value.questions[qIndex].options.push({ text: '' })
}

const removeOption = (qIndex: number, oIndex: number) => {
  if (survey.value.questions[qIndex].options.length > 2) {
      survey.value.questions[qIndex].options.splice(oIndex, 1)
  }
}

const submitSurvey = async () => {
  if (!survey.value.title) return alert("Anket başlığı giriniz.")
  if (survey.value.questions.length !== 1) return alert("Sistemde anketler sadece 1 soru içerebilir.")
  if (!survey.value.questions[0].text) return alert("Lütfen sorunuzun metnini giriniz.")
  
  isSubmitting.value = true
  try {
    const res = await fetch('http://localhost:8080/api/surveys', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(survey.value)
    })
    
    if (res.ok) {
      router.push('/dashboard')
    } else {
      const text = await res.text()
      alert("Hata: " + text)
    }
  } catch (e) {
    alert("Kayıt sırasında bir hata oluştu")
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="app-wrapper animate-fade-in" style="background-color: var(--color-white);">
    <header class="header">
      <div class="container flex justify-between items-center header-inner">
        <div class="logo">optiyoo</div>
        <router-link to="/dashboard"><button class="btn btn-outline" style="border:none;">İptal</button></router-link>
      </div>
    </header>

    <main class="container main-content mt-8" style="max-width: 800px;">
      <h2>Yeni Bir Soru Sor</h2>
      <p class="text-muted mb-8">Katılımcılara yöneltmek istediğiniz tek soruluk mini-anketinizi (poll) tasarlayın.</p>

      <div class="card mb-8" style="padding-bottom: var(--spacing-4);">
        <div class="form-group">
          <label class="form-label">Anket Konusu / Başlığı</label>
          <input type="text" class="form-control" v-model="survey.title" placeholder="Örn: Hafta Sonu Planı" required />
        </div>
        <div class="form-group">
          <label class="form-label">Açıklama (Opsiyonel)</label>
          <input type="text" class="form-control" v-model="survey.description" placeholder="Katılımcıların bilmesi gereken ekstra detaylar var mı?" />
        </div>
      </div>

      <h3 class="mb-4">Oylama Sorunuz</h3>
      <div v-for="(q, qIndex) in survey.questions" :key="qIndex" class="card mb-8" style="border-left: 4px solid var(--primary-color);">
        <input type="text" class="form-control mb-4" v-model="q.text" placeholder="Katılımcılara tam olarak ne sormak istiyorsunuz?..." required />

        <!-- Seçenekler -->
        <div v-if="q.type === 'single_choice'">
          <div v-for="(opt, oIndex) in q.options" :key="oIndex" class="flex items-center gap-4 mb-2">
            <input type="radio" disabled />
            <input type="text" class="form-control" style="padding: var(--spacing-2); flex:1;" v-model="opt.text" :placeholder="`Seçenek ${Number(oIndex) + 1}`" required />
            <button class="btn btn-outline" style="padding: 4px 8px; border:none;" @click="removeOption(Number(qIndex), Number(oIndex))" v-if="q.options.length > 2">x</button>
          </div>
          <button class="btn btn-outline mt-2" style="font-size: 12px; padding: 4px 12px;" @click="addOption(Number(qIndex))">+ Yeni Seçenek Ekle</button>
        </div>
      </div>

      <div class="flex justify-center mt-8 pt-8" style="border-top:1px solid var(--border-color);">
        <button class="btn btn-primary" style="font-size: var(--font-size-lg); padding: var(--spacing-3) var(--spacing-12);" @click="submitSurvey" :disabled="isSubmitting">{{ isSubmitting ? 'Yayınlanıyor...' : 'Hemen Yayınla' }}</button>
      </div>
      <br><br>
    </main>
  </div>
</template>

<style scoped>
.app-wrapper {
  min-height: 100vh;
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
</style>
