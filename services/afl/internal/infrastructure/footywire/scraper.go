package footywire

import (
	"context"
	"fmt"
	"io"
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
	return ParseMatchStatsHTML(body)
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
	return ParseFixtureMid(body, roundName, homeClub, awayClub)
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
// adjust the selectors in collectStatsTable.
func ParseMatchStatsHTML(r io.Reader) (application.MatchStats, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse HTML: %w", err)
	}

	tables := findStatsTables(doc)
	if len(tables) < 2 {
		return application.MatchStats{}, fmt.Errorf("expected 2 stats tables, found %d", len(tables))
	}

	home, err := parseStatsTable(tables[0])
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse home stats: %w", err)
	}
	away, err := parseStatsTable(tables[1])
	if err != nil {
		return application.MatchStats{}, fmt.Errorf("parse away stats: %w", err)
	}

	var stats application.MatchStats
	stats.HomeClubName = home.clubName
	stats.HomeTeamGoals = home.goals
	stats.HomeTeamBehinds = home.behinds
	stats.AwayClubName = away.clubName
	stats.AwayTeamGoals = away.goals
	stats.AwayTeamBehinds = away.behinds
	stats.Players = append(home.players, away.players...)
	return stats, nil
}

// ParseFixtureMid parses the FootyWire fixture list page and returns the mid
// for the match identified by roundName, homeClub and awayClub.
//
// FootyWire's fixture list groups matches by round in tables. Each match row
// contains links to ft_match_statistics?mid=XXXXX. We look for the round
// section and then find the row whose home/away club cells match our clubs.
func ParseFixtureMid(r io.Reader, roundName, homeClub, awayClub string) (string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", fmt.Errorf("parse HTML: %w", err)
	}

	mid := findMidInFixture(doc, roundName, homeClub, awayClub)
	if mid == "" {
		return "", fmt.Errorf("match not found: round=%q home=%q away=%q", roundName, homeClub, awayClub)
	}
	return mid, nil
}

// ---- internal parsing helpers ----

type tableStats struct {
	clubName string
	goals    int
	behinds  int
	players  []application.PlayerStats
}

// findStatsTables returns the two <table> nodes that contain player stats.
// FootyWire uses <table class="ft"> for both stats tables; we select the ones
// that contain a column header row with "K" and "HB" cells.
func findStatsTables(n *html.Node) []*html.Node {
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
	walk(n)
	return tables
}

// isStatsTable returns true when the table contains a header row with both "K" and "HB" cells.
func isStatsTable(table *html.Node) bool {
	var found bool
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && (n.Data == "th" || n.Data == "td") {
			text := strings.TrimSpace(textContent(n))
			if text == "K" || text == "HB" {
				found = true
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)
	return found
}

func parseStatsTable(table *html.Node) (tableStats, error) {
	rows := tableRows(table)
	if len(rows) < 2 {
		return tableStats{}, fmt.Errorf("too few rows in stats table")
	}

	var ts tableStats

	// First row: club heading with score "Goals.Behinds (Total)" embedded in text.
	headingText := strings.TrimSpace(textContent(rows[0]))
	ts.clubName, ts.goals, ts.behinds = parseClubHeading(headingText)

	// Find the header row (contains "K" cell) to build column index map.
	var colIdx map[string]int
	var dataRows []*html.Node
	for i, row := range rows[1:] {
		cells := rowCells(row)
		idx := buildColIndex(cells)
		if _, ok := idx["K"]; ok {
			colIdx = idx
			dataRows = rows[i+2:] // rows after the header
			break
		}
	}
	if colIdx == nil {
		return tableStats{}, fmt.Errorf("could not find column header row in stats table")
	}

	for _, row := range dataRows {
		cells := rowCells(row)
		if len(cells) < 3 {
			continue
		}
		// Player name is the first non-numeric cell (skip jersey number).
		playerName := strings.TrimSpace(textContent(cells[1]))
		if playerName == "" || playerName == "Totals" || playerName == "Opposition" {
			continue
		}

		ps := application.PlayerStats{
			Name:      playerName,
			ClubName:  ts.clubName,
			Kicks:     cellInt(cells, colIdx, "K"),
			Handballs: cellInt(cells, colIdx, "HB"),
			Marks:     cellInt(cells, colIdx, "M"),
			Hitouts:   cellInt(cells, colIdx, "HO"),
			Tackles:   cellInt(cells, colIdx, "T"),
			Goals:     cellInt(cells, colIdx, "G"),
			Behinds:   cellInt(cells, colIdx, "B"),
		}
		ts.players = append(ts.players, ps)
	}

	return ts, nil
}

// parseClubHeading extracts the club name, goals, and behinds from a heading like
// "Carlton 14.9 (93)" or "Greater Western Sydney Giants 8.12 (60)".
func parseClubHeading(text string) (clubName string, goals, behinds int) {
	// Find the score pattern: digits.digits possibly followed by (total).
	// The club name is everything before the score.
	parts := strings.Fields(text)
	for i, p := range parts {
		if idx := strings.Index(p, "."); idx > 0 {
			g, errG := strconv.Atoi(p[:idx])
			rest := p[idx+1:]
			// behinds may have trailing "(total)" — strip it
			rest = strings.Split(rest, "(")[0]
			rest = strings.TrimSuffix(rest, ")")
			b, errB := strconv.Atoi(strings.TrimSpace(rest))
			if errG == nil && errB == nil {
				clubName = strings.Join(parts[:i], " ")
				goals = g
				behinds = b
				return
			}
		}
	}
	// Fallback: return whole text as club name.
	clubName = text
	return
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

// findMidInFixture walks the fixture list DOM looking for a link to
// ft_match_statistics?mid=XXXXX that sits within the correct round section
// and near the expected club names.
// findMidInFixture walks the fixture DOM in document order, tracking the most
// recently seen round heading. When a match link is found, it checks whether
// the surrounding row text contains both club names and the current heading
// matches the requested round.
func findMidInFixture(doc *html.Node, roundName, homeClub, awayClub string) string {
	normRound := normStr(roundName)
	normHome := normStr(homeClub)
	normAway := normStr(awayClub)

	var currentRound string

	var walk func(*html.Node) string
	walk = func(n *html.Node) string {
		if n.Type == html.ElementNode {
			// Track round headings (h1–h3).
			if n.Data == "h1" || n.Data == "h2" || n.Data == "h3" {
				currentRound = normStr(textContent(n))
			}

			if n.Data == "a" {
				href := attrVal(n, "href")
				if strings.Contains(href, "ft_match_statistics?mid=") {
					// Gather text of the nearest table row containing this link.
					rowText := normStr(nearestRowText(n))
					if strings.Contains(currentRound, normRound) &&
						strings.Contains(rowText, normHome) &&
						strings.Contains(rowText, normAway) {
						return extractMid(href)
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
