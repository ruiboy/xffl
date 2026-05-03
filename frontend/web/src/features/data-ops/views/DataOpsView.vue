<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">Data Ops</h1>
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
    <!-- Tab: AFL Stats Import                       -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'afl-stats'">
      <div v-if="loadingLiveRound" class="text-text-faint">Loading…</div>
      <div v-else-if="liveRoundError" class="text-red-400">{{ liveRoundError.message }}</div>
      <template v-else>
        <!-- Round selector -->
        <div class="mb-5 max-w-xs">
          <label class="block text-xs font-medium text-text-muted mb-1">Round</label>
          <select
            v-model="selectedAflRoundId"
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
          >
            <option v-for="r in aflRounds" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
        </div>

        <!-- Match list -->
        <div v-if="loadingRoundStats" class="text-text-faint text-sm">Loading matches…</div>
        <div v-else-if="!aflMatches.length" class="text-text-faint text-sm">No matches in this round.</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-border">
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Match</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Status</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Score</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Players</th>
                <th class="pb-2 text-right text-xs font-medium text-text-faint"></th>
              </tr>
            </thead>
            <tbody>
              <template v-for="match in aflMatches" :key="match.id">
                <tr class="border-b border-border" :class="{ 'border-b-0': scrapeResult[match.id] || scrapeError[match.id] }">
                  <td class="py-3 pr-4 text-sm font-semibold whitespace-nowrap">
                    <router-link
                      v-if="aflSeasonId"
                      :to="{ name: 'afl-match', params: { seasonId: aflSeasonId, matchId: match.id } }"
                      class="text-text hover:text-active transition-colors"
                    >
                      {{ abbrevClub(match.homeClubMatch?.club.name) }}
                      <span class="font-normal text-text-faint text-xs mx-1.5">vs</span>
                      {{ abbrevClub(match.awayClubMatch?.club.name) }}
                    </router-link>
                    <template v-else>
                      {{ abbrevClub(match.homeClubMatch?.club.name) }}
                      <span class="font-normal text-text-faint text-xs mx-1.5">vs</span>
                      {{ abbrevClub(match.awayClubMatch?.club.name) }}
                    </template>
                  </td>
                  <td class="py-3 pr-4 whitespace-nowrap">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="statusBadge(match.dataStatus)"
                    >{{ statusLabel(match.dataStatus) }}</span>
                  </td>
                  <td class="py-3 pr-4 text-xs text-text-faint whitespace-nowrap tabular-nums">
                    <template v-if="match.homeClubMatch?.score != null && match.awayClubMatch?.score != null && match.dataStatus !== 'no_data'">
                      {{ match.homeClubMatch.score }}
                      <span class="text-text-faint mx-1">vs</span>
                      {{ match.awayClubMatch.score }}
                    </template>
                    <template v-else>—</template>
                  </td>
                  <td class="py-3 pr-4 text-xs text-text-faint whitespace-nowrap">
                    <template v-if="match.dataStatus === 'partial' || match.dataStatus === 'final'">
                      {{ match.homeClubMatch?.playerMatches?.length ?? 0 }} - {{ match.awayClubMatch?.playerMatches?.length ?? 0 }}
                    </template>
                    <template v-else>—</template>
                  </td>
                  <td class="py-3 text-right whitespace-nowrap">
                    <div class="flex items-center justify-end gap-2">
                      <button
                        v-if="match.dataStatus !== 'no_data'"
                        @click="toggleFinal(match)"
                        :disabled="togglingFinal[match.id]"
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ match.dataStatus === 'final' ? 'Mark Partial' : 'Mark Final' }}</button>
                      <button
                        @click="scrape(match)"
                        :disabled="scraping[match.id]"
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ scraping[match.id] ? 'Getting Stats…' : 'Get Stats' }}</button>
                    </div>
                  </td>
                </tr>
                <!-- Feedback / unmatched review sub-row -->
                <tr v-if="scrapeResult[match.id] || scrapeError[match.id]" class="border-b border-border">
                  <td colspan="5" class="pb-3 pt-0">
                    <span v-if="scrapeError[match.id]" class="text-xs text-red-400">{{ scrapeError[match.id] }}</span>
                    <template v-else-if="scrapeResult[match.id]">
                      <div class="text-xs mb-1">
                        <span class="text-green-500 font-medium">Imported</span>
                        <span class="text-text-faint ml-1">
                          · {{ scrapeResult[match.id].homePlayerCount + scrapeResult[match.id].awayPlayerCount }} players
                        </span>
                        <span v-if="scrapeResult[match.id].unmatchedPlayers.length > 0" class="text-yellow-500 ml-1">
                          · {{ scrapeResult[match.id].unmatchedPlayers.length }} unmatched — review below
                        </span>
                      </div>
                      <!-- Unmatched player review table -->
                      <div v-if="scrapeResult[match.id].unmatchedPlayers.length > 0" class="mt-2 border border-border rounded-lg overflow-hidden">
                        <table class="w-full">
                          <thead>
                            <tr class="border-b border-border bg-surface-raised">
                              <th class="px-3 py-2 text-left text-xs font-medium text-text-faint">Parsed name</th>
                              <th class="px-3 py-2 text-left text-xs font-medium text-text-faint">Stats</th>
                              <th class="px-3 py-2 text-left text-xs font-medium text-text-faint">Best match</th>
                              <th class="px-3 py-2 text-right text-xs font-medium text-text-faint"></th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr
                              v-for="(up, ui) in scrapeResult[match.id].unmatchedPlayers"
                              :key="ui"
                              class="border-b border-border last:border-0"
                              :class="unmatchedConfirmed[match.id]?.[ui] ? 'bg-green-500/5' : ''"
                            >
                              <td class="px-3 py-2 text-sm font-medium whitespace-nowrap">{{ up.parsedName }}</td>
                              <td class="px-3 py-2 text-xs text-text-faint whitespace-nowrap tabular-nums">
                                {{ up.goals }}g {{ up.kicks }}k {{ up.handballs }}hb {{ up.marks }}m {{ up.tackles }}t {{ up.hitouts }}ho
                              </td>
                              <td class="px-3 py-2">
                                <template v-if="unmatchedConfirmed[match.id]?.[ui]">
                                  <span class="text-green-500 text-xs font-medium">✓ Confirmed</span>
                                </template>
                                <template v-else-if="up.candidates.length === 0">
                                  <span class="text-text-faint text-xs">No candidates</span>
                                </template>
                                <template v-else>
                                  <select
                                    v-model="unmatchedSelections[match.id + ':' + ui]"
                                    class="rounded border border-border bg-surface px-2 py-1 text-xs text-text focus:outline-none focus:ring-1 focus:ring-active"
                                  >
                                    <option value="">Select player…</option>
                                    <option
                                      v-for="c in up.candidates"
                                      :key="c.playerSeasonId"
                                      :value="c.playerSeasonId"
                                    >{{ c.name }} ({{ (c.confidence * 100).toFixed(0) }}%)</option>
                                  </select>
                                </template>
                              </td>
                              <td class="px-3 py-2 text-right">
                                <button
                                  v-if="!unmatchedConfirmed[match.id]?.[ui] && up.candidates.length > 0"
                                  @click="confirmUnmatched(match.id, ui, up)"
                                  :disabled="!unmatchedSelections[match.id + ':' + ui] || confirmingUnmatched[match.id + ':' + ui]"
                                  class="rounded border border-border px-2 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                                >{{ confirmingUnmatched[match.id + ':' + ui] ? '…' : 'Confirm' }}</button>
                                <button
                                  v-else-if="!unmatchedConfirmed[match.id]?.[ui]"
                                  class="rounded border border-border px-2 py-1 text-xs text-text-faint"
                                  disabled
                                >Skip</button>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </template>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: FFL Teams                              -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'team-submission'">
      <div v-if="loadingSeasonData" class="text-text-faint">Loading...</div>
      <div v-else-if="seasonError" class="text-red-400">{{ seasonError.message }}</div>
      <template v-else-if="season">

        <!-- Round selector -->
        <div class="mb-5 max-w-xs">
          <label class="block text-xs font-medium text-text-muted mb-1">Round</label>
          <select
            v-model="selectedRoundId"
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
          >
            <option v-for="r in season.rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
        </div>

        <!-- Club list for selected round -->
        <div v-if="!selectedRound" class="text-text-faint text-sm">Select a round.</div>
        <div v-else-if="!fflClubRows.length" class="text-text-faint text-sm">No matches in this round.</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-border">
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Club</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Status</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Score</th>
                <th class="pb-2 text-right text-xs font-medium text-text-faint"></th>
              </tr>
            </thead>
            <tbody>
              <template v-for="row in fflClubRows" :key="row.clubMatchId">
                <!-- Club row -->
                <tr
                  class="border-b border-border"
                  :class="{ 'border-b-0': activeImportClubMatchId === row.clubMatchId }"
                >
                  <td class="py-3 pr-4 text-sm font-semibold">{{ row.clubName }}</td>
                  <td class="py-3 pr-4 whitespace-nowrap">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="fflStatusBadge(row.dataStatus)"
                    >{{ fflStatusLabel(row.dataStatus) }}</span>
                  </td>
                  <td class="py-3 pr-4 text-xs text-text-faint tabular-nums">
                    {{ row.dataStatus !== 'no_data' ? row.score : '—' }}
                  </td>
                  <td class="py-3 text-right">
                    <button
                      @click="toggleImportPanel(row.clubMatchId, row.clubSeasonId, row.clubName)"
                      class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors"
                    >{{ activeImportClubMatchId === row.clubMatchId ? 'Cancel' : 'Import Team' }}</button>
                  </td>
                </tr>

                <!-- Inline import panel -->
                <tr v-if="activeImportClubMatchId === row.clubMatchId" class="border-b border-border">
                  <td colspan="4" class="pb-4 pt-1">
                    <div v-if="importPhase === 'input'" class="space-y-4 max-w-xl pl-1">
                      <!-- Team format -->
                      <div>
                        <label class="block text-xs font-medium text-text-muted mb-1">Team format</label>
                        <select
                          v-model="teamName"
                          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
                        >
                          <option value="">Auto-detect</option>
                          <option value="Ruiboys">Ruiboys</option>
                          <option value="Slashers">Slashers</option>
                          <option value="Cheetahs">Cheetahs</option>
                          <option value="THC">THC</option>
                        </select>
                      </div>
                      <!-- Paste area -->
                      <div>
                        <label class="block text-xs font-medium text-text-muted mb-1">Team (e.g. forum post)</label>
                        <textarea
                          v-model="post"
                          rows="14"
                          placeholder="Paste team here…"
                          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-active resize-y"
                        />
                      </div>
                      <p v-if="parseError" class="text-sm text-red-400">{{ parseError }}</p>
                      <button
                        @click="onParse"
                        :disabled="!canParse || parsing"
                        class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                      >{{ parsing ? 'Parsing…' : 'Parse' }}</button>
                    </div>

                    <!-- Review table -->
                    <div v-else-if="importPhase === 'review'" class="pl-1">
                      <div class="mb-3 flex items-center gap-4">
                        <button
                          @click="importPhase = 'input'"
                          class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
                        >← Back</button>
                        <span class="text-sm text-text-muted">
                          {{ resolvedPlayers.length }} players ·
                          <span :class="needsReview.length > 0 ? 'text-yellow-500' : 'text-green-500'">
                            {{ needsReview.length }} need review
                          </span>
                        </span>
                        <button
                          @click="onConfirm"
                          :disabled="confirming || unresolvedCount > 0 || imported"
                          class="ml-auto rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                        >{{ confirming ? 'Saving…' : 'Confirm & Import' }}</button>
                      </div>
                      <p v-if="unresolvedCount > 0" class="mb-2 text-sm text-red-400">
                        {{ unresolvedCount }} player(s) unresolved — confirm disabled.
                      </p>
                      <p v-if="compositionWarnings.length > 0" class="mb-2 text-sm text-yellow-500">
                        Invalid team: {{ compositionWarnings.join(' · ') }}
                      </p>
                      <p v-if="confirmError" class="mb-2 text-sm text-red-400">{{ confirmError }}</p>
                      <p v-if="confirmSuccess" class="mb-2 text-sm text-green-500">{{ confirmSuccess }}</p>

                      <table class="w-full">
                        <thead>
                          <tr class="border-b border-border">
                            <th colspan="2" class="px-4 pb-2 text-left text-xs font-medium text-text-faint">Posted</th>
                            <th class="pb-2"></th>
                            <th colspan="2" class="px-4 pb-2 text-left text-xs font-medium text-text-faint">Resolved</th>
                            <th class="px-4 pb-2 text-left text-xs font-medium text-text-faint">Position</th>
                            <th class="px-4 pb-2 text-right text-xs font-medium text-text-faint">Score</th>
                            <th class="px-4 pb-2 text-right text-xs font-medium text-text-faint">Confidence</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr
                            v-for="(rp, i) in resolvedPlayers"
                            :key="i"
                            class="border-b border-border last:border-0"
                            :class="rowClass(i)"
                          >
                            <td class="pl-4 pr-2 py-3 font-mono text-sm font-medium text-text whitespace-nowrap">{{ rp.parsedName }}</td>
                            <td class="pr-3 py-3 text-xs text-text-faint whitespace-nowrap">{{ rp.clubHint }}</td>
                            <td class="py-3 text-text-faint text-sm select-none px-1">→</td>
                            <td class="pl-4 pr-2 py-3 text-sm font-semibold text-text whitespace-nowrap">
                              <span v-if="rp.resolvedName">{{ rp.resolvedName }}</span>
                              <span v-else class="text-red-400 font-normal">Unresolved</span>
                            </td>
                            <td class="pr-4 py-3 text-xs text-text-muted whitespace-nowrap">{{ rp.resolvedClub ?? '—' }}</td>
                            <td class="px-4 py-3 text-sm whitespace-nowrap">
                              <span :class="POSITION_COLORS[rp.position] ?? 'text-text-muted'">{{ displayPosition(rp) }}</span>
                            </td>
                            <td class="px-4 py-3 text-right tabular-nums text-sm text-text">
                              {{ rp.score ?? '—' }}
                            </td>
                            <td class="pl-4 pr-4 py-3 text-right">
                              <span
                                class="inline-block rounded-full px-2.5 py-0.5 text-xs font-medium"
                                :class="confidenceBadge(rp.confidence)"
                              >{{ (rp.confidence * 100).toFixed(0) }}%</span>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_DATA_OPS, GET_AFL_ROUND_STATS } from '../api/queries'
import { PARSE_TEAM_SUBMISSION, CONFIRM_TEAM_SUBMISSION, IMPORT_AFL_MATCH_STATS, MARK_AFL_MATCH_STATS_COMPLETE, CONFIRM_AFL_PLAYER_MATCH } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { GET_AFL_LIVE_ROUND } from '@/features/afl/api/queries'
import { POSITION_COLORS, POSITION_LABEL, POSITION_SLOTS } from '@/features/ffl/utils/position'

const { liveSeasonId, liveRoundId } = useFflState()

// ---- Tabs ----
const tabs = [
  { id: 'team-submission', label: 'FFL Teams' },
  { id: 'afl-stats', label: 'AFL Stats' },
]
const activeTab = ref('afl-stats')

// ════════════════════════════════════════════
// AFL Stats Import
// ════════════════════════════════════════════

const { result: liveRoundResult, loading: loadingLiveRound, error: liveRoundError } = useQuery(GET_AFL_LIVE_ROUND)

const aflRounds = computed(() => liveRoundResult.value?.aflLiveRound?.round?.season?.rounds ?? [])
const selectedAflRoundId = ref('')

watch(liveRoundResult, (val) => {
  if (val?.aflLiveRound?.round?.id && !selectedAflRoundId.value) {
    selectedAflRoundId.value = val.aflLiveRound.round.id
  }
}, { immediate: true })

const { result: roundStatsResult, loading: loadingRoundStats, refetch: refetchRoundStats } = useQuery(
  GET_AFL_ROUND_STATS,
  () => ({ roundId: selectedAflRoundId.value }),
  () => ({ enabled: !!selectedAflRoundId.value }),
)

const aflMatches = computed(() => roundStatsResult.value?.aflRound?.matches ?? [])
const aflSeasonId = computed(() => roundStatsResult.value?.aflRound?.season?.id ?? '')

type UnmatchedAFLPlayer = {
  parsedName: string
  clubMatchId: string
  kicks: number; handballs: number; marks: number; hitouts: number; tackles: number; goals: number; behinds: number
  candidates: { playerSeasonId: string; name: string; confidence: number }[]
}

type ScrapeResult = {
  homeClubName: string; awayClubName: string
  homePlayerCount: number; awayPlayerCount: number
  unmatchedPlayers: UnmatchedAFLPlayer[]
}

const scraping = ref<Record<string, boolean>>({})
const scrapeResult = ref<Record<string, ScrapeResult>>({})
const scrapeError = ref<Record<string, string>>({})
const togglingFinal = ref<Record<string, boolean>>({})

// unmatched review: keyed by "matchId:rowIndex"
const unmatchedSelections = ref<Record<string, string>>({})
const confirmingUnmatched = ref<Record<string, boolean>>({})
// keyed by matchId → rowIndex → true when confirmed
const unmatchedConfirmed = ref<Record<string, Record<number, boolean>>>({})

const { mutate: importStatsMutation } = useMutation(IMPORT_AFL_MATCH_STATS)
const { mutate: markCompleteMutation } = useMutation(MARK_AFL_MATCH_STATS_COMPLETE)
const { mutate: confirmPlayerMatchMutation } = useMutation(CONFIRM_AFL_PLAYER_MATCH)

async function scrape(match: any) {
  scraping.value[match.id] = true
  scrapeError.value[match.id] = ''
  scrapeResult.value[match.id] = undefined as any
  try {
    const res = await importStatsMutation({ matchId: match.id })
    const data = res?.data?.importAFLMatchStats
    if (data) scrapeResult.value[match.id] = data
    await refetchRoundStats()
  } catch (e: any) {
    scrapeError.value[match.id] = e.message ?? 'Scrape failed'
  } finally {
    scraping.value[match.id] = false
  }
}

async function toggleFinal(match: any) {
  togglingFinal.value[match.id] = true
  try {
    await markCompleteMutation({
      matchId: match.id,
      complete: match.dataStatus !== 'final',
    })
    await refetchRoundStats()
  } catch (e: any) {
    scrapeError.value[match.id] = e.message ?? 'Update failed'
  } finally {
    togglingFinal.value[match.id] = false
  }
}

async function confirmUnmatched(matchId: string, rowIndex: number, up: UnmatchedAFLPlayer) {
  const key = `${matchId}:${rowIndex}`
  const playerSeasonId = unmatchedSelections.value[key]
  if (!playerSeasonId) return
  confirmingUnmatched.value[key] = true
  try {
    await confirmPlayerMatchMutation({
      input: {
        playerSeasonId,
        clubMatchId: up.clubMatchId,
        status: 'played',
        kicks: up.kicks,
        handballs: up.handballs,
        marks: up.marks,
        hitouts: up.hitouts,
        tackles: up.tackles,
        goals: up.goals,
        behinds: up.behinds,
      },
    })
    if (!unmatchedConfirmed.value[matchId]) unmatchedConfirmed.value[matchId] = {}
    unmatchedConfirmed.value[matchId][rowIndex] = true
    await refetchRoundStats()
  } catch (e: any) {
    scrapeError.value[matchId] = e.message ?? 'Confirm failed'
  } finally {
    confirmingUnmatched.value[key] = false
  }
}

const CLUB_ABBREV: Record<string, string> = {
  'Adelaide Crows': 'Adelaide',
  'Brisbane Lions': 'Brisbane',
  'Gold Coast Suns': 'Gold Coast',
  'Greater Western Sydney Giants': 'GWS',
  'Greater Western Sydney': 'GWS',
  'North Melbourne': 'North Melb.',
  'Port Adelaide Power': 'Port Adelaide',
  'Sydney Swans': 'Sydney',
  'West Coast Eagles': 'West Coast',
  'Western Bulldogs': 'W. Bulldogs',
}

function abbrevClub(name: string | undefined): string {
  if (!name) return '—'
  return CLUB_ABBREV[name] ?? name
}

function statusLabel(status: string): string {
  if (status === 'final') return 'Final'
  if (status === 'partial') return 'Partial'
  return 'No data'
}

function statusBadge(status: string): string {
  if (status === 'final') return 'bg-green-500/15 text-green-500'
  if (status === 'partial') return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-surface-raised text-text-faint'
}

function fflStatusLabel(status: string): string {
  if (status === 'final') return 'Final'
  if (status === 'submitted') return 'Submitted'
  return 'Not submitted'
}

function fflStatusBadge(status: string): string {
  if (status === 'final') return 'bg-green-500/15 text-green-500'
  if (status === 'submitted') return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-surface-raised text-text-faint'
}

// ════════════════════════════════════════════
// FFL Team Submission
// ════════════════════════════════════════════

const { result: seasonResult, loading: loadingSeasonData, error: seasonError, refetch: refetchSeasonData } = useQuery(
  GET_FFL_DATA_OPS,
  () => ({ seasonId: liveSeasonId.value }),
  () => ({ enabled: !!liveSeasonId.value }),
)

const season = computed(() => seasonResult.value?.fflSeason ?? null)

const selectedRoundId = ref(liveRoundId.value)

watch(liveRoundId, (val) => {
  if (val && !selectedRoundId.value) selectedRoundId.value = val
}, { immediate: true })

const selectedRound = computed(() =>
  season.value?.rounds.find((r: any) => r.id === selectedRoundId.value) ?? null,
)

type FflClubRow = {
  clubMatchId: string
  clubSeasonId: string
  clubName: string
  dataStatus: string
  score: number
}

const fflClubRows = computed<FflClubRow[]>(() => {
  if (!selectedRound.value) return []
  const rows: FflClubRow[] = []
  for (const match of selectedRound.value.matches) {
    if (match.homeClubMatch) {
      rows.push({
        clubMatchId: match.homeClubMatch.id,
        clubSeasonId: match.homeClubMatch.club.id,
        clubName: match.homeClubMatch.club.name,
        dataStatus: match.homeClubMatch.dataStatus ?? 'no_data',
        score: match.homeClubMatch.score ?? 0,
      })
    }
    if (match.awayClubMatch) {
      rows.push({
        clubMatchId: match.awayClubMatch.id,
        clubSeasonId: match.awayClubMatch.club.id,
        clubName: match.awayClubMatch.club.name,
        dataStatus: match.awayClubMatch.dataStatus ?? 'no_data',
        score: match.awayClubMatch.score ?? 0,
      })
    }
  }
  return rows
})

// ---- Import panel (inline, per-club-match) ----

const activeImportClubMatchId = ref('')
const activeImportClubSeasonId = ref('')

function toggleImportPanel(clubMatchId: string, clubSeasonId: string, clubName: string) {
  if (activeImportClubMatchId.value === clubMatchId) {
    activeImportClubMatchId.value = ''
    activeImportClubSeasonId.value = ''
    resetImportState()
  } else {
    activeImportClubMatchId.value = clubMatchId
    activeImportClubSeasonId.value = clubSeasonId
    resetImportState()
    // Pre-select team format matching club name
    teamName.value = ['Ruiboys', 'Slashers', 'Cheetahs', 'THC'].find(n =>
      clubName.toLowerCase().includes(n.toLowerCase())
    ) ?? ''
  }
}

function resetImportState() {
  importPhase.value = 'input'
  post.value = ''
  parseError.value = ''
  confirmError.value = ''
  confirmSuccess.value = ''
  imported.value = false
  resolvedPlayers.value = []
  needsReview.value = []
}

const importPhase = ref<'input' | 'review'>('input')
const teamName = ref('')
const post = ref('')
const parseError = ref('')
const parsing = ref(false)

const canParse = computed(() =>
  !!activeImportClubMatchId.value && post.value.trim().length > 0,
)

type ResolvedPlayer = {
  parsedName: string; clubHint: string; resolvedName: string | null; resolvedClub: string | null
  position: string; backupPositions: string; interchangePosition: string
  score: number | null; notes: string; playerSeasonId: string | null; confidence: number
}

const resolvedPlayers = ref<ResolvedPlayer[]>([])
const needsReview = ref<number[]>([])

const { mutate: parseMutation } = useMutation(PARSE_TEAM_SUBMISSION)

async function onParse() {
  parseError.value = ''
  confirmError.value = ''
  confirmSuccess.value = ''
  imported.value = false
  parsing.value = true
  try {
    const res = await parseMutation({
      input: {
        clubSeasonId: activeImportClubSeasonId.value,
        clubMatchId: activeImportClubMatchId.value,
        teamName: teamName.value,
        post: post.value,
      },
    })
    const data = res?.data?.parseFFLTeamSubmission
    if (!data) throw new Error('No result returned')
    resolvedPlayers.value = data.resolvedPlayers
    needsReview.value = data.needsReview
    importPhase.value = 'review'
  } catch (e: any) {
    parseError.value = e.message ?? 'Parse failed'
  } finally {
    parsing.value = false
  }
}

const confirming = ref(false)
const confirmError = ref('')
const confirmSuccess = ref('')
const imported = ref(false)

const { mutate: confirmMutation } = useMutation(CONFIRM_TEAM_SUBMISSION)

const unresolvedCount = computed(() => resolvedPlayers.value.filter(rp => !rp.playerSeasonId).length)

const compositionWarnings = computed(() => {
  const counts: Record<string, number> = {}
  for (const rp of resolvedPlayers.value) {
    if (!rp.playerSeasonId) continue
    if (rp.backupPositions || rp.interchangePosition) continue
    counts[rp.position] = (counts[rp.position] ?? 0) + 1
  }
  return Object.entries(POSITION_SLOTS)
    .filter(([pos, expected]) => (counts[pos] ?? 0) !== expected)
    .map(([pos, expected]) => `${POSITION_LABEL[pos] ?? pos} ${counts[pos] ?? 0}/${expected}`)
})

async function onConfirm() {
  confirmError.value = ''
  confirmSuccess.value = ''
  confirming.value = true
  try {
    const players = resolvedPlayers.value
      .filter(rp => rp.playerSeasonId)
      .map(rp => ({
        playerSeasonId: rp.playerSeasonId!,
        position: rp.position,
        backupPositions: rp.backupPositions || null,
        interchangePosition: rp.interchangePosition || null,
        score: rp.score,
      }))
    const res = await confirmMutation({ input: { clubMatchId: activeImportClubMatchId.value, players } })
    const saved = res?.data?.confirmFFLTeamSubmission ?? []
    confirmSuccess.value = `Imported ${saved.length} player records.`
    imported.value = true
    await refetchSeasonData()
  } catch (e: any) {
    confirmError.value = e.message ?? 'Confirm failed'
  } finally {
    confirming.value = false
  }
}

function displayPosition(rp: ResolvedPlayer): string {
  if (rp.backupPositions) {
    const labels = rp.backupPositions.split(',').map(p => {
      const key = p.trim()
      const label = POSITION_LABEL[key] ?? key
      return key === rp.interchangePosition ? `${label} (Int)` : label
    }).join(', ')
    return `Bench - ${labels}`
  }
  return POSITION_LABEL[rp.position] ?? rp.position
}

function confidenceBadge(confidence: number): string {
  if (confidence >= 0.85) return 'bg-green-500/15 text-green-500'
  if (confidence >= 0.6) return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-red-500/15 text-red-400'
}

function rowClass(i: number): string {
  const rp = resolvedPlayers.value[i]
  if (!rp.playerSeasonId) return 'bg-red-500/5'
  if (needsReview.value.includes(i)) return 'bg-yellow-500/5'
  return ''
}
</script>
