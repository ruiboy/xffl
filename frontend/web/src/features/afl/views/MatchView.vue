<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-6">
        <div v-if="matchData" class="mb-2">
          <router-link
            :to="{ name: 'afl-round', params: { seasonId: props.seasonId, roundId: matchData.roundId } }"
            class="text-sm text-text-muted hover:text-text transition-colors"
          >
            ← Back to round
          </router-link>
        </div>
        <h1 class="text-2xl font-bold flex items-center gap-3">
          <img v-if="match.homeClubMatch" :src="clubLogoUrl(match.homeClubMatch.club.name)" :alt="match.homeClubMatch.club.name" class="w-10 h-10 object-contain" />
          {{ match.homeClubMatch?.club.name ?? '—' }}
          <span class="text-text-faint mx-1">v</span>
          <img v-if="match.awayClubMatch" :src="clubLogoUrl(match.awayClubMatch.club.name)" :alt="match.awayClubMatch.club.name" class="w-10 h-10 object-contain" />
          {{ match.awayClubMatch?.club.name ?? '—' }}
        </h1>
        <p v-if="match.venue" class="text-sm text-text-muted mt-1">{{ match.venue }}</p>
        <p v-if="match.result" class="text-lg font-semibold mt-2">
          {{ match.homeClubMatch?.score }} – {{ match.awayClubMatch?.score }}
        </p>
        <div class="mt-3 flex items-center gap-4">
          <button
            @click="managing = !managing"
            class="rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors"
            :class="managing
              ? 'border-active bg-active text-active-text'
              : 'border-border bg-surface text-text hover:bg-surface-hover'"
          >
            {{ managing ? 'Done' : 'Manage' }}
          </button>
          <span v-if="saveMessage" class="text-sm text-green-500">{{ saveMessage }}</span>
        </div>
      </div>

      <div v-for="side in sides" :key="side.label" class="mb-10">
        <h2 class="text-lg font-semibold mb-3">{{ side.label }}</h2>
        <PlayerStatsTable
          v-if="side.clubMatch"
          :club-match="side.clubMatch"
          :readonly="!managing"
          @update="handleUpdate"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_MATCH } from '../api/queries'
import { UPDATE_PLAYER_MATCH } from '../api/mutations'
import PlayerStatsTable from '../components/PlayerStatsTable.vue'
import { clubLogoUrl } from '../utils/clubLogos'

const props = defineProps<{ seasonId: string; matchId: string }>()

const managing = ref(false)

const { result, loading, error } = useQuery(GET_MATCH, () => ({ seasonId: props.seasonId }))

const matchData = computed(() => {
  const season = result.value?.aflSeason
  if (!season) return null
  for (const round of season.rounds) {
    const found = round.matches.find((m: { id: string }) => m.id === props.matchId)
    if (found) return { match: found, roundId: round.id as string }
  }
  return null
})

const match = computed(() => matchData.value?.match ?? null)

const sides = computed(() => {
  if (!match.value) return []
  return [
    { label: match.value.homeClubMatch?.club.name ?? 'Home', clubMatch: match.value.homeClubMatch },
    { label: match.value.awayClubMatch?.club.name ?? 'Away', clubMatch: match.value.awayClubMatch },
  ]
})

const { mutate } = useMutation(UPDATE_PLAYER_MATCH)

const saveMessage = ref('')
let saveMessageTimer: ReturnType<typeof setTimeout> | null = null

async function handleUpdate(input: { playerSeasonId: string; clubMatchId: string; [key: string]: unknown }) {
  await mutate({ input })
  if (saveMessageTimer) clearTimeout(saveMessageTimer)
  saveMessage.value = 'Saved'
  saveMessageTimer = setTimeout(() => { saveMessage.value = '' }, 3000)
}
</script>
