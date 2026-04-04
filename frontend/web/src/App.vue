<template>
  <div class="min-h-screen bg-surface text-text">
    <nav class="border-b border-border bg-surface-raised px-6 py-4">
      <div class="max-w-5xl mx-auto flex items-center gap-6">
        <router-link to="/ffl" class="text-lg font-semibold tracking-tight text-text hover:text-text-muted transition-colors">
          xFFL
        </router-link>
        <router-link to="/ffl" class="text-sm text-text-muted hover:text-text transition-colors">
          FFL
        </router-link>
<router-link to="/afl" class="text-sm text-text-muted hover:text-text transition-colors">
          AFL
        </router-link>
        <button
          class="ml-auto text-sm text-text-muted hover:text-text transition-colors"
          @click="toggleTheme"
        >
          {{ isDark ? 'Light' : 'Dark' }}
        </button>
      </div>
    </nav>
    <main class="max-w-5xl mx-auto px-6 py-8">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const isDark = ref(false)

function applyTheme(dark: boolean) {
  document.documentElement.classList.toggle('dark', dark)
  localStorage.setItem('theme', dark ? 'dark' : 'light')
  isDark.value = dark
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

onMounted(() => {
  const saved = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(saved ? saved === 'dark' : prefersDark)
})
</script>
