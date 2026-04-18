<template>
  <div
    class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-3 hover:border-border-strong transition-colors cursor-pointer"
    @click="router.push(to)"
  >
    <div class="flex items-center gap-3 font-medium">
      <img v-if="homeLogo" :src="homeLogo" :alt="match.homeClubMatch?.club.name" class="w-8 h-8 object-contain" />
      <span :class="{ 'font-bold': winner === 'home' }">{{ match.homeClubMatch?.club.name ?? '—' }}</span>
      <button
        v-if="buildTeamTo && myClubSide === 'home'"
        @click.stop="router.push(buildTeamTo)"
        title="Team Builder"
        class="rounded p-1 text-active hover:bg-active/10 transition-colors"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437 1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008Z"/>
        </svg>
      </button>
      <span class="text-text-faint">v</span>
      <img v-if="awayLogo" :src="awayLogo" :alt="match.awayClubMatch?.club.name" class="w-8 h-8 object-contain" />
      <span :class="{ 'font-bold': winner === 'away' }">{{ match.awayClubMatch?.club.name ?? '—' }}</span>
      <button
        v-if="buildTeamTo && myClubSide === 'away'"
        @click.stop="router.push(buildTeamTo)"
        title="Team Builder"
        class="rounded p-1 text-active hover:bg-active/10 transition-colors"
      >
        <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437 1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008Z"/>
        </svg>
      </button>
    </div>
    <span v-if="match.result" class="text-sm tabular-nums text-text-muted font-semibold">
      {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { RouteLocationRaw } from 'vue-router'
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
  to: RouteLocationRaw
  myClubId?: string
  buildTeamTo?: RouteLocationRaw
}>()

const router = useRouter()

const homeLogo = computed(() => props.match.homeClubMatch ? clubLogoUrl(props.match.homeClubMatch.club.name) : '')
const awayLogo = computed(() => props.match.awayClubMatch ? clubLogoUrl(props.match.awayClubMatch.club.name) : '')

const winner = computed(() => {
  if (!props.match.result) return null
  if (props.match.result === 'home_win') return 'home'
  if (props.match.result === 'away_win') return 'away'
  return null
})

const myClubSide = computed(() => {
  if (!props.myClubId) return null
  if (props.match.homeClubMatch?.club.id === props.myClubId) return 'home'
  if (props.match.awayClubMatch?.club.id === props.myClubId) return 'away'
  return null
})
</script>
