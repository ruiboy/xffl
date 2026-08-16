package dataops

import (
	"context"
	"fmt"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// FixtureImportRound is one parsed round prepared for import: the reviewer has
// mapped it to an AFL round and a round type; the sheet supplies the pairings and
// the reference (spreadsheet) scores as ParsedFixture values. Club names are still
// as written — ImportFixtures resolves them to the season's club_seasons.
type FixtureImportRound struct {
	Name       string
	AFLRoundID int
	Type       domain.RoundType
	Fixtures   []application.ParsedFixture
}

// ImportFixturesParams is the confirmed input to ImportFixtures. Rounds is the
// complete set of minor rounds for the season: ImportFixtures reconciles the whole
// season fixture to it (via SaveFixtures), so omitting a round removes it.
type ImportFixturesParams struct {
	SeasonID int
	Rounds   []FixtureImportRound
}

// ImportFixturesResult reports what the import did. When Unresolved is non-empty
// nothing is written — every club name in the sheet must map to a club_season
// first (the reviewer registers the club or fixes the spelling).
type ImportFixturesResult struct {
	RoundsCreated int
	ScoresWritten int
	Unresolved    []string
}

// refScore is a reference (spreadsheet) score to stamp on a club_match once the
// fixtures exist. Keyed by round name + club_season because that pair is unique.
type refScore struct {
	RoundName    string
	ClubSeasonID int
	Score        int
}

// ImportFixtures builds a season's rounds and fixtures from the parsed sheet and
// records the reference scores in club_match.notes ("spreadsheet:NN"). Fixture
// creation reuses SaveFixtures — the one fixture write path — so byes, immutability
// of rounds with entered teams, and reconciliation all behave identically to the
// manual builder. The bye club each round (the season club not named in any
// fixture) gets a bye match; superbye and finals nuance is left to the manual
// builder afterward.
func (b *Builder) ImportFixtures(ctx context.Context, params ImportFixturesParams) (ImportFixturesResult, error) {
	seasonCS, nameToCS, err := b.loadSeasonClubs(ctx, params.SeasonID)
	if err != nil {
		return ImportFixturesResult{}, err
	}

	specs, refs, unresolved := planFixtureImport(params.Rounds, seasonCS, nameToCS)
	if len(unresolved) > 0 {
		// Refuse a partial import: a mis-spelled or unregistered club would silently
		// drop a fixture and skew the ladder.
		return ImportFixturesResult{Unresolved: unresolved}, nil
	}

	if err := b.SaveFixtures(ctx, params.SeasonID, specs); err != nil {
		return ImportFixturesResult{}, err
	}

	written, err := b.writeReferenceScores(ctx, params.SeasonID, refs)
	if err != nil {
		return ImportFixturesResult{}, err
	}
	return ImportFixturesResult{RoundsCreated: len(specs), ScoresWritten: written}, nil
}

// loadSeasonClubs returns the season's club_season ids and a normalised
// club-name → club_season_id lookup for resolving the sheet's club names.
func (b *Builder) loadSeasonClubs(ctx context.Context, seasonID int) ([]int, map[string]int, error) {
	var ids []int
	nameToCS := map[string]int{}
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		css, err := repos.ClubSeasons.FindBySeasonID(ctx, seasonID)
		if err != nil {
			return err
		}
		clubIDs := make([]int, len(css))
		for i, cs := range css {
			clubIDs[i] = cs.ClubID
		}
		clubs, err := repos.Clubs.FindByIDs(ctx, clubIDs)
		if err != nil {
			return err
		}
		for _, cs := range css {
			ids = append(ids, cs.ID)
			if club, ok := clubs[cs.ClubID]; ok {
				nameToCS[canonicalClub(club.Name)] = cs.ID
			}
		}
		return nil
	})
	return ids, nameToCS, err
}

// planFixtureImport turns resolved rounds into the RoundSpecs SaveFixtures wants
// and the reference scores to stamp afterward. It is pure so the resolution and
// bye-inference rules are testable without a database. Any club name that does not
// map to a club_season is collected in unresolved (deduplicated, first-seen order).
func planFixtureImport(rounds []FixtureImportRound, seasonClubSeasonIDs []int, nameToCS map[string]int) ([]RoundSpec, []refScore, []string) {
	specs := make([]RoundSpec, 0, len(rounds))
	var refs []refScore
	var unresolved []string
	seenUnresolved := map[string]bool{}

	noteUnresolved := func(name string) {
		if !seenUnresolved[name] {
			seenUnresolved[name] = true
			unresolved = append(unresolved, name)
		}
	}

	for _, r := range rounds {
		spec := RoundSpec{Name: r.Name, AFLRoundID: r.AFLRoundID, Type: r.Type}
		used := map[int]bool{}
		for _, fx := range r.Fixtures {
			home, homeOK := nameToCS[canonicalClub(fx.HomeClub)]
			away, awayOK := nameToCS[canonicalClub(fx.AwayClub)]
			if !homeOK {
				noteUnresolved(fx.HomeClub)
			}
			if !awayOK {
				noteUnresolved(fx.AwayClub)
			}
			if !homeOK || !awayOK {
				continue
			}
			spec.Matches = append(spec.Matches, MatchSpec{Style: domain.MatchStyleVersus, ClubSeasonIDs: []int{home, away}})
			used[home], used[away] = true, true
			if fx.HomeScore != nil {
				refs = append(refs, refScore{RoundName: r.Name, ClubSeasonID: home, Score: *fx.HomeScore})
			}
			if fx.AwayScore != nil {
				refs = append(refs, refScore{RoundName: r.Name, ClubSeasonID: away, Score: *fx.AwayScore})
			}
		}
		// Home-and-away rounds give every club not playing a bye; finals do not —
		// clubs missing from a final are eliminated, not on a bye.
		if r.Type == domain.RoundTypeMinor {
			for _, cs := range seasonClubSeasonIDs {
				if !used[cs] {
					spec.Matches = append(spec.Matches, MatchSpec{Style: domain.MatchStyleBye, ClubSeasonIDs: []int{cs}})
				}
			}
		}
		specs = append(specs, spec)
	}
	return specs, refs, unresolved
}

// writeReferenceScores stamps each reference score onto its club_match once the
// fixtures exist, matching on (round name, club_season). Returns the count written.
func (b *Builder) writeReferenceScores(ctx context.Context, seasonID int, refs []refScore) (int, error) {
	if len(refs) == 0 {
		return 0, nil
	}
	scoreByKey := make(map[string]int, len(refs))
	for _, r := range refs {
		scoreByKey[refKey(r.RoundName, r.ClubSeasonID)] = r.Score
	}

	written := 0
	err := b.tx.WithTx(ctx, func(repos application.WriteRepos) error {
		rounds, err := repos.Rounds.FindBySeasonID(ctx, seasonID)
		if err != nil {
			return err
		}
		for _, rnd := range rounds {
			matches, err := repos.Matches.FindByRoundID(ctx, rnd.ID)
			if err != nil {
				return err
			}
			for _, m := range matches {
				cms, err := repos.ClubMatches.FindByMatchID(ctx, m.ID)
				if err != nil {
					return err
				}
				for _, cm := range cms {
					score, ok := scoreByKey[refKey(rnd.Name, cm.ClubSeasonID)]
					if !ok {
						continue
					}
					merged := cm.UpsertNote("spreadsheet", score)
					if err := repos.ClubMatches.UpdateNotes(ctx, cm.ID, merged); err != nil {
						return err
					}
					written++
				}
			}
		}
		return nil
	})
	return written, err
}

func refKey(roundName string, clubSeasonID int) string {
	return fmt.Sprintf("%s\x00%d", roundName, clubSeasonID)
}
