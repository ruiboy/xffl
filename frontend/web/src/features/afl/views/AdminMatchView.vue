<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-6">
        <Breadcrumb v-if="match" :items="breadcrumbs" />
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

      <div v-for="side in sides" :key="side.label" class="mb-10">
        <h2 class="text-lg font-semibold mb-3">{{ side.label }}</h2>
        <PlayerStatsTable
          v-if="side.clubMatch"
          :club-match="side.clubMatch"
          :readonly="false"
          @update="handleUpdate"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_AFL_MATCH } from '../api/queries'
import { UPDATE_PLAYER_MATCH } from '../api/mutations'
import Breadcrumb from '../components/Breadcrumb.vue'
import PlayerStatsTable from '../components/PlayerStatsTable.vue'
import { clubLogoUrl } from '../utils/clubLogos'

const props = defineProps<{ seasonId: string; matchId: string }>()

const { result, loading, error } = useQuery(GET_AFL_MATCH, () => ({ matchId: props.matchId }))

const match = computed(() => result.value?.aflMatch ?? null)

const breadcrumbs = computed(() => {
  if (!match.value) return []
  const round = match.value.round
  return [
    { label: 'AFL' },
    { label: round.season.name, to: { name: 'afl-home' } },
    { label: round.name, to: { name: 'afl-round', params: { seasonId: props.seasonId, roundId: round.id } } },
  ]
})

const sides = computed(() => {
  if (!match.value) return []
  return [
    { label: match.value.homeClubMatch?.club.name ?? 'Home', clubMatch: match.value.homeClubMatch },
    { label: match.value.awayClubMatch?.club.name ?? 'Away', clubMatch: match.value.awayClubMatch },
  ]
})

const { mutate } = useMutation(UPDATE_PLAYER_MATCH)

function handleUpdate(input: { playerSeasonId: string; clubMatchId: string; [key: string]: unknown }) {
  mutate({ input })
}
</script>
