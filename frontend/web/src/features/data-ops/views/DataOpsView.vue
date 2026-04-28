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
        <div v-else class="space-y-3">
          <div
            v-for="match in aflMatches"
            :key="match.id"
            class="rounded-lg border border-border bg-surface p-4"
          >
            <!-- Match header -->
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-3">
                <span class="text-sm font-medium text-text">
                  {{ match.homeClubMatch?.club.name ?? '—' }}
                  <span class="text-text-faint mx-2">vs</span>
                  {{ match.awayClubMatch?.club.name ?? '—' }}
                </span>
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="statusBadge(match.statsImportStatus)"
                >
                  {{ statusLabel(match.statsImportStatus) }}
                </span>
                <span v-if="match.statsImportedAt" class="text-xs text-text-faint">
                  last import {{ formatDate(match.statsImportedAt) }}
                </span>
              </div>

              <!-- Actions -->
              <div class="flex items-center gap-2">
                <button
                  v-if="match.statsImportStatus !== 'no_data'"
                  @click="toggleComplete(match)"
                  :disabled="togglingComplete[match.id]"
                  class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                >
                  {{ match.statsImportStatus === 'complete' ? 'Mark Partial' : 'Mark Complete' }}
                </button>
                <button
                  @click="scrape(match)"
                  :disabled="scraping[match.id]"
                  class="rounded border border-active bg-active px-3 py-1 text-xs font-medium text-active-text transition-colors disabled:opacity-40"
                >
                  {{ scraping[match.id] ? 'Getting Stats…' : 'Get Stats' }}
                </button>
              </div>
            </div>

            <!-- Scrape result -->
            <div v-if="scrapeResult[match.id]" class="mt-2 text-xs text-text-muted border-t border-border pt-2 space-y-1">
              <div>
                <span class="text-green-500 font-medium">Imported</span>
                · {{ scrapeResult[match.id].homeClubName }} {{ scrapeResult[match.id].homePlayerCount }} players
                · {{ scrapeResult[match.id].awayClubName }} {{ scrapeResult[match.id].awayPlayerCount }} players
              </div>
              <div v-if="scrapeResult[match.id].unmatchedPlayers.length > 0" class="text-yellow-500">
                Unmatched ({{ scrapeResult[match.id].unmatchedPlayers.length }}):
                {{ scrapeResult[match.id].unmatchedPlayers.join(', ') }}
              </div>
            </div>
            <div v-if="scrapeError[match.id]" class="mt-2 text-xs text-red-400 border-t border-border pt-2">
              {{ scrapeError[match.id] }}
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Team Submission                        -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'team-submission'">
      <div v-if="loadingSeasonData" class="text-text-faint">Loading...</div>
      <div v-else-if="seasonError" class="text-red-400">{{ seasonError.message }}</div>
      <template v-else-if="season">

        <!-- Step 1: Input -->
        <template v-if="phase === 'input'">
          <div class="space-y-5 max-w-xl">

            <!-- Club + Round selectors -->
            <div class="flex gap-4">
              <div class="flex-1">
                <label class="block text-xs font-medium text-text-muted mb-1">Club</label>
                <select
                  v-model="selectedClubSeasonId"
                  class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
                >
                  <option value="">Select club…</option>
                  <option v-for="cs in season.ladder" :key="cs.id" :value="cs.id">
                    {{ cs.club.name }}
                  </option>
                </select>
              </div>
              <div class="flex-1">
                <label class="block text-xs font-medium text-text-muted mb-1">Round</label>
                <select
                  v-model="selectedRoundId"
                  class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
                >
                  <option value="">Select round…</option>
                  <option v-for="r in season.rounds" :key="r.id" :value="r.id">
                    {{ r.name }}
                  </option>
                </select>
              </div>
            </div>

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
              <label class="block text-xs font-medium text-text-muted mb-1">Forum post</label>
              <textarea
                v-model="post"
                rows="16"
                placeholder="Paste the forum post here…"
                class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-active resize-y"
              />
            </div>

            <!-- Error -->
            <p v-if="parseError" class="text-sm text-red-400">{{ parseError }}</p>

            <!-- Warning: no club_match found -->
            <p v-if="selectedClubSeasonId && selectedRoundId && !clubMatchId" class="text-sm text-yellow-500">
              No match found for this club in the selected round.
            </p>

            <button
              @click="onParse"
              :disabled="!canParse || parsing"
              class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ parsing ? 'Parsing…' : 'Parse' }}
            </button>
          </div>
        </template>

        <!-- Step 2: Review -->
        <template v-if="phase === 'review'">
          <div class="mb-4 flex items-center gap-4">
            <button
              @click="goBack"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >
              ← Back
            </button>
            <span class="text-sm text-text-muted">
              {{ resolvedPlayers.length }} players parsed ·
              <span :class="needsReview.length > 0 ? 'text-yellow-500' : 'text-green-500'">
                {{ needsReview.length }} need review
              </span>
            </span>
            <button
              @click="onConfirm"
              :disabled="confirming || unresolvedCount > 0 || imported"
              class="ml-auto rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ confirming ? 'Saving…' : 'Confirm & Import' }}
            </button>
          </div>

          <p v-if="unresolvedCount > 0" class="mb-3 text-sm text-red-400">
            {{ unresolvedCount }} player(s) could not be resolved and will be skipped.
            Confirm is disabled until all are resolved or remove them from the post.
          </p>
          <p v-if="compositionWarnings.length > 0" class="mb-3 text-sm text-yellow-500">
            Invalid team: {{ compositionWarnings.join(' · ') }}
          </p>
          <p v-if="confirmError" class="mb-3 text-sm text-red-400">{{ confirmError }}</p>
          <p v-if="confirmSuccess" class="mb-3 text-sm text-green-500">{{ confirmSuccess }}</p>

          <div class="overflow-x-auto rounded-lg border border-border">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border bg-surface-raised">
                  <th colspan="2" class="px-3 py-2 text-left font-medium text-text-muted">Posted</th>
                  <th class="py-2"></th>
                  <th colspan="2" class="px-3 py-2 text-left font-medium text-text-muted">Resolved</th>
                  <th class="px-3 py-2 text-left font-medium text-text-muted">Position</th>
                  <th class="px-3 py-2 text-right font-medium text-text-muted">Score</th>
                  <th class="px-3 py-2 text-center font-medium text-text-muted">Confidence</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(rp, i) in resolvedPlayers"
                  :key="i"
                  class="border-b border-border last:border-0"
                  :class="rowClass(i)"
                >
                  <td class="px-3 py-2 font-mono text-xs text-text">{{ rp.parsedName }}</td>
                  <td class="pr-3 py-2 text-xs text-text-muted">{{ rp.clubHint }}</td>
                  <td class="py-2 text-text-faint text-xs select-none">→</td>
                  <td class="px-3 py-2 text-text text-sm">
                    <span v-if="rp.resolvedName">{{ rp.resolvedName }}</span>
                    <span v-else class="text-red-400">Unresolved</span>
                  </td>
                  <td class="pr-3 py-2 text-xs text-text-muted">{{ rp.resolvedClub ?? '—' }}</td>
                  <td class="px-3 py-2 text-xs">
                    <span :class="POSITION_COLORS[rp.position] ?? 'text-text-muted'">{{ displayPosition(rp) }}</span>
                  </td>
                  <td class="px-3 py-2 text-right tabular-nums text-text-muted">
                    {{ rp.score ?? '—' }}
                  </td>
                  <td class="px-3 py-2 text-center">
                    <span
                      class="inline-block rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="confidenceBadge(rp.confidence)"
                    >
                      {{ (rp.confidence * 100).toFixed(0) }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_DATA_OPS, GET_AFL_ROUND_STATS } from '../api/queries'
import { PARSE_TEAM_SUBMISSION, CONFIRM_TEAM_SUBMISSION, IMPORT_AFL_MATCH_STATS, MARK_AFL_MATCH_STATS_COMPLETE } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { GET_AFL_LIVE_ROUND } from '@/features/afl/api/queries'
import { POSITION_COLORS, POSITION_LABEL, POSITION_SLOTS } from '@/features/ffl/utils/position'

const { liveSeasonId, liveRoundId } = useFflState()

// ---- Tabs ----
const tabs = [
  { id: 'afl-stats', label: 'AFL Stats' },
  { id: 'team-submission', label: 'Team Submission' },
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

// Track per-match scrape state
const scraping = ref<Record<string, boolean>>({})
const scrapeResult = ref<Record<string, { homeClubName: string; awayClubName: string; homePlayerCount: number; awayPlayerCount: number; unmatchedPlayers: string[] }>>({})
const scrapeError = ref<Record<string, string>>({})
const togglingComplete = ref<Record<string, boolean>>({})

const { mutate: importStatsMutation } = useMutation(IMPORT_AFL_MATCH_STATS)
const { mutate: markCompleteMutation } = useMutation(MARK_AFL_MATCH_STATS_COMPLETE)

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

async function toggleComplete(match: any) {
  togglingComplete.value[match.id] = true
  try {
    await markCompleteMutation({
      matchId: match.id,
      complete: match.statsImportStatus !== 'complete',
    })
    await refetchRoundStats()
  } catch (e: any) {
    scrapeError.value[match.id] = e.message ?? 'Update failed'
  } finally {
    togglingComplete.value[match.id] = false
  }
}

function statusLabel(status: string): string {
  if (status === 'complete') return 'Complete'
  if (status === 'partial') return 'Partial'
  return 'No data'
}

function statusBadge(status: string): string {
  if (status === 'complete') return 'bg-green-500/15 text-green-500'
  if (status === 'partial') return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-surface-raised text-text-faint'
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString('en-AU', { dateStyle: 'short', timeStyle: 'short' })
}

// ════════════════════════════════════════════
// Team Submission (unchanged)
// ════════════════════════════════════════════

const { result: seasonResult, loading: loadingSeasonData, error: seasonError } = useQuery(
  GET_FFL_DATA_OPS,
  () => ({ seasonId: liveSeasonId.value }),
  () => ({ enabled: !!liveSeasonId.value }),
)

const season = computed(() => seasonResult.value?.fflSeason ?? null)

const selectedClubSeasonId = ref('')
const selectedRoundId = ref(liveRoundId.value)
const teamName = ref('')
const post = ref('')

const selectedClubId = computed(() => {
  const cs = season.value?.ladder.find((l: any) => l.id === selectedClubSeasonId.value)
  return cs?.club.id ?? ''
})

const selectedRound = computed(() =>
  season.value?.rounds.find((r: any) => r.id === selectedRoundId.value) ?? null,
)

const clubMatchId = computed(() => {
  if (!selectedRound.value || !selectedClubId.value) return ''
  for (const match of selectedRound.value.matches) {
    if (match.homeClubMatch?.club.id === selectedClubId.value) return match.homeClubMatch.id
    if (match.awayClubMatch?.club.id === selectedClubId.value) return match.awayClubMatch.id
  }
  return ''
})

const canParse = computed(() =>
  !!selectedClubSeasonId.value && !!selectedRoundId.value && !!clubMatchId.value && post.value.trim().length > 0,
)

type ResolvedPlayer = {
  parsedName: string; clubHint: string; resolvedName: string | null; resolvedClub: string | null
  position: string; backupPositions: string; interchangePosition: string
  score: number | null; notes: string; playerSeasonId: string | null; confidence: number
}

const phase = ref<'input' | 'review'>('input')
const resolvedPlayers = ref<ResolvedPlayer[]>([])
const needsReview = ref<number[]>([])
const parseError = ref('')
const parsing = ref(false)

const { mutate: parseMutation } = useMutation(PARSE_TEAM_SUBMISSION)

async function onParse() {
  parseError.value = ''
  confirmError.value = ''
  confirmSuccess.value = ''
  imported.value = false
  parsing.value = true
  try {
    const res = await parseMutation({
      input: { clubSeasonId: selectedClubSeasonId.value, clubMatchId: clubMatchId.value, teamName: teamName.value, post: post.value },
    })
    const data = res?.data?.parseFFLTeamSubmission
    if (!data) throw new Error('No result returned')
    resolvedPlayers.value = data.resolvedPlayers
    needsReview.value = data.needsReview
    phase.value = 'review'
  } catch (e: any) {
    parseError.value = e.message ?? 'Parse failed'
  } finally {
    parsing.value = false
  }
}

function goBack() {
  phase.value = 'input'
  confirmError.value = ''
  confirmSuccess.value = ''
  imported.value = false
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
    const res = await confirmMutation({ input: { clubMatchId: clubMatchId.value, players } })
    const saved = res?.data?.confirmFFLTeamSubmission ?? []
    confirmSuccess.value = `Imported ${saved.length} player records.`
    imported.value = true
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
