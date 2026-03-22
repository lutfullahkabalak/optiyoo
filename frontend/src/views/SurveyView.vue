<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const survey = ref<any>(null)
const isLoading = ref(true)
const answers = ref<Record<string, any>>({})
const submitMessage = ref('')

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/')
    return
  }
  
  try {
    const res = await fetch(`http://localhost:8080/api/surveys/${route.params.id}`)
    if (res.ok) {
      survey.value = await res.json()
      // Initialize model answers
      if (survey.value.questions) {
        survey.value.questions.forEach((q: any) => {
          answers.value[q.id] = null
        })
      }
    }
  } catch(e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
})

const handleSubmit = async () => {
  const payload = {
    user_id: authStore.user?.id,
    answers: Object.entries(answers.value).map(([qId, val]) => ({
      question_id: qId,
      value: String(val || '')
    }))
  }

  try {
    const res = await fetch(`http://localhost:8080/api/surveys/${route.params.id}/answers`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    const data = await res.json()
    if (res.ok) {
      submitMessage.value = data.message || "Anket başarıyla gönerildi!"
      if(authStore.user) {
         authStore.user.points = Number(authStore.user.points) + 5
      }
      setTimeout(() => router.push('/dashboard'), 2500)
    }
  } catch (e) {
     console.error(e)
  }
}
</script>

<template>
  <div class="app-wrapper animate-fade-in" style="background-color: var(--color-white); display: flex; align-items: center; padding: var(--spacing-8) 0; flex-direction:column;">
    <main class="container main-content" style="max-width: 700px; width: 100%;">
        
      <div style="text-align: center; margin-bottom: var(--spacing-8);">
          <div class="logo" style="font-size: var(--font-size-2xl); font-weight: 700; color: var(--primary-color);">optiyoo</div>
      </div>

      <div v-if="isLoading" class="text-center card">Lütfen bekleyin, anket yükleniyor...</div>
      
      <div v-else-if="!survey" class="text-center card">
          Aradığınız anket bulunamadı.
          <br><br>
          <router-link to="/dashboard"><button class="btn btn-outline">Panele Dön</button></router-link>
      </div>

      <div v-else class="card" style="border-top: 8px solid var(--primary-color); padding: var(--spacing-8);">
        
        <div v-if="submitMessage" class="text-center" style="padding: var(--spacing-8) 0;">
            <h2 style="color:var(--secondary-color)">🎉 {{ submitMessage }}</h2>
            <p style="margin-top:var(--spacing-4); color:var(--text-color-muted)">Panele yönlendiriliyorsunuz...</p>
        </div>
        
        <div v-else>
            <h1 class="mb-2">{{ survey.title }}</h1>
            <p class="text-muted mb-8" style="color: var(--text-color-muted);">{{ survey.description }}</p>
            
            <div v-if="!survey.questions || survey.questions.length === 0" class="text-center text-muted">
                Oluşturucu bu ankete henüz soru eklemedi.
            </div>

            <div v-for="(q, index) in survey.questions" :key="q.id" class="form-group mb-8">
              <label class="form-label" style="font-size: var(--font-size-lg);">{{ Number(index) + 1 }}. {{ q.text }}</label>
              
              <!-- Çoktan Seçmeli Tip -->
              <div v-if="q.type !== 'text'" class="flex gap-4 mt-4 answer-grid">
                <label v-for="opt in q.options" :key="opt.id" class="card option-card" :class="{ 'active': answers[q.id] === opt.id }">
                  <input type="radio" :name="'q_'+q.id" :value="opt.id" v-model="answers[q.id]" style="margin-right: 8px;"> {{ opt.text }}
                </label>
              </div>

              <!-- Kısa Yanıt / Açık Uçlu Tip -->
              <input v-else type="text" class="form-control mt-4" placeholder="Kısa kelimelerle ifade edin..." v-model="answers[q.id]" />
            </div>
            
            <div class="flex justify-between mt-8" style="border-top: 1px solid var(--border-color); padding-top: var(--spacing-6);">
              <router-link to="/dashboard" style="text-decoration:none;">
                  <button class="btn btn-outline">İptal</button>
              </router-link>
              <button class="btn btn-primary" style="padding-left: var(--spacing-8); padding-right: var(--spacing-8);" @click="handleSubmit">Cevapla & +5 OPT Kazan</button>
            </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-wrapper {
  min-height: 100vh;
}
.answer-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}
.option-card {
    cursor: pointer;
    display: flex;
    justify-content: flex-start;
    align-items: center;
    border: 2px solid var(--border-color);
    box-shadow: none;
    transition: all 0.2s;
}
.option-card:hover {
    border-color: var(--secondary-color);
}
.option-card.active {
    border-color: var(--primary-color);
    background-color: var(--bg-color);
}
</style>
