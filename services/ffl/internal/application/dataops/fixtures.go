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
//   - An existing round without teams is reconciled match-by-match against the
//     spec — unchanged matches keep their rows, edited ones are updated in place,
//     and only genuinely added/removed matches are inserted/soft-deleted. Round
//     metadata is updated only when it changed.
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
			old := existingByID[id]
			if old.Name != r.Name || old.AFLRoundID != r.AFLRoundID || old.Type != r.Type {
				if err := repos.Rounds.Update(ctx, id, r.Name, r.AFLRoundID, r.Type); err != nil {
					return err
				}
			}
			if err := reconcileRoundMatches(ctx, repos, id, r.Matches); err != nil {
				return err
			}
		}
		return nil
	})
}

// deleteRoundFixtures removes every match in a round; club_matches follow via
// ON DELETE CASCADE.
func deleteRoundFixtures(ctx context.Context, repos application.WriteRepos, roundID int) error {
	return repos.Matches.DeleteByRoundID(ctx, roundID)
}

// writeRoundMatches creates every match of a brand-new round.
func writeRoundMatches(ctx context.Context, repos application.WriteRepos, roundID int, r RoundSpec) error {
	for _, ms := range r.Matches {
		if err := createMatch(ctx, repos, roundID, ms); err != nil {
			return err
		}
	}
	return nil
}

// createMatch inserts one match plus a club_match per participating club.
func createMatch(ctx context.Context, repos application.WriteRepos, roundID int, ms MatchSpec) error {
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
	return nil
}

// existingMatch is a round's stored match with its club_matches (home-first for
// versus), used to reconcile against the desired spec without churning rows.
type existingMatch struct {
	matchID int
	style   domain.MatchStyle
	cms     []domain.ClubMatch
}

// reconcileRoundMatches brings a round's matches into line with the spec while
// preserving unchanged rows. It pairs each desired match with a stored one —
// preferring an identical match (no writes at all), then any stored match of the
// same style (reused in place) — creates the leftovers, and soft-deletes any
// stored match the spec no longer wants. This replaces the previous
// delete-everything-then-reinsert approach, so a save that changes nothing (or
// only edits a pairing) no longer churns every match and club_match row.
func reconcileRoundMatches(ctx context.Context, repos application.WriteRepos, roundID int, desired []MatchSpec) error {
	matches, err := repos.Matches.FindByRoundID(ctx, roundID)
	if err != nil {
		return err
	}
	existing := make([]existingMatch, len(matches))
	for i, m := range matches {
		cms, err := repos.ClubMatches.FindByMatchID(ctx, m.ID)
		if err != nil {
			return err
		}
		existing[i] = existingMatch{matchID: m.ID, style: m.MatchStyle, cms: cms}
	}
	used := make([]bool, len(existing))
	paired := make([]bool, len(desired))

	// Pass 1: identical matches — reuse with zero writes.
	for di, d := range desired {
		for ei := range existing {
			if used[ei] || !matchesExactly(existing[ei], d) {
				continue
			}
			used[ei], paired[di] = true, true
			break
		}
	}
	// Pass 2: reuse a stored match of the same style, editing its club_matches.
	for di, d := range desired {
		if paired[di] {
			continue
		}
		for ei := range existing {
			if used[ei] || existing[ei].style != d.Style {
				continue
			}
			if err := reconcileClubMatches(ctx, repos, existing[ei], d); err != nil {
				return err
			}
			used[ei], paired[di] = true, true
			break
		}
	}
	// Pass 3: create desired matches with no stored counterpart.
	for di, d := range desired {
		if paired[di] {
			continue
		}
		if err := createMatch(ctx, repos, roundID, d); err != nil {
			return err
		}
	}
	// Pass 4: drop stored matches the spec no longer wants (club_matches cascade).
	for ei := range existing {
		if used[ei] {
			continue
		}
		if err := repos.Matches.DeleteByID(ctx, existing[ei].matchID); err != nil {
			return err
		}
	}
	return nil
}

// matchesExactly reports whether a stored match already equals the spec. Versus
// and bye compare club_seasons in order (home vs away is meaningful); a superbye
// compares membership as a set (its clubs are unordered).
func matchesExactly(ex existingMatch, d MatchSpec) bool {
	if ex.style != d.Style || len(ex.cms) != len(d.ClubSeasonIDs) {
		return false
	}
	if d.Style == domain.MatchStyleSuperbye {
		set := make(map[int]bool, len(ex.cms))
		for _, cm := range ex.cms {
			set[cm.ClubSeasonID] = true
		}
		for _, id := range d.ClubSeasonIDs {
			if !set[id] {
				return false
			}
		}
		return true
	}
	for i, cm := range ex.cms {
		if cm.ClubSeasonID != d.ClubSeasonIDs[i] {
			return false
		}
	}
	return true
}

// reconcileClubMatches edits a reused match's club_matches to match the spec,
// keyed on the club rather than the slot: a club that is staying keeps its row
// and only has its side rewritten, clubs no longer in the match are deleted, and
// newly added clubs are inserted.
//
// Keying on the club is what makes a home/away swap safe. Repointing rows by
// slot instead would rewrite club_season_id, and setting the home row to the
// club already sitting in the away row trips uni_ffl_club_match
// (club_season_id, match_id) mid-transaction. Swapping the sides of two rows
// touches no unique column at all.
func reconcileClubMatches(ctx context.Context, repos application.WriteRepos, ex existingMatch, d MatchSpec) error {
	desiredSide := make(map[int]string, len(d.ClubSeasonIDs))
	for i, cs := range d.ClubSeasonIDs {
		desiredSide[cs] = sideFor(d.Style, i)
	}
	kept := make(map[int]bool, len(ex.cms))
	for _, cm := range ex.cms {
		side, wanted := desiredSide[cm.ClubSeasonID]
		if !wanted {
			if err := repos.ClubMatches.DeleteByID(ctx, cm.ID); err != nil {
				return err
			}
			continue
		}
		kept[cm.ClubSeasonID] = true
		if cm.Side != side {
			if err := repos.ClubMatches.UpdateSide(ctx, cm.ID, side); err != nil {
				return err
			}
		}
	}
	for _, cs := range d.ClubSeasonIDs {
		if kept[cs] {
			continue
		}
		if _, err := repos.ClubMatches.Create(ctx, ex.matchID, cs, desiredSide[cs]); err != nil {
			return err
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
