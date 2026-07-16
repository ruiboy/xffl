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
      <div v-for="(rnd, ri) in rounds" :key="rnd.key" class="rounded-lg border border-border bg-surface-raised p-3 space-y-2">
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-xs text-text-faint w-6">{{ ri + 1 }}</span>
          <input
            v-model="rnd.name" :disabled="rnd.locked"
            class="w-40 rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60"
          />
          <label class="text-xs text-text-muted flex items-center gap-1">
            AFL
            <select v-model="rnd.aflRoundId" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
              <option value="">—</option>
              <option v-for="ar in aflRounds" :key="ar.id" :value="ar.id">{{ ar.name }}</option>
            </select>
          </label>
          <select v-model="rnd.roundType" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
            <option value="MINOR">MINOR</option>
            <option value="GRAND_FINAL">GRAND FINAL</option>
          </select>
          <label class="text-xs text-text-muted flex items-center gap-1" title="All clubs submit a team; the round's top scorer earns 1 point">
            <input type="checkbox" v-model="rnd.superbye" :disabled="rnd.locked" /> superbye
          </label>
          <span v-if="rnd.locked" class="text-xs rounded-full bg-surface px-2 py-0.5 text-text-faint" title="Has submitted teams — fixtures are locked">🔒 locked</span>
          <button
            @click="removeRound(ri)" :disabled="rnd.locked"
            class="ml-auto text-xs px-2 py-1 rounded border border-border hover:bg-surface disabled:opacity-40"
          >Remove round</button>
        </div>

        <!-- Superbye: every club submits, no head-to-head -->
        <p v-if="rnd.superbye" class="pl-6 text-xs text-text-faint">
          Superbye — all {{ clubs.length }} clubs submit a team; the highest scorer earns 1 point.
        </p>

        <!-- Fixtures -->
        <template v-if="!rnd.superbye">
        <div v-for="(f, fi) in rnd.fixtures" :key="fi" class="flex items-center gap-2 pl-6">
          <select v-model="f.home" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
            <option value="">home…</option>
            <option v-for="c in clubs" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <span class="text-text-faint text-xs">v</span>
          <select v-model="f.away" :disabled="rnd.locked" class="rounded border border-border bg-surface px-2 py-1 text-sm disabled:opacity-60">
            <option value="">away…</option>
            <option v-for="c in clubs" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
          <button v-if="!rnd.locked" @click="rnd.fixtures.splice(fi, 1)" class="text-xs text-text-faint hover:text-red-400">✕</button>
        </div>

        <div class="flex items-center gap-3 pl-6">
          <button
            v-if="!rnd.locked"
            @click="addMatch(rnd)" :disabled="remainingClubs(rnd).length < 2"
            class="text-xs px-2 py-1 rounded border border-border hover:bg-surface disabled:opacity-40"
          >+ match</button>
          <span v-if="byeClubs(rnd).length" class="text-xs text-text-faint">
            Bye: {{ byeClubs(rnd).map((c) => c.name).join(', ') }}
          </span>
        </div>
        </template>
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
  superbye: boolean
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
  rounds.value = loaded.map((r: any): StagedRound => ({
    key: keySeq++,
    roundId: r.roundId,
    name: r.name,
    aflRoundId: r.aflRoundId,
    roundType: r.roundType || 'MINOR',
    locked: r.locked,
    fixtures: r.fixtures.map((f: any) => ({ home: f.homeClubSeasonId, away: f.awayClubSeasonId })),
    superbye: (r.superbye?.length ?? 0) > 0,
  }))
}, { immediate: true })

function clubName(id: string) {
  return clubs.value.find((c) => c.id === id)?.name ?? id
}

function usedClubIds(r: StagedRound): Set<string> {
  const s = new Set<string>()
  for (const f of r.fixtures) {
    if (f.home) s.add(f.home)
    if (f.away) s.add(f.away)
  }
  return s
}
function remainingClubs(r: StagedRound) {
  const used = usedClubIds(r)
  return clubs.value.filter((c) => !used.has(c.id))
}
// Clubs not in any fixture are on a (scoring) bye this round.
function byeClubs(r: StagedRound) {
  return remainingClubs(r)
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
    aflRoundId, roundType: 'MINOR', locked: false, fixtures: [], superbye: false,
  })
}

function addMatch(r: StagedRound) {
  const rem = remainingClubs(r)
  if (rem.length >= 2) r.fixtures.push({ home: rem[0].id, away: rem[1].id })
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
      aflRoundId, roundType: src.roundType, locked: false, superbye: src.superbye,
      fixtures: src.fixtures.map((f) => repeatReverse.value ? { home: f.away, away: f.home } : { home: f.home, away: f.away }),
    })
  }
}

// ---- Validation + save ----
const validationError = computed(() => {
  for (const [i, r] of rounds.value.entries()) {
    if (r.locked) continue
    if (!r.aflRoundId) return `Round ${i + 1}: pick an AFL round.`
    if (r.superbye) continue // all clubs submit; nothing else to validate
    const seen = new Set<string>()
    for (const f of r.fixtures) {
      if (!f.home || !f.away) return `Round ${i + 1}: every match needs both clubs.`
      if (f.home === f.away) return `Round ${i + 1}: a club can't play itself.`
      for (const id of [f.home, f.away]) {
        if (seen.has(id)) return `Round ${i + 1}: ${clubName(id)} appears twice.`
        seen.add(id)
      }
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
        // A superbye round: every club is in the superbye, no head-to-head or byes.
        fixtures: r.superbye ? [] : r.fixtures.map((f) => ({ homeClubSeasonId: f.home, awayClubSeasonId: f.away })),
        byes: r.superbye ? [] : byeClubs(r).map((c) => c.id),
        superbye: r.superbye ? clubs.value.map((c) => c.id) : [],
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
