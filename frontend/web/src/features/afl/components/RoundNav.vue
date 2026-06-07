<template>
  <nav class="flex flex-wrap gap-2">
    <!-- Ladder pill -->
    <router-link
      :to="{ name: 'afl-home' }"
      class="w-8 h-8 rounded-full flex items-center justify-center transition-colors bg-control text-text-muted hover:bg-control-hover hover:text-text"
      title="Ladder"
    >
      <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="13" width="4" height="8" rx="0.5" />
        <rect x="9.5" y="8" width="4" height="13" rx="0.5" />
        <rect x="16" y="3" width="4" height="18" rx="0.5" />
      </svg>
    </router-link>

    <!-- Round pills -->
    <router-link
      v-for="round in rounds"
      :key="round.id"
      :to="toRound ? toRound(round) : { name: 'afl-round', params: { roundId: round.id } }"
      class="relative w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors"
      :class="effectiveActiveId === round.id
        ? 'bg-active text-active-text'
        : round.id === liveRoundId
          ? ['ring-2 ring-active ring-offset-2 ring-offset-surface', 'bg-control text-text-muted hover:bg-control-hover hover:text-text']
          : 'bg-control text-text-muted hover:bg-control-hover hover:text-text'"
    >
      {{ round.name === 'Opening Round' ? '0' : round.name.replace(/^Round\s+/i, '') }}
      <!-- Pulsing dot when this is the live round and it starts today -->
      <span
        v-if="round.id === liveRoundId && isLiveToday"
        class="absolute top-0 right-0 w-2 h-2 rounded-full bg-active animate-pulse"
      />
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { RouteLocationRaw } from 'vue-router'

interface Round {
  id: string
  name: string
}

const props = defineProps<{
  rounds: Round[]
  liveRoundId: string
  liveStartDate?: string
  activeId?: string
  toRound?: (r: Round) => RouteLocationRaw
}>()

const route = useRoute()
const effectiveActiveId = computed(() => props.activeId ?? route.params.roundId as string | undefined)

const isLiveToday = computed(() => {
  if (!props.liveStartDate) return false
  const today = new Date().toISOString().slice(0, 10)
  return props.liveStartDate.slice(0, 10) === today
})
</script>
