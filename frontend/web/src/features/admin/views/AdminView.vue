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

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Calculate                              -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'calculate'" class="space-y-6 max-w-lg">
      <!-- AFL Ladder -->
      <div class="rounded-xl border border-border bg-surface-raised p-5">
        <h2 class="text-sm font-semibold mb-1">AFL Ladder</h2>
        <p class="text-xs text-text-faint mb-4">Rebuilds AFL club season standings from all final matches for the current season.</p>
        <div class="flex items-center gap-3">
          <button
            @click="recalculateAFLLadder"
            :disabled="recalcAflLoading || !aflSeasonId"
            class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ recalcAflLoading ? 'Recalculating…' : 'Recalculate AFL Ladder' }}</button>
          <span v-if="recalcAflDone" class="text-sm text-green-500">Done</span>
          <span v-if="recalcAflError" class="text-sm text-red-400">{{ recalcAflError }}</span>
        </div>
      </div>

      <!-- FFL Ladder -->
      <div class="rounded-xl border border-border bg-surface-raised p-5">
        <h2 class="text-sm font-semibold mb-1">FFL Ladder</h2>
        <p class="text-xs text-text-faint mb-4">Rebuilds FFL club season standings from all final matches for the current season.</p>
        <div class="flex items-center gap-3">
          <button
            @click="recalculateFFLLadder"
            :disabled="recalcFflLoading || !liveSeasonId"
            class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ recalcFflLoading ? 'Recalculating…' : 'Recalculate FFL Ladder' }}</button>
          <span v-if="recalcFflDone" class="text-sm text-green-500">Done</span>
          <span v-if="recalcFflError" class="text-sm text-red-400">{{ recalcFflError }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation } from '@vue/apollo-composable'
import SeasonBuilder from '../components/SeasonBuilder.vue'
import SeasonsList from '../components/SeasonsList.vue'
import FixtureBuilder from '../components/FixtureBuilder.vue'
import { RECALCULATE_AFL_LADDER, RECALCULATE_FFL_LADDER } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { GET_AFL_LIVE_ROUND } from '@/features/afl/api/queries'

const { liveSeasonId } = useFflState()
const route = useRoute()

// ---- Tabs ----
const tabs = [
  { id: 'seasons', label: 'Seasons' },
  { id: 'fixtures', label: 'Fixtures' },
  { id: 'calculate', label: 'Calculate' },
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

// ════════════════════════════════════════════
// Calculate
// ════════════════════════════════════════════

const { result: liveRoundResult } = useQuery(GET_AFL_LIVE_ROUND)
const aflSeasonId = computed(() => liveRoundResult.value?.aflLiveRound?.round?.season?.id ?? '')

const recalcAflLoading = ref(false)
const recalcAflError = ref('')
const recalcAflDone = ref(false)
const recalcFflLoading = ref(false)
const recalcFflError = ref('')
const recalcFflDone = ref(false)

const { mutate: recalcAflMutation } = useMutation(RECALCULATE_AFL_LADDER)
const { mutate: recalcFflMutation } = useMutation(RECALCULATE_FFL_LADDER)

async function recalculateAFLLadder() {
  recalcAflLoading.value = true
  recalcAflError.value = ''
  recalcAflDone.value = false
  try {
    await recalcAflMutation({ seasonId: aflSeasonId.value })
    recalcAflDone.value = true
  } catch (e: any) {
    recalcAflError.value = e.message ?? 'Failed'
  } finally {
    recalcAflLoading.value = false
  }
}

async function recalculateFFLLadder() {
  recalcFflLoading.value = true
  recalcFflError.value = ''
  recalcFflDone.value = false
  try {
    await recalcFflMutation({ seasonId: liveSeasonId.value })
    recalcFflDone.value = true
  } catch (e: any) {
    recalcFflError.value = e.message ?? 'Failed'
  } finally {
    recalcFflLoading.value = false
  }
}
</script>
