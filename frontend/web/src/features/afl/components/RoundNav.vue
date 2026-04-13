<template>
  <nav class="flex flex-wrap gap-2">
    <router-link
      :to="{ name: 'afl-home' }"
      class="w-8 h-8 rounded-full flex items-center justify-center transition-colors"
      :class="isHome
        ? 'bg-active text-active-text'
        : 'bg-control text-text-muted hover:bg-control-hover hover:text-text'"
      title="Ladder"
    >
      <svg class="w-4 h-4" viewBox="0 0 12 14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
        <line x1="2" y1="1" x2="2" y2="13"/>
        <line x1="10" y1="1" x2="10" y2="13"/>
        <line x1="2" y1="4" x2="10" y2="4"/>
        <line x1="2" y1="7" x2="10" y2="7"/>
        <line x1="2" y1="10" x2="10" y2="10"/>
      </svg>
    </router-link>
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

const props = defineProps<{
  rounds: Round[]
  seasonId: string
  liveRoundId: string
}>()

const route = useRoute()
const isHome = computed(() => route.name === 'afl-home')
const activeRoundId = computed(() => route.params.roundId as string | undefined)
</script>
