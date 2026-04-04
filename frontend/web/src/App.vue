<template>
  <div class="min-h-screen bg-surface text-text">
    <nav class="border-b border-border bg-surface-raised px-6 py-4">
      <div class="max-w-5xl mx-auto flex items-center gap-6">
        <!-- Left: brand + service links -->
        <router-link to="/ffl" class="text-lg font-semibold tracking-tight text-text hover:text-text-muted transition-colors">
          xFFL
        </router-link>
        <router-link to="/ffl" class="text-sm text-text-muted hover:text-text transition-colors">
          FFL
        </router-link>
        <router-link to="/afl" class="text-sm text-text-muted hover:text-text transition-colors">
          AFL
        </router-link>

        <!-- Right: FFL nav + theme toggle -->
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
            <router-link
              :to="currentSeasonId && currentRoundId ? { name: 'ffl-team-builder', params: { seasonId: currentSeasonId, roundId: currentRoundId } } : '/ffl'"
              class="text-sm transition-colors"
              :class="currentSeasonId && currentRoundId ? 'text-text-muted hover:text-text' : 'text-text-faint pointer-events-none'"
            >
              Team Builder
            </router-link>

            <!-- Club selector -->
            <select
              v-if="clubs.length > 0"
              :value="selectedClubId"
              @change="onClubChange"
              class="rounded-lg border border-border bg-surface px-3 py-1 text-sm text-text focus:border-active focus:outline-none"
            >
              <option v-for="cs in clubs" :key="cs.club.id" :value="cs.club.id">
                {{ cs.club.name }}
              </option>
            </select>
          </template>

          <button
            class="text-sm text-text-muted hover:text-text transition-colors"
            @click="toggleTheme"
          >
            {{ isDark ? 'Light' : 'Dark' }}
          </button>
        </div>
      </div>
    </nav>
    <main class="max-w-5xl mx-auto px-6 py-8">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASON_CLUBS } from '@/features/ffl/api/queries'
import { useFflState } from '@/features/ffl/composables/useFflState'

const route = useRoute()
const { selectedClubId, currentSeasonId, currentRoundId, setClub } = useFflState()

const isFfl = computed(() => route.path.startsWith('/ffl'))

// Load clubs for the current FFL season (driven by state, not route params)
const { result: clubsResult } = useQuery(
  GET_FFL_SEASON_CLUBS,
  () => ({ seasonId: currentSeasonId.value }),
  () => ({ enabled: isFfl.value && !!currentSeasonId.value })
)

const clubs = computed(() => clubsResult.value?.fflSeason?.ladder ?? [])

// Auto-select first club if nothing stored
watch(clubs, (list) => {
  if (list.length > 0 && !selectedClubId.value) {
    setClub(list[0].club.id)
  }
})

function onClubChange(event: Event) {
  setClub((event.target as HTMLSelectElement).value)
}

// Theme
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
