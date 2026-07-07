import { ref, readonly } from 'vue'
import { getCookie, setCookie } from '@/utils/cookie'

const isDark = ref(false)

function applyTheme(dark: boolean) {
  document.documentElement.classList.toggle('dark', dark)
  setCookie('xffl_dark_mode', dark ? '1' : '0', 365 * 10)
  isDark.value = dark
}

export function initTheme() {
  const saved = getCookie('xffl_dark_mode')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(saved !== '' ? saved === '1' : prefersDark)
}

export function useTheme() {
  return {
    isDark: readonly(isDark),
    toggleTheme: () => applyTheme(!isDark.value),
  }
}
