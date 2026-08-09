<template>
  <div>
    <Breadcrumb :items="[{ label: 'FFL', to: { name: 'ffl-ladder' } }, { label: 'Seasons' }]" />
    <h1 class="text-2xl font-bold mb-6">Seasons</h1>

    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <ul v-else class="divide-y divide-border-subtle">
      <li v-for="season in seasons" :key="season.id">
        <router-link
          :to="{ name: 'ffl-season', params: { seasonId: season.id } }"
          class="flex items-center justify-between py-3 text-sm hover:text-active transition-colors"
        >
          {{ season.name }}
        </router-link>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASONS } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'

const { result, loading, error } = useQuery(GET_FFL_SEASONS)

const seasons = computed<{ id: string; name: string }[]>(
  () => [...(result.value?.fflSeasons ?? [])].sort((a, b) => b.name.localeCompare(a.name)),
)
</script>
