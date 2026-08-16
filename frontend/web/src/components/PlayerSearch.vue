<template>
  <Teleport to="body">
    <div
      v-if="show"
      class="fixed inset-0 z-50 flex items-start justify-center bg-black/60 pt-24"
      @click.self="$emit('close')"
    >
      <div class="relative z-10 w-full max-w-lg mx-4 rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-base font-semibold text-text">Find Player</h3>
          <button @click="$emit('close')" class="text-text-faint hover:text-text text-lg leading-none">×</button>
        </div>

        <input
          ref="input"
          v-model="searchQuery"
          @input="onSearchInput"
          type="text"
          placeholder="Search by name..."
          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none mb-3"
        />

        <div class="h-64 overflow-y-auto -mx-1 px-1">
          <div v-if="searching" class="text-text-faint text-sm py-2">Searching…</div>
          <div v-else-if="searchQuery.length >= 2 && results.length === 0" class="text-text-faint text-sm py-2">No players found.</div>
          <div v-else-if="searchQuery.length < 2" class="text-text-faint text-sm py-2">Type at least two characters.</div>
          <div v-else>
            <!-- The season is shown because a player's most recent record may
                 predate the current one; clicking lands on the season named here. -->
            <button
              v-for="p in results"
              :key="p.id"
              :disabled="!p.latestPlayerSeason"
              @click="goToPlayer(p)"
              class="w-full flex items-center justify-between border-b border-border-subtle py-2 text-left transition-colors disabled:cursor-default enabled:hover:bg-surface-hover"
            >
              <div class="min-w-0 mr-3">
                <div class="text-sm text-text font-medium">{{ p.name }}</div>
                <div v-if="p.latestPlayerSeason" class="text-xs text-text-muted">
                  {{ p.latestPlayerSeason.clubSeason.club.name }} · {{ p.latestPlayerSeason.clubSeason.season.name }}
                </div>
                <div v-else class="text-xs text-text-faint">No AFL season data</div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApolloClient } from '@vue/apollo-composable'
import { SEARCH_AFL_PLAYERS } from '@/features/ffl/api/queries'

const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: [] }>()

const router = useRouter()
const { resolveClient } = useApolloClient()

type PlayerResult = {
  id: string
  name: string
  latestPlayerSeason: {
    id: string
    clubSeason: { id: string; club: { name: string }; season: { id: string; name: string } }
  } | null
}

const searchQuery = ref('')
const results = ref<PlayerResult[]>([])
const searching = ref(false)
const input = ref<HTMLInputElement | null>(null)

let searchTimeout: ReturnType<typeof setTimeout> | null = null

// Opening resets the previous search and takes focus, so the modal is always
// ready to type into rather than showing a stale result list.
watch(() => props.show, (open) => {
  if (!open) return
  searchQuery.value = ''
  results.value = []
  nextTick(() => input.value?.focus())
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.show) emit('close')
}
onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))

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

// Navigation only — unlike the squad-building search, this never creates a
// player_season for players the AFL side hasn't recorded this year.
function goToPlayer(p: PlayerResult) {
  if (!p.latestPlayerSeason) return
  router.push({
    name: 'ffl-afl-player-season',
    params: { aflPlayerSeasonId: p.latestPlayerSeason.id },
  })
  emit('close')
}
</script>
