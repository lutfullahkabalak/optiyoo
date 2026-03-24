<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const emit = defineEmits(['close', 'created'])
const authStore = useAuthStore()

const survey = ref({
  creator_id: '',
  questions: [
    { type: 'single_choice', text: '', order: 1, options: [{ text: '' }, { text: '' }] }
  ] as any[]
})

const isSubmitting = ref(false)

onMounted(() => {
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
      // Başarılı
      emit('created')
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
  <div class="modal-overlay animate-fade-in" @click.self="emit('close')">
    <div class="modal-content card" style="border-radius: 20px; padding: 32px;">
      <div class="flex justify-between items-center mb-4">
        <h2 style="margin: 0; font-size: 22px; font-weight: 700;">Yeni Bir Soru Sor</h2>
        <button class="btn btn-outline btn-close" @click="emit('close')">✕</button>
      </div>

      <p class="text-muted mb-8" style="font-size: 15px; line-height: 1.5;">Katılımcılara yöneltmek istediğiniz tek soruluk mini-anketinizi (poll) tasarlayın.</p>

      <div v-for="(q, qIndex) in survey.questions" :key="qIndex">
        
        <label style="font-weight: 600; font-size: 14px; margin-bottom: 8px; display: block; color: var(--text-color);">Soru Metni</label>
        <textarea 
          class="form-control mb-6" 
          v-model="q.text" 
          placeholder="Katılımcılara tam olarak ne sormak istiyorsunuz?" 
          required 
          rows="2"
          style="font-size: 16px; padding: 14px; resize: none;"
        ></textarea>

        <!-- Seçenekler -->
        <div v-if="q.type === 'single_choice'">
          <label style="font-weight: 600; font-size: 14px; margin-bottom: 12px; display: block; color: var(--text-color);">Seçenekler</label>
          
          <div style="display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px;">
            <div v-for="(opt, oIndex) in q.options" :key="oIndex" class="flex items-center gap-3">
              <input 
                type="text" 
                class="form-control" 
                style="padding: 12px 14px; flex: 1;" 
                v-model="opt.text" 
                :placeholder="`Seçenek ${Number(oIndex) + 1}`" 
                required 
              />
              <button 
                class="btn btn-outline" 
                style="padding: 8px 12px; border: none; color: var(--text-color-muted); font-weight: bold; background: transparent;" 
                @click="removeOption(Number(qIndex), Number(oIndex))" 
                v-if="q.options.length > 2"
                title="Kaldır"
              >✕</button>
            </div>
          </div>
          
          <button class="btn btn-outline" style="font-size: 14px; padding: 8px 16px; border-style: dashed;" @click="addOption(Number(qIndex))">
            + Yeni Seçenek Ekle
          </button>
        </div>
      </div>

      <div class="flex justify-center mt-8 pt-8" style="border-top:1px solid var(--border-color);">
        <button class="btn btn-primary" style="width: 100%; font-size: 16px; padding: 14px;" @click="submitSurvey" :disabled="isSubmitting">
          {{ isSubmitting ? 'Yayınlanıyor...' : 'Hemen Yayınla' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
}

.modal-content {
  background: var(--color-white);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
  border-radius: var(--radius-lg);
  padding: var(--spacing-6);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
}

.btn-close {
  padding: 4px 10px;
  border: none;
  background: var(--bg-color);
  font-weight: bold;
}

.btn-close:hover {
  background: #e2e8f0;
}


</style>
