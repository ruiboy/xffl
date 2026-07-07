<template>
  <div class="flex items-end justify-end gap-1">
    <span :class="statSource === 'season' ? '' : 'text-xs'" :style="avgStyle">{{ fmtStat(avg) }}</span>
    <template v-if="last3 != null">
      <span v-if="trendIndicator" class="text-[10px]" :class="trendIndicatorCls">{{ trendIndicator }}</span>
      <span :class="statSource === 'form' ? '' : 'text-xs'" :style="last3Style">{{ fmtStat(last3) }}</span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { fmtStat, trendDir } from '../utils/playerStats'
import { useStatSource } from '../composables/useStatSource'

const props = defineProps<{
  avg: number | null | undefined
  last3: number | null | undefined
  avgStyle?: Record<string, string>
  last3Style?: Record<string, string>
}>()

// The active source (global toggle) renders at full size, the other small.
const { statSource } = useStatSource()

// Shared trend semantics (utils/playerStats): '~' when within thresholds.
const trend = computed(() => {
  if (props.last3 == null || props.avg == null) return null
  return trendDir(props.last3, props.avg)
})

const trendIndicator = computed(() => {
  if (props.last3 == null || props.avg == null) return ''
  if (trend.value === 'up') return '↑'
  if (trend.value === 'down') return '↓'
  return '~'
})

const trendIndicatorCls = computed(() => {
  if (trend.value === 'up') return 'text-green-400'
  if (trend.value === 'down') return 'text-red-400'
  return 'text-blue-400'
})
</script>
