package dataops

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// MatchSpec is one match in a round: its style and the club_seasons taking part.
// The club_season ordering is meaningful for versus ([home, away]); a bye has one
// club, a superbye lists every participating club.
type MatchSpec struct {
	Style         domain.MatchStyle
	ClubSeasonIDs []int
}

// RoundSpec is the desired state of a single round. A nil RoundID means a new
// round; a set RoundID targets an existing round in the season. A round is just
// its list of matches — versus, bye and superbye are all matches, distinguished
// by Style.
type RoundSpec struct {
	RoundID    *int
	Name       string
	AFLRoundID int
	Type       domain.RoundType
	Matches    []MatchSpec
}

// FixtureRound is a round loaded for the builder: its matches, plus whether it is
// locked (has submitted teams, so its fixtures are immutable).
type FixtureRound struct {
	RoundID    int
	Name       string
	AFLRoundID int
	Type       domain.RoundType
	Locked     bool
	Matches    []MatchSpec
}

// LoadFixtures reads a season's rounds and their matches for the builder to edit.
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
				// FindByMatchID orders home before away, so versus ids read [home, away].
				cms, err := repos.ClubMatches.FindByMatchID(ctx, m.ID)
				if err != nil {
					return err
				}
				ids := make([]int, len(cms))
				for i, cm := range cms {
					ids[i] = cm.ClubSeasonID
				}
				fr.Matches = append(fr.Matches, MatchSpec{Style: m.MatchStyle, ClubSeasonIDs: ids})
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
//   - An existing round without teams is replaced wholesale (its matches are
//     rebuilt from the spec) and its metadata updated.
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
				if err := writeRoundMatches(ctx, repos, rnd.ID, r); err != nil {
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
			if err := writeRoundMatches(ctx, repos, id, r); err != nil {
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

// writeRoundMatches creates a round's matches, each with its style and a
// club_match per participating club.
func writeRoundMatches(ctx context.Context, repos application.WriteRepos, roundID int, r RoundSpec) error {
	for _, ms := range r.Matches {
		style := string(ms.Style)
		m, err := repos.Matches.Create(ctx, roundID, &style)
		if err != nil {
			return err
		}
		for i, cs := range ms.ClubSeasonIDs {
			if _, err := repos.ClubMatches.Create(ctx, m.ID, cs, sideFor(ms.Style, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// sideFor is the club_match side for the i-th club of a match: home/away for a
// versus match, otherwise the style itself ('bye' / 'superbye').
func sideFor(style domain.MatchStyle, index int) string {
	if style == domain.MatchStyleVersus {
		if index == 0 {
			return "home"
		}
		return "away"
	}
	return string(style)
}
