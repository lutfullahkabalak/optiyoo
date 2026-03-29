<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  file: File | null
}>()

const emit = defineEmits<{
  'update:modelValue': [boolean]
  apply: [Blob]
}>()

const V = 260
const OUT = 400

const objectUrl = ref('')
const imgEl = ref<HTMLImageElement | null>(null)
const naturalW = ref(0)
const naturalH = ref(0)
const scale = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)
const dragging = ref(false)
let dragStart = { x: 0, y: 0, ox: 0, oy: 0 }

function resetState() {
  scale.value = 1
  offsetX.value = 0
  offsetY.value = 0
  naturalW.value = 0
  naturalH.value = 0
}

watch(
  () => [props.modelValue, props.file] as const,
  ([open, file]) => {
    if (open && file) {
      if (objectUrl.value) URL.revokeObjectURL(objectUrl.value)
      objectUrl.value = URL.createObjectURL(file)
      resetState()
    }
    if (!open && objectUrl.value) {
      URL.revokeObjectURL(objectUrl.value)
      objectUrl.value = ''
    }
  }
)

onUnmounted(() => {
  if (objectUrl.value) URL.revokeObjectURL(objectUrl.value)
})

function onImgLoad() {
  const el = imgEl.value
  if (!el) return
  naturalW.value = el.naturalWidth
  naturalH.value = el.naturalHeight
}

function baseScale() {
  const iw = naturalW.value
  const ih = naturalH.value
  if (!iw || !ih) return 1
  return Math.max(V / iw, V / ih)
}

function imgStyle() {
  const iw = naturalW.value
  const ih = naturalH.value
  if (!iw || !ih) return {}
  const s = baseScale() * scale.value
  const w = iw * s
  const h = ih * s
  return {
    width: `${w}px`,
    height: `${h}px`,
    position: 'absolute' as const,
    left: `${V / 2 - w / 2 + offsetX.value}px`,
    top: `${V / 2 - h / 2 + offsetY.value}px`,
    maxWidth: 'none',
    userSelect: 'none' as const,
    touchAction: 'none' as const
  }
}

function onPointerDown(e: PointerEvent) {
  dragging.value = true
  ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
  dragStart = { x: e.clientX, y: e.clientY, ox: offsetX.value, oy: offsetY.value }
}

function onPointerMove(e: PointerEvent) {
  if (!dragging.value) return
  offsetX.value = dragStart.ox + (e.clientX - dragStart.x)
  offsetY.value = dragStart.oy + (e.clientY - dragStart.y)
}

function onPointerUp(e: PointerEvent) {
  dragging.value = false
  try {
    ;(e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId)
  } catch {
    /* ignore */
  }
}

function close() {
  emit('update:modelValue', false)
}

function applyCrop() {
  const img = imgEl.value
  if (!img || !naturalW.value || !naturalH.value) return
  const canvas = document.createElement('canvas')
  canvas.width = OUT
  canvas.height = OUT
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const k = OUT / V
  const iw = naturalW.value
  const ih = naturalH.value
  const s = baseScale() * scale.value
  const dw = iw * s * k
  const dh = ih * s * k
  const dx = OUT / 2 - dw / 2 + offsetX.value * k
  const dy = OUT / 2 - dh / 2 + offsetY.value * k
  ctx.save()
  ctx.beginPath()
  ctx.arc(OUT / 2, OUT / 2, OUT / 2, 0, Math.PI * 2)
  ctx.closePath()
  ctx.clip()
  ctx.drawImage(img, dx, dy, dw, dh)
  ctx.restore()
  canvas.toBlob(
    (blob) => {
      if (blob) emit('apply', blob)
      close()
    },
    'image/jpeg',
    0.9
  )
}
</script>

<template>
  <Teleport to="body">
    <div v-if="modelValue" class="crop-overlay" @click.self="close">
      <div class="crop-dialog card">
        <h3 class="crop-title">Profil fotoğrafı</h3>
        <p class="crop-hint">Daire içinde nasıl görüneceğini sürükleyerek ve yakınlaştırarak ayarlayın.</p>
        <div class="crop-stage-outer">
          <div class="crop-stage" :style="{ width: V + 'px', height: V + 'px' }">
            <img
              v-show="objectUrl"
              ref="imgEl"
              :src="objectUrl"
              alt=""
              draggable="false"
              :style="imgStyle()"
              @load="onImgLoad"
              @pointerdown.prevent="onPointerDown"
              @pointermove="onPointerMove"
              @pointerup="onPointerUp"
              @pointercancel="onPointerUp"
            />
          </div>
        </div>
        <label class="crop-zoom-label">
          Yakınlaştırma
          <input v-model.number="scale" type="range" min="1" max="3" step="0.02" />
        </label>
        <div class="crop-actions">
          <button type="button" class="btn btn-outline" @click="close">İptal</button>
          <button type="button" class="btn btn-primary" @click="applyCrop">Kırp ve yükle</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.crop-overlay {
  position: fixed;
  inset: 0;
  z-index: 2000;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-4);
  box-sizing: border-box;
}
.crop-dialog {
  max-width: 100%;
  width: 400px;
  padding: var(--spacing-6);
}
.crop-title {
  margin: 0 0 var(--spacing-2);
  font-size: var(--font-size-lg);
  font-weight: 700;
  color: var(--text-color);
}
.crop-hint {
  margin: 0 0 var(--spacing-4);
  font-size: var(--font-size-sm);
  color: var(--text-color-muted);
  line-height: 1.45;
}
.crop-stage-outer {
  display: flex;
  justify-content: center;
  margin-bottom: var(--spacing-4);
}
.crop-stage {
  position: relative;
  border-radius: 50%;
  overflow: hidden;
  background: var(--border-color, #e5e7eb);
  box-shadow: inset 0 0 0 3px rgba(255, 255, 255, 0.85);
  cursor: grab;
}
.crop-stage:active {
  cursor: grabbing;
}
.crop-zoom-label {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-2);
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--text-color);
  margin-bottom: var(--spacing-4);
}
.crop-zoom-label input {
  width: 100%;
}
.crop-actions {
  display: flex;
  gap: var(--spacing-3);
  justify-content: flex-end;
  flex-wrap: wrap;
}
</style>
