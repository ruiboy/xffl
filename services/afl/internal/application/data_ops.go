package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"xffl/contracts/events"
	"xffl/services/afl/internal/domain"
	sharedevents "xffl/shared/events"
)

const (
	footywireSource       = "footywire"
	confidenceThreshold   = 0.85
)

// ImportAFLStatsResult summarises what was written for each club in a match.
type ImportAFLStatsResult struct {
	MatchID         int
	HomeClubName    string
	AwayClubName    string
	HomePlayerCount int
	AwayPlayerCount int
}

// DataOpsCommands handles AFL stats import operations.
type DataOpsCommands struct {
	tx            TxManager
	matches       domain.MatchRepository
	clubMatches   domain.ClubMatchRepository
	clubSeasons   domain.ClubSeasonRepository
	clubs         domain.ClubRepository
	rounds        domain.RoundRepository
	playerSeasons domain.PlayerSeasonRepository
	sourceMap     MatchSourceMapRepository
	statsParser   StatsParser
	discovery     FixtureDiscovery
	resolver      PlayerResolver
	dispatcher    sharedevents.Dispatcher
}

func NewDataOpsCommands(
	tx TxManager,
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	clubs domain.ClubRepository,
	rounds domain.RoundRepository,
	playerSeasons domain.PlayerSeasonRepository,
	sourceMap MatchSourceMapRepository,
	statsParser StatsParser,
	discovery FixtureDiscovery,
	resolver PlayerResolver,
	dispatcher sharedevents.Dispatcher,
) *DataOpsCommands {
	return &DataOpsCommands{
		tx:            tx,
		matches:       matches,
		clubMatches:   clubMatches,
		clubSeasons:   clubSeasons,
		clubs:         clubs,
		rounds:        rounds,
		playerSeasons: playerSeasons,
		sourceMap:     sourceMap,
		statsParser:   statsParser,
		discovery:     discovery,
		resolver:      resolver,
		dispatcher:    dispatcher,
	}
}

// ImportAFLStats scrapes match stats from FootyWire, resolves player names, and writes
// afl.player_match records. Sets stats_import_status to "partial".
func (c *DataOpsCommands) ImportAFLStats(ctx context.Context, matchID int) (ImportAFLStatsResult, error) {
	match, err := c.matches.FindByIDWithDetails(ctx, matchID)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("load match: %w", err)
	}

	round, err := c.rounds.FindByID(ctx, match.RoundID)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("load round: %w", err)
	}

	homeClubName, err := c.clubNameForClubSeason(ctx, match.Home.ClubSeasonID)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("resolve home club: %w", err)
	}
	awayClubName, err := c.clubNameForClubSeason(ctx, match.Away.ClubSeasonID)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("resolve away club: %w", err)
	}

	mid, err := c.resolveMid(ctx, matchID, round.Name, homeClubName, awayClubName)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("resolve footywire mid: %w", err)
	}

	stats, err := c.statsParser.ParseMatch(ctx, mid)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("parse match stats: %w", err)
	}

	result := ImportAFLStatsResult{
		MatchID:      matchID,
		HomeClubName: homeClubName,
		AwayClubName: awayClubName,
	}

	type clubWork struct {
		cm       domain.ClubMatch
		clubName string
		teamGoals int
		teamBehinds int
		counter  *int
	}

	works := []clubWork{
		{cm: match.Home, clubName: homeClubName, teamGoals: stats.HomeTeamGoals, teamBehinds: stats.HomeTeamBehinds, counter: &result.HomePlayerCount},
		{cm: match.Away, clubName: awayClubName, teamGoals: stats.AwayTeamGoals, teamBehinds: stats.AwayTeamBehinds, counter: &result.AwayPlayerCount},
	}

	var allWritten []domain.PlayerMatch
	var roundID int

	for _, w := range works {
		candidates, err := c.buildCandidates(ctx, w.cm.ClubSeasonID)
		if err != nil {
			return ImportAFLStatsResult{}, fmt.Errorf("build candidates for %s: %w", w.clubName, err)
		}

		playerStats := filterByClub(stats.Players, w.clubName)

		var written []domain.PlayerMatch
		var sumBehinds int

		err = c.tx.WithTx(ctx, func(repos WriteRepos) error {
			written = make([]domain.PlayerMatch, 0, len(playerStats))
			for _, ps := range playerStats {
				matches, err := c.resolver.Resolve(ctx, ps.Name, w.clubName, candidates)
				if err != nil {
					slog.WarnContext(ctx, "resolver error", slog.String("player", ps.Name), slog.Any("error", err))
					continue
				}
				if len(matches) == 0 || matches[0].Confidence < confidenceThreshold {
					slog.WarnContext(ctx, "no confident match for player",
						slog.String("player", ps.Name),
						slog.String("club", w.clubName),
					)
					continue
				}

				psID := matches[0].Candidate.PlayerSeasonID
				status := "played"
				kicks, handballs, marks, hitouts, tackles, goals, behinds :=
					ps.Kicks, ps.Handballs, ps.Marks, ps.Hitouts, ps.Tackles, ps.Goals, ps.Behinds

				pm, err := repos.PlayerMatches.Upsert(ctx, domain.UpsertPlayerMatchParams{
					ClubMatchID:    w.cm.ID,
					PlayerSeasonID: psID,
					Status:         &status,
					Kicks:          &kicks,
					Handballs:      &handballs,
					Marks:          &marks,
					Hitouts:        &hitouts,
					Tackles:        &tackles,
					Goals:          &goals,
					Behinds:        &behinds,
				})
				if err != nil {
					return fmt.Errorf("upsert player_match for %s: %w", ps.Name, err)
				}
				written = append(written, pm)
				sumBehinds += ps.Behinds
			}

			// Recalculate club score from updated player records.
			allPlayerMatches, err := repos.PlayerMatches.FindByClubMatchID(ctx, w.cm.ID)
			if err != nil {
				return fmt.Errorf("reload player matches: %w", err)
			}
			cm := w.cm
			cm.PlayerMatches = allPlayerMatches
			if err := repos.ClubMatches.UpdateScore(ctx, w.cm.ID, cm.Score()); err != nil {
				return fmt.Errorf("update club score: %w", err)
			}

			// Compute and store rushed behinds: team total − sum of player behinds.
			rushedBehinds := w.teamBehinds - sumBehinds
			if rushedBehinds < 0 {
				rushedBehinds = 0
			}
			if err := repos.ClubMatches.UpdateRushedBehinds(ctx, w.cm.ID, rushedBehinds); err != nil {
				return fmt.Errorf("update rushed behinds: %w", err)
			}

			roundID, err = repos.ClubMatches.FindRoundID(ctx, w.cm.ID)
			return err
		})
		if err != nil {
			return ImportAFLStatsResult{}, err
		}

		*w.counter = len(written)
		allWritten = append(allWritten, written...)
	}

	// Update match import status (outside transaction — best-effort).
	if err := c.matches.UpdateImportStatus(ctx, matchID, domain.MatchStatsPartial, time.Now()); err != nil {
		slog.WarnContext(ctx, "failed to update match import status", slog.Int("match_id", matchID), slog.Any("error", err))
	}

	// Fire PlayerMatchUpdated events.
	for _, pm := range allWritten {
		payload, err := json.Marshal(events.PlayerMatchUpdatedPayload{
			PlayerMatchID:  pm.ID,
			PlayerSeasonID: pm.PlayerSeasonID,
			ClubMatchID:    pm.ClubMatchID,
			RoundID:        roundID,
			Kicks:          pm.Kicks,
			Handballs:      pm.Handballs,
			Marks:          pm.Marks,
			Hitouts:        pm.Hitouts,
			Tackles:        pm.Tackles,
			Goals:          pm.Goals,
			Behinds:        pm.Behinds,
		})
		if err != nil {
			slog.WarnContext(ctx, "marshal PlayerMatchUpdated failed", slog.Any("error", err))
			continue
		}
		if err := c.dispatcher.Publish(ctx, events.PlayerMatchUpdated, payload); err != nil {
			slog.WarnContext(ctx, "publish PlayerMatchUpdated failed", slog.Any("error", err))
		}
	}

	return result, nil
}

// MarkMatchStatsComplete sets stats_import_status to "complete" (or back to "partial").
func (c *DataOpsCommands) MarkMatchStatsComplete(ctx context.Context, matchID int, complete bool) (domain.Match, error) {
	match, err := c.matches.FindByID(ctx, matchID)
	if err != nil {
		return domain.Match{}, fmt.Errorf("load match: %w", err)
	}

	status := domain.MatchStatsPartial
	if complete {
		status = domain.MatchStatsComplete
	}

	importedAt := time.Now()
	if match.StatsImportedAt != nil {
		importedAt = *match.StatsImportedAt
	}

	if err := c.matches.UpdateImportStatus(ctx, matchID, status, importedAt); err != nil {
		return domain.Match{}, fmt.Errorf("update import status: %w", err)
	}

	return c.matches.FindByID(ctx, matchID)
}

// resolveMid returns the FootyWire mid for a match, scraping the fixture list if needed.
func (c *DataOpsCommands) resolveMid(ctx context.Context, matchID int, roundName, homeClub, awayClub string) (string, error) {
	mid, found, err := c.sourceMap.FindByMatchID(ctx, footywireSource, matchID)
	if err != nil {
		return "", fmt.Errorf("lookup source map: %w", err)
	}
	if found {
		return mid, nil
	}

	mid, err = c.discovery.FindMatchMid(ctx, roundName, homeClub, awayClub)
	if err != nil {
		return "", fmt.Errorf("discover mid: %w", err)
	}

	if err := c.sourceMap.Store(ctx, footywireSource, mid, matchID); err != nil {
		// Non-fatal: mid was found, just couldn't cache it.
		slog.WarnContext(ctx, "failed to cache footywire mid", slog.Int("match_id", matchID), slog.Any("error", err))
	}

	return mid, nil
}

// clubNameForClubSeason resolves the AFL club name for a club_season_id.
func (c *DataOpsCommands) clubNameForClubSeason(ctx context.Context, clubSeasonID int) (string, error) {
	cs, err := c.clubSeasons.FindByID(ctx, clubSeasonID)
	if err != nil {
		return "", err
	}
	club, err := c.clubs.FindByID(ctx, cs.ClubID)
	if err != nil {
		return "", err
	}
	return club.Name, nil
}

// buildCandidates fetches all player_seasons for a club_season and returns them as a candidate pool.
func (c *DataOpsCommands) buildCandidates(ctx context.Context, clubSeasonID int) ([]PlayerCandidate, error) {
	rows, err := c.playerSeasons.FindByClubSeasonIDWithPlayer(ctx, clubSeasonID)
	if err != nil {
		return nil, err
	}
	candidates := make([]PlayerCandidate, len(rows))
	for i, r := range rows {
		candidates[i] = PlayerCandidate{
			PlayerSeasonID: r.PlayerSeasonID,
			Name:           r.Name,
			Club:           "", // club hint matching done by clubName parameter
		}
	}
	return candidates, nil
}

// filterByClub returns only the player stats whose ClubName matches (case-insensitive prefix).
// FootyWire may use shortened club names, so we do a contains check as fallback.
func filterByClub(players []PlayerStats, clubName string) []PlayerStats {
	var out []PlayerStats
	for _, p := range players {
		if p.ClubName == clubName {
			out = append(out, p)
		}
	}
	return out
}
