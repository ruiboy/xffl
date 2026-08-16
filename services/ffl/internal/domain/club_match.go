package domain

import (
	"context"
	"fmt"
	"strings"
)

// ClubMatchDataStatus tracks whether a team has been submitted and scored.
type ClubMatchDataStatus string

const (
	ClubMatchDataNoData    ClubMatchDataStatus = "no_data"
	ClubMatchDataSubmitted ClubMatchDataStatus = "submitted"
	ClubMatchDataFinal     ClubMatchDataStatus = "final"
)

type ClubMatch struct {
	ID            int
	MatchID       int
	ClubSeasonID  int
	Side          string // 'home' | 'away' | 'bye' | 'superbye'
	DataStatus    ClubMatchDataStatus
	Notes         *string
	StoredScore   int
	PlayerMatches []PlayerMatch
}

// TeamSubmitted is the domain event raised when a club match team is submitted.
// Yeah, this is not adding a lot of value right now, but it demonstrates the
// concept of domain events (cf. integration events).
type TeamSubmitted struct {
	ClubMatchID int
}

// SubmitTeam validates the player list against the season's rules, replaces the
// club match's player matches, and transitions data_status to submitted. Returns
// a TeamSubmitted domain event on success.
func (cm *ClubMatch) SubmitTeam(players []PlayerMatch, rules Rules) (TeamSubmitted, error) {
	if err := rules.Validate(players); err != nil {
		return TeamSubmitted{}, err
	}
	cm.PlayerMatches = players
	cm.DataStatus = ClubMatchDataSubmitted
	return TeamSubmitted{ClubMatchID: cm.ID}, nil
}

// containsPosition checks whether a comma-separated positions string contains pos.
func containsPosition(positions string, pos Position) bool {
	for _, p := range strings.Split(positions, ",") {
		if Position(strings.TrimSpace(p)) == pos {
			return true
		}
	}
	return false
}

// SubPairing is an explicit TM declaration pairing a starter being replaced with the bench
// player covering them.
type SubPairing struct {
	ReplacedPMID  int // starter going out (must be DNP for subs; any status for interchange)
	ReplacingPMID int // bench player coming in
}

// DeclareSubs records explicit TM sub/interchange decisions for this club match.
// subs lists starter→bench pairings for substitutions (replaced starter must be DNP).
// interchange is an optional single starter→bench pairing for the interchange slot.
// All prior sub/interchange statuses are reset to named before applying new declarations.
// Returns the full player match slice with updated statuses.
func (cm ClubMatch) DeclareSubs(subs []SubPairing, interchange *SubPairing) ([]PlayerMatch, error) {
	if cm.DataStatus == ClubMatchDataFinal {
		return nil, fmt.Errorf("club match %d is already final", cm.ID)
	}

	pmByID := make(map[int]*PlayerMatch, len(cm.PlayerMatches))
	updated := make([]PlayerMatch, len(cm.PlayerMatches))
	copy(updated, cm.PlayerMatches)
	for i := range updated {
		pmByID[updated[i].ID] = &updated[i]
	}

	// Validate subs before mutating anything.
	for _, pair := range subs {
		replaced, ok := pmByID[pair.ReplacedPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", pair.ReplacedPMID)
		}
		if replaced.BackupPositions != nil {
			return nil, fmt.Errorf("player_match %d is a bench player, cannot be subbed out", pair.ReplacedPMID)
		}
		if replaced.AFLStatus == nil || *replaced.AFLStatus != AFLStatusDNP {
			return nil, fmt.Errorf("player_match %d is not DNP, cannot be subbed out", pair.ReplacedPMID)
		}
		replacing, ok := pmByID[pair.ReplacingPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", pair.ReplacingPMID)
		}
		if replacing.BackupPositions == nil {
			return nil, fmt.Errorf("player_match %d is not a bench player, cannot be subbed in", pair.ReplacingPMID)
		}
	}
	if interchange != nil {
		replaced, ok := pmByID[interchange.ReplacedPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", interchange.ReplacedPMID)
		}
		if replaced.BackupPositions != nil {
			return nil, fmt.Errorf("player_match %d is a bench player, cannot be interchanged out", interchange.ReplacedPMID)
		}
		replacing, ok := pmByID[interchange.ReplacingPMID]
		if !ok {
			return nil, fmt.Errorf("player_match %d not found in club match", interchange.ReplacingPMID)
		}
		if replacing.BackupPositions == nil {
			return nil, fmt.Errorf("player_match %d is not a bench player, cannot be interchanged in", interchange.ReplacingPMID)
		}
	}

	// Reset all prior sub/interchange statuses to named.
	// Bench players that were activated also have their inherited position cleared.
	named := PlayerMatchStatusNamed
	for i := range updated {
		if s := updated[i].Status; s != nil {
			switch *s {
			case PlayerMatchStatusSubbedOut, PlayerMatchStatusInterchangedOut:
				updated[i].Status = &named
			case PlayerMatchStatusSubbedIn, PlayerMatchStatusInterchangedIn:
				updated[i].Status = &named
				updated[i].Position = nil // clear position inherited from the starter
			}
		}
	}

	// Apply new declarations.
	for _, pair := range subs {
		subbedOut := PlayerMatchStatusSubbedOut
		subbedIn := PlayerMatchStatusSubbedIn
		replaced := pmByID[pair.ReplacedPMID]
		replacing := pmByID[pair.ReplacingPMID]
		replaced.Status = &subbedOut
		// Bench player inherits the starter's position so scoring works correctly.
		starterPos := *replaced.Position
		replacing.Status = &subbedIn
		replacing.Position = &starterPos
	}
	if interchange != nil {
		interchangedOut := PlayerMatchStatusInterchangedOut
		interchangedIn := PlayerMatchStatusInterchangedIn
		replaced := pmByID[interchange.ReplacedPMID]
		replacing := pmByID[interchange.ReplacingPMID]
		replaced.Status = &interchangedOut
		// Bench player scores at the declared interchange position (always one of their
		// backup positions and validated by SubmitTeam). Fall back to the displaced
		// starter's position if InterchangePosition is somehow unset.
		var icPos Position
		if replacing.InterchangePosition != nil {
			icPos = Position(*replacing.InterchangePosition)
		} else {
			icPos = *replaced.Position
		}
		replacing.Status = &interchangedIn
		replacing.Position = &icPos
	}

	return updated, nil
}

// SubstitutionKind distinguishes an interchange suggestion from a substitution suggestion.
type SubstitutionKind string

const (
	SubstitutionKindInterchange SubstitutionKind = "interchange"
	SubstitutionKindSub         SubstitutionKind = "sub"
)

// SuggestedSubstitution is a recommended TM action: either an interchange (bench outscores
// a starter at the interchange position) or a sub (starter is DNP and bench covers them).
type SuggestedSubstitution struct {
	Kind          SubstitutionKind
	ReplacedPMID  int
	ReplacingPMID int
}

// SuggestedSubstitutions returns all recommended TM actions for this club match.
// Interchange: the interchange bench player strictly outscores the lowest-scoring active
// starter at the interchange position and has not yet been activated.
// Sub: a starter is DNP and a backup bench player covers their position, is not already
// activated, and has not also missed their AFL match.
// Both suggestions may be returned for the same bench player if applicable — the TM decides.
func (cm ClubMatch) SuggestedSubstitutions() []SuggestedSubstitution {
	var result []SuggestedSubstitution

	// ── Interchange suggestion ───────────────────────────────────────────────
	var icBench *PlayerMatch
	for i := range cm.PlayerMatches {
		pm := &cm.PlayerMatches[i]
		if pm.isBench() && pm.InterchangePosition != nil {
			icBench = pm
			break
		}
	}
	if icBench != nil && (icBench.Status == nil || *icBench.Status == PlayerMatchStatusNamed) {
		icPos := Position(*icBench.InterchangePosition)
		var lowestStarter *PlayerMatch
		for i := range cm.PlayerMatches {
			pm := &cm.PlayerMatches[i]
			if pm.isBench() || pm.Position == nil || *pm.Position != icPos {
				continue
			}
			if pm.Status != nil && *pm.Status == PlayerMatchStatusInterchangedOut {
				continue
			}
			if lowestStarter == nil || pm.Score < lowestStarter.Score {
				lowestStarter = pm
			}
		}
		if lowestStarter != nil && icBench.Score > lowestStarter.Score {
			result = append(result, SuggestedSubstitution{
				Kind:          SubstitutionKindInterchange,
				ReplacedPMID:  lowestStarter.ID,
				ReplacingPMID: icBench.ID,
			})
		}
	}

	// ── Sub suggestions ──────────────────────────────────────────────────────
	// Map each position to the backup bench player covering it (non-interchange, named, not DNP).
	benchByPos := make(map[Position]*PlayerMatch)
	for i := range cm.PlayerMatches {
		pm := &cm.PlayerMatches[i]
		if !pm.isBench() || pm.InterchangePosition != nil {
			continue
		}
		if pm.Status != nil && *pm.Status != PlayerMatchStatusNamed {
			continue
		}
		if pm.AFLStatus != nil && *pm.AFLStatus == AFLStatusDNP {
			continue
		}
		if pm.BackupPositions == nil {
			continue
		}
		for _, part := range strings.Split(*pm.BackupPositions, ",") {
			benchByPos[Position(strings.TrimSpace(part))] = pm
		}
	}
	for i := range cm.PlayerMatches {
		pm := &cm.PlayerMatches[i]
		if pm.isBench() || pm.Position == nil {
			continue
		}
		if pm.AFLStatus == nil || *pm.AFLStatus != AFLStatusDNP {
			continue
		}
		if pm.Status != nil && (*pm.Status == PlayerMatchStatusSubbedOut || *pm.Status == PlayerMatchStatusInterchangedOut) {
			continue
		}
		if bench, ok := benchByPos[*pm.Position]; ok {
			result = append(result, SuggestedSubstitution{
				Kind:          SubstitutionKindSub,
				ReplacedPMID:  pm.ID,
				ReplacingPMID: bench.ID,
			})
		}
	}

	return result
}

// Score computes the total fantasy score for this club match.
// Bench players (BackupPositions set) only score when explicitly activated (subbed_in or
// interchanged_in). Starters score when named or status is nil; excluded starters
// (subbed_out, interchanged_out) do not score.
func (cm ClubMatch) Score() int {
	total := 0
	for _, pm := range cm.PlayerMatches {
		if pm.Position == nil {
			continue
		}
		if pm.isBench() {
			if pm.Status != nil && (*pm.Status == PlayerMatchStatusSubbedIn || *pm.Status == PlayerMatchStatusInterchangedIn) {
				total += pm.Score
			}
			continue
		}
		if pm.Status == nil || *pm.Status == PlayerMatchStatusNamed ||
			*pm.Status == PlayerMatchStatusSubbedIn || *pm.Status == PlayerMatchStatusInterchangedIn {
			total += pm.Score
		}
	}
	return total
}

// Notes reference scores: club_match.notes stores space-separated "source:value"
// tokens (e.g. "posted:452 spreadsheet:450"). References coexist — the forum
// import records "posted", the fixture import records "spreadsheet", and neither
// clobbers the other. drv_score stays the evaluated truth; these are references.

// UpsertNote sets source:value in the club_match's notes, replacing an existing
// token for that source and preserving every other token, and returns the merged
// string. A brand-new source is appended; non-"key:value" text is left in place.
func (cm ClubMatch) UpsertNote(source string, value int) string {
	existing := ""
	if cm.Notes != nil {
		existing = *cm.Notes
	}
	newTok := fmt.Sprintf("%s:%d", source, value)
	fields := strings.Fields(existing)
	for i, f := range fields {
		if k, ok := noteTokenKey(f); ok && k == source {
			fields[i] = newTok
			return strings.Join(fields, " ")
		}
	}
	return strings.TrimSpace(strings.Join(append(fields, newTok), " "))
}

// noteTokenKey returns the key of a "key:value" token, or ok=false otherwise.
func noteTokenKey(f string) (string, bool) {
	i := strings.IndexByte(f, ':')
	if i <= 0 {
		return "", false
	}
	return f[:i], true
}

type ClubMatchRepository interface {
	FindByMatchID(ctx context.Context, matchID int) ([]ClubMatch, error)
	FindByID(ctx context.Context, id int) (ClubMatch, error)
	FindByIDs(ctx context.Context, ids []int) (map[int]ClubMatch, error)
	UpdateScore(ctx context.Context, id int, score int) error
	UpdatePremiershipPoints(ctx context.Context, id int, points int) error
	UpdateNotes(ctx context.Context, id int, notes string) error
	UpdateDataStatus(ctx context.Context, id int, status ClubMatchDataStatus) error
	CountFinalByMatchID(ctx context.Context, matchID int) (int, error)
	GetRulesID(ctx context.Context, clubMatchID int) (string, error)
	Create(ctx context.Context, matchID int, clubSeasonID int, side string) (ClubMatch, error)
	UpdateSide(ctx context.Context, id int, side string) error
	// A club dropped from a match is removed outright: uni_ffl_club_match ignores
	// deleted_at, so a tombstone would block that club rejoining the match.
	DeleteByID(ctx context.Context, id int) error
	FindFinalClubMatchesBySeasonID(ctx context.Context, seasonID int) ([]ScoredClubMatch, error)
}
