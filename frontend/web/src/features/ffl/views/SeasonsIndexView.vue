<template>
  <div>
    <Breadcrumb :items="[{ label: 'FFL', to: { name: 'ffl-ladder' } }, { label: 'Seasons' }]" />
    <h1 class="text-2xl font-bold mb-6">Seasons</h1>

    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <ul v-else class="divide-y divide-border-subtle">
      <li v-for="season in seasons" :key="season.id" class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1 py-3">
        <router-link
          :to="{ name: 'ffl-season', params: { seasonId: season.id } }"
          class="text-sm font-medium hover:text-active transition-colors"
        >
          {{ season.name }}
        </router-link>
        <div class="flex flex-wrap items-center justify-end gap-x-4 gap-y-1">
          <router-link
            v-for="entry in sortedClubs(season)"
            :key="entry.id"
            :to="{ name: 'ffl-club-season', params: { clubSeasonId: entry.id } }"
            class="flex items-center gap-1 text-xs text-text-muted hover:text-active transition-colors"
          >
            <img :src="clubLogoUrl(entry.club.name)" :alt="entry.club.name" class="w-4 h-4 object-contain" />
            {{ entry.club.name }}
          </router-link>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASONS } from '../api/queries'
import { clubLogoUrl } from '../utils/clubLogos'
import Breadcrumb from '../components/Breadcrumb.vue'

interface ClubSeasonEntry {
  id: string
  club: { id: string; name: string }
}

interface Season {
  id: string
  name: string
  ladder: ClubSeasonEntry[]
}

const { result, loading, error } = useQuery(GET_FFL_SEASONS)

const seasons = computed<Season[]>(
  () => [...(result.value?.fflSeasons ?? [])].sort((a, b) => b.name.localeCompare(a.name)),
)

function sortedClubs(season: Season): ClubSeasonEntry[] {
  return [...season.ladder].sort((a, b) => a.club.name.localeCompare(b.club.name))
}
</script>
