<template>
  <section class="rounded-lg border border-border p-4 space-y-4">
    <div class="flex items-center gap-3">
      <h3 class="text-sm font-semibold">Fixtures</h3>
      <label class="text-sm ml-auto flex items-center gap-2">
        <span class="text-text-muted">Season</span>
        <select v-model="seasonId" class="rounded border border-border bg-surface-raised px-2 py-1 text-sm">
          <option value="">—</option>
          <option v-for="s in seasons" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </label>
    </div>

    <p v-if="!seasonId" class="text-sm text-text-faint">Pick a season to edit its fixtures.</p>

    <template v-else>
      <!-- Rounds -->
      <div v-for="(rnd, ri) in rounds" :key="rnd.key" data-testid="round" class="rounded-lg border border-border bg-surface-raised">
        <!-- Header: what this round is -->
        <div class="flex flex-wrap items-center gap-2 p-3">
          <span class="w-5 shrink-0 text-xs tabular-nums text-text-faint">{{ ri + 1 }}</span>
          <input
            v-model="rnd.name" :disabled="rnd.locked" placeholder="Round name"
            class="w-40 rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60"
          />
          <label class="flex items-center gap-1.5 text-xs text-text-muted">
            AFL
            <select v-model="rnd.aflRoundId" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
              <option value="">—</option>
              <option v-for="ar in aflRounds" :key="ar.id" :value="ar.id">{{ ar.name }}</option>
            </select>
          </label>
          <select v-model="rnd.roundType" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
            <option value="MINOR">Minor</option>
            <option value="GRAND_FINAL">Grand final</option>
          </select>
          <div class="ml-auto flex items-center gap-2">
            <span
              v-if="rnd.locked"
              class="rounded-full bg-surface px-2 py-0.5 text-xs text-text-faint"
              title="Has submitted teams — fixtures are locked"
            >🔒 Locked</span>
            <button
              @click="removeRound(ri)" :disabled="rnd.locked"
              class="rounded border border-border px-2 py-1 text-xs hover:bg-surface disabled:opacity-40"
            >Remove</button>
          </div>
        </div>

        <!-- Body: matches are the round. Byes and the superbye are what's left
             over, so they sit below and stay quiet until they hold something. -->
        <div class="space-y-3 border-t border-border p-3">
          <!-- Matches: the one thing you do most, so it leads and owns the -->
          <!-- only button that looks like a button. -->
          <div class="space-y-1.5">
            <div v-for="(f, fi) in rnd.fixtures" :key="fi" data-testid="match" class="flex items-center gap-2">
              <select v-model="f.home" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
                <option value="">home…</option>
                <option v-for="c in pickable(rnd, f.home)" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
              <span class="text-xs text-text-faint">v</span>
              <select v-model="f.away" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
                <option value="">away…</option>
                <option v-for="c in pickable(rnd, f.away)" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
              <button
                v-if="!rnd.locked" @click="rnd.fixtures.splice(fi, 1)"
                class="text-xs text-text-faint hover:text-red-400" title="Remove match"
              >✕</button>
            </div>
            <button
              v-if="!rnd.locked"
              @click="addMatch(rnd)" :disabled="unassignedClubs(rnd).length < 2"
              class="rounded border border-border px-3 py-1.5 text-sm hover:bg-surface disabled:opacity-40"
            >+ Add match</button>
          </div>

          <!-- Clubs credited a bye: they score For but play nobody. -->
          <div v-if="rnd.byes.length" data-testid="byes" class="flex gap-3 pt-1">
            <span class="w-16 shrink-0 pt-0.5 text-xs uppercase tracking-wide text-text-faint">Byes</span>
            <div class="flex flex-1 flex-wrap gap-1.5">
              <span
                v-for="id in rnd.byes" :key="id" data-testid="bye-club"
                class="inline-flex items-center gap-1.5 rounded-full bg-surface px-2.5 py-0.5 text-xs"
              >
                {{ clubName(id) }}
                <button
                  v-if="!rnd.locked" @click="removeBye(rnd, id)"
                  class="text-text-faint hover:text-red-400" title="Remove bye"
                >✕</button>
              </span>
            </div>
          </div>

          <!-- The superbye: these clubs each submit, and the best of them takes a point. -->
          <div v-if="rnd.superbyeClubs.length" data-testid="superbye" class="flex gap-3 pt-1">
            <span class="w-16 shrink-0 pt-0.5 text-xs uppercase tracking-wide text-text-faint">Superbye</span>
            <div class="flex-1 space-y-1.5">
              <div class="flex flex-wrap gap-1.5">
                <span
                  v-for="id in rnd.superbyeClubs" :key="id" data-testid="superbye-club"
                  class="inline-flex items-center gap-1.5 rounded-full bg-surface px-2.5 py-0.5 text-xs"
                >
                  {{ clubName(id) }}
                  <button
                    v-if="!rnd.locked" @click="removeFromSuperbye(rnd, id)"
                    class="text-text-faint hover:text-red-400" title="Remove from superbye"
                  >✕</button>
                </span>
              </div>
              <p class="text-xs text-text-faint">
                These {{ rnd.superbyeClubs.length }} each submit a team; the highest scorer earns 1 point.
              </p>
            </div>
          </div>

          <!-- Whatever is left over. Plain sentence, link-styled names: the two
               group actions cover the usual cases, a name covers the odd one. -->
          <p
            v-if="!rnd.locked && unassignedClubs(rnd).length"
            data-testid="unplaced"
            class="flex flex-wrap items-baseline gap-x-1.5 gap-y-1 pt-1 text-xs text-text-faint"
          >
            <span>Not placed:</span>
            <template v-for="(c, ci) in unassignedClubs(rnd)" :key="c.id">
              <button
                @click="addBye(rnd, c.id)"
                data-testid="unplaced-club"
                class="underline decoration-dotted underline-offset-2 hover:text-active"
                title="Give this club a bye"
              >{{ c.name }}</button><span v-if="ci < unassignedClubs(rnd).length - 1">,</span>
            </template>
            <span class="text-text-faint/60">— click a club to give it a bye, or</span>
            <button
              @click="byeAll(rnd)" data-testid="bye-all"
              class="underline underline-offset-2 hover:text-active"
            >bye all</button>
            <span v-if="unassignedClubs(rnd).length > 1" class="text-text-faint/60">/</span>
            <button
              v-if="unassignedClubs(rnd).length > 1"
              @click="addSuperbye(rnd)" data-testid="superbye-all"
              class="underline underline-offset-2 hover:text-active"
            >superbye all</button>
          </p>
        </div>
      </div>

      <!-- Add round + repeat -->
      <div class="flex flex-wrap items-center gap-4 pt-1">
        <button @click="addRound" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised">+ Add round</button>

        <div class="flex items-center gap-2 text-sm">
          <span class="text-text-muted">Repeat rounds</span>
          <input v-model.number="repeatFrom" type="number" min="1" class="w-14 rounded border border-border bg-surface-raised px-2 py-1 text-sm" />
          <span class="text-text-faint">to</span>
          <input v-model.number="repeatTo" type="number" min="1" class="w-14 rounded border border-border bg-surface-raised px-2 py-1 text-sm" />
          <label class="flex items-center gap-1 text-text-muted">
            <input type="checkbox" v-model="repeatReverse" /> reverse H/A
          </label>
          <button @click="repeat" :disabled="!canRepeat" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised disabled:opacity-40">Repeat</button>
        </div>
      </div>

      <!-- Save -->
      <div class="flex items-center gap-3 border-t border-border pt-3">
        <button
          @click="save" :disabled="saving || validationError !== ''"
          class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >{{ saving ? 'Saving…' : 'Save fixtures' }}</button>
        <span v-if="validationError" class="text-sm text-red-400">{{ validationError }}</span>
        <span v-else-if="saveError" class="text-sm text-red-400">{{ saveError }}</span>
        <span v-else-if="savedAt" class="text-sm text-green-500">Saved ✓</span>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_SEASON_LIST, GET_FFL_BUILDER_SEASON, GET_FFL_SEASON_FIXTURES } from '../api/queries'
import { SAVE_FFL_FIXTURES } from '../api/mutations'

const props = defineProps<{ initialSeasonId?: string | null }>()

type StagedFixture = { home: string; away: string }
type StagedRound = {
  key: number
  roundId: string | null
  name: string
  aflRoundId: string
  roundType: string
  locked: boolean
  fixtures: StagedFixture[]
  byes: string[] // club_season ids explicitly given a bye
  superbyeClubs: string[] // club_season ids in the round's superbye, if it has one
}

const seasonId = ref(props.initialSeasonId ?? '')
watch(() => props.initialSeasonId, (v) => { if (v) seasonId.value = v })

const { result: seasonsResult } = useQuery(GET_FFL_SEASON_LIST)
const seasons = computed<{ id: string; name: string }[]>(
  () => [...(seasonsResult.value?.fflSeasons ?? [])].sort((a, b) => b.name.localeCompare(a.name)),
)

const { result: ctxResult } = useQuery(
  GET_FFL_BUILDER_SEASON,
  () => ({ id: seasonId.value }),
  () => ({ enabled: !!seasonId.value }),
)
const clubs = computed<{ id: string; name: string }[]>(
  () => (ctxResult.value?.fflSeason?.ladder ?? [])
    .map((cs: any) => ({ id: cs.id, name: cs.club.name }))
    .sort((a: any, b: any) => a.name.localeCompare(b.name)),
)
const aflRounds = computed<{ id: string; name: string }[]>(
  () => ctxResult.value?.fflSeason?.aflSeason?.rounds ?? [],
)

const { result: fixturesResult, refetch: refetchFixtures } = useQuery(
  GET_FFL_SEASON_FIXTURES,
  () => ({ seasonId: seasonId.value }),
  () => ({ enabled: !!seasonId.value }),
)

let keySeq = 0
const rounds = ref<StagedRound[]>([])

// Rebuild the staged model whenever the loaded fixtures change (season switch / save).
watch(fixturesResult, (val) => {
  const loaded = val?.fflSeasonFixtures ?? []
  rounds.value = loaded.map((r: any): StagedRound => {
    // The API models every match uniformly (style + clubSeasonIds). The editor
    // keeps versus pairings, explicit byes, and a superbye toggle.
    const versus = r.matches.filter((m: any) => m.style === 'versus')
    const byes = r.matches.filter((m: any) => m.style === 'bye').map((m: any) => m.clubSeasonIds[0])
    const superbye = r.matches.find((m: any) => m.style === 'superbye')
    return {
      key: keySeq++,
      roundId: r.roundId,
      name: r.name,
      aflRoundId: r.aflRoundId,
      roundType: r.roundType || 'MINOR',
      locked: r.locked,
      fixtures: versus.map((m: any) => ({ home: m.clubSeasonIds[0], away: m.clubSeasonIds[1] })),
      byes,
      superbyeClubs: superbye ? [...superbye.clubSeasonIds] : [],
    }
  })
}, { immediate: true })

function clubName(id: string) {
  return clubs.value.find((c) => c.id === id)?.name ?? id
}

// Clubs already placed in a fixture or given a bye this round.
function usedClubIds(r: StagedRound): Set<string> {
  const s = new Set<string>()
  for (const f of r.fixtures) {
    if (f.home) s.add(f.home)
    if (f.away) s.add(f.away)
  }
  for (const id of r.byes) s.add(id)
  for (const id of r.superbyeClubs) s.add(id)
  return s
}
// Clubs in neither a match nor a bye — they simply don't feature this round
// (e.g. a Grand Final round with only the two finalists).
function unassignedClubs(r: StagedRound) {
  const used = usedClubIds(r)
  return clubs.value.filter((c) => !used.has(c.id))
}

// A match dropdown offers the clubs still free, plus whatever it already holds —
// so a filled match stays readable without listing clubs placed elsewhere.
function pickable(r: StagedRound, currentId: string) {
  const free = unassignedClubs(r)
  if (!currentId || free.some((c) => c.id === currentId)) return free
  const current = clubs.value.find((c) => c.id === currentId)
  return current ? [current, ...free] : free
}

function addBye(r: StagedRound, clubId: string) {
  if (!r.byes.includes(clubId)) r.byes.push(clubId)
}
function byeAll(r: StagedRound) {
  for (const c of unassignedClubs(r)) addBye(r, c.id)
}
function removeBye(r: StagedRound, clubId: string) {
  r.byes = r.byes.filter((id) => id !== clubId)
}

// A round has a superbye when any club is in it — there is no separate mode. The
// common case (the whole competition in one) is the "put all" shortcut.
function addToSuperbye(r: StagedRound, clubId: string) {
  if (!r.superbyeClubs.includes(clubId)) r.superbyeClubs.push(clubId)
}
function removeFromSuperbye(r: StagedRound, clubId: string) {
  r.superbyeClubs = r.superbyeClubs.filter((id) => id !== clubId)
}
// Adding a superbye enrols everyone not already playing — the usual case is the
// whole competition. Clubs can then be dropped, or added back from Not playing.
function addSuperbye(r: StagedRound) {
  for (const c of unassignedClubs(r)) addToSuperbye(r, c.id)
}

// Translate the editor's staged round into the API's uniform match list: a
// superbye is one superbye match of all clubs; otherwise versus matches from the
// pairings plus a bye match per explicitly-chosen bye. Unassigned clubs are sent
// as nothing — they don't feature this round.
function roundMatches(r: StagedRound): { style: string; clubSeasonIds: string[] }[] {
  const matches = r.fixtures.map((f) => ({ style: 'versus', clubSeasonIds: [f.home, f.away] }))
  for (const id of r.byes) {
    matches.push({ style: 'bye', clubSeasonIds: [id] })
  }
  if (r.superbyeClubs.length) {
    matches.push({ style: 'superbye', clubSeasonIds: [...r.superbyeClubs] })
  }
  return matches
}

function aflRoundIndex(id: string) {
  return aflRounds.value.findIndex((ar) => ar.id === id)
}
// Default the next round's AFL round to the one following `afterId` in sequence.
function nextAflRoundId(afterId: string): string {
  const i = aflRoundIndex(afterId)
  if (i >= 0 && i + 1 < aflRounds.value.length) return aflRounds.value[i + 1].id
  return afterId
}

function addRound() {
  const last = rounds.value[rounds.value.length - 1]
  const aflRoundId = last ? nextAflRoundId(last.aflRoundId) : (aflRounds.value[0]?.id ?? '')
  rounds.value.push({
    key: keySeq++, roundId: null, name: `Round ${rounds.value.length + 1}`,
    aflRoundId, roundType: 'MINOR', locked: false, fixtures: [], byes: [], superbyeClubs: [],
  })
}

// Seed the new match with the first two free clubs — usually right, and always
// faster to correct than to fill from empty.
function addMatch(r: StagedRound) {
  const free = unassignedClubs(r)
  if (free.length >= 2) r.fixtures.push({ home: free[0].id, away: free[1].id })
}

function removeRound(i: number) {
  if (!rounds.value[i].locked) rounds.value.splice(i, 1)
}

// ---- Repeat rounds X–Y ----
const repeatFrom = ref(1)
const repeatTo = ref(1)
const repeatReverse = ref(false)
const canRepeat = computed(() =>
  repeatFrom.value >= 1 && repeatTo.value >= repeatFrom.value && repeatTo.value <= rounds.value.length,
)
function repeat() {
  if (!canRepeat.value) return
  const slice = rounds.value.slice(repeatFrom.value - 1, repeatTo.value)
  for (const src of slice) {
    const last = rounds.value[rounds.value.length - 1]
    const aflRoundId = last ? nextAflRoundId(last.aflRoundId) : src.aflRoundId
    rounds.value.push({
      key: keySeq++, roundId: null, name: `Round ${rounds.value.length + 1}`,
      aflRoundId, roundType: src.roundType, locked: false,
      fixtures: src.fixtures.map((f) => repeatReverse.value ? { home: f.away, away: f.home } : { home: f.home, away: f.away }),
      byes: [...src.byes],
      superbyeClubs: [...src.superbyeClubs],
    })
  }
}

// ---- Validation + save ----
const validationError = computed(() => {
  for (const [i, r] of rounds.value.entries()) {
    if (r.locked) continue
    if (!r.aflRoundId) return `Round ${i + 1}: pick an AFL round.`
    // A club takes part in a round exactly once, whichever way it takes part.
    const seen = new Set<string>()
    for (const f of r.fixtures) {
      if (!f.home || !f.away) return `Round ${i + 1}: every match needs both clubs.`
      if (f.home === f.away) return `Round ${i + 1}: a club can't play itself.`
      for (const id of [f.home, f.away]) {
        if (seen.has(id)) return `Round ${i + 1}: ${clubName(id)} appears twice.`
        seen.add(id)
      }
    }
    for (const id of r.byes) {
      if (seen.has(id)) return `Round ${i + 1}: ${clubName(id)} is in a match and a bye.`
      seen.add(id)
    }
    for (const id of r.superbyeClubs) {
      if (seen.has(id)) return `Round ${i + 1}: ${clubName(id)} is in the superbye and a match or bye.`
      seen.add(id)
    }
    // One club alone has no field to top — that's a bye, not a superbye.
    if (r.superbyeClubs.length === 1) {
      return `Round ${i + 1}: a superbye needs at least two clubs — give ${clubName(r.superbyeClubs[0])} a bye instead.`
    }
  }
  return ''
})

const saving = ref(false)
const saveError = ref('')
const savedAt = ref(0)
const { mutate: saveMut } = useMutation(SAVE_FFL_FIXTURES)

async function save() {
  if (validationError.value) return
  saving.value = true
  saveError.value = ''
  try {
    const input = {
      seasonId: seasonId.value,
      rounds: rounds.value.map((r) => ({
        roundId: r.roundId,
        name: r.name,
        aflRoundId: r.aflRoundId,
        roundType: r.roundType,
        matches: roundMatches(r),
      })),
    }
    await saveMut({ input })
    await refetchFixtures()
    savedAt.value = Date.now()
  } catch (e: any) {
    saveError.value = e?.message ?? 'Save failed'
  } finally {
    saving.value = false
  }
}
</script>
