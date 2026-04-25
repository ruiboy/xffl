<template>
  <nav class="flex flex-wrap gap-2">
    <router-link
      v-for="round in rounds"
      :key="round.id"
      :to="{ name: 'afl-round', params: { seasonId, roundId: round.id } }"
      class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors"
      :class="activeRoundId === round.id
        ? 'bg-active text-active-text'
        : round.id === liveRoundId
          ? ['ring-2 ring-active ring-offset-2 ring-offset-surface', 'bg-control text-text-muted hover:bg-control-hover hover:text-text']
          : 'bg-control text-text-muted hover:bg-control-hover hover:text-text'"
    >
      {{ round.name === 'Opening Round' ? '0' : round.name.replace(/^Round\s+/i, '') }}
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

interface Round {
  id: string
  name: string
}

defineProps<{
  rounds: Round[]
  seasonId: string
  liveRoundId: string
}>()

const route = useRoute()
const activeRoundId = computed(() => route.params.roundId as string | undefined)
</script>
