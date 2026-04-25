<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-6">
        <Breadcrumb v-if="round" :items="breadcrumbs" />
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
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <div v-for="side in sides" :key="side.label">
          <div class="flex items-center gap-2 mb-1">
            <img v-if="side.clubMatch" :src="clubLogoUrl(side.clubMatch.club.name)" :alt="side.clubMatch.club.name" class="w-8 h-8 object-contain" />
            <h2 class="text-lg font-semibold">
              <router-link
                v-if="side.clubMatch"
                :to="{ name: 'ffl-squad', params: { seasonId: props.seasonId, clubId: side.clubMatch.club.id } }"
                class="hover:text-active transition-colors"
              >{{ side.label }}</router-link>
              <span v-else>{{ side.label }}</span>
            </h2>
            <router-link
              v-if="isMyClubMatch && side.clubMatch?.club.id === selectedClubId"
              :to="{ name: 'ffl-team-builder', params: { seasonId: props.seasonId, roundId: round!.id } }"
              title="Team Builder"
              class="rounded p-1 text-active hover:bg-active/10 transition-colors"
            >
              <IconTeamBuilder class="w-4 h-4" />
            </router-link>
          </div>
          <p class="text-sm text-text-muted mb-3">
            Score: <span class="font-semibold text-text">{{ side.clubMatch?.score ?? 0 }}</span>
          </p>
          <SquadTable v-if="side.clubMatch" :player-matches="side.clubMatch.playerMatches" />
        </div>
      </div>

      <div v-if="aflRoundTo" class="mt-8">
        <router-link :to="aflRoundTo" class="text-sm text-text-muted hover:text-text transition-colors">
          AFL Round ↗
        </router-link>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_MATCH } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'
import SquadTable from '../components/SquadTable.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { useFflState } from '../composables/useFflState'
import IconTeamBuilder from '../components/icons/IconTeamBuilder.vue'
import { useAflState } from '../../afl/composables/useAflState'

const props = defineProps<{ seasonId: string; matchId: string }>()

const { selectedClubId } = useFflState()
const { liveSeasonId: aflSeasonId } = useAflState()
const { result, loading, error } = useQuery(GET_FFL_MATCH, () => ({ id: props.matchId }))

const match = computed(() => result.value?.fflMatch ?? null)
const round = computed(() => match.value?.round ?? null)

const breadcrumbs = computed(() => {
  if (!match.value || !round.value) return []
  return [
    { label: 'FFL' },
    { label: round.value.season.name, to: { name: 'home' } },
    { label: round.value.name, to: { name: 'ffl-round', params: { seasonId: props.seasonId, roundId: round.value.id } } },
  ]
})

const isMyClubMatch = computed(() => {
  if (!match.value || !selectedClubId.value) return false
  return match.value.homeClubMatch?.club.id === selectedClubId.value ||
    match.value.awayClubMatch?.club.id === selectedClubId.value
})

const aflRoundTo = computed(() => {
  const aflRoundId = round.value?.aflRoundId
  if (!aflRoundId || !aflSeasonId.value) return null
  return { name: 'afl-round', params: { seasonId: aflSeasonId.value, roundId: aflRoundId } }
})

const sides = computed(() => {
  if (!match.value) return []
  return [
    { label: match.value.homeClubMatch?.club.name ?? 'Home', clubMatch: match.value.homeClubMatch },
    { label: match.value.awayClubMatch?.club.name ?? 'Away', clubMatch: match.value.awayClubMatch },
  ]
})
</script>
