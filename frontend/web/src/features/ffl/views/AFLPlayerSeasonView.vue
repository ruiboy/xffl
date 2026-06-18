<template>
  <div>
    <Breadcrumb v-if="playerSeason" :items="breadcrumbs" />

    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="playerSeason">

      <!-- Header -->
      <div class="mb-6">
        <h1 class="text-2xl font-bold text-text">{{ playerSeason.player.name }}</h1>
        <p class="text-sm text-text-muted mt-0.5">{{ playerSeason.clubSeason.club.name }} · {{ playerSeason.clubSeason.season.name }}</p>
      </div>

      <!-- Averages bar -->
      <div v-if="playedMatches.length > 0" class="mb-6 flex flex-wrap gap-4">
        <div v-for="col in statCols" :key="col.key" class="flex flex-col items-center gap-0.5">
          <span class="text-[10px] uppercase tracking-wide text-text-faint font-medium">{{ col.label }}</span>
          <span class="text-lg font-semibold tabular-nums text-text">{{ avg(col.key) }}</span>
        </div>
        <div class="flex flex-col items-center gap-0.5">
          <span class="text-[10px] uppercase tracking-wide text-text-faint font-medium">★</span>
          <span class="text-lg font-semibold tabular-nums text-active">{{ fflStarAvg() }}</span>
        </div>
        <div class="flex flex-col items-center gap-0.5 ml-2 pl-2 border-l border-border">
          <span class="text-[10px] uppercase tracking-wide text-text-faint font-medium">Games</span>
          <span class="text-lg font-semibold tabular-nums text-text">{{ playedMatches.length }}</span>
        </div>
      </div>

      <!-- AFL stats table -->
      <section class="mb-8">
        <h2 class="text-sm font-semibold uppercase tracking-wide text-text-muted mb-3">AFL Stats</h2>
        <div v-if="sortedMatches.length === 0" class="text-text-faint text-sm">No matches recorded.</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-border text-left text-text-muted">
                <th class="py-2 pr-3 font-medium">Round</th>
                <th class="py-2 pr-4 font-medium">Vs</th>
                <th v-for="col in statCols" :key="col.key" class="py-2 px-2 font-medium text-right w-10">{{ col.label }}</th>
                <th class="py-2 px-2 font-medium text-right w-10">★</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="m in sortedMatches"
                :key="m.id"
                class="border-b border-border-subtle"
                :class="m.status === 'dnp' ? 'opacity-40' : 'hover:bg-surface-hover'"
              >
                <td class="py-2 pr-3 tabular-nums text-text-muted text-xs">{{ m.clubMatch.match.round.name }}</td>
                <td class="py-2 pr-4 text-sm">{{ opponent(m) }}</td>
                <td v-for="col in statCols" :key="col.key" class="py-2 px-2 text-right tabular-nums">
                  <span v-if="m.status !== 'dnp'">{{ m[col.key] }}</span>
                  <span v-else class="text-text-faint">—</span>
                </td>
                <td class="py-2 px-2 text-right tabular-nums font-medium text-active">
                  <span v-if="m.status !== 'dnp'">{{ fflStar(m) }}</span>
                  <span v-else class="text-text-faint">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- FFL history -->
      <section>
        <h2 class="text-sm font-semibold uppercase tracking-wide text-text-muted mb-3">FFL History</h2>
        <div v-if="stintsLoading" class="text-text-faint text-sm">Loading...</div>
        <div v-else-if="stints.length === 0" class="text-text-faint text-sm">Not in any FFL squad this season.</div>
        <div v-else class="flex flex-col gap-6">
          <div v-for="stint in stints" :key="stint.id">
            <div class="flex items-center gap-2 mb-2">
              <span class="font-semibold text-text">{{ stint.club.name }}</span>
              <span v-if="stintRoundRange(stint)" class="text-xs text-text-muted">{{ stintRoundRange(stint) }}</span>
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-border text-left text-text-muted">
                    <th class="py-1.5 pr-3 font-medium">Round</th>
                    <th class="py-1.5 px-2 font-medium">Pos</th>
                    <th class="py-1.5 px-2 font-medium text-right">*</th>
                    <th class="py-1.5 px-2 font-medium text-text-faint text-xs">Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="pm in sortedStintMatches(stint)"
                    :key="pm.id"
                    class="border-b border-border-subtle"
                    :class="pm.aflStatus === 'dnp' ? 'opacity-40' : 'hover:bg-surface-hover'"
                  >
                    <td class="py-1.5 pr-3 text-xs text-text-muted tabular-nums">{{ fflMatchRound(pm) }}</td>
                    <td class="py-1.5 px-2">
                      <span v-if="pm.position" :class="positionColor(pm.position)" class="font-medium text-xs">
                        {{ positionLetter(pm.position) }}
                      </span>
                      <span v-else class="text-text-faint">—</span>
                    </td>
                    <td class="py-1.5 px-2 text-right tabular-nums font-medium text-active">
                      <span v-if="pm.aflStatus !== 'dnp'">{{ pm.score }}</span>
                      <span v-else class="text-text-faint">—</span>
                    </td>
                    <td class="py-1.5 px-2 text-xs text-text-faint">{{ pm.aflStatus ?? pm.status ?? '—' }}</td>
                  </tr>
                </tbody>
                <tfoot v-if="stintPlayedCount(stint) > 0">
                  <tr class="border-t border-border text-text-muted text-xs">
                    <td class="py-1.5 pr-3" colspan="2">avg</td>
                    <td class="py-1.5 px-2 text-right tabular-nums font-medium text-active">{{ stintAvg(stint) }}</td>
                    <td></td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
        </div>
      </section>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_AFL_PLAYER_SEASON_STATS, GET_FFL_PLAYER_STINTS } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'
import { POSITION_LETTERS, POSITION_COLORS } from '../utils/position'

const props = defineProps<{ aflPlayerSeasonId: string }>()

// --- AFL stats query ---

const { result: aflResult, loading, error } = useQuery(
  GET_AFL_PLAYER_SEASON_STATS,
  () => ({ id: props.aflPlayerSeasonId }),
)

const playerSeason = computed(() => aflResult.value?.aflPlayerSeason ?? null)

// --- FFL stints query ---

const { result: fflResult, loading: stintsLoading } = useQuery(
  GET_FFL_PLAYER_STINTS,
  () => ({ aflPlayerSeasonId: props.aflPlayerSeasonId }),
)

const stints = computed(() => fflResult.value?.fflPlayerSeasonsByAflPlayerSeason ?? [])

// --- Breadcrumb ---

const breadcrumbs = computed(() => {
  if (!playerSeason.value) return []
  return [
    { label: 'FFL', to: { name: 'home' } },
    { label: playerSeason.value.clubSeason.season.name },
    { label: playerSeason.value.player.name },
  ]
})

// --- AFL stats ---

interface AFLPlayerMatch {
  id: string
  status: string
  kicks: number
  handballs: number
  marks: number
  tackles: number
  hitouts: number
  goals: number
  behinds: number
  score: number
  clubMatch: {
    club: { id: string; name: string }
    match: {
      round: { id: string; name: string }
      homeClubMatch: { club: { id: string; name: string } }
      awayClubMatch: { club: { id: string; name: string } }
    }
  }
}

const statCols = [
  { key: 'kicks' as const,    label: 'K' },
  { key: 'handballs' as const, label: 'HB' },
  { key: 'marks' as const,    label: 'M' },
  { key: 'hitouts' as const,  label: 'HO' },
  { key: 'tackles' as const,  label: 'T' },
  { key: 'goals' as const,    label: 'G' },
  { key: 'behinds' as const,  label: 'B' },
]

type StatKey = typeof statCols[number]['key']

const sortedMatches = computed((): AFLPlayerMatch[] => {
  const ms = playerSeason.value?.matches ?? []
  return [...ms].sort((a, b) => parseInt(a.clubMatch.match.round.id) - parseInt(b.clubMatch.match.round.id))
})

const playedMatches = computed(() =>
  sortedMatches.value.filter(m => m.status === 'played' || m.status === 'playing'),
)

function avg(key: StatKey): string {
  const n = playedMatches.value.length
  if (n === 0) return '—'
  const sum = playedMatches.value.reduce((s, m) => s + m[key], 0)
  return (sum / n).toFixed(1)
}

function fflStar(m: AFLPlayerMatch): number {
  return m.goals * 5 + m.kicks + m.handballs + m.marks * 2 + m.tackles * 4
}

function fflStarAvg(): string {
  const n = playedMatches.value.length
  if (n === 0) return '—'
  const sum = playedMatches.value.reduce((s, m) => s + fflStar(m), 0)
  return (sum / n).toFixed(1)
}

function opponent(m: AFLPlayerMatch): string {
  const myId = playerSeason.value?.clubSeason.club.id
  const home = m.clubMatch.match.homeClubMatch.club
  const away = m.clubMatch.match.awayClubMatch.club
  return home.id === myId ? away.name : home.name
}

// --- FFL stints ---

interface FFLPlayerMatchData {
  id: string
  position: string | null
  status: string | null
  aflStatus: string | null
  score: number
  aflPlayerMatch: {
    id: string
    clubMatch: {
      match: {
        round: { id: string; name: string }
      }
    }
  } | null
}

interface FFLStint {
  id: string
  club: { id: string; name: string }
  fromRoundId: string | null
  toRoundId: string | null
  playerMatches: FFLPlayerMatchData[]
}

function fflMatchRound(pm: FFLPlayerMatchData): string {
  if (pm.aflPlayerMatch) return pm.aflPlayerMatch.clubMatch.match.round.name
  if (pm.aflStatus === 'bye') return 'Bye'
  return '—'
}

function fflMatchRoundId(pm: FFLPlayerMatchData): number {
  if (pm.aflPlayerMatch) return parseInt(pm.aflPlayerMatch.clubMatch.match.round.id)
  return 0
}

function sortedStintMatches(stint: FFLStint): FFLPlayerMatchData[] {
  return [...stint.playerMatches].sort((a, b) => fflMatchRoundId(a) - fflMatchRoundId(b))
}

function stintRoundRange(stint: FFLStint): string {
  const sorted = sortedStintMatches(stint)
  const first = sorted.find(pm => fflMatchRound(pm) !== '—')
  const last = [...sorted].reverse().find(pm => fflMatchRound(pm) !== '—')
  if (!first) return ''
  const from = fflMatchRound(first)
  const to = last ? fflMatchRound(last) : from
  return from === to ? from : `${from} – ${to}`
}

function stintPlayedCount(stint: FFLStint): number {
  return stint.playerMatches.filter(pm => pm.aflStatus !== 'dnp').length
}

function stintAvg(stint: FFLStint): string {
  const played = stint.playerMatches.filter(pm => pm.aflStatus !== 'dnp')
  if (played.length === 0) return '—'
  const sum = played.reduce((s, pm) => s + pm.score, 0)
  return (sum / played.length).toFixed(1)
}

function positionLetter(pos: string): string {
  return POSITION_LETTERS[pos] ?? pos
}

function positionColor(pos: string): string {
  return POSITION_COLORS[pos] ?? 'text-text'
}
</script>
