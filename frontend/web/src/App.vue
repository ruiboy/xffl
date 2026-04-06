<template>
  <div class="min-h-screen bg-surface text-text">
    <nav class="border-b border-border bg-surface-raised px-6 py-4">
      <div class="max-w-5xl mx-auto flex items-center gap-6">
        <!-- Left: brand + service links -->
        <router-link to="/ffl" class="flex items-center">
          <img src="/images/ffl-eagle-logo.png" alt="FFL" class="h-10 w-auto transition-transform duration-200 hover:scale-[3] origin-top-left" />
        </router-link>
        <router-link to="/ffl" class="text-sm text-text-muted hover:text-text transition-colors">
          FFL
        </router-link>
        <router-link to="/afl" class="text-sm text-text-muted hover:text-text transition-colors">
          AFL
        </router-link>

        <!-- Right: FFL nav + settings -->
        <div class="ml-auto flex items-center gap-4">
          <!-- Squad + Team Builder (always visible on FFL pages once season is known) -->
          <template v-if="isFfl">
            <router-link
              :to="currentSeasonId ? { name: 'ffl-squad', params: { seasonId: currentSeasonId } } : '/ffl'"
              class="text-sm transition-colors"
              :class="currentSeasonId ? 'text-text-muted hover:text-text' : 'text-text-faint pointer-events-none'"
            >
              Squad
            </router-link>
            <!-- Club selector -->
            <ClubSelector
              v-if="clubs.length > 0"
              :model-value="selectedClubId"
              :clubs="clubs"
              @update:model-value="setClub"
            />
          </template>

          <!-- Settings cog -->
          <div class="relative" ref="settingsContainer">
            <button
              @click="settingsOpen = !settingsOpen"
              class="text-text-muted hover:text-text transition-colors translate-y-0.5"
              title="Settings"
            >
              <svg class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd" />
              </svg>
            </button>

            <div
              v-if="settingsOpen"
              class="absolute right-0 top-full mt-2 w-48 rounded-lg border border-border bg-surface-raised shadow-lg py-2 z-50"
            >
              <div class="px-3 py-1.5 flex items-center justify-between">
                <span class="text-sm text-text-muted">Dark mode</span>
                <button
                  @click="toggleTheme"
                  class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors"
                  :class="isDark ? 'bg-active' : 'bg-control-ring'"
                >
                  <span
                    class="inline-block h-3.5 w-3.5 transform rounded-full bg-white shadow transition-transform"
                    :class="isDark ? 'translate-x-4' : 'translate-x-1'"
                  />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>
    <main class="max-w-5xl mx-auto px-6 py-8">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASON_CLUBS } from '@/features/ffl/api/queries'
import { useFflState } from '@/features/ffl/composables/useFflState'
import ClubSelector from '@/features/ffl/components/ClubSelector.vue'

const route = useRoute()
const { selectedClubId, currentSeasonId, setClub } = useFflState()

const isFfl = computed(() => route.path.startsWith('/ffl'))

// Load clubs for the current FFL season (driven by state, not route params)
const { result: clubsResult } = useQuery(
  GET_FFL_SEASON_CLUBS,
  () => ({ seasonId: currentSeasonId.value }),
  () => ({ enabled: isFfl.value && !!currentSeasonId.value })
)

const rawClubs = computed(() => clubsResult.value?.fflSeason?.ladder ?? [])

// Persist the last non-empty clubs list so the ClubSelector doesn't
// disappear during transient cache re-evaluations after mutations.
const clubs = ref<typeof rawClubs.value>([])
watch(rawClubs, (list) => {
  if (list.length > 0) clubs.value = list
}, { immediate: true })

// Auto-select first club if nothing stored
watch(clubs, (list) => {
  if (list.length > 0 && !selectedClubId.value) {
    setClub(list[0].club.id)
  }
})

// Settings dropdown
const settingsOpen = ref(false)
const settingsContainer = ref<HTMLElement | null>(null)

function onClickOutside(e: MouseEvent) {
  if (settingsContainer.value && !settingsContainer.value.contains(e.target as Node)) {
    settingsOpen.value = false
  }
}

onMounted(() => document.addEventListener('mousedown', onClickOutside))
onUnmounted(() => document.removeEventListener('mousedown', onClickOutside))

// Theme
const isDark = ref(false)

function getThemeCookie(): string {
  const match = document.cookie.match(/(^| )xffl_dark_mode=([^;]+)/)
  return match ? match[2] : ''
}

function setThemeCookie(dark: boolean) {
  const expires = new Date()
  expires.setFullYear(expires.getFullYear() + 10)
  document.cookie = `xffl_dark_mode=${dark ? '1' : '0'};expires=${expires.toUTCString()};path=/`
}

function applyTheme(dark: boolean) {
  document.documentElement.classList.toggle('dark', dark)
  setThemeCookie(dark)
  isDark.value = dark
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

onMounted(() => {
  const saved = getThemeCookie()
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  applyTheme(saved !== '' ? saved === '1' : prefersDark)
})
</script>
