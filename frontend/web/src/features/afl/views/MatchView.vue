<template>
  <div>
    <div v-if="loading" class="text-gray-400">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-8">
        <h1 class="text-2xl font-bold flex items-center gap-3">
          <img v-if="match.homeClubMatch" :src="clubLogoUrl(match.homeClubMatch.club.name)" :alt="match.homeClubMatch.club.name" class="w-10 h-10 object-contain" />
          {{ match.homeClubMatch?.club.name ?? '—' }}
          <span class="text-gray-400 mx-1">v</span>
          <img v-if="match.awayClubMatch" :src="clubLogoUrl(match.awayClubMatch.club.name)" :alt="match.awayClubMatch.club.name" class="w-10 h-10 object-contain" />
          {{ match.awayClubMatch?.club.name ?? '—' }}
        </h1>
        <p v-if="match.venue" class="text-sm text-gray-500 mt-1">{{ match.venue }}</p>
        <p v-if="match.result" class="text-lg font-semibold mt-2">
          {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
        </p>
        <router-link
          :to="{ name: 'admin-match', params: { seasonId: props.seasonId, matchId: props.matchId } }"
          class="inline-block mt-3 text-sm text-gray-400 hover:text-gray-700 transition-colors"
        >
          Edit stats
        </router-link>
      </div>

      <div v-for="side in sides" :key="side.label" class="mb-10">
        <h2 class="text-lg font-semibold mb-3">{{ side.label }}</h2>
        <PlayerStatsTable
          v-if="side.clubMatch"
          :club-match="side.clubMatch"
          :readonly="true"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_MATCH } from '../api/queries'
import PlayerStatsTable from '../components/PlayerStatsTable.vue'
import { clubLogoUrl } from '../utils/clubLogos'

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
</script>
