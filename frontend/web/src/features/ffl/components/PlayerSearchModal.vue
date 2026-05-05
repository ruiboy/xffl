<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      @click.self="$emit('close')"
    >
      <div class="relative z-10 w-full max-w-lg mx-4 rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">

        <!-- Search step -->
        <template v-if="step === 'search'">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-text">Add Player</h3>
            <button @click="$emit('close')" class="text-text-faint hover:text-text text-lg leading-none">×</button>
          </div>

          <input
            v-model="searchQuery"
            @input="onSearchInput"
            type="text"
            placeholder="Search by name…"
            autofocus
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none mb-3"
          />

          <div class="min-h-[8rem] max-h-64 overflow-y-auto -mx-1 px-1 mb-3">
            <div v-if="searching" class="text-text-faint text-sm py-2">Searching…</div>
            <div v-else-if="searchQuery.length >= 2 && results.length === 0" class="text-text-faint text-sm py-2">No players found.</div>
            <div v-else>
              <div
                v-for="p in results"
                :key="p.id"
                class="flex items-center justify-between border-b border-border-subtle py-2"
              >
                <div class="min-w-0 mr-3">
                  <div class="text-sm text-text font-medium">{{ p.name }}</div>
                  <div v-if="p.latestPlayerSeason" class="text-xs text-text-muted">
                    {{ p.latestPlayerSeason.clubSeason.club.name }} · {{ p.latestPlayerSeason.clubSeason.season.name }}
                  </div>
                </div>
                <button
                  @click="selectPlayer(p)"
                  class="shrink-0 rounded border border-active px-2 py-0.5 text-xs font-medium text-active hover:bg-active hover:text-active-text transition-colors"
                >Add</button>
              </div>
            </div>
          </div>

          <div class="flex items-center justify-between border-t border-border pt-3">
            <button
              @click="step = 'add-new'"
              class="text-xs text-active hover:underline"
            >+ Add new player</button>
            <button
              @click="$emit('close')"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >Cancel</button>
          </div>
        </template>

        <!-- Confirm add step -->
        <template v-else-if="step === 'confirm-add' && pendingPlayer">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-text">Add Player</h3>
            <button @click="$emit('close')" class="text-text-faint hover:text-text text-lg leading-none">×</button>
          </div>

          <p class="text-sm font-medium text-text mb-0.5">{{ pendingPlayer.name }}</p>
          <p v-if="pendingPlayer.latestPlayerSeason" class="text-xs text-text-muted mb-4">
            {{ pendingPlayer.latestPlayerSeason.clubSeason.club.name }}
          </p>

          <label class="text-xs text-text-muted block mb-1">From round</label>
          <select
            v-model="fromRoundId"
            class="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text mb-5 focus:outline-none focus:border-active"
          >
            <option v-for="r in rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>

          <p v-if="addError" class="mb-2 text-sm text-red-400">{{ addError }}</p>

          <div class="flex gap-2 justify-end">
            <button
              @click="step = 'search'"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >← Back</button>
            <button
              @click="confirmAdd"
              :disabled="adding"
              class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >{{ adding ? 'Adding…' : 'Add Player' }}</button>
          </div>
        </template>

        <!-- Add new player step -->
        <template v-else-if="step === 'add-new'">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-base font-semibold text-text">Add New Player</h3>
            <button @click="$emit('close')" class="text-text-faint hover:text-text text-lg leading-none">×</button>
          </div>

          <div class="space-y-3 mb-4">
            <div>
              <label class="text-xs text-text-muted block mb-1">Player name</label>
              <input
                v-model="newPlayerName"
                placeholder="e.g. Toby Greene"
                autofocus
                class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:border-active"
              />
            </div>
            <div>
              <label class="text-xs text-text-muted block mb-1">AFL club</label>
              <div v-if="loadingClubSeasons" class="text-xs text-text-faint py-1">Loading…</div>
              <select
                v-else
                v-model="newPlayerClubSeasonId"
                class="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:outline-none focus:border-active"
              >
                <option value="">Select club…</option>
                <option
                  v-for="cs in aflClubSeasons"
                  :key="cs.id"
                  :value="cs.id"
                >{{ cs.club.name }}</option>
              </select>
            </div>
            <div>
              <label class="text-xs text-text-muted block mb-1">From round</label>
              <select
                v-model="fromRoundId"
                class="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:outline-none focus:border-active"
              >
                <option v-for="r in rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
              </select>
            </div>
          </div>

          <p v-if="addError" class="mb-2 text-sm text-red-400">{{ addError }}</p>

          <div class="flex gap-2 justify-end">
            <button
              @click="step = 'search'"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >← Back</button>
            <button
              @click="confirmAddNew"
              :disabled="!newPlayerName.trim() || !newPlayerClubSeasonId || adding"
              class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >{{ adding ? 'Adding…' : 'Add Player' }}</button>
          </div>
        </template>

      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useQuery, useApolloClient } from '@vue/apollo-composable'
import { GET_AFL_SEASON_CLUB_SEASONS, SEARCH_AFL_PLAYERS } from '../api/queries'
import { ADD_AFL_PLAYER, ADD_FFL_PLAYER_TO_SEASON } from '../api/mutations'

const props = defineProps<{
  show: boolean
  fflSeasonId: string
  fflClubSeasonId: string
  rounds: { id: string; name: string }[]
  defaultFromRoundId: string
}>()

const emit = defineEmits<{
  close: []
  added: []
}>()

const { resolveClient } = useApolloClient()

type PlayerResult = {
  id: string
  name: string
  latestPlayerSeason: {
    id: string
    clubSeason: { id: string; club: { name: string }; season: { name: string } }
  } | null
}

const step = ref<'search' | 'confirm-add' | 'add-new'>('search')
const searchQuery = ref('')
const results = ref<PlayerResult[]>([])
const searching = ref(false)
const pendingPlayer = ref<PlayerResult | null>(null)
const fromRoundId = ref('')
const newPlayerName = ref('')
const newPlayerClubSeasonId = ref('')
const adding = ref(false)
const addError = ref('')

let searchTimeout: ReturnType<typeof setTimeout> | null = null

const { result: clubSeasonsResult, loading: loadingClubSeasons } = useQuery(
  GET_AFL_SEASON_CLUB_SEASONS,
  () => ({ fflSeasonId: props.fflSeasonId }),
  () => ({ enabled: step.value === 'add-new' }),
)

const aflClubSeasons = ref<{ id: string; club: { name: string } }[]>([])
watch(clubSeasonsResult, (v) => {
  aflClubSeasons.value = v?.fflSeason?.aflSeason?.ladder ?? []
})

watch(() => props.show, (v) => {
  if (v) {
    step.value = 'search'
    searchQuery.value = ''
    results.value = []
    addError.value = ''
    pendingPlayer.value = null
    newPlayerName.value = ''
    newPlayerClubSeasonId.value = ''
    fromRoundId.value = props.defaultFromRoundId
  }
})

watch(() => props.defaultFromRoundId, (v) => {
  if (!fromRoundId.value) fromRoundId.value = v
})

function onSearchInput() {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (searchQuery.value.length < 2) {
    results.value = []
    return
  }
  searchTimeout = setTimeout(runSearch, 300)
}

async function runSearch() {
  searching.value = true
  try {
    const client = resolveClient()
    const res = await client.query({
      query: SEARCH_AFL_PLAYERS,
      variables: { query: searchQuery.value },
      fetchPolicy: 'network-only',
    })
    results.value = res.data?.aflPlayerSearch ?? []
  } catch {
    results.value = []
  } finally {
    searching.value = false
  }
}

function selectPlayer(p: PlayerResult) {
  pendingPlayer.value = p
  step.value = 'confirm-add'
}

async function confirmAdd() {
  if (!pendingPlayer.value) return
  adding.value = true
  addError.value = ''
  try {
    const client = resolveClient()
    await client.mutate({
      mutation: ADD_FFL_PLAYER_TO_SEASON,
      variables: {
        input: {
          clubSeasonId: props.fflClubSeasonId,
          aflPlayerSeasonId: pendingPlayer.value.latestPlayerSeason?.id ?? pendingPlayer.value.id,
          fromRoundId: fromRoundId.value || null,
        },
      },
    })
    emit('added')
  } catch (e: any) {
    addError.value = e.message ?? 'Failed to add player'
  } finally {
    adding.value = false
  }
}

async function confirmAddNew() {
  const name = newPlayerName.value.trim()
  if (!name || !newPlayerClubSeasonId.value) return
  adding.value = true
  addError.value = ''
  try {
    const client = resolveClient()
    const addRes = await client.mutate({
      mutation: ADD_AFL_PLAYER,
      variables: { input: { name, clubSeasonId: newPlayerClubSeasonId.value } },
    })
    const aflPlayerSeasonId = addRes?.data?.addAFLPlayer?.id
    if (!aflPlayerSeasonId) throw new Error('No player season returned')

    await client.mutate({
      mutation: ADD_FFL_PLAYER_TO_SEASON,
      variables: {
        input: {
          clubSeasonId: props.fflClubSeasonId,
          aflPlayerSeasonId,
          fromRoundId: fromRoundId.value || null,
        },
      },
    })
    emit('added')
  } catch (e: any) {
    addError.value = e.message ?? 'Failed to add player'
  } finally {
    adding.value = false
  }
}
</script>
