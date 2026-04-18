<template>
  <nav class="flex items-center gap-1.5 text-sm mb-1">
    <template v-for="(item, i) in items" :key="i">
      <span v-if="i > 0" class="text-text-faint select-none">›</span>
      <router-link
        v-if="item.to"
        :to="item.to"
        class="text-text-muted hover:text-text transition-colors"
      >{{ item.label }}</router-link>
      <span v-else-if="i === items.length - 1 && items.length > 1" class="font-semibold text-text">
        {{ item.label }}<span v-if="item.sub" class="font-normal text-text-faint"> · {{ item.sub }}</span>
      </span>
      <span v-else class="text-text-muted">{{ item.label }}</span>
    </template>
  </nav>
</template>

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'

defineProps<{
  items: Array<{ label: string; to?: RouteLocationRaw; sub?: string }>
}>()
</script>
