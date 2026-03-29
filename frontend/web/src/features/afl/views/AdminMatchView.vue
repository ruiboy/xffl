<template>
  <div>
    <div v-if="loading" class="text-gray-400">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-8">
        <p class="text-sm text-gray-500 mb-2">Admin</p>
        <h1 class="text-2xl font-bold">
          {{ match.homeClubMatch?.club.name ?? '—' }}
          <span class="text-gray-500 mx-2">v</span>
          {{ match.awayClubMatch?.club.name ?? '—' }}
        </h1>
        <p v-if="match.venue" class="text-sm text-gray-400 mt-1">{{ match.venue }}</p>
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
import { GET_MATCH } from '../api/queries'
import { UPDATE_PLAYER_MATCH } from '../api/mutations'
import PlayerStatsTable from '../components/PlayerStatsTable.vue'

const props = defineProps<{ seasonId: string; matchId: string }>()

const { result, loading, error } = useQuery(GET_MATCH, () => ({ seasonId: props.seasonId }))

const match = computed(() => {
  const season = result.value?.aflSeason
  if (!season) return null
  for (const round of season.rounds) {
    const found = round.matches.find((m: { id: string }) => m.id === props.matchId)
    if (found) return found
  }
  return null
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
