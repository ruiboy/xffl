<template>
  <div class="mb-3">
    <div class="flex items-center gap-2 mb-1">
      <img v-if="side.clubMatch" :src="clubLogoUrl(side.clubMatch.club.name)" :alt="side.clubMatch.club.name" class="w-8 h-8 object-contain" />
      <h2 class="text-lg font-semibold">
        <router-link
          v-if="side.clubMatch"
          :to="{ name: side.clubMatch.club.id === selectedClubId ? 'ffl-club-match-edit' : 'ffl-club-match', params: { clubMatchId: side.clubMatch.id } }"
          class="inline-flex items-center gap-1.5 hover:text-active transition-colors"
        >
          {{ side.label }}
          <IconTeamBuilder v-if="side.clubMatch.club.id === selectedClubId" class="w-4 h-4" />
        </router-link>
        <span v-else>{{ side.label }}</span>
      </h2>
    </div>
    <p class="text-sm text-text-muted mb-3">
      Score: <span class="font-semibold text-text">{{ side.clubMatch?.score ?? 0 }}</span>
      <span
        v-if="matchStyle === 'superbye' && topScore > 0 && (side.clubMatch?.score ?? 0) === topScore"
        class="ml-2 text-xs font-medium rounded-full bg-green-500/15 text-green-500 px-2 py-0.5"
      >top scorer · +1</span>
    </p>
    <div
      v-if="side.clubMatch?.club.id === selectedClubId && (side.clubMatch?.suggestedSubstitutions?.length ?? 0) > 0"
      class="rounded-lg border border-sky-500/30 bg-sky-500/10 px-4 py-3"
    >
      <p class="text-xs font-semibold text-sky-400 mb-1">Improve your score:</p>
      <ul class="space-y-0.5">
        <li
          v-for="s in side.clubMatch.suggestedSubstitutions"
          :key="s.replacingPmId"
          class="flex items-start gap-1.5 text-sm text-sky-300"
        >
          <span class="mt-px">·</span>
          <span>{{ s.kind === 'interchange' ? 'Interchange' : 'Sub' }}: {{ playerName(s.replacingPmId) }} in for {{ playerName(s.replacedPmId) }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { clubLogoUrl } from '../utils/clubLogos'
import IconTeamBuilder from './icons/IconTeamBuilder.vue'

interface PlayerMatch {
  id: string
  player: { aflPlayer: { name: string } }
}
interface ClubMatch {
  id: string
  club: { id: string; name: string }
  score: number
  suggestedSubstitutions?: { kind: string; replacedPmId: string; replacingPmId: string }[]
  playerMatches: PlayerMatch[]
}

const props = defineProps<{
  side: { label: string; clubMatch: ClubMatch | null }
  selectedClubId?: string | null
  matchStyle: string
  topScore: number
}>()

function playerName(pmId: string): string {
  return props.side.clubMatch?.playerMatches.find((pm) => pm.id === pmId)?.player.aflPlayer.name ?? pmId
}
</script>
