<template>
  <span
    class="inline-block rounded px-1.5 py-0.5 text-xs font-medium"
    :class="statusClass"
  >
    {{ label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ status: string | null }>()

const label = computed(() => {
  switch (props.status) {
    case 'named':        return 'Named'
    case 'playing':      return 'Playing'
    case 'played':       return 'Played'
    case 'dnp':          return 'DNP'
    case 'subbed':       return 'Subbed'
    case 'interchanged': return 'IC'
    default: return '—'
  }
})

const statusClass = computed(() => {
  switch (props.status) {
    case 'named':        return 'bg-yellow-500/15 text-yellow-500'
    case 'playing':      return 'bg-blue-500/15 text-blue-400'
    case 'played':       return 'bg-green-500/15 text-green-500'
    case 'dnp':          return 'bg-red-500/15 text-red-500'
    case 'subbed':       return 'bg-orange-500/15 text-orange-500'
    case 'interchanged': return 'bg-purple-500/15 text-purple-400'
    default: return 'bg-surface text-text-faint'
  }
})
</script>
