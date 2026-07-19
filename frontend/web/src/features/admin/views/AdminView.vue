<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">Admin</h1>
    </div>

    <!-- Tab navigation -->
    <div class="flex gap-1 mb-6 border-b border-border">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px"
        :class="activeTab === tab.id
          ? 'border-active text-active'
          : 'border-transparent text-text-muted hover:text-text'"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Seasons                                -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'seasons'" class="space-y-6">
      <SeasonsList ref="seasonsList" @create="showCreate = true" @edit-fixtures="onEditFixtures" />
    </div>

    <!-- Create season dialog -->
    <div
      v-if="showCreate"
      class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4 pt-24"
      @click.self="showCreate = false"
    >
      <div class="w-full max-w-2xl rounded-lg border border-border bg-surface shadow-xl">
        <div class="flex items-center justify-between border-b border-border px-4 py-3">
          <h3 class="text-sm font-semibold">Create season</h3>
          <button @click="showCreate = false" class="text-text-faint hover:text-text" title="Close">✕</button>
        </div>
        <div class="p-4">
          <SeasonBuilder @created="onSeasonCreated" />
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Fixtures                               -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'fixtures'">
      <FixtureBuilder :initial-season-id="fixtureSeasonId" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import SeasonBuilder from '../components/SeasonBuilder.vue'
import SeasonsList from '../components/SeasonsList.vue'
import FixtureBuilder from '../components/FixtureBuilder.vue'

const route = useRoute()

// ---- Tabs ----
const tabs = [
  { id: 'seasons', label: 'Seasons' },
  { id: 'fixtures', label: 'Fixtures' },
]
const activeTab = ref((route.query.tab as string) || 'seasons')

// Create season is a dialog off the Seasons list; the season a row's "Edit"
// link opens is handed to the fixture builder.
const showCreate = ref(false)
const fixtureSeasonId = ref<string | null>(null)
const seasonsList = ref<{ refetch: () => void } | null>(null)

function onSeasonCreated() {
  showCreate.value = false
  seasonsList.value?.refetch()
}

function onEditFixtures(seasonId: string) {
  fixtureSeasonId.value = seasonId
  activeTab.value = 'fixtures'
}
</script>
