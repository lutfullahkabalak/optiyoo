<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'

interface AppConfig {
  active_theme?: string
  default_theme?: string
  themes?: string[]
}

onMounted(async () => {
  try {
    const res = await fetch('http://localhost:8080/api/config')
    if (res.ok) {
      const config = await res.json() as AppConfig
      const themeClass = config.active_theme || config.default_theme

      if (themeClass) {
        // Remove previous theme classes
        document.body.classList.forEach(c => {
          if (c.startsWith('theme-')) document.body.classList.remove(c)
        })

        // Use backend-provided theme directly (no concatenation)
        document.body.classList.add(themeClass)
      }
    }
  } catch (e) {
    console.error('Config fetch error:', e)
  }
})
</script>

<template>
  <RouterView />
</template>
