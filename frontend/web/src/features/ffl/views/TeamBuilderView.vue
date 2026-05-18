<template>
  <div>
    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <template v-else-if="round">
      <div class="mb-6">
        <Breadcrumb v-if="currentRound" :items="breadcrumbs" />
        <div class="flex items-center">
          <h1 class="text-2xl font-bold flex items-center gap-3">
            <img v-if="selectedClubSeason" :src="clubLogoUrl(selectedClubSeason.club.name)" :alt="selectedClubSeason.club.name" class="w-10 h-10 object-contain" />
            {{ selectedClubSeason?.club.name ?? '' }}<span class="font-normal text-text-muted"> · Team Builder</span>
          </h1>
          <div class="flex items-center gap-3 ml-auto">
            <div class="flex items-center gap-1 rounded-lg border border-border px-1">
              <router-link
                v-if="prevClubMatchId"
                :to="{ name: 'ffl-club-match-edit', params: { clubMatchId: prevClubMatchId } }"
                class="w-6 h-6 flex items-center justify-center rounded text-text-muted hover:bg-control-hover hover:text-text transition-colors text-sm"
                title="Previous round"
              >‹</router-link>
              <span v-else class="w-6 h-6 flex items-center justify-center text-text-faint text-sm opacity-30">‹</span>
              <span class="text-sm text-text-muted tabular-nums">{{ currentRound?.name }}</span>
              <router-link
                v-if="nextClubMatchId"
                :to="{ name: 'ffl-club-match-edit', params: { clubMatchId: nextClubMatchId } }"
                class="w-6 h-6 flex items-center justify-center rounded text-text-muted hover:bg-control-hover hover:text-text transition-colors text-sm"
                title="Next round"
              >›</router-link>
              <span v-else class="w-6 h-6 flex items-center justify-center text-text-faint text-sm opacity-30">›</span>
            </div>
            <router-link
              v-if="selectedClubSeason"
              :to="{ name: 'ffl-club-season', params: { clubSeasonId: bootstrapClubSeasonId } }"
              class="flex items-center gap-1.5 text-sm text-text-muted hover:text-text transition-colors"
            >
              <IconSquad class="w-4 h-4" />
              Squad
            </router-link>
          </div>
        </div>
      </div>

      <template v-if="selectedClubSeason && clubMatch">
        <div class="mb-6 flex items-center gap-4">
          <!-- Subs mode -->
          <template v-if="subsMode">
            <button
              @click="exitSubsMode"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
              :disabled="subsSaving"
            >Cancel</button>
            <button
              @click="onSaveSubs"
              class="rounded-lg border border-active bg-active px-3 py-1.5 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              :disabled="subsSaving"
            >{{ subsSaving ? 'Saving...' : 'Save Subs' }}</button>
            <span v-if="subsMessage" class="text-sm" :class="subsMessage.startsWith('Failed') ? 'text-red-400' : 'text-green-500'">{{ subsMessage }}</span>
          </template>

          <!-- Manage mode -->
          <template v-else-if="managing">
            <button
              @click="cancelManage"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
              :disabled="submitting"
            >
              Cancel
            </button>
            <button
              @click="onSaveTeam"
              class="rounded-lg border border-active bg-active px-3 py-1.5 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              :disabled="submitting || !isDirty || !!benchValidationError"
            >
              {{ submitting ? 'Saving...' : 'Save Team' }}
            </button>
            <span v-if="benchValidationError" class="text-sm text-red-400">{{ benchValidationError }}</span>
            <span v-else-if="submitMessage" class="text-sm text-green-500">{{ submitMessage }}</span>
          </template>

          <!-- Default mode buttons -->
          <template v-else>
            <button
              @click="managing = true"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >
              <span class="flex items-center gap-1.5">
                <IconManage class="w-3.5 h-3.5" />
                Manage
              </span>
            </button>
            <button
              v-if="aflMatchStarted"
              @click="enterSubsMode"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >
              <span class="flex items-center gap-1.5">
                <IconSubs class="w-3.5 h-3.5" />
                Substitutions
              </span>
            </button>
          </template>
        </div>

        <!-- Summary bar -->
        <div class="mb-8 rounded-lg border border-border bg-surface-raised px-4 py-3">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-text-heading">Team</h2>
            <div class="flex items-center gap-3">
              <span class="text-sm tabular-nums text-text-muted">{{ starterCount }}/18 starters · {{ benchCount }}/4 bench</span>
              <span class="text-sm font-semibold tabular-nums">{{ grandTotal }}</span>
            </div>
          </div>
        </div>

        <div class="grid gap-8" :class="managing ? 'grid-cols-1 sm:grid-cols-2' : 'grid-cols-1'">
          <!-- Team (left col) -->
          <div>

            <!-- Starter position groups -->
            <div v-for="pos in positions" :key="pos.key" class="mb-6">
              <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold text-text-faint">
                  {{ pos.label }}<span v-if="positionTotal(pos.key) > 0" class="font-normal ml-3">({{ positionTotal(pos.key) }})</span>
                </h3>
              </div>
              <div class="space-y-1">
                <div
                  v-for="(slot, index) in teamSlots[pos.key]"
                  :key="index"
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="slot.player
                    ? (subsMode && slot.player.aflStatus === 'dnp'
                      ? (subbedOutIds.has(slot.player.pmId ?? '') ? 'border-sky-500/40 bg-sky-500/5 cursor-pointer' : 'border-amber-600/30 bg-amber-500/5 cursor-pointer')
                      : 'border-border bg-surface-raised')
                    : 'border-dashed border-border-subtle bg-surface'"
                  @click="onStarterClick(slot.player)"
                >
                  <div v-if="slot.player" class="flex items-center gap-3">
                    <span v-if="pos.key === 'star'" class="text-yellow-400 text-xs">★</span>
                    <div v-if="managing">
                      <div class="font-medium text-sm">{{ slot.player.name }}</div>
                      <div v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</div>
                    </div>
                    <div v-else class="flex flex-col">
                      <div class="flex items-baseline gap-2">
                        <component
                          :is="playerAflMatchRoute(slot.player) ? 'router-link' : 'span'"
                          :to="playerAflMatchRoute(slot.player) ?? undefined"
                          class="font-medium text-sm hover:text-active transition-colors"
                        >{{ slot.player.name }}</component>
                        <span v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</span>
                      </div>
                      <span v-if="effectiveCovering(slot.player.pmId)" class="text-xs text-sky-400">
                        ↑ {{ effectiveCovering(slot.player.pmId)!.name }}
                      </span>
                    </div>
                  </div>
                  <span v-else class="text-text-faint text-sm">Empty slot</span>
                  <div v-if="slot.player && managing" class="flex items-center gap-2">
                    <button
                      v-for="target in positions.filter(p => p.key !== pos.key)"
                      :key="target.key"
                      class="rounded px-1.5 py-0.5 text-xs transition-colors"
                      :disabled="isPositionFull(target.key)"
                      :class="[
                        isPositionFull(target.key) ? 'opacity-30 cursor-not-allowed' : '',
                        target.key === 'star' ? 'text-yellow-400 hover:bg-control-hover hover:text-yellow-300' : 'text-text-faint hover:bg-control-hover hover:text-text'
                      ]"
                      :title="`Move to ${target.label}`"
                      @click="moveToPosition(pos.key, index, target.key)"
                    >
                      {{ target.short }}
                    </button>
                    <button
                      aria-label="Remove"
                      class="text-xs text-red-400 hover:text-red-300 transition-colors"
                      @click="removeFromTeam(pos.key, index)"
                    >
                      <IconBin class="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <div v-else-if="slot.player" class="flex items-center gap-2 shrink-0">
                    <span class="w-28 shrink-0"></span>
                    <span class="w-16 shrink-0">
                      <StatusBadge :status="playerStatus(slot.player)" />
                    </span>
                    <span class="w-28 text-right text-xs tabular-nums text-text-faint shrink-0">{{ playerShowScore(slot.player) ? (positionFormula(pos.key, slot.player) ?? '') : '' }}</span>
                    <span class="w-12 text-right text-sm tabular-nums text-text shrink-0">{{ starterDisplayScore(slot.player, pos.key) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Bench -->
            <div class="mb-6">
              <h3 class="text-sm font-semibold text-text-faint mb-2">Bench</h3>

              <div v-for="(slot, index) in benchDualSlots" :key="index" class="mb-1">
                <div
                  class="flex items-center justify-between rounded-lg border px-4 py-2 transition-colors"
                  :class="[
                    slot.player
                      ? (subsMode && isInterchangeSlot(slot)
                        ? (interchangeApplied ? 'border-sky-500/40 bg-sky-500/5 cursor-pointer' : 'border-amber-600/30 bg-amber-500/5 cursor-pointer')
                        : 'border-border bg-surface-raised')
                      : 'border-dashed border-border-subtle bg-surface',
                    recentlyClearedSlot === index ? '!border-orange-400' : ''
                  ]"
                  @click="onBenchRowClick(slot)"
                >
                  <!-- Left: name -->
                  <div class="flex items-center gap-3 min-w-0">
                    <div v-if="slot.player">
                      <span v-if="!managing && effectiveSubbedForStarter(slot.player.pmId)" class="block text-xs text-sky-400">
                        ↑ {{ effectiveSubbedForStarter(slot.player.pmId)!.name }}
                      </span>
                      <div class="flex items-baseline gap-2" :class="managing ? 'flex-col gap-0' : ''">
                        <component
                          :is="!managing && playerAflMatchRoute(slot.player) ? 'router-link' : 'span'"
                          :to="!managing && playerAflMatchRoute(slot.player) ? playerAflMatchRoute(slot.player) : undefined"
                          class="font-medium text-sm text-text-muted hover:text-active transition-colors"
                        >{{ slot.player.name }}</component>
                        <span v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</span>
                      </div>
                    </div>
                    <span v-else class="text-text-faint text-sm">Empty slot</span>
                  </div>
                  <!-- Right: selectors + remove (manage) or read-only tags -->
                  <div class="flex items-center gap-2 ml-4 shrink-0">
                    <template v-if="slot.player && managing">
                      <select
                        class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                        :value="slot.positions[0] ?? ''"
                        @change="setBenchPosition(index, 0, ($event.target as HTMLSelectElement).value)"
                        aria-label="Position 1"
                      >
                        <option value=""></option>
                        <option v-for="pos in positions" :key="pos.key" :value="pos.key">
                          {{ pos.short }}{{ isBenchPositionUsed(pos.key, index, 0) ? ' ·' : '' }}
                        </option>
                      </select>
                      <select
                        v-if="slot.positions[0] !== 'star'"
                        class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                        :value="slot.positions[1] ?? ''"
                        @change="setBenchPosition(index, 1, ($event.target as HTMLSelectElement).value)"
                        aria-label="Position 2"
                      >
                        <option value=""></option>
                        <option v-for="pos in nonStarPositions" :key="pos.key" :value="pos.key">
                          {{ pos.short }}{{ isBenchPositionUsed(pos.key, index, 1) ? ' ·' : '' }}
                        </option>
                      </select>
                      <button
                        aria-label="Remove"
                        class="text-xs text-red-400 hover:text-red-300 transition-colors"
                        @click="removeBenchDual(index)"
                      >
                        <svg class="w-3.5 h-3.5" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
                          <path d="M2 3.5h10M5.5 3.5V2.5a.5.5 0 01.5-.5h2a.5.5 0 01.5.5v1M6 6.5v4M8 6.5v4M3 3.5l.7 7.5a.5.5 0 00.5.5h5.6a.5.5 0 00.5-.5L11 3.5"/>
                        </svg>
                      </button>
                    </template>
                    <template v-else-if="slot.player">
                      <div class="flex items-center gap-2 shrink-0">
                        <div class="w-28 flex items-center justify-end gap-1 shrink-0">
                          <span v-if="slot.positions[0]" :class="interchangePosition === slot.positions[0] ? 'text-xs rounded px-1.5 py-0.5 bg-sky-500/10 text-sky-400' : effectiveCoveredPosition(slot.player?.pmId) === slot.positions[0] ? 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted ring-1 ring-sky-400/60' : 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted'">
                            {{ positionShort(slot.positions[0]) }}<template v-if="interchangePosition === slot.positions[0]"> · Int</template>
                          </span>
                          <span v-if="slot.positions[1]" :class="interchangePosition === slot.positions[1] ? 'text-xs rounded px-1.5 py-0.5 bg-sky-500/10 text-sky-400' : effectiveCoveredPosition(slot.player?.pmId) === slot.positions[1] ? 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted ring-1 ring-sky-400/60' : 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted'">
                            {{ positionShort(slot.positions[1]) }}<template v-if="interchangePosition === slot.positions[1]"> · Int</template>
                          </span>
                        </div>
                        <span class="w-16 shrink-0">
                          <StatusBadge :status="playerStatus(slot.player)" />
                        </span>
                        <span class="w-28 shrink-0"></span>
                        <span class="w-12 text-right text-sm tabular-nums text-text shrink-0">{{ playerShowScore(slot.player) ? benchScoreDisplay(slot) : '' }}</span>
                      </div>
                    </template>
                  </div>
                </div>
              </div>

              <!-- Interchange -->
              <div v-if="managing" class="mt-3 flex items-center gap-2 justify-end">
                <span class="text-xs text-text-faint">Interchange</span>
                <select
                  class="text-xs rounded bg-control text-text px-1 py-0.5 border border-border"
                  aria-label="Interchange"
                  :value="interchangePosition ?? ''"
                  @change="setInterchange(($event.target as HTMLSelectElement).value)"
                >
                  <option value=""></option>
                  <option v-for="pos in positions" :key="pos.key" :value="pos.key">{{ pos.short }}</option>
                </select>
              </div>
            </div>

          </div>

          <!-- Squad panel (right col, manage mode only) -->
          <div v-if="managing">
            <h2 class="text-lg font-semibold text-text-heading mb-3">Squad ({{ availablePlayers.length }})</h2>
            <div class="space-y-1">
              <div
                v-for="player in availablePlayers"
                :key="player.id"
                class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-2"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <div>
                    <div class="font-medium text-sm">{{ player.name }}</div>
                    <div v-if="player.club" class="text-xs text-text-muted">{{ player.club }}</div>
                  </div>
                  <span v-if="playerShowScore(player)" class="text-sm tabular-nums text-text shrink-0">{{ player.score }}</span>
                </div>
                <div class="flex items-center gap-1">
                  <!-- Position buttons (starters) -->
                  <button
                    v-for="pos in positions"
                    :key="pos.key"
                    class="rounded px-2 py-0.5 text-xs transition-colors"
                    :class="[
                      isPositionFull(pos.key) ? 'opacity-30 cursor-not-allowed' : '',
                      pos.key === 'star' ? 'text-yellow-400 hover:bg-control-hover hover:text-yellow-300' : 'text-text-muted hover:bg-control-hover hover:text-text'
                    ]"
                    :disabled="isPositionFull(pos.key)"
                    @click="addToTeam(pos.key, player)"
                  >
                    {{ pos.short }}
                  </button>
                  <span class="w-px h-4 bg-border mx-0.5 shrink-0"></span>
                  <!-- Bench -->
                  <button
                    class="rounded px-2 py-0.5 text-xs text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                    :disabled="benchDualFull"
                    :class="{ 'opacity-30 cursor-not-allowed': benchDualFull }"
                    title="Add to bench"
                    @click="addBenchDual(player)"
                  >
                    B
                  </button>
                </div>
              </div>
              <p v-if="availablePlayers.length === 0" class="text-sm text-text-faint">All players assigned</p>
            </div>
            <template v-if="tradedPlayers.length > 0">
              <button
                @click="showTraded = !showTraded"
                class="flex items-center gap-1.5 text-xs font-medium text-text-faint uppercase tracking-wide mt-4 mb-2 hover:text-text-muted transition-colors"
              >
                <span>{{ showTraded ? '▾' : '▸' }}</span>
                Traded ({{ tradedPlayers.length }})
              </button>
              <div v-if="showTraded" class="space-y-1">
                <div
                  v-for="player in tradedPlayers"
                  :key="player.id"
                  class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-2 opacity-40"
                >
                  <div class="flex items-center gap-3 min-w-0">
                    <div>
                      <div class="font-medium text-sm">{{ player.name }}</div>
                      <div v-if="player.club" class="text-xs text-text-muted">{{ player.club }}</div>
                    </div>
                    <span v-if="playerShowScore(player)" class="text-sm tabular-nums text-text shrink-0">{{ player.score }}</span>
                  </div>
                  <div class="flex items-center gap-1">
                    <button
                      v-for="pos in positions"
                      :key="pos.key"
                      class="rounded px-2 py-0.5 text-xs transition-colors"
                      :class="[
                        isPositionFull(pos.key) ? 'opacity-30 cursor-not-allowed' : '',
                        pos.key === 'star' ? 'text-yellow-400 hover:bg-control-hover hover:text-yellow-300' : 'text-text-muted hover:bg-control-hover hover:text-text'
                      ]"
                      :disabled="isPositionFull(pos.key)"
                      @click="addToTeam(pos.key, player)"
                    >{{ pos.short }}</button>
                    <span class="w-px h-4 bg-border mx-0.5 shrink-0"></span>
                    <button
                      class="rounded px-2 py-0.5 text-xs text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                      :disabled="benchDualFull"
                      :class="{ 'opacity-30 cursor-not-allowed': benchDualFull }"
                      @click="addBenchDual(player)"
                    >B</button>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </template>
      <p v-else class="text-text-faint">No club selected. Choose a club in the nav bar.</p>

      <div v-if="bootstrapRoundId" class="mt-8">
        <router-link
          :to="{ name: 'ffl-data-ops', query: { tab: 'team-submission', round: bootstrapRoundId } }"
          class="flex items-center gap-1.5 text-sm text-text-muted hover:text-text transition-colors"
        >
          <IconDataOps class="w-4 h-4" />
          Data Ops
        </router-link>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_ROUND, GET_FFL_SEASON_CLUBS, GET_FFL_CLUB_SEASON, GET_FFL_CLUB_MATCH } from '../api/queries'
import { SET_FFL_TEAM, DECLARE_FFL_SUBSTITUTIONS } from '../api/mutations'
import Breadcrumb from '../components/Breadcrumb.vue'
import StatusBadge from '../components/StatusBadge.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { positionFormula } from '../utils/position'
import IconSquad from '../components/icons/IconSquad.vue'
import IconManage from '../components/icons/IconManage.vue'
import IconSubs from '../components/icons/IconSubs.vue'
import IconBin from '../components/icons/IconBin.vue'
import IconDataOps from '@/features/data-ops/components/icons/IconDataOps.vue'
import { useFflState } from '../composables/useFflState'
import { POSITION_MULTIPLIERS } from '../utils/position'

const props = defineProps<{ clubMatchId: string }>()

const positions = [
  { key: 'goals',     label: 'Goals',     short: 'G',  count: 3 },
  { key: 'kicks',     label: 'Kicks',     short: 'K',  count: 4 },
  { key: 'handballs', label: 'Handballs', short: 'H',  count: 4 },
  { key: 'marks',     label: 'Marks',     short: 'M',  count: 2 },
  { key: 'tackles',   label: 'Tackles',   short: 'T',  count: 2 },
  { key: 'hitouts',   label: 'Hitouts',   short: 'R',  count: 2 },
  { key: 'star',      label: 'Star',      short: '★',  count: 1 },
] as const

type PositionKey = typeof positions[number]['key']
type NonStarPositionKey = Exclude<PositionKey, 'star'>

const nonStarPositions = positions.filter(p => p.key !== 'star')

interface SquadPlayer {
  id: string
  name: string
  club: string | null
  status: string | null
  aflStatus: string | null
  score: number | null
  aflMatchId: string | null
  toRoundId: string | null
  pmId: string | null
  goals: number | null
  kicks: number | null
  handballs: number | null
  marks: number | null
  tackles: number | null
  hitouts: number | null
}

interface Slot {
  player: SquadPlayer | null
}

interface BenchDualSlot {
  player: SquadPlayer | null
  positions: [PositionKey | null, NonStarPositionKey | null]
}

const { selectedClubId, setClub } = useFflState()
const managing = ref(false)
const subsMode = ref(false)

const { result: clubMatchBootstrap, loading: bootstrapLoading } = useQuery(
  GET_FFL_CLUB_MATCH,
  () => ({ id: props.clubMatchId }),
  { errorPolicy: 'all' },
)

const bootstrapRoundId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.roundId ?? '')
const bootstrapSeasonId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.seasonId ?? '')
const bootstrapClubSeasonId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.clubSeasonId ?? '')
const bootstrapClubId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.club?.id ?? '')

watch(bootstrapClubId, (id) => {
  if (id) setClub(id)
}, { immediate: true })

const { result: roundResult, loading: roundLoading, error: roundError } = useQuery(
  GET_FFL_ROUND,
  () => ({ id: bootstrapRoundId.value }),
  () => ({ enabled: !!bootstrapRoundId.value, errorPolicy: 'all' }),
)
const { result: seasonResult, loading: seasonLoading } = useQuery(
  GET_FFL_SEASON_CLUBS,
  () => ({ seasonId: bootstrapSeasonId.value }),
  () => ({ enabled: !!bootstrapSeasonId.value, errorPolicy: 'all' }),
)

const round = computed(() => roundResult.value?.fflRound ?? null)
const currentRound = round
const season = computed(() => seasonResult.value?.fflSeason ?? null)

const loading = computed(() => bootstrapLoading.value || roundLoading.value || seasonLoading.value)
const error = computed(() => roundError.value ?? null)

const selectedClubSeason = computed(() =>
  season.value?.ladder.find((cs: { club: { id: string } }) => cs.club.id === bootstrapClubId.value) ?? null
)

const { result: clubSeasonResult } = useQuery(
  GET_FFL_CLUB_SEASON,
  () => ({ id: bootstrapClubSeasonId.value }),
  () => ({ enabled: !!bootstrapClubSeasonId.value, errorPolicy: 'all' }),
)

const clubSeasonData = computed(() => clubSeasonResult.value?.fflClubSeason ?? null)

const allRounds = computed(() => round.value?.season.rounds ?? [])

const breadcrumbs = computed(() => {
  if (!round.value) return []
  const crumbs: { label: string; to?: object }[] = [
    { label: 'FFL' },
    { label: round.value.season.name, to: { name: 'home' } },
    { label: round.value.name, to: { name: 'ffl-round', params: { roundId: bootstrapRoundId.value } } },
  ]
  if (currentMatch.value) {
    const home = currentMatch.value.homeClubMatch?.club.name ?? '?'
    const away = currentMatch.value.awayClubMatch?.club.name ?? '?'
    crumbs.push({ label: `${home} v ${away}`, to: { name: 'ffl-match', params: { matchId: currentMatch.value.id } } })
  }
  return crumbs
})

const prevRound = computed(() => {
  const rounds = allRounds.value
  const idx = rounds.findIndex((r: { id: string }) => r.id === bootstrapRoundId.value)
  return idx > 0 ? rounds[idx - 1] : null
})

const nextRound = computed(() => {
  const rounds = allRounds.value
  const idx = rounds.findIndex((r: { id: string }) => r.id === bootstrapRoundId.value)
  return idx >= 0 && idx < rounds.length - 1 ? rounds[idx + 1] : null
})

type RoundMatchEntry = { homeClubMatch?: { id: string; clubSeasonId: string } | null; awayClubMatch?: { id: string; clubSeasonId: string } | null }

function clubMatchIdForRound(roundId: string): string | null {
  const roundData = round.value?.season.rounds.find((r: { id: string }) => r.id === roundId)
  if (!roundData) return null
  const csId = bootstrapClubSeasonId.value
  const m: RoundMatchEntry | undefined = roundData.matches?.find((m: RoundMatchEntry) =>
    m.homeClubMatch?.clubSeasonId === csId || m.awayClubMatch?.clubSeasonId === csId
  )
  if (!m) return null
  return m.homeClubMatch?.clubSeasonId === csId ? (m.homeClubMatch?.id ?? null) : (m.awayClubMatch?.id ?? null)
}

const prevClubMatchId = computed(() => prevRound.value ? clubMatchIdForRound(prevRound.value.id) : null)
const nextClubMatchId = computed(() => nextRound.value ? clubMatchIdForRound(nextRound.value.id) : null)

const currentMatch = computed(() => {
  if (!round.value) return null
  return round.value.matches.find((m: { homeClubMatch?: { id: string } | null; awayClubMatch?: { id: string } | null }) =>
    m.homeClubMatch?.id === props.clubMatchId || m.awayClubMatch?.id === props.clubMatchId
  ) ?? null
})

const clubMatch = computed(() => {
  if (!currentMatch.value) return null
  const m = currentMatch.value
  if (m.homeClubMatch?.id === props.clubMatchId) return m.homeClubMatch
  if (m.awayClubMatch?.id === props.clubMatchId) return m.awayClubMatch
  return null
})

const playerMatchBySeasonId = computed(() => {
  const map = new Map<string, { pmId: string; score: number | null; club: string | null; status: string | null; aflStatus: string | null; aflMatchId: string | null; goals: number | null; kicks: number | null; handballs: number | null; marks: number | null; tackles: number | null; hitouts: number | null }>()
  for (const pm of clubMatch.value?.playerMatches ?? []) {
    map.set(pm.playerSeasonId, {
      pmId: pm.id,
      score: pm.score ?? null,
      club: pm.playerSeason?.aflPlayerSeason?.clubSeason?.club?.name ?? null,
      status: pm.status ?? null,
      aflStatus: pm.aflStatus ?? null,
      aflMatchId: pm.aflPlayerMatch?.clubMatch?.match?.id ?? null,
      goals: pm.aflPlayerMatch?.goals ?? null,
      kicks: pm.aflPlayerMatch?.kicks ?? null,
      handballs: pm.aflPlayerMatch?.handballs ?? null,
      marks: pm.aflPlayerMatch?.marks ?? null,
      tackles: pm.aflPlayerMatch?.tackles ?? null,
      hitouts: pm.aflPlayerMatch?.hitouts ?? null,
    })
  }
  return map
})

const squad = computed<SquadPlayer[]>(() => {
  if (!clubSeasonData.value) return []
  return clubSeasonData.value.players.nodes.map((r: {
    id: string
    player: { aflPlayer: { name: string } }
    aflPlayerSeason?: { clubSeason?: { club?: { name: string } | null } | null } | null
    toRoundId?: string | null
  }) => {
    const pm = playerMatchBySeasonId.value.get(r.id)
    return {
      id: r.id,
      name: r.player.aflPlayer.name,
      club: pm?.club ?? r.aflPlayerSeason?.clubSeason?.club?.name ?? null,
      status: pm?.status ?? null,
      aflStatus: pm?.aflStatus ?? null,
      score: pm?.score ?? null,
      aflMatchId: pm?.aflMatchId ?? null,
      toRoundId: r.toRoundId ?? null,
      pmId: pm?.pmId ?? null,
      goals: pm?.goals ?? null,
      kicks: pm?.kicks ?? null,
      handballs: pm?.handballs ?? null,
      marks: pm?.marks ?? null,
      tackles: pm?.tackles ?? null,
      hitouts: pm?.hitouts ?? null,
    }
  })
})

function playerStatus(player: SquadPlayer): string | null {
  if (player.status === 'subbed') return 'subbed'
  if (player.status === 'interchanged') return 'interchanged'
  return player.aflStatus
}

function playerShowScore(player: SquadPlayer): boolean {
  return player.aflStatus === 'played' || player.aflStatus === 'playing'
}

function benchPositionScore(player: SquadPlayer, pos: string): number | null {
  if (!playerShowScore(player)) return null
  if (pos === 'star') {
    if (player.goals === null) return null
    return (player.goals ?? 0) * 5 + (player.kicks ?? 0) + (player.handballs ?? 0) + (player.marks ?? 0) * 2 + (player.tackles ?? 0) * 4
  }
  const statMap: Record<string, number | null> = {
    goals: player.goals, kicks: player.kicks, handballs: player.handballs,
    marks: player.marks, tackles: player.tackles, hitouts: player.hitouts,
  }
  const stat = statMap[pos] ?? null
  if (stat === null) return null
  return stat * (POSITION_MULTIPLIERS[pos] ?? 1)
}

function benchScoreDisplay(slot: BenchDualSlot): string {
  if (!slot.player) return ''
  const parts: string[] = []
  for (const pos of slot.positions) {
    if (!pos) continue
    const s = benchPositionScore(slot.player, pos)
    parts.push(s !== null ? String(s) : '?')
  }
  return parts.join('/')
}

function playerAflMatchRoute(player: SquadPlayer): { name: string; params: { matchId: string } } | null {
  if (!player.aflMatchId) return null
  return { name: 'afl-match', params: { matchId: player.aflMatchId } }
}

// ── Team state ──────────────────────────────────────────────────────────────

const createSlots = (count: number): Slot[] => Array.from({ length: count }, () => ({ player: null }))

const teamSlots = ref<Record<PositionKey, Slot[]>>(
  Object.fromEntries(positions.map(p => [p.key, createSlots(p.count)])) as Record<PositionKey, Slot[]>
)

const benchDualSlots = ref<BenchDualSlot[]>([
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
  { player: null, positions: [null, null] },
])

// The position that acts as the free interchange slot.
const interchangePosition = ref<string | null>(null)

// Highlight recently-stolen bench slot index (orange border flash).
const recentlyClearedSlot = ref<number | null>(null)
let clearHighlightTimer: ReturnType<typeof setTimeout> | null = null

// Track the match ID we last loaded from to avoid Apollo cache updates wiping local edits.
const initializedMatchId = ref<string | null>(null)

// Dirty tracking — snapshot taken after load or save; compared to detect unsaved changes.
const isDirty = ref(false)

function takeSnapshot() {
  isDirty.value = false
}

function markDirty() {
  isDirty.value = true
}

function resetTeamState() {
  for (const pos of positions) {
    teamSlots.value[pos.key] = createSlots(pos.count)
  }
  benchDualSlots.value = [
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
    { player: null, positions: [null, null] },
  ]
  interchangePosition.value = null
}

function loadTeamFromMatch(cm: NonNullable<typeof clubMatch.value>) {
  resetTeamState()
  takeSnapshot()
  if (!cm.playerMatches) return

  let dualIndex = 0
  for (const pm of cm.playerMatches) {
    const squadEntry = squad.value.find(s => s.id === pm.playerSeasonId)
    const player: SquadPlayer = {
      id: pm.playerSeasonId,
      name: pm.player.aflPlayer.name,
      club: pm.playerSeason?.aflPlayerSeason?.clubSeason?.club?.name ?? squadEntry?.club ?? null,
      status: pm.status ?? null,
      aflStatus: pm.aflStatus ?? null,
      score: pm.score ?? null,
      aflMatchId: pm.aflPlayerMatch?.clubMatch?.match?.id ?? null,
      toRoundId: squadEntry?.toRoundId ?? null,
      pmId: pm.id,
      goals: pm.aflPlayerMatch?.goals ?? null,
      kicks: pm.aflPlayerMatch?.kicks ?? null,
      handballs: pm.aflPlayerMatch?.handballs ?? null,
      marks: pm.aflPlayerMatch?.marks ?? null,
      tackles: pm.aflPlayerMatch?.tackles ?? null,
      hitouts: pm.aflPlayerMatch?.hitouts ?? null,
    }
    const isBench = pm.backupPositions != null || pm.interchangePosition != null

    if (!isBench) {
      const posSlots = teamSlots.value[pm.position as PositionKey]
      if (posSlots) {
        const slot = posSlots.find((s: Slot) => !s.player)
        if (slot) slot.player = player
      }
    } else if (dualIndex < 4) {
      if (pm.backupPositions === 'star') {
        benchDualSlots.value[dualIndex].player = player
        benchDualSlots.value[dualIndex].positions = ['star', null]
      } else if (pm.backupPositions) {
        const parts = pm.backupPositions.split(',').map((p: string) => p.trim()) as NonStarPositionKey[]
        benchDualSlots.value[dualIndex].player = player
        benchDualSlots.value[dualIndex].positions = [parts[0] ?? null, parts[1] ?? null]
      }
      if (pm.interchangePosition) interchangePosition.value = pm.interchangePosition
      dualIndex++
    }
  }
}

// Load existing team from server data — only when the match changes, not on every Apollo cache update.
// { immediate: true } ensures this fires on component remount when Apollo cache already has data
// (without it, watch only fires on changes — a cache hit on remount produces no change event).
watch(clubMatch, (cm) => {
  if (!cm) return
  if (cm.id === initializedMatchId.value) return  // already initialised for this match; don't reset local edits
  initializedMatchId.value = cm.id
  loadTeamFromMatch(cm)
}, { immediate: true })

// ── Computed helpers ──────────────────────────────────────────────────────────

const assignedPlayerIds = computed(() => {
  const ids = new Set<string>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player) ids.add(slot.player.id)
    }
  }
  for (const slot of benchDualSlots.value) {
    if (slot.player) ids.add(slot.player.id)
  }
  return ids
})

const availablePlayers = computed(() =>
  squad.value.filter(p => !p.toRoundId && !assignedPlayerIds.value.has(p.id))
)

const tradedPlayers = computed(() =>
  squad.value.filter(p => !!p.toRoundId && !assignedPlayerIds.value.has(p.id))
)

const showTraded = ref(false)

const starterCount = computed(() => {
  let count = 0
  for (const pos of positions) {
    count += teamSlots.value[pos.key].filter(s => s.player).length
  }
  return count
})

const benchCount = computed(() => benchDualSlots.value.filter(s => s.player).length)

function positionTotal(key: PositionKey): number {
  return teamSlots.value[key].reduce((sum: number, s: Slot) => {
    if (!s.player) return sum
    const score = starterDisplayScore(s.player, key)
    return sum + (score === '' ? 0 : Number(score))
  }, 0)
}

const grandTotal = computed(() => {
  if (!subsMode.value) return clubMatch.value?.score ?? 0
  let total = 0
  for (const pos of positions) total += positionTotal(pos.key)
  return total
})

const benchDualFull = computed(() => benchDualSlots.value.every(s => s.player !== null))

const benchValidationError = computed<string | null>(() => {
  for (const slot of benchDualSlots.value) {
    if (!slot.player) continue
    const [p1, p2] = slot.positions
    if (!p1) return 'Each bench player must have a position assigned'
    if (p1 !== 'star' && !p2) return 'Non-star bench players need two backup positions'
  }
  const filledCount = benchDualSlots.value.filter(s => s.player).length
  if (filledCount > 1 && !interchangePosition.value) return 'Choose an interchange position'
  return null
})

const isPositionFull = (key: PositionKey) =>
  teamSlots.value[key].every(s => s.player !== null)

// Returns true if posKey is already used by another bench slot (excluding slotIndex+sideIndex).
function isBenchPositionUsed(posKey: string, slotIndex: number, sideIndex: number): boolean {
  for (let i = 0; i < benchDualSlots.value.length; i++) {
    const slot = benchDualSlots.value[i]
    for (const j of [0, 1] as const) {
      if (i === slotIndex && j === sideIndex) continue
      if (slot.positions[j] === posKey) return true
    }
  }
  return false
}

function positionShort(key: string): string {
  return positions.find(p => p.key === key)?.short ?? key
}

// ── Team management ─────────────────────────────────────────────────────────

function addToTeam(key: PositionKey, player: SquadPlayer) {
  const slot = teamSlots.value[key].find(s => !s.player)
  if (slot) { slot.player = player; markDirty() }
}

function removeFromTeam(key: PositionKey, index: number) {
  teamSlots.value[key][index].player = null
  markDirty()
}

function moveToPosition(fromKey: PositionKey, fromIndex: number, toKey: PositionKey) {
  const player = teamSlots.value[fromKey][fromIndex].player
  if (!player) return
  const toSlot = teamSlots.value[toKey].find(s => !s.player)
  if (!toSlot) return
  teamSlots.value[fromKey][fromIndex].player = null
  toSlot.player = player
  markDirty()
}

function addBenchDual(player: SquadPlayer) {
  const slot = benchDualSlots.value.find(s => !s.player)
  if (slot) { slot.player = player; markDirty() }
}

function removeBenchDual(index: number) {
  benchDualSlots.value[index].player = null
  benchDualSlots.value[index].positions = [null, null]
  markDirty()
}

function setBenchPosition(slotIndex: number, sideIndex: 0 | 1, value: string) {
  const slot = benchDualSlots.value[slotIndex]
  // Steal position from any other slot that already has it, and flash that slot
  if (value) {
    for (let i = 0; i < benchDualSlots.value.length; i++) {
      const other = benchDualSlots.value[i]
      if (other.positions[0] === value && !(i === slotIndex && sideIndex === 0)) {
        other.positions[0] = null
        flashClearedSlot(i)
      } else if (other.positions[1] === value && !(i === slotIndex && sideIndex === 1)) {
        other.positions[1] = null
        flashClearedSlot(i)
      }
    }
  }
  if (sideIndex === 0) {
    slot.positions[0] = (value || null) as PositionKey | null
    if (value === 'star') slot.positions[1] = null
  } else {
    slot.positions[1] = (value || null) as NonStarPositionKey | null
  }
  markDirty()
}

function flashClearedSlot(index: number) {
  if (clearHighlightTimer) clearTimeout(clearHighlightTimer)
  recentlyClearedSlot.value = index
  clearHighlightTimer = setTimeout(() => { recentlyClearedSlot.value = null }, 2000)
}

function setInterchange(value: string) {
  interchangePosition.value = value || null
  markDirty()
}

// ── Subs mode ────────────────────────────────────────────────────────────────

// True when the AFL match is underway or complete (excludes 'named' — pre-match only).
const aflMatchStarted = computed(() => {
  const pms = clubMatch.value?.playerMatches ?? []
  return pms.some((pm: { aflStatus: string | null }) =>
    pm.aflStatus === 'playing' || pm.aflStatus === 'played' || pm.aflStatus === 'dnp'
  )
})

// Bench player with InterchangePosition set (at most one per team).
const interchangeBenchPlayer = computed(() => {
  const pms = clubMatch.value?.playerMatches ?? []
  return pms.find((pm: { interchangePosition: string | null }) => pm.interchangePosition != null) ?? null
})

// Starter at the interchange position with the lowest score (the one that would be displaced).
const interchangeTargetStarter = computed(() => {
  const bench = interchangeBenchPlayer.value
  if (!bench) return null
  const pms = clubMatch.value?.playerMatches ?? []
  const starters = pms.filter((pm: { backupPositions: string | null; interchangePosition: string | null; position: string | null }) =>
    pm.backupPositions == null && pm.interchangePosition == null && pm.position === bench.interchangePosition
  )
  if (!starters.length) return null
  return starters.reduce((lowest: typeof starters[0], pm: typeof starters[0]) => pm.score < lowest.score ? pm : lowest)
})

// Whether the interchange bench player currently outscores the target starter.
const interchangeBeneficial = computed(() => {
  const bench = interchangeBenchPlayer.value
  const starter = interchangeTargetStarter.value
  if (!bench || !starter) return false
  return bench.score > starter.score
})

// Subs UI state.
const subbedOutIds = ref<Set<string>>(new Set())
const interchangeApplied = ref(false)
const subsSaving = ref(false)
const subsMessage = ref('')

function initSubsState() {
  const pms = clubMatch.value?.playerMatches ?? []
  // Pre-populate from stored TM decisions.
  subbedOutIds.value = new Set(
    pms
      .filter((pm: { status: string | null }) => pm.status === 'subbed')
      .map((pm: { id: string }) => pm.id)
  )
  // Check if interchange is currently applied.
  interchangeApplied.value = pms.some((pm: { status: string | null }) => pm.status === 'interchanged')
  // Default interchange to checked if beneficial and no decision stored yet.
  if (!pms.some((pm: { status: string | null }) => pm.status === 'subbed' || pm.status === 'interchanged')) {
    interchangeApplied.value = interchangeBeneficial.value
  }
}

function isInterchangeSlot(slot: BenchDualSlot): boolean {
  if (!interchangePosition.value) return false
  return slot.positions[0] === interchangePosition.value || slot.positions[1] === interchangePosition.value
}

function enterSubsMode() {
  initSubsState()
  subsMode.value = true
}

function exitSubsMode() {
  subsMode.value = false
  subsMessage.value = ''
}

function toggleSub(pmId: string) {
  if (subbedOutIds.value.has(pmId)) {
    subbedOutIds.value.delete(pmId)
  } else {
    subbedOutIds.value.add(pmId)
  }
}

// Maps subbed-out starter pmId → the first bench player whose backup positions cover that starter's position.
const subsMapping = computed(() => {
  const map = new Map<string, SquadPlayer>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (!slot.player || !subbedOutIds.value.has(slot.player.pmId ?? '')) continue
      for (const bSlot of benchDualSlots.value) {
        if (!bSlot.player) continue
        if ((bSlot.positions as (string | null)[]).includes(pos.key)) {
          map.set(slot.player.pmId!, bSlot.player)
          break
        }
      }
    }
  }
  return map
})

// Covering map based on saved server state (status === 'subbed') — used in normal (non-subs) mode.
const savedSubsMap = computed(() => {
  const map = new Map<string, SquadPlayer>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (!slot.player?.pmId || slot.player.status !== 'subbed') continue
      for (const bSlot of benchDualSlots.value) {
        if (!bSlot.player) continue
        if ((bSlot.positions as (string | null)[]).includes(pos.key)) {
          map.set(slot.player.pmId, bSlot.player)
          break
        }
      }
    }
  }
  return map
})

// Returns the covering bench player for a starter — subs mode uses live UI state, normal mode uses saved server state.
function effectiveCovering(pmId: string | null): SquadPlayer | null {
  if (!pmId) return null
  return subsMode.value
    ? (subsMapping.value.get(pmId) ?? null)
    : (savedSubsMap.value.get(pmId) ?? null)
}

// Maps bench pmId → the starter they are subbing for (subs mode — live UI state).
const subsStarterMap = computed(() => {
  const map = new Map<string, SquadPlayer>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (!slot.player?.pmId || !subbedOutIds.value.has(slot.player.pmId)) continue
      const cp = subsMapping.value.get(slot.player.pmId)
      if (cp?.pmId) map.set(cp.pmId, slot.player)
    }
  }
  return map
})

// Maps bench pmId → the starter they are subbing for (normal mode — saved server state).
const savedSubsStarterMap = computed(() => {
  const map = new Map<string, SquadPlayer>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (!slot.player?.pmId || slot.player.status !== 'subbed') continue
      const cp = savedSubsMap.value.get(slot.player.pmId)
      if (cp?.pmId) map.set(cp.pmId, slot.player)
    }
  }
  return map
})

// Returns the starter a bench player is subbing for — subs mode uses live UI state, normal mode uses saved server state.
function effectiveSubbedForStarter(benchPmId: string | null): SquadPlayer | null {
  if (!benchPmId) return null
  return subsMode.value
    ? (subsStarterMap.value.get(benchPmId) ?? null)
    : (savedSubsStarterMap.value.get(benchPmId) ?? null)
}

// Returns the position key the bench player is actively covering (used to border-highlight the right pill).
function effectiveCoveredPosition(benchPmId: string | null): string | null {
  const starter = effectiveSubbedForStarter(benchPmId)
  if (!starter?.pmId) return null
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player?.pmId === starter.pmId) return pos.key
    }
  }
  return null
}

function onStarterClick(player: SquadPlayer | null) {
  if (!player || !subsMode.value || player.aflStatus !== 'dnp') return
  toggleSub(player.pmId ?? '')
}

function onBenchRowClick(slot: BenchDualSlot) {
  if (!subsMode.value || !slot.player || !isInterchangeSlot(slot)) return
  interchangeApplied.value = !interchangeApplied.value
}

function starterDisplayScore(player: SquadPlayer, posKey: string): number | string {
  if (subsMode.value && subbedOutIds.value.has(player.pmId ?? '')) {
    const cp = subsMapping.value.get(player.pmId ?? '')
    if (cp) return benchPositionScore(cp, posKey) ?? ''
  } else if (!subsMode.value) {
    const cp = savedSubsMap.value.get(player.pmId ?? '')
    if (cp) return benchPositionScore(cp, posKey) ?? ''
  }
  return playerShowScore(player) ? (player.score ?? '') : ''
}

const { mutate: declareSubs } = useMutation(DECLARE_FFL_SUBSTITUTIONS, () => ({
  refetchQueries: [{ query: GET_FFL_ROUND, variables: { id: bootstrapRoundId.value } }],
  awaitRefetchQueries: true,
}))

async function onSaveSubs() {
  if (!clubMatch.value) return
  subsSaving.value = true
  subsMessage.value = ''
  try {
    await declareSubs({
      input: {
        clubMatchId: clubMatch.value.id,
        subbedOutPlayerMatchIds: Array.from(subbedOutIds.value),
        interchangeApplied: interchangeApplied.value,
      },
    })
    if (clubMatch.value) {
      loadTeamFromMatch(clubMatch.value)
      initializedMatchId.value = clubMatch.value.id
    }
    exitSubsMode()
  } catch {
    subsMessage.value = 'Failed to save substitutions'
  } finally {
    subsSaving.value = false
  }
}

// ── Submit ────────────────────────────────────────────────────────────────────

const { mutate: setTeam } = useMutation(SET_FFL_TEAM, () => ({
  refetchQueries: [{ query: GET_FFL_ROUND, variables: { id: bootstrapRoundId.value } }],
  awaitRefetchQueries: true,
}))
const submitting = ref(false)
const submitMessage = ref('')

async function onSaveTeam() {
  await submitTeam()
  managing.value = false
}

function cancelManage() {
  if (clubMatch.value) loadTeamFromMatch(clubMatch.value)
  managing.value = false
}

async function submitTeam() {
  if (!clubMatch.value) return
  submitting.value = true
  submitMessage.value = ''

  const players: {
    playerSeasonId: string
    position: string
    backupPositions?: string
    interchangePosition?: string
  }[] = []

  // Starters
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player) {
        players.push({ playerSeasonId: slot.player.id, position: pos.key })
      }
    }
  }

  // Bench slots
  for (const slot of benchDualSlots.value) {
    if (!slot.player) continue
    const [p1, p2] = slot.positions
    const isStar = p1 === 'star'
    const bp = isStar ? 'star' : [p1, p2].filter(Boolean).join(',')
    const entry: (typeof players)[number] = {
      playerSeasonId: slot.player.id,
      position: p1 ?? p2 ?? 'goals',
      backupPositions: bp || undefined,
    }
    if (interchangePosition.value && (p1 === interchangePosition.value || p2 === interchangePosition.value)) {
      entry.interchangePosition = interchangePosition.value ?? undefined
    }
    players.push(entry)
  }

  try {
    await setTeam({ input: { clubMatchId: clubMatch.value.id, players } })
    takeSnapshot()
    submitMessage.value = 'Saved'
    setTimeout(() => { submitMessage.value = '' }, 3000)
  } catch (e) {
    submitMessage.value = 'Failed to save team'
  } finally {
    submitting.value = false
  }
}
</script>
