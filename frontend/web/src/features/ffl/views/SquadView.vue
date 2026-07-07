<template>
  <div>
    <Breadcrumb v-if="clubSeason" :items="breadcrumbs" />
    <div class="mb-6 flex items-center">
      <h1 class="text-2xl font-bold flex items-center gap-3">
        <img v-if="clubSeason" :src="clubLogoUrl(clubSeason.club.name)" :alt="clubSeason.club.name" class="w-10 h-10 object-contain" />
        {{ clubSeason?.club.name ?? '' }}
      </h1>
      <router-link
        v-if="isMyClub && liveClubMatchId"
        :to="{ name: 'ffl-club-match-edit', params: { clubMatchId: liveClubMatchId } }"
        class="ml-auto flex items-center gap-1.5 text-sm text-text-muted hover:text-text transition-colors"
      >
        <IconTeamBuilder class="w-4 h-4" />
        Team Builder
      </router-link>
    </div>

    <div v-if="squadLoading" class="text-text-faint">Loading...</div>
    <div v-else-if="squadError" class="text-red-400">{{ squadError.message }}</div>
    <template v-else>
      <div v-if="players.length > 0">
        <div class="mb-4 flex items-center gap-4">
          <template v-if="isMyClub">
            <button
              @click="managing = !managing"
              class="rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors"
              :class="managing
                ? 'border-active bg-active text-active-text'
                : 'border-border bg-surface text-text hover:bg-surface-hover'"
            >
              <span class="flex items-center gap-1.5">
                <IconManage v-if="!managing" class="w-3.5 h-3.5" />
                {{ managing ? 'Done' : 'Manage' }}
              </span>
            </button>
            <button
              v-if="managing"
              @click="openAddSearch"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >
              + Add Player
            </button>
            <span v-if="saveMessage" class="text-sm text-green-500">{{ saveMessage }}</span>
            <span class="w-px h-5 bg-border"></span>
          </template>
          <div class="flex">
            <button
              v-for="seg in segments"
              :key="seg.id"
              @click="statsView = seg.id"
              class="px-3 py-1 text-sm transition-colors border-b-2"
              :class="statsView === seg.id
                ? 'border-active text-text font-medium'
                : 'border-transparent text-text-muted hover:text-text hover:border-border'"
            >{{ seg.label }}</button>
          </div>
          <StatSourceToggle v-show="statsView === 'stats'" class="ml-auto" />
        </div>

        <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-border text-left text-text-muted">
              <th class="sticky left-0 z-20 bg-surface py-2 pr-3 font-medium w-36">
                <button
                  v-if="statsView === 'stats'"
                  class="transition-colors"
                  :class="statsSortKey === 'name' ? 'text-sky-400' : 'hover:text-text'"
                  title="Sort by name (click again for position order)"
                  @click="toggleStatsSort('name')"
                >Player</button>
                <template v-else>Player</template>
              </th>
              <th class="sticky left-36 z-20 bg-surface py-2 pr-3 font-medium whitespace-nowrap">
                <button
                  v-if="statsView === 'stats'"
                  class="transition-colors"
                  :class="statsSortKey === 'club' ? 'text-sky-400' : 'hover:text-text'"
                  title="Sort by club (click again for position order)"
                  @click="toggleStatsSort('club')"
                >Club</button>
                <template v-else>Club</template>
              </th>
              <!-- Squad headers -->
              <th v-show="statsView === 'squad'" class="py-2 pl-4 w-full">
                <div class="flex">
                  <span v-for="round in rounds" :key="round.id" class="flex-1 text-center text-[10px] text-text-faint font-normal">{{ roundLabel(round.name) }}</span>
                </div>
              </th>
              <th v-show="statsView === 'squad' && isMyClub && managing" class="py-2 px-2"></th>
              <!-- Stats headers — click to sort, click again for position order -->
              <template v-for="col in statCols" :key="col.key">
              <th v-show="statsView === 'stats'" class="py-2 px-2 font-medium text-right">
                <button
                  class="transition-colors"
                  :class="statsSortKey === col.key ? 'text-sky-400' : 'hover:text-text'"
                  :title="`Sort by ${POSITION_LABEL[col.key]} (click again for position order)`"
                  @click="toggleStatsSort(col.key)"
                >{{ POSITION_LABEL[col.key] }}</button>
              </th>
              </template>
              <th v-show="statsView === 'stats'" class="py-2 pl-2 pr-3 font-medium text-right">
                <button
                  class="transition-colors"
                  :class="statsSortKey === 'star' ? 'text-sky-400' : 'text-text-muted hover:text-text'"
                  title="Sort by Star score (click again for position order)"
                  @click="toggleStatsSort('star')"
                >Star <span class="text-yellow-400">★</span></button>
              </th>
            </tr>
          </thead>
          <tbody>
            <template v-for="(group, gi) in displayedGroups" :key="group.pos ?? 'bench'">
              <tr v-if="gi > 0"><td colspan="10" class="pt-3"></td></tr>
              <template v-for="row in group.players" :key="row.id">
                <tr
                  class="group hover:bg-surface-hover"
                  :class="[
                    statsView === 'squad' && managing ? 'cursor-pointer' : '',
                    expandedId === row.id ? '' : 'border-b border-border-subtle'
                  ]"
                  @click="statsView === 'squad' && managing && toggleRow(row)"
                >
                  <td class="sticky left-0 z-10 bg-surface group-hover:bg-surface-hover py-2 pr-3 font-medium w-36 truncate max-w-[144px]">
                    <PlayerStatsCard :name="row.player.aflPlayer.name" :club="row.aflPlayerSeason?.clubSeason?.club?.name ?? null" :afl-status="null" :afl-player-season-id="row.aflPlayerSeason?.id ?? null" :afl-round-id="null">
                      <router-link v-if="row.aflPlayerSeason?.id" :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: row.aflPlayerSeason.id } }" class="hover:text-text-muted transition-colors" @click.stop>{{ row.player.aflPlayer.name }}</router-link>
                      <span v-else>{{ row.player.aflPlayer.name }}</span>
                    </PlayerStatsCard>
                  </td>
                  <td class="sticky left-36 z-10 bg-surface group-hover:bg-surface-hover py-2 pr-3 text-xs text-text-muted whitespace-nowrap">
                    <router-link v-if="row.aflPlayerSeason?.clubSeason?.id" :to="{ name: 'ffl-afl-club-season', params: { clubSeasonId: row.aflPlayerSeason.clubSeason.id } }" class="inline-flex items-center gap-1 hover:text-text transition-colors" @click.stop>
                      <img :src="aflClubLogoUrl(row.aflPlayerSeason.clubSeason.club?.name ?? '')" class="w-4 h-4 object-contain shrink-0" />
                      {{ row.aflPlayerSeason.clubSeason.club?.name ?? '—' }}
                    </router-link>
                    <span v-else>{{ row.aflPlayerSeason?.clubSeason?.club?.name ?? '—' }}</span>
                  </td>
                  <!-- Squad cells -->
                  <td v-show="statsView === 'squad'" class="py-2 pl-4 w-full">
                    <div class="flex">
                      <span v-for="round in rounds" :key="round.id" class="flex-1 text-center text-xs font-mono" :class="roundColor(row.id, round.id)">{{ roundLetter(row.id, round.id) }}</span>
                    </div>
                  </td>
                  <td v-show="statsView === 'squad' && isMyClub && managing" class="py-2 px-2 text-right" @click.stop>
                    <button @click="openRemoveModal(row.id, row.player.aflPlayer.name, row.aflPlayerSeason?.clubSeason?.club?.name ?? '')" aria-label="Remove" class="text-red-400 hover:text-red-300 transition-colors">
                      <IconBin class="w-3.5 h-3.5" />
                    </button>
                  </td>
                  <!-- Stats cells -->
                  <template v-for="col in statCols" :key="col.key">
                  <td v-show="statsView === 'stats'" class="py-2 px-2 tabular-nums">
                    <StatCell
                      :avg="row.aflPlayerSeason?.statsAll?.[col.key]"
                      :last3="row.aflPlayerSeason?.statsLastN?.[col.key]"
                      :avg-style="statHeat(row.aflPlayerSeason?.statsAll?.[col.key], col.key)"
                      :last3-style="statHeat(row.aflPlayerSeason?.statsLastN?.[col.key], col.key)"
                    />
                  </td>
                  </template>
                  <td v-show="statsView === 'stats'" class="py-2 pl-2 pr-3 tabular-nums">
                    <StatCell
                      :avg="row.aflPlayerSeason?.statsAll ? starScore(row.aflPlayerSeason.statsAll) : null"
                      :last3="row.aflPlayerSeason?.statsLastN ? starScore(row.aflPlayerSeason.statsLastN) : null"
                      :avg-style="statHeat(row.aflPlayerSeason?.statsAll ? starScore(row.aflPlayerSeason.statsAll) : null, 'star')"
                      :last3-style="statHeat(row.aflPlayerSeason?.statsLastN ? starScore(row.aflPlayerSeason.statsLastN) : null, 'star')"
                    />
                  </td>
                </tr>
                <tr v-if="expandedId === row.id && statsView === 'squad'" class="border-b border-border-subtle">
                  <td colspan="4" class="pb-3 pt-1 px-0">
                    <div class="flex gap-4">
                      <div class="flex flex-col gap-2 text-xs shrink-0">
                        <div><div class="text-text-muted">From</div><div class="text-text">{{ roundName(row.fromRoundId) }}</div></div>
                        <div><div class="text-text-muted">To</div><div class="text-text">{{ roundName(row.toRoundId) }}</div></div>
                        <div><div class="text-text-muted">Cost</div><div class="text-text">{{ formatCost(row.costCents) }}</div></div>
                      </div>
                      <div class="flex-1 flex flex-col" @click.stop>
                        <textarea v-model="expandedNotes" rows="3" placeholder="Add notes..." class="flex-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none resize-none" />
                        <div class="flex justify-end mt-2">
                          <button v-if="expandedDirty" @click="saveExpanded" :disabled="expandedSubmitting" class="rounded-lg px-3 py-1.5 text-xs font-medium bg-active text-active-text hover:opacity-90 transition-colors disabled:opacity-40">{{ expandedSubmitting ? '…' : 'Save' }}</button>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </template>
            <template v-if="tradedPlayers.length > 0">
              <tr><td colspan="10" class="pt-4 pb-1">
                <button @click="showTraded = !showTraded" class="flex items-center gap-1.5 text-xs font-medium text-text-faint uppercase tracking-wide hover:text-text-muted transition-colors">
                  <span>{{ showTraded ? '▾' : '▸' }}</span>
                  Traded ({{ tradedPlayers.length }})
                </button>
              </td></tr>
              <template v-if="showTraded" v-for="row in tradedPlayers" :key="row.id">
                <tr
                  class="opacity-40 group hover:opacity-60"
                  :class="[
                    statsView === 'squad' && managing ? 'cursor-pointer' : '',
                    expandedId === row.id ? '' : 'border-b border-border-subtle'
                  ]"
                  @click="statsView === 'squad' && managing && toggleRow(row)"
                >
                  <td class="sticky left-0 z-10 bg-surface py-2 pr-3 font-medium w-36 truncate max-w-[144px]">
                    <PlayerStatsCard :name="row.player.aflPlayer.name" :club="row.aflPlayerSeason?.clubSeason?.club?.name ?? null" :afl-status="null" :afl-player-season-id="row.aflPlayerSeason?.id ?? null" :afl-round-id="null">
                      <router-link v-if="row.aflPlayerSeason?.id" :to="{ name: 'ffl-afl-player-season', params: { aflPlayerSeasonId: row.aflPlayerSeason.id } }" class="hover:text-text-muted transition-colors" @click.stop>{{ row.player.aflPlayer.name }}</router-link>
                      <span v-else>{{ row.player.aflPlayer.name }}</span>
                    </PlayerStatsCard>
                  </td>
                  <td class="sticky left-36 z-10 bg-surface py-2 pr-3 text-xs text-text-muted whitespace-nowrap">
                    <router-link v-if="row.aflPlayerSeason?.clubSeason?.id" :to="{ name: 'ffl-afl-club-season', params: { clubSeasonId: row.aflPlayerSeason.clubSeason.id } }" class="inline-flex items-center gap-1 hover:text-text transition-colors" @click.stop>
                      <img :src="aflClubLogoUrl(row.aflPlayerSeason.clubSeason.club?.name ?? '')" class="w-4 h-4 object-contain shrink-0" />
                      {{ row.aflPlayerSeason.clubSeason.club?.name ?? '—' }}
                    </router-link>
                    <span v-else>{{ row.aflPlayerSeason?.clubSeason?.club?.name ?? '—' }}</span>
                  </td>
                  <td v-show="statsView === 'squad'" class="py-2 pl-4 w-full">
                    <div class="flex">
                      <span v-for="round in rounds" :key="round.id" class="flex-1 text-center text-xs font-mono" :class="roundColor(row.id, round.id)">{{ roundLetter(row.id, round.id) }}</span>
                    </div>
                  </td>
                  <td v-show="statsView === 'squad' && isMyClub && managing" class="py-2 px-2"></td>
                  <template v-for="col in statCols" :key="col.key">
                  <td v-show="statsView === 'stats'" class="py-2 px-2 tabular-nums">
                    <StatCell
                      :avg="row.aflPlayerSeason?.statsAll?.[col.key]"
                      :last3="row.aflPlayerSeason?.statsLastN?.[col.key]"
                      :avg-style="statHeat(row.aflPlayerSeason?.statsAll?.[col.key], col.key)"
                      :last3-style="statHeat(row.aflPlayerSeason?.statsLastN?.[col.key], col.key)"
                    />
                  </td>
                  </template>
                  <td v-show="statsView === 'stats'" class="py-2 pl-2 pr-3 tabular-nums">
                    <StatCell
                      :avg="row.aflPlayerSeason?.statsAll ? starScore(row.aflPlayerSeason.statsAll) : null"
                      :last3="row.aflPlayerSeason?.statsLastN ? starScore(row.aflPlayerSeason.statsLastN) : null"
                      :avg-style="statHeat(row.aflPlayerSeason?.statsAll ? starScore(row.aflPlayerSeason.statsAll) : null, 'star')"
                      :last3-style="statHeat(row.aflPlayerSeason?.statsLastN ? starScore(row.aflPlayerSeason.statsLastN) : null, 'star')"
                    />
                  </td>
                </tr>
                <tr v-if="expandedId === row.id && statsView === 'squad'" class="border-b border-border-subtle opacity-40">
                  <td colspan="4" class="pb-3 pt-1 px-0">
                    <div class="flex gap-4">
                      <div class="flex flex-col gap-2 text-xs shrink-0">
                        <div><div class="text-text-muted">From</div><div class="text-text">{{ roundName(row.fromRoundId) }}</div></div>
                        <div><div class="text-text-muted">To</div><div class="text-text">{{ roundName(row.toRoundId) }}</div></div>
                        <div><div class="text-text-muted">Cost</div><div class="text-text">{{ formatCost(row.costCents) }}</div></div>
                      </div>
                      <div class="flex-1 flex flex-col" @click.stop>
                        <textarea v-model="expandedNotes" rows="3" placeholder="Add notes..." class="flex-1 w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text placeholder-text-faint focus:border-active focus:outline-none resize-none" />
                        <div class="flex justify-end mt-2">
                          <button v-if="expandedDirty" @click="saveExpanded" :disabled="expandedSubmitting" class="rounded-lg px-3 py-1.5 text-xs font-medium bg-active text-active-text hover:opacity-90 transition-colors disabled:opacity-40">{{ expandedSubmitting ? '…' : 'Save' }}</button>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </template>
          </tbody>
        </table>
        </div>
      </div>
      <p v-else class="text-text-faint">No players on squad.</p>
    </template>

    <PlayerSearchModal
      :show="addModalOpen"
      :ffl-season-id="clubSeason?.season.id ?? ''"
      :ffl-club-season-id="props.clubSeasonId"
      :rounds="rounds"
      :default-from-round-id="defaultRoundId()"
      @close="addModalOpen = false"
      @added="onPlayerAdded"
    />

    <Teleport to="body">
      <!-- Remove player modal -->
      <div v-if="modal" class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/60" @click="closeModal" />
        <div class="relative z-10 w-80 rounded-xl border border-border bg-surface-raised p-6 shadow-2xl">
          <h3 class="text-base font-semibold text-text mb-1">Remove Player</h3>
          <p class="text-sm font-medium text-text mb-0.5">{{ modal.playerName }}</p>
          <p class="text-xs text-text-muted mb-4">{{ modal.playerClub }}</p>
          <label class="text-xs text-text-muted block mb-1">Round</label>
          <select
            v-model="modal.roundId"
            class="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text mb-5 focus:outline-none focus:border-active"
          >
            <option v-for="r in rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
          </select>
          <div class="flex gap-2 justify-end">
            <button
              @click="closeModal"
              class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
            >Cancel</button>
            <button
              @click="confirmModal"
              :disabled="modalSubmitting || !modal.roundId"
              class="rounded-lg px-3 py-1.5 text-sm font-medium bg-red-600 text-white hover:bg-red-500 transition-colors disabled:opacity-40"
            >{{ modalSubmitting ? '…' : 'Remove' }}</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation, useApolloClient } from '@vue/apollo-composable'
import { useTheme } from '@/composables/useTheme'
import { heatStyle } from '@/utils/heatmap'
import { statCols, starScore, type StatSummary, type StatKey } from '../utils/playerStats'
import StatCell from '../components/StatCell.vue'
import PlayerStatsCard from '../components/PlayerStatsCard.vue'
import StatSourceToggle from '../components/StatSourceToggle.vue'
import { useStatSource } from '../composables/useStatSource'
import { GET_FFL_CLUB_SEASON, GET_FFL_SEASON_POSITIONS, GET_FFL_ROUND_CLUB_MATCHES } from '../api/queries'
import { REMOVE_FFL_PLAYER_FROM_SEASON, UPDATE_FFL_PLAYER_SEASON } from '../api/mutations'
import { useFflState } from '../composables/useFflState'
import Breadcrumb from '../components/Breadcrumb.vue'
import IconTeamBuilder from '../components/icons/IconTeamBuilder.vue'
import IconManage from '../components/icons/IconManage.vue'
import IconBin from '../components/icons/IconBin.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { clubLogoUrl as aflClubLogoUrl } from '@/features/afl/utils/clubLogos'
import { POSITION_LETTERS, POSITION_COLORS, POSITION_ORDER, POSITION_LABEL, primaryPosition, type RoundEntry } from '../utils/position'
import PlayerSearchModal from '../components/PlayerSearchModal.vue'

const props = defineProps<{ clubSeasonId: string }>()

const { selectedClubId, liveRoundId } = useFflState()
const { client } = useApolloClient()
const managing = ref(false)
const showTraded = ref(false)

type StatsView = 'squad' | 'stats'
const statsView = ref<StatsView>('squad')
const segments: { id: StatsView; label: string }[] = [
  { id: 'squad', label: 'Positions' },
  { id: 'stats', label: 'Stats' },
]

const { isDark } = useTheme()
const { statSource } = useStatSource()

const isMyClub = computed(() => !!selectedClubId.value && clubSeason.value?.club.id === selectedClubId.value)

// Squad query — driven by route clubSeasonId
const { result: squadResult, loading: squadLoading, error: squadError, refetch: refetchSquad } = useQuery(
  GET_FFL_CLUB_SEASON,
  () => ({ id: props.clubSeasonId }),
)

const clubSeason = computed(() => squadResult.value?.fflClubSeason ?? null)

// Live round club match — only fetched when viewing your own club
const { result: liveRoundMatchesResult } = useQuery(
  GET_FFL_ROUND_CLUB_MATCHES,
  () => ({ id: liveRoundId.value }),
  () => ({ enabled: isMyClub.value && !!liveRoundId.value }),
)

const liveClubMatchId = computed(() => {
  const matches = liveRoundMatchesResult.value?.fflRound?.matches ?? []
  for (const m of matches) {
    if (m.homeClubMatch.clubSeasonId === props.clubSeasonId) return m.homeClubMatch.id
    if (m.awayClubMatch.clubSeasonId === props.clubSeasonId) return m.awayClubMatch.id
  }
  return null
})

const breadcrumbs = computed(() => {
  if (!clubSeason.value) return []
  return [
    { label: 'FFL' },
    { label: clubSeason.value.season.name, to: { name: 'home' } },
    { label: clubSeason.value.club.name },
  ]
})

const byLastName = (a: { player: { aflPlayer: { name: string } } }, b: { player: { aflPlayer: { name: string } } }) => {
  const lastA = a.player.aflPlayer.name.split(' ').pop()?.toLowerCase() ?? ''
  const lastB = b.player.aflPlayer.name.split(' ').pop()?.toLowerCase() ?? ''
  return lastA.localeCompare(lastB)
}

const players = computed(() => {
  const nodes = clubSeason.value?.players?.nodes ?? []
  return [...nodes].sort(byLastName)
})

const activePlayers = computed(() => players.value.filter((p: PlayerSeasonRow) => !p.toRoundId))
const tradedPlayers = computed(() => [...players.value.filter((p: PlayerSeasonRow) => !!p.toRoundId)].sort(byLastName))

// --- Round history ---

const { result: seasonResult } = useQuery(
  GET_FFL_SEASON_POSITIONS,
  () => ({ id: clubSeason.value?.season.id ?? '' }),
  () => ({ enabled: !!clubSeason.value?.season.id }),
)

const rounds = computed(() => seasonResult.value?.fflSeason?.rounds ?? [])

const playerRoundMap = computed((): Map<string, Map<string, RoundEntry>> => {
  const map = new Map<string, Map<string, RoundEntry>>()
  for (const round of rounds.value) {
    for (const match of round.matches ?? []) {
      for (const side of [match.homeClubMatch, match.awayClubMatch]) {
        if (!side) continue
        for (const pm of side.playerMatches ?? []) {
          const isBench = pm.backupPositions != null || pm.interchangePosition != null
          if (!map.has(pm.playerSeasonId)) map.set(pm.playerSeasonId, new Map())
          map.get(pm.playerSeasonId)!.set(round.id, { position: pm.position ?? null, isBench })
        }
      }
    }
  }
  return map
})

function roundLetter(playerSeasonId: string, roundId: string): string {
  const e = playerRoundMap.value.get(playerSeasonId)?.get(roundId)
  if (!e) return '–'
  if (e.isBench) return 'B'
  return e.position ? (POSITION_LETTERS[e.position] ?? '?') : '–'
}

function roundColor(playerSeasonId: string, roundId: string): string {
  const e = playerRoundMap.value.get(playerSeasonId)?.get(roundId)
  if (!e) return 'text-text-faint'
  if (e.isBench) return 'text-text-muted'
  return e.position ? (POSITION_COLORS[e.position] ?? 'text-text') : 'text-text-faint'
}

function roundLabel(name: string): string {
  return name.replace(/\D+/g, '')
}

// --- Position grouping (recency-weighted) ---

const groupedPlayers = computed(() => {
  type P = typeof activePlayers.value[number]
  const buckets = new Map<string | null, P[]>(
    [...POSITION_ORDER.map(p => [p, []] as [string, P[]]), [null, []]]
  )
  for (const p of activePlayers.value) {
    const pos = primaryPosition(p.id, playerRoundMap.value, rounds.value)
    buckets.get(POSITION_ORDER.includes(pos as typeof POSITION_ORDER[number]) ? pos : null)!.push(p)
  }
  return [...POSITION_ORDER, null].flatMap(pos => {
    const group = buckets.get(pos) ?? []
    if (!group.length) return []
    return [{ pos, label: pos ? POSITION_LABEL[pos] : 'Bench / Unassigned', players: group }]
  })
})

// --- Stats view sorting ---

// Click a header to sort flat (stats descending, name/club ascending);
// click it again to return to the default position-group order.
type StatsSortKey = StatKey | 'star' | 'name' | 'club'
const statsSortKey = ref<StatsSortKey | null>(null)

function toggleStatsSort(key: StatsSortKey) {
  statsSortKey.value = statsSortKey.value === key ? null : key
}

const displayedGroups = computed(() => {
  const key = statsSortKey.value
  if (statsView.value !== 'stats' || !key) return groupedPlayers.value

  let sorted: PlayerSeasonRow[]
  if (key === 'name') {
    sorted = [...activePlayers.value].sort(byLastName)
  } else if (key === 'club') {
    const club = (r: PlayerSeasonRow) => r.aflPlayerSeason?.clubSeason?.club?.name ?? ''
    sorted = [...activePlayers.value].sort((a, b) => club(a).localeCompare(club(b)) || byLastName(a, b))
  } else {
    const val = (r: PlayerSeasonRow): number => {
      const ps = r.aflPlayerSeason
      const s = statSource.value === 'form'
        ? (ps?.statsLastN ?? ps?.statsAll)
        : (ps?.statsAll ?? ps?.statsLastN)
      if (!s) return -1
      return key === 'star' ? starScore(s) : s[key]
    }
    sorted = [...activePlayers.value].sort((a, b) => val(b) - val(a))
  }
  return [{ pos: null, label: '', players: sorted }]
})

// --- Manage mode: add/remove ---

// Saved flash
const saveMessage = ref('')
let saveMessageTimer: ReturnType<typeof setTimeout> | null = null
function flashSaved() {
  if (saveMessageTimer) clearTimeout(saveMessageTimer)
  saveMessage.value = 'Saved'
  saveMessageTimer = setTimeout(() => { saveMessage.value = '' }, 3000)
}

// Remove modal
interface RemoveModalState { playerSeasonId: string; playerName: string; playerClub: string; roundId: string }
const modal = ref<RemoveModalState | null>(null)
const modalSubmitting = ref(false)

const { mutate: removePlayerMutation } = useMutation(REMOVE_FFL_PLAYER_FROM_SEASON)
const { mutate: updatePlayerSeasonMutation } = useMutation(UPDATE_FFL_PLAYER_SEASON)

function defaultRoundId(): string {
  return liveRoundId.value || (rounds.value.at(-1)?.id ?? '')
}

function openRemoveModal(playerSeasonId: string, playerName: string, playerClub: string) {
  modal.value = { playerSeasonId, playerName, playerClub, roundId: defaultRoundId() }
}

function closeModal() {
  if (!modalSubmitting.value) modal.value = null
}

async function confirmModal() {
  if (!modal.value) return
  modalSubmitting.value = true
  try {
    await removePlayerMutation({ input: { id: modal.value.playerSeasonId, toRoundId: modal.value.roundId } })
    await refetchSquad()
    flashSaved()
    modal.value = null
  } finally {
    modalSubmitting.value = false
  }
}

// Add player modal
const addModalOpen = ref(false)

function openAddSearch() {
  addModalOpen.value = true
}

function cancelAddSearch() {
  addModalOpen.value = false
}

async function onPlayerAdded() {
  addModalOpen.value = false
  // The federation fields added to GET_FFL_CLUB_SEASON (~300ms) can cause a pre-mutation
  // in-flight request to still be running when the mutation completes. Apollo deduplicates
  // refetchSquad() against that stale in-flight and returns pre-mutation data.
  // Fix: client.query with network-only creates an independent request that bypasses
  // deduplication and waits past the stale in-flight. Once it resolves, there is no
  // competing in-flight so the subsequent refetchSquad() fires cleanly.
  await client.query({ query: GET_FFL_CLUB_SEASON, variables: { id: props.clubSeasonId }, fetchPolicy: 'network-only' })
  await refetchSquad()
  flashSaved()
}

// Inline row expansion

interface PlayerSeasonRow {
  id: string
  player: { id: string; aflPlayerId: string; aflPlayer: { id: string; name: string } }
  aflPlayerSeason?: {
    id: string
    clubSeason?: { id: string; club?: { name: string } } | null
    statsAll?: StatSummary | null
    statsLastN?: StatSummary | null
  } | null
  fromRoundId?: string | null
  toRoundId?: string | null
  notes?: string | null
  costCents?: number | null
}

// Heatmap ranges scaled to the active stat source (global Last N/Season toggle).
const columnRange = computed(() => {
  const allRows = [
    ...groupedPlayers.value.flatMap(g => g.players),
    ...tradedPlayers.value,
  ] as PlayerSeasonRow[]
  const src = (r: PlayerSeasonRow) =>
    statSource.value === 'form' ? r.aflPlayerSeason?.statsLastN : r.aflPlayerSeason?.statsAll
  const range = {} as Record<StatKey | 'star', { min: number; max: number }>
  for (const col of statCols) {
    const vals = allRows.map(r => src(r)?.[col.key]).filter((v): v is number => v != null)
    if (vals.length) range[col.key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  const starVals = allRows
    .map(r => { const s = src(r); return s ? starScore(s) : null })
    .filter((v): v is number => v != null)
  if (starVals.length) range['star'] = { min: Math.min(...starVals), max: Math.max(...starVals) }
  return range
})

function statHeat(value: number | null | undefined, key: StatKey | 'star'): Record<string, string> {
  if (value == null) return {}
  const r = columnRange.value[key]
  if (!r) return {}
  return heatStyle(value, r.min, r.max, isDark.value)
}

const expandedId = ref<string | null>(null)
const expandedNotes = ref('')
const expandedSubmitting = ref(false)
const expandedDirty = computed(() => {
  const row = players.value.find((p: PlayerSeasonRow) => p.id === expandedId.value)
  return row != null && expandedNotes.value !== (row.notes ?? '')
})

function roundName(id: string | null | undefined): string {
  if (!id) return '—'
  return rounds.value.find((r: { id: string; name: string }) => r.id === id)?.name ?? '—'
}

function formatCost(cents: number | null | undefined): string {
  if (cents == null) return '—'
  return `$${(cents / 100).toFixed(2)}`
}

function toggleRow(row: PlayerSeasonRow) {
  if (expandedId.value === row.id) {
    expandedId.value = null
    expandedNotes.value = ''
  } else {
    expandedId.value = row.id
    expandedNotes.value = row.notes ?? ''
  }
}

async function saveExpanded() {
  if (!expandedId.value) return
  expandedSubmitting.value = true
  try {
    await updatePlayerSeasonMutation({
      input: {
        id: expandedId.value,
        notes: expandedNotes.value || null,
      },
    })
    await refetchSquad()
    flashSaved()
  } finally {
    expandedSubmitting.value = false
  }
}

watch(managing, (val) => {
  if (!val) {
    cancelAddSearch()
    expandedId.value = null
    expandedNotes.value = ''
  }
})

watch(isMyClub, (val) => {
  if (!val) managing.value = false
})

watch(statsView, (val) => {
  if (val === 'stats') managing.value = false
})
</script>
