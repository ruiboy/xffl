<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-6">
        <Breadcrumb v-if="round" :items="breadcrumbs" />
        <h1 class="text-2xl font-bold flex items-center gap-3">
          <img v-if="match.homeClubMatch" :src="clubLogoUrl(match.homeClubMatch.club.name)" :alt="match.homeClubMatch.club.name" class="w-10 h-10 object-contain" />
          <router-link v-if="match.homeClubMatch" :to="{ name: 'ffl-club-season', params: { clubSeasonId: match.homeClubMatch.clubSeasonId } }" class="hover:text-text-muted transition-colors">{{ match.homeClubMatch.club.name }}</router-link>
          <span v-else>—</span>
          <span class="text-text-faint mx-1">v</span>
          <img v-if="match.awayClubMatch" :src="clubLogoUrl(match.awayClubMatch.club.name)" :alt="match.awayClubMatch.club.name" class="w-10 h-10 object-contain" />
          <router-link v-if="match.awayClubMatch" :to="{ name: 'ffl-club-season', params: { clubSeasonId: match.awayClubMatch.clubSeasonId } }" class="hover:text-text-muted transition-colors">{{ match.awayClubMatch.club.name }}</router-link>
          <span v-else>—</span>
        </h1>
        <p v-if="match.venue" class="text-sm text-text-muted mt-1">{{ match.venue }}</p>
        <p v-if="match.result" class="text-lg font-semibold mt-2">
          {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
        </p>
      </div>

      <!-- Headers row: both cells share the same row height, so SquadTables below always align -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-x-8 mb-0">
        <div v-for="side in sides" :key="side.label + '-hd'" class="mb-3">
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
          </p>
          <div
            v-if="side.clubMatch?.club.id === selectedClubId && (side.clubMatch?.suggestedSubstitutions?.length ?? 0) > 0"
            class="rounded-lg border border-sky-500/30 bg-sky-500/10 px-4 py-3"
          >
            <p class="text-xs font-semibold text-sky-400 mb-1">Improve your score:</p>
            <ul class="space-y-0.5">
              <li
                v-for="s in side.clubMatch!.suggestedSubstitutions"
                :key="s.replacingPmId"
                class="flex items-start gap-1.5 text-sm text-sky-300"
              >
                <span class="mt-px">·</span>
                <span>{{ s.kind === 'interchange' ? 'Interchange' : 'Sub' }}: {{ playerName(side.clubMatch!, s.replacingPmId) }} in for {{ playerName(side.clubMatch!, s.replacedPmId) }}</span>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <!-- Tables row: starts at the same Y for both columns -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-x-8">
        <div v-for="side in sides" :key="side.label + '-tbl'">
          <SquadTable v-if="side.clubMatch" :player-matches="side.clubMatch.playerMatches" />
        </div>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_MATCH } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'
import SquadTable from '../components/SquadTable.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { useFflState } from '../composables/useFflState'
import IconTeamBuilder from '../components/icons/IconTeamBuilder.vue'

const props = defineProps<{ matchId: string }>()

const { selectedClubId } = useFflState()
const { result, loading, error } = useQuery(GET_FFL_MATCH, () => ({ id: props.matchId }))

const match = computed(() => result.value?.fflMatch ?? null)
const round = computed(() => match.value?.round ?? null)

const breadcrumbs = computed(() => {
  if (!match.value || !round.value) return []
  return [
    { label: 'FFL' },
    { label: round.value.season.name, to: { name: 'home' } },
    { label: round.value.name, to: { name: 'ffl-round', params: { roundId: round.value.id } } },
  ]
})

const sides = computed(() => {
  if (!match.value) return []
  return [
    { label: match.value.homeClubMatch?.club.name ?? 'Home', clubMatch: match.value.homeClubMatch },
    { label: match.value.awayClubMatch?.club.name ?? 'Away', clubMatch: match.value.awayClubMatch },
  ]
})

function playerName(clubMatch: { playerMatches: { id: string; player: { aflPlayer: { name: string } } }[] }, pmId: string): string {
  return clubMatch.playerMatches.find(pm => pm.id === pmId)?.player.aflPlayer.name ?? pmId
}
</script>
