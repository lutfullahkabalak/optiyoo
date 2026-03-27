<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <div class="app-wrapper animate-fade-in">
    <header class="header">
      <div class="container flex justify-between items-center header-inner">
        <RouterLink to="/" class="logo-link">
          <span class="logo">optiyoo</span>
        </RouterLink>
        <nav class="nav">
          <RouterLink to="/" style="text-decoration: none">
            <button class="btn btn-outline" type="button" style="font-size: 13px">← Ana sayfa</button>
          </RouterLink>
        </nav>
      </div>
    </header>

    <main class="main flex flex-col items-center justify-center text-center">
      <div class="not-found-inner">
        <h1 class="code">
          <span>4</span><span>0</span><span>4</span>
        </h1>
        <p class="subtitle">Aradığın sayfayı bulamadık</p>
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
.logo-link {
  text-decoration: none;
}
.logo {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--primary-color);
  letter-spacing: -0.05em;
}
.main {
  flex: 1;
  padding: var(--spacing-8) var(--spacing-4);
}
.not-found-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  transform: translateY(calc(-1 * clamp(1.25rem, 7vh, 4.5rem)));
}
.code {
  display: flex;
  justify-content: center;
  gap: 0.14em;
  width: max-content;
  font-size: clamp(5.5rem, 28vw, 14rem);
  font-weight: 700;
  line-height: 0.9;
  color: var(--primary-color);
  margin: 0 0 var(--spacing-6);
}
.subtitle {
  font-size: var(--font-size-xl);
  font-weight: 500;
  color: var(--text-color-muted);
  margin: 0;
  max-width: min(22rem, 100%);
  text-align: center;
  text-wrap: balance;
}
</style>
