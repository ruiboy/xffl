<template>
  <router-link
    :to="to"
    class="flex items-center justify-between rounded-lg border border-gray-800 px-4 py-3 hover:border-gray-600 transition-colors"
  >
    <div class="flex items-center gap-2 font-medium">
      <span :class="{ 'text-white': winner === 'home' }">
        {{ match.homeClubMatch?.club.name ?? '—' }}
      </span>
      <span class="text-gray-500">v</span>
      <span :class="{ 'text-white': winner === 'away' }">
        {{ match.awayClubMatch?.club.name ?? '—' }}
      </span>
    </div>
    <span v-if="match.result" class="text-sm tabular-nums text-gray-400">
      {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
    </span>
  </router-link>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface ClubMatch {
  id: string
  club: { id: string; name: string }
  score: number
}

interface Match {
  id: string
  result?: string | null
  homeClubMatch?: ClubMatch | null
  awayClubMatch?: ClubMatch | null
}

const props = defineProps<{
  match: Match
  to: { name: string; params: Record<string, string> }
}>()

const winner = computed(() => {
  if (!props.match.result) return null
  if (props.match.result === 'home_win') return 'home'
  if (props.match.result === 'away_win') return 'away'
  return null
})
</script>
