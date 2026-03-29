<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-border text-left text-text-muted">
          <th class="py-2 pr-4 font-medium w-8">#</th>
          <th class="py-2 pr-4 font-medium">Club</th>
          <th class="py-2 px-2 font-medium text-right">P</th>
          <th class="py-2 px-2 font-medium text-right">W</th>
          <th class="py-2 px-2 font-medium text-right">L</th>
          <th class="py-2 px-2 font-medium text-right">D</th>
          <th class="py-2 px-2 font-medium text-right">F</th>
          <th class="py-2 px-2 font-medium text-right">A</th>
          <th class="py-2 px-2 font-medium text-right">Pts</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(entry, index) in ladder"
          :key="entry.id"
          class="border-b border-border-subtle hover:bg-surface-hover"
        >
          <td class="py-2 pr-4 tabular-nums text-text-faint">{{ index + 1 }}</td>
          <td class="py-2 pr-4 font-medium">
            <div class="flex items-center gap-2">
              <img :src="clubLogoUrl(entry.club.name)" :alt="entry.club.name" class="w-6 h-6 object-contain" />
              {{ entry.club.name }}
            </div>
          </td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.played }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.won }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.lost }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.drawn }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.for }}</td>
          <td class="py-2 px-2 text-right tabular-nums">{{ entry.against }}</td>
          <td class="py-2 px-2 text-right tabular-nums font-semibold">{{ entry.premiershipPoints }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { clubLogoUrl } from '../utils/clubLogos'

interface LadderEntry {
  id: string
  club: { id: string; name: string }
  played: number
  won: number
  lost: number
  drawn: number
  for: number
  against: number
  premiershipPoints: number
}

defineProps<{ ladder: LadderEntry[] }>()
</script>
