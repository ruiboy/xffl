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
    <!-- Tab: Seasons + Fixtures                     -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'seasons-fixtures'" class="space-y-6">
      <SeasonBuilder @created="onSeasonCreated" />
      <FixtureBuilder :initial-season-id="createdSeasonId" />
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
import FixtureBuilder from '../components/FixtureBuilder.vue'
import { RECALCULATE_AFL_LADDER, RECALCULATE_FFL_LADDER } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { GET_AFL_LIVE_ROUND } from '@/features/afl/api/queries'

const { liveSeasonId } = useFflState()
const route = useRoute()

// ---- Tabs ----
const tabs = [
  { id: 'seasons-fixtures', label: 'Seasons + Fixtures' },
  { id: 'calculate', label: 'Calculate' },
]
const activeTab = ref((route.query.tab as string) || 'seasons-fixtures')

// Hand a just-created season to the fixture builder so it opens ready to edit.
const createdSeasonId = ref<string | null>(null)
function onSeasonCreated(seasonId: string) {
  createdSeasonId.value = seasonId
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
