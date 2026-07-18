<template>
  <!-- Create an FFL season: name + scoring era + the clubs playing it. Framed by its dialog. -->
  <div class="space-y-3">
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_AFL_SEASONS } from '@/features/afl/api/queries'
import { GET_FFL_BUILDER_REFS } from '../api/queries'
import { BUILD_FFL_SEASON } from '../api/mutations'

const emit = defineEmits<{ (e: 'created', seasonId: string): void }>()

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
  if (built.value) emit('created', built.value.seasonId)
}
</script>
