<template>
  <nav class="flex flex-wrap gap-2">
    <!-- Round pills -->
    <router-link
      v-for="round in rounds"
      :key="round.id"
      :to="toRound ? toRound(round) : { name: 'ffl-round', params: { roundId: round.id } }"
      :title="round.name"
      class="relative w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors"
      :class="effectiveActiveId === round.id
        ? 'bg-active text-active-text'
        : round.id === liveRoundId
          ? 'border border-active text-active hover:bg-active/10'
          : 'bg-control text-text-muted hover:bg-control-hover hover:text-text'"
    >
      {{ roundPillLabel(round.name) }}
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
import { roundPillLabel } from '@/utils/roundLabel'

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
