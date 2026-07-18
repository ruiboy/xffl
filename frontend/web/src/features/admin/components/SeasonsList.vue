<template>
  <section class="rounded-lg border border-border p-4 space-y-3">
    <div class="flex items-center justify-between">
      <h3 class="text-sm font-semibold">Seasons</h3>
      <button
        @click="emit('create')"
        class="rounded-lg border border-active bg-active px-3 py-1.5 text-sm font-medium text-active-text transition-colors"
      >New season</button>
    </div>
    <div v-if="seasons.length === 0" class="text-sm text-text-faint">No seasons yet.</div>
    <div v-else class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-border text-left text-text-muted">
            <th class="py-2 pr-4 font-medium">Name</th>
            <th class="py-2 pr-4 font-medium">Rules</th>
            <th class="py-2 pr-4 font-medium">Clubs</th>
            <th class="py-2 font-medium text-right">Fixtures</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in seasons" :key="s.id" class="border-b border-border-subtle">
            <td class="py-2 pr-4 font-medium">{{ s.name }}</td>
            <td class="py-2 pr-4 tabular-nums text-text-muted">{{ s.rulesId }}</td>
            <td class="py-2 pr-4 text-text-muted">{{ clubNames(s) }}</td>
            <td class="py-2 text-right">
              <button @click="emit('editFixtures', s.id)" class="text-sm font-medium text-active hover:underline">Edit</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASONS_ADMIN } from '../api/queries'

const emit = defineEmits<{ (e: 'create'): void; (e: 'editFixtures', seasonId: string): void }>()

const { result, refetch } = useQuery(GET_FFL_SEASONS_ADMIN)
const seasons = computed<{ id: string; name: string; rulesId: string; ladder: { club: { name: string } }[] }[]>(
  () => [...(result.value?.fflSeasons ?? [])].sort((a: any, b: any) => b.name.localeCompare(a.name)),
)

function clubNames(s: { ladder: { club: { name: string } }[] }): string {
  return [...s.ladder].map((cs) => cs.club.name).sort((a, b) => a.localeCompare(b)).join(', ')
}

// Let the parent refresh the list after a season is created.
defineExpose({ refetch })
</script>
