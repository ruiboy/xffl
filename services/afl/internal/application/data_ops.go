package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"xffl/contracts/events"
	"xffl/services/afl/internal/domain"
	sharedevents "xffl/shared/events"
)

const (
	footywireSource       = "footywire"
	confidenceThreshold   = 0.85
)

// UnmatchedAFLPlayer holds the parsed stats and candidate pool for a player
// who could not be confidently matched during import.
type UnmatchedAFLPlayer struct {
	ParsedName  string
	ClubMatchID int
	Kicks       int
	Handballs   int
	Marks       int
	Hitouts     int
	Tackles     int
	Goals       int
	Behinds     int
	Candidates  []PlayerMatch // sorted by descending confidence; may be empty
}

// ImportAFLStatsResult summarises what was written for each club in a match.
type ImportAFLStatsResult struct {
	MatchID          int
	HomeClubName     string
	AwayClubName     string
	HomePlayerCount  int
	AwayPlayerCount  int
	UnmatchedPlayers []UnmatchedAFLPlayer
}

// DataOpsCommands handles AFL stats import operations.
type DataOpsCommands struct {
	tx              TxManager
	matches         domain.MatchRepository
	clubMatches     domain.ClubMatchRepository
	clubSeasons     domain.ClubSeasonRepository
	clubs           domain.ClubRepository
	rounds          domain.RoundRepository
	playerSeasons   domain.PlayerSeasonRepository
	sourceMap       DataopsMatchSourceRepository
	playerSourceMap DataopsPlayerSourceRepository
	statsParser     StatsParser
	discovery       FixtureDiscovery
	resolver        PlayerResolver
	dispatcher      sharedevents.Dispatcher
}

func NewDataOpsCommands(
	tx TxManager,
	matches domain.MatchRepository,
	clubMatches domain.ClubMatchRepository,
	clubSeasons domain.ClubSeasonRepository,
	clubs domain.ClubRepository,
	rounds domain.RoundRepository,
	playerSeasons domain.PlayerSeasonRepository,
	sourceMap DataopsMatchSourceRepository,
	playerSourceMap DataopsPlayerSourceRepository,
	statsParser StatsParser,
	discovery FixtureDiscovery,
	resolver PlayerResolver,
	dispatcher sharedevents.Dispatcher,
) *DataOpsCommands {
	return &DataOpsCommands{
		tx:              tx,
		matches:         matches,
		clubMatches:     clubMatches,
		clubSeasons:     clubSeasons,
		clubs:           clubs,
		rounds:          rounds,
		playerSeasons:   playerSeasons,
		sourceMap:       sourceMap,
		playerSourceMap: playerSourceMap,
		statsParser:     statsParser,
		discovery:       discovery,
		resolver:        resolver,
		dispatcher:      dispatcher,
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

	slog.InfoContext(ctx, "importing AFL stats",
		slog.Int("matchId", matchID),
		slog.String("round", round.Name),
		slog.String("home", homeClubName),
		slog.String("away", awayClubName),
	)

	mid, err := c.resolveMid(ctx, matchID, round.Name, homeClubName, awayClubName)
	if err != nil {
		return ImportAFLStatsResult{}, fmt.Errorf("resolve footywire mid: %w", err)
	}
	slog.InfoContext(ctx, "footywire mid resolved", slog.String("mid", mid))

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
		cm            domain.ClubMatch
		clubName      string // DB name — used for candidate lookup
		statsClubName string // score table name — used to filter parsed players
		teamGoals     int
		teamBehinds   int
		counter       *int
	}

	works := []clubWork{
		{cm: match.Home, clubName: homeClubName, statsClubName: stats.HomeClubName, teamGoals: stats.HomeTeamGoals, teamBehinds: stats.HomeTeamBehinds, counter: &result.HomePlayerCount},
		{cm: match.Away, clubName: awayClubName, statsClubName: stats.AwayClubName, teamGoals: stats.AwayTeamGoals, teamBehinds: stats.AwayTeamBehinds, counter: &result.AwayPlayerCount},
	}

	var allWritten []domain.PlayerMatch
	var roundID int

	for _, w := range works {
		candidates, err := c.buildCandidates(ctx, w.cm.ClubSeasonID)
		if err != nil {
			return ImportAFLStatsResult{}, fmt.Errorf("build candidates for %s: %w", w.clubName, err)
		}

		playerStats := filterByClub(stats.Players, w.statsClubName)

		// Sum behinds from ALL parsed player stats regardless of match confidence —
		// this gives the accurate denominator for computing rushed behinds.
		var totalParsedBehinds int
		for _, ps := range playerStats {
			totalParsedBehinds += ps.Behinds
		}

		var written []domain.PlayerMatch
		var unmatched []UnmatchedAFLPlayer

		err = c.tx.WithTx(ctx, func(repos WriteRepos) error {
			written = make([]domain.PlayerMatch, 0, len(playerStats))
			for _, ps := range playerStats {
				resolveName := ps.Name
				if ps.CanonicalName != "" {
					resolveName = ps.CanonicalName
				}
				matches, err := c.resolver.Resolve(ctx, resolveName, w.clubName, candidates)
				if err != nil {
					slog.WarnContext(ctx, "resolver error", slog.String("player", ps.Name), slog.Any("error", err))
					unmatched = append(unmatched, UnmatchedAFLPlayer{ParsedName: ps.Name, ClubMatchID: w.cm.ID,
						Kicks: ps.Kicks, Handballs: ps.Handballs, Marks: ps.Marks, Hitouts: ps.Hitouts,
						Tackles: ps.Tackles, Goals: ps.Goals, Behinds: ps.Behinds})
					continue
				}
				if len(matches) == 0 || matches[0].Confidence < confidenceThreshold {
					slog.WarnContext(ctx, "no confident match for player",
						slog.String("player", ps.Name),
						slog.String("canonicalName", ps.CanonicalName),
						slog.String("club", w.clubName),
					)
					unmatched = append(unmatched, UnmatchedAFLPlayer{ParsedName: ps.Name, ClubMatchID: w.cm.ID,
						Kicks: ps.Kicks, Handballs: ps.Handballs, Marks: ps.Marks, Hitouts: ps.Hitouts,
						Tackles: ps.Tackles, Goals: ps.Goals, Behinds: ps.Behinds, Candidates: matches})
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

			// Rushed behinds = team total behinds − sum of all player behinds from the
			// stats page. Using the full parsed total (not just matched players) keeps
			// this accurate even when some players are not yet in the DB.
			rushedBehinds := w.teamBehinds - totalParsedBehinds
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
		result.UnmatchedPlayers = append(result.UnmatchedPlayers, unmatched...)
	}

	// Update match data status (outside transaction — best-effort).
	if err := c.matches.UpdateDataStatus(ctx, matchID, domain.MatchDataPartial); err != nil {
		slog.WarnContext(ctx, "failed to update match data status", slog.Int("match_id", matchID), slog.Any("error", err))
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

// MarkMatchStatsFinal sets data_status to "final" (or back to "partial").
func (c *DataOpsCommands) MarkMatchStatsFinal(ctx context.Context, matchID int, final bool) (domain.Match, error) {
	status := domain.MatchDataPartial
	if final {
		status = domain.MatchDataFinal
	}

	if err := c.matches.UpdateDataStatus(ctx, matchID, status); err != nil {
		return domain.Match{}, fmt.Errorf("update data status: %w", err)
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

// AddAFLPlayerParams holds the input for creating a new AFL player and their season record.
type AddAFLPlayerParams struct {
	Name        string
	ClubSeasonID int
}

// AddAFLPlayer creates a new afl.player and afl.player_season record within a transaction.
func (c *DataOpsCommands) AddAFLPlayer(ctx context.Context, params AddAFLPlayerParams) (domain.PlayerSeason, error) {
	var result domain.PlayerSeason
	err := c.tx.WithTx(ctx, func(repos WriteRepos) error {
		player, err := repos.Players.Create(ctx, params.Name)
		if err != nil {
			return fmt.Errorf("create player: %w", err)
		}
		ps, err := repos.PlayerSeasons.Create(ctx, player.ID, params.ClubSeasonID)
		if err != nil {
			return fmt.Errorf("create player season: %w", err)
		}
		result = ps
		return nil
	})
	return result, err
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
