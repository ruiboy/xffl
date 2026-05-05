<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
    @click.self="$emit('close')"
  >
    <div class="bg-surface rounded-xl border border-border w-full max-w-lg mx-4 p-6">
      <div class="flex items-start justify-between mb-4">
        <div>
          <h2 class="text-base font-semibold text-text">Resolve player</h2>
          <p class="text-xs text-text-muted mt-0.5">{{ player.parsedName }}<span v-if="player.clubName" class="ml-2 text-text-faint">· {{ player.clubName }}</span></p>
        </div>
        <button @click="$emit('close')" class="text-text-faint hover:text-text text-lg leading-none">×</button>
      </div>

      <!-- Search mode -->
      <div v-if="!addingNew" class="space-y-3">
        <input
          v-model="searchQuery"
          @input="onSearchInput"
          placeholder="Search players…"
          autofocus
          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
        />

        <div class="border border-border rounded-lg overflow-hidden h-48 overflow-y-auto">
          <div v-if="searching" class="flex items-center justify-center h-full text-xs text-text-faint">Searching…</div>
          <div v-else-if="searchQuery.length >= 2 && results.length === 0" class="flex items-center justify-center h-full text-xs text-text-faint">No players found.</div>
          <div v-else-if="!results.length" class="flex items-center justify-center h-full text-xs text-text-faint">Type to search…</div>
          <div
            v-for="p in results"
            :key="p.id"
            class="flex items-center justify-between px-3 py-2 border-b border-border last:border-0 hover:bg-surface-raised"
          >
            <div class="min-w-0 mr-3">
              <span class="text-sm text-text font-medium">{{ p.name }}</span>
              <template v-if="p.latestPlayerSeason">
                <span class="text-xs text-text-faint ml-2 whitespace-nowrap">
                  {{ p.latestPlayerSeason.clubSeason.club.name }} · {{ p.latestPlayerSeason.clubSeason.season.name }}
                </span>
              </template>
              <span v-else class="text-xs text-text-faint ml-2">no season on record</span>
            </div>
            <button
              @click="selectPlayer(p)"
              :disabled="resolving"
              class="shrink-0 rounded border border-border px-2 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
            >Select</button>
          </div>
        </div>

        <div class="pt-1 border-t border-border">
          <button
            @click="addingNew = true"
            class="text-xs text-active hover:underline"
          >+ Add new player</button>
        </div>

        <p v-if="error" class="text-xs text-red-400">{{ error }}</p>
      </div>

      <!-- Add new player mode -->
      <div v-else class="space-y-3">
        <p class="text-xs text-text-muted">
          Adding as a new AFL player to the same club as this match.
        </p>
        <input
          v-model="newPlayerName"
          placeholder="Player name…"
          autofocus
          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
        />
        <div class="flex gap-2">
          <button
            @click="addingNew = false"
            class="rounded-lg border border-border px-3 py-2 text-sm text-text hover:bg-surface-hover transition-colors"
          >← Back</button>
          <button
            @click="addAndResolve"
            :disabled="!newPlayerName.trim() || resolving"
            class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ resolving ? 'Adding…' : 'Add & Resolve' }}</button>
        </div>
        <p v-if="error" class="text-xs text-red-400">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useApolloClient } from '@vue/apollo-composable'
import { SEARCH_AFL_PLAYERS } from '../api/queries'
import { RESOLVE_AFL_PLAYER_MATCH, ADD_AFL_PLAYER, ADD_AFL_PLAYER_SEASON } from '../api/mutations'

type UnmatchedPlayer = {
  parsedName: string
  clubMatchId: string
  clubSeasonId: string
  clubName: string
  kicks: number
  handballs: number
  marks: number
  hitouts: number
  tackles: number
  goals: number
  behinds: number
}

const props = defineProps<{
  show: boolean
  player: UnmatchedPlayer
}>()

const emit = defineEmits<{
  close: []
  resolved: []
}>()

const { resolveClient } = useApolloClient()

const searchQuery = ref('')
const results = ref<any[]>([])
const searching = ref(false)
const resolving = ref(false)
const error = ref('')
const addingNew = ref(false)
const newPlayerName = ref('')

let searchTimeout: ReturnType<typeof setTimeout> | null = null

watch(() => props.show, (v) => {
  if (v) {
    searchQuery.value = ''
    results.value = []
    error.value = ''
    addingNew.value = false
    newPlayerName.value = ''
  }
})

function onSearchInput() {
  error.value = ''
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
  } catch (e: any) {
    error.value = e.message ?? 'Search failed'
  } finally {
    searching.value = false
  }
}

async function selectPlayer(p: any) {
  resolving.value = true
  error.value = ''
  try {
    const client = resolveClient()
    let playerSeasonId: string

    const isCurrentSeason = p.latestPlayerSeason?.clubSeason?.id === props.player.clubSeasonId
    if (isCurrentSeason) {
      playerSeasonId = p.latestPlayerSeason.id
    } else {
      // Flow B/C-variant: player exists but not registered for this club/season yet
      const addRes = await client.mutate({
        mutation: ADD_AFL_PLAYER_SEASON,
        variables: { input: { playerId: p.id, clubSeasonId: props.player.clubSeasonId } },
      })
      playerSeasonId = addRes?.data?.addAFLPlayerSeason?.id
      if (!playerSeasonId) throw new Error('Failed to register player for this season')
    }

    await client.mutate({
      mutation: RESOLVE_AFL_PLAYER_MATCH,
      variables: {
        input: {
          clubMatchId: props.player.clubMatchId,
          playerSeasonId,
          kicks: props.player.kicks,
          handballs: props.player.handballs,
          marks: props.player.marks,
          hitouts: props.player.hitouts,
          tackles: props.player.tackles,
          goals: props.player.goals,
          behinds: props.player.behinds,
          parsedName: props.player.parsedName,
        },
      },
    })
    emit('resolved')
  } catch (e: any) {
    error.value = e.message ?? 'Failed to resolve player'
  } finally {
    resolving.value = false
  }
}

async function addAndResolve() {
  const name = newPlayerName.value.trim()
  if (!name) return
  resolving.value = true
  error.value = ''
  try {
    const client = resolveClient()
    const addRes = await client.mutate({
      mutation: ADD_AFL_PLAYER,
      variables: { input: { name, clubSeasonId: props.player.clubSeasonId } },
    })
    const playerSeasonId = addRes?.data?.addAFLPlayer?.id
    if (!playerSeasonId) throw new Error('No player season returned')

    await client.mutate({
      mutation: RESOLVE_AFL_PLAYER_MATCH,
      variables: {
        input: {
          clubMatchId: props.player.clubMatchId,
          playerSeasonId,
          kicks: props.player.kicks,
          handballs: props.player.handballs,
          marks: props.player.marks,
          hitouts: props.player.hitouts,
          tackles: props.player.tackles,
          goals: props.player.goals,
          behinds: props.player.behinds,
          parsedName: props.player.parsedName,
        },
      },
    })
    emit('resolved')
  } catch (e: any) {
    error.value = e.message ?? 'Failed to add player'
  } finally {
    resolving.value = false
  }
}
</script>
