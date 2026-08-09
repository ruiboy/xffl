<template>
  <nav class="flex flex-wrap gap-2">
    <!-- Round pills -->
    <router-link
      v-for="round in rounds"
      :key="round.id"
      :to="toRound ? toRound(round) : { name: 'ffl-round', params: { roundId: round.id } }"
      :title="round.name"
      class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium transition-colors"
      :class="[
        effectiveActiveId === round.id
          ? 'bg-active text-active-text'
          : 'bg-control text-text-muted hover:bg-control-hover hover:text-text',
        round.id === liveRoundId ? 'ring-2 ring-active ring-offset-2 ring-offset-surface' : '',
      ]"
    >
      {{ roundPillLabel(round.name) }}
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
</script>
