<template>
  <div>
    <!-- Click-away backdrop for popup menus -->
    <div v-if="openMenuKey" class="fixed inset-0 z-40" @click="closeMenu" />
    <div v-if="loading" class="text-text-faint">Loading...</div>
    <div v-else-if="error" class="text-red-400">{{ error.message }}</div>
    <NotFound v-else-if="notFound" entity="Team" />
    <template v-else-if="round">
      <div class="mb-6">
        <Breadcrumb v-if="currentRound" :items="breadcrumbs" />
        <div class="flex items-center">
          <h1 class="text-2xl font-bold flex items-center gap-3">
            <img v-if="selectedClubSeason" :src="clubLogoUrl(selectedClubSeason.club.name)" :alt="selectedClubSeason.club.name" class="w-10 h-10 object-contain" />
            {{ selectedClubSeason?.club.name ?? '' }}<span v-if="!readonly" class="font-normal text-text-muted"> · Team Builder</span>
          </h1>
          <div class="flex items-center gap-3 ml-auto">
            <div class="flex items-center gap-1 rounded-lg border border-border px-1">
              <router-link
                v-if="prevClubMatchId"
                :to="{ name: readonly ? 'ffl-club-match' : 'ffl-club-match-edit', params: { clubMatchId: prevClubMatchId } }"
                class="w-6 h-6 flex items-center justify-center rounded text-text-muted hover:bg-control-hover hover:text-text transition-colors text-sm"
                title="Previous round"
              >‹</router-link>
              <span v-else class="w-6 h-6 flex items-center justify-center text-text-faint text-sm opacity-30">‹</span>
              <span class="text-sm text-text-muted tabular-nums">{{ currentRound?.name }}</span>
              <router-link
                v-if="nextClubMatchId"
                :to="{ name: readonly ? 'ffl-club-match' : 'ffl-club-match-edit', params: { clubMatchId: nextClubMatchId } }"
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
            <router-link
              v-if="readonly && isMyClub && bootstrapClubMatchId"
              :to="{ name: 'ffl-club-match-edit', params: { clubMatchId: bootstrapClubMatchId } }"
              class="flex items-center gap-1.5 text-sm text-text-muted hover:text-text transition-colors"
            >
              <IconTeamBuilder class="w-4 h-4" />
              Team Builder
            </router-link>
          </div>
        </div>
      </div>

      <template v-if="selectedClubSeason && clubMatch">
        <div class="mb-6 flex items-center gap-4 flex-wrap">
          <!-- Subs mode -->
          <template v-if="!readonly && subsMode">
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
            <span v-if="subsMessage" class="text-sm" :class="subsError ? 'text-red-400' : 'text-green-500'">{{ subsMessage }}</span>
          </template>

          <!-- Manage mode -->
          <template v-else-if="!readonly && managing">
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
            <span class="w-2 shrink-0" />
            <button
              v-if="prevClubMatchId && !clubMatchLocked"
              @click="copyPreviousTeam"
              :disabled="prevTeamLoading"
              :title="`Set team from ${prevRound?.name}`"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text-muted hover:text-text hover:bg-surface-hover transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >
              {{ prevTeamLoading ? 'Replicating…' : `${prevRound?.name} Team` }}
            </button>
            <button
              v-if="!clubMatchLocked"
              @click="applyBestTeam"
              title="Set team for the highest projected total (uses active Last 5/Season stat source)"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text-muted hover:text-text hover:bg-surface-hover transition-colors"
            >
              Best Team
            </button>
            <button
              v-if="!clubMatchLocked"
              @click="() => { resetTeamState(); markDirty() }"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text-muted hover:text-text hover:bg-surface-hover transition-colors"
            >
              Clear
            </button>
          </template>

          <!-- Default mode buttons -->
          <template v-else-if="!readonly">
            <button
              @click="managing = true"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >
              <span class="flex items-center gap-1.5">
                <IconManage class="w-3.5 h-3.5" />
                Build Team
              </span>
            </button>
            <button
              v-if="aflMatchStarted"
              @click="enterSubsMode"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >
              <span class="flex items-center gap-1.5">
                <IconSubs class="w-3.5 h-3.5" />
                Make Substitutions
              </span>
            </button>
            <button
              v-if="starterCount > 0 && (clubMatchDataStatus === 'no_data' || clubMatchDataStatus === 'submitted')"
              @click="markFinal"
              :disabled="markingFinal"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
            >
              <span class="flex items-center gap-1.5">
                <IconLock class="w-3.5 h-3.5" />
                {{ markingFinal ? 'Marking…' : 'Mark Final' }}
              </span>
            </button>
            <button
              v-if="clubMatchDataStatus === 'final'"
              @click="markSubmitted"
              :disabled="markingSubmitted"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
            >
              <span class="flex items-center gap-1.5">
                <IconLock :open="true" class="w-3.5 h-3.5" />
                {{ markingSubmitted ? 'Marking…' : 'Mark Submitted' }}
              </span>
            </button>
            <span class="w-2 shrink-0" />
            <button
              v-if="starterCount > 0"
              @click="copyTeamToClipboard"
              title="Copy to Clipboard"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text-muted hover:text-text hover:bg-surface-hover transition-colors"
            >
              <span class="flex items-center gap-1.5">
                <IconCopy class="w-3.5 h-3.5" />
                {{ copyToClipboardLabel }}
              </span>
            </button>
          </template>

          <!-- Data status pill -->
          <div class="ml-auto flex items-center gap-2 shrink-0">
            <span v-if="markStatusError" class="text-xs text-red-400">{{ markStatusError }}</span>
            <span
              v-if="clubMatchDataStatus"
              class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
              :class="dataStatusClass[clubMatchDataStatus] ?? 'bg-surface-raised text-text-faint'"
            >{{ dataStatusLabel[clubMatchDataStatus] ?? clubMatchDataStatus }}</span>
          </div>
        </div>

        <!-- Substitution suggestions -->
        <div
          v-if="!readonly && !managing && suggestedSubstitutionHints.length > 0"
          class="mb-4 rounded-lg border border-sky-500/30 bg-sky-500/10 px-4 py-3"
        >
          <p class="text-xs font-semibold text-sky-400 mb-1">Improve your score:</p>
          <ul class="space-y-0.5">
            <li
              v-for="hint in suggestedSubstitutionHints"
              :key="hint"
              class="flex items-start gap-1.5 text-sm text-sky-300"
            >
              <span class="mt-px">·</span>
              <span>{{ hint }}</span>
            </li>
          </ul>
        </div>

        <!-- Skipped players notice (copy from round) -->
        <p v-if="skippedOnCopy.length > 0" class="mb-4 text-sm text-amber-400">
          Skipped (no longer in squad): {{ skippedOnCopy.join(', ') }}
        </p>

        <!-- Summary bar. In manage mode the projected pill sits at the right edge of
             the team column (mirroring the grid below); counts stay on the far right. -->
        <div class="mb-8 rounded-lg border border-border bg-surface-raised px-4 py-3">
          <div v-if="managing" class="grid grid-cols-1 sm:grid-cols-2 gap-8 items-center">
            <div class="flex items-center justify-between">
              <h2 class="text-sm font-semibold text-text-heading">Team</h2>
              <span
                v-if="projectedTotal > 0"
                class="rounded-md bg-sky-500/15 ring-1 ring-sky-400/40 px-2.5 py-0.5 text-sm tabular-nums text-sky-400"
                :title="`Estimated total from ${statSourceTitle.toLowerCase()}`"
              >Projected ~{{ projectedTotal }}</span>
            </div>
            <div class="flex items-center justify-end gap-3">
              <span class="text-sm tabular-nums text-text-muted">{{ starterCount }}/18 starters · {{ benchCount }}/4 bench</span>
              <span class="text-sm font-semibold tabular-nums">{{ grandTotal }}</span>
            </div>
          </div>
          <div v-else class="flex items-center justify-between">
            <h2 class="text-sm font-semibold text-text-heading">Team</h2>
            <div class="flex items-center gap-3">
              <span
                v-if="!readonly && projectedTotal > 0"
                class="rounded-md bg-sky-500/15 ring-1 ring-sky-400/40 px-2.5 py-0.5 text-sm tabular-nums text-sky-400"
                :title="`Estimated total from ${statSourceTitle.toLowerCase()}`"
              >Projected ~{{ projectedTotal }}</span>
              <PlayedCount :club-match="clubMatch" class="text-xs" />
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
                  :class="[slot.player
                    ? (subsMode && slot.player.aflStatus === 'dnp'
                      ? (subbedOutIds.has(slot.player.pmId ?? '') ? 'border-sky-500/40 bg-sky-500/5 cursor-pointer' : 'border-amber-600/30 bg-amber-500/5 cursor-pointer')
                      : 'border-border bg-surface-raised')
                    : 'border-dashed border-border-subtle bg-surface',
                    managing && slot.player ? 'cursor-grab active:cursor-grabbing' : '',
                    dragOverKey === `s:${pos.key}:${index}` ? '!border-sky-400 bg-sky-500/10' : '']"
                  :draggable="managing && !!slot.player"
                  @dragstart="onDragStart($event, { kind: 'starter', pos: pos.key, index })"
                  @dragend="onDragEnd"
                  @dragover="onDragOverTarget($event, `s:${pos.key}:${index}`, starterDropAction(pos.key, index))"
                  @dragleave="onDragLeave(`s:${pos.key}:${index}`)"
                  @drop.prevent="onDropTarget(starterDropAction(pos.key, index))"
                  @click="onStarterClick(slot.player)"
                >
                  <div v-if="slot.player" class="flex items-center gap-3">
                    <span v-if="pos.key === 'star'" class="text-yellow-400 text-xs">★</span>
                    <div v-if="managing">
                      <PlayerStatsCard :name="slot.player.name" :club="slot.player.club" :afl-status="slot.player.aflStatus" :afl-player-season-id="slot.player.aflPlayerSeasonId" :afl-round-id="bootstrapAflRoundId">
                        <div class="font-medium text-sm">{{ slot.player.name }}</div>
                        <div v-if="slot.player.club" class="text-xs text-text-muted">{{ slot.player.club }}</div>
                      </PlayerStatsCard>
                    </div>
                    <div v-else class="flex flex-col">
                      <PlayerStatsCard :name="slot.player.name" :club="slot.player.club" :afl-status="slot.player.aflStatus" :afl-player-season-id="slot.player.aflPlayerSeasonId" :afl-round-id="bootstrapAflRoundId">
                        <div class="flex items-baseline gap-2">
                          <component
                            :is="playerAflMatchRoute(slot.player) ? 'router-link' : 'span'"
                            :to="playerAflMatchRoute(slot.player) ?? undefined"
                            class="font-medium text-sm hover:text-active transition-colors"
                            :class="{ 'line-through': effectiveCovering(slot.player.pmId) }"
                          >{{ slot.player.name }}</component>
                          <span v-if="slot.player.club" class="text-xs text-text-muted" :class="{ 'line-through': effectiveCovering(slot.player.pmId) }">{{ slot.player.club }}</span>
                        </div>
                      </PlayerStatsCard>
                      <div v-if="effectiveCovering(slot.player.pmId)" class="text-sky-400">
                        <span class="text-xs mr-1">↑</span>
                        <span class="font-medium text-sm">{{ effectiveCovering(slot.player.pmId)!.name }}</span>
                        <span v-if="effectiveCovering(slot.player.pmId)!.club" class="ml-2 text-xs">{{ effectiveCovering(slot.player.pmId)!.club }}</span>
                      </div>
                    </div>
                  </div>
                  <span v-else class="text-text-faint text-sm">Empty slot</span>
                  <div v-if="slot.player && managing" class="relative flex items-center gap-2 shrink-0">
                    <span class="flex items-center gap-0.5" :title="statSourceTitle">
                      <span
                        v-for="col in statSummaryCols"
                        :key="col.key"
                        class="w-9 text-right text-xs tabular-nums whitespace-nowrap"
                        :class="col.key === 'star' ? 'text-yellow-400/70' : 'text-text-muted'"
                        :style="statHeat(slot.player, col.key)"
                      ><span v-if="statTrend(slot.player, col.key) === 'up'" class="text-[10px] text-green-400 mr-0.5" :title="trendUpTitle">↑</span><span v-else-if="statTrend(slot.player, col.key) === 'down'" class="text-[10px] text-red-400 mr-0.5" :title="trendDownTitle">↓</span><span
                        :class="col.key === pos.key ? 'rounded bg-sky-500/15 ring-1 ring-sky-400/40 px-1 py-0.5' : ''"
                      >{{ squadStat(slot.player, col.key) }}</span></span>
                    </span>
                    <button
                      aria-label="Player actions"
                      class="w-6 h-6 flex items-center justify-center rounded text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                      @click.stop="toggleMenu(`starter:${pos.key}:${index}`)"
                    >
                      <IconMenu class="w-3.5 h-3.5" />
                    </button>
                    <div
                      v-if="openMenuKey === `starter:${pos.key}:${index}`"
                      class="absolute right-0 top-full mt-1 z-50 w-44 rounded-lg border border-border bg-surface shadow-lg p-1"
                      @click.stop
                    >
                      <p class="px-2 py-1 text-[10px] uppercase tracking-wide text-text-faint">Move to</p>
                      <button
                        v-for="target in positions.filter(p => p.key !== pos.key)"
                        :key="target.key"
                        class="flex w-full items-center justify-between rounded px-2 py-1 text-xs transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                        :class="target.key === 'star' ? 'text-yellow-400' : 'text-text'"
                        :disabled="isPositionFull(target.key)"
                        @click.stop="moveToPosition(pos.key, index, target.key); closeMenu()"
                      >
                        <span>{{ target.label }}</span>
                        <span class="flex items-center gap-1.5">
                          <span class="flex items-center gap-0.5">
                            <span v-for="n in emptySlotCount(target.key)" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                          </span>
                          <span class="text-text-faint">{{ target.short }}</span>
                        </span>
                      </button>
                      <button
                        class="flex w-full items-center justify-between rounded px-2 py-1 text-xs text-text transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                        :disabled="benchDualFull"
                        @click.stop="moveStarterToBench(pos.key, index); closeMenu()"
                      >
                        <span>Bench</span>
                        <span class="flex items-center gap-1.5">
                          <span class="flex items-center gap-0.5">
                            <span v-for="n in benchEmptyCount" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                          </span>
                          <span class="text-text-faint">B</span>
                        </span>
                      </button>
                      <template v-if="canMoveUp(pos.key, index) || canMoveDown(pos.key, index)">
                        <div class="my-1 h-px bg-border-subtle" />
                        <button
                          v-if="canMoveUp(pos.key, index)"
                          class="flex w-full items-center rounded px-2 py-1 text-xs text-text transition-colors hover:bg-control-hover"
                          @click.stop="swapStarters(pos.key, index, index - 1); closeMenu()"
                        >Move up</button>
                        <button
                          v-if="canMoveDown(pos.key, index)"
                          class="flex w-full items-center rounded px-2 py-1 text-xs text-text transition-colors hover:bg-control-hover"
                          @click.stop="swapStarters(pos.key, index, index + 1); closeMenu()"
                        >Move down</button>
                      </template>
                      <div class="my-1 h-px bg-border-subtle" />
                      <button
                        class="flex w-full items-center gap-1.5 rounded px-2 py-1 text-xs text-red-400 transition-colors hover:bg-control-hover"
                        @click.stop="removeFromTeam(pos.key, index); closeMenu()"
                      >
                        <IconBin class="w-3.5 h-3.5" />
                        Remove
                      </button>
                    </div>
                  </div>
                  <div v-else-if="slot.player" class="flex items-center gap-2 shrink-0">
                    <span class="w-28 shrink-0"></span>
                    <span class="w-16 shrink-0">
                      <StatusBadge :status="playerStatus(slot.player)" />
                    </span>
                    <span class="w-28 text-right text-xs tabular-nums text-text-faint shrink-0">{{ playerShowScore(slot.player) ? (positionFormula(pos.key, effectivePlayerStats(slot.player)) ?? '') : '' }}</span>
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
                    recentlyClearedSlot === index ? '!border-orange-400' : '',
                    managing && slot.player ? 'cursor-grab active:cursor-grabbing' : '',
                    dragOverKey === `b:${index}` ? '!border-sky-400 bg-sky-500/10' : ''
                  ]"
                  :draggable="managing && !!slot.player"
                  @dragstart="onDragStart($event, { kind: 'bench', index })"
                  @dragend="onDragEnd"
                  @dragover="onDragOverTarget($event, `b:${index}`, benchDropAction(index))"
                  @dragleave="onDragLeave(`b:${index}`)"
                  @drop.prevent="onDropTarget(benchDropAction(index))"
                  @click="onBenchRowClick(slot)"
                >
                  <!-- Left: name -->
                  <div class="flex items-center gap-3 min-w-0">
                    <div v-if="slot.player">
                      <div class="flex items-baseline gap-2" :class="managing ? 'flex-col gap-0' : ''">
                        <span v-if="!managing && effectiveSubbedForStarter(slot.player.pmId)" class="text-xs mr-1 text-sky-400">↑</span>
                        <PlayerStatsCard v-if="!managing" :name="slot.player.name" :club="slot.player.club" :afl-status="slot.player.aflStatus" :afl-player-season-id="slot.player.aflPlayerSeasonId" :afl-round-id="bootstrapAflRoundId">
                          <component
                            :is="playerAflMatchRoute(slot.player) ? 'router-link' : 'span'"
                            :to="playerAflMatchRoute(slot.player) ?? undefined"
                            class="font-medium text-sm hover:text-active transition-colors"
                            :class="effectiveSubbedForStarter(slot.player.pmId) ? 'text-sky-400' : 'text-text-muted'"
                          >{{ slot.player.name }}</component>
                        </PlayerStatsCard>
                        <PlayerStatsCard v-else :name="slot.player.name" :club="slot.player.club" :afl-status="slot.player.aflStatus" :afl-player-season-id="slot.player.aflPlayerSeasonId" :afl-round-id="bootstrapAflRoundId">
                          <span class="font-medium text-sm text-text-muted">{{ slot.player.name }}</span>
                        </PlayerStatsCard>
                        <span
                          v-if="slot.player.club"
                          class="text-xs"
                          :class="!managing && effectiveSubbedForStarter(slot.player.pmId) ? 'text-sky-400' : 'text-text-muted'"
                        >{{ slot.player.club }}</span>
                      </div>
                    </div>
                    <span v-else class="text-text-faint text-sm">Empty slot</span>
                  </div>
                  <!-- Right: selectors + actions menu (manage) or read-only tags -->
                  <div class="relative flex items-center gap-2 ml-4 shrink-0">
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
                        aria-label="Bench player actions"
                        class="w-6 h-6 flex items-center justify-center rounded text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                        @click.stop="toggleMenu(`bench:${index}`)"
                      >
                        <IconMenu class="w-3.5 h-3.5" />
                      </button>
                      <div
                        v-if="openMenuKey === `bench:${index}`"
                        class="absolute right-0 top-full mt-1 z-50 w-44 rounded-lg border border-border bg-surface shadow-lg p-1"
                        @click.stop
                      >
                        <p class="px-2 py-1 text-[10px] uppercase tracking-wide text-text-faint">Move to</p>
                        <button
                          v-for="target in positions"
                          :key="target.key"
                          class="flex w-full items-center justify-between rounded px-2 py-1 text-xs transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                          :class="target.key === 'star' ? 'text-yellow-400' : 'text-text'"
                          :disabled="isPositionFull(target.key)"
                          @click.stop="moveBenchToStarter(index, target.key); closeMenu()"
                        >
                          <span>{{ target.label }}</span>
                          <span class="flex items-center gap-1.5">
                            <span class="flex items-center gap-0.5">
                              <span v-for="n in emptySlotCount(target.key)" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                            </span>
                            <span class="text-text-faint">{{ target.short }}</span>
                          </span>
                        </button>
                        <div class="my-1 h-px bg-border-subtle" />
                        <button
                          class="flex w-full items-center gap-1.5 rounded px-2 py-1 text-xs text-red-400 transition-colors hover:bg-control-hover"
                          @click.stop="removeBenchDual(index); closeMenu()"
                        >
                          <IconBin class="w-3.5 h-3.5" />
                          Remove
                        </button>
                      </div>
                    </template>
                    <template v-else-if="slot.player">
                      <div class="flex items-center gap-2 shrink-0">
                        <div class="w-28 flex items-center justify-end gap-1 shrink-0">
                          <span v-if="slot.positions[0]" :class="effectiveCoveredPosition(slot.player?.pmId) === slot.positions[0] ? 'text-xs rounded px-1.5 py-0.5 bg-sky-500/10 text-sky-400' : slot.positions[0] === 'star' ? 'text-xs bg-control rounded px-1.5 py-0.5 text-yellow-400' : 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted'">
                            {{ positionShort(slot.positions[0]) }}<template v-if="interchangePosition === slot.positions[0]"> · Int</template>
                          </span>
                          <span v-if="slot.positions[1]" :class="effectiveCoveredPosition(slot.player?.pmId) === slot.positions[1] ? 'text-xs rounded px-1.5 py-0.5 bg-sky-500/10 text-sky-400' : slot.positions[1] === 'star' ? 'text-xs bg-control rounded px-1.5 py-0.5 text-yellow-400' : 'text-xs bg-control rounded px-1.5 py-0.5 text-text-muted'">
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
            <div class="flex items-center justify-between mb-3">
              <h2 class="text-lg font-semibold text-text-heading">Squad ({{ availablePlayers.length }})</h2>
              <StatSourceToggle />
            </div>
            <!-- Sort headers: name on the left, trend explainer in the gap, stat columns on the right -->
            <div class="flex items-center justify-between px-4 mb-1 text-[10px] text-text-faint">
              <button
                class="transition-colors"
                :class="squadSortKey === 'name' ? 'text-sky-400 font-semibold' : 'hover:text-text'"
                title="Sort by name"
                @click="squadSortKey = 'name'"
              >Name</button>
              <div class="flex items-center gap-0.5" :title="statSourceTitle">
                <button
                  v-for="col in statSummaryCols"
                  :key="col.key"
                  class="w-9 text-right transition-colors"
                  :class="squadSortKey === col.key
                    ? 'text-sky-400 font-semibold'
                    : (col.key === 'star' ? 'text-yellow-400/70 hover:text-yellow-300' : 'hover:text-text')"
                  :title="`Sort by ${col.label}`"
                  @click="toggleSquadSort(col.key)"
                >{{ col.label }}</button>
                <span class="w-6" />
              </div>
            </div>
            <div
              class="space-y-1"
              :class="dragOverKey === 'squad' ? 'rounded-lg ring-1 ring-red-400/50' : ''"
              @dragover="onDragOverTarget($event, 'squad', squadDropAction())"
              @dragleave="onDragLeave('squad')"
              @drop.prevent="onDropTarget(squadDropAction())"
            >
              <div
                v-for="player in sortedAvailablePlayers"
                :key="player.id"
                class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-2 cursor-grab active:cursor-grabbing"
                draggable="true"
                @dragstart="onDragStart($event, { kind: 'squad', player })"
                @dragend="onDragEnd"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <PlayerStatsCard :name="player.name" :club="player.club" :afl-status="player.aflStatus" :afl-player-season-id="player.aflPlayerSeasonId" :afl-round-id="bootstrapAflRoundId">
                    <div>
                      <div class="font-medium text-sm">{{ player.name }}</div>
                      <div v-if="player.club" class="text-xs text-text-muted">{{ player.club }}</div>
                    </div>
                  </PlayerStatsCard>
                  <span v-if="playerShowScore(player)" class="text-sm tabular-nums text-text shrink-0">{{ player.score }}</span>
                </div>
                <div class="relative flex items-center gap-0.5 shrink-0">
                  <!-- Stats summary -->
                  <span
                    v-for="col in statSummaryCols"
                    :key="col.key"
                    class="w-9 text-right text-xs tabular-nums whitespace-nowrap"
                    :class="col.key === 'star' ? 'text-yellow-400/70' : 'text-text-muted'"
                    :style="statHeat(player, col.key)"
                  ><span v-if="statTrend(player, col.key) === 'up'" class="text-[10px] text-green-400 mr-0.5" :title="trendUpTitle">↑</span><span v-else-if="statTrend(player, col.key) === 'down'" class="text-[10px] text-red-400 mr-0.5" :title="trendDownTitle">↓</span>{{ squadStat(player, col.key) }}</span>
                  <!-- Popup fallback for drag and drop -->
                  <button
                    aria-label="Add to team"
                    class="w-6 h-6 flex items-center justify-center rounded text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                    @click.stop="toggleMenu(`squad:${player.id}`)"
                  >
                    <IconMenu class="w-3.5 h-3.5" />
                  </button>
                  <div
                    v-if="openMenuKey === `squad:${player.id}`"
                    class="absolute right-0 top-full mt-1 z-50 w-44 rounded-lg border border-border bg-surface shadow-lg p-1"
                    @click.stop
                  >
                    <p class="px-2 py-1 text-[10px] uppercase tracking-wide text-text-faint">Add to</p>
                    <button
                      v-for="pos in positions"
                      :key="pos.key"
                      class="flex w-full items-center justify-between rounded px-2 py-1 text-xs transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                      :class="pos.key === 'star' ? 'text-yellow-400' : 'text-text'"
                      :disabled="isPositionFull(pos.key)"
                      @click.stop="addToTeam(pos.key, player); closeMenu()"
                    >
                      <span>{{ pos.label }}</span>
                      <span class="flex items-center gap-1.5">
                        <span class="flex items-center gap-0.5">
                          <span v-for="n in emptySlotCount(pos.key)" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                        </span>
                        <span class="text-text-faint">{{ pos.short }}</span>
                      </span>
                    </button>
                    <button
                      class="flex w-full items-center justify-between rounded px-2 py-1 text-xs text-text transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                      :disabled="benchDualFull"
                      @click.stop="addBenchDual(player); closeMenu()"
                    >
                      <span>Bench</span>
                      <span class="flex items-center gap-1.5">
                        <span class="flex items-center gap-0.5">
                          <span v-for="n in benchEmptyCount" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                        </span>
                        <span class="text-text-faint">B</span>
                      </span>
                    </button>
                  </div>
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
                  v-for="player in sortedTradedPlayers"
                  :key="player.id"
                  class="flex items-center justify-between rounded-lg border border-border bg-surface-raised px-4 py-2 cursor-grab active:cursor-grabbing"
                  draggable="true"
                  @dragstart="onDragStart($event, { kind: 'squad', player })"
                  @dragend="onDragEnd"
                >
                  <div class="flex items-center gap-3 min-w-0 opacity-40">
                    <PlayerStatsCard :name="player.name" :club="player.club" :afl-status="player.aflStatus" :afl-player-season-id="player.aflPlayerSeasonId" :afl-round-id="bootstrapAflRoundId">
                      <div>
                        <div class="font-medium text-sm">{{ player.name }}</div>
                        <div v-if="player.club" class="text-xs text-text-muted">{{ player.club }}</div>
                      </div>
                    </PlayerStatsCard>
                    <span v-if="playerShowScore(player)" class="text-sm tabular-nums text-text shrink-0">{{ player.score }}</span>
                  </div>
                  <div class="relative flex items-center gap-0.5 shrink-0">
                    <span
                      v-for="col in statSummaryCols"
                      :key="col.key"
                      class="w-9 text-right text-xs tabular-nums whitespace-nowrap opacity-40"
                      :class="col.key === 'star' ? 'text-yellow-400/70' : 'text-text-muted'"
                      :style="statHeat(player, col.key)"
                    ><span v-if="statTrend(player, col.key) === 'up'" class="text-[10px] text-green-400 mr-0.5" :title="trendUpTitle">↑</span><span v-else-if="statTrend(player, col.key) === 'down'" class="text-[10px] text-red-400 mr-0.5" :title="trendDownTitle">↓</span>{{ squadStat(player, col.key) }}</span>
                    <button
                      aria-label="Add to team"
                      class="w-6 h-6 flex items-center justify-center rounded text-text-faint hover:bg-control-hover hover:text-text transition-colors"
                      @click.stop="toggleMenu(`squad:${player.id}`)"
                    >
                      <IconMenu class="w-3.5 h-3.5" />
                    </button>
                    <div
                      v-if="openMenuKey === `squad:${player.id}`"
                      class="absolute right-0 top-full mt-1 z-50 w-44 rounded-lg border border-border bg-surface shadow-lg p-1"
                      @click.stop
                    >
                      <p class="px-2 py-1 text-[10px] uppercase tracking-wide text-text-faint">Add to</p>
                      <button
                        v-for="pos in positions"
                        :key="pos.key"
                        class="flex w-full items-center justify-between rounded px-2 py-1 text-xs transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                        :class="pos.key === 'star' ? 'text-yellow-400' : 'text-text'"
                        :disabled="isPositionFull(pos.key)"
                        @click.stop="addToTeam(pos.key, player); closeMenu()"
                      >
                        <span>{{ pos.label }}</span>
                        <span class="flex items-center gap-1.5">
                          <span class="flex items-center gap-0.5">
                            <span v-for="n in emptySlotCount(pos.key)" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                          </span>
                          <span class="text-text-faint">{{ pos.short }}</span>
                        </span>
                      </button>
                      <button
                        class="flex w-full items-center justify-between rounded px-2 py-1 text-xs text-text transition-colors hover:bg-control-hover disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:bg-transparent"
                        :disabled="benchDualFull"
                        @click.stop="addBenchDual(player); closeMenu()"
                      >
                        <span>Bench</span>
                        <span class="flex items-center gap-1.5">
                          <span class="flex items-center gap-0.5">
                            <span v-for="n in benchEmptyCount" :key="n" class="w-1 h-1 rounded-full bg-text-faint" />
                          </span>
                          <span class="text-text-faint">B</span>
                        </span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </template>
      <p v-else class="text-text-faint">{{ readonly ? 'No data.' : 'No club selected. Choose a club in the nav bar.' }}</p>

    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { useQuery, useMutation, useLazyQuery } from '@vue/apollo-composable'
import { GET_FFL_ROUND, GET_FFL_SEASON_CLUBS, GET_FFL_CLUB_SEASON, GET_FFL_CLUB_MATCH, GET_FFL_CLUB_MATCH_TEAM } from '../api/queries'
import { SET_FFL_TEAM, DECLARE_FFL_SUBSTITUTIONS } from '../api/mutations'
import { MARK_FFL_TEAM_FINAL, MARK_FFL_TEAM_SUBMITTED } from '../../data-ops/api/mutations'
import Breadcrumb from '../components/Breadcrumb.vue'
import StatusBadge from '../components/StatusBadge.vue'
import PlayedCount from '../components/PlayedCount.vue'
import { clubLogoUrl } from '../utils/clubLogos'
import { clubAbbrev } from '../../afl/utils/clubAbbrev'
import { positionFormula } from '../utils/position'
import { isScoring } from '../utils/scoring'
import IconSquad from '../components/icons/IconSquad.vue'
import IconTeamBuilder from '../components/icons/IconTeamBuilder.vue'
import IconLock from '../components/icons/IconLock.vue'
import IconManage from '../components/icons/IconManage.vue'
import IconSubs from '../components/icons/IconSubs.vue'
import IconBin from '../components/icons/IconBin.vue'
import IconCopy from '../components/icons/IconCopy.vue'
import IconMenu from '../components/icons/IconMenu.vue'
import { heatStyle } from '@/utils/heatmap'
import { useTheme } from '@/composables/useTheme'
import { useNotFound } from '@/composables/useNotFound'
import NotFound from '@/components/NotFound.vue'
import { solveBestAssignment } from '../utils/bestTeam'
import PlayerStatsCard from '../components/PlayerStatsCard.vue'
import { useFflState } from '../composables/useFflState'
import { useStatSource } from '../composables/useStatSource'
import StatSourceToggle from '../components/StatSourceToggle.vue'
import { POSITION_MULTIPLIERS } from '../utils/position'
import { starScore, fmtStat, statCols, trendDir, LAST_N, TREND_PCT, type StatSummary } from '../utils/playerStats'

const props = defineProps<{ clubMatchId: string; readonly?: boolean }>()

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
  byeStats: { goals: number; kicks: number; handballs: number; marks: number; tackles: number; hitouts: number; games: number } | null
  aflPlayerSeasonId: string | null
  statsAll: StatSummary | null
  statsLastN: StatSummary | null
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

const bootstrapClubMatch = computed(() => clubMatchBootstrap.value?.fflClubMatch ?? null)
const notFound = useNotFound(bootstrapClubMatch, bootstrapLoading, ref(null))

const bootstrapRoundId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.roundId ?? '')
const bootstrapAflRoundId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.aflRoundId ?? null)
const bootstrapSeasonId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.seasonId ?? '')
const bootstrapClubSeasonId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.clubSeasonId ?? '')
const bootstrapClubMatchId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.id ?? '')
const bootstrapClubId = computed(() => clubMatchBootstrap.value?.fflClubMatch?.club?.id ?? '')
const isMyClub = computed(() => !!bootstrapClubId.value && bootstrapClubId.value === selectedClubId.value)

watch(bootstrapClubId, (id) => {
  if (id && !props.readonly) setClub(id)
}, { immediate: true })

const { result: roundResult, loading: roundLoading, error: roundError } = useQuery(
  GET_FFL_ROUND,
  () => ({ id: bootstrapRoundId.value, aflRoundId: bootstrapAflRoundId.value }),
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
    const m = currentMatch.value
    const cms = m.clubMatches ?? []
    let label: string
    if (m.matchStyle === 'bye') label = `${cms[0]?.club.name ?? '?'} (Bye)`
    else if (m.matchStyle === 'superbye') label = 'Superbye'
    else label = `${cms[0]?.club.name ?? '?'} v ${cms[1]?.club.name ?? '?'}`
    crumbs.push({ label, to: { name: 'ffl-match', params: { matchId: m.id } } })
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

type RoundMatchEntry = { clubMatches?: { id: string; clubSeasonId: string }[] | null }

function clubMatchIdForRound(roundId: string): string | null {
  const roundData = round.value?.season.rounds.find((r: { id: string }) => r.id === roundId)
  if (!roundData) return null
  const csId = bootstrapClubSeasonId.value
  for (const m of (roundData.matches ?? []) as RoundMatchEntry[]) {
    const cm = (m.clubMatches ?? []).find((cm) => cm.clubSeasonId === csId)
    if (cm) return cm.id
  }
  return null
}

const prevClubMatchId = computed(() => prevRound.value ? clubMatchIdForRound(prevRound.value.id) : null)
const nextClubMatchId = computed(() => nextRound.value ? clubMatchIdForRound(nextRound.value.id) : null)

const currentMatch = computed(() => {
  if (!round.value) return null
  return round.value.matches.find((m: { clubMatches?: { id: string }[] | null }) =>
    (m.clubMatches ?? []).some((cm) => cm.id === props.clubMatchId)
  ) ?? null
})

const clubMatch = computed(() => {
  if (!currentMatch.value) return null
  return (currentMatch.value.clubMatches ?? []).find((cm: { id: string }) => cm.id === props.clubMatchId) ?? null
})

const playerMatchBySeasonId = computed(() => {
  const map = new Map<string, { pmId: string; score: number | null; club: string | null; status: string | null; aflStatus: string | null; aflMatchId: string | null; goals: number | null; kicks: number | null; handballs: number | null; marks: number | null; tackles: number | null; hitouts: number | null; byeStats: { goals: number; kicks: number; handballs: number; marks: number; tackles: number; hitouts: number } | null; aflPlayerSeasonId: string | null }>()
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
      byeStats: pm.playerSeason?.aflPlayerSeason?.stats ?? null,
      aflPlayerSeasonId: pm.playerSeason?.aflPlayerSeason?.id ?? null,
    })
  }
  return map
})

const squad = computed<SquadPlayer[]>(() => {
  if (!clubSeasonData.value) return []
  return clubSeasonData.value.players.nodes.map((r: {
    id: string
    player: { aflPlayer: { name: string } }
    aflPlayerSeason?: { id?: string; clubSeason?: { club?: { name: string } | null } | null; statsAll?: StatSummary | null; statsLastN?: StatSummary | null } | null
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
      byeStats: pm?.byeStats ?? null,
      aflPlayerSeasonId: pm?.aflPlayerSeasonId ?? r.aflPlayerSeason?.id ?? null,
      statsAll: r.aflPlayerSeason?.statsAll ?? null,
      statsLastN: r.aflPlayerSeason?.statsLastN ?? null,
    }
  })
})

function playerStatus(player: SquadPlayer): string | null {
  if (player.status && player.status !== 'named') return player.status
  return player.aflStatus
}

function playerShowScore(player: SquadPlayer): boolean {
  return isScoring(player.aflStatus)
}

function benchPositionScore(player: SquadPlayer, pos: string): number | null {
  if (!playerShowScore(player)) return null
  const s = player.goals !== null ? player : player.byeStats
  if (!s) return null
  if (pos === 'star') {
    if (s.goals === null) return null
    return Math.floor(s.goals ?? 0) * 5 + Math.floor(s.kicks ?? 0) + Math.floor(s.handballs ?? 0) + Math.floor(s.marks ?? 0) * 2 + Math.floor(s.tackles ?? 0) * 4
  }
  const statMap: Record<string, number | null> = {
    goals: s.goals, kicks: s.kicks, handballs: s.handballs,
    marks: s.marks, tackles: s.tackles, hitouts: s.hitouts,
  }
  const stat = statMap[pos] ?? null
  if (stat === null) return null
  return Math.floor(stat) * (POSITION_MULTIPLIERS[pos] ?? 1)
}

function effectivePlayerStats(player: SquadPlayer) {
  if (player.goals !== null) return player
  const s = player.byeStats
  if (!s) return player
  return {
    goals: Math.floor(s.goals), kicks: Math.floor(s.kicks), handballs: Math.floor(s.handballs),
    marks: Math.floor(s.marks), tackles: Math.floor(s.tackles), hitouts: Math.floor(s.hitouts),
  }
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
      byeStats: pm.playerSeason?.aflPlayerSeason?.stats ?? null,
      aflPlayerSeasonId: pm.playerSeason?.aflPlayerSeason?.id ?? squadEntry?.aflPlayerSeasonId ?? null,
      statsAll: squadEntry?.statsAll ?? null,
      statsLastN: squadEntry?.statsLastN ?? null,
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

// ── Copy to clipboard ────────────────────────────────────────────────────────

const copyToClipboardLabel = ref('Copy')

const STATUS_TAG: Record<string, string> = {
  bye: 'Bye', dnp: 'DNP', subbed_out: 'Subbed', subbed_in: 'Sub In',
  interchanged_out: "IC'd", interchanged_in: 'IC In',
}

function hasSubScore(player: SquadPlayer): boolean {
  return player.status === 'subbed_out' || player.status === 'subbed_in' ||
    player.status === 'interchanged_out' || player.status === 'interchanged_in'
}

function formatTeamText(): string {
  const showTotals = grandTotal.value > 0
  const clubName = (selectedClubSeason.value?.club.name ?? 'TEAM').toUpperCase()
  const lines: string[] = [`${clubName} ${showTotals ? grandTotal.value: ''}`]
  lines.push('')

  for (const pos of positions) {
    const slots = teamSlots.value[pos.key].filter((s: Slot) => s.player)
    if (!slots.length) continue
    lines.push(pos.label.toUpperCase())
    for (const slot of slots) {
      const club = clubAbbrev(slot.player!.club)
      const tag = STATUS_TAG[playerStatus(slot.player!) ?? ''] ?? ''
      const showScore = playerShowScore(slot.player!) || hasSubScore(slot.player!)
      const score = showScore ? ` ${starterDisplayScore(slot.player!, pos.key)}` : ''
      lines.push(`${slot.player!.name}${club ? ` (${club})` : ''}${tag ? ` ${tag}` : ''}${score}`)
    }
    lines.push(String(showTotals ? positionTotal(pos.key) : ''))
  }

  const benchSlots = benchDualSlots.value.filter((s: BenchDualSlot) => s.player)
  if (benchSlots.length) {
    lines.push('BENCH')
    for (const slot of benchSlots) {
      const club = clubAbbrev(slot.player!.club)
      const tag = STATUS_TAG[playerStatus(slot.player!) ?? ''] ?? ''
      const isIc = isInterchangeSlot(slot)
      const posLabel = isIc ? '*' : slot.positions.filter((p): p is NonNullable<typeof p> => p != null).map(positionShort).join('/')
      const showScore = playerShowScore(slot.player!) || hasSubScore(slot.player!)
      const score = showScore ? ` ${benchScoreDisplay(slot)}` : ''
      lines.push(`${slot.player!.name}${club ? ` (${club})` : ''} ${posLabel}${tag ? ` ${tag}` : ''}${score}`)
    }
    if (interchangePosition.value) lines.push('Interchange = *')
  }

  return lines.join('\n')
}

async function copyTeamToClipboard() {
  await navigator.clipboard.writeText(formatTeamText())
  copyToClipboardLabel.value = 'Copied!'
  setTimeout(() => { copyToClipboardLabel.value = 'Copy' }, 2000)
}

// ── Suggested substitutions ──────────────────────────────────────────────────

const suggestedSubstitutionHints = computed(() => {
  const subs = clubMatch.value?.suggestedSubstitutions ?? []
  if (!subs.length) return []
  const pms = clubMatch.value?.playerMatches ?? []
  const find = (id: string) => pms.find((pm: { id: string }) => pm.id === id)?.player?.aflPlayer?.name ?? id
  return subs.map((s: { kind: string; replacedPmId: string; replacingPmId: string }) =>
    `${s.kind === 'interchange' ? 'Interchange' : 'Sub'}: ${find(s.replacingPmId)} in for ${find(s.replacedPmId)}`
  )
})

// ── Club match data status ───────────────────────────────────────────────────

const clubMatchDataStatus = computed(() => clubMatch.value?.dataStatus as string | undefined ?? null)

const clubMatchLocked = computed(() => clubMatchDataStatus.value === 'final')

const dataStatusLabel: Record<string, string> = { no_data: 'Not submitted', submitted: 'Submitted', final: 'Final' }

const markingFinal = ref(false)
const markingSubmitted = ref(false)
const markStatusError = ref('')

const roundRefetchVars = () => ({ query: GET_FFL_ROUND, variables: { id: bootstrapRoundId.value, aflRoundId: bootstrapAflRoundId.value } })

const { mutate: markFinalMutation } = useMutation(MARK_FFL_TEAM_FINAL, () => ({ refetchQueries: [roundRefetchVars()], awaitRefetchQueries: true }))
const { mutate: markSubmittedMutation } = useMutation(MARK_FFL_TEAM_SUBMITTED, () => ({ refetchQueries: [roundRefetchVars()], awaitRefetchQueries: true }))

async function markFinal() {
  if (!currentMatch.value) return
  markStatusError.value = ''
  markingFinal.value = true
  try {
    await markFinalMutation({ input: { clubMatchId: props.clubMatchId, matchId: currentMatch.value.id, roundId: bootstrapRoundId.value } })
  } catch (e: any) {
    markStatusError.value = e.message ?? 'Failed'
  } finally {
    markingFinal.value = false
  }
}

async function markSubmitted() {
  if (!currentMatch.value) return
  markStatusError.value = ''
  markingSubmitted.value = true
  try {
    await markSubmittedMutation({ input: { clubMatchId: props.clubMatchId, matchId: currentMatch.value.id, roundId: bootstrapRoundId.value } })
  } catch (e: any) {
    markStatusError.value = e.message ?? 'Failed'
  } finally {
    markingSubmitted.value = false
  }
}
const dataStatusClass: Record<string, string> = {
  no_data:   'bg-surface-raised text-text-faint',
  submitted: 'bg-yellow-500/15 text-yellow-500',
  final:     'bg-green-500/15 text-green-500',
}

// ── Copy from previous round ─────────────────────────────────────────────────

interface PrevPlayerMatch {
  playerSeasonId: string
  position: string
  backupPositions: string | null
  interchangePosition: string | null
  player: { aflPlayer: { name: string } }
}

const skippedOnCopy = ref<string[]>([])

const { load: loadPrevTeam, result: prevTeamResult, loading: prevTeamLoading } = useLazyQuery<{
  fflClubMatch: { playerMatches: PrevPlayerMatch[] } | null
}>(GET_FFL_CLUB_MATCH_TEAM)

async function copyPreviousTeam() {
  if (!prevClubMatchId.value) return
  skippedOnCopy.value = []
  await loadPrevTeam(GET_FFL_CLUB_MATCH_TEAM, { id: prevClubMatchId.value }, { fetchPolicy: 'network-only' })
  const pms = prevTeamResult.value?.fflClubMatch?.playerMatches ?? []
  resetTeamState()
  const squadById = new Map(squad.value.map(p => [p.id, p]))
  const skipped: string[] = []
  let dualIndex = 0
  for (const pm of pms) {
    const player = squadById.get(pm.playerSeasonId)
    if (!player || player.toRoundId) { skipped.push(pm.player.aflPlayer.name); continue }
    const isBench = pm.backupPositions != null || pm.interchangePosition != null
    if (!isBench) {
      const slot = teamSlots.value[pm.position as PositionKey]?.find((s: Slot) => !s.player)
      if (slot) slot.player = player
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
  skippedOnCopy.value = skipped
  markDirty()
}

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

function swapStarters(key: PositionKey, indexA: number, indexB: number) {
  const slots = teamSlots.value[key]
  const tmp = slots[indexA].player
  slots[indexA].player = slots[indexB].player
  slots[indexB].player = tmp
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

function moveBenchToStarter(index: number, pos: PositionKey) {
  const bSlot = benchDualSlots.value[index]
  if (!bSlot.player) return
  const slot = teamSlots.value[pos].find(s => !s.player)
  if (!slot) return
  slot.player = bSlot.player
  bSlot.player = null
  bSlot.positions = [null, null]
  markDirty()
}

function moveStarterToBench(pos: PositionKey, index: number) {
  const player = teamSlots.value[pos][index].player
  if (!player) return
  const slot = benchDualSlots.value.find(s => !s.player)
  if (!slot) return
  slot.player = player
  teamSlots.value[pos][index].player = null
  markDirty()
}

function canMoveUp(pos: PositionKey, index: number): boolean {
  return index > 0 && !!teamSlots.value[pos][index - 1].player
}

function canMoveDown(pos: PositionKey, index: number): boolean {
  const slots = teamSlots.value[pos]
  return index < slots.length - 1 && !!slots[index + 1].player
}

// ── Stats summary (manage mode) ──────────────────────────────────────────────

// Same column order as the squad page (statCols), with the star score appended.
const statSummaryCols = [...statCols, { key: 'star', label: '★' }] as const

const squadStatsById = computed(() => {
  const map = new Map<string, { statsAll: StatSummary | null; statsLastN: StatSummary | null }>()
  for (const p of squad.value) map.set(p.id, { statsAll: p.statsAll, statsLastN: p.statsLastN })
  return map
})

// Which stat set drives the summary columns, heatmap, sort, and projections —
// global toggle shared with the other stats pages.
const { statSource } = useStatSource()

const statSourceTitle = computed(() =>
  statSource.value === 'form' ? `Last ${LAST_N} form averages` : 'Season averages'
)

// Both stat sets for a player. Team-slot players are looked up by id because
// they may have been loaded before the club season stats arrived.
function playerStatSets(player: SquadPlayer): { form: StatSummary | null; season: StatSummary | null } {
  const s = squadStatsById.value.get(player.id)
  return {
    form: s?.statsLastN ?? player.statsLastN ?? null,
    season: s?.statsAll ?? player.statsAll ?? null,
  }
}

// The active stat set per the toggle; form falls back to season when a player
// has no recent games.
function formStats(player: SquadPlayer): StatSummary | null {
  const { form, season } = playerStatSets(player)
  return statSource.value === 'season' ? season : (form ?? season)
}

// Trend of form vs season for one stat (see trendDir in utils/playerStats).
const trendUpTitle = `Trending up: last ${LAST_N} average is at least ${TREND_PCT * 100}% above season average`
const trendDownTitle = `Trending down: last ${LAST_N} average is at least ${TREND_PCT * 100}% below season average`

function statTrend(player: SquadPlayer, key: typeof statSummaryCols[number]['key']): 'up' | 'down' | null {
  const { form, season } = playerStatSets(player)
  if (!form || !season) return null
  const f = key === 'star' ? starScore(form) : form[key]
  const s = key === 'star' ? starScore(season) : season[key]
  return trendDir(f, s)
}

function squadStat(player: SquadPlayer, key: typeof statSummaryCols[number]['key']): string {
  const s = formStats(player)
  if (!s) return '—'
  return fmtStat(key === 'star' ? starScore(s) : s[key])
}

// Per-column heatmap over the whole squad (same palette as the Squad page).
const { isDark } = useTheme()

const statColumnRange = computed(() => {
  const range = {} as Record<typeof statSummaryCols[number]['key'], { min: number; max: number }>
  for (const col of statSummaryCols) {
    const vals = squad.value
      .map(p => {
        const s = formStats(p)
        return s ? (col.key === 'star' ? starScore(s) : s[col.key]) : null
      })
      .filter((v): v is number => v != null)
    if (vals.length) range[col.key] = { min: Math.min(...vals), max: Math.max(...vals) }
  }
  return range
})

function statHeat(player: SquadPlayer, key: typeof statSummaryCols[number]['key']): Record<string, string> {
  const s = formStats(player)
  if (!s) return {}
  const r = statColumnRange.value[key]
  if (!r) return {}
  return heatStyle(key === 'star' ? starScore(s) : s[key], r.min, r.max, isDark.value)
}

// Squad panel sorting — click a stat header to sort descending, click again to
// return to last-name order. Defaults to star score.
const squadSortKey = ref<typeof statSummaryCols[number]['key'] | 'name'>('star')

function toggleSquadSort(key: typeof statSummaryCols[number]['key']) {
  squadSortKey.value = squadSortKey.value === key ? 'name' : key
}

function sortBySquadKey(players: SquadPlayer[]): SquadPlayer[] {
  const key = squadSortKey.value
  if (key === 'name') {
    return [...players].sort((a, b) => {
      const lastA = a.name.split(' ').pop()?.toLowerCase() ?? ''
      const lastB = b.name.split(' ').pop()?.toLowerCase() ?? ''
      return lastA.localeCompare(lastB)
    })
  }
  const val = (p: SquadPlayer): number => {
    const s = formStats(p)
    if (!s) return -1
    return key === 'star' ? starScore(s) : s[key]
  }
  return [...players].sort((a, b) => val(b) - val(a))
}

const sortedAvailablePlayers = computed(() => sortBySquadKey(availablePlayers.value))
const sortedTradedPlayers = computed(() => sortBySquadKey(tradedPlayers.value))

// Projected points for a starter slot: form average for the position stat × multiplier.
function projectedScore(player: SquadPlayer, pos: PositionKey): number | null {
  const s = formStats(player)
  if (!s) return null
  return pos === 'star' ? starScore(s) : s[pos] * (POSITION_MULTIPLIERS[pos] ?? 1)
}

// Estimated team total: form-based projections summed across all filled starter slots.
const projectedTotal = computed(() => {
  let total = 0
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (!slot.player) continue
      total += projectedScore(slot.player, pos.key) ?? 0
    }
  }
  return Math.round(total)
})

// ── Best team ────────────────────────────────────────────────────────────────

// Fills the starter slots with the highest-scoring assignment of squad players
// to positions (Hungarian algorithm over projected scores from the active stat
// source). Bench slots whose player becomes a starter are cleared; the rest of
// the bench is left alone. Local state only — Save still applies it.
function applyBestTeam() {
  const pool = squad.value.filter(p => !p.toRoundId && formStats(p))
  if (!pool.length) return

  const slotPositions: PositionKey[] = []
  for (const pos of positions) {
    for (let i = 0; i < pos.count; i++) slotPositions.push(pos.key)
  }

  const values = slotPositions.map(pos => pool.map(p => projectedScore(p, pos) ?? 0))
  const assignment = solveBestAssignment(values)

  const byPos = new Map<PositionKey, SquadPlayer[]>(positions.map(p => [p.key, []]))
  assignment.forEach((playerIdx, slotIdx) => {
    if (playerIdx >= 0) byPos.get(slotPositions[slotIdx])!.push(pool[playerIdx])
  })

  const starterIds = new Set<string>()
  for (const pos of positions) {
    const chosen = byPos.get(pos.key)!
      .sort((a, b) => (projectedScore(b, pos.key) ?? 0) - (projectedScore(a, pos.key) ?? 0))
    teamSlots.value[pos.key] = Array.from({ length: pos.count }, (_, i) => ({ player: chosen[i] ?? null }))
    for (const p of chosen) starterIds.add(p.id)
  }

  for (const bSlot of benchDualSlots.value) {
    if (bSlot.player && starterIds.has(bSlot.player.id)) {
      bSlot.player = null
      bSlot.positions = [null, null]
    }
  }
  markDirty()
}

// ── Popup menus (fallback for drag and drop) ─────────────────────────────────

// Remaining open slots per target — rendered as dots in the popup menus.
function emptySlotCount(key: PositionKey): number {
  return teamSlots.value[key].filter(s => !s.player).length
}

const benchEmptyCount = computed(() => benchDualSlots.value.filter(s => !s.player).length)

const openMenuKey = ref<string | null>(null)

function toggleMenu(key: string) {
  openMenuKey.value = openMenuKey.value === key ? null : key
}

function closeMenu() {
  openMenuKey.value = null
}

// ── Drag and drop ────────────────────────────────────────────────────────────

type DragSource =
  | { kind: 'squad'; player: SquadPlayer }
  | { kind: 'starter'; pos: PositionKey; index: number }
  | { kind: 'bench'; index: number }

const dragSource = ref<DragSource | null>(null)
const dragOverKey = ref<string | null>(null)

function onDragStart(ev: DragEvent, src: DragSource) {
  if (!managing.value) {
    ev.preventDefault()
    return
  }
  dragSource.value = src
  closeMenu()
  if (ev.dataTransfer) {
    ev.dataTransfer.setData('text/plain', '')
    ev.dataTransfer.effectAllowed = 'move'
  }
}

function onDragEnd() {
  dragSource.value = null
  dragOverKey.value = null
}

function onDragLeave(key: string) {
  if (dragOverKey.value === key) dragOverKey.value = null
}

// Returns the action a drop on the given starter slot would perform, or null if invalid.
function starterDropAction(pos: PositionKey, index: number): (() => void) | null {
  const src = dragSource.value
  if (!src) return null
  const slot = teamSlots.value[pos][index]
  if (src.kind === 'squad') {
    const player = src.player
    if (!slot.player) return () => { slot.player = player; markDirty() }
    if (!isPositionFull(pos)) return () => addToTeam(pos, player)
    return null
  }
  if (src.kind === 'starter') {
    const { pos: fromPos, index: fromIndex } = src
    if (fromPos === pos) {
      if (fromIndex === index || !slot.player) return null
      return () => swapStarters(pos, fromIndex, index)
    }
    // Cross-position: fill an empty slot, or swap occupants.
    const fromSlot = teamSlots.value[fromPos][fromIndex]
    if (!slot.player) return () => { slot.player = fromSlot.player; fromSlot.player = null; markDirty() }
    return () => {
      const tmp = fromSlot.player
      fromSlot.player = slot.player
      slot.player = tmp
      markDirty()
    }
  }
  // Bench → starter slot.
  const bSlot = benchDualSlots.value[src.index]
  if (!bSlot.player) return null
  const place = (target: Slot) => {
    target.player = bSlot.player
    bSlot.player = null
    bSlot.positions = [null, null]
    markDirty()
  }
  if (!slot.player) return () => place(slot)
  const empty = teamSlots.value[pos].find(s => !s.player)
  if (empty) return () => place(empty)
  return null
}

// Returns the action a drop on the given bench slot would perform, or null if invalid.
function benchDropAction(index: number): (() => void) | null {
  const src = dragSource.value
  if (!src) return null
  const slot = benchDualSlots.value[index]
  if (src.kind === 'squad') {
    const player = src.player
    if (!slot.player) return () => { slot.player = player; markDirty() }
    if (!benchDualFull.value) return () => addBenchDual(player)
    return null
  }
  if (src.kind === 'starter') {
    const fromSlot = teamSlots.value[src.pos][src.index]
    if (!fromSlot.player) return null
    if (!slot.player) return () => { slot.player = fromSlot.player; fromSlot.player = null; markDirty() }
    if (!benchDualFull.value) return () => { addBenchDual(fromSlot.player!); fromSlot.player = null }
    return null
  }
  // Bench → bench: reorder (player and backup positions travel together).
  if (src.index === index) return null
  const other = benchDualSlots.value[src.index]
  return () => {
    const tmpPlayer = other.player
    const tmpPositions = other.positions
    other.player = slot.player
    other.positions = slot.positions
    slot.player = tmpPlayer
    slot.positions = tmpPositions
    markDirty()
  }
}

// Dropping a team/bench player back on the squad list removes them from the team.
function squadDropAction(): (() => void) | null {
  const src = dragSource.value
  if (!src) return null
  if (src.kind === 'starter') return () => removeFromTeam(src.pos, src.index)
  if (src.kind === 'bench') return () => removeBenchDual(src.index)
  return null
}

function onDragOverTarget(ev: DragEvent, key: string, action: (() => void) | null) {
  if (!action) return
  ev.preventDefault()
  if (ev.dataTransfer) ev.dataTransfer.dropEffect = 'move'
  dragOverKey.value = key
}

function onDropTarget(action: (() => void) | null) {
  if (action) action()
  onDragEnd()
}

// ── Subs mode ────────────────────────────────────────────────────────────────

// True when the AFL match is underway or complete (excludes 'named' — pre-match only).
const aflMatchStarted = computed(() => {
  const pms = clubMatch.value?.playerMatches ?? []
  return pms.some((pm: { aflStatus: string | null }) =>
    pm.aflStatus === 'playing' || pm.aflStatus === 'played' || pm.aflStatus === 'dnp' || pm.aflStatus === 'bye'
  )
})

// Subs UI state.
const subbedOutIds = ref<Set<string>>(new Set())
const interchangeApplied = ref(false)
const subsSaving = ref(false)
const subsMessage = ref('')
const subsError = ref(false)

function initSubsState() {
  const pms = clubMatch.value?.playerMatches ?? []
  // Pre-populate from stored TM decisions.
  subbedOutIds.value = new Set(
    pms
      .filter((pm: { status: string | null }) => pm.status === 'subbed_out')
      .map((pm: { id: string }) => pm.id)
  )
  const savedApplied = pms.some((pm: { status: string | null }) => pm.status === 'interchanged_out')
  interchangeApplied.value = savedApplied
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
  subsError.value = false
}

function toggleSub(pmId: string) {
  const next = new Set(subbedOutIds.value)
  if (next.has(pmId)) {
    next.delete(pmId)
  } else {
    next.add(pmId)
  }
  subbedOutIds.value = next
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

// Covering map based on saved server state (status === 'subbed_out') — used in normal (non-subs) mode.
const savedSubsMap = computed(() => {
  const map = new Map<string, SquadPlayer>()
  for (const pos of positions) {
    for (const slot of teamSlots.value[pos.key]) {
      if (!slot.player?.pmId || slot.player.status !== 'subbed_out') continue
      for (const bSlot of benchDualSlots.value) {
        if (!bSlot.player || bSlot.player.status !== 'subbed_in') continue
        if ((bSlot.positions as (string | null)[]).includes(pos.key)) {
          map.set(slot.player.pmId, bSlot.player)
          break
        }
      }
    }
  }
  return map
})

// The bench slot that holds the interchange player.
const interchangeBenchSlot = computed(() =>
  benchDualSlots.value.find(s => isInterchangeSlot(s)) ?? null
)

// Displaced starter in subs mode: lowest-scoring active starter at interchangePosition.
const interchangeDisplacedStarterSubsMode = computed((): SquadPlayer | null => {
  if (!interchangeApplied.value || !interchangePosition.value) return null
  const posSlots = teamSlots.value[interchangePosition.value as PositionKey]
  if (!posSlots) return null
  const active = posSlots.filter(s => s.player && !subbedOutIds.value.has(s.player.pmId ?? ''))
  if (!active.length) return null
  return active.reduce((low, s) => {
    const ls = playerShowScore(low.player!) ? (low.player!.score ?? 0) : 0
    const ss = playerShowScore(s.player!) ? (s.player!.score ?? 0) : 0
    return ss < ls ? s : low
  }, active[0]).player ?? null
})

// Displaced starter in normal mode: starter with status='interchanged_out' at interchangePosition.
const interchangeDisplacedStarterNormal = computed((): SquadPlayer | null => {
  if (!interchangePosition.value) return null
  const posSlots = teamSlots.value[interchangePosition.value as PositionKey]
  return posSlots?.find(s => s.player?.status === 'interchanged_out')?.player ?? null
})

// Returns the covering bench player for a starter — covers both subs and interchange.
function effectiveCovering(pmId: string | null): SquadPlayer | null {
  if (!pmId) return null
  if (subsMode.value) {
    const sub = subsMapping.value.get(pmId)
    if (sub) return sub
    if (interchangeApplied.value && interchangeDisplacedStarterSubsMode.value?.pmId === pmId)
      return interchangeBenchSlot.value?.player ?? null
    return null
  } else {
    const sub = savedSubsMap.value.get(pmId)
    if (sub) return sub
    if (interchangeDisplacedStarterNormal.value?.pmId === pmId)
      return interchangeBenchSlot.value?.player ?? null
    return null
  }
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
      if (!slot.player?.pmId || slot.player.status !== 'subbed_out') continue
      const cp = savedSubsMap.value.get(slot.player.pmId)
      if (cp?.pmId) map.set(cp.pmId, slot.player)
    }
  }
  return map
})

// Returns the starter a bench player is covering (sub or interchange) — covers both modes.
function effectiveSubbedForStarter(benchPmId: string | null): SquadPlayer | null {
  if (!benchPmId) return null
  if (subsMode.value) {
    const sub = subsStarterMap.value.get(benchPmId)
    if (sub) return sub
    if (interchangeApplied.value && interchangeBenchSlot.value?.player?.pmId === benchPmId)
      return interchangeDisplacedStarterSubsMode.value ?? null
    return null
  } else {
    const sub = savedSubsStarterMap.value.get(benchPmId)
    if (sub) return sub
    if (interchangeBenchSlot.value?.player?.pmId === benchPmId)
      return interchangeDisplacedStarterNormal.value ?? null
    return null
  }
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
  if (subsMode.value) {
    // Regular sub
    const subCp = subsMapping.value.get(player.pmId ?? '')
    if (subCp && subbedOutIds.value.has(player.pmId ?? '')) return benchPositionScore(subCp, posKey) ?? ''
    // Interchange displaced starter
    if (interchangeApplied.value && interchangeDisplacedStarterSubsMode.value?.pmId === player.pmId) {
      const cp = interchangeBenchSlot.value?.player
      if (cp) return benchPositionScore(cp, posKey) ?? ''
    }
  } else {
    // Saved sub
    const subCp = savedSubsMap.value.get(player.pmId ?? '')
    if (subCp) return benchPositionScore(subCp, posKey) ?? ''
    // Saved interchange displaced starter
    if (interchangeDisplacedStarterNormal.value?.pmId === player.pmId) {
      const cp = interchangeBenchSlot.value?.player
      if (cp) return benchPositionScore(cp, posKey) ?? ''
    }
  }
  return playerShowScore(player) ? (player.score ?? '') : ''
}

const { mutate: declareSubs } = useMutation(DECLARE_FFL_SUBSTITUTIONS, () => ({
  refetchQueries: [{ query: GET_FFL_ROUND, variables: { id: bootstrapRoundId.value, aflRoundId: bootstrapAflRoundId.value } }],
  awaitRefetchQueries: true,
}))

async function onSaveSubs() {
  if (!clubMatch.value) return
  subsSaving.value = true
  subsMessage.value = ''
  subsError.value = false
  try {
    const subs = Array.from(subsMapping.value.entries()).map(([replacedPmId, benchPlayer]) => ({
      replacedPmId,
      replacingPmId: benchPlayer.pmId!,
    }))

    let interchange: { replacedPmId: string; replacingPmId: string } | null = null
    if (interchangeApplied.value) {
      const displacedPmId = interchangeDisplacedStarterSubsMode.value?.pmId
      const icPmId = interchangeBenchSlot.value?.player?.pmId
      if (displacedPmId && icPmId) {
        interchange = { replacedPmId: displacedPmId, replacingPmId: icPmId }
      }
    }

    await declareSubs({
      input: {
        clubMatchId: clubMatch.value.id,
        subs,
        interchange,
      },
    })
    // nextTick lets Vue flush the Apollo cache → reactive update before we read
    // clubMatch.value, so loadTeamFromMatch sees the new statuses.
    await nextTick()
    if (clubMatch.value) {
      loadTeamFromMatch(clubMatch.value)
      initializedMatchId.value = clubMatch.value.id
    }
    exitSubsMode()
  } catch (e: unknown) {
    const gqlErr = (e as { graphQLErrors?: { message: string }[] })?.graphQLErrors?.[0]
    subsMessage.value = gqlErr?.message ?? 'Failed to save substitutions'
    subsError.value = true
  } finally {
    subsSaving.value = false
  }
}

// ── Submit ────────────────────────────────────────────────────────────────────

const { mutate: setTeam } = useMutation(SET_FFL_TEAM, () => ({
  refetchQueries: [{ query: GET_FFL_ROUND, variables: { id: bootstrapRoundId.value, aflRoundId: bootstrapAflRoundId.value } }],
  awaitRefetchQueries: true,
}))
const submitting = ref(false)
const submitMessage = ref('')

async function onSaveTeam() {
  const ok = await submitTeam()
  if (ok) { managing.value = false; skippedOnCopy.value = [] }
}

function cancelManage() {
  if (clubMatch.value) loadTeamFromMatch(clubMatch.value)
  managing.value = false
  skippedOnCopy.value = []
}

async function submitTeam(): Promise<boolean> {
  if (!clubMatch.value) return false
  submitting.value = true
  submitMessage.value = ''

  const players: {
    playerSeasonId: string
    position: string
    backupPositions?: string
    interchangePosition?: string
    displayOrder: number
  }[] = []

  // Starters — displayOrder is 1-based within each position group, in slot order.
  for (const pos of positions) {
    let posOrder = 0
    for (const slot of teamSlots.value[pos.key]) {
      if (slot.player) {
        posOrder++
        players.push({ playerSeasonId: slot.player.id, position: pos.key, displayOrder: posOrder })
      }
    }
  }

  // Bench slots — displayOrder is 1-based across all bench slots.
  let benchOrder = 0
  for (const slot of benchDualSlots.value) {
    if (!slot.player) continue
    benchOrder++
    const [p1, p2] = slot.positions
    const isStar = p1 === 'star'
    const bp = isStar ? 'star' : [p1, p2].filter(Boolean).join(',')
    const entry: (typeof players)[number] = {
      playerSeasonId: slot.player.id,
      position: p1 ?? p2 ?? 'goals',
      backupPositions: bp || undefined,
      displayOrder: benchOrder,
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
    return true
  } catch (e: unknown) {
    const gqlErr = (e as { graphQLErrors?: { message: string; extensions?: Record<string, string> }[] })?.graphQLErrors?.[0]
    if (gqlErr?.extensions?.code === 'BYE_INELIGIBLE') {
      const psId = gqlErr.extensions.playerSeasonId
      const player = squad.value.find(p => p.id === psId)
      submitMessage.value = `${player?.name ?? 'A player'} is on a bye but didn't play last round — remove them to save`
    } else {
      submitMessage.value = gqlErr?.message ?? 'Failed to save team'
    }
    return false
  } finally {
    submitting.value = false
  }
}
</script>
