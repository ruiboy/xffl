<template>
  <!-- Skeleton builder for the historical import (slice 2c). Functional but plain —
       intended as a base to redirect the UX (round-robin copy-fill, byes, finals). -->
  <div class="space-y-6">
    <!-- 1. Create season -->
    <section class="rounded-lg border border-border p-4 space-y-3">
      <h3 class="text-sm font-semibold">1 · Create season</h3>
      <div class="flex flex-wrap items-start gap-4">
        <label class="text-sm">
          <span class="block text-text-muted mb-1">FFL season name</span>
          <input v-model="seasonName" class="w-40 rounded border border-border bg-surface-raised px-2 py-1 text-sm" placeholder="2024" />
        </label>
        <label class="text-sm">
          <span class="block text-text-muted mb-1">AFL season</span>
          <select v-model="aflSeasonId" @change="preselectRules" class="rounded border border-border bg-surface-raised px-2 py-1 text-sm">
            <option value="">—</option>
            <option v-for="s in aflSeasons" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
        </label>
        <label class="text-sm min-w-[18rem]">
          <span class="block text-text-muted mb-1">Scoring rules</span>
          <select v-model="rulesId" class="w-full rounded border border-border bg-surface-raised px-2 py-1 text-sm">
            <option value="">—</option>
            <option v-for="e in rulesEras" :key="e.id" :value="e.id">{{ e.id }} · {{ e.label }}</option>
          </select>
        </label>
      </div>

      <div class="text-sm">
        <span class="block text-text-muted mb-1">Clubs</span>
        <div class="flex flex-wrap gap-x-4 gap-y-1.5 rounded border border-border bg-surface-raised p-3 max-h-56 overflow-y-auto">
          <label v-for="c in clubs" :key="c.id" class="flex items-center gap-1.5 cursor-pointer min-w-[10rem]">
            <input type="checkbox" :value="c.id" v-model="selectedClubIds" />
            <span>{{ c.name }}</span>
          </label>
          <span v-if="clubs.length === 0" class="text-text-faint">No clubs registered.</span>
        </div>
      </div>

      <div class="flex items-center gap-3">
        <button
          @click="createSeason"
          :disabled="!seasonName || !aflSeasonId || !rulesId || selectedClubIds.length === 0"
          class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        >Create</button>
        <span v-if="built" class="text-sm text-text-muted">
          ✓ season <b>{{ built.seasonId }}</b>, era <b>{{ built.rulesId }}</b> ·
          {{ built.clubSeasons.map((c) => c.clubName).join(', ') }}
        </span>
      </div>
    </section>

    <!-- 2. Auto-generate the home-and-away season -->
    <section v-if="built" class="rounded-lg border border-border p-4 space-y-3">
      <h3 class="text-sm font-semibold">2 · Generate home-and-away (round-robin)</h3>
      <div class="flex flex-wrap items-end gap-3">
        <label class="text-sm">
          <span class="block text-text-muted mb-1">Rounds</span>
          <input v-model.number="genRounds" type="number" class="w-24 rounded border border-border bg-surface-raised px-2 py-1 text-sm" />
        </label>
        <label class="text-sm">
          <span class="block text-text-muted mb-1">AFL round start id</span>
          <input v-model.number="genAflStart" type="number" class="w-32 rounded border border-border bg-surface-raised px-2 py-1 text-sm" />
        </label>
        <button @click="generate" :disabled="!genRounds || !genAflStart" class="text-sm px-3 py-1.5 rounded bg-active text-white disabled:opacity-40">Generate season</button>
        <span v-if="generated.length" class="text-sm text-text-muted">✓ created {{ generated.length }} rounds</span>
      </div>
      <p class="text-xs text-text-faint">Fills {{ genRounds }} rounds of round-robin fixtures over all {{ built.clubSeasons.length }} clubs. Byes, superbye and finals are added by hand below.</p>
    </section>

    <!-- 3. Add rounds + fixtures manually (finals, byes, adjustments) -->
    <section v-if="built" class="rounded-lg border border-border p-4 space-y-4">
      <h3 class="text-sm font-semibold">3 · Rounds & fixtures (manual)</h3>

      <div class="flex flex-wrap items-end gap-3">
        <label class="text-sm">
          <span class="block text-text-muted mb-1">Round name</span>
          <input v-model="roundName" class="w-40 rounded border border-border bg-surface-raised px-2 py-1 text-sm" placeholder="Round 1" />
        </label>
        <label class="text-sm">
          <span class="block text-text-muted mb-1">AFL round id</span>
          <input v-model.number="aflRoundId" type="number" class="w-28 rounded border border-border bg-surface-raised px-2 py-1 text-sm" />
        </label>
        <label class="text-sm">
          <span class="block text-text-muted mb-1">Type</span>
          <select v-model="roundType" class="rounded border border-border bg-surface-raised px-2 py-1 text-sm">
            <option value="MINOR">MINOR</option>
            <option value="GRAND_FINAL">GRAND_FINAL</option>
          </select>
        </label>
        <button @click="addRound" :disabled="!roundName || !aflRoundId" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised disabled:opacity-40">Add round</button>
      </div>

      <div v-for="rnd in rounds" :key="rnd.roundId" class="rounded border border-border p-3 space-y-2">
        <div class="text-sm font-medium">{{ rnd.name }} <span class="text-text-faint">(round {{ rnd.roundId }})</span></div>
        <div class="flex flex-wrap items-end gap-2">
          <select v-model="fixtureHome[rnd.roundId]" class="rounded border border-border bg-surface-raised px-2 py-1 text-sm">
            <option value="">home…</option>
            <option v-for="c in built.clubSeasons" :key="c.clubSeasonId" :value="c.clubSeasonId">{{ c.clubName }}</option>
          </select>
          <span class="text-text-faint">v</span>
          <select v-model="fixtureAway[rnd.roundId]" class="rounded border border-border bg-surface-raised px-2 py-1 text-sm">
            <option value="">away…</option>
            <option v-for="c in built.clubSeasons" :key="c.clubSeasonId" :value="c.clubSeasonId">{{ c.clubName }}</option>
          </select>
          <button
            @click="addFixture(rnd.roundId)"
            :disabled="!fixtureHome[rnd.roundId] || !fixtureAway[rnd.roundId] || fixtureHome[rnd.roundId] === fixtureAway[rnd.roundId]"
            class="text-sm px-3 py-1 rounded border border-border hover:bg-surface-raised disabled:opacity-40"
          >Add fixture</button>
        </div>
        <div v-for="f in fixturesByRound[rnd.roundId] || []" :key="f.matchId" class="text-xs text-text-faint">
          match {{ f.matchId }} · {{ clubName(f.home) }} v {{ clubName(f.away) }}
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_AFL_SEASONS } from '@/features/afl/api/queries'
import { GET_FFL_BUILDER_REFS } from '../api/queries'
import { BUILD_FFL_SEASON, ADD_FFL_ROUND, ADD_FFL_FIXTURE, GENERATE_FFL_HOME_AND_AWAY } from '../api/mutations'

const { result: aflSeasonsResult } = useQuery(GET_AFL_SEASONS)
const aflSeasons = computed<{ id: string; name: string }[]>(
  () => [...(aflSeasonsResult.value?.aflSeasons ?? [])].sort((a, b) => b.name.localeCompare(a.name)),
)

const { result: refsResult } = useQuery(GET_FFL_BUILDER_REFS)
const clubs = computed<{ id: string; name: string }[]>(
  () => [...(refsResult.value?.fflClubs ?? [])].sort((a, b) => a.name.localeCompare(b.name)),
)
// Eras arrive ascending by id; show newest first so the common recent pick is on top.
const rulesEras = computed<{ id: string; label: string }[]>(
  () => [...(refsResult.value?.fflRulesEras ?? [])].reverse(),
)

const seasonName = ref('')
const aflSeasonId = ref('')
const rulesId = ref('')
const selectedClubIds = ref<string[]>([])

// Pre-select the era for the AFL season's year: the latest era whose start year
// (era ids are start years) is ≤ that year. Best-effort — fully overridable.
function preselectRules() {
  const s = aflSeasons.value.find((x) => x.id === aflSeasonId.value)
  const year = Number(s?.name.match(/(\d{4})/)?.[1])
  if (!year || rulesEras.value.length === 0) return
  const ascending = [...rulesEras.value].sort((a, b) => Number(a.id) - Number(b.id))
  const chosen = ascending.filter((e) => Number(e.id) <= year).at(-1) ?? ascending[0]
  rulesId.value = chosen.id
}

type Built = { seasonId: string; rulesId: string; clubSeasons: { clubName: string; clubSeasonId: string }[] }
const built = ref<Built | null>(null)

const { mutate: buildSeasonMut } = useMutation(BUILD_FFL_SEASON)
async function createSeason() {
  const res = await buildSeasonMut({
    input: {
      seasonName: seasonName.value,
      rulesId: rulesId.value,
      aflSeasonId: aflSeasonId.value,
      clubIds: selectedClubIds.value,
    },
  })
  built.value = res?.data?.buildFFLSeason ?? null
}

// auto-generate H&A
const genRounds = ref<number | null>(null)
const genAflStart = ref<number | null>(null)
const generated = ref<{ roundId: string; name: string }[]>([])
const { mutate: generateMut } = useMutation(GENERATE_FFL_HOME_AND_AWAY)
async function generate() {
  const res = await generateMut({
    input: {
      seasonId: built.value!.seasonId,
      clubSeasonIds: built.value!.clubSeasons.map((c) => c.clubSeasonId),
      rounds: genRounds.value,
      aflRoundStartId: String(genAflStart.value),
    },
  })
  generated.value = res?.data?.generateFFLHomeAndAway ?? []
}

const roundName = ref('')
const aflRoundId = ref<number | null>(null)
const roundType = ref('MINOR')
const rounds = ref<{ roundId: string; name: string }[]>([])

const { mutate: addRoundMut } = useMutation(ADD_FFL_ROUND)
async function addRound() {
  const res = await addRoundMut({
    input: { seasonId: built.value!.seasonId, name: roundName.value, aflRoundId: String(aflRoundId.value), roundType: roundType.value },
  })
  const r = res?.data?.addFFLRound
  if (r) rounds.value.push(r)
  roundName.value = ''
}

const fixtureHome = reactive<Record<string, string>>({})
const fixtureAway = reactive<Record<string, string>>({})
const fixturesByRound = reactive<Record<string, { matchId: string; home: string; away: string }[]>>({})

const { mutate: addFixtureMut } = useMutation(ADD_FFL_FIXTURE)
async function addFixture(roundId: string) {
  const home = fixtureHome[roundId]
  const away = fixtureAway[roundId]
  const res = await addFixtureMut({ input: { roundId, homeClubSeasonId: home, awayClubSeasonId: away } })
  const f = res?.data?.addFFLFixture
  if (f) {
    ;(fixturesByRound[roundId] ||= []).push({ matchId: f.matchId, home, away })
  }
}

function clubName(clubSeasonId: string): string {
  return built.value?.clubSeasons.find((c) => c.clubSeasonId === clubSeasonId)?.clubName ?? clubSeasonId
}
</script>
