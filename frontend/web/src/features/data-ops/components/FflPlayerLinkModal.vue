<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60"
      @click.self="$emit('close')"
    >
      <div class="relative z-10 w-full max-w-lg mx-4 rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">
        <div class="flex items-start justify-between mb-4">
          <div>
            <h3 class="text-base font-semibold text-text">Link Player</h3>
            <p class="text-xs text-text-muted mt-0.5">{{ parsedName }}<span v-if="clubHint" class="ml-2 text-text-faint">· {{ clubHint }}</span></p>
          </div>
          <button @click="$emit('close')" class="text-text-faint hover:text-text text-lg leading-none">×</button>
        </div>

        <!-- Search mode -->
        <div v-if="!addingNew">
          <input
            v-model="searchQuery"
            @input="onSearchInput"
            placeholder="Search players…"
            autofocus
            class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none mb-3"
          />

          <div class="h-52 overflow-y-auto -mx-1 px-1 mb-3">
            <div v-if="searching" class="text-text-faint text-sm py-2">Searching…</div>
            <div v-else-if="searchQuery.length >= 2 && results.length === 0" class="text-text-faint text-sm py-2">No players found.</div>
            <div v-else-if="!results.length" class="text-text-faint text-sm py-2">Type to search…</div>
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
                  <div v-else class="text-xs text-text-faint">no season on record</div>
                </div>
                <button
                  @click="selectPlayer(p)"
                  :disabled="linking"
                  class="shrink-0 rounded border border-active px-2 py-0.5 text-xs font-medium text-active hover:bg-active hover:text-active-text transition-colors disabled:opacity-40"
                >Link</button>
              </div>
            </div>
          </div>

          <p v-if="error" class="mb-3 text-sm text-red-400">{{ error }}</p>

          <div class="flex items-center justify-between border-t border-border pt-3">
            <button
              @click="addingNew = true"
              class="text-xs text-active hover:underline"
            >+ Add new player</button>
            <button
              @click="$emit('close')"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >Cancel</button>
          </div>
        </div>

        <!-- Add new player mode -->
        <div v-else class="space-y-3">
          <div>
            <label class="text-xs text-text-muted block mb-1">Player name</label>
            <input
              v-model="newPlayerName"
              placeholder="e.g. Toby Greene"
              autofocus
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none"
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
              <option v-for="cs in aflClubSeasons" :key="cs.id" :value="cs.id">{{ cs.club.name }}</option>
            </select>
          </div>

          <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

          <div class="flex gap-2 justify-end">
            <button
              @click="addingNew = false"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >← Back</button>
            <button
              @click="addAndLink"
              :disabled="!newPlayerName.trim() || !newPlayerClubSeasonId || linking"
              class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >{{ linking ? 'Adding…' : 'Add & Link' }}</button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useApolloClient } from '@vue/apollo-composable'
import { SEARCH_AFL_PLAYERS, GET_AFL_SEASON_CLUB_SEASONS } from '../api/queries'
import { ADD_AFL_PLAYER, ADD_FFL_PLAYER_TO_SEASON } from '../api/mutations'

const props = defineProps<{
  show: boolean
  fflSeasonId: string
  fflClubSeasonId: string
  parsedName: string
  clubHint: string
}>()

const emit = defineEmits<{
  close: []
  linked: [{ playerSeasonId: string; resolvedName: string; resolvedClub: string }]
}>()

const { resolveClient } = useApolloClient()

const searchQuery = ref('')
const results = ref<any[]>([])
const searching = ref(false)
const linking = ref(false)
const error = ref('')
const addingNew = ref(false)
const newPlayerName = ref('')
const newPlayerClubSeasonId = ref('')

let searchTimeout: ReturnType<typeof setTimeout> | null = null

const { result: clubSeasonsResult, loading: loadingClubSeasons } = useQuery(
  GET_AFL_SEASON_CLUB_SEASONS,
  () => ({ fflSeasonId: props.fflSeasonId }),
  () => ({ enabled: addingNew.value && !!props.fflSeasonId }),
)

const aflClubSeasons = computed(() => clubSeasonsResult.value?.fflSeason?.aflSeason?.ladder ?? [])

watch(() => props.show, (v) => {
  if (v) {
    searchQuery.value = ''
    results.value = []
    error.value = ''
    addingNew.value = false
    newPlayerName.value = ''
    newPlayerClubSeasonId.value = ''
  }
})

function onSearchInput() {
  error.value = ''
  if (searchTimeout) clearTimeout(searchTimeout)
  if (searchQuery.value.length < 2) { results.value = []; return }
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

async function selectPlayer(p: any) {
  linking.value = true
  error.value = ''
  try {
    const client = resolveClient()
    const aflPlayerSeasonId = p.latestPlayerSeason?.id ?? p.id
    const res = await client.mutate({
      mutation: ADD_FFL_PLAYER_TO_SEASON,
      variables: { input: { clubSeasonId: props.fflClubSeasonId, aflPlayerSeasonId } },
    })
    const playerSeasonId = res?.data?.addFFLPlayerToSeason?.id
    if (!playerSeasonId) throw new Error('Failed to add player to season')
    emit('linked', {
      playerSeasonId,
      resolvedName: p.name,
      resolvedClub: p.latestPlayerSeason?.clubSeason?.club?.name ?? '',
    })
  } catch (e: any) {
    error.value = e.message ?? 'Failed to link player'
  } finally {
    linking.value = false
  }
}

async function addAndLink() {
  const name = newPlayerName.value.trim()
  if (!name || !newPlayerClubSeasonId.value) return
  linking.value = true
  error.value = ''
  try {
    const client = resolveClient()
    const addRes = await client.mutate({
      mutation: ADD_AFL_PLAYER,
      variables: { input: { name, clubSeasonId: newPlayerClubSeasonId.value } },
    })
    const aflPlayerSeasonId = addRes?.data?.addAFLPlayer?.id
    if (!aflPlayerSeasonId) throw new Error('No player season returned')

    const linkRes = await client.mutate({
      mutation: ADD_FFL_PLAYER_TO_SEASON,
      variables: { input: { clubSeasonId: props.fflClubSeasonId, aflPlayerSeasonId } },
    })
    const playerSeasonId = linkRes?.data?.addFFLPlayerToSeason?.id
    if (!playerSeasonId) throw new Error('Failed to add player to FFL season')

    const selectedClub = aflClubSeasons.value.find((cs: any) => cs.id === newPlayerClubSeasonId.value)
    emit('linked', {
      playerSeasonId,
      resolvedName: name,
      resolvedClub: selectedClub?.club?.name ?? '',
    })
  } catch (e: any) {
    error.value = e.message ?? 'Failed to add player'
  } finally {
    linking.value = false
  }
}
</script>
