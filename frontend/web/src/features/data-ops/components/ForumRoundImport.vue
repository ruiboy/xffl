<template>
  <div class="space-y-5">
    <!-- Season + round + capture controls -->
    <div class="flex flex-wrap items-end gap-4">
      <div>
        <label class="block text-xs font-medium text-text-muted mb-1">Season</label>
        <select v-model="seasonId" class="rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active">
          <option v-if="!seasonOptions.length" value="">Loading…</option>
          <option v-for="s in seasonOptions" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>
      <div>
        <label class="block text-xs font-medium text-text-muted mb-1">Round</label>
        <select v-model="roundId" class="rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active">
          <option value="">Select round…</option>
          <option v-for="r in rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
        </select>
      </div>
      <div class="ml-auto flex gap-2">
        <button @click="refetchCaptured()" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised">Refresh capture</button>
        <button @click="clearCaptures()" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised">Clear</button>
      </div>
    </div>

    <!-- Capture summary + build -->
    <div class="rounded-lg border border-border p-3 text-sm">
      <div v-if="!capturedPages.length" class="text-text-faint">
        No pages captured this session. Use <code>ffl-forum-capture.user.js</code> on a round thread — capture <em>every</em> page, then Refresh.
      </div>
      <div v-else class="flex flex-wrap items-center gap-3">
        <span class="text-text-muted">{{ totalPosts }} posts across {{ capturedPages.length }} page(s){{ threadSummary }} · {{ teamPostStats.initial }} look like initial teams, {{ teamPostStats.final }} look like final teams.</span>
        <button
          @click="buildTeams"
          class="ml-auto rounded-lg border border-active bg-active px-4 py-1.5 text-sm font-medium text-active-text transition-colors"
        >Build teams</button>
      </div>
    </div>

    <p v-if="!roundId && candidates.length" class="text-sm text-yellow-500">Select a round to resolve and import into.</p>

    <!-- Per-club team candidates -->
    <div
      v-for="(c, ci) in candidates"
      :key="c.key"
      class="rounded-xl border p-4"
      :class="c.clubSeasonId ? 'border-border bg-surface-raised' : 'border-yellow-500/60 ring-1 ring-yellow-500/30 bg-surface-raised'"
    >
      <!-- Header: attribution -->
      <div class="flex flex-wrap items-center gap-3">
        <div v-if="c.clubSeasonId" class="text-sm font-semibold text-text">{{ c.clubName }}</div>
        <div v-else class="text-sm text-yellow-500">Unknown club — posted by <span class="font-mono">@{{ c.author }}</span></div>
        <select
          v-model="c.clubSeasonId"
          @change="onClubChange(c)"
          :disabled="c.imported"
          class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:outline-none focus:border-active disabled:opacity-50"
        >
          <option value="">Assign club…</option>
          <option v-for="cs in clubSeasons" :key="cs.id" :value="cs.id">{{ cs.club.name }}</option>
        </select>
        <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs" :class="c.scored ? 'bg-green-500/15 text-green-500' : 'bg-surface px-2 text-text-faint'">
          {{ c.scored ? 'scored' : 'no scores' }}
        </span>
        <span class="text-xs text-text-faint">{{ c.players.length }} players · forum {{ c.postedScore ?? '—' }}</span>

        <div class="ml-auto flex items-center gap-2">
          <span v-if="c.error" class="text-xs text-red-400">{{ c.error }}</span>
          <button
            v-if="c.phase === 'preview'"
            @click="resolve(c)"
            :disabled="!c.clubSeasonId || !roundId || c.busy"
            class="rounded-lg border border-active bg-active px-4 py-1.5 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ c.busy ? 'Reading…' : 'Resolve' }}</button>
        </div>
      </div>

      <!-- Imported: derived vs forum vs fixtures, above the team -->
      <div v-if="c.imported" class="mt-3 flex flex-wrap items-center gap-4 text-sm">
        <span class="text-green-500 font-medium">Imported</span>
        <span class="text-text-muted">Derived <span class="tabular-nums text-text">{{ c.derived ?? '—' }}</span></span>
        <span class="text-text-muted">Forum <span class="tabular-nums text-text">{{ c.postedScore ?? '—' }}</span></span>
        <span class="text-text-muted">Fixtures <span class="tabular-nums text-text">{{ c.fixturesRef ?? '—' }}</span></span>
        <span v-if="hasDelta(c)" class="text-xs text-yellow-500">differs — noted, no action</span>
        <router-link
          v-if="c.clubMatchId"
          :to="{ name: 'ffl-club-match-edit', params: { clubMatchId: c.clubMatchId } }"
          target="_blank"
          class="ml-auto text-active hover:underline text-xs"
        >Open team builder ↗</router-link>
      </div>

      <!-- Review phase: resolved players -->
      <div v-if="c.phase === 'review'" class="mt-3">
        <div class="mb-2 flex items-center gap-3 text-sm">
          <button @click="c.phase = 'preview'" class="rounded border border-border px-2 py-1 text-xs hover:bg-surface-hover">← Back</button>
          <span :class="unresolvedCount(c) > 0 ? 'text-red-400' : 'text-green-500'">{{ unresolvedCount(c) }} unresolved</span>
          <button
            @click="confirmImport(c, ci)"
            :disabled="unresolvedCount(c) > 0 || c.busy"
            class="ml-auto rounded-lg border border-active bg-active px-4 py-1.5 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ c.busy ? 'Importing…' : 'Confirm & Import' }}</button>
        </div>
        <table class="w-full">
          <tbody>
            <tr
              v-for="(rp, ri) in c.resolved"
              :key="ri"
              class="border-b border-border last:border-0"
              :class="!rp.playerSeasonId ? 'bg-red-500/5' : (rp.confidence < 1 ? 'bg-yellow-500/5' : '')"
            >
              <td class="py-1.5 pr-3 font-mono text-sm">{{ rp.parsedName }} <span class="text-xs text-text-faint">{{ rp.clubHint }}</span></td>
              <td class="py-1.5 px-1 text-text-faint text-sm select-none">→</td>
              <td class="py-1.5 pr-3 text-sm">
                <span v-if="rp.resolvedName" class="font-semibold text-text">{{ rp.resolvedName }}</span>
                <span v-else class="text-red-400">Unresolved</span>
                <span class="ml-2 text-xs text-text-muted">{{ rp.resolvedClub ?? '' }}</span>
              </td>
              <td class="py-1.5 pr-3 text-xs text-text-muted whitespace-nowrap">{{ rp.position }}{{ rp.backupPositions ? ' (' + rp.backupPositions + ')' : '' }}</td>
              <td class="py-1.5 pr-3 text-right text-sm tabular-nums text-text-muted">{{ rp.score ?? '' }}</td>
              <td class="py-1.5 text-right">
                <span class="inline-block rounded-full px-2 py-0.5 text-xs" :class="confidenceBadge(rp.confidence)">{{ (rp.confidence * 100).toFixed(0) }}%</span>
              </td>
              <td class="py-1.5 pl-2 text-right">
                <button v-if="rp.confidence < 1" @click="openFix(ci, ri)" class="rounded border border-border px-2 py-0.5 text-xs hover:bg-surface-hover">Fix</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

    </div>

    <!-- Step 9: round link + move on, once all imported -->
    <div v-if="candidates.length && allImported && roundId" class="rounded-lg border border-green-500/40 bg-green-500/5 p-3 text-sm flex flex-wrap items-center gap-3">
      <span class="text-green-500 font-medium">All teams imported.</span>
      <router-link :to="{ name: 'ffl-round', params: { roundId } }" target="_blank" class="text-active hover:underline">Open round page ↗</router-link>
      <button
        @click="nextRound"
        class="ml-auto rounded-lg border border-active bg-active px-4 py-1.5 text-sm font-medium text-active-text transition-colors"
      >Done — next round</button>
    </div>

    <FflPlayerLinkModal
      :show="fixTarget !== null"
      :ffl-season-id="seasonId"
      :ffl-club-season-id="fixCandidate?.clubSeasonId ?? ''"
      :parsed-name="fixRow?.parsedName ?? ''"
      :club-hint="fixRow?.clubHint ?? ''"
      @close="fixTarget = null"
      @linked="onFixLinked"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation, useApolloClient } from '@vue/apollo-composable'
import { GET_FFL_CAPTURED_PAGES, GET_FFL_CAPTURE_SEASON, GET_FFL_CLUB_MATCH_SCORE } from '../api/queries'
import { GET_FFL_SEASONS } from '@/features/ffl/api/queries'
import { PARSE_TEAM_SUBMISSION, CONFIRM_TEAM_SUBMISSION, CLEAR_FFL_FORUM_CAPTURES } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { canonicalClub } from '../utils/clubAliases'
import { parseReferenceScores } from '../utils/referenceScores'
import FflPlayerLinkModal from './FflPlayerLinkModal.vue'

const { liveSeasonId } = useFflState()
const { resolveClient } = useApolloClient()

// ---- Season + round ----
const { result: seasonsResult } = useQuery(GET_FFL_SEASONS)
const seasonOptions = computed<{ id: string; name: string }[]>(() =>
  [...(seasonsResult.value?.fflSeasons ?? [])].sort((a: any, b: any) => b.name.localeCompare(a.name)),
)

const seasonId = ref(liveSeasonId.value)
watch(liveSeasonId, (v) => { if (v && !seasonId.value) seasonId.value = v }, { immediate: true })
const roundId = ref('')

const { result: seasonResult, refetch: refetchSeason } = useQuery(
  GET_FFL_CAPTURE_SEASON,
  () => ({ id: seasonId.value }),
  () => ({ enabled: !!seasonId.value }),
)
const season = computed(() => seasonResult.value?.fflSeason ?? null)
const clubSeasons = computed(() => season.value?.ladder ?? [])
const rounds = computed(() => season.value?.rounds ?? [])
const selectedRound = computed(() => rounds.value.find((r: any) => r.id === roundId.value) ?? null)

// clubSeasonId → { clubMatchId } for the selected round.
const clubMatchBySeason = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const m of selectedRound.value?.matches ?? []) {
    for (const cm of m.clubMatches ?? []) map[cm.clubSeasonId] = cm.id
  }
  return map
})

// Switching season resets built candidates (different clubs); round can be picked
// before or after building without discarding the work.
watch(seasonId, () => { candidates.value = []; roundId.value = '' })

// ---- Captured pages ----
const { result: capturedResult, refetch: refetchCaptured } = useQuery(GET_FFL_CAPTURED_PAGES)
const capturedPages = computed(() => capturedResult.value?.fflCapturedPages ?? [])
const totalPosts = computed(() => capturedPages.value.reduce((n: number, p: any) => n + p.posts.length, 0))
// Team posts split by whether they carry scores: a scored post is a "final" team
// (points included, posted after the round), an unscored one an "initial" team.
const teamPostStats = computed(() => {
  let initial = 0, final = 0
  for (const p of capturedPages.value) {
    for (const post of p.posts) {
      if (!post.players?.length) continue
      if (post.players.some((pl: any) => pl.score !== null && pl.score !== undefined)) final++
      else initial++
    }
  }
  return { initial, final }
})
const threadSummary = computed(() => {
  const topics = new Set(capturedPages.value.map((p: any) => p.topicId))
  return topics.size > 1 ? ` (${topics.size} threads)` : ''
})

const { mutate: clearMutation } = useMutation(CLEAR_FFL_FORUM_CAPTURES)
async function clearCaptures() {
  await clearMutation()
  await refetchCaptured()
  candidates.value = []
}

// Finish this round and set up for the next one in the same season: drop the
// in-session capture buffer and built teams, and clear the round so the next
// thread's captures start clean.
async function nextRound() {
  await clearCaptures()
  roundId.value = ''
}

// ---- Build teams: classify → attribute → authoritative post per club ----
type Candidate = {
  key: string
  clubSeasonId: string
  clubName: string
  clubMatchId: string
  author: string
  format: string
  text: string
  players: any[]
  postedScore: number | null
  scored: boolean
  phase: 'preview' | 'review'
  imported: boolean
  busy: boolean
  error: string
  resolved: any[]
  needsReview: number[]
  derived: number | null
  fixturesRef: number | null
}

const candidates = ref<Candidate[]>([])

function attributeClub(format: string): { clubSeasonId: string; clubName: string } {
  if (!format) return { clubSeasonId: '', clubName: '' }
  const key = canonicalClub(format)
  const cs = clubSeasons.value.find((c: any) => canonicalClub(c.club.name) === key)
  return cs ? { clubSeasonId: cs.id, clubName: cs.club.name } : { clubSeasonId: '', clubName: '' }
}

function buildTeams() {
  // Flatten every captured team post (has players) in capture order.
  const posts: any[] = []
  for (const page of capturedPages.value) {
    for (const post of page.posts) {
      if (!post.players?.length) continue // banter / non-team
      posts.push(post)
    }
  }

  // Group by club (attributed), else by format, else by author; keep the latest
  // scored post per group (the authoritative submission), else the latest team.
  const groups = new Map<string, Candidate>()
  for (const post of posts) {
    const scored = post.players.some((p: any) => p.score !== null && p.score !== undefined)
    const postedScore = scored ? post.players.reduce((n: number, p: any) => n + (p.score ?? 0), 0) : null
    const attr = attributeClub(post.team)
    const key = attr.clubSeasonId ? `cs:${attr.clubSeasonId}` : (post.team ? `fmt:${post.team}` : `author:${post.author}`)

    const cand: Candidate = {
      key,
      clubSeasonId: attr.clubSeasonId,
      clubName: attr.clubName,
      clubMatchId: '',
      author: post.author,
      format: post.team,
      text: post.text,
      players: post.players,
      postedScore,
      scored,
      phase: 'preview',
      imported: false,
      busy: false,
      error: '',
      resolved: [],
      needsReview: [],
      derived: null,
      fixturesRef: null,
    }

    const existing = groups.get(key)
    // Prefer a scored post; between two of the same scored-ness, the later wins.
    if (!existing || (cand.scored && !existing.scored) || cand.scored === existing.scored) {
      groups.set(key, cand)
    }
  }

  candidates.value = [...groups.values()].sort((a, b) =>
    (a.clubSeasonId ? 0 : 1) - (b.clubSeasonId ? 0 : 1) || a.clubName.localeCompare(b.clubName),
  )
}

function onClubChange(c: Candidate) {
  const cs = clubSeasons.value.find((x: any) => x.id === c.clubSeasonId)
  c.clubName = cs?.club.name ?? ''
}

// ---- Resolve (reuses the team-submission parse) ----
const { mutate: parseMutation } = useMutation(PARSE_TEAM_SUBMISSION)
const { mutate: confirmMutation } = useMutation(CONFIRM_TEAM_SUBMISSION)

async function resolve(c: Candidate) {
  const clubMatchId = clubMatchBySeason.value[c.clubSeasonId]
  if (!clubMatchId) { c.error = 'This club has no match in the selected round.'; return }
  c.clubMatchId = clubMatchId
  c.error = ''
  c.busy = true
  try {
    const res = await parseMutation({
      input: { clubSeasonId: c.clubSeasonId, clubMatchId, teamName: c.format, post: c.text },
    })
    const data = res?.data?.parseFFLTeamSubmission
    if (!data) throw new Error('No result returned')
    c.resolved = data.resolvedPlayers.map((p: any) => ({ ...p }))
    c.needsReview = data.needsReview
    c.phase = 'review'
  } catch (e: any) {
    c.error = e.message ?? 'Resolve failed'
  } finally {
    c.busy = false
  }
}

function unresolvedCount(c: Candidate): number {
  return c.resolved.filter(rp => !rp.playerSeasonId).length
}

async function confirmImport(c: Candidate, _ci: number) {
  const clubMatchId = clubMatchBySeason.value[c.clubSeasonId]
  if (!clubMatchId) { c.error = 'This club has no match in the selected round.'; return }
  c.error = ''
  c.busy = true
  try {
    const players = c.resolved
      .filter(rp => rp.playerSeasonId)
      .map(rp => ({
        playerSeasonId: rp.playerSeasonId,
        position: rp.position,
        backupPositions: rp.backupPositions || null,
        interchangePosition: rp.interchangePosition || null,
        score: rp.score,
      }))
    await confirmMutation({ input: { clubMatchId, players } })

    // Read back the derived score + reference notes for the inline comparison.
    const client = resolveClient()
    const back = await client.query({
      query: GET_FFL_CLUB_MATCH_SCORE,
      variables: { id: clubMatchId },
      fetchPolicy: 'network-only',
    })
    const cm = back.data?.fflClubMatch
    c.derived = cm?.score ?? null
    c.fixturesRef = parseReferenceScores(cm?.notes).spreadsheet
    c.imported = true
    await refetchSeason()
  } catch (e: any) {
    const gqlErr = e?.graphQLErrors?.[0]
    c.error = gqlErr?.message ?? e.message ?? 'Import failed'
  } finally {
    c.busy = false
  }
}

function hasDelta(c: Candidate): boolean {
  if (c.derived === null) return false
  return (c.postedScore !== null && c.derived !== c.postedScore) ||
    (c.fixturesRef !== null && c.derived !== c.fixturesRef)
}

const allImported = computed(() => candidates.value.length > 0 && candidates.value.every(c => c.imported))

// ---- Fix modal ----
const fixTarget = ref<{ ci: number; ri: number } | null>(null)
const fixCandidate = computed(() => fixTarget.value ? candidates.value[fixTarget.value.ci] : null)
const fixRow = computed(() => fixTarget.value ? candidates.value[fixTarget.value.ci]?.resolved[fixTarget.value.ri] : null)

function openFix(ci: number, ri: number) { fixTarget.value = { ci, ri } }

function onFixLinked(data: { playerSeasonId: string; resolvedName: string; resolvedClub: string }) {
  if (!fixTarget.value) return
  const { ci, ri } = fixTarget.value
  const rp = candidates.value[ci]?.resolved[ri]
  if (rp) {
    rp.playerSeasonId = data.playerSeasonId
    rp.resolvedName = data.resolvedName
    rp.resolvedClub = data.resolvedClub
    rp.confidence = 1
  }
  fixTarget.value = null
}

function confidenceBadge(confidence: number): string {
  if (confidence >= 0.85) return 'bg-green-500/15 text-green-500'
  if (confidence >= 0.6) return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-red-500/15 text-red-400'
}
</script>
