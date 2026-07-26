<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold">Data Ops</h1>
    </div>

    <!-- Tab navigation -->
    <div class="flex gap-1 mb-6 border-b border-border">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        class="px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px"
        :class="activeTab === tab.id
          ? 'border-active text-active'
          : 'border-transparent text-text-muted hover:text-text'"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: AFL Stats Import                       -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'afl-stats'">
      <div v-if="loadingLiveRound" class="text-text-faint">Loading…</div>
      <div v-else-if="liveRoundError" class="text-red-400">{{ liveRoundError.message }}</div>
      <template v-else>
        <!-- Round selector -->
        <AflRoundNav
          class="mb-5"
          :rounds="aflRounds"
          :live-round-id="aflLiveRoundId"
          :active-id="selectedAflRoundId"
          :to-round="(r) => ({ name: 'ffl-data-ops', query: { tab: 'afl-stats', round: r.id } })"
        />

        <!-- Match list -->
        <div v-if="loadingRoundStats" class="text-text-faint text-sm">Loading matches…</div>
        <div v-else-if="!aflMatches.length" class="text-text-faint text-sm">No matches in this round.</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-border">
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Match</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Status</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Score</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Players</th>
                <th class="pb-2 text-right text-xs font-medium text-text-faint"></th>
              </tr>
            </thead>
            <tbody>
              <template v-for="match in aflMatches" :key="match.id">
                <tr class="border-b border-border" :class="{ 'border-b-0': scrapeResult[match.id] || scrapeError[match.id] }">
                  <td class="py-3 pr-4 text-sm font-semibold whitespace-nowrap">
                    <router-link
                      v-if="match.id"
                      :to="{ name: 'afl-match', params: { matchId: match.id } }"
                                            class="inline-flex items-center gap-1.5 text-text hover:text-active transition-colors"
                    >
                      <img v-if="match.homeClubMatch?.club.name" :src="clubLogoUrl(match.homeClubMatch.club.name)" class="w-4 h-4 object-contain" />
                      {{ match.homeClubMatch?.club.name ?? '—' }}
                      <span class="font-normal text-text-faint text-xs mx-1">vs</span>
                      <img v-if="match.awayClubMatch?.club.name" :src="clubLogoUrl(match.awayClubMatch.club.name)" class="w-4 h-4 object-contain" />
                      {{ match.awayClubMatch?.club.name ?? '—' }}
                                          </router-link>
                    <template v-else>
                      <span class="inline-flex items-center gap-1.5">
                        <img v-if="match.homeClubMatch?.club.name" :src="clubLogoUrl(match.homeClubMatch.club.name)" class="w-4 h-4 object-contain" />
                        {{ match.homeClubMatch?.club.name ?? '—' }}
                        <span class="font-normal text-text-faint text-xs mx-1">vs</span>
                        <img v-if="match.awayClubMatch?.club.name" :src="clubLogoUrl(match.awayClubMatch.club.name)" class="w-4 h-4 object-contain" />
                        {{ match.awayClubMatch?.club.name ?? '—' }}
                      </span>
                    </template>
                  </td>
                  <td class="py-3 pr-4 whitespace-nowrap">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="statusBadge(match.dataStatus)"
                    >{{ statusLabel(match.dataStatus) }}</span>
                  </td>
                  <td class="py-3 pr-4 text-xs text-text-faint whitespace-nowrap tabular-nums">
                    <template v-if="match.homeClubMatch?.score != null && match.awayClubMatch?.score != null && match.dataStatus !== 'no_data'">
                      {{ match.homeClubMatch.score }}
                      <span class="text-text-faint mx-1">vs</span>
                      {{ match.awayClubMatch.score }}
                    </template>
                    <template v-else>—</template>
                  </td>
                  <td class="py-3 pr-4 text-xs text-text-faint whitespace-nowrap">
                    <template v-if="match.dataStatus === 'partial' || match.dataStatus === 'final'">
                      {{ match.homeClubMatch?.playerMatches?.length ?? 0 }} - {{ match.awayClubMatch?.playerMatches?.length ?? 0 }}
                    </template>
                    <template v-else>—</template>
                  </td>
                  <td class="py-3 text-right whitespace-nowrap">
                    <div class="flex items-center justify-end gap-2">
                      <button
                        v-if="match.dataStatus !== 'no_data'"
                        @click="toggleFinal(match)"
                        :disabled="togglingFinal[match.id]"
                        :title="match.dataStatus === 'final' ? 'Reverts match statuss from .' : 'Marks match stats as complete. Triggers AFL ladder recalculation.'"
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ match.dataStatus === 'final' ? 'Mark Partial' : 'Mark Final' }}</button>
                      <button
                        @click="scrape(match)"
                        :disabled="scraping[match.id]"
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ scraping[match.id] ? 'Getting Stats…' : 'Get Stats' }}</button>
                    </div>
                  </td>
                </tr>
                <!-- Feedback / unmatched review sub-row -->
                <tr v-if="scrapeResult[match.id] || scrapeError[match.id]" class="border-b border-border">
                  <td colspan="5" class="pb-3 pt-0">
                    <span v-if="scrapeError[match.id]" class="text-xs text-red-400">{{ scrapeError[match.id] }}</span>
                    <template v-else-if="scrapeResult[match.id]">
                      <div class="text-xs mb-1">
                        <span class="text-green-500 font-medium">Imported</span>
                        <span class="text-text-faint ml-1">
                          · {{ scrapeResult[match.id].homePlayerCount + scrapeResult[match.id].awayPlayerCount }} players
                        </span>
                        <span v-if="scrapeResult[match.id].unmatchedPlayers.length > 0" class="text-yellow-500 ml-1">
                          · {{ scrapeResult[match.id].unmatchedPlayers.length }} unlinked — review below
                        </span>
                      </div>
                      <!-- Unmatched players -->
                      <div v-if="scrapeResult[match.id].unmatchedPlayers.length > 0" class="mt-2 border border-border rounded-lg overflow-hidden">
                        <table class="w-full">
                          <thead>
                            <tr class="border-b border-border bg-surface-raised">
                              <th class="px-3 py-2 text-left text-xs font-medium text-text-faint">Unlinked player</th>
                              <th class="px-3 py-2 text-left text-xs font-medium text-text-faint">Stats</th>
                              <th class="px-3 py-2 text-right text-xs font-medium text-text-faint"></th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr
                              v-for="(up, ui) in scrapeResult[match.id].unmatchedPlayers"
                              :key="ui"
                              class="border-b border-border last:border-0"
                              :class="resolvedUnmatched[match.id]?.[ui] ? 'bg-green-500/5' : ''"
                            >
                              <td class="px-3 py-2 whitespace-nowrap">
                                <span class="text-sm font-medium">{{ up.parsedName }}</span>
                                <span class="text-xs text-text-faint ml-2">{{ clubMatchClubMap[up.clubMatchId] }}</span>
                              </td>
                              <td class="px-3 py-2 text-xs text-text-faint whitespace-nowrap tabular-nums">
                                {{ up.goals }}g {{ up.kicks }}k {{ up.handballs }}hb {{ up.marks }}m {{ up.tackles }}t {{ up.hitouts }}ho
                              </td>
                              <td class="px-3 py-2 text-right">
                                <span v-if="resolvedUnmatched[match.id]?.[ui]" class="text-green-500 text-xs font-medium">✓ Linked</span>
                                <button
                                  v-else
                                  @click="openResolveModal(match, ui, up)"
                                  class="rounded border border-border px-2 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors"
                                >Link Player</button>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </template>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- Bye clubs -->
        <div v-if="aflByes.length" class="mt-4 flex items-center gap-2 flex-wrap">
          <span class="text-sm text-text-faint">Byes</span>
          <template v-for="bye in aflByes" :key="bye.id">
            <span class="text-sm text-border">·</span>
            <span class="flex items-center gap-1.5 opacity-70">
              <img :src="clubLogoUrl(bye.club.name)" :alt="bye.club.name" class="w-5 h-5 object-contain" />
              <span class="text-sm text-text-faint">{{ bye.club.name }}</span>
            </span>
          </template>
        </div>
      </template>
    </div>

    <!-- Player search modal -->
    <PlayerSearchModal
      v-if="resolveModalPlayer"
      :show="!!resolveModalPlayer"
      :player="resolveModalPlayer"
      @close="resolveModalPlayer = null"
      @resolved="onPlayerResolved"
    />

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: FFL Teams                              -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'team-submission'">
      <div v-if="loadingSeasonData" class="text-text-faint">Loading...</div>
      <div v-else-if="seasonError" class="text-red-400">{{ seasonError.message }}</div>
      <template v-else-if="season">

        <!-- Round selector -->
        <FflRoundNav
          class="mb-5"
          :rounds="season.rounds"
          :live-round-id="liveRoundId"
          :active-id="selectedRoundId"
          :to-round="(r) => ({ name: 'ffl-data-ops', query: { tab: 'team-submission', round: r.id } })"
        />

        <!-- Club list for selected round -->
        <div v-if="!selectedRound" class="text-text-faint text-sm">Select a round.</div>
        <div v-else-if="!fflClubRows.length" class="text-text-faint text-sm">No matches in this round.</div>
        <div v-else class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-border">
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Club</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint w-px whitespace-nowrap">Status</th>
                <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint w-px whitespace-nowrap">Score</th>
                <th class="pb-2 text-right text-xs font-medium text-text-faint w-[32rem]"></th>
              </tr>
            </thead>
            <tbody>
              <template v-for="row in fflClubRows" :key="row.clubMatchId">
                <!-- Club row -->
                <tr
                  class="border-b border-border"
                  :class="{ 'border-b-0': activeImportClubMatchId === row.clubMatchId }"
                >
                  <td class="py-3 pr-4">
                    <router-link
                      :to="{ name: 'ffl-club-match-edit', params: { clubMatchId: row.clubMatchId } }"
                                            class="inline-flex items-center gap-1.5 text-sm font-semibold text-text hover:text-active transition-colors"
                    >
                      <img v-if="fflClubLogoUrl(row.clubName)" :src="fflClubLogoUrl(row.clubName)" :alt="row.clubName" class="w-5 h-5 object-contain" />
                      {{ row.clubName }}                    </router-link>
                    <span
                      v-if="row.matchStyle === 'bye' || row.matchStyle === 'superbye'"
                      class="ml-2 inline-flex items-center rounded-full bg-surface-raised px-2 py-0.5 text-xs text-text-faint"
                    >{{ row.matchStyle }}</span>
                  </td>
                  <td class="py-3 pr-4 whitespace-nowrap">
                    <span
                      class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                      :class="fflStatusBadge(row.dataStatus)"
                    >{{ fflStatusLabel(row.dataStatus) }}</span>
                  </td>
                  <td class="py-3 pr-4 text-xs text-text-faint tabular-nums">
                    {{ row.dataStatus !== 'no_data' ? row.score : '—' }}
                  </td>
                  <td class="py-3 text-right whitespace-nowrap">
                    <div v-if="activeImportClubMatchId !== row.clubMatchId" class="flex items-center justify-end gap-2">
                      <span v-if="importedResult[row.clubMatchId]" class="text-xs text-green-500">{{ importedResult[row.clubMatchId] }}</span>
                      <span v-if="markFinalError[row.clubMatchId]" class="text-xs text-red-400">{{ markFinalError[row.clubMatchId] }}</span>
                      <span v-if="markSubmittedError[row.clubMatchId]" class="text-xs text-red-400">{{ markSubmittedError[row.clubMatchId] }}</span>
                      <span v-if="recalcScoreError[row.clubMatchId]" class="text-xs text-red-400">{{ recalcScoreError[row.clubMatchId] }}</span>
                      <span v-if="recalcScoreDone[row.clubMatchId]" class="text-xs text-green-500">Recalculated</span>
                      <button
                        v-if="row.dataStatus === 'submitted' || row.dataStatus === 'final'"
                        @click="recalculateScore(row)"
                        :disabled="recalcScoreLoading[row.clubMatchId]"
                        title="Recalculates the club's FFL score from latest AFL stats."
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ recalcScoreLoading[row.clubMatchId] ? 'Recalculating…' : 'Recalculate' }}</button>
                      <button
                        v-if="row.dataStatus === 'submitted'"
                        @click="markTeamFinal(row)"
                        :disabled="markingFinal[row.clubMatchId]"
                        title="Locks this club's team submission as final for scoring."
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ markingFinal[row.clubMatchId] ? 'Marking…' : 'Mark Final' }}</button>
                      <button
                        v-if="row.dataStatus === 'final'"
                        @click="markTeamSubmitted(row)"
                        :disabled="markingSubmitted[row.clubMatchId]"
                        title="Reverts this club's team submission from final."
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors disabled:opacity-40"
                      >{{ markingSubmitted[row.clubMatchId] ? 'Marking…' : 'Mark Submitted' }}</button>
                      <button
                        @click="toggleImportPanel(row.clubMatchId, row.clubSeasonId, row.clubName)"
                        class="rounded border border-border px-3 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors"
                      >Import Team</button>
                    </div>
                  </td>
                </tr>

                <!-- Inline import panel -->
                <tr v-if="activeImportClubMatchId === row.clubMatchId" class="border-b border-border">
                  <td colspan="4" class="pb-4 pt-2 px-2">
                    <div class="rounded-xl border border-border bg-surface-raised shadow-sm p-5">
                    <div v-if="importPhase === 'input'" class="space-y-4 max-w-xl mx-auto">
                      <!-- Team format -->
                      <div>
                        <label class="block text-xs font-medium text-text-muted mb-1">Team format</label>
                        <select
                          v-model="teamName"
                          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
                        >
                          <option value="">Auto-detect</option>
                          <option value="Ruiboys">Ruiboys</option>
                          <option value="Slashers">Slashers</option>
                          <option value="Cheetahs">Cheetahs</option>
                          <option value="THC">THC</option>
                        </select>
                      </div>
                      <!-- Paste area -->
                      <div>
                        <label class="block text-xs font-medium text-text-muted mb-1">Team (e.g. forum post)</label>
                        <textarea
                          v-model="post"
                          rows="14"
                          placeholder="Paste team here…"
                          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-active resize-y"
                        />
                      </div>
                      <p v-if="parseError" class="text-sm text-red-400">{{ parseError }}</p>
                      <div class="flex items-center justify-between">
                        <button
                          @click="toggleImportPanel(row.clubMatchId, row.clubSeasonId, row.clubName)"
                          class="rounded-lg border border-border px-3 py-1.5 text-sm text-text hover:bg-surface-hover transition-colors"
                        >Cancel</button>
                        <button
                          @click="onParse"
                          :disabled="!canParse || parsing"
                          class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                        >{{ parsing ? 'Reading…' : 'Read Team' }}</button>
                      </div>
                    </div>

                    <!-- Review table -->
                    <div v-else-if="importPhase === 'review'">
                      <div class="mb-3 flex items-center gap-3">
                        <button
                          @click="toggleImportPanel(row.clubMatchId, row.clubSeasonId, row.clubName)"
                          class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
                        >Cancel</button>
                        <button
                          @click="importPhase = 'input'"
                          class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
                        >← Back</button>
                        <span class="text-sm text-text-muted">
                          {{ resolvedPlayers.length }} players ·
                          <span :class="needsReview.length > 0 ? 'text-yellow-500' : 'text-green-500'">
                            {{ needsReview.length }} need review
                          </span>
                        </span>
                        <button
                          @click="onConfirm"
                          :disabled="confirming || unresolvedCount > 0 || imported"
                          class="ml-auto rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                        >{{ confirming ? 'Saving…' : 'Confirm & Import' }}</button>
                      </div>
                      <p v-if="unresolvedCount > 0" class="mb-2 text-sm text-red-400">
                        {{ unresolvedCount }} player(s) unresolved — confirm disabled.
                      </p>
                      <p v-if="compositionWarnings.length > 0" class="mb-2 text-sm text-yellow-500">
                        Invalid team: {{ compositionWarnings.join(' · ') }}
                      </p>
                      <p v-if="confirmError" class="mb-2 text-sm text-red-400">{{ confirmError }}</p>

                      <table class="w-full">
                        <thead>
                          <tr class="border-b border-border">
                            <th class="px-4 pb-2 text-left text-xs font-medium text-text-faint">Posted</th>
                            <th class="pb-2"></th>
                            <th class="px-4 pb-2 text-left text-xs font-medium text-text-faint">Resolved</th>
                            <th class="px-4 pb-2 text-left text-xs font-medium text-text-faint">Position</th>
                            <th class="px-4 pb-2 text-right text-xs font-medium text-text-faint">Score</th>
                            <th class="px-4 pb-2 text-right text-xs font-medium text-text-faint">Confidence</th>
                            <th class="pb-2"></th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr
                            v-for="(rp, i) in resolvedPlayers"
                            :key="i"
                            class="border-b border-border last:border-0"
                            :class="rowClass(i)"
                          >
                            <td class="pl-4 pr-4 py-3">
                              <div class="font-mono text-sm font-medium text-text">{{ rp.parsedName }}</div>
                              <div class="text-xs text-text-faint">{{ rp.clubHint }}</div>
                            </td>
                            <td class="py-3 text-text-faint text-sm select-none px-1">→</td>
                            <td class="pl-4 pr-4 py-3">
                              <div class="text-sm font-semibold text-text">
                                <span v-if="rp.resolvedName">{{ rp.resolvedName }}</span>
                                <span v-else class="text-red-400 font-normal">Unresolved</span>
                              </div>
                              <div class="text-xs text-text-muted">{{ rp.resolvedClub ?? '—' }}</div>
                            </td>
                            <td class="px-4 py-3 text-sm whitespace-nowrap">
                              <span :class="POSITION_COLORS[rp.position] ?? 'text-text-muted'">{{ displayPosition(rp) }}</span>
                            </td>
                            <td class="px-4 py-3 text-right tabular-nums text-sm text-text">
                              {{ rp.score ?? '—' }}
                            </td>
                            <td class="pl-4 pr-4 py-3 text-right">
                              <span
                                class="inline-block rounded-full px-2.5 py-0.5 text-xs font-medium"
                                :class="confidenceBadge(rp.confidence)"
                              >{{ (rp.confidence * 100).toFixed(0) }}%</span>
                            </td>
                            <td class="pr-4 py-3 text-right">
                              <button
                                v-if="rp.confidence < 1"
                                @click="openLinkModal(i)"
                                class="rounded border border-border px-2 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors"
                              >Fix</button>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

      </template>

      <FflPlayerLinkModal
        :show="linkModalRowIndex !== null"
        :ffl-season-id="season?.id ?? ''"
        :ffl-club-season-id="activeImportClubSeasonId"
        :parsed-name="linkModalPlayer?.parsedName ?? ''"
        :club-hint="linkModalPlayer?.clubHint ?? ''"
        @close="linkModalRowIndex = null"
        @linked="onPlayerLinked"
      />
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Import Squads                          -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'squads'">
      <!-- Season selector (local to this tab) -->
      <div class="mb-5 max-w-2xl">
        <label class="block text-xs font-medium text-text-muted mb-1">Season</label>
        <select
          v-model="squadSeasonId"
          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
        >
          <option v-if="!fflSeasonOptions.length" value="">Loading…</option>
          <option v-for="s in fflSeasonOptions" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>

      <div v-if="loadingSquadSeason" class="text-text-faint">Loading…</div>
      <div v-else-if="squadSeasonError" class="text-red-400">{{ squadSeasonError.message }}</div>
      <template v-else-if="squadSeason">
        <!-- Paste + parse -->
        <div v-if="squadPhase === 'input'" class="space-y-4 max-w-2xl">
          <p class="text-sm text-text-muted">
            Paste this season's squads thread. Members resolve against
            <span class="font-medium text-text">{{ squadSeason.name }}</span>’s AFL players; assign each squad to a club and import.
          </p>
          <div>
            <label class="block text-xs font-medium text-text-muted mb-1">Effective from</label>
            <select
              v-model="squadFromRoundId"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
            >
              <option value="">Season start</option>
              <option v-for="r in squadSeason.rounds" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium text-text-muted mb-1">Squads thread</label>
            <textarea
              v-model="squadThread"
              rows="16"
              placeholder="Paste squads thread here…"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-active resize-y"
            />
          </div>
          <p v-if="squadParseError" class="text-sm text-red-400">{{ squadParseError }}</p>
          <div class="flex justify-end">
            <button
              @click="onParseSquads"
              :disabled="!canParseSquads || parsingSquads"
              class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >{{ parsingSquads ? 'Reading…' : 'Read Squads' }}</button>
          </div>
        </div>

        <!-- Review -->
        <div v-else-if="squadPhase === 'review'" class="space-y-5">
          <div class="flex items-center gap-3">
            <button
              @click="squadPhase = 'input'"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >← Back</button>
            <span class="text-sm text-text-muted">
              {{ resolvedSquads.length }} squads · effective {{ squadFromRoundLabel }}
            </span>
          </div>

          <div v-for="(sq, si) in resolvedSquads" :key="si" class="rounded-xl border border-border bg-surface-raised p-5">
            <!-- Squad header -->
            <div class="flex flex-wrap items-center gap-3 mb-3">
              <div class="font-semibold text-text">{{ sq.clubName }}</div>
              <span class="text-xs text-text-faint">{{ sq.members.length }} members</span>
              <span class="text-xs" :class="squadUnresolvedCount(sq) > 0 ? 'text-yellow-500' : 'text-green-500'">
                {{ squadUnresolvedCount(sq) }} unresolved
              </span>
              <div class="ml-auto flex items-center gap-2">
                <span v-if="sq.importedMsg" class="text-xs text-green-500">{{ sq.importedMsg }}</span>
                <span v-if="sq.importError" class="text-xs text-red-400">{{ sq.importError }}</span>
                <select
                  v-model="sq.clubSeasonId"
                  :disabled="sq.imported"
                  class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:outline-none focus:border-active disabled:opacity-50"
                >
                  <option value="">Assign to club…</option>
                  <option v-for="cs in fflClubSeasons" :key="cs.id" :value="cs.id">{{ cs.club.name }}</option>
                </select>
                <button
                  @click="onImportSquad(si)"
                  :disabled="!sq.clubSeasonId || squadUnresolvedCount(sq) > 0 || sq.importing || sq.imported"
                  class="rounded-lg border border-active bg-active px-4 py-1.5 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                >{{ sq.importing ? 'Importing…' : sq.imported ? 'Imported' : 'Import Squad' }}</button>
              </div>
            </div>

            <!-- Members -->
            <table class="w-full">
              <thead>
                <tr class="border-b border-border">
                  <th class="pb-2 pr-2 text-right text-xs font-medium text-text-faint w-px">#</th>
                  <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Posted</th>
                  <th class="pb-2"></th>
                  <th class="pb-2 pr-4 text-left text-xs font-medium text-text-faint">Resolved</th>
                  <th class="pb-2 pr-4 text-right text-xs font-medium text-text-faint">Cost</th>
                  <th class="pb-2 pr-4 text-right text-xs font-medium text-text-faint">Confidence</th>
                  <th class="pb-2"></th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(m, mi) in sq.members"
                  :key="mi"
                  class="border-b border-border last:border-0"
                  :class="!m.aflPlayerSeasonId ? 'bg-red-500/5' : (m.confidence < 1 ? 'bg-yellow-500/5' : '')"
                >
                  <td class="py-2 pr-2 text-right text-xs text-text-faint tabular-nums">{{ m.rank }}</td>
                  <td class="py-2 pr-4">
                    <span class="font-mono text-sm font-medium text-text">{{ m.parsedName }}</span>
                    <span class="ml-2 text-xs text-text-faint">{{ m.clubHint }}</span>
                  </td>
                  <td class="py-2 text-text-faint text-sm select-none px-1">→</td>
                  <td class="py-2 pr-4">
                    <span v-if="m.resolvedName" class="text-sm font-semibold text-text">{{ m.resolvedName }}</span>
                    <span v-else class="text-sm text-red-400">Unresolved</span>
                    <span class="ml-2 text-xs text-text-muted">{{ m.resolvedClub ?? '' }}</span>
                  </td>
                  <td class="py-2 pr-4 text-right text-sm tabular-nums text-text-muted">
                    {{ m.costCents != null ? (m.costCents / 100).toFixed(1) : '—' }}
                  </td>
                  <td class="py-2 pr-4 text-right">
                    <span
                      class="inline-block rounded-full px-2.5 py-0.5 text-xs font-medium"
                      :class="confidenceBadge(m.confidence)"
                    >{{ (m.confidence * 100).toFixed(0) }}%</span>
                  </td>
                  <td class="py-2 text-right">
                    <button
                      v-if="m.confidence < 1 && !sq.imported"
                      @click="openSquadLink(si, mi)"
                      class="rounded border border-border px-2 py-1 text-xs font-medium text-text hover:bg-surface-hover transition-colors"
                    >Fix</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>

      <SquadMemberLinkModal
        :show="squadLinkTarget !== null"
        :ffl-season-id="squadSeasonId"
        :parsed-name="squadLinkMember?.parsedName ?? ''"
        :club-hint="squadLinkMember?.clubHint ?? ''"
        @close="squadLinkTarget = null"
        @linked="onSquadMemberLinked"
      />
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Import Fixtures                        -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'fixtures'">
      <!-- Season selector (local to this tab) -->
      <div class="mb-5 max-w-2xl">
        <label class="block text-xs font-medium text-text-muted mb-1">Season</label>
        <select
          v-model="fixtureSeasonId"
          class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:outline-none focus:ring-1 focus:ring-active"
        >
          <option v-if="!fflSeasonOptions.length" value="">Loading…</option>
          <option v-for="s in fflSeasonOptions" :key="s.id" :value="s.id">{{ s.name }}</option>
        </select>
      </div>

      <div v-if="loadingFixtureSeason" class="text-text-faint">Loading…</div>
      <div v-else-if="fixtureSeasonError" class="text-red-400">{{ fixtureSeasonError.message }}</div>
      <template v-else-if="fixtureSeason">
        <!-- Paste + parse -->
        <div v-if="fixturePhase === 'input'" class="space-y-4 max-w-2xl">
          <p class="text-sm text-text-muted">
            Paste <span class="font-medium text-text">{{ fixtureSeason.name }}</span>’s fixture sheet.
            Rounds and pairings are read from it; club-level scores are kept as reference notes. Map each round to an AFL round, then import.
          </p>
          <div>
            <label class="block text-xs font-medium text-text-muted mb-1">Fixture sheet</label>
            <textarea
              v-model="fixtureSheet"
              rows="16"
              placeholder="Paste the tab-separated fixture sheet here…"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text font-mono focus:outline-none focus:ring-1 focus:ring-active resize-y"
            />
          </div>
          <p v-if="fixtureParseError" class="text-sm text-red-400">{{ fixtureParseError }}</p>
          <div class="flex justify-end">
            <button
              @click="onParseFixtures"
              :disabled="!canParseFixtures || parsingFixtures"
              class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
            >{{ parsingFixtures ? 'Reading…' : 'Read Fixtures' }}</button>
          </div>
        </div>

        <!-- Review -->
        <div v-else-if="fixturePhase === 'review'" class="space-y-5">
          <div class="flex flex-wrap items-center gap-3">
            <button
              @click="fixturePhase = 'input'"
              class="rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text hover:bg-surface-hover transition-colors"
            >← Back</button>
            <span class="text-sm text-text-muted">
              {{ includedFixtureRounds.length }} of {{ fixtureRounds.length }} rounds selected
            </span>
            <div class="ml-auto flex items-center gap-3">
              <span v-if="fixtureImportResult" class="text-xs text-green-500">
                {{ fixtureImportResult.roundsCreated }} rounds · {{ fixtureImportResult.scoresWritten }} reference scores
              </span>
              <span v-if="fixtureImportError" class="text-xs text-red-400">{{ fixtureImportError }}</span>
              <button
                @click="onImportFixtures"
                :disabled="!canImportFixtures || importingFixtures"
                class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
              >{{ importingFixtures ? 'Importing…' : 'Import Fixtures' }}</button>
            </div>
          </div>

          <p v-if="fixtureUnresolved.length" class="text-sm text-red-400">
            Unresolved club names — nothing was written. Fix the spelling in the sheet or register the club:
            {{ fixtureUnresolved.join(', ') }}
          </p>

          <div
            v-for="(rd, ri) in fixtureRounds"
            :key="ri"
            class="rounded-xl border border-border p-5"
            :class="rd.include ? 'bg-surface-raised' : 'bg-surface-raised/40 opacity-60'"
          >
            <!-- Round header -->
            <div class="flex flex-wrap items-center gap-3 mb-3">
              <label class="flex items-center gap-2 text-sm font-semibold text-text">
                <input type="checkbox" v-model="rd.include" class="accent-active" />
                {{ rd.round > 0 ? `Round ${rd.round}` : 'Finals' }}
                <span v-if="rd.label" class="text-xs font-normal text-text-faint">{{ rd.label }}</span>
              </label>
              <span class="text-xs text-text-faint">{{ rd.fixtures.length }} fixtures</span>
              <div class="ml-auto flex items-center gap-2">
                <input
                  v-model="rd.name"
                  placeholder="Round name"
                  class="w-36 rounded-lg border border-border bg-surface px-3 py-1.5 text-sm text-text focus:outline-none focus:border-active"
                />
                <select
                  v-model="rd.aflRoundId"
                  class="rounded-lg border bg-surface px-3 py-1.5 text-sm text-text focus:outline-none focus:border-active"
                  :class="rd.include && !rd.aflRoundId ? 'border-red-400' : 'border-border'"
                >
                  <option value="">Map to AFL round…</option>
                  <option v-for="ar in fixtureAflRounds" :key="ar.id" :value="ar.id">{{ ar.name }}</option>
                </select>
              </div>
            </div>

            <!-- Fixtures -->
            <table class="w-full">
              <tbody>
                <tr v-for="(fx, fi) in rd.fixtures" :key="fi" class="border-b border-border last:border-0">
                  <td class="py-2 pr-3 text-right text-sm" :class="clubResolves(fx.homeClub) ? 'text-text' : 'text-red-400'">
                    {{ fx.homeClub }}
                  </td>
                  <td class="py-2 px-2 text-right text-sm tabular-nums text-text-muted w-px">{{ fx.homeScore ?? '' }}</td>
                  <td class="py-2 px-1 text-center text-xs text-text-faint select-none w-px">vs</td>
                  <td class="py-2 px-2 text-left text-sm tabular-nums text-text-muted w-px">{{ fx.awayScore ?? '' }}</td>
                  <td class="py-2 pl-3 text-sm" :class="clubResolves(fx.awayClub) ? 'text-text' : 'text-red-400'">
                    {{ fx.awayClub }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>
    </div>

    <!-- Tab: Forum Capture (historical import, slice 1) -->
    <div v-if="activeTab === 'forum-capture'" class="space-y-4">
      <div class="flex items-center justify-between gap-4">
        <p class="text-sm text-text-muted">
          Pages captured this session via the forum userscript. Ephemeral preview — nothing is saved.
        </p>
        <div class="flex gap-2 shrink-0">
          <button @click="refetchCaptured()" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised">Refresh</button>
          <button @click="clearCaptures()" class="text-sm px-3 py-1.5 rounded border border-border hover:bg-surface-raised">Clear</button>
        </div>
      </div>

      <div v-if="capturedPages.length === 0" class="text-text-faint text-sm">
        No pages captured yet. Install <code>dev/userscripts/ffl-forum-capture.user.js</code>, open a round thread, and click “Capture round”.
      </div>

      <div v-for="page in capturedPages" :key="page.topicId" class="rounded-lg border border-border p-4 space-y-3">
        <div class="text-sm font-semibold">
          {{ page.season || '—' }} · {{ page.roundTitle }}
          <span class="text-text-faint font-normal">(topic {{ page.topicId }}, {{ page.posts.length }} posts)</span>
        </div>

        <div v-for="post in page.posts" :key="post.postId" class="rounded border border-border p-3">
          <div class="flex items-center gap-2 text-sm">
            <span class="font-medium">{{ post.author }}</span>
            <span
              class="px-1.5 py-0.5 rounded text-xs"
              :class="post.team ? 'bg-active text-white' : 'bg-surface-raised text-text-faint'"
            >{{ post.team || 'unknown author' }}</span>
            <span class="text-xs text-text-faint">
              {{ post.isTeamSubmission ? `${post.players.length} players` : 'not a team submission' }}
            </span>
          </div>
          <div v-if="post.parseError" class="mt-1 text-xs text-text-muted">parse error: {{ post.parseError }}</div>
          <table v-if="post.players.length" class="mt-2 w-full text-xs">
            <tbody>
              <tr v-for="(pl, i) in post.players" :key="i" class="border-t border-border">
                <td class="py-0.5 pr-3 text-text-faint">{{ pl.position }}{{ pl.backupPositions ? ' (' + pl.backupPositions + ')' : '' }}</td>
                <td class="py-0.5 pr-3">{{ pl.name }}</td>
                <td class="py-0.5 pr-3 text-text-faint">{{ pl.clubHint }}</td>
                <td class="py-0.5 tabular-nums text-right">{{ pl.score ?? '' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ═══════════════════════════════════════════ -->
    <!-- Tab: Calculate                              -->
    <!-- ═══════════════════════════════════════════ -->
    <div v-if="activeTab === 'calculate'" class="space-y-6 max-w-lg">
      <!-- AFL Ladder -->
      <div class="rounded-xl border border-border bg-surface-raised p-5">
        <h2 class="text-sm font-semibold mb-1">AFL Ladder</h2>
        <p class="text-xs text-text-faint mb-4">Rebuilds AFL club season standings from all final matches for the current season.</p>
        <div class="flex items-center gap-3">
          <button
            @click="recalculateAFLLadder"
            :disabled="recalcAflLoading || !aflSeasonId"
            class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ recalcAflLoading ? 'Recalculating…' : 'Recalculate AFL Ladder' }}</button>
          <span v-if="recalcAflDone" class="text-sm text-green-500">Done</span>
          <span v-if="recalcAflError" class="text-sm text-red-400">{{ recalcAflError }}</span>
        </div>
      </div>

      <!-- FFL Ladder -->
      <div class="rounded-xl border border-border bg-surface-raised p-5">
        <h2 class="text-sm font-semibold mb-1">FFL Ladder</h2>
        <p class="text-xs text-text-faint mb-4">Rebuilds FFL club season standings from all final matches for the current season.</p>
        <div class="flex items-center gap-3">
          <button
            @click="recalculateFFLLadder"
            :disabled="recalcFflLoading || !liveSeasonId"
            class="rounded-lg border border-active bg-active px-4 py-2 text-sm font-medium text-active-text transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
          >{{ recalcFflLoading ? 'Recalculating…' : 'Recalculate FFL Ladder' }}</button>
          <span v-if="recalcFflDone" class="text-sm text-green-500">Done</span>
          <span v-if="recalcFflError" class="text-sm text-red-400">{{ recalcFflError }}</span>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useQuery, useMutation } from '@vue/apollo-composable'
import { GET_FFL_DATA_OPS, GET_AFL_ROUND_STATS, GET_FFL_CAPTURED_PAGES, GET_AFL_SEASON_CLUB_SEASONS, GET_FFL_FIXTURE_IMPORT_SEASON } from '../api/queries'
import { PARSE_TEAM_SUBMISSION, CONFIRM_TEAM_SUBMISSION, IMPORT_AFL_MATCH_STATS, MARK_AFL_MATCH_STATS_COMPLETE, MARK_FFL_TEAM_FINAL, MARK_FFL_TEAM_SUBMITTED, RECALCULATE_FFL_CLUB_MATCH_SCORE, CLEAR_FFL_FORUM_CAPTURES, RECALCULATE_AFL_LADDER, RECALCULATE_FFL_LADDER, PARSE_FFL_SQUAD_THREAD, IMPORT_FFL_SQUAD, PARSE_FFL_FIXTURE_SHEET, IMPORT_FFL_FIXTURES } from '../api/mutations'
import { useFflState } from '@/features/ffl/composables/useFflState'
import { GET_FFL_SEASONS } from '@/features/ffl/api/queries'
import { canonicalClub } from '../utils/clubAliases'
import { useAflState } from '@/features/afl/composables/useAflState'
import { GET_AFL_LIVE_ROUND } from '@/features/afl/api/queries'
import { clubLogoUrl } from '@/features/afl/utils/clubLogos'
import { clubLogoUrl as fflClubLogoUrl } from '@/features/ffl/utils/clubLogos'
import { POSITION_COLORS, POSITION_LABEL, POSITION_SLOTS } from '@/features/ffl/utils/position'
import PlayerSearchModal from '../components/PlayerSearchModal.vue'
import FflPlayerLinkModal from '../components/FflPlayerLinkModal.vue'
import SquadMemberLinkModal from '../components/SquadMemberLinkModal.vue'
import AflRoundNav from '@/features/afl/components/RoundNav.vue'
import FflRoundNav from '@/features/ffl/components/RoundNav.vue'

const { liveSeasonId, liveRoundId } = useFflState()
const { liveRoundId: aflLiveRoundId } = useAflState()
const route = useRoute()

const initialTab = (route.query.tab as string) || 'team-submission'
const initialRound = (route.query.round as string) || ''

// ---- Tabs ----
const tabs = [
  { id: 'team-submission', label: 'Import FFL Teams' },
  { id: 'afl-stats', label: 'Import AFL Stats' },
  { id: 'calculate', label: 'Calculate' },
  { id: 'squads', label: 'Import FFL Squads' },
  { id: 'fixtures', label: 'Import FFL Fixtures' },
  { id: 'forum-capture', label: 'FFL Round Forum Capture' },
]
const activeTab = ref(initialTab)

// ---- Forum Capture (historical import, slice 1) ----
const { result: capturedResult, refetch: refetchCaptured } = useQuery(GET_FFL_CAPTURED_PAGES)
const capturedPages = computed(() => capturedResult.value?.fflCapturedPages ?? [])
const { mutate: clearCapturesMutate } = useMutation(CLEAR_FFL_FORUM_CAPTURES)
async function clearCaptures() {
  await clearCapturesMutate()
  await refetchCaptured()
}

// ════════════════════════════════════════════
// AFL Stats Import
// ════════════════════════════════════════════

const { result: liveRoundResult, loading: loadingLiveRound, error: liveRoundError } = useQuery(GET_AFL_LIVE_ROUND)

const aflRounds = computed(() => liveRoundResult.value?.aflLiveRound?.round?.season?.rounds ?? [])
const selectedAflRoundId = ref(initialTab === 'afl-stats' && initialRound ? initialRound : '')

watch(liveRoundResult, (val) => {
  if (val?.aflLiveRound?.round?.id && !selectedAflRoundId.value) {
    selectedAflRoundId.value = val.aflLiveRound.round.id
  }
}, { immediate: true })

const { result: roundStatsResult, loading: loadingRoundStats, refetch: refetchRoundStats } = useQuery(
  GET_AFL_ROUND_STATS,
  () => ({ roundId: selectedAflRoundId.value }),
  () => ({ enabled: !!selectedAflRoundId.value }),
)

const aflMatches = computed(() => roundStatsResult.value?.aflRound?.matches ?? [])
const aflByes = computed(() => roundStatsResult.value?.aflRound?.byes ?? [])

type UnmatchedAFLPlayer = {
  parsedName: string
  clubMatchId: string
  kicks: number; handballs: number; marks: number; hitouts: number; tackles: number; goals: number; behinds: number
}

type ScrapeResult = {
  homeClubName: string; awayClubName: string
  homePlayerCount: number; awayPlayerCount: number
  unmatchedPlayers: UnmatchedAFLPlayer[]
}

const scraping = ref<Record<string, boolean>>({})
const scrapeResult = ref<Record<string, ScrapeResult>>({})
const scrapeError = ref<Record<string, string>>({})
const togglingFinal = ref<Record<string, boolean>>({})

// keyed by matchId → rowIndex → true when resolved
const resolvedUnmatched = ref<Record<string, Record<number, boolean>>>({})

// resolve modal state
type ModalPlayer = UnmatchedAFLPlayer & { clubName: string } & {
  clubSeasonId: string
  matchId: string
  rowIndex: number
}
const resolveModalPlayer = ref<ModalPlayer | null>(null)

const { mutate: importStatsMutation } = useMutation(IMPORT_AFL_MATCH_STATS)
const { mutate: markCompleteMutation } = useMutation(MARK_AFL_MATCH_STATS_COMPLETE)

async function scrape(match: any) {
  scraping.value[match.id] = true
  scrapeError.value[match.id] = ''
  scrapeResult.value[match.id] = undefined as any
  resolvedUnmatched.value[match.id] = {}
  try {
    const res = await importStatsMutation({ matchId: match.id })
    const data = res?.data?.importAFLMatchStats
    if (data) scrapeResult.value[match.id] = data
  } catch (e: any) {
    scrapeError.value[match.id] = e.message ?? 'Scrape failed'
  } finally {
    scraping.value[match.id] = false
  }
}

async function toggleFinal(match: any) {
  togglingFinal.value[match.id] = true
  try {
    await markCompleteMutation({
      matchId: match.id,
      complete: match.dataStatus !== 'final',
    })
  } catch (e: any) {
    scrapeError.value[match.id] = e.message ?? 'Update failed'
  } finally {
    togglingFinal.value[match.id] = false
  }
}

// Build a map from clubMatchId → clubSeasonId using the round query result
const clubMatchSeasonMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const m of aflMatches.value) {
    if (m.homeClubMatch) map[m.homeClubMatch.id] = m.homeClubMatch.clubSeasonId
    if (m.awayClubMatch) map[m.awayClubMatch.id] = m.awayClubMatch.clubSeasonId
  }
  return map
})

const clubMatchClubMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {}
  for (const m of aflMatches.value) {
    if (m.homeClubMatch) map[m.homeClubMatch.id] = m.homeClubMatch.club.name
    if (m.awayClubMatch) map[m.awayClubMatch.id] = m.awayClubMatch.club.name
  }
  return map
})

function openResolveModal(match: any, rowIndex: number, up: UnmatchedAFLPlayer) {
  resolveModalPlayer.value = {
    ...up,
    clubSeasonId: clubMatchSeasonMap.value[up.clubMatchId] ?? '',
    clubName: clubMatchClubMap.value[up.clubMatchId] ?? '',
    matchId: match.id,
    rowIndex,
  }
}

async function onPlayerResolved() {
  const modal = resolveModalPlayer.value
  if (modal) {
    if (!resolvedUnmatched.value[modal.matchId]) resolvedUnmatched.value[modal.matchId] = {}
    resolvedUnmatched.value[modal.matchId][modal.rowIndex] = true
  }
  resolveModalPlayer.value = null
  await refetchRoundStats()
}


function statusLabel(status: string): string {
  if (status === 'final') return 'Final'
  if (status === 'partial') return 'Partial'
  return 'No data'
}

function statusBadge(status: string): string {
  if (status === 'final') return 'bg-green-500/15 text-green-500'
  if (status === 'partial') return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-surface-raised text-text-faint'
}

function fflStatusLabel(status: string): string {
  if (status === 'final') return 'Final'
  if (status === 'submitted') return 'Submitted'
  return 'Not submitted'
}

function fflStatusBadge(status: string): string {
  if (status === 'final') return 'bg-green-500/15 text-green-500'
  if (status === 'submitted') return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-surface-raised text-text-faint'
}

// ════════════════════════════════════════════
// FFL Team Submission
// ════════════════════════════════════════════

const { result: seasonResult, loading: loadingSeasonData, error: seasonError, refetch: refetchSeasonData } = useQuery(
  GET_FFL_DATA_OPS,
  () => ({ seasonId: liveSeasonId.value }),
  () => ({ enabled: !!liveSeasonId.value }),
)

const season = computed(() => seasonResult.value?.fflSeason ?? null)

const selectedRoundId = ref(initialTab === 'team-submission' && initialRound ? initialRound : liveRoundId.value)

watch(liveRoundId, (val) => {
  if (val && !selectedRoundId.value) selectedRoundId.value = val
}, { immediate: true })

// Sync tab + round from route query params when RoundNav navigates within DataOps
watch(() => route.query.tab, (val) => { if (val) activeTab.value = val as string })
watch(() => route.query.round, (val) => {
  if (!val) return
  if (activeTab.value === 'afl-stats') selectedAflRoundId.value = val as string
  if (activeTab.value === 'team-submission') selectedRoundId.value = val as string
})

const selectedRound = computed(() =>
  season.value?.rounds.find((r: any) => r.id === selectedRoundId.value) ?? null,
)

type FflClubRow = {
  clubMatchId: string
  matchId: string
  roundId: string
  clubSeasonId: string
  clubName: string
  dataStatus: string
  score: number
  matchStyle: string // '' | 'bye' | 'superbye'
}

const fflClubRows = computed<FflClubRow[]>(() => {
  if (!selectedRound.value) return []
  const rows: FflClubRow[] = []
  // One row per club_match — covers home/away fixtures plus bye and superbye
  // teams, which submit like any other club.
  for (const match of selectedRound.value.matches) {
    for (const cm of match.clubMatches ?? []) {
      rows.push({
        clubMatchId: cm.id,
        matchId: match.id,
        roundId: selectedRoundId.value,
        clubSeasonId: cm.clubSeasonId,
        clubName: cm.club.name,
        dataStatus: cm.dataStatus ?? 'no_data',
        score: cm.score ?? 0,
        matchStyle: match.matchStyle ?? '',
      })
    }
  }
  return rows.map(r => ({ ...r, ...(rowOverrides.value[r.clubMatchId] ?? {}) }))
})

// Local overrides to update individual rows without refetching the whole table.
const rowOverrides = ref<Record<string, Partial<FflClubRow>>>({})

function patchRow(clubMatchId: string, patch: Partial<FflClubRow>) {
  rowOverrides.value[clubMatchId] = { ...(rowOverrides.value[clubMatchId] ?? {}), ...patch }
}

function clearRowStatus(clubMatchId: string) {
  markFinalError.value[clubMatchId] = ''
  markSubmittedError.value[clubMatchId] = ''
  recalcScoreError.value[clubMatchId] = ''
  recalcScoreDone.value[clubMatchId] = false
  importedResult.value[clubMatchId] = ''
}

// ---- Import panel (inline, per-club-match) ----

const activeImportClubMatchId = ref('')
const activeImportClubSeasonId = ref('')
const importedResult = ref<Record<string, string>>({})

function toggleImportPanel(clubMatchId: string, clubSeasonId: string, clubName: string) {
  if (activeImportClubMatchId.value === clubMatchId) {
    activeImportClubMatchId.value = ''
    activeImportClubSeasonId.value = ''
    resetImportState()
  } else {
    activeImportClubMatchId.value = clubMatchId
    activeImportClubSeasonId.value = clubSeasonId
    delete importedResult.value[clubMatchId]
    resetImportState()
    // Pre-select team format matching club name
    teamName.value = ['Ruiboys', 'Slashers', 'Cheetahs', 'THC'].find(n =>
      clubName.toLowerCase().includes(n.toLowerCase())
    ) ?? ''
  }
}

function resetImportState() {
  importPhase.value = 'input'
  post.value = ''
  parseError.value = ''
  confirmError.value = ''
  imported.value = false
  resolvedPlayers.value = []
  needsReview.value = []
}

const importPhase = ref<'input' | 'review'>('input')
const teamName = ref('')
const post = ref('')
const parseError = ref('')
const parsing = ref(false)

const canParse = computed(() =>
  !!activeImportClubMatchId.value && post.value.trim().length > 0,
)

type ResolvedPlayer = {
  parsedName: string; clubHint: string; resolvedName: string | null; resolvedClub: string | null
  position: string; backupPositions: string; interchangePosition: string
  score: number | null; notes: string; playerSeasonId: string | null; confidence: number
}

const resolvedPlayers = ref<ResolvedPlayer[]>([])
const needsReview = ref<number[]>([])

const { mutate: parseMutation } = useMutation(PARSE_TEAM_SUBMISSION)

async function onParse() {
  parseError.value = ''
  confirmError.value = ''
  imported.value = false
  parsing.value = true
  try {
    const res = await parseMutation({
      input: {
        clubSeasonId: activeImportClubSeasonId.value,
        clubMatchId: activeImportClubMatchId.value,
        teamName: teamName.value,
        post: post.value,
      },
    })
    const data = res?.data?.parseFFLTeamSubmission
    if (!data) throw new Error('No result returned')
    resolvedPlayers.value = data.resolvedPlayers
    needsReview.value = data.needsReview
    importPhase.value = 'review'
  } catch (e: any) {
    parseError.value = e.message ?? 'Parse failed'
  } finally {
    parsing.value = false
  }
}

const confirming = ref(false)
const confirmError = ref('')
const imported = ref(false)

const { mutate: confirmMutation } = useMutation(CONFIRM_TEAM_SUBMISSION)

const unresolvedCount = computed(() => resolvedPlayers.value.filter(rp => !rp.playerSeasonId).length)

const compositionWarnings = computed(() => {
  const counts: Record<string, number> = {}
  for (const rp of resolvedPlayers.value) {
    if (!rp.playerSeasonId) continue
    if (rp.backupPositions || rp.interchangePosition) continue
    counts[rp.position] = (counts[rp.position] ?? 0) + 1
  }
  return Object.entries(POSITION_SLOTS)
    .filter(([pos, expected]) => (counts[pos] ?? 0) !== expected)
    .map(([pos, expected]) => `${POSITION_LABEL[pos] ?? pos} ${counts[pos] ?? 0}/${expected}`)
})

async function onConfirm() {
  confirmError.value = ''
  confirming.value = true
  try {
    const players = resolvedPlayers.value
      .filter(rp => rp.playerSeasonId)
      .map(rp => ({
        playerSeasonId: rp.playerSeasonId!,
        position: rp.position,
        backupPositions: rp.backupPositions || null,
        interchangePosition: rp.interchangePosition || null,
        score: rp.score,
      }))
    const res = await confirmMutation({ input: { clubMatchId: activeImportClubMatchId.value, players } })
    const saved = res?.data?.confirmFFLTeamSubmission ?? []
    importedResult.value[activeImportClubMatchId.value] = `Imported ${saved.length} player records.`
    imported.value = true
    activeImportClubMatchId.value = ''
    activeImportClubSeasonId.value = ''
    await refetchSeasonData()
  } catch (e: any) {
    const gqlErr = e?.graphQLErrors?.[0]
    if (gqlErr?.extensions?.code === 'BYE_INELIGIBLE') {
      const psId = gqlErr.extensions.playerSeasonId
      const rp = resolvedPlayers.value.find(p => p.playerSeasonId === psId)
      confirmError.value = `${rp?.resolvedName ?? 'A player'} is on a bye but didn't play last round — remove them to import`
    } else {
      confirmError.value = gqlErr?.message ?? e.message ?? 'Confirm failed'
    }
  } finally {
    confirming.value = false
  }
}

// ---- Mark FFL team final ----

const markingFinal = ref<Record<string, boolean>>({})
const markFinalError = ref<Record<string, string>>({})
const markingSubmitted = ref<Record<string, boolean>>({})
const markSubmittedError = ref<Record<string, string>>({})

const { mutate: markFinalMutation } = useMutation(MARK_FFL_TEAM_FINAL)
const { mutate: markSubmittedMutation } = useMutation(MARK_FFL_TEAM_SUBMITTED)

async function markTeamFinal(row: FflClubRow) {
  clearRowStatus(row.clubMatchId)
  markingFinal.value[row.clubMatchId] = true
  try {
    await markFinalMutation({ input: { clubMatchId: row.clubMatchId, matchId: row.matchId, roundId: row.roundId } })
    patchRow(row.clubMatchId, { dataStatus: 'final' })
  } catch (e: any) {
    markFinalError.value[row.clubMatchId] = e.message ?? 'Failed to mark final'
  } finally {
    markingFinal.value[row.clubMatchId] = false
  }
}

async function markTeamSubmitted(row: FflClubRow) {
  clearRowStatus(row.clubMatchId)
  markingSubmitted.value[row.clubMatchId] = true
  try {
    await markSubmittedMutation({ input: { clubMatchId: row.clubMatchId, matchId: row.matchId, roundId: row.roundId } })
    patchRow(row.clubMatchId, { dataStatus: 'submitted' })
  } catch (e: any) {
    markSubmittedError.value[row.clubMatchId] = e.message ?? 'Failed to mark submitted'
  } finally {
    markingSubmitted.value[row.clubMatchId] = false
  }
}

const recalcScoreLoading = ref<Record<string, boolean>>({})
const recalcScoreError = ref<Record<string, string>>({})
const recalcScoreDone = ref<Record<string, boolean>>({})

const { mutate: recalcScoreMutation } = useMutation(RECALCULATE_FFL_CLUB_MATCH_SCORE)

async function recalculateScore(row: FflClubRow) {
  clearRowStatus(row.clubMatchId)
  recalcScoreLoading.value[row.clubMatchId] = true
  try {
    const res = await recalcScoreMutation({ clubMatchId: row.clubMatchId })
    const newScore = res?.data?.recalculateFFLClubMatchScore
    if (newScore != null) patchRow(row.clubMatchId, { score: newScore })
    recalcScoreDone.value[row.clubMatchId] = true
  } catch (e: any) {
    recalcScoreError.value[row.clubMatchId] = e.message ?? 'Failed to recalculate'
  } finally {
    recalcScoreLoading.value[row.clubMatchId] = false
  }
}

// ---- FFL player link modal ----

const linkModalRowIndex = ref<number | null>(null)

const linkModalPlayer = computed(() =>
  linkModalRowIndex.value !== null ? resolvedPlayers.value[linkModalRowIndex.value] : null,
)

function openLinkModal(i: number) {
  linkModalRowIndex.value = i
}

function onPlayerLinked(data: { playerSeasonId: string; resolvedName: string; resolvedClub: string }) {
  if (linkModalRowIndex.value === null) return
  const i = linkModalRowIndex.value
  resolvedPlayers.value[i] = {
    ...resolvedPlayers.value[i],
    playerSeasonId: data.playerSeasonId,
    resolvedName: data.resolvedName,
    resolvedClub: data.resolvedClub,
    confidence: 1,
  }
  needsReview.value = needsReview.value.filter(idx => idx !== i)
  linkModalRowIndex.value = null
}

function displayPosition(rp: ResolvedPlayer): string {
  if (rp.backupPositions) {
    const labels = rp.backupPositions.split(',').map(p => {
      const key = p.trim()
      const label = POSITION_LABEL[key] ?? key
      return key === rp.interchangePosition ? `${label} (Int)` : label
    }).join(', ')
    return `Bench - ${labels}`
  }
  return POSITION_LABEL[rp.position] ?? rp.position
}

function confidenceBadge(confidence: number): string {
  if (confidence >= 0.85) return 'bg-green-500/15 text-green-500'
  if (confidence >= 0.6) return 'bg-yellow-500/15 text-yellow-500'
  return 'bg-red-500/15 text-red-400'
}

function rowClass(i: number): string {
  const rp = resolvedPlayers.value[i]
  if (!rp.playerSeasonId) return 'bg-red-500/5'
  if (needsReview.value.includes(i)) return 'bg-yellow-500/5'
  return ''
}

// ════════════════════════════════════════════
// Import Squads (historical import, slice 3)
// ════════════════════════════════════════════

type SquadMember = {
  rank: number; parsedName: string; clubHint: string; costCents: number | null
  resolvedName: string | null; resolvedClub: string | null
  aflPlayerSeasonId: string | null; confidence: number
}
type ReviewSquad = {
  clubName: string
  clubSeasonId: string
  members: SquadMember[]
  importing: boolean; imported: boolean; importedMsg: string; importError: string
}

const squadPhase = ref<'input' | 'review'>('input')
const squadThread = ref('')
const squadFromRoundId = ref('')
const parsingSquads = ref(false)
const squadParseError = ref('')
const resolvedSquads = ref<ReviewSquad[]>([])

// ---- Season selection (local to this tab; not the global live season) ----
// Squad import targets any historical season, so this tab picks its own —
// defaulting to the live one — without disturbing the live-season-driven Teams tab.
const { result: fflSeasonsResult } = useQuery(GET_FFL_SEASONS)
const fflSeasonOptions = computed<{ id: string; name: string }[]>(() =>
  [...(fflSeasonsResult.value?.fflSeasons ?? [])].sort((a: any, b: any) => b.name.localeCompare(a.name)),
)

const squadSeasonId = ref(liveSeasonId.value)
watch(liveSeasonId, (val) => { if (val && !squadSeasonId.value) squadSeasonId.value = val }, { immediate: true })

const { result: squadSeasonResult, loading: loadingSquadSeason, error: squadSeasonError } = useQuery(
  GET_FFL_DATA_OPS,
  () => ({ seasonId: squadSeasonId.value }),
  () => ({ enabled: !!squadSeasonId.value }),
)
const squadSeason = computed(() => squadSeasonResult.value?.fflSeason ?? null)

// The AFL season backing the selected FFL season — the resolution candidate pool.
const { result: aflSeasonResult } = useQuery(
  GET_AFL_SEASON_CLUB_SEASONS,
  () => ({ fflSeasonId: squadSeasonId.value }),
  () => ({ enabled: !!squadSeasonId.value }),
)
const squadAflSeasonId = computed(() => aflSeasonResult.value?.fflSeason?.aflSeason?.id ?? '')

// FFL club_seasons for the selected season, for assigning each parsed squad to a club.
const fflClubSeasons = computed(() => squadSeason.value?.ladder ?? [])

// Switching season abandons any in-progress review.
watch(squadSeasonId, () => {
  squadPhase.value = 'input'
  resolvedSquads.value = []
  squadThread.value = ''
  squadFromRoundId.value = ''
  squadParseError.value = ''
})

const canParseSquads = computed(() => !!squadAflSeasonId.value && squadThread.value.trim().length > 0)

const squadFromRoundLabel = computed(() => {
  if (!squadFromRoundId.value) return 'from season start'
  const r = squadSeason.value?.rounds.find((r: any) => r.id === squadFromRoundId.value)
  return r ? `from ${r.name}` : 'from season start'
})

function squadUnresolvedCount(sq: ReviewSquad): number {
  return sq.members.filter(m => !m.aflPlayerSeasonId).length
}

// Match a parsed club header to an FFL club_season by name (case-insensitive,
// either direction contains) so the assignment dropdown is pre-filled.
function autoAssignClubSeasonId(clubName: string): string {
  const key = clubName.trim().toLowerCase()
  const match = fflClubSeasons.value.find((cs: any) => {
    const name = cs.club.name.toLowerCase()
    return name === key || name.includes(key) || key.includes(name)
  })
  return match?.id ?? ''
}

const { mutate: parseSquadsMutation } = useMutation(PARSE_FFL_SQUAD_THREAD)
const { mutate: importSquadMutation } = useMutation(IMPORT_FFL_SQUAD)

async function onParseSquads() {
  squadParseError.value = ''
  parsingSquads.value = true
  try {
    const res = await parseSquadsMutation({
      input: { aflSeasonId: squadAflSeasonId.value, thread: squadThread.value },
    })
    const data = res?.data?.parseFFLSquadThread
    if (!data) throw new Error('No result returned')
    resolvedSquads.value = data.squads.map((sq: any) => ({
      clubName: sq.clubName,
      clubSeasonId: autoAssignClubSeasonId(sq.clubName),
      members: sq.members.map((m: any) => ({ ...m })),
      importing: false, imported: false, importedMsg: '', importError: '',
    }))
    squadPhase.value = 'review'
  } catch (e: any) {
    squadParseError.value = e.message ?? 'Parse failed'
  } finally {
    parsingSquads.value = false
  }
}

async function onImportSquad(si: number) {
  const sq = resolvedSquads.value[si]
  if (!sq || !sq.clubSeasonId) return
  sq.importError = ''
  sq.importing = true
  try {
    const members = sq.members
      .filter(m => m.aflPlayerSeasonId)
      .map(m => ({
        aflPlayerSeasonId: m.aflPlayerSeasonId!,
        name: m.resolvedName ?? m.parsedName,
        costCents: m.costCents,
      }))
    const res = await importSquadMutation({
      input: {
        clubSeasonId: sq.clubSeasonId,
        fromRoundId: squadFromRoundId.value || null,
        members,
      },
    })
    const saved = res?.data?.importFFLSquad ?? []
    sq.imported = true
    sq.importedMsg = `Imported ${saved.length} players`
    await refetchSeasonData()
  } catch (e: any) {
    sq.importError = e?.graphQLErrors?.[0]?.message ?? e.message ?? 'Import failed'
  } finally {
    sq.importing = false
  }
}

// ---- Squad member link modal ----

const squadLinkTarget = ref<{ si: number; mi: number } | null>(null)

const squadLinkMember = computed(() =>
  squadLinkTarget.value ? resolvedSquads.value[squadLinkTarget.value.si]?.members[squadLinkTarget.value.mi] : null,
)

function openSquadLink(si: number, mi: number) {
  squadLinkTarget.value = { si, mi }
}

function onSquadMemberLinked(data: { aflPlayerSeasonId: string; resolvedName: string; resolvedClub: string }) {
  if (!squadLinkTarget.value) return
  const { si, mi } = squadLinkTarget.value
  const m = resolvedSquads.value[si]?.members[mi]
  if (m) {
    m.aflPlayerSeasonId = data.aflPlayerSeasonId
    m.resolvedName = data.resolvedName
    m.resolvedClub = data.resolvedClub
    m.confidence = 1
  }
  squadLinkTarget.value = null
}

// ════════════════════════════════════════════
// Import Fixtures (historical import, slice 4)
// ════════════════════════════════════════════

type ReviewFixture = { homeClub: string; homeScore: number | null; awayClub: string; awayScore: number | null }
type FixtureReviewRound = {
  round: number
  label: string
  name: string       // editable FFL round name
  aflRoundId: string // reviewer-mapped AFL round
  include: boolean
  fixtures: ReviewFixture[]
}

const fixturePhase = ref<'input' | 'review'>('input')
const fixtureSheet = ref('')
const parsingFixtures = ref(false)
const fixtureParseError = ref('')
const fixtureRounds = ref<FixtureReviewRound[]>([])
const importingFixtures = ref(false)
const fixtureImportError = ref('')
const fixtureImportResult = ref<{ roundsCreated: number; scoresWritten: number } | null>(null)
const fixtureUnresolved = ref<string[]>([])

// Season selection (local to this tab), and its club_seasons + AFL rounds.
const fixtureSeasonId = ref(liveSeasonId.value)
watch(liveSeasonId, (val) => { if (val && !fixtureSeasonId.value) fixtureSeasonId.value = val }, { immediate: true })

const { result: fixtureSeasonResult, loading: loadingFixtureSeason, error: fixtureSeasonError } = useQuery(
  GET_FFL_FIXTURE_IMPORT_SEASON,
  () => ({ id: fixtureSeasonId.value }),
  () => ({ enabled: !!fixtureSeasonId.value }),
)
const fixtureSeason = computed(() => fixtureSeasonResult.value?.fflSeason ?? null)
const fixtureAflRounds = computed(() => fixtureSeason.value?.aflSeason?.rounds ?? [])

// Canonical club keys registered for the season, for pre-import resolution hints.
// Alias-aware (via clubAliases) so the hint matches the backend's resolution.
const fixtureClubNames = computed(() => {
  const set = new Set<string>()
  for (const cs of fixtureSeason.value?.ladder ?? []) set.add(canonicalClub(cs.club.name))
  return set
})
function clubResolves(name: string): boolean {
  return fixtureClubNames.value.has(canonicalClub(name))
}

const canParseFixtures = computed(() => !!fixtureSeasonId.value && fixtureSheet.value.trim().length > 0)

const includedFixtureRounds = computed(() => fixtureRounds.value.filter(r => r.include))
const canImportFixtures = computed(() =>
  includedFixtureRounds.value.length > 0 && includedFixtureRounds.value.every(r => !!r.aflRoundId),
)

// Best-effort map of a parsed round to an AFL round by matching the round number
// against the AFL round name; the reviewer adjusts anything it misses.
function autoMapAflRound(round: number): string {
  if (round <= 0) return ''
  const match = fixtureAflRounds.value.find((ar: any) => {
    const name = ar.name.toLowerCase()
    return name === `round ${round}` || new RegExp(`\\b${round}\\b`).test(name)
  })
  return match?.id ?? ''
}

// Switching season abandons any in-progress review.
watch(fixtureSeasonId, () => {
  fixturePhase.value = 'input'
  fixtureSheet.value = ''
  fixtureRounds.value = []
  fixtureParseError.value = ''
  fixtureImportResult.value = null
  fixtureImportError.value = ''
  fixtureUnresolved.value = []
})

const { mutate: parseFixturesMutation } = useMutation(PARSE_FFL_FIXTURE_SHEET)
const { mutate: importFixturesMutation } = useMutation(IMPORT_FFL_FIXTURES)

async function onParseFixtures() {
  fixtureParseError.value = ''
  fixtureImportResult.value = null
  fixtureImportError.value = ''
  fixtureUnresolved.value = []
  parsingFixtures.value = true
  try {
    const res = await parseFixturesMutation({ input: { sheet: fixtureSheet.value } })
    const data = res?.data?.parseFFLFixtureSheet
    if (!data) throw new Error('No result returned')
    fixtureRounds.value = data.rounds.map((r: any) => ({
      round: r.round,
      label: r.label,
      name: r.round > 0 ? `Round ${r.round}` : (r.label || 'Finals'),
      aflRoundId: autoMapAflRound(r.round),
      include: r.round > 0, // finals (round 0) left to the manual builder — off by default
      fixtures: r.fixtures.map((fx: any) => ({ ...fx })),
    }))
    fixturePhase.value = 'review'
  } catch (e: any) {
    fixtureParseError.value = e.message ?? 'Parse failed'
  } finally {
    parsingFixtures.value = false
  }
}

async function onImportFixtures() {
  fixtureImportError.value = ''
  fixtureImportResult.value = null
  fixtureUnresolved.value = []
  importingFixtures.value = true
  try {
    const rounds = includedFixtureRounds.value.map(r => ({
      name: r.name,
      aflRoundId: r.aflRoundId,
      fixtures: r.fixtures.map(fx => ({
        homeClub: fx.homeClub,
        homeScore: fx.homeScore,
        awayClub: fx.awayClub,
        awayScore: fx.awayScore,
      })),
    }))
    const res = await importFixturesMutation({ input: { seasonId: fixtureSeasonId.value, rounds } })
    const data = res?.data?.importFFLFixtures
    if (!data) throw new Error('No result returned')
    if (data.unresolved?.length) {
      fixtureUnresolved.value = data.unresolved
      return
    }
    fixtureImportResult.value = { roundsCreated: data.roundsCreated, scoresWritten: data.scoresWritten }
  } catch (e: any) {
    fixtureImportError.value = e?.graphQLErrors?.[0]?.message ?? e.message ?? 'Import failed'
  } finally {
    importingFixtures.value = false
  }
}

// ════════════════════════════════════════════
// Calculate
// ════════════════════════════════════════════

// Reuses the live-round query the AFL Stats tab already runs.
const aflSeasonId = computed(() => liveRoundResult.value?.aflLiveRound?.round?.season?.id ?? '')

const recalcAflLoading = ref(false)
const recalcAflError = ref('')
const recalcAflDone = ref(false)
const recalcFflLoading = ref(false)
const recalcFflError = ref('')
const recalcFflDone = ref(false)

const { mutate: recalcAflMutation } = useMutation(RECALCULATE_AFL_LADDER)
const { mutate: recalcFflMutation } = useMutation(RECALCULATE_FFL_LADDER)

async function recalculateAFLLadder() {
  recalcAflLoading.value = true
  recalcAflError.value = ''
  recalcAflDone.value = false
  try {
    await recalcAflMutation({ seasonId: aflSeasonId.value })
    recalcAflDone.value = true
  } catch (e: any) {
    recalcAflError.value = e.message ?? 'Failed'
  } finally {
    recalcAflLoading.value = false
  }
}

async function recalculateFFLLadder() {
  recalcFflLoading.value = true
  recalcFflError.value = ''
  recalcFflDone.value = false
  try {
    await recalcFflMutation({ seasonId: liveSeasonId.value })
    recalcFflDone.value = true
  } catch (e: any) {
    recalcFflError.value = e.message ?? 'Failed'
  } finally {
    recalcFflLoading.value = false
  }
}
</script>
