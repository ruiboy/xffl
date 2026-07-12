// Package afltables parses historical AFL match data from afltables.com HTML
// pages into flat records suitable for CSV export. It is a one-time data-prep
// tool (see ADR-024 direction / Phase 24) and deliberately has no dependency on
// the application or domain layers — it turns HTML into rows, nothing more.
//
// afltables renders each match ("game") page with a metadata table (round,
// venue, date, and each club's quarter-by-quarter score) followed by one
// "<Club> Match Statistics" table per club, home first. Player rows carry a
// numeric jumper in the first cell; a trailing "Totals" row is ignored.
// Columns are located by their afltables abbreviation (KI, MK, HB, …) so the
// parser tolerates column reordering across eras.
package afltables

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Game is one parsed match: its metadata plus a flat player list (both clubs).
type Game struct {
	Round    string // e.g. "1", or a final's name like "Elimination Final"
	Date     string // ISO "2006-01-02"
	Venue    string
	HomeClub string
	AwayClub string
	Players  []PlayerLine
}

// PlayerLine is one player's counting stats for a match. Only the seven stats
// the AFL schema stores are captured; afltables' advanced columns are dropped.
type PlayerLine struct {
	Club      string
	Name      string // normalised to "First Last"
	Kicks     int
	Marks     int
	Handballs int
	Goals     int
	Behinds   int
	Hitouts   int
	Tackles   int
}

// ParseSeasonIndex returns the game-page paths linked from a season index page
// (afltables.com/afl/seas/YYYY.html), de-duplicated and in document order.
func ParseSeasonIndex(r io.Reader) ([]string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse season index: %w", err)
	}
	var out []string
	seen := map[string]bool{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if href := attr(n, "href"); strings.Contains(href, "stats/games/") && strings.HasSuffix(href, ".html") {
				clean := strings.TrimPrefix(href, "../")
				if !seen[clean] {
					seen[clean] = true
					out = append(out, clean)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out, nil
}

// ParseGamePage parses a single afltables game stats page.
func ParseGamePage(r io.Reader) (Game, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return Game{}, fmt.Errorf("parse game page: %w", err)
	}

	var g Game
	statsTables := []*html.Node{}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			switch {
			case g.Round == "" && tableHasMeta(n):
				parseMeta(n, &g)
			case statsTableClub(n) != "":
				statsTables = append(statsTables, n)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if g.Round == "" {
		return Game{}, fmt.Errorf("no metadata table found")
	}
	if len(statsTables) < 2 {
		return Game{}, fmt.Errorf("expected 2 stats tables, found %d", len(statsTables))
	}
	for _, t := range statsTables {
		club := statsTableClub(t)
		lines, err := parsePlayerRows(t, club)
		if err != nil {
			return Game{}, fmt.Errorf("parse %s players: %w", club, err)
		}
		g.Players = append(g.Players, lines...)
	}
	return g, nil
}

// tableHasMeta reports whether a table's first row is the metadata line.
func tableHasMeta(t *html.Node) bool {
	rows := tableRows(t)
	if len(rows) == 0 {
		return false
	}
	return strings.Contains(textContent(rows[0]), "Round:") && strings.Contains(textContent(rows[0]), "Venue:")
}

// parseMeta reads round/venue/date from row 0 and home/away clubs from the two
// club-score rows that follow.
func parseMeta(t *html.Node, g *Game) {
	rows := tableRows(t)
	meta := normSpace(textContent(rows[0]))
	g.Round = between(meta, "Round:", "Venue:")
	g.Venue = between(meta, "Venue:", "Date:")
	g.Date = parseDate(between(meta, "Date:", "Attendance:"))

	// Rows 1 and 2 are the home and away club score lines: "<Club> q1 q2 q3 final".
	var clubs []string
	for _, row := range rows[1:] {
		cells := rowCells(row)
		if len(cells) == 0 {
			continue
		}
		name := strings.TrimSpace(textContent(cells[0]))
		if name == "" || strings.HasPrefix(name, "Qrt") || strings.Contains(name, "umpire") {
			continue
		}
		clubs = append(clubs, name)
		if len(clubs) == 2 {
			break
		}
	}
	if len(clubs) > 0 {
		g.HomeClub = clubs[0]
	}
	if len(clubs) > 1 {
		g.AwayClub = clubs[1]
	}
}

// statsTableClub returns the club name for a "<Club> Match Statistics" table,
// or "" if the table isn't one.
func statsTableClub(t *html.Node) string {
	rows := tableRows(t)
	if len(rows) == 0 {
		return ""
	}
	head := normSpace(textContent(rows[0]))
	i := strings.Index(head, "Match Statistics")
	if i <= 0 {
		return ""
	}
	return strings.TrimSpace(head[:i])
}

func parsePlayerRows(t *html.Node, club string) ([]PlayerLine, error) {
	rows := tableRows(t)
	var colIdx map[string]int
	var dataRows []*html.Node
	for i, row := range rows {
		idx := buildColIndex(rowCells(row))
		if _, ok := idx["KI"]; ok {
			colIdx = idx
			dataRows = rows[i+1:]
			break
		}
	}
	if colIdx == nil {
		return nil, fmt.Errorf("no header row (KI) found")
	}

	var out []PlayerLine
	for _, row := range dataRows {
		cells := rowCells(row)
		if len(cells) < 2 {
			continue
		}
		// Player rows start with a numeric jumper; skip Totals/Opponent rows.
		// The jumper cell may carry a milestone marker (e.g. "26 ↓"), so match a
		// leading digit rather than requiring a fully-numeric cell.
		if !startsWithDigit(strings.TrimSpace(textContent(cells[0]))) {
			continue
		}
		out = append(out, PlayerLine{
			Club:      club,
			Name:      flipName(strings.TrimSpace(textContent(cells[1]))),
			Kicks:     cellInt(cells, colIdx, "KI"),
			Marks:     cellInt(cells, colIdx, "MK"),
			Handballs: cellInt(cells, colIdx, "HB"),
			Goals:     cellInt(cells, colIdx, "GL"),
			Behinds:   cellInt(cells, colIdx, "BH"),
			Hitouts:   cellInt(cells, colIdx, "HO"),
			Tackles:   cellInt(cells, colIdx, "TK"),
		})
	}
	return out, nil
}

func startsWithDigit(s string) bool {
	return s != "" && s[0] >= '0' && s[0] <= '9'
}

// flipName converts afltables' "Last, First" to "First Last". Names without a
// comma are returned unchanged.
func flipName(s string) string {
	if i := strings.Index(s, ", "); i >= 0 {
		return strings.TrimSpace(s[i+2:]) + " " + strings.TrimSpace(s[:i])
	}
	return s
}

// parseDate turns "Thu, 16-Mar-2023 7:20 PM (6:20 PM)" into ISO "2023-03-16".
// Returns the trimmed input on failure so nothing is silently lost.
func parseDate(s string) string {
	s = strings.TrimSpace(s)
	// Isolate the "DD-Mon-YYYY" token.
	fields := strings.Fields(s)
	for _, f := range fields {
		parts := strings.Split(f, "-")
		if len(parts) == 3 && len(parts[2]) == 4 {
			mon := monthNum[parts[1]]
			if mon == "" {
				break
			}
			day := parts[0]
			if len(day) == 1 {
				day = "0" + day
			}
			return parts[2] + "-" + mon + "-" + day
		}
	}
	return s
}

var monthNum = map[string]string{
	"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04", "May": "05", "Jun": "06",
	"Jul": "07", "Aug": "08", "Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
}

// --- small HTML helpers (mirrors the footywire package's approach) ---

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

func buildColIndex(cells []*html.Node) map[string]int {
	idx := map[string]int{}
	for i, c := range cells {
		idx[strings.TrimSpace(textContent(c))] = i
	}
	return idx
}

// cellInt reads the integer at column `col`; blank/&nbsp;/missing cells are 0.
func cellInt(cells []*html.Node, colIdx map[string]int, col string) int {
	i, ok := colIdx[col]
	if !ok || i >= len(cells) {
		return 0
	}
	v := strings.TrimSpace(textContent(cells[i]))
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return n
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

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func normSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// between returns the text between markers `from` and `to` (exclusive), trimmed.
func between(s, from, to string) string {
	i := strings.Index(s, from)
	if i < 0 {
		return ""
	}
	i += len(from)
	j := strings.Index(s[i:], to)
	if j < 0 {
		return strings.TrimSpace(s[i:])
	}
	return strings.TrimSpace(s[i : i+j])
}
