<template>
  <svg :width="width" :height="height" class="overflow-visible shrink-0">
    <polyline
      :points="points"
      fill="none"
      :stroke="color"
      stroke-width="1.5"
      stroke-linejoin="round"
      stroke-linecap="round"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// Values are chronological (oldest first) — the line reads left-to-right as
// time passing, same convention as the round-history charts elsewhere.
const props = withDefaults(defineProps<{ values: number[]; width?: number; height?: number; color?: string }>(), {
  width: 72,
  height: 20,
  color: '#94a3b8',
})

const points = computed(() => {
  const vals = props.values
  if (vals.length < 2) return ''
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const range = max - min || 1
  const stepX = props.width / (vals.length - 1)
  // A little vertical inset so the line doesn't clip on min/max points.
  const inset = 2
  return vals
    .map((v, i) => {
      const x = i * stepX
      const y = inset + (props.height - inset * 2) * (1 - (v - min) / range)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
})
</script>
