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
    case 'named':             return 'Named'
    case 'playing':           return 'Playing'
    case 'played':            return 'Played'
    case 'bye':               return 'Bye'
    case 'dnp':               return 'DNP'
    case 'subbed_out':        return 'Subbed'
    case 'subbed_in':         return 'Sub In'
    case 'interchanged_out':  return 'IC\'ed'
    case 'interchanged_in':   return 'IC In'
    default: return '—'
  }
})

const statusClass = computed(() => {
  switch (props.status) {
    case 'named':             return 'bg-yellow-500/15 text-yellow-500'
    case 'playing':           return 'bg-blue-500/15 text-blue-400'
    case 'played':            return 'bg-green-500/15 text-green-500'
    case 'bye':               return 'bg-violet-500/15 text-violet-400'
    case 'dnp':               return 'bg-red-500/15 text-red-500'
    case 'subbed_out':        return 'bg-green-500/15 text-green-500'
    case 'subbed_in':         return 'bg-sky-500/15 text-sky-400'
    case 'interchanged_out':  return 'bg-green-500/15 text-green-500'
    case 'interchanged_in':   return 'bg-sky-500/15 text-sky-400'
    default: return 'bg-surface text-text-faint'
  }
})
</script>
