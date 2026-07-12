package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// AfltablesSource is the source key used in the player xref (afl.dataops_player_source)
// for decisions made during the historical import.
const AfltablesSource = "afltables"

// A fuzzy match at or above this confidence is logged as a near-miss when a new
// player is auto-created, so a genuine spelling-variant can be reviewed/relinked.
const historicalNearMissThreshold = 0.75

// HistoricalRow is one player-match line to ingest (mapped from the CSV by the
// CLI). Defined here so the use case does not depend on the infrastructure
// adapter that produces it.
type HistoricalRow struct {
	Round     string
	Date      string // ISO "2006-01-02"
	Venue     string
	HomeClub  string
	AwayClub  string
	Club      string
	Player    string
	Kicks     int
	Marks     int
	Handballs int
	Goals     int
	Behinds   int
	Hitouts   int
	Tackles   int
}

// PlayerRef identifies an existing afl.player.
type PlayerRef struct {
	ID   int
	Name string
}

// PlayerMatchInput is the stat line written for a resolved player in a match.
type PlayerMatchInput struct {
	ClubMatchID    int
	PlayerSeasonID int
	Kicks          int
	Marks          int
	Handballs      int
	Goals          int
	Behinds        int
	Hitouts        int
	Tackles        int
}

// HistoricalRepo is the persistence port for the historical import — idempotent
// get-or-create over the season→player_match scaffold, plus name/xref lookups.
type HistoricalRepo interface {
	UpsertLeague(ctx context.Context, name string) (int, error)
	GetOrCreateSeason(ctx context.Context, leagueID int, name string) (int, error)
	GetOrCreateRound(ctx context.Context, seasonID int, name string) (int, error)
	UpsertClub(ctx context.Context, name string) (int, error)
	GetOrCreateClubSeason(ctx context.Context, clubID, seasonID int) (int, error)
	FindMatchByRoundAndHomeClubSeason(ctx context.Context, roundID, homeClubSeasonID int) (int, bool, error)
	InsertMatch(ctx context.Context, roundID int, venue string, startDt time.Time) (int, error)
	UpsertClubMatch(ctx context.Context, matchID, clubSeasonID int, side string) (int, error)
	FindPlayersByExactName(ctx context.Context, name string) ([]PlayerRef, error)
	ClubsForNamedPlayers(ctx context.Context, name string) (map[int][]string, error)
	AllPlayers(ctx context.Context) ([]PlayerRef, error)
	PlayerSeasonYears(ctx context.Context, playerID int) ([]int, error)
	CreatePlayer(ctx context.Context, name string) (int, error)
	GetOrCreatePlayerSeason(ctx context.Context, playerID, clubSeasonID int) (int, error)
	UpsertPlayerMatch(ctx context.Context, p PlayerMatchInput) error
	FindXref(ctx context.Context, source, season, club, player string) (int, bool, error)
	StoreXref(ctx context.Context, source, season, club, player string, playerSeasonID int) error
}

// PlayerChoice is a candidate presented to the operator for an ambiguous name.
type PlayerChoice struct {
	PlayerID   int
	Name       string
	Confidence float64 // 0 for exact-name / gap prompts
	Detail     string  // e.g. "played 2004–2015" — context for the operator
}

// PlayerPrompter asks the operator to resolve an ambiguous player. It returns
// the chosen existing player ID, or 0 to create a new player.
type PlayerPrompter interface {
	Choose(ctx context.Context, name, club, season string, candidates []PlayerChoice) (playerID int, err error)
}

// HistoricalReviewLogger records auto-decisions for later audit.
type HistoricalReviewLogger interface {
	NewPlayer(season, club, name string, playerID int)
	NearMiss(season, club, name, candidateName string, confidence float64)
	// Gap records a same-name match auto-linked across a season gap (missed
	// seasons between this row's year and the player's known career) — the audit
	// trail for possible same-name-different-player merges.
	Gap(name string, year, missedSeasons int, existingSpan string, playerID int)
}

// ImportSeasonSummary reports what a season import did.
type ImportSeasonSummary struct {
	Season        int
	Matches       int
	PlayerMatches int
	NewPlayers    int
	Prompted      int
	Gaps          int // same-name matches auto-linked across a season gap (logged)
}

// HistoricalImporter ingests one season of afltables CSV rows into the AFL
// schema, resolving player identity interactively where ambiguous.
type HistoricalImporter struct {
	repo     HistoricalRepo
	resolver PlayerResolver
	prompter PlayerPrompter
	log      HistoricalReviewLogger
}

func NewHistoricalImporter(repo HistoricalRepo, resolver PlayerResolver, prompter PlayerPrompter, log HistoricalReviewLogger) *HistoricalImporter {
	return &HistoricalImporter{repo: repo, resolver: resolver, prompter: prompter, log: log}
}

// ImportSeason ingests all rows for one season. Idempotent: re-running upserts.
func (h *HistoricalImporter) ImportSeason(ctx context.Context, year int, rows []HistoricalRow) (ImportSeasonSummary, error) {
	sum := ImportSeasonSummary{Season: year}

	leagueID, err := h.repo.UpsertLeague(ctx, "AFL")
	if err != nil {
		return sum, fmt.Errorf("upsert league: %w", err)
	}
	seasonName := fmt.Sprintf("AFL %d", year)
	seasonID, err := h.repo.GetOrCreateSeason(ctx, leagueID, seasonName)
	if err != nil {
		return sum, fmt.Errorf("get-or-create season: %w", err)
	}

	// Fuzzy-match pool; grows as new players are created this run.
	allPlayers, err := h.repo.AllPlayers(ctx)
	if err != nil {
		return sum, fmt.Errorf("load players: %w", err)
	}

	clubSeasonID := map[string]int{}
	ensureClubSeason := func(clubName string) (int, error) {
		if id, ok := clubSeasonID[clubName]; ok {
			return id, nil
		}
		clubID, err := h.repo.UpsertClub(ctx, clubName)
		if err != nil {
			return 0, fmt.Errorf("upsert club %q: %w", clubName, err)
		}
		csID, err := h.repo.GetOrCreateClubSeason(ctx, clubID, seasonID)
		if err != nil {
			return 0, fmt.Errorf("club season %q: %w", clubName, err)
		}
		clubSeasonID[clubName] = csID
		return csID, nil
	}

	// Group rows into matches (a round's home-vs-away fixture).
	type matchKey struct{ round, home, away string }
	groups := map[matchKey][]HistoricalRow{}
	var order []matchKey
	for _, r := range rows {
		k := matchKey{r.Round, r.HomeClub, r.AwayClub}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], r)
	}

	// player_season id per external (club|player) for this run.
	psCache := map[string]int{}

	for _, k := range order {
		grp := groups[k]
		roundID, err := h.repo.GetOrCreateRound(ctx, seasonID, roundName(k.round))
		if err != nil {
			return sum, fmt.Errorf("round %q: %w", k.round, err)
		}
		homeCS, err := ensureClubSeason(k.home)
		if err != nil {
			return sum, err
		}
		awayCS, err := ensureClubSeason(k.away)
		if err != nil {
			return sum, err
		}

		matchID, found, err := h.repo.FindMatchByRoundAndHomeClubSeason(ctx, roundID, homeCS)
		if err != nil {
			return sum, fmt.Errorf("find match: %w", err)
		}
		if !found {
			matchID, err = h.repo.InsertMatch(ctx, roundID, grp[0].Venue, parseDate(grp[0].Date))
			if err != nil {
				return sum, fmt.Errorf("insert match: %w", err)
			}
		}
		homeCM, err := h.repo.UpsertClubMatch(ctx, matchID, homeCS, "home")
		if err != nil {
			return sum, fmt.Errorf("home club match: %w", err)
		}
		awayCM, err := h.repo.UpsertClubMatch(ctx, matchID, awayCS, "away")
		if err != nil {
			return sum, fmt.Errorf("away club match: %w", err)
		}
		sum.Matches++

		for _, r := range grp {
			psID, err := h.resolvePlayerSeason(ctx, year, r, seasonName, ensureClubSeason, psCache, &allPlayers, &sum)
			if err != nil {
				return sum, err
			}
			clubMatchID := homeCM
			if r.Club == k.away {
				clubMatchID = awayCM
			}
			if err := h.repo.UpsertPlayerMatch(ctx, PlayerMatchInput{
				ClubMatchID: clubMatchID, PlayerSeasonID: psID,
				Kicks: r.Kicks, Marks: r.Marks, Handballs: r.Handballs,
				Goals: r.Goals, Behinds: r.Behinds, Hitouts: r.Hitouts, Tackles: r.Tackles,
			}); err != nil {
				return sum, fmt.Errorf("upsert player_match for %q: %w", r.Player, err)
			}
			sum.PlayerMatches++
		}
	}
	return sum, nil
}

// resolvePlayerSeason resolves a row's player to a player_season_id, creating
// the player and/or player_season as needed, and caching the decision.
func (h *HistoricalImporter) resolvePlayerSeason(
	ctx context.Context, year int, r HistoricalRow, seasonName string,
	ensureClubSeason func(string) (int, error), psCache map[string]int,
	allPlayers *[]PlayerRef, sum *ImportSeasonSummary,
) (int, error) {
	cacheKey := r.Club + "|" + r.Player
	if id, ok := psCache[cacheKey]; ok {
		return id, nil
	}
	csID, err := ensureClubSeason(r.Club)
	if err != nil {
		return 0, err
	}

	// Previously-recorded decision (also makes re-runs non-interactive).
	if id, found, err := h.repo.FindXref(ctx, AfltablesSource, seasonName, r.Club, r.Player); err != nil {
		return 0, fmt.Errorf("xref lookup: %w", err)
	} else if found {
		psCache[cacheKey] = id
		return id, nil
	}

	playerID, err := h.resolvePlayerID(ctx, year, r, seasonName, allPlayers, sum)
	if err != nil {
		return 0, err
	}
	psID, err := h.repo.GetOrCreatePlayerSeason(ctx, playerID, csID)
	if err != nil {
		return 0, fmt.Errorf("player season for %q: %w", r.Player, err)
	}
	if err := h.repo.StoreXref(ctx, AfltablesSource, seasonName, r.Club, r.Player, psID); err != nil {
		return 0, fmt.Errorf("store xref: %w", err)
	}
	psCache[cacheKey] = psID
	return psID, nil
}

// resolvePlayerID maps a name to an afl.player id: exact-unique auto-links,
// no-match auto-creates (logging any near-miss), and ambiguity (duplicate exact
// names, or a high-confidence fuzzy match) is prompted.
func (h *HistoricalImporter) resolvePlayerID(
	ctx context.Context, year int, r HistoricalRow, seasonName string,
	allPlayers *[]PlayerRef, sum *ImportSeasonSummary,
) (int, error) {
	exact, err := h.repo.FindPlayersByExactName(ctx, r.Player)
	if err != nil {
		return 0, fmt.Errorf("exact lookup for %q: %w", r.Player, err)
	}
	if len(exact) == 1 {
		// One same-named player exists. If this season is adjacent to their known
		// career it's the same person (auto-link); a gap suggests a different
		// player sharing the name, so prompt.
		years, err := h.repo.PlayerSeasonYears(ctx, exact[0].ID)
		if err != nil {
			return 0, fmt.Errorf("season years for %q: %w", r.Player, err)
		}
		if yearConnected(year, years) {
			return exact[0].ID, nil
		}
		// Season gap: overwhelmingly the same player returning after a layoff, so
		// auto-link — but log it so genuine same-name-different-player collisions
		// can be found and split afterward.
		h.log.Gap(r.Player, year, nearestGap(year, years)-1, spanDetail(years), exact[0].ID)
		sum.Gaps++
		return exact[0].ID, nil
	}
	if len(exact) > 1 {
		// Multiple players share this name. If exactly one of them played for the
		// row's club, that disambiguates it (e.g. the two Bailey Williamses split
		// by Bulldogs/West Coast) — auto-link. Otherwise prompt.
		clubs, err := h.repo.ClubsForNamedPlayers(ctx, r.Player)
		if err != nil {
			return 0, fmt.Errorf("clubs for %q: %w", r.Player, err)
		}
		var byClub []PlayerRef
		for _, p := range exact {
			if containsFold(clubs[p.ID], r.Club) {
				byClub = append(byClub, p)
			}
		}
		if len(byClub) == 1 {
			return byClub[0].ID, nil
		}
		choices := make([]PlayerChoice, len(exact))
		for i, p := range exact {
			choices[i] = PlayerChoice{PlayerID: p.ID, Name: p.Name, Detail: spanDetail(mustYears(ctx, h.repo, p.ID))}
		}
		return h.promptOrCreate(ctx, r, seasonName, choices, allPlayers, sum)
	}

	// No exact match: create a new player. Never silently link a fuzzy match —
	// instead log a near-miss (any similar existing name) for post-hoc review,
	// so a genuine spelling-variant split can be relinked afterward.
	best := h.fuzzyBest(ctx, r.Player, r.Club, *allPlayers)
	id, err := h.createPlayer(ctx, seasonName, r.Club, r.Player, allPlayers, sum)
	if err != nil {
		return 0, err
	}
	if best.PlayerID != 0 && best.Confidence >= historicalNearMissThreshold {
		h.log.NearMiss(seasonName, r.Club, r.Player, best.Name, best.Confidence)
	}
	return id, nil
}

// promptOrCreate asks the operator; a 0 result means create a new player.
func (h *HistoricalImporter) promptOrCreate(
	ctx context.Context, r HistoricalRow, seasonName string, choices []PlayerChoice,
	allPlayers *[]PlayerRef, sum *ImportSeasonSummary,
) (int, error) {
	sum.Prompted++
	chosen, err := h.prompter.Choose(ctx, r.Player, r.Club, seasonName, choices)
	if err != nil {
		return 0, fmt.Errorf("prompt for %q: %w", r.Player, err)
	}
	if chosen != 0 {
		return chosen, nil
	}
	return h.createPlayer(ctx, seasonName, r.Club, r.Player, allPlayers, sum)
}

func (h *HistoricalImporter) createPlayer(ctx context.Context, season, club, name string, allPlayers *[]PlayerRef, sum *ImportSeasonSummary) (int, error) {
	id, err := h.repo.CreatePlayer(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("create player %q: %w", name, err)
	}
	*allPlayers = append(*allPlayers, PlayerRef{ID: id, Name: name})
	sum.NewPlayers++
	h.log.NewPlayer(season, club, name, id)
	return id, nil
}

// fuzzyBest returns the highest-confidence existing player for a name, or a zero
// PlayerChoice if there are none. The PlayerResolver is generic over an int
// identifier — here the candidate's PlayerSeasonID field carries the player id.
func (h *HistoricalImporter) fuzzyBest(ctx context.Context, name, club string, pool []PlayerRef) PlayerChoice {
	if len(pool) == 0 {
		return PlayerChoice{}
	}
	candidates := make([]PlayerCandidate, len(pool))
	for i, p := range pool {
		candidates[i] = PlayerCandidate{PlayerSeasonID: p.ID, Name: p.Name}
	}
	matches, err := h.resolver.Resolve(ctx, name, club, candidates)
	if err != nil || len(matches) == 0 {
		return PlayerChoice{}
	}
	top := matches[0]
	return PlayerChoice{PlayerID: top.Candidate.PlayerSeasonID, Name: top.Candidate.Name, Confidence: top.Confidence}
}

// yearConnected reports whether `year` is adjacent to (within one year of) any
// season the player already has — i.e. their career reaches this season. No
// known seasons is treated as connected (nothing to doubt).
func yearConnected(year int, years []int) bool {
	if len(years) == 0 {
		return true
	}
	for _, y := range years {
		if y-year <= 1 && year-y <= 1 {
			return true
		}
	}
	return false
}

func containsFold(list []string, s string) bool {
	for _, v := range list {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}

// mustYears fetches a player's season years for display, swallowing errors.
func mustYears(ctx context.Context, repo HistoricalRepo, playerID int) []int {
	years, _ := repo.PlayerSeasonYears(ctx, playerID)
	return years
}

// nearestGap returns the distance in years from `year` to the closest season
// the player already has (0 means overlapping/present). Assumes years non-empty.
func nearestGap(year int, years []int) int {
	best := -1
	for _, y := range years {
		d := y - year
		if d < 0 {
			d = -d
		}
		if best < 0 || d < best {
			best = d
		}
	}
	if best < 0 {
		return 0
	}
	return best
}

// spanDetail summarises a player's known seasons for an operator prompt.
func spanDetail(years []int) string {
	if len(years) == 0 {
		return "no recorded seasons"
	}
	lo, hi := years[0], years[0]
	for _, y := range years {
		if y < lo {
			lo = y
		}
		if y > hi {
			hi = y
		}
	}
	if lo == hi {
		return fmt.Sprintf("played %d", lo)
	}
	return fmt.Sprintf("played %d–%d", lo, hi)
}

// roundName maps an afltables round token to the DB convention: numeric rounds
// become "Round N"; finals keep their name ("Grand Final", etc.).
func roundName(round string) string {
	if _, err := strconv.Atoi(round); err == nil {
		return "Round " + round
	}
	return round
}

// parseDate parses ISO "2006-01-02"; an unparseable/empty date yields the zero
// time (stored as NULL start_dt).
func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
