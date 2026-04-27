<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">Data Ops <span class="font-normal text-text-muted">· Team Submission</span></h1>
    </div>

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
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_DATA_OPS } from '../api/queries'
import { PARSE_TEAM_SUBMISSION, CONFIRM_TEAM_SUBMISSION } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { POSITION_COLORS, POSITION_LABEL, POSITION_SLOTS } from '@/features/ffl/utils/position'

const { liveSeasonId, liveRoundId } = useFflState()

// ---- Season data ----
const { result: seasonResult, loading: loadingSeasonData, error: seasonError } = useQuery(
  GET_FFL_DATA_OPS,
  () => ({ seasonId: liveSeasonId.value }),
  () => ({ enabled: !!liveSeasonId.value }),
)

const season = computed(() => seasonResult.value?.fflSeason ?? null)

// ---- Selections ----
const selectedClubSeasonId = ref('')
const selectedRoundId = ref(liveRoundId.value)
const teamName = ref('')
const post = ref('')

// Derived: find the club_match_id for the selected club in the selected round
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

// ---- Parse ----
type ResolvedPlayer = {
  parsedName: string
  clubHint: string
  resolvedName: string | null
  resolvedClub: string | null
  position: string
  backupPositions: string
  interchangePosition: string
  score: number | null
  notes: string
  playerSeasonId: string | null
  confidence: number
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
      input: {
        clubSeasonId: selectedClubSeasonId.value,
        clubMatchId: clubMatchId.value,
        teamName: teamName.value,
        post: post.value,
      },
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

// ---- Confirm ----
const confirming = ref(false)
const confirmError = ref('')
const confirmSuccess = ref('')
const imported = ref(false)

const { mutate: confirmMutation } = useMutation(CONFIRM_TEAM_SUBMISSION)

const unresolvedCount = computed(() =>
  resolvedPlayers.value.filter(rp => !rp.playerSeasonId).length,
)

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

    const res = await confirmMutation({
      input: {
        clubMatchId: clubMatchId.value,
        players,
      },
    })
    const saved = res?.data?.confirmFFLTeamSubmission ?? []
    confirmSuccess.value = `Imported ${saved.length} player records.`
    imported.value = true
  } catch (e: any) {
    confirmError.value = e.message ?? 'Confirm failed'
  } finally {
    confirming.value = false
  }
}

// ---- Display helpers ----
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
