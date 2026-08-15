<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading…</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <NotFound v-else-if="notFound" entity="Club" />
    <template v-else-if="club">
      <Breadcrumb :items="[{ label: 'FFL' }]" />
      <div class="flex items-center gap-3 mb-6">
        <img :src="clubLogoUrl(club.name)" :alt="club.name" class="w-10 h-10 object-contain" />
        <h1 class="text-2xl font-bold">{{ club.name }}</h1>
      </div>

      <section>
        <h2 class="text-lg font-semibold text-text-heading mb-3">Seasons</h2>
        <ul class="divide-y divide-border-subtle">
          <li v-for="entry in club.seasons" :key="entry.id">
            <router-link
              :to="{ name: 'ffl-club-season', params: { clubSeasonId: entry.id } }"
              class="flex items-center justify-between py-3 text-sm hover:text-active transition-colors"
            >
              <span class="font-medium">{{ entry.season.name }}</span>
              <span class="text-text-muted tabular-nums">{{ entry.won }}-{{ entry.lost }}-{{ entry.drawn }}</span>
            </router-link>
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useQuery } from '@vue/apollo-composable'
import { GET_FFL_CLUB } from '../api/queries'
import { useNotFound } from '@/composables/useNotFound'
import { clubLogoUrl } from '../utils/clubLogos'
import NotFound from '@/components/NotFound.vue'
import Breadcrumb from '../components/Breadcrumb.vue'

const props = defineProps<{ clubId: string }>()

const { result, loading, error } = useQuery(
  GET_FFL_CLUB,
  () => ({ id: props.clubId }),
)

const club = computed(() => result.value?.fflClub ?? null)
const notFound = useNotFound(club, loading, error)
</script>
