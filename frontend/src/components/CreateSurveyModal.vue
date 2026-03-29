<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const emit = defineEmits(['close', 'created'])
const authStore = useAuthStore()

const newOption = () => ({ text: '', imageFile: null as File | null, _previewUrl: '' as string })
const newQuestion = () => ({
  type: 'single_choice',
  text: '',
  order: 1,
  imageFile: null as File | null,
  _previewUrl: '' as string,
  options: [newOption(), newOption()]
})

const survey = ref({
  creator_id: '',
  questions: [newQuestion()] as any[]
})

const isSubmitting = ref(false)

onMounted(() => {
  survey.value.creator_id = authStore.user?.id || ''
})

const canCreateMultiQuestionSurveys = () =>
  Boolean(authStore.user?.can_create_multi_question_surveys)

const revokePreview = (obj: { _previewUrl?: string }) => {
  if (obj._previewUrl) URL.revokeObjectURL(obj._previewUrl)
  obj._previewUrl = ''
}

const onQuestionImage = (qIndex: number, e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  const q = survey.value.questions[qIndex]
  revokePreview(q)
  q.imageFile = file
  q._previewUrl = file ? URL.createObjectURL(file) : ''
  input.value = ''
}

const onOptionImage = (qIndex: number, oIndex: number, e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  const opt = survey.value.questions[qIndex].options[oIndex]
  revokePreview(opt)
  opt.imageFile = file
  opt._previewUrl = file ? URL.createObjectURL(file) : ''
  input.value = ''
}

const clearQuestionImage = (qIndex: number) => {
  const q = survey.value.questions[qIndex]
  revokePreview(q)
  q.imageFile = null
}

const clearOptionImage = (qIndex: number, oIndex: number) => {
  const opt = survey.value.questions[qIndex].options[oIndex]
  revokePreview(opt)
  opt.imageFile = null
}

const addOption = (qIndex: number) => {
  survey.value.questions[qIndex].options.push(newOption())
}

const removeOption = (qIndex: number, oIndex: number) => {
  if (survey.value.questions[qIndex].options.length > 2) {
    const opt = survey.value.questions[qIndex].options[oIndex]
    revokePreview(opt)
    survey.value.questions[qIndex].options.splice(oIndex, 1)
  }
}

const addQuestion = () => {
  if (!canCreateMultiQuestionSurveys()) return
  const nq = newQuestion()
  nq.order = survey.value.questions.length + 1
  survey.value.questions.push(nq)
}

const removeQuestion = (qIndex: number) => {
  if (!canCreateMultiQuestionSurveys()) return
  if (survey.value.questions.length <= 1) return
  const removed = survey.value.questions[qIndex]
  revokePreview(removed)
  ;(removed.options || []).forEach((o: any) => revokePreview(o))
  survey.value.questions.splice(qIndex, 1)
  survey.value.questions.forEach((question, index) => {
    question.order = index + 1
  })
}

const surveyPayloadJSON = () => ({
  creator_id: survey.value.creator_id,
  questions: survey.value.questions.map((q: any) => ({
    type: q.type,
    text: q.text,
    order: q.order,
    options: (q.options || []).map((o: any) => ({ text: o.text }))
  }))
})

function remoteQuestionFor(localQ: any, created: any, index: number) {
  const list = created.questions as any[]
  if (!Array.isArray(list)) return null
  const byOrder = list.find((rq: any) => Number(rq.order) === Number(localQ.order))
  return byOrder ?? list[index] ?? null
}

async function uploadPendingMedia(created: any): Promise<{ failed: number; firstError: string }> {
  const uid = authStore.user?.id
  if (!uid || !created?.id || !Array.isArray(created.questions)) {
    return { failed: 0, firstError: '' }
  }
  let failed = 0
  let firstError = ''
  const localQs = survey.value.questions
  for (let qi = 0; qi < localQs.length; qi++) {
    const localQ = localQs[qi]
    const remoteQ = remoteQuestionFor(localQ, created, qi)
    if (!remoteQ?.id) {
      if (localQ.imageFile) {
        failed++
        const msg = 'Sunucu yanıtında soru id eşleşmedi; görseller atlandı.'
        if (!firstError) firstError = msg
        console.error('[CreateSurveyModal]', msg, { order: localQ.order, created })
      }
      continue
    }
    if (localQ.imageFile) {
      const fd = new FormData()
      fd.append('survey_id', created.id)
      fd.append('kind', 'question')
      fd.append('ref_id', String(remoteQ.id))
      fd.append('file', localQ.imageFile)
      const r = await fetch(`${API_BASE}/api/media`, {
        method: 'POST',
        headers: authStore.authHeadersMultipart(),
        body: fd
      })
      if (!r.ok) {
        failed++
        const t = await r.text()
        if (!firstError) firstError = t || `${r.status} soru görseli`
        console.error('[CreateSurveyModal] POST /api/media question', r.status, t)
      }
    }
    const opts = localQ.options || []
    const rOpts = remoteQ.options || []
    for (let oi = 0; oi < opts.length; oi++) {
      const localO = opts[oi]
      const remoteO = rOpts[oi]
      if (!remoteO?.id || !localO?.imageFile) continue
      const fd = new FormData()
      fd.append('survey_id', created.id)
      fd.append('kind', 'option')
      fd.append('ref_id', String(remoteO.id))
      fd.append('file', localO.imageFile)
      const r = await fetch(`${API_BASE}/api/media`, {
        method: 'POST',
        headers: authStore.authHeadersMultipart(),
        body: fd
      })
      if (!r.ok) {
        failed++
        const t = await r.text()
        if (!firstError) firstError = t || `${r.status} seçenek görseli`
        console.error('[CreateSurveyModal] POST /api/media option', r.status, t)
      }
    }
  }
  return { failed, firstError }
}

const submitSurvey = async () => {
  if (!canCreateMultiQuestionSurveys() && survey.value.questions.length !== 1) {
    return alert("Bu hesap için sadece tek sorulu anket oluşturabilirsiniz.")
  }
  if (survey.value.questions.length < 1) return alert("Anket en az bir soru içermelidir.")

  for (let i = 0; i < survey.value.questions.length; i++) {
    const q = survey.value.questions[i]
    if (!q.text?.trim()) return alert(`${i + 1}. sorunun metnini giriniz.`)
    if (!Array.isArray(q.options) || q.options.length < 2) {
      return alert(`${i + 1}. soru için en az 2 seçenek giriniz.`)
    }
    const hasEmptyOption = q.options.some((opt: any) => !opt.text?.trim())
    if (hasEmptyOption) return alert(`${i + 1}. sorunun tüm seçeneklerini doldurunuz.`)
    q.order = i + 1
  }

  survey.value.creator_id = authStore.user?.id || ''

  isSubmitting.value = true
  try {
    const res = await fetch(`${API_BASE}/api/surveys`, {
      method: 'POST',
      headers: authStore.authHeadersJson(),
      body: JSON.stringify(surveyPayloadJSON())
    })

    if (res.ok) {
      const created = await res.json()
      const { failed, firstError } = await uploadPendingMedia(created)
      if (failed > 0) {
        alert(
          `Anket oluşturuldu fakat ${failed} görsel yüklenemedi.${firstError ? '\n\n' + firstError : ''}`
        )
      }
      emit('created')
    } else {
      const text = await res.text()
      alert('Hata: ' + text)
    }
  } catch (e) {
    alert('Kayıt sırasında bir hata oluştu')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="modal-overlay animate-fade-in" @click.self="emit('close')">
    <div class="modal-content card">
      <div class="flex justify-between items-center mb-4">
        <h2 style="margin: 0; font-size: 22px; font-weight: 700;">Yeni Bir Soru Sor</h2>
        <button class="btn btn-outline btn-close" @click="emit('close')">✕</button>
      </div>

      <p class="text-muted mb-8" style="font-size: 15px; line-height: 1.5;">
        {{ canCreateMultiQuestionSurveys()
          ? 'Katılımcılara yöneltmek istediğiniz anket sorularını oluşturun.'
          : 'Katılımcılara yöneltmek istediğiniz tek soruluk mini-anketinizi (poll) tasarlayın.' }}
      </p>

      <div v-for="(q, qIndex) in survey.questions" :key="qIndex" style="margin-bottom: 18px;">
        <div class="flex justify-between items-center" style="margin-bottom: 8px;">
          <label style="font-weight: 600; font-size: 14px; color: var(--text-color);">Soru {{ Number(qIndex) + 1 }}</label>
          <button
            v-if="canCreateMultiQuestionSurveys() && survey.questions.length > 1"
            class="btn btn-outline"
            style="padding: 4px 10px; border: none; color: var(--text-color-muted); background: transparent;"
            @click="removeQuestion(Number(qIndex))"
            type="button"
          >
            Soruyu Kaldır
          </button>
        </div>
        
        <label style="font-weight: 600; font-size: 14px; margin-bottom: 8px; display: block; color: var(--text-color);">Soru Metni</label>
        <textarea 
          class="form-control mb-6" 
          v-model="q.text" 
          placeholder="Katılımcılara tam olarak ne sormak istiyorsunuz?" 
          required 
          rows="2"
          style="font-size: 16px; padding: 14px; resize: none;"
        ></textarea>

        <div class="media-field">
          <span class="media-label">Soru görseli (isteğe bağlı)</span>
          <div class="media-row">
            <label class="btn btn-outline media-file-btn">
              Dosya seç
              <input
                type="file"
                accept="image/*"
                class="sr-only"
                @change="onQuestionImage(Number(qIndex), $event)"
              />
            </label>
            <button
              v-if="q.imageFile || q._previewUrl"
              type="button"
              class="btn btn-outline media-clear"
              @click="clearQuestionImage(Number(qIndex))"
            >
              Kaldır
            </button>
          </div>
          <img v-if="q._previewUrl" :src="q._previewUrl" alt="" class="media-preview" />
        </div>

        <!-- Seçenekler -->
        <div v-if="q.type === 'single_choice'">
          <label style="font-weight: 600; font-size: 14px; margin-bottom: 12px; display: block; color: var(--text-color);">Seçenekler</label>
          
          <div style="display: flex; flex-direction: column; gap: 12px; margin-bottom: 16px;">
            <div v-for="(opt, oIndex) in q.options" :key="oIndex" class="option-block">
              <div class="flex items-center gap-3 option-row">
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
                  type="button"
                  title="Kaldır"
                >✕</button>
              </div>
              <div class="media-field media-field-compact">
                <span class="media-label">Seçenek görseli</span>
                <div class="media-row">
                  <label class="btn btn-outline media-file-btn">
                    Görsel
                    <input
                      type="file"
                      accept="image/*"
                      class="sr-only"
                      @change="onOptionImage(Number(qIndex), Number(oIndex), $event)"
                    />
                  </label>
                  <button
                    v-if="opt.imageFile || opt._previewUrl"
                    type="button"
                    class="btn btn-outline media-clear"
                    @click="clearOptionImage(Number(qIndex), Number(oIndex))"
                  >
                    Kaldır
                  </button>
                </div>
                <img v-if="opt._previewUrl" :src="opt._previewUrl" alt="" class="media-preview media-preview-sm" />
              </div>
            </div>
          </div>
          
          <button class="btn btn-outline" style="font-size: 14px; padding: 8px 16px; border-style: dashed;" @click="addOption(Number(qIndex))" type="button">
            + Yeni Seçenek Ekle
          </button>
        </div>
      </div>

      <div v-if="canCreateMultiQuestionSurveys()" class="flex justify-start mb-4">
        <button class="btn btn-outline" style="font-size: 14px; padding: 8px 16px; border-style: dashed;" @click="addQuestion" type="button">
          + Yeni Soru Ekle
        </button>
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
  inset: 0;
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  background-color: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 9999;
  padding: max(12px, env(safe-area-inset-top, 0px)) max(12px, env(safe-area-inset-right, 0px))
    max(12px, env(safe-area-inset-bottom, 0px)) max(12px, env(safe-area-inset-left, 0px));
  box-sizing: border-box;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}

.modal-content {
  background: var(--color-white);
  width: 100%;
  max-width: 600px;
  max-height: min(90vh, 100dvh - 24px);
  overflow-y: auto;
  border-radius: 20px;
  padding: 32px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
  margin: auto;
  flex-shrink: 0;
}

@media (max-width: 540px) {
  .modal-overlay {
    align-items: stretch;
    padding: 0;
  }
  .modal-content {
    max-height: none;
    min-height: 100vh;
    min-height: 100dvh;
    max-width: none;
    border-radius: 0;
    padding: max(var(--spacing-4), env(safe-area-inset-top, 0px)) max(var(--spacing-4), env(safe-area-inset-right, 0px))
      max(var(--spacing-4), env(safe-area-inset-bottom, 0px)) max(var(--spacing-4), env(safe-area-inset-left, 0px));
  }
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

.sr-only {
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
.media-field {
  margin-bottom: 14px;
}
.media-field-compact {
  margin-top: 8px;
  margin-bottom: 0;
}
.media-label {
  display: block;
  font-weight: 600;
  font-size: 13px;
  color: var(--text-color-muted);
  margin-bottom: 6px;
}
.media-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.media-file-btn {
  position: relative;
  font-size: 13px;
  padding: 8px 14px;
  cursor: pointer;
  overflow: hidden;
}
.media-clear {
  font-size: 13px;
  padding: 8px 12px;
  border: none;
  background: transparent;
  color: var(--text-color-muted);
}
.media-preview {
  display: block;
  max-width: 100%;
  max-height: 140px;
  width: auto;
  height: auto;
  margin-top: 10px;
  border-radius: 12px;
  object-fit: contain;
  object-position: center;
  border: 1px solid var(--border-color);
  background: var(--bg-color);
}
.media-preview-sm {
  max-height: 72px;
  max-width: 120px;
  object-fit: contain;
}
.option-block {
  padding-bottom: 8px;
  border-bottom: 1px dashed var(--border-color);
}
.option-block:last-of-type {
  border-bottom: none;
}

@media (max-width: 480px) {
  .option-row {
    flex-direction: column;
    align-items: stretch !important;
  }
  .option-row .btn {
    align-self: flex-end;
  }
}

</style>
