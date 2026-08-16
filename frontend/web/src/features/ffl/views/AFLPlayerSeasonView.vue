<template>
  <div>
    <Breadcrumb v-if="playerSeason" :items="breadcrumbs" />

    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <NotFound v-else-if="notFound" entity="Player season" />
    <template v-else-if="playerSeason">

      <!-- Header -->
      <div class="mb-8 flex items-start gap-4">
        <img :src="aflClubLogoUrl(playerSeason.clubSeason.club.name)" class="w-12 h-12 object-contain mt-0.5 flex-shrink-0" />
        <div>
          <h1 class="text-2xl font-bold text-text">{{ playerSeason.player.name }}</h1>
          <router-link
            :to="{ name: 'ffl-afl-club-season', params: { clubSeasonId: playerSeason.clubSeason.id } }"
            class="text-sm text-text-muted mt-0.5 hover:text-text transition-colors"
          >{{ playerSeason.clubSeason.club.name }}</router-link>
          <!-- Other AFL seasons for this player, most recent first. -->
          <div v-if="otherSeasons.length > 0" class="flex flex-wrap items-center gap-2 mt-2">
            <router-link
              v-for="s in otherSeasons"
              :key="s.id"
              :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: s.id } }"
              class="flex items-center gap-1.5 rounded-full border border-border bg-surface px-2.5 py-1 text-xs text-text-muted hover:text-text hover:bg-surface-hover transition-colors"
              :title="`${s.clubName} · ${s.seasonName}`"
            >
              <img :src="aflClubLogoUrl(s.clubName)" class="w-4 h-4 object-contain" />
              <span>{{ s.seasonName }}</span>
            </router-link>
          </div>

          <div v-if="stintEvents.length > 0" class="flex flex-wrap gap-5 mt-2">
            <div v-for="(event, i) in stintEvents" :key="i" class="flex items-center gap-2">
              <img :src="fflClubLogoUrl(event.clubName)" class="w-5 h-5 object-contain" />
              <router-link
                :to="{ name: 'ffl-club-season', params: { clubSeasonId: event.clubSeasonId } }"
                class="text-base text-text font-semibold hover:text-text-muted transition-colors"
              >{{ event.clubName }}</router-link>
              <span v-if="event.from" class="text-sm text-text-faint">{{ event.from }} – {{ event.to }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Analysis -->
      <div v-if="playerSeason?.statsAll" class="mb-8">
        <p class="text-[10px] font-semibold uppercase tracking-widest text-text-faint mb-2">Season Stats</p>
          <table class="w-full table-fixed text-sm">
            <thead>
              <tr class="border-b border-border">
                <th class="py-1 text-left text-xs font-normal text-text-faint">
                  {{ playerSeason.statsAll.games }} games
                </th>
                <th v-for="col in statCols" :key="col.key" class="px-2 py-1 text-right font-medium text-text-muted">{{ col.label }}</th>
                <th class="px-2 py-1 text-right font-medium text-yellow-400">★</th>
              </tr>
            </thead>
            <tbody>
              <tr class="border-b border-border-subtle">
                <td class="py-2 text-xs text-text-faint whitespace-nowrap">Season avg</td>
                <td v-for="col in statCols" :key="col.key" class="px-2 py-2 text-right tabular-nums text-base font-bold">{{ statAvg(col.key) }}</td>
                <td class="px-2 py-2 text-right tabular-nums text-base font-bold text-yellow-400">{{ starSeasonAvg }}</td>
              </tr>
              <tr class="border-b border-border-subtle">
                <td class="py-2 text-xs text-text-faint whitespace-nowrap">Last {{ LAST_N }} avg</td>
                <td v-for="col in statCols" :key="col.key" class="px-2 py-2 text-right tabular-nums text-base font-bold">
                  <span v-if="statLastNTrend(col.key) === 'up'" class="text-xs font-normal text-green-400">↑ </span>
                  <span v-else-if="statLastNTrend(col.key) === 'down'" class="text-xs font-normal text-red-400">↓ </span>{{ statLastNAvg(col.key) }}
                </td>
                <td class="px-2 py-2 text-right tabular-nums text-base font-bold text-yellow-400">
                  <span v-if="starLastNTrend === 'up'" class="text-xs font-normal text-green-400">↑ </span>
                  <span v-else-if="starLastNTrend === 'down'" class="text-xs font-normal text-red-400">↓ </span>{{ starLastNAvg }}
                </td>
              </tr>
              <tr>
                <td class="py-2 text-xs text-text-faint whitespace-nowrap">Median</td>
                <td v-for="col in statCols" :key="col.key" class="px-2 py-2 text-right tabular-nums text-base font-bold">{{ statMedian(col.key) }}</td>
                <td class="px-2 py-2 text-right tabular-nums text-base font-bold text-yellow-400">{{ starMedian }}</td>
              </tr>
            </tbody>
          </table>
      </div>

      <!-- Merged table -->
      <p class="text-[10px] font-semibold uppercase tracking-widest text-text-faint mb-3">Match by Match</p>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border text-left text-text-muted">
              <th class="py-2 pr-3 font-medium whitespace-nowrap">Round</th>
              <th class="py-2 pr-4 font-medium">Vs</th>
              <th v-for="col in statCols" :key="col.key" class="py-2 px-2 font-medium text-right w-10">{{ col.label }}</th>
              <th class="py-2 pl-2 pr-3 font-medium text-right w-10 text-yellow-400">★</th>
              <th class="py-2 w-5"></th>
              <th class="py-2 pl-5 pr-3 font-medium bg-white/[0.03] border-l border-border">FFL Club</th>
              <th class="py-2 px-2 font-medium bg-white/[0.03]">Position</th>
              <th class="py-2 px-2 font-medium bg-white/[0.03]">Bench</th>
              <th class="py-2 px-2 font-medium bg-white/[0.03]">IC</th>
              <th class="py-2 px-2 font-medium bg-white/[0.03]">Status</th>
              <th class="py-2 pl-2 pr-1 font-medium text-right bg-white/[0.03]">Score</th>
              <th class="py-2 w-5 bg-white/[0.03]"></th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, i) in displayRows"
              :key="row.id"
              class="border-b border-border-subtle group"
              :class="row.status === 'dnp' ? 'opacity-40' : 'hover:bg-surface-hover'"
            >
              <td class="py-2 pr-3 tabular-nums text-text-muted text-xs whitespace-nowrap">{{ shortRound(row.clubMatch.match.round.name) }}</td>
              <td class="py-2 pr-4">
                <span class="inline-flex items-center gap-1.5">
                  <img :src="aflClubLogoUrl(opponentName(row))" class="w-4 h-4 object-contain" />
                  <span class="text-xs text-text-muted">{{ aflClubAbbrev(opponentName(row)) }}</span>
                </span>
              </td>
              <td v-for="col in statCols" :key="col.key" class="py-2 px-2 text-right tabular-nums">
                <span v-if="row.status !== 'dnp'" :class="statColor(col.key, row.fflPosition)" :style="roundHeat(row[col.key], col.key)">{{ row[col.key] }}</span>
                <span v-else class="text-text-faint">—</span>
              </td>
              <td class="py-2 pl-2 pr-3 text-right tabular-nums font-medium">
                <span v-if="row.status !== 'dnp'" :style="roundHeat(row.starScore, 'star')">{{ row.starScore }}</span>
                <span v-else class="text-text-faint">—</span>
              </td>
              <td class="py-2 pr-2 w-5">
                <router-link
                  :to="{ name: 'afl-match', params: { matchId: row.aflMatchId } }"
                  class="opacity-0 group-hover:opacity-100 transition-opacity text-text-faint hover:text-text"
                  title="View AFL match"
                >
                  <svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M6 3H3a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1h9a1 1 0 0 0 1-1v-3M9 2h5m0 0v5m0-5L7 9"/>
                  </svg>
                </router-link>
              </td>
              <td class="py-2 pl-5 pr-3 whitespace-nowrap border-l border-border bg-white/[0.03]">
                <span v-if="row.fflClubName" class="inline-flex items-center gap-1.5">
                  <img :src="fflClubLogoUrl(row.fflClubName)" class="w-4 h-4 object-contain" />
                  <span v-if="i === 0 || displayRows[i - 1].fflClubName !== row.fflClubName" class="text-xs text-text-muted">{{ row.fflClubName }}</span>
                </span>
                <span v-else class="text-text-faint text-xs">—</span>
              </td>
              <td class="py-2 px-2 bg-white/[0.03]">
                <span v-if="row.fflPosition" :class="POSITION_COLORS[row.fflPosition]" class="text-xs font-medium">
                  {{ POSITION_LETTERS[row.fflPosition] }}
                </span>
                <span v-else class="text-text-faint text-xs">—</span>
              </td>
              <td class="py-2 px-2 bg-white/[0.03]">
                <span v-if="row.fflBackupPositions" class="text-xs font-medium text-text-muted">
                  {{ posLetters(row.fflBackupPositions) }}
                </span>
                <span v-else class="text-text-faint text-xs">—</span>
              </td>
              <td class="py-2 px-2 bg-white/[0.03]">
                <span v-if="row.fflInterchangePosition" :class="POSITION_COLORS[row.fflInterchangePosition]" class="text-xs font-medium">
                  {{ posLetters(row.fflInterchangePosition) }}
                </span>
                <span v-else class="text-text-faint text-xs">—</span>
              </td>
              <td class="py-2 px-2 bg-white/[0.03]">
                <StatusBadge v-if="row.fflClubName" :status="row.fflAflStatus" />
              </td>
              <td class="py-2 pl-2 pr-1 text-right tabular-nums font-medium bg-white/[0.03]">
                <span v-if="row.fflScore !== null" class="text-active">{{ row.fflScore }}</span>
                <span v-else class="text-text-faint">—</span>
              </td>
              <td class="py-2 pr-2 w-5 bg-white/[0.03]">
                <router-link
                  v-if="row.fflMatchId"
                  :to="{ name: 'ffl-match', params: { matchId: row.fflMatchId } }"
                  class="opacity-0 group-hover:opacity-100 transition-opacity text-text-faint hover:text-text"
                  title="View FFL match"
                >
                  <svg class="w-3.5 h-3.5" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M6 3H3a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1h9a1 1 0 0 0 1-1v-3M9 2h5m0 0v5m0-5L7 9"/>
                  </svg>
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { useNotFound } from '@/composables/useNotFound'
import NotFound from '@/components/NotFound.vue'
import { GET_AFL_PLAYER_SEASON_STATS, GET_FFL_PLAYER_STINTS } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { POSITION_LETTERS, POSITION_COLORS } from '../utils/position'
import { fmtStat, statCols, starScore, trendDir, LAST_N, type StatKey } from '../utils/playerStats'
import { heatStyle } from '@/utils/heatmap'
import { useTheme } from '@/composables/useTheme'
import { clubLogoUrl as aflClubLogoUrl } from '@/features/afl/utils/clubLogos'
import { clubAbbrev as aflClubAbbrev } from '@/features/afl/utils/clubAbbrev'
import { clubLogoUrl as fflClubLogoUrl } from '../utils/clubLogos'

const props = defineProps<{ aflPlayerSeasonId: string }>()

const { isDark } = useTheme()

const { result: aflResult, loading, error } = useQuery(
  GET_AFL_PLAYER_SEASON_STATS,
  () => ({ id: props.aflPlayerSeasonId }),
)

const { result: fflResult } = useQuery(
  GET_FFL_PLAYER_STINTS,
  () => ({ aflPlayerSeasonId: props.aflPlayerSeasonId }),
)

interface PlayerSeasonLink {
  id: string
  clubSeason: { id: string; club: { id: string; name: string }; season: { id: string; name: string } }
}

const playerSeason = computed(() => aflResult.value?.aflPlayerSeason as {
  id: string
  player: { id: string; name: string; playerSeasons: PlayerSeasonLink[] }
  clubSeason: { id: string; club: { id: string; name: string }; season: { id: string; name: string } }
  matches: AFLPlayerMatch[]
  statsAll: AFLStatSummary | null
  statsLastN: AFLStatSummary | null
  statsMedian: AFLStatSummary | null
} | null)
const notFound = useNotFound(playerSeason, loading, error)

// The player's other AFL seasons — the one being viewed is dropped, since it's
// already the page you're on. Server orders these most recent first.
const otherSeasons = computed(() => {
  const all = playerSeason.value?.player.playerSeasons ?? []
  return all
    .filter(ps => ps.id !== playerSeason.value?.id)
    .map(ps => ({
      id: ps.id,
      clubName: ps.clubSeason.club.name,
      seasonName: ps.clubSeason.season.name,
    }))
})
const stints = computed(() => fflResult.value?.fflPlayerSeasonsByAflPlayerSeason ?? [])

const breadcrumbs = computed(() => {
  if (!playerSeason.value) return []
  return [
    { label: 'FFL', to: { name: 'home' } },
    { label: playerSeason.value.clubSeason.season.name, to: { name: 'afl-home' } },
    { label: playerSeason.value.player.name },
  ]
})

// ── Types ────────────────────────────────────────────────────────────────────

interface AFLStatSummary {
  goals: number; kicks: number; handballs: number
  marks: number; tackles: number; hitouts: number
  games?: number
}

interface AFLPlayerMatch {
  id: string
  status: string
  kicks: number
  handballs: number
  marks: number
  tackles: number
  hitouts: number
  goals: number
  clubMatch: {
    club: { id: string; name: string }
    match: {
      id: string
      round: { id: string; name: string }
      homeClubMatch: { club: { id: string; name: string } }
      awayClubMatch: { club: { id: string; name: string } }
    }
  }
}

interface AFLMatchRef {
  round: { id: string; name: string }
  homeClubMatch: { club: { id: string; name: string } }
  awayClubMatch: { club: { id: string; name: string } }
}

interface FFLPlayerMatchData {
  id: string
  matchId: string | null
  position: string | null
  backupPositions: string | null
  interchangePosition: string | null
  status: string | null
  aflStatus: string | null
  score: number
  aflPlayerMatch: { id: string; clubMatch?: { match?: AFLMatchRef } } | null
}

interface FFLStint {
  id: string
  club: { id: string; name: string }
  clubSeasonId: string
  fromRoundId: string | null
  toRoundId: string | null
  playerMatches: FFLPlayerMatchData[]
}

interface MergedRow extends AFLPlayerMatch {
  starScore: number
  aflMatchId: string
  fflMatchId: string | null
  fflClubName: string | null
  fflPosition: string | null
  fflScore: number | null
  fflAflStatus: string | null
  fflBackupPositions: string | null
  fflInterchangePosition: string | null
}

// ── Data ─────────────────────────────────────────────────────────────────────

const sortedMatches = computed((): AFLPlayerMatch[] => {
  const ms = playerSeason.value?.matches ?? []
  return [...ms].sort((a, b) => parseInt(a.clubMatch.match.round.id) - parseInt(b.clubMatch.match.round.id))
})

const playedMatches = computed(() =>
  sortedMatches.value.filter(m => m.status === 'played' || m.status === 'playing'),
)

// Build lookup: AFL player match ID → FFL data + club name
const fflByAflMatchId = computed(() => {
  const map = new Map<string, {
    matchId: string | null
    clubName: string
    position: string | null
    score: number
    aflStatus: string | null
    backupPositions: string | null
    interchangePosition: string | null
  }>()
  for (const stint of stints.value as FFLStint[]) {
    for (const pm of stint.playerMatches) {
      if (pm.aflPlayerMatch) {
        map.set(pm.aflPlayerMatch.id, {
          matchId: pm.matchId,
          clubName: stint.club.name,
          position: pm.position,
          score: pm.score,
          aflStatus: pm.aflStatus,
          backupPositions: pm.backupPositions,
          interchangePosition: pm.interchangePosition,
        })
      }
    }
  }
  return map
})

const mergedRows = computed((): MergedRow[] =>
  sortedMatches.value.map(m => {
    const ffl = fflByAflMatchId.value.get(m.id) ?? null
    return {
      ...m,
      starScore: starScore(m),
      aflMatchId: m.clubMatch.match.id,
      fflMatchId: ffl?.matchId ?? null,
      fflClubName: ffl?.clubName ?? null,
      fflPosition: ffl?.position ?? null,
      fflScore: ffl !== null ? ffl.score : null,
      fflAflStatus: ffl?.aflStatus ?? null,
      fflBackupPositions: ffl?.backupPositions ?? null,
      fflInterchangePosition: ffl?.interchangePosition ?? null,
    }
  }),
)

// The match-by-match table shows most recent round first; averages/median above
// stay off the chronological mergedRows.
const displayRows = computed((): MergedRow[] => [...mergedRows.value].reverse())

// ── Analysis ─────────────────────────────────────────────────────────────────

function starFromSummary(s: AFLStatSummary | null): string {
  if (!s) return '—'
  return starScore(s).toFixed(1)
}

function statAvg(key: StatKey): string {
  return fmtStat(playerSeason.value?.statsAll?.[key])
}

function statLastNAvg(key: StatKey): string {
  return fmtStat(playerSeason.value?.statsLastN?.[key])
}

// Same trend semantics as the team builder: ↑/↓ per trendDir thresholds.
function statLastNTrend(key: StatKey): 'up' | 'down' | null {
  const all = playerSeason.value?.statsAll?.[key]
  const lastN = playerSeason.value?.statsLastN?.[key]
  if (all == null || lastN == null) return null
  return trendDir(lastN, all)
}

const starSeasonAvg = computed(() => starFromSummary(playerSeason.value?.statsAll ?? null))
const starLastNAvg  = computed(() => starFromSummary(playerSeason.value?.statsLastN ?? null))

const starLastNTrend = computed((): 'up' | 'down' | null => {
  const all = playerSeason.value?.statsAll
  const lastN = playerSeason.value?.statsLastN
  if (!all || !lastN) return null
  return trendDir(starScore(lastN), starScore(all))
})

const starMedian = computed(() => starFromSummary(playerSeason.value?.statsMedian ?? null))

function statMedian(key: StatKey): string {
  return fmtStat(playerSeason.value?.statsMedian?.[key])
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function opponentName(m: AFLPlayerMatch): string {
  const myId = playerSeason.value?.clubSeason.club.id
  const home = m.clubMatch.match.homeClubMatch.club
  const away = m.clubMatch.match.awayClubMatch.club
  return home.id === myId ? away.name : home.name
}

function shortRound(name: string): string {
  return name.replace(/^Round\s+/i, 'R')
}

function statColor(statKey: string, position: string | null): string {
  if (!position || position === 'star') return ''
  return statKey === position ? (POSITION_COLORS[position] ?? '') : ''
}

type HeatKey = typeof statCols[number]['key'] | 'star'

const roundHeatRanges = computed(() => {
  const played = mergedRows.value.filter(r => r.status !== 'dnp')
  const range = {} as Record<HeatKey, { min: number; max: number }>
  for (const col of statCols) {
    const vals = played.map(r => r[col.key]).filter((v): v is number => v != null)
    if (vals.length) range[col.key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  const starVals = played.map(r => r.starScore)
  if (starVals.length) range['star'] = { min: Math.min(...starVals), max: Math.max(...starVals) }
  return range
})

function roundHeat(value: number | null | undefined, key: HeatKey): Record<string, string> {
  if (value == null) return {}
  const r = roundHeatRanges.value[key]
  if (!r || r.min === r.max) return {}
  return heatStyle(value, r.min, r.max, isDark.value)
}


function posLetters(pos: string | null): string {
  if (!pos) return ''
  return pos.split(',').map(p => POSITION_LETTERS[p.trim()] ?? p.trim()).join('')
}

const stintEvents = computed(() =>
  (stints.value as FFLStint[]).map(stint => {
    const sorted = [...stint.playerMatches]
      .filter(pm => pm.aflPlayerMatch?.clubMatch?.match?.round)
      .sort((a, b) => parseInt(a.aflPlayerMatch!.clubMatch!.match!.round!.id) - parseInt(b.aflPlayerMatch!.clubMatch!.match!.round!.id))
    const first = sorted[0]?.aflPlayerMatch?.clubMatch?.match?.round
    const last = sorted[sorted.length - 1]?.aflPlayerMatch?.clubMatch?.match?.round
    return {
      clubName: stint.club.name,
      clubSeasonId: stint.clubSeasonId,
      from: first ? shortRound(first.name) : null,
      to: stint.toRoundId === null ? 'present' : (last ? shortRound(last.name) : null),
    }
  })
)
</script>
