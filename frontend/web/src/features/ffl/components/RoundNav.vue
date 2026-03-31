<template>
  <nav class="flex flex-wrap gap-2">
    <router-link
      v-for="round in rounds"
      :key="round.id"
      :to="{ name: 'ffl-round', params: { seasonId, roundId: round.id } }"
      class="rounded px-3 py-1 text-sm transition-colors"
      :class="round.id === currentRoundId
        ? 'bg-active text-active-text'
        : 'bg-control text-text-muted hover:bg-control-hover hover:text-text'"
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
