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
          <div class="flex items-center justify-between mb-1">
            <h2 class="text-lg font-semibold">{{ side.label }}</h2>
            <router-link
              v-if="isMyClubMatch && side.clubMatch?.club.id === selectedClubId"
              :to="{ name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: matchData!.roundId } }"
              class="text-xs text-active hover:text-active-hover transition-colors"
            >
              Build Team →
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
