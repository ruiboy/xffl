<template>
  <div>
    <h3 class="text-sm font-semibold text-text-faint mb-2">{{ label }}</h3>
    <ol class="space-y-1">
      <li
        v-for="(entry, index) in players"
        :key="index"
        class="flex items-center justify-between text-sm"
      >
        <span>
          <span class="text-text-muted tabular-nums mr-2">{{ index + 1 }}.</span>
          <span class="font-medium">{{ entry.name }}</span>
          <span class="text-text-muted ml-1">({{ entry.club }})</span>
        </span>
        <span class="tabular-nums font-semibold">{{ entry.value }}</span>
      </li>
    </ol>
  </div>
</template>

<script setup lang="ts">
interface PlayerEntry {
  name: string
  club: string
  value: number
}

defineProps<{
  label: string
  players: PlayerEntry[]
}>()
</script>
