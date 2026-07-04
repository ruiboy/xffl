<template>
  <div class="flex items-end justify-end gap-1">
    <span class="text-xs" :style="avgStyle">{{ fmtStat(avg) }}</span>
    <template v-if="last3 != null">
      <span v-if="trendIndicator" class="text-[10px]" :class="trendIndicatorCls">{{ trendIndicator }}</span>
      <span :style="last3Style">{{ fmtStat(last3) }}</span>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { fmtStat } from '../utils/playerStats'

const props = defineProps<{
  avg: number | null | undefined
  last3: number | null | undefined
  avgStyle?: Record<string, string>
  last3Style?: Record<string, string>
}>()

const trendIndicator = computed(() => {
  if (props.last3 == null || props.avg == null) return ''
  if (props.last3 === props.avg) return '='
  return props.last3 > props.avg ? '↑' : '↓'
})

const trendIndicatorCls = computed(() => {
  if (props.last3 == null || props.avg == null) return ''
  if (props.last3 === props.avg) return 'text-blue-400'
  return props.last3 > props.avg ? 'text-green-400' : 'text-red-400'
})
</script>
