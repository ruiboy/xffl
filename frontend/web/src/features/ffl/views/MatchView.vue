<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading match…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="match">
      <div class="mb-6">
        <Breadcrumb v-if="round" :items="breadcrumbs" />

        <!-- Versus: A v B -->
        <h1 v-if="matchStyle === 'versus'" class="text-2xl font-bold flex items-center gap-3">
          <img v-if="clubMatches[0]" :src="clubLogoUrl(clubMatches[0].club.name)" :alt="clubMatches[0].club.name" class="w-10 h-10 object-contain" />
          <router-link v-if="clubMatches[0]" :to="{ name: 'ffl-club-season', params: { clubSeasonId: clubMatches[0].clubSeasonId } }" class="hover:text-text-muted transition-colors">{{ clubMatches[0].club.name }}</router-link>
          <span v-else>—</span>
          <span class="text-text-faint mx-1">v</span>
          <img v-if="clubMatches[1]" :src="clubLogoUrl(clubMatches[1].club.name)" :alt="clubMatches[1].club.name" class="w-10 h-10 object-contain" />
          <router-link v-if="clubMatches[1]" :to="{ name: 'ffl-club-season', params: { clubSeasonId: clubMatches[1].clubSeasonId } }" class="hover:text-text-muted transition-colors">{{ clubMatches[1].club.name }}</router-link>
          <span v-else>—</span>
        </h1>

        <!-- Bye: a single club, no opponent -->
        <h1 v-else-if="matchStyle === 'bye'" class="text-2xl font-bold flex items-center gap-3">
          <img v-if="clubMatches[0]" :src="clubLogoUrl(clubMatches[0].club.name)" :alt="clubMatches[0].club.name" class="w-10 h-10 object-contain" />
          <router-link v-if="clubMatches[0]" :to="{ name: 'ffl-club-season', params: { clubSeasonId: clubMatches[0].clubSeasonId } }" class="hover:text-text-muted transition-colors">{{ clubMatches[0].club.name }}</router-link>
          <span class="text-sm font-medium rounded-full bg-surface-raised px-2.5 py-1 text-text-muted">Bye</span>
        </h1>

        <!-- Superbye: every club submits -->
        <h1 v-else class="text-2xl font-bold flex items-center gap-3">
          Superbye
          <span class="text-sm font-medium rounded-full bg-surface-raised px-2.5 py-1 text-text-muted">{{ clubMatches.length }} teams · top scorer +1</span>
        </h1>

        <p v-if="match.venue" class="text-sm text-text-muted mt-1">{{ match.venue }}</p>
        <p v-if="matchStyle === 'versus' && match.result" class="text-lg font-semibold mt-2">
          {{ clubMatches[0]?.score }} – {{ clubMatches[1]?.score }}
        </p>
      </div>

      <!-- Versus & bye: headers on one row so the two tables below start aligned. -->
      <template v-if="matchStyle !== 'superbye'">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-x-8 mb-0">
          <MatchSideHeader
            v-for="side in sides" :key="side.label + '-hd'"
            :side="side" :selected-club-id="selectedClubId" :match-style="matchStyle" :top-score="topScore"
          />
        </div>
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-x-8">
          <div v-for="side in sides" :key="side.label + '-tbl'">
            <SquadTable v-if="side.clubMatch" :player-matches="side.clubMatch.playerMatches" :highlight-pm-id="highlightPmId" />
          </div>
        </div>
      </template>

      <!-- Superbye: one self-contained card per team (header above its own table). -->
      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-x-8 gap-y-6">
        <div v-for="side in sides" :key="side.clubMatch.id">
          <MatchSideHeader :side="side" :selected-club-id="selectedClubId" :match-style="matchStyle" :top-score="topScore" />
          <SquadTable :player-matches="side.clubMatch.playerMatches" :highlight-pm-id="highlightPmId" />
        </div>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_MATCH } from '../api/queries'
import Breadcrumb from '../components/Breadcrumb.vue'
import SquadTable from '../components/SquadTable.vue'
import MatchSideHeader from '../components/MatchSideHeader.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { useFflState } from '../composables/useFflState'

const props = defineProps<{ matchId: string }>()

const route = useRoute()
const { selectedClubId } = useFflState()
const { result, loading, error } = useQuery(GET_FFL_MATCH, () => ({ id: props.matchId }))

const match = computed(() => result.value?.fflMatch ?? null)
const round = computed(() => match.value?.round ?? null)
const matchStyle = computed<string>(() => match.value?.matchStyle ?? 'versus')
const clubMatches = computed<any[]>(() => match.value?.clubMatches ?? [])
const topScore = computed(() => Math.max(0, ...clubMatches.value.map((cm: any) => cm.score ?? 0)))

const highlightPmId = computed(() => (route.query.highlight as string) || null)

let scrolledToHighlight = false
watch(match, (m) => {
  if (!m || !highlightPmId.value || scrolledToHighlight) return
  scrolledToHighlight = true
  const id = highlightPmId.value
  setTimeout(() => {
    const el = document.getElementById(`pm-${id}`)
    if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, 300)
}, { immediate: true })

const breadcrumbs = computed(() => {
  if (!match.value || !round.value) return []
  return [
    { label: 'FFL' },
    { label: round.value.season.name, to: { name: 'home' } },
    { label: round.value.name, to: { name: 'ffl-round', params: { roundId: round.value.id } } },
  ]
})

// One column per participating club_match — two for versus, one for a bye, N for
// a superbye. clubMatches orders home before away, so versus reads [home, away].
const sides = computed(() =>
  clubMatches.value.map((cm: any) => ({ label: cm.club.name, clubMatch: cm })),
)
</script>
