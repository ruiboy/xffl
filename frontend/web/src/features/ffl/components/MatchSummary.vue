<template>
  <div
    class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-3 hover:border-border-strong transition-colors cursor-pointer"
    :style="rowStyle"
    @click="router.push(to)"
  >
    <!-- Versus: home left · scores centred · away right (logos on the outer edges) -->
    <div v-if="style === 'versus'" class="flex flex-1 items-center gap-3 font-medium text-lg min-w-0">
      <!-- Home: logo, name -->
      <div class="flex flex-1 items-center gap-3 min-w-0">
        <img v-if="home" :src="clubLogoUrl(home.club.name)" :alt="home.club.name" class="w-8 h-8 object-contain shrink-0" />
        <span class="truncate" :class="winner === 'home' ? 'underline decoration-green-500 decoration-2 underline-offset-4' : ''">
          {{ home?.club.name ?? '—' }}
        </span>
        <BuildButton v-if="buildTeamTo && isMyClub(home)" :to="buildTeamTo" />
      </div>
      <!-- Scores: equal-width boxes flank the 'v' so it stays centred regardless of score widths.
           A parenthesised count of on-field players who have played sits after each score. -->
      <div class="flex items-center gap-3 shrink-0">
        <span v-if="hasScores" class="tabular-nums text-base w-24 text-right" :class="winner === 'home' ? 'font-bold text-text' : 'font-semibold text-text-muted'"><PlayedCount :club-match="home" class="text-xs mr-2" />{{ home?.score }}</span>
        <span class="text-text-faint">v</span>
        <span v-if="hasScores" class="tabular-nums text-base w-24 text-left" :class="winner === 'away' ? 'font-bold text-text' : 'font-semibold text-text-muted'">{{ away?.score }}<PlayedCount :club-match="away" class="text-xs ml-2" /></span>
      </div>
      <!-- Away: name, logo -->
      <div class="flex flex-1 items-center justify-end gap-3 min-w-0">
        <BuildButton v-if="buildTeamTo && isMyClub(away)" :to="buildTeamTo" />
        <span class="truncate text-right" :class="winner === 'away' ? 'underline decoration-green-500 decoration-2 underline-offset-4' : ''">
          {{ away?.club.name ?? '—' }}
        </span>
        <img v-if="away" :src="clubLogoUrl(away.club.name)" :alt="away.club.name" class="w-8 h-8 object-contain shrink-0" />
      </div>
    </div>

    <!-- Bye: a single club, no opponent -->
    <div v-else-if="style === 'bye'" class="flex items-center gap-3 font-medium text-lg">
      <img v-if="home" :src="clubLogoUrl(home.club.name)" :alt="home.club.name" class="w-8 h-8 object-contain shrink-0" />
      <span>{{ home?.club.name ?? '—' }}</span>
      <span class="text-xs font-medium rounded-full bg-surface px-2 py-0.5 text-text-faint">Bye</span>
      <BuildButton v-if="buildTeamTo && isMyClub(home)" :to="buildTeamTo" />
      <span v-if="hasScores" class="tabular-nums text-base font-semibold text-text-muted">{{ home?.score }}</span>
    </div>

    <!-- Superbye: every club, top scorer highlighted -->
    <div v-else class="flex items-center gap-x-4 gap-y-1.5 flex-wrap font-medium text-lg">
      <span class="text-xs font-medium rounded-full bg-surface px-2 py-0.5 text-text-faint shrink-0">Superbye</span>
      <span v-for="cm in clubMatches" :key="cm.id" class="flex items-center gap-1.5">
        <img :src="clubLogoUrl(cm.club.name)" :alt="cm.club.name" class="w-8 h-8 object-contain shrink-0" />
        <span :class="isTop(cm) ? 'underline decoration-green-500 decoration-2 underline-offset-4' : ''">{{ cm.club.name }}</span>
        <BuildButton v-if="buildTeamTo && isMyClub(cm)" :to="buildTeamTo" />
        <span v-if="hasScores" class="tabular-nums text-base text-text-muted">{{ cm.score }}</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useRouter } from 'vue-router'
import type { RouteLocationRaw } from 'vue-router'
import { clubLogoUrl } from '../utils/clubLogos'
import { clubColorRgba } from '../utils/clubColors'
import PlayedCount from './PlayedCount.vue'
import IconTeamBuilder from './icons/IconTeamBuilder.vue'

interface PlayerMatch {
  status?: string | null
  aflStatus?: string | null
  backupPositions?: string | null
  interchangePosition?: string | null
}

interface ClubMatch {
  id: string
  club: { id: string; name: string }
  score: number
  dataStatus?: string | null
  playerMatches?: PlayerMatch[] | null
}

interface Match {
  id: string
  result?: string | null
  matchStyle?: string | null
  clubMatches?: ClubMatch[] | null
}

const props = defineProps<{
  match: Match
  to: RouteLocationRaw
  myClubId?: string
  buildTeamTo?: RouteLocationRaw
}>()

const router = useRouter()

// A small Team Builder link that stops propagation so it doesn't trigger the card.
const BuildButton = (p: { to: RouteLocationRaw }) =>
  h('button', {
    title: 'Team Builder',
    class: 'rounded p-1 text-active hover:bg-active/10 transition-colors shrink-0',
    onClick: (e: Event) => { e.stopPropagation(); router.push(p.to) },
  }, [h(IconTeamBuilder, { class: 'w-4 h-4' })])

const style = computed(() => props.match.matchStyle ?? 'versus')
const clubMatches = computed<ClubMatch[]>(() => props.match.clubMatches ?? [])
const home = computed(() => clubMatches.value[0] ?? null)
const away = computed(() => clubMatches.value[1] ?? null)

// Experimental: tint the row with each club's colour, fading in from its side so
// the middle stays neutral. Just to relieve the visual monotony of the list.
const rowStyle = computed(() => {
  const homeTint = clubColorRgba(clubMatches.value[0]?.club.name, 0.18)
  const awayTint = clubColorRgba(clubMatches.value[1]?.club.name, 0.18)
  if (!homeTint && !awayTint) return {}
  return {
    backgroundImage: `linear-gradient(to right, ${homeTint ?? 'transparent'}, transparent 42%, transparent 58%, ${awayTint ?? 'transparent'})`,
  }
})

const hasScores = computed(() => clubMatches.value.some((cm) => (cm.score ?? 0) > 0))
const topScore = computed(() => Math.max(0, ...clubMatches.value.map((cm) => cm.score ?? 0)))

const winner = computed(() => {
  if (style.value !== 'versus' || !hasScores.value) return null
  const hs = home.value?.score ?? 0
  const as = away.value?.score ?? 0
  if (hs > as) return 'home'
  if (as > hs) return 'away'
  return null
})

function isMyClub(cm: ClubMatch | null): boolean {
  return !!props.myClubId && !!cm && cm.club.id === props.myClubId
}
function isTop(cm: ClubMatch): boolean {
  return topScore.value > 0 && (cm.score ?? 0) === topScore.value
}
</script>
