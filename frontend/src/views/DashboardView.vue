<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const surveys = ref<any[]>([])
const isLoading = ref(true)

// Local state for interactive feed
const givenAnswers = ref<Record<string, string>>({})
const submittingState = ref<Record<string, boolean>>({})
const completedSurveys = ref<Set<string>>(new Set(JSON.parse(localStorage.getItem('completed_polls_' + authStore.user?.id) || '[]')))

onMounted(async () => {
  if (!authStore.isAuthenticated) {
    router.push('/')
    return
  }
  
  try {
    const res = await fetch(`http://localhost:8080/api/surveys?user_id=${authStore.user?.id}`)
    if (res.ok) {
      surveys.value = await res.json()
      // Initialize completed state directly from reliable Backend data!
      surveys.value.forEach((s: any) => {
          if (s.user_answer) {
              givenAnswers.value[s.questions[0].id] = s.user_answer
              completedSurveys.value.add(s.id)
          }
      })
    }
  } catch(e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
})

const handleLogout = () => {
  authStore.logout()
  router.push('/')
}

const getTotalVotes = (q: any) => {
    if (!q || !q.options) return 0;
    return q.options.reduce((sum: number, o: any) => sum + (o.vote_count || 0), 0);
}

const submitPoll = async (survey: any, val: string) => {
  if (!val || completedSurveys.value.has(survey.id)) return;
  
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
      completedSurveys.value.add(survey.id)
      localStorage.setItem('completed_polls_' + authStore.user?.id, JSON.stringify(Array.from(completedSurveys.value)))
      
      // Update local vote count manually for immediate visual feedback
      const q = survey.questions[0]
      const opt = q.options.find((o: any) => o.id === val)
      if (opt) opt.vote_count = (opt.vote_count || 0) + 1
      // Force reactivity on the option object
      survey.questions[0] = { ...q }

      if(authStore.user) {
         authStore.user.points = Number(authStore.user.points) + 5
      }
    } else {
        const text = await res.text()
        alert("Oylama İptali: " + text)
        if (res.status === 403) {
            completedSurveys.value.add(survey.id)
            localStorage.setItem('completed_polls_' + authStore.user?.id, JSON.stringify(Array.from(completedSurveys.value)))
        }
    }
  } catch (e) {
     console.error(e)
  } finally {
     submittingState.value[survey.id] = false
  }
}
</script>

<template>
  <div class="app-wrapper animate-fade-in" v-if="authStore.user">
    <header class="header">
      <div class="container flex justify-between items-center header-inner">
        <div class="logo">optiyoo</div>
        <nav class="nav">
            <span style="font-weight:600; margin-right: var(--spacing-4);">Ziyaretçi: <strong>{{ authStore.user.name }}</strong></span>
            <button class="btn btn-outline" style="border:none;" @click="handleLogout">Çıkış Yap</button>
        </nav>
      </div>
    </header>

    <main class="container main-content mt-8" style="max-width:900px;">
      <div class="flex justify-between items-center mb-6">
        <h2 style="font-size: var(--font-size-3xl);">🏠 Ana Akış (Feed)</h2>
        <router-link to="/create-survey" style="text-decoration:none;">
          <button class="btn btn-primary">Yeni Anket Oluştur +</button>
        </router-link>
      </div>
      
      <div class="flex gap-8" style="align-items: flex-start;">
        <!-- Sol taraf: Feed -->
        <div style="flex: 2;">
          
          <div v-if="isLoading" class="text-center py-4 text-muted">Akış yenileniyor...</div>
          <div v-else-if="surveys.length === 0" class="card text-center text-muted py-8">Şu an aktif bir oylama bulunmuyor.</div>
          
          <div v-for="s in surveys" :key="s.id" class="card feed-card">
            
            <div v-if="submittingState[s.id]" class="completed-overlay animate-fade-in" style="background: rgba(255,255,255,0.7);">
                 <p style="font-size:18px; font-weight:600; color:var(--primary-color);">Oyunuz İletiliyor...</p>
            </div>

            <div>
                <!-- Anket Başlığı ve Soru -->
                <div class="flex justify-end items-center">
                    <span v-if="!completedSurveys.has(s.id)" class="badge" style="background:var(--bg-color); padding:4px 8px; font-size:12px; border-radius:4px;">Yeni</span>
                </div>
                
                <h3 style="margin: var(--spacing-2) 0; color:var(--text-color);">{{ s.title }}</h3>
                <p class="text-muted" style="margin-bottom: var(--spacing-4);" v-if="s.description">{{ s.description }}</p>

                <!-- Şıklar veya Yüzde Barları -->
                <div v-if="s.questions && s.questions.length > 0" class="options-container">
                    
                    <!-- Tekten Seçmeli Mantığı -->
                    <div v-if="s.questions[0].type === 'single_choice'" style="display: flex; flex-direction:column; gap:8px;">
                        
                        <label v-for="opt in s.questions[0].options" :key="opt.id" class="option-row" :class="{ 'active': givenAnswers[s.questions[0].id] === opt.id, 'disabled': completedSurveys.has(s.id) }">
                            
                            <!-- Progress Bar Background (Sadece tamamlanmışsa görünür) -->
                            <div v-if="completedSurveys.has(s.id)" class="inline-progress-fill" :class="{'selected-fill': givenAnswers[s.questions[0].id] === opt.id}" :style="`width: ${getTotalVotes(s.questions[0]) > 0 ? Math.round((opt.vote_count / getTotalVotes(s.questions[0])) * 100) : 0}%;`"></div>
                            
                            <!-- İçerik -->
                            <div class="option-content">
                                <div style="display: flex; align-items: center;">
                                    <input type="radio" :name="'survey_'+s.id" :value="opt.id" v-model="givenAnswers[s.questions[0].id]" 
                                           @change="submitPoll(s, opt.id)" 
                                           :disabled="completedSurveys.has(s.id)" 
                                           style="display: none;" />
                                    <!-- Custom Radio Indicator for aesthetics -->
                                    <div class="radio-indicator" :class="{'checked': givenAnswers[s.questions[0].id] === opt.id}"></div>
                                    <span style="font-weight: 500;">{{ opt.text }}</span>
                                </div>
                                
                                <!-- Yüzdelik Metin (Sadece tamamlanmışsa) -->
                                <span v-if="completedSurveys.has(s.id)" style="font-weight: 700; font-size: 14px; text-align:right;">
                                    {{ getTotalVotes(s.questions[0]) > 0 ? Math.round((opt.vote_count / getTotalVotes(s.questions[0])) * 100) : 0 }}%
                                </span>
                            </div>
                        </label>
                        
                        <!-- Toplam Oy Sayısı -->
                        <p v-if="completedSurveys.has(s.id)" style="font-size:12px; font-weight: 500; color:var(--text-color-muted); text-align:right; margin-top:4px;">
                           Toplam {{ getTotalVotes(s.questions[0]) }} Oy
                        </p>

                    </div>
                </div>
                
            </div>
          </div>

        </div>

        <!-- Sağ taraf: Profil/Cüzdan Widget -->
        <div class="card" style="flex: 1; position: sticky; top: 20px;">
          <h3 class="mb-4">Profil Cüzdanı</h3>
          <p style="font-size: var(--font-size-sm); color: var(--text-color-muted);">Güncel Bakiyeniz:</p>
          <div style="font-size: 32px; font-weight:700; color:var(--primary-color); margin-top:8px;">{{ authStore.user.points }} OPT</div>
          <button class="btn btn-outline mt-6" style="width: 100%; border: 2px solid var(--border-color);">Ödülleri Keşfet</button>
        </div>
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

/* Feed Styling */
.feed-card {
    border: 1px solid var(--border-color);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
    margin-bottom: var(--spacing-6);
    position: relative;
    overflow: hidden;
    padding: var(--spacing-6);
}

.option-row {
    position: relative;
    padding: 0;
    border: 1px solid var(--border-color);
    border-radius: var(--radius-md);
    cursor: pointer;
    overflow: hidden;
    transition: all 0.2s ease;
}
.option-row:not(.disabled):hover {
    background-color: var(--bg-color);
    border-color: var(--secondary-color);
}
.option-row.active {
    border-color: var(--primary-color);
}
.option-row.disabled {
    cursor: default;
    background-color: transparent !important;
}

.option-content {
    position: relative;
    z-index: 2;
    padding: var(--spacing-3) var(--spacing-4);
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.inline-progress-fill {
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    background-color: var(--border-color);
    opacity: 0.6;
    z-index: 1;
    transition: width 0.8s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.inline-progress-fill.selected-fill {
    background-color: var(--color-pale-beige);
    opacity: 1;
}

/* Görsel Özgünleştirme: Yeni Radyo Butonu */
.radio-indicator {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    border: 2px solid var(--border-color);
    margin-right: 12px;
    transition: all 0.2s;
    background: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}
.radio-indicator.checked {
    border-color: var(--primary-color);
}
.radio-indicator.checked::after {
    content: '';
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background-color: var(--primary-color);
}

.completed-overlay {
    position: absolute;
    top:0; left:0; right:0; bottom:0;
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 10;
    backdrop-filter: blur(2px);
}
</style>
