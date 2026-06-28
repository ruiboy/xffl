<template>
  <div>
    <div v-if="aflLoading" class="text-text-faint">Loading…</div>
    <div v-else-if="aflError" class="text-red-400">{{ aflError.message }}</div>
    <div v-else-if="fflError" class="text-red-400">{{ fflError.message }}</div>
    <div v-else-if="aflResult && !fflRound" class="text-text-muted">
      Cannot determine round to display. Consult your admin.
    </div>
    <template v-else-if="fflRound">
      <Breadcrumb :items="[{ label: 'FFL' }]" />
      <h1 class="text-2xl font-bold mb-6">{{ fflRound.season.name }}</h1>

      <RoundNav
        class="mb-8"
        :rounds="fflRound.season.rounds"
        :live-round-id="liveRoundId"
        :live-start-date="liveStartDate"
      />

      <section>
        <div class="flex items-center justify-between mb-3">
          <h2 class="text-lg font-semibold text-text-heading">Ladder</h2>
          <button
            @click="copyLadderToClipboard"
            title="Copy to Clipboard"
            class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text-muted hover:text-text hover:bg-surface-hover transition-colors"
          >
            <span class="flex items-center gap-1.5">
              <IconCopy class="w-3.5 h-3.5" />
              {{ copyLadderLabel }}
            </span>
          </button>
        </div>
        <LadderTable :ladder="fflRound.season.ladder" />
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_AFL_LIVE_ROUND, GET_FFL_ROUND_BY_AFL_ROUND } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import LadderTable from '../components/LadderTable.vue'
import RoundNav from '../components/RoundNav.vue'
import IconCopy from '../components/icons/IconCopy.vue'

const { liveRoundId, liveStartDate, setLiveRound } = useFflState()

// Step 1: get the live AFL round
const { result: aflResult, loading: aflLoading, error: aflError } = useQuery(GET_AFL_LIVE_ROUND)

const aflRoundId = computed(() => aflResult.value?.aflLiveRound?.round?.id ?? null)

// Step 2: find the FFL round linked to that AFL round (skipped until AFL round is known)
const { result: fflResult, error: fflError } = useQuery(
  GET_FFL_ROUND_BY_AFL_ROUND,
  () => ({ aflRoundId: aflRoundId.value }),
  () => ({ enabled: !!aflRoundId.value }),
)

const fflRound = computed(() => fflResult.value?.fflRoundByAflRound ?? null)

watch([fflRound, () => aflResult.value], ([round, afl]) => {
  if (!round || !afl) return
  setLiveRound(round.season.id, round.id, afl.aflLiveRound.startDate)
})

const copyLadderLabel = ref('Copy')

function formatLadderText(): string {
  const round = fflRound.value
  if (!round) return ''
  const ladder = round.season.ladder as {
    club: { name: string }; played: number; won: number; lost: number
    drawn: number; for: number; against: number; percentage: number
  }[]
  const header = `${round.season.name} LADDER — ${round.name}`
  const pad = (s: string, n: number) => s.padEnd(n)
  const rpad = (s: string, n: number) => s.padStart(n)
  const maxName = Math.max(...ladder.map(e => e.club.name.length), 4)
  const cols = `${pad('', 4)}${pad('Club', maxName + 2)} ${rpad('P', 3)} ${rpad('W', 3)} ${rpad('L', 3)} ${rpad('D', 3)} ${rpad('F', 6)} ${rpad('A', 6)} ${rpad('%', 7)}`
  const rows = ladder.map((e, i) =>
    `${rpad(String(i + 1), 2)}.  ${pad(e.club.name, maxName + 2)}${rpad(String(e.played), 3)} ${rpad(String(e.won), 3)} ${rpad(String(e.lost), 3)} ${rpad(String(e.drawn), 3)} ${rpad(String(e.for), 6)} ${rpad(String(e.against), 6)} ${rpad(e.percentage.toFixed(1), 7)}`
  )
  return [header, '', cols, ...rows].join('\n')
}

async function copyLadderToClipboard() {
  await navigator.clipboard.writeText(formatLadderText())
  copyLadderLabel.value = 'Copied!'
  setTimeout(() => { copyLadderLabel.value = 'Copy' }, 2000)
}
</script>
