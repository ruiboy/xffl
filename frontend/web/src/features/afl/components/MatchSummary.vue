<template>
  <router-link
    :to="to"
    class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-3 hover:border-border-strong transition-colors"
    :style="rowStyle"
  >
    <!-- Versus: home left · scores centred · away right (logos on the outer edges) -->
    <div class="flex flex-1 items-center gap-3 font-medium text-lg min-w-0">
      <!-- Home: logo, name -->
      <div class="flex flex-1 items-center gap-3 min-w-0">
        <img v-if="homeLogo" :src="homeLogo" :alt="match.homeClubMatch?.club.name" class="w-8 h-8 object-contain shrink-0" />
        <span class="truncate" :class="{ 'underline decoration-green-500 decoration-2 underline-offset-4': winner === 'home' }">
          {{ match.homeClubMatch?.club.name ?? '—' }}
        </span>
      </div>
      <!-- Scores: equal-width boxes flank the 'v' so it stays centred regardless of score widths -->
      <div class="flex items-center gap-3 shrink-0">
        <span v-if="hasScores" class="tabular-nums text-base w-10 text-right" :class="winner === 'home' ? 'font-bold text-text' : 'font-semibold text-text-muted'">
          {{ match.homeClubMatch?.score }}
        </span>
        <span class="text-text-faint">v</span>
        <span v-if="hasScores" class="tabular-nums text-base w-10 text-left" :class="winner === 'away' ? 'font-bold text-text' : 'font-semibold text-text-muted'">
          {{ match.awayClubMatch?.score }}
        </span>
      </div>
      <!-- Away: name, logo -->
      <div class="flex flex-1 items-center justify-end gap-3 min-w-0">
        <span class="truncate text-right" :class="{ 'underline decoration-green-500 decoration-2 underline-offset-4': winner === 'away' }">
          {{ match.awayClubMatch?.club.name ?? '—' }}
        </span>
        <img v-if="awayLogo" :src="awayLogo" :alt="match.awayClubMatch?.club.name" class="w-8 h-8 object-contain shrink-0" />
      </div>
    </div>
  </router-link>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { clubColorRgba } from '../utils/clubColors'
import { useTheme } from '@/composables/useTheme'

const { isDark } = useTheme()

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

// Tint the row with each club's colour, fading in from its side so the middle
// stays neutral — just to relieve the visual monotony of the list.
const rowStyle = computed(() => {
  const homeTint = clubColorRgba(props.match.homeClubMatch?.club.name, 0.18, isDark.value)
  const awayTint = clubColorRgba(props.match.awayClubMatch?.club.name, 0.18, isDark.value)
  if (!homeTint && !awayTint) return {}
  return {
    backgroundImage: `linear-gradient(to right, ${homeTint ?? 'transparent'}, transparent 42%, transparent 58%, ${awayTint ?? 'transparent'})`,
  }
})

const hasScores = computed(() =>
  props.match.result === 'home_win' || props.match.result === 'away_win' || props.match.result === 'draw'
)

const winner = computed(() => {
  if (props.match.result === 'home_win') return 'home'
  if (props.match.result === 'away_win') return 'away'
  return null
})
</script>
