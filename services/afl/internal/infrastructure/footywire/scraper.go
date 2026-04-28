package footywire

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"xffl/services/afl/internal/application"
)

const (
	baseURL      = "https://www.footywire.com/afl/footy"
	statsPathFmt = "/ft_match_statistics?mid=%s"
	fixturePath  = "/ft_match_list"
)

// FootywireClient fetches and parses AFL match data from FootyWire.
// It implements application.StatsParser and application.FixtureDiscovery.
type FootywireClient struct {
	http *http.Client
}

func NewFootywireClient() *FootywireClient {
	return &FootywireClient{http: &http.Client{}}
}

// ParseMatch fetches the match statistics page for the given mid and returns parsed stats.
func (c *FootywireClient) ParseMatch(ctx context.Context, mid string) (application.MatchStats, error) {
	url := baseURL + fmt.Sprintf(statsPathFmt, mid)
	body, err := c.fetch(ctx, url)
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("fetch match stats page: %w", err)
	}
	defer body.Close()
	return ParseMatchStatsHTML(ctx, body)
}

// FindMatchMid scrapes the fixture list to find the FootyWire match ID for the given
// round and clubs. Returns an error if no matching match is found.
func (c *FootywireClient) FindMatchMid(ctx context.Context, roundName, homeClub, awayClub string) (string, error) {
	url := baseURL + fixturePath
	body, err := c.fetch(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetch fixture list: %w", err)
	}
	defer body.Close()
	return ParseFixtureMid(ctx, body, roundName, homeClub, awayClub)
}

func (c *FootywireClient) fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; xffl-stats-importer/1.0)")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}
	return resp.Body, nil
}

// ParseMatchStatsHTML parses the FootyWire match statistics HTML page.
//
// FootyWire renders two stats tables (one per club) each with a header row
// identifying columns by their AFL abbreviation (K, HB, M, G, B, HO, T, …).
// The club name and score ("Goals.Behinds (Total)") appear in a preceding
// heading row within the same table.
//
// Column detection is done by matching the header row, so the parser is robust
// to column reordering. If the actual page structure differs from this model,
// The page has a Team|Q1|Q2|Q3|Q4|Final score table and two separate player
// stats tables (one per club, home first). Club names and team totals come from
// the score table; player rows come from the stats tables.
func ParseMatchStatsHTML(ctx context.Context, r io.Reader) (application.MatchStats, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse HTML: %w", err)
	}

	scoreTable := findScoreTable(doc)
	if scoreTable == nil {
		return application.MatchStats{}, fmt.Errorf("could not find score summary table (Team|Q1..Final)")
	}
	homeClub, awayClub, homeGoals, homeBehinds, awayGoals, awayBehinds, err := parseScoreTable(scoreTable)
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse score table: %w", err)
	}
	slog.DebugContext(ctx, "score parsed",
		slog.String("home", homeClub), slog.Int("homeGoals", homeGoals), slog.Int("homeBehinds", homeBehinds),
		slog.String("away", awayClub), slog.Int("awayGoals", awayGoals), slog.Int("awayBehinds", awayBehinds),
	)

	tables := findStatsTables(doc)
	slog.DebugContext(ctx, "player stats tables found", slog.Int("count", len(tables)))
	if len(tables) < 2 {
		return application.MatchStats{}, fmt.Errorf("expected 2 player stats tables, found %d", len(tables))
	}

	homePlayers, err := parsePlayerRows(tables[0], homeClub)
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse home player stats: %w", err)
	}
	awayPlayers, err := parsePlayerRows(tables[1], awayClub)
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse away player stats: %w", err)
	}

	return application.MatchStats{
		HomeClubName:    homeClub,
		HomeTeamGoals:   homeGoals,
		HomeTeamBehinds: homeBehinds,
		AwayClubName:    awayClub,
		AwayTeamGoals:   awayGoals,
		AwayTeamBehinds: awayBehinds,
		Players:         append(homePlayers, awayPlayers...),
	}, nil
}

// ParseFixtureMid parses the FootyWire fixture list page and returns the mid
// for the match identified by roundName, homeClub and awayClub.
//
// FootyWire's fixture list groups matches by round in tables. Each match row
// contains links to ft_match_statistics?mid=XXXXX. We look for the round
// section and then find the row whose home/away club cells match our clubs.
func ParseFixtureMid(ctx context.Context, r io.Reader, roundName, homeClub, awayClub string) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", fmt.Errorf("parse HTML: %w", err)
	}

	mid := findMidInFixture(ctx, doc, roundName, homeClub, awayClub)
	if mid == "" {
		return "", fmt.Errorf("match not found: round=%q home=%q away=%q", roundName, homeClub, awayClub)
	}
	return mid, nil
}

// ---- internal parsing helpers ----

// findScoreTable finds the Team|Q1|Q2|Q3|Q4|Final summary table.
func findScoreTable(doc *html.Node) *html.Node {
	var result *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if result != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "table" && isScoreTable(n) {
			result = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return result
}

// isScoreTable returns true when the table's first row has both "Team" and "Final" cells.
func isScoreTable(table *html.Node) bool {
	rows := tableRows(table)
	if len(rows) < 3 {
		return false
	}
	hasTeam, hasFinal := false, false
	for _, cell := range rowCells(rows[0]) {
		switch strings.TrimSpace(textContent(cell)) {
		case "Team":
			hasTeam = true
		case "Final":
			hasFinal = true
		}
	}
	return hasTeam && hasFinal
}

// parseScoreTable extracts home/away club names and final-quarter goals.behinds.
func parseScoreTable(table *html.Node) (homeClub, awayClub string, homeGoals, homeBehinds, awayGoals, awayBehinds int, err error) {
	rows := tableRows(table)
	if len(rows) < 3 {
		err = fmt.Errorf("score table has %d rows, need at least 3", len(rows))
		return
	}
	colIdx := buildColIndex(rowCells(rows[0]))
	q4Idx, ok := colIdx["Q4"]
	if !ok {
		err = fmt.Errorf("score table missing Q4 column")
		return
	}
	parseCells := func(row *html.Node) (club string, goals, behinds int) {
		cells := rowCells(row)
		if len(cells) == 0 {
			return
		}
		club = strings.TrimSpace(textContent(cells[0]))
		if q4Idx < len(cells) {
			goals, behinds = parseGoalsBehinds(strings.TrimSpace(textContent(cells[q4Idx])))
		}
		return
	}
	homeClub, homeGoals, homeBehinds = parseCells(rows[1])
	awayClub, awayGoals, awayBehinds = parseCells(rows[2])
	return
}

// parseGoalsBehinds parses "9.6" → (9, 6). Returns zeros on bad input.
func parseGoalsBehinds(s string) (goals, behinds int) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) == 2 {
		goals, _ = strconv.Atoi(parts[0])
		behinds, _ = strconv.Atoi(parts[1])
	}
	return
}

// findStatsTables returns all <table> nodes whose own rows (not in nested tables)
// contain a cell with text exactly "K" (Kicks column header).
func findStatsTables(doc *html.Node) []*html.Node {
	var tables []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			if isStatsTable(n) {
				tables = append(tables, n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return tables
}

// isStatsTable returns true when one of the table's own cells (not in nested tables)
// has text exactly "K". Uses tableRows so outer layout tables that merely contain
// a stats table inside a cell are not matched.
func isStatsTable(table *html.Node) bool {
	for _, row := range tableRows(table) {
		for _, cell := range rowCells(row) {
			if strings.TrimSpace(textContent(cell)) == "K" {
				return true
			}
		}
	}
	return false
}

// parsePlayerRows extracts player stats from a player stats table.
// The table must have a header row containing "K"; club name is injected from
// the score table rather than parsed from a heading row.
func parsePlayerRows(table *html.Node, clubName string) ([]application.PlayerStats, error) {
	rows := tableRows(table)

	var colIdx map[string]int
	var dataRows []*html.Node
	for i, row := range rows {
		idx := buildColIndex(rowCells(row))
		if _, ok := idx["K"]; ok {
			colIdx = idx
			dataRows = rows[i+1:]
			break
		}
	}
	if colIdx == nil {
		return nil, fmt.Errorf("could not find column header row in player stats table for %q", clubName)
	}

	// "Player" column is first on the real page; fall back to index 0.
	nameIdx := 0
	if i, ok := colIdx["Player"]; ok {
		nameIdx = i
	}

	var players []application.PlayerStats
	for _, row := range dataRows {
		cells := rowCells(row)
		if len(cells) <= nameIdx {
			continue
		}
		name := strings.TrimSpace(textContent(cells[nameIdx]))
		if name == "" || name == "Totals" || name == "Opposition" {
			continue
		}
		players = append(players, application.PlayerStats{
			Name:      name,
			ClubName:  clubName,
			Kicks:     cellInt(cells, colIdx, "K"),
			Handballs: cellInt(cells, colIdx, "HB"),
			Marks:     cellInt(cells, colIdx, "M"),
			Hitouts:   cellInt(cells, colIdx, "HO"),
			Tackles:   cellInt(cells, colIdx, "T"),
			Goals:     cellInt(cells, colIdx, "G"),
			Behinds:   cellInt(cells, colIdx, "B"),
		})
	}
	return players, nil
}

func buildColIndex(cells []*html.Node) map[string]int {
	idx := make(map[string]int, len(cells))
	for i, c := range cells {
		text := strings.TrimSpace(textContent(c))
		if text != "" {
			idx[text] = i
		}
	}
	return idx
}

func cellInt(cells []*html.Node, colIdx map[string]int, col string) int {
	i, ok := colIdx[col]
	if !ok || i >= len(cells) {
		return 0
	}
	v, _ := strconv.Atoi(strings.TrimSpace(textContent(cells[i])))
	return v
}

// footywireClubAliases maps FootyWire's abbreviated fixture names to a normalised
// prefix of the full club name. Only needed for clubs whose fixture abbreviation
// bears no prefix relation to their full name (e.g. GWS acronym).
var footywireClubAliases = map[string]string{
	"gws": "greaterwestern", // Greater Western Sydney Giants
}

// matchesClub reports whether normClub (full normalised club name) matches the
// abbreviated name that FootyWire uses in fixture row text.
//
// FootyWire drops "mascot" words in some cases (e.g. "sydney" for "Sydney Swans",
// "brisbane" for "Brisbane Lions"), so the fixture token is often a prefix of the
// full normalised name. For GWS, an explicit alias handles the acronym form.
func matchesClub(rowText, normClub string) bool {
	if strings.Contains(rowText, normClub) {
		return true
	}
	// rowText is the output of normStr so spaces are gone but newlines survive —
	// split on newlines to extract individual tokens.
	for _, tok := range strings.Split(rowText, "\n") {
		if len(tok) < 3 {
			continue
		}
		if strings.HasPrefix(normClub, tok) {
			return true
		}
		if alias, ok := footywireClubAliases[tok]; ok && strings.HasPrefix(normClub, alias) {
			return true
		}
	}
	return false
}

// findMidInFixture walks the fixture DOM in document order. When a match link is
// found it checks whether the surrounding row text contains both club names.
//
// Note: the FootyWire fixture page does not use h1–h3 headings to mark rounds, so
// round-based disambiguation is not applied. Club-name matching is sufficient for
// the regular season where each matchup occurs at most twice.
func findMidInFixture(ctx context.Context, doc *html.Node, roundName, homeClub, awayClub string) string {
	normHome := normStr(homeClub)
	normAway := normStr(awayClub)

	slog.DebugContext(ctx, "fixture discovery start",
		slog.String("round", roundName),
		slog.String("home", homeClub), slog.String("normHome", normHome),
		slog.String("away", awayClub), slog.String("normAway", normAway),
	)

	var walk func(*html.Node) string
	walk = func(n *html.Node) string {
		if n.Type == html.ElementNode {
			if n.Data == "a" {
				href := attrVal(n, "href")
				if strings.Contains(href, "ft_match_statistics?mid=") {
					rowText := normStr(nearestRowText(n))
					homeMatch := matchesClub(rowText, normHome)
					awayMatch := matchesClub(rowText, normAway)
					slog.DebugContext(ctx, "fixture link",
						slog.String("href", href),
						slog.String("rowText", rowText),
						slog.Bool("homeMatch", homeMatch),
						slog.Bool("awayMatch", awayMatch),
					)
					if homeMatch && awayMatch {
						mid := extractMid(href)
						slog.DebugContext(ctx, "fixture match found", slog.String("mid", mid))
						return mid
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if mid := walk(c); mid != "" {
				return mid
			}
		}
		return ""
	}

	return walk(doc)
}

// nearestRowText walks up from n to find the enclosing <tr> and returns its text.
// Falls back to the parent element's text if no <tr> is found.
func nearestRowText(n *html.Node) string {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == "tr" {
			return textContent(p)
		}
	}
	if n.Parent != nil {
		return textContent(n.Parent)
	}
	return ""
}

func normStr(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", ""))
}

func extractMid(href string) string {
	const prefix = "mid="
	idx := strings.Index(href, prefix)
	if idx < 0 {
		return ""
	}
	mid := href[idx+len(prefix):]
	// Stop at any query separator.
	if i := strings.IndexAny(mid, "&# "); i >= 0 {
		mid = mid[:i]
	}
	return mid
}

// ---- DOM traversal utilities ----

func tableRows(table *html.Node) []*html.Node {
	var rows []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			rows = append(rows, n)
			return // don't recurse into nested tables
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)
	return rows
}

func rowCells(row *html.Node) []*html.Node {
	var cells []*html.Node
	for c := row.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			cells = append(cells, c)
		}
	}
	return cells
}

func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textContent(c))
	}
	return sb.String()
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
