<template>
  <div>
    <router-link to="/" class="text-sm text-gray-500 hover:text-gray-300 transition-colors">
      ← Seasons
    </router-link>

    <div v-if="loading" class="mt-6 text-gray-400">Loading season…</div>
    <div v-else-if="error" class="mt-6 text-red-400">{{ error.message }}</div>
    <template v-else-if="result?.aflSeason">
      <h1 class="text-2xl font-bold mt-4 mb-6">{{ result.aflSeason.name }}</h1>

      <div v-for="round in result.aflSeason.rounds" :key="round.id" class="mb-6">
        <h2 class="text-lg font-semibold text-gray-300 mb-3">{{ round.name }}</h2>
        <div class="space-y-2">
          <router-link
            v-for="match in round.matches"
            :key="match.id"
            :to="{ name: 'match', params: { seasonId: props.seasonId, matchId: match.id } }"
            class="flex items-center justify-between rounded-lg border border-gray-800 px-4 py-3 hover:border-gray-600 transition-colors"
          >
            <span class="font-medium">
              {{ match.homeClubMatch?.club.name ?? '—' }}
              <span class="text-gray-500 mx-2">v</span>
              {{ match.awayClubMatch?.club.name ?? '—' }}
            </span>
            <span v-if="match.result" class="text-sm text-gray-400">
              {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
            </span>
          </router-link>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useQuery } from '@vue/apollo-composable'
import { GET_SEASON } from '../api/queries'

const props = defineProps<{ seasonId: string }>()
const { result, loading, error } = useQuery(GET_SEASON, () => ({ id: props.seasonId }))
</script>
