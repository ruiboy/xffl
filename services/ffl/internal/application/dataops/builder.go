package dataops

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// Builder creates the season scaffolding:
// the season + its club_seasons here; fixtures via the fixture builder.
type Builder struct {
	tx application.TxManager
}

func NewBuilder(tx application.TxManager) *Builder {
	return &Builder{tx: tx}
}

// BuildSeasonParams: the FFL season name (free text), the scoring era to apply
// (chosen explicitly), the AFL season it maps to (resolved by the caller), and
// the existing clubs playing this season (by id — clubs are pre-registered).
type BuildSeasonParams struct {
	SeasonName  string
	RulesID     string
	AFLSeasonID int
	ClubIDs     []int
}

type BuiltClubSeason struct {
	ClubName     string
	ClubSeasonID int
}

type BuiltSeason struct {
	SeasonID    int
	RulesID     string
	ClubSeasons []BuiltClubSeason
}

// BuildSeason creates the season — with the chosen `rules_id` — and a
// club_season per selected (pre-existing) club, atomically.
func (b *Builder) BuildSeason(ctx context.Context, params BuildSeasonParams) (BuiltSeason, error) {
	if _, err := domain.RulesFor(params.RulesID); err != nil {
		return BuiltSeason{}, err
	}
	var out BuiltSeason
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		leagueID, err := ensureLeague(ctx, repos)
		if err != nil {
			return err
		}
		season, err := repos.Seasons.Create(ctx, leagueID, params.SeasonName, params.AFLSeasonID, params.RulesID)
		if err != nil {
			return err
		}
		out.SeasonID = season.ID
		out.RulesID = params.RulesID
		clubs, err := repos.Clubs.FindByIDs(ctx, params.ClubIDs)
		if err != nil {
			return err
		}
		for _, clubID := range params.ClubIDs {
			club, ok := clubs[clubID]
			if !ok {
				return fmt.Errorf("club %d not found", clubID)
			}
			cs, err := repos.ClubSeasons.Create(ctx, clubID, season.ID)
			if err != nil {
				return err
			}
			out.ClubSeasons = append(out.ClubSeasons, BuiltClubSeason{ClubName: club.Name, ClubSeasonID: cs.ID})
		}
		return nil
	})
	if err != nil {
		return BuiltSeason{}, err
	}
	return out, nil
}

// AddRound creates a round in a season (defaults to MINOR; the grand final is
// set explicitly). afl_round_id is resolved by the caller.
func (b *Builder) AddRound(ctx context.Context, seasonID int, name string, aflRoundID int, roundType domain.RoundType) (domain.Round, error) {
	var out domain.Round
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		r, err := repos.Rounds.Create(ctx, seasonID, name, aflRoundID, roundType)
		out = r
		return err
	})
	return out, err
}

// BuiltFixture is a created match with its two club_matches.
type BuiltFixture struct {
	MatchID         int
	HomeClubMatchID int
	AwayClubMatchID int
}

// AddFixture creates a regular home-vs-away fixture in a round: one match plus a
// home and away club_match, atomically. (The superbye variant comes later.)
func (b *Builder) AddFixture(ctx context.Context, roundID, homeClubSeasonID, awayClubSeasonID int) (BuiltFixture, error) {
	var out BuiltFixture
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		m, err := repos.Matches.Create(ctx, roundID, nil)
		if err != nil {
			return err
		}
		home, err := repos.ClubMatches.Create(ctx, m.ID, homeClubSeasonID, "home")
		if err != nil {
			return err
		}
		away, err := repos.ClubMatches.Create(ctx, m.ID, awayClubSeasonID, "away")
		if err != nil {
			return err
		}
		out = BuiltFixture{MatchID: m.ID, HomeClubMatchID: home.ID, AwayClubMatchID: away.ID}
		return nil
	})
	return out, err
}

// GeneratedRound is a round created by GenerateHomeAndAway with its fixtures.
type GeneratedRound struct {
	RoundID  int
	Name     string
	Fixtures []BuiltFixture
}

// GenerateHomeAndAway creates a full home-and-away season — `rounds` rounds of
// round-robin fixtures over the given club_seasons — atomically. Round N is named
// "<namePrefix>N" and maps to afl_round_id (aflRoundStart + N-1).
func (b *Builder) GenerateHomeAndAway(ctx context.Context, seasonID int, clubSeasonIDs []int, rounds, aflRoundStart int, namePrefix string) ([]GeneratedRound, error) {
	sched := RoundRobinPairings(len(clubSeasonIDs), rounds)
	if sched == nil {
		return nil, fmt.Errorf("cannot schedule %d clubs over %d rounds (need an even club count ≥ 2 and rounds ≥ 1)", len(clubSeasonIDs), rounds)
	}
	if namePrefix == "" {
		namePrefix = "Round "
	}
	var out []GeneratedRound
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		for i, pairings := range sched {
			rnd, err := repos.Rounds.Create(ctx, seasonID, fmt.Sprintf("%s%d", namePrefix, i+1), aflRoundStart+i, domain.RoundTypeMinor)
			if err != nil {
				return err
			}
			gr := GeneratedRound{RoundID: rnd.ID, Name: rnd.Name}
			for _, p := range pairings {
				m, err := repos.Matches.Create(ctx, rnd.ID, nil)
				if err != nil {
					return err
				}
				home, err := repos.ClubMatches.Create(ctx, m.ID, clubSeasonIDs[p.HomeIdx], "home")
				if err != nil {
					return err
				}
				away, err := repos.ClubMatches.Create(ctx, m.ID, clubSeasonIDs[p.AwayIdx], "away")
				if err != nil {
					return err
				}
				gr.Fixtures = append(gr.Fixtures, BuiltFixture{MatchID: m.ID, HomeClubMatchID: home.ID, AwayClubMatchID: away.ID})
			}
			out = append(out, gr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ensureLeague returns the single FFL league, creating it if none exists yet.
func ensureLeague(ctx context.Context, repos application.WriteRepos) (int, error) {
	leagues, err := repos.Leagues.FindAll(ctx)
	if err != nil {
		return 0, err
	}
	if len(leagues) > 0 {
		return leagues[0].ID, nil
	}
	l, err := repos.Leagues.Create(ctx, "FFL")
	if err != nil {
		return 0, err
	}
	return l.ID, nil
}
