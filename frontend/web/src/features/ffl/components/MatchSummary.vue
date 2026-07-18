<template>
  <div
    class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-3 hover:border-border-strong transition-colors cursor-pointer"
    @click="router.push(to)"
  >
    <!-- Versus: A [score] v [score] B -->
    <div v-if="style === 'versus'" class="flex items-center gap-3 font-medium">
      <img v-if="home" :src="clubLogoUrl(home.club.name)" :alt="home.club.name" class="w-8 h-8 object-contain shrink-0" />
      <span :class="winner === 'home' ? 'underline decoration-green-500 decoration-2 underline-offset-4' : ''">
        {{ home?.club.name ?? '—' }}
      </span>
      <BuildButton v-if="buildTeamTo && isMyClub(home)" :to="buildTeamTo" />
      <span v-if="hasScores" class="tabular-nums text-sm" :class="winner === 'home' ? 'font-bold text-text' : 'font-semibold text-text-muted'">{{ home?.score }}</span>
      <span class="text-text-faint">v</span>
      <span v-if="hasScores" class="tabular-nums text-sm" :class="winner === 'away' ? 'font-bold text-text' : 'font-semibold text-text-muted'">{{ away?.score }}</span>
      <img v-if="away" :src="clubLogoUrl(away.club.name)" :alt="away.club.name" class="w-8 h-8 object-contain shrink-0" />
      <span :class="winner === 'away' ? 'underline decoration-green-500 decoration-2 underline-offset-4' : ''">
        {{ away?.club.name ?? '—' }}
      </span>
      <BuildButton v-if="buildTeamTo && isMyClub(away)" :to="buildTeamTo" />
    </div>

    <!-- Bye: a single club, no opponent -->
    <div v-else-if="style === 'bye'" class="flex items-center gap-3 font-medium">
      <img v-if="home" :src="clubLogoUrl(home.club.name)" :alt="home.club.name" class="w-8 h-8 object-contain shrink-0" />
      <span>{{ home?.club.name ?? '—' }}</span>
      <span class="text-xs font-medium rounded-full bg-surface px-2 py-0.5 text-text-faint">Bye</span>
      <BuildButton v-if="buildTeamTo && isMyClub(home)" :to="buildTeamTo" />
      <span v-if="hasScores" class="tabular-nums text-sm font-semibold text-text-muted">{{ home?.score }}</span>
    </div>

    <!-- Superbye: every club, top scorer highlighted -->
    <div v-else class="flex items-center gap-x-4 gap-y-1.5 flex-wrap font-medium">
      <span class="text-xs font-medium rounded-full bg-surface px-2 py-0.5 text-text-faint shrink-0">Superbye</span>
      <span v-for="cm in clubMatches" :key="cm.id" class="flex items-center gap-1.5">
        <img :src="clubLogoUrl(cm.club.name)" :alt="cm.club.name" class="w-6 h-6 object-contain shrink-0" />
        <span :class="isTop(cm) ? 'underline decoration-green-500 decoration-2 underline-offset-4' : ''">{{ cm.club.name }}</span>
        <BuildButton v-if="buildTeamTo && isMyClub(cm)" :to="buildTeamTo" />
        <span v-if="hasScores" class="tabular-nums text-sm text-text-muted">{{ cm.score }}</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h } from 'vue'
import { useRouter } from 'vue-router'
import type { RouteLocationRaw } from 'vue-router'
import { clubLogoUrl } from '../utils/clubLogos'
import IconTeamBuilder from './icons/IconTeamBuilder.vue'

interface ClubMatch {
  id: string
  club: { id: string; name: string }
  score: number
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
