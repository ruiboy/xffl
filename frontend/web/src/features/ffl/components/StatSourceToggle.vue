<template>
  <div class="flex items-center gap-3">
    <span
      class="text-xs text-text-faint whitespace-nowrap"
      :title="`↑/↓ = last ${LAST_N} form at least ${TREND_PCT * 100}% above/below season average`"
    >
      <span class="text-green-400">↑</span>/<span class="text-red-400">↓</span> = Last {{ LAST_N }} form
    </span>
    <div class="flex items-center rounded border border-border overflow-hidden text-[10px] text-text-faint">
      <button
        class="px-1.5 py-0.5 transition-colors"
        :class="statSource === 'season' ? 'bg-control text-text' : 'hover:text-text'"
        title="Use season averages"
        @click="setStatSource('season')"
      >Season</button>
      <button
        class="px-1.5 py-0.5 transition-colors"
        :class="statSource === 'form' ? 'bg-control text-text' : 'hover:text-text'"
        :title="`Use last ${LAST_N} averages`"
        @click="setStatSource('form')"
      >Last {{ LAST_N }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { LAST_N, TREND_PCT } from '../utils/playerStats'
import { useStatSource } from '../composables/useStatSource'

const { statSource, setStatSource } = useStatSource()
</script>
