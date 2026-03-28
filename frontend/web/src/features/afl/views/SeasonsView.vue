<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">AFL Seasons</h1>

    <div v-if="loading" class="text-gray-400">Loading seasons…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <div v-else class="space-y-2">
      <router-link
        v-for="season in result?.aflSeasons"
        :key="season.id"
        :to="{ name: 'season', params: { seasonId: season.id } }"
        class="block rounded-lg border border-gray-800 px-4 py-3 hover:border-gray-600 transition-colors"
      >
        {{ season.name }}
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useQuery } from '@vue/apollo-composable'
import { GET_SEASONS } from '../api/queries'

const { result, loading, error } = useQuery(GET_SEASONS)
</script>
