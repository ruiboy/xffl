<template>
  <div
    class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-3 hover:border-border-strong transition-colors cursor-pointer"
    @click="router.push(to)"
  >
    <div class="flex items-center gap-3 font-medium">
      <img v-if="homeLogo" :src="homeLogo" :alt="match.homeClubMatch?.club.name" class="w-8 h-8 object-contain shrink-0" />
      <span :class="{ 'font-bold': winner === 'home' }">{{ match.homeClubMatch?.club.name ?? '—' }}</span>
      <button
        v-if="buildTeamTo && myClubSide === 'home'"
        @click.stop="router.push(buildTeamTo)"
        title="Team Builder"
        class="rounded p-1 text-active hover:bg-active/10 transition-colors"
      >
        <IconWrench class="w-4 h-4" />
      </button>
      <span class="text-text-faint">v</span>
      <img v-if="awayLogo" :src="awayLogo" :alt="match.awayClubMatch?.club.name" class="w-8 h-8 object-contain shrink-0" />
      <span :class="{ 'font-bold': winner === 'away' }">{{ match.awayClubMatch?.club.name ?? '—' }}</span>
      <button
        v-if="buildTeamTo && myClubSide === 'away'"
        @click.stop="router.push(buildTeamTo)"
        title="Team Builder"
        class="rounded p-1 text-active hover:bg-active/10 transition-colors"
      >
        <IconWrench class="w-4 h-4" />
      </button>
    </div>
    <span v-if="hasScores" class="text-sm tabular-nums text-text-muted font-semibold">
      {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { RouteLocationRaw } from 'vue-router'
import { clubLogoUrl } from '../utils/clubLogos'
import IconWrench from './IconWrench.vue'

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

const hasScores = computed(() =>
  (props.match.homeClubMatch?.score ?? 0) > 0 || (props.match.awayClubMatch?.score ?? 0) > 0
)

const myClubSide = computed(() => {
  if (!props.myClubId) return null
  if (props.match.homeClubMatch?.club.id === props.myClubId) return 'home'
  if (props.match.awayClubMatch?.club.id === props.myClubId) return 'away'
  return null
})

const winner = computed(() => {
  if (!hasScores.value) return null
  const home = props.match.homeClubMatch?.score ?? 0
  const away = props.match.awayClubMatch?.score ?? 0
  if (home > away) return 'home'
  if (away > home) return 'away'
  return null
})

</script>
