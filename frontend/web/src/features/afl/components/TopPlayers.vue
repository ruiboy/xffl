<template>
  <div>
    <h3 class="text-sm font-semibold text-text-faint mb-2">{{ label }}</h3>
    <ol class="space-y-1">
      <li
        v-for="(entry, index) in players"
        :key="index"
        class="flex items-center justify-between text-sm gap-2"
      >
        <span class="flex items-center gap-1.5 min-w-0">
          <img :src="clubLogoUrl(entry.club)" :alt="entry.club" class="w-4 h-4 object-contain shrink-0" />
          <router-link
            v-if="entry.matchId"
            :to="{ name: 'afl-match', params: { matchId: entry.matchId }, query: entry.pmId ? { highlight: entry.pmId } : undefined }"
            class="font-medium hover:text-text-muted transition-colors truncate"
          >{{ entry.name }}</router-link>
          <span v-else class="font-medium truncate">{{ entry.name }}</span>
        </span>
        <span class="tabular-nums font-semibold shrink-0">{{ entry.value }}</span>
      </li>
    </ol>
  </div>
</template>

<script setup lang="ts">
import { clubLogoUrl } from '../utils/clubLogos'

interface PlayerEntry {
  name: string
  club: string
  value: number
  matchId?: string
  pmId?: string
}

defineProps<{
  label: string
  players: PlayerEntry[]
}>()
</script>
