<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-6">
        <Breadcrumb v-if="matchData" :items="breadcrumbs" />
        <h1 class="text-2xl font-bold flex items-center gap-3">
          <img v-if="match.homeClubMatch" :src="clubLogoUrl(match.homeClubMatch.club.name)" :alt="match.homeClubMatch.club.name" class="w-10 h-10 object-contain" />
          {{ match.homeClubMatch?.club.name ?? '—' }}
          <span class="text-text-faint mx-1">v</span>
          <img v-if="match.awayClubMatch" :src="clubLogoUrl(match.awayClubMatch.club.name)" :alt="match.awayClubMatch.club.name" class="w-10 h-10 object-contain" />
          {{ match.awayClubMatch?.club.name ?? '—' }}
        </h1>
        <p v-if="match.venue" class="text-sm text-text-muted mt-1">{{ match.venue }}</p>
        <p v-if="match.result" class="text-lg font-semibold mt-2">
          {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
        </p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div v-for="side in sides" :key="side.label">
          <div class="flex items-center gap-2 mb-1">
            <img v-if="side.clubMatch" :src="clubLogoUrl(side.clubMatch.club.name)" :alt="side.clubMatch.club.name" class="w-8 h-8 object-contain" />
            <h2 class="text-lg font-semibold">
              <router-link
                v-if="side.clubMatch"
                :to="{ name: 'ffl-squad', params: { seasonId: props.seasonId, clubId: side.clubMatch.club.id } }"
                class="hover:text-active transition-colors"
              >{{ side.label }}</router-link>
              <span v-else>{{ side.label }}</span>
            </h2>
            <router-link
              v-if="isMyClubMatch && side.clubMatch?.club.id === selectedClubId"
              :to="{ name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: matchData!.roundId } }"
              title="Team Builder"
              class="rounded p-1 text-active hover:bg-active/10 transition-colors"
            >
              <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                <path d="M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437 1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008Z"/>
              </svg>
            </router-link>
          </div>
          <p class="text-sm text-text-muted mb-3">
            Fantasy score: <span class="font-semibold text-text">{{ side.clubMatch?.score ?? 0 }}</span>
          </p>
          <SquadTable v-if="side.clubMatch" :player-matches="side.clubMatch.playerMatches" />
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASON } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'
import SquadTable from '../components/SquadTable.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { useFflState } from '../composables/useFflState'

const props = defineProps<{ seasonId: string; matchId: string }>()

const { selectedClubId } = useFflState()
const { result, loading, error } = useQuery(GET_FFL_SEASON, () => ({ id: props.seasonId }))

const matchData = computed(() => {
  const season = result.value?.fflSeason
  if (!season) return null
  for (const round of season.rounds) {
    const found = round.matches.find((m: { id: string }) => m.id === props.matchId)
    if (found) return { match: found, roundId: round.id as string, roundName: round.name as string, seasonName: season.name as string }
  }
  return null
})

const breadcrumbs = computed(() => {
  if (!matchData.value) return []
  return [
    { label: 'FFL' },
    { label: matchData.value.seasonName, to: { name: 'home' } },
    { label: matchData.value.roundName, to: { name: 'ffl-round', params: { seasonId: props.seasonId, roundId: matchData.value.roundId } } },
  ]
})

const match = computed(() => matchData.value?.match ?? null)

const isMyClubMatch = computed(() => {
  if (!match.value || !selectedClubId.value) return false
  return match.value.homeClubMatch?.club.id === selectedClubId.value ||
    match.value.awayClubMatch?.club.id === selectedClubId.value
})

const sides = computed(() => {
  if (!match.value) return []
  return [
    { label: match.value.homeClubMatch?.club.name ?? 'Home', clubMatch: match.value.homeClubMatch },
    { label: match.value.awayClubMatch?.club.name ?? 'Away', clubMatch: match.value.awayClubMatch },
  ]
})
</script>
