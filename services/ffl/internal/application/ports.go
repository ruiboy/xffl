package application

import "context"

// PlayerCandidate is a known player that can be matched against a parsed name.
type PlayerCandidate struct {
	PlayerID          int
	AFLPlayerID       int
	AFLPlayerSeasonID int    // AFL player_season handle; used by squad import to call AddPlayerToSeason. 0 when not sourced.
	Name              string
	Club              string // AFL club name from afl.club
}

// PlayerNameMatch is the result of resolving a parsed name against a candidate pool.
type PlayerNameMatch struct {
	Candidate  PlayerCandidate
	Confidence float64 // 0.0–1.0
}

// PlayerMatchStats holds the AFL stats for a single player match, returned by LookupPlayerMatch.
type PlayerMatchStats struct {
	ID             int
	Status         string
	Goals          int
	Kicks          int
	Handballs      int
	Marks          int
	Tackles        int
	Hitouts        int
	PlayerSeasonID int // populated by LookupPlayerMatchBySeasonRound; 0 when looked up by ID
}

// ByePlayerInfo holds bye status, eligibility, and season averages for a player.
type ByePlayerInfo struct {
	PlayerSeasonID int
	HasBye         bool
	PlayedLast     bool // true if player played in their club's most recent non-bye final match
	AvgGoals       float64
	AvgKicks       float64
	AvgHandballs   float64
	AvgMarks       float64
	AvgTackles     float64
	AvgHitouts     float64
}

// PlayerLookup fetches entities from the AFL service by ID, to return information required cross-service.
type PlayerLookup interface {
	// LookupPlayers fetches an AFL Player by ID.
	LookupPlayers(ctx context.Context, aflPlayerIDs []int) ([]PlayerCandidate, error)
	// LookupPlayerSeason fetches an AFL Player Season by ID.
	LookupPlayerSeason(ctx context.Context, aflPlayerSeasonID int) (int, error)
	// LookupPlayerMatch fetches AFL match stats for a batch of AFL player match IDs.
	LookupPlayerMatch(ctx context.Context, aflPlayerMatchIDs []int) ([]PlayerMatchStats, error)
	// LookupPlayerMatchBySeasonRound fetches AFL match stats for a set of AFL player_season IDs
	// within a specific AFL round. PlayerSeasonID is populated in each returned stat.
	LookupPlayerMatchBySeasonRound(ctx context.Context, aflPlayerSeasonIDs []int, aflRoundID int) ([]PlayerMatchStats, error)
	// LookupByeInfo returns bye status, eligibility, and season averages for a batch of
	// AFL player_season_ids for a given AFL round.
	LookupByeInfo(ctx context.Context, aflPlayerSeasonIDs []int, aflRoundID int) ([]ByePlayerInfo, error)
	// LookupPlayerSeasonsBySeasonID returns every AFL player_season in a season as a
	// candidate pool (with AFLPlayerSeasonID set), for fuzzy-matching imported names
	// such as a pasted squads thread against a whole season's players.
	LookupPlayerSeasonsBySeasonID(ctx context.Context, aflSeasonID int) ([]PlayerCandidate, error)
}

// PlayerResolver fuzzy-matches a parsed name (with optional club hint) against
// a caller-supplied candidate pool. Decoupled from the record type being matched.
type PlayerResolver interface {
	Resolve(ctx context.Context, name, clubHint string, candidates []PlayerCandidate) ([]PlayerNameMatch, error)
}

// TeamParser parses a raw forum post into structured player rows.
// The caller supplies the FFL team name and round number; the parser
// identifies player names, positions, and optional scores.
type TeamParser interface {
	Parse(ctx context.Context, teamName, post string) ([]ParsedPlayerRow, error)
}

// ForumProcessor turns raw forum capture (post content HTML + author) into
// parseable text and team identity, then parses it. Implemented by the forum
// adapter, keeping HTML/forum specifics out of the application layer.
type ForumProcessor interface {
	TeamParser
	// HTMLToText converts a post's content HTML into newline-separated text.
	HTMLToText(html string) string
	// TeamForAuthor returns the parser format for a forum author, or "" if unknown.
	TeamForAuthor(author string) string
	// DetectFormat guesses the parser format from a post's content, or "" if none
	// is recognised — used when the author is unknown (a team posted on another's
	// behalf), so attribution can still come from the post itself.
	DetectFormat(post string) string
}

// ParsedPlayerRow is one player line extracted from a forum post.
type ParsedPlayerRow struct {
	Name                string
	ClubHint            string // AFL club code as written in the post (e.g. "Geel", "WB")
	Position            string // primary position ("goals", "kicks", …, "bench")
	BackupPositions     string // comma-separated, bench players only
	InterchangePosition string // bench players with interchange designation
	Score               *int   // nil if not present in the post
	Notes               string
}

// SquadThreadParser parses a squads thread — many clubs, ~30 members each — into
// per-club squads. This is a distinct forum format from the four team-submission
// formats handled by TeamParser.
type SquadThreadParser interface {
	ParseSquads(ctx context.Context, text string) ([]ParsedSquad, error)
}

// ParsedSquad is one club's squad as read from the thread. ClubName is the FFL
// club name exactly as written in the header line — resolving it to a
// club_season is a later step, not the parser's job.
type ParsedSquad struct {
	ClubName string // FFL club name as written (e.g. "Cheetahs", "RUIBOYS")
	Members  []ParsedSquadMember
}

// ParsedSquadMember is one player line from a squads thread:
// "<rank> <name> <price> <club>", e.g. "1 Darcy Fogarty 0.6 Adel".
type ParsedSquadMember struct {
	Rank      int
	Name      string
	ClubHint  string // AFL club code as written in the thread (e.g. "Adel", "WB")
	CostCents *int   // squad price in cents ("0.6" → 60); nil if absent/unparseable
}

// FixtureSheetParser parses a pasted season fixture sheet into rounds and their
// fixtures, carrying the reference (spreadsheet) scores. Resolving club names to
// club_seasons and creating rounds/matches is the operation's job, not the parser's.
type FixtureSheetParser interface {
	ParseFixtures(ctx context.Context, text string) ([]ParsedFixtureRound, error)
}

// ParsedFixtureRound is one round from the sheet. Round is the printed round
// number (0 for a labelled finals round with no number); Label carries any text
// beside the number or a finals name ("Super Bye", "Semi-Final", "Grand Final").
type ParsedFixtureRound struct {
	Round    int
	Label    string
	Fixtures []ParsedFixture
}

// ParsedFixture is one fixture line: two clubs and their reference scores as
// written. A score is nil when the sheet leaves it blank (e.g. an unplayed final).
// Club names are exactly as written; placeholder ("n/a") lines are not emitted.
type ParsedFixture struct {
	HomeClub  string
	HomeScore *int
	AwayClub  string
	AwayScore *int
}
