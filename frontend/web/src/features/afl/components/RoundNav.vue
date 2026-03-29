<template>
  <nav class="flex flex-wrap gap-2">
    <router-link
      v-for="round in rounds"
      :key="round.id"
      :to="{ name: 'round', params: { seasonId, roundId: round.id } }"
      class="rounded px-3 py-1 text-sm transition-colors"
      :class="round.id === currentRoundId
        ? 'bg-gray-800 text-white'
        : 'bg-gray-100 text-gray-600 hover:bg-gray-200 hover:text-gray-900'"
    >
      {{ round.name }}
    </router-link>
  </nav>
</template>

<script setup lang="ts">
interface Round {
  id: string
  name: string
}

defineProps<{
  rounds: Round[]
  currentRoundId: string
  seasonId: string
}>()
</script>
