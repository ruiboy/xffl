<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="season">
      <Breadcrumb :items="[{ label: 'FFL' }]" />
      <h1 class="text-2xl font-bold mb-6">{{ season.name }}</h1>

      <RoundNav
        class="mb-8"
        :rounds="season.rounds"
        :live-round-id="liveRoundId"
        :live-start-date="liveStartDate"
      />

      <section>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Ladder</h2>
        <LadderTable :ladder="season.ladder" />
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_SEASON_LADDER } from '../api/queries'
import { useFflState } from '../composables/useFflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import LadderTable from '../components/LadderTable.vue'
import RoundNav from '../components/RoundNav.vue'

const props = defineProps<{ seasonId: string }>()

const { liveRoundId, liveStartDate } = useFflState()

const { result, loading, error } = useQuery(
  GET_FFL_SEASON_LADDER,
  () => ({ id: props.seasonId }),
)

const season = computed(() => result.value?.fflSeason ?? null)
</script>
