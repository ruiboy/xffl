<template>
  <router-link
    :to="to"
    class="flex items-center justify-between rounded-lg border border-gray-200 bg-white px-4 py-3 hover:border-gray-400 transition-colors"
  >
    <div class="flex items-center gap-3 font-medium">
      <img v-if="homeLogo" :src="homeLogo" :alt="match.homeClubMatch?.club.name" class="w-8 h-8 object-contain" />
      <span :class="{ 'font-bold': winner === 'home' }">
        {{ match.homeClubMatch?.club.name ?? '—' }}
      </span>
      <span class="text-gray-400">v</span>
      <img v-if="awayLogo" :src="awayLogo" :alt="match.awayClubMatch?.club.name" class="w-8 h-8 object-contain" />
      <span :class="{ 'font-bold': winner === 'away' }">
        {{ match.awayClubMatch?.club.name ?? '—' }}
      </span>
    </div>
    <span v-if="match.result" class="text-sm tabular-nums text-gray-500 font-semibold">
      {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
    </span>
  </router-link>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { clubLogoUrl } from '../utils/clubLogos'

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

const homeLogo = computed(() => props.match.homeClubMatch ? clubLogoUrl(props.match.homeClubMatch.club.name) : '')
const awayLogo = computed(() => props.match.awayClubMatch ? clubLogoUrl(props.match.awayClubMatch.club.name) : '')

const winner = computed(() => {
  if (!props.match.result) return null
  if (props.match.result === 'home_win') return 'home'
  if (props.match.result === 'away_win') return 'away'
  return null
})
</script>
