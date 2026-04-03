<template>
  <router-link
    :to="to"
    class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-3 hover:border-border-strong transition-colors"
  >
    <div class="flex items-center gap-3 font-medium">
      <span :class="{ 'font-bold': winner === 'home' }">
        {{ match.homeClubMatch?.club.name ?? '—' }}
      </span>
      <span class="text-text-faint">v</span>
      <span :class="{ 'font-bold': winner === 'away' }">
        {{ match.awayClubMatch?.club.name ?? '—' }}
      </span>
    </div>
    <span v-if="match.result" class="text-sm tabular-nums text-text-muted font-semibold">
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
