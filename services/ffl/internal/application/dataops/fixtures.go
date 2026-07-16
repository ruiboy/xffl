package dataops

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// FixtureSpec is one home-vs-away pairing in a round, by club_season id.
type FixtureSpec struct {
	HomeClubSeasonID int
	AwayClubSeasonID int
}

// RoundSpec is the desired state of a single round. A nil RoundID means a new
// round; a set RoundID targets an existing round in the season.
//
// Byes are the clubs sitting out the head-to-head this round: unlike an AFL bye
// they still field a scoring team, so each becomes a single-sided bye match
// (one club_match, no opponent).
type RoundSpec struct {
	RoundID    *int
	Name       string
	AFLRoundID int
	Type       domain.RoundType
	Fixtures   []FixtureSpec
	Byes       []int // club_season ids on a (scoring) bye
	Superbye   []int // club_season ids in the round's superbye (empty = none)
}

// FixtureRound is a round loaded for the builder: its fixtures and byes, plus
// whether it is locked (has submitted teams, so its fixtures are immutable).
type FixtureRound struct {
	RoundID    int
	Name       string
	AFLRoundID int
	Type       domain.RoundType
	Locked     bool
	Fixtures   []FixtureSpec
	Byes       []int
	Superbye   []int
}

// LoadFixtures reads a season's rounds with their fixtures and byes for the
// builder to edit. Byes (single-sided bye matches) are surfaced separately from
// home-vs-away fixtures.
func (b *Builder) LoadFixtures(ctx context.Context, seasonID int) ([]FixtureRound, error) {
	var out []FixtureRound
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		rounds, err := repos.Rounds.FindBySeasonID(ctx, seasonID)
		if err != nil {
			return err
		}
		for _, rnd := range rounds {
			n, err := repos.PlayerMatches.CountByRoundID(ctx, rnd.ID)
			if err != nil {
				return err
			}
			fr := FixtureRound{
				RoundID: rnd.ID, Name: rnd.Name, AFLRoundID: rnd.AFLRoundID,
				Type: rnd.Type, Locked: n > 0,
			}
			matches, err := repos.Matches.FindByRoundID(ctx, rnd.ID)
			if err != nil {
				return err
			}
			for _, m := range matches {
				cms, err := repos.ClubMatches.FindByMatchID(ctx, m.ID)
				if err != nil {
					return err
				}
				switch m.MatchStyle {
				case "superbye":
					for _, cm := range cms {
						fr.Superbye = append(fr.Superbye, cm.ClubSeasonID)
					}
				case "bye":
					if len(cms) == 1 {
						fr.Byes = append(fr.Byes, cms[0].ClubSeasonID)
					}
				default:
					// Regular match: FindByMatchID orders home before away.
					if len(cms) == 2 {
						fr.Fixtures = append(fr.Fixtures, FixtureSpec{
							HomeClubSeasonID: cms[0].ClubSeasonID,
							AwayClubSeasonID: cms[1].ClubSeasonID,
						})
					}
				}
			}
			out = append(out, fr)
		}
		return nil
	})
	return out, err
}

// SaveFixtures reconciles a season's rounds to the given desired state in a
// single transaction. It is the one write path for the fixture builder — create,
// edit and delete all flow through here.
//
// Reconciliation is round-granular and protects entered teams:
//   - A round with submitted teams (any player_match) is immutable: it is left
//     untouched, and removing it is refused.
//   - An existing round without teams is replaced wholesale (its fixtures/byes
//     are rebuilt from the spec) and its metadata updated.
//   - A new round (nil RoundID) is created.
//   - An existing round absent from the spec is deleted (unless it has teams).
func (b *Builder) SaveFixtures(ctx context.Context, seasonID int, rounds []RoundSpec) error {
	return b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		existing, err := repos.Rounds.FindBySeasonID(ctx, seasonID)
		if err != nil {
			return err
		}
		existingByID := make(map[int]domain.Round, len(existing))
		hasTeams := make(map[int]bool, len(existing))
		for _, e := range existing {
			existingByID[e.ID] = e
			n, err := repos.PlayerMatches.CountByRoundID(ctx, e.ID)
			if err != nil {
				return err
			}
			hasTeams[e.ID] = n > 0
		}

		incomingIDs := make(map[int]bool)
		for _, r := range rounds {
			if r.RoundID != nil {
				if _, ok := existingByID[*r.RoundID]; !ok {
					return fmt.Errorf("round %d does not belong to season %d", *r.RoundID, seasonID)
				}
				incomingIDs[*r.RoundID] = true
			}
		}

		// Deletions: existing rounds no longer wanted.
		for _, e := range existing {
			if incomingIDs[e.ID] {
				continue
			}
			if hasTeams[e.ID] {
				return fmt.Errorf("cannot delete round %q: it has submitted teams", e.Name)
			}
			if err := deleteRoundFixtures(ctx, repos, e.ID); err != nil {
				return err
			}
			if err := repos.Rounds.SoftDelete(ctx, e.ID); err != nil {
				return err
			}
		}

		// Creates and replacements.
		for _, r := range rounds {
			if r.RoundID == nil {
				rnd, err := repos.Rounds.Create(ctx, seasonID, r.Name, r.AFLRoundID, r.Type)
				if err != nil {
					return err
				}
				if err := writeRoundFixtures(ctx, repos, rnd.ID, r); err != nil {
					return err
				}
				continue
			}
			id := *r.RoundID
			if hasTeams[id] {
				continue // immutable — teams entered
			}
			if err := repos.Rounds.Update(ctx, id, r.Name, r.AFLRoundID, r.Type); err != nil {
				return err
			}
			if err := deleteRoundFixtures(ctx, repos, id); err != nil {
				return err
			}
			if err := writeRoundFixtures(ctx, repos, id, r); err != nil {
				return err
			}
		}
		return nil
	})
}

// deleteRoundFixtures soft-deletes every match (and its club_matches) in a round.
func deleteRoundFixtures(ctx context.Context, repos application.WriteRepos, roundID int) error {
	matches, err := repos.Matches.FindByRoundID(ctx, roundID)
	if err != nil {
		return err
	}
	for _, m := range matches {
		if err := repos.ClubMatches.SoftDeleteByMatchID(ctx, m.ID); err != nil {
			return err
		}
	}
	return repos.Matches.SoftDeleteByRoundID(ctx, roundID)
}

// writeRoundFixtures creates the round's home-vs-away matches and single-sided
// bye matches.
func writeRoundFixtures(ctx context.Context, repos application.WriteRepos, roundID int, r RoundSpec) error {
	for _, f := range r.Fixtures {
		m, err := repos.Matches.Create(ctx, roundID, nil)
		if err != nil {
			return err
		}
		if _, err := repos.ClubMatches.Create(ctx, m.ID, f.HomeClubSeasonID, "home"); err != nil {
			return err
		}
		if _, err := repos.ClubMatches.Create(ctx, m.ID, f.AwayClubSeasonID, "away"); err != nil {
			return err
		}
	}
	byeStyle := "bye"
	for _, cs := range r.Byes {
		m, err := repos.Matches.Create(ctx, roundID, &byeStyle)
		if err != nil {
			return err
		}
		if _, err := repos.ClubMatches.Create(ctx, m.ID, cs, "bye"); err != nil {
			return err
		}
	}
	// A superbye is one match containing a club_match for every participating club.
	if len(r.Superbye) > 0 {
		superbyeStyle := "superbye"
		m, err := repos.Matches.Create(ctx, roundID, &superbyeStyle)
		if err != nil {
			return err
		}
		for _, cs := range r.Superbye {
			if _, err := repos.ClubMatches.Create(ctx, m.ID, cs, "superbye"); err != nil {
				return err
			}
		}
	}
	return nil
}
