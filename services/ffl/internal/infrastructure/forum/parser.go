package forum

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"xffl/services/ffl/internal/application"
)

// Parser implements application.TeamParser for Tapatalk FFL forum posts.
// Supports all four team formats: Ruiboys, Slashers, Cheetahs, THC.
type Parser struct{}

func NewParser() *Parser { return &Parser{} }

// position aliases: raw token → canonical position name
var positionAliases = map[string]string{
	"GOALS": "goals", "GOAL": "goals",
	"KICKS": "kicks", "KICK": "kicks",
	"HANDBALLS": "handballs", "HANDBALL": "handballs", "HB": "handballs", "HBS": "handballs",
	"MARKS": "marks", "MARK": "marks",
	"TACKLES": "tackles", "TACKLE": "tackles",
	"HITOUTS": "hitouts", "HITOUT": "hitouts",
	"RUCK": "hitouts", "RUCKS": "hitouts", "HO": "hitouts",
	"STAR": "star",
	"BENCH": "bench", "INTERCHANGE": "bench",
}

// bench letter codes → position
var benchLetter = map[string]string{
	"G": "goals", "K": "kicks",
	"H": "handballs", "HB": "handballs",
	"M": "marks", "T": "tackles",
	"R": "hitouts", "HO": "hitouts",
	"S": "star",
}

// nicknames: lowercase key → canonical name
var nicknames = map[string][2]string{
	"tdk":         {"Tom De Koning", "SK"},
	"t de koning": {"Tom De Koning", "SK"},
	"the bont":    {"M Bontempelli", "WB"},
}

var stripWords = []string{"Journeyman", "Mountain Goat"}

var (
	artifactRE   = regexp.MustCompile(`(?i)^(Quote|Edit|Share|Like|Dislike|Pin\s+Topic|TATLTWDNMTS|Bloody\s+Legend|hugs?\s*$|reacted\s+to|\w+\s+reacted\s+to|\w+\s+likes?\s+this\s+post|likes?\s+this\s+post|\d{1,2}:\d{2}\s*(AM|PM))`)
	memberNumRE  = regexp.MustCompile(`^\d[\d,]{3,}$`)
	subtotalRE   = regexp.MustCompile(`^\s*\d+\s*$`)
	sectionRE    = regexp.MustCompile(`(?i)^\s*(?:I/C[–-]\s*)?(GOALS?|KICKS?|HANDBALLS?|HANDBALL|HB|HBS|MARKS?|TACKLES?|HITOUTS?|RUCK[S]?|HO|STAR|BENCH|INTERCHANGE)\b[\s\d=]*$`)
	icSectionRE  = regexp.MustCompile(`(?i)^\s*I/C[–\-]\s*\w+`)
	thcScoreRE   = regexp.MustCompile(`(?i)^THC[-–\s]+(\d+)`)
	cheetahRE    = regexp.MustCompile(`(?i)^CHEETAHS\s+(\d+)`)
	totalRE      = regexp.MustCompile(`(?i)^TOTAL\s*:\s*(\d+)`)
	ruiHeaderRE  = regexp.MustCompile(`^R\d+\s+(\d+)$`)
	bareScoreRE  = regexp.MustCompile(`^(\d{3,})\s*$`)
	icStarMRE    = regexp.MustCompile(`^\*\s*=\s*(.+)`)
	icStarLabelRE = regexp.MustCompile(`(?i)^Star[-–]\s*(.+)`)
	benchCodeRE  = regexp.MustCompile(`(?i)^([A-Z]+/[A-Z]+)\s*[-=]\s*(.+)`)
)

func (p *Parser) Parse(_ context.Context, teamName, post string) ([]application.ParsedPlayerRow, error) {
	lines := splitLines(post)
	team := teamName
	if team == "" {
		team = detectTeam(lines)
	}
	if team == "" {
		return nil, fmt.Errorf("could not identify team from post")
	}
	return parseBlock(team, lines), nil
}

// --- team detection ---

func detectTeam(lines []string) string {
	text := strings.Join(lines[:min(5, len(lines))], "\n")
	if strings.Contains(strings.ToUpper(text), "THC") {
		return "THC"
	}
	if strings.Contains(strings.ToUpper(text), "CHEETAHS") {
		return "Cheetahs"
	}
	for _, l := range lines {
		if totalRE.MatchString(l) {
			return "Slashers"
		}
	}
	for _, l := range lines {
		if strings.Contains(l, "–") {
			return "Ruiboys"
		}
	}
	return ""
}

// --- block parser ---

func parseBlock(team string, lines []string) []application.ParsedPlayerRow {
	var rows []application.ParsedPlayerRow
	currentPos := ""
	inIC := false

	isRuiboys := team == "Ruiboys"
	isSlashers := team == "Slashers"
	isCheetahs := team == "Cheetahs"
	isTHC := team == "THC"

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// skip score header lines — we only want player scores
		if isRuiboys && (ruiHeaderRE.MatchString(line) || (bareScoreRE.MatchString(line) && currentPos == "")) {
			continue
		}
		if (isTHC || isSlashers) && bareScoreRE.MatchString(line) && currentPos == "" {
			continue
		}
		if isCheetahs && cheetahRE.MatchString(line) {
			continue
		}
		if isTHC && thcScoreRE.MatchString(line) {
			continue
		}
		if isSlashers && totalRE.MatchString(line) {
			continue
		}

		if isArtifact(line) {
			continue
		}

		// position subtotals — numeric-only lines after position is set
		if subtotalRE.MatchString(line) {
			continue
		}

		// THC I/C section header
		if isTHC && icSectionRE.MatchString(line) {
			inIC = true
			currentPos = "bench"
			continue
		}

		// "Interchange = * *" — for Cheetahs, marks the bench star player as interchange
		if regexp.MustCompile(`(?i)^Interchange\s*=\s*\*`).MatchString(line) {
			if isCheetahs {
				for i := len(rows) - 1; i >= 0; i-- {
					if rows[i].BackupPositions == "star" {
						rows[i].InterchangePosition = "star"
						break
					}
				}
			}
			continue
		}

		// position section header
		if m := sectionRE.FindStringSubmatch(line); m != nil {
			currentPos = normalisePosition(m[1])
			inIC = false
			continue
		}

		if currentPos == "" {
			continue
		}

		var row *application.ParsedPlayerRow
		switch {
		case isRuiboys:
			row = parseRuiboys(line, currentPos)
		case isSlashers:
			row = parseSlashers(line, currentPos)
		case isCheetahs:
			row = parseCheetahs(line, currentPos)
		case isTHC:
			row = parseTHC(line, currentPos, inIC)
		}
		if row != nil {
			rows = append(rows, *row)
		}
	}
	return rows
}

// --- Ruiboys ---

var ruiPlayerRE = regexp.MustCompile(`^(.+?)\s*[–\-]\s*([A-Z][a-zA-Z]+)\s*(.*)`)

func parseRuiboys(line, position string) *application.ParsedPlayerRow {
	m := ruiPlayerRE.FindStringSubmatch(line)
	if m == nil {
		return nil
	}
	name, clubHint, rest := strings.TrimSpace(m[1]), strings.TrimSpace(m[2]), strings.TrimSpace(m[3])
	name, clubHint = resolveNickname(name, clubHint)

	row := &application.ParsedPlayerRow{Name: name, ClubHint: clubHint, Position: position}

	if position == "bench" {
		bm := regexp.MustCompile(`(?i)^(\*\s*\(?\s*INT\s*\)?|\*|[A-Z]+/[A-Z]+)\s*(.*)`).FindStringSubmatch(rest)
		if bm != nil {
			bp, ic := decodeBenchCode(strings.TrimSpace(bm[1]))
			row.BackupPositions = bp
			row.InterchangePosition = ic
			tail := strings.TrimSpace(bm[2])
			if tail != "" && !regexp.MustCompile(`\d+[\\|]\d+`).MatchString(tail) {
				row.Score = extractScore(tail)
			}
		}
		return row
	}

	subM := regexp.MustCompile(`(?i)(\d+)\s+sub\s+(\d+)`).FindStringSubmatch(rest)
	if subM != nil {
		s, _ := strconv.Atoi(subM[2])
		row.Score = &s
		row.Notes = fmt.Sprintf("starter score %s; interchange/sub used, slot score = %d", subM[1], s)
	} else {
		row.Score = lastNumber(rest)
	}
	return row
}

// --- Slashers ---

var slashersNameClubRE = regexp.MustCompile(`^([A-Z][A-Za-z\s'\-]+?)\s*\(([A-Z][a-zA-Z]+)\)`)

func parseSlashers(line, position string) *application.ParsedPlayerRow {
	// bench code prefix: "K/G - Name"
	if position == "bench" {
		if bm := benchCodeRE.FindStringSubmatch(line); bm != nil {
			bp, ic := decodeBenchCode(bm[1])
			n, club := slashersNameClub(bm[2])
			n, club = resolveNickname(n, club)
			return &application.ParsedPlayerRow{
				Name: n, ClubHint: club, Position: position,
				BackupPositions: bp, InterchangePosition: ic,
				Score: extractScore(strings.SplitN(bm[2], "(", 2)[len(strings.SplitN(bm[2], "(", 2))-1]),
			}
		}
		// interchange player: ***Name (Club)***
		if m := regexp.MustCompile(`^\*{2,3}\s*(.+?)\s*\*{2,3}\s*(.*)`).FindStringSubmatch(line); m != nil {
			n, club := slashersNameClub(m[1])
			n, club = resolveNickname(n, club)
			s := extractScore(m[2])
			return &application.ParsedPlayerRow{
				Name: n, ClubHint: club, Position: position,
				BackupPositions: "star", InterchangePosition: "star", Score: s,
			}
		}
	}

	// DNP line
	if m := regexp.MustCompile(`(?i)^(.+?)\s+dnp\s*[-–]\s*interchange\s+(\w+)\s+(\d+)`).FindStringSubmatch(line); m != nil {
		n, club := slashersNameClub(m[1])
		zero := 0
		return &application.ParsedPlayerRow{
			Name: n, ClubHint: club, Position: position,
			Score: &zero, Notes: fmt.Sprintf("DNP; %s subbed in for %s pts", m[2], m[3]),
		}
	}

	n, club := slashersNameClub(line)
	if n == "" {
		return nil
	}
	n, club = resolveNickname(n, club)
	sm := regexp.MustCompile(`\)\s+(\d+)`).FindStringSubmatch(line)
	var score *int
	if sm != nil {
		v, _ := strconv.Atoi(sm[1])
		score = &v
	}

	// interchange annotation on STAR line
	notes := ""
	if icm := regexp.MustCompile(`(?i)[-–]\s*interchange[d]?\s+(?:with\s+)?(\w+)\s+(\d+)`).FindStringSubmatch(line); icm != nil {
		icScore, _ := strconv.Atoi(icm[2])
		if score != nil && icScore > *score {
			notes = fmt.Sprintf("interchange occurred: %s %d > starter %d; slot score = %d", icm[1], icScore, *score, icScore)
			score = &icScore
		}
	}
	return &application.ParsedPlayerRow{Name: n, ClubHint: club, Position: position, Score: score, Notes: notes}
}

func slashersNameClub(s string) (string, string) {
	s = strings.TrimSpace(s)
	if m := slashersNameClubRE.FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	if m := regexp.MustCompile(`^([A-Z][A-Za-z\s'\-]+)`).FindStringSubmatch(s); m != nil {
		return strings.TrimSpace(m[1]), ""
	}
	return "", ""
}

// --- Cheetahs ---

var cheetahsPlayerRE = regexp.MustCompile(`^(.+?)\s*\(([A-Z][a-zA-Z]+)\)\s*(.*)`)

// Cheetahs show raw stats for these positions; multiply to get FFL pts.
var cheetahsRawMultipliers = map[string]int{"goals": 5, "marks": 2, "tackles": 4}

func parseCheetahs(line, position string) *application.ParsedPlayerRow {
	m := cheetahsPlayerRE.FindStringSubmatch(line)
	if m == nil {
		return nil
	}
	name := stripInlineNicknames(strings.TrimSpace(m[1]))
	club := strings.TrimSpace(m[2])
	rest := strings.TrimSpace(m[3])
	name, club = resolveNickname(name, club)

	row := &application.ParsedPlayerRow{Name: name, ClubHint: club, Position: position}

	if position == "bench" {
		tokens := strings.Fields(rest)
		if len(tokens) > 0 {
			code := strings.TrimRight(tokens[0], ",")
			bp, ic := decodeBenchCode(code)
			row.BackupPositions = bp
			row.InterchangePosition = ic
			tail := strings.Join(tokens[1:], " ")
			if tail != "" && !regexp.MustCompile(`\d+[,]\d+`).MatchString(tail) {
				row.Score = extractScore(tail)
			}
		}
		return row
	}

	rawScore := extractScore(rest)
	if rawScore != nil {
		if mult, ok := cheetahsRawMultipliers[position]; ok {
			ffl := *rawScore * mult
			row.Score = &ffl
			row.Notes = fmt.Sprintf("raw %s stat: %d × %d = %d", position, *rawScore, mult, ffl)
		} else {
			row.Score = rawScore
		}
	}
	return row
}

// --- THC ---

func parseTHC(line, position string, inIC bool) *application.ParsedPlayerRow {
	// I/C section: "Star- Name CLUB= score"
	if inIC {
		if m := icStarLabelRE.FindStringSubmatch(line); m != nil {
			n, club, score, notes := thcNameScore(strings.TrimSpace(m[1]))
			return &application.ParsedPlayerRow{
				Name: n, ClubHint: club, Position: "bench",
				BackupPositions: "star", InterchangePosition: "star", Score: score, Notes: notes,
			}
		}
		// "*= Name CLUB= score"
		if m := icStarMRE.FindStringSubmatch(line); m != nil {
			n, club, score, notes := thcNameScore(strings.TrimSpace(m[1]))
			return &application.ParsedPlayerRow{
				Name: n, ClubHint: club, Position: "bench",
				BackupPositions: "star", InterchangePosition: "star", Score: score, Notes: notes,
			}
		}
		// "K/HB- Name CLUB"
		if m := benchCodeRE.FindStringSubmatch(line); m != nil {
			bp, ic := decodeBenchCode(m[1])
			n, club, score, notes := thcNameScore(strings.TrimSpace(m[2]))
			return &application.ParsedPlayerRow{
				Name: n, ClubHint: club, Position: "bench",
				BackupPositions: bp, InterchangePosition: ic, Score: score, Notes: notes,
			}
		}
	}

	n, club, score, notes := thcNameScore(line)
	if n == "" {
		return nil
	}
	return &application.ParsedPlayerRow{Name: n, ClubHint: club, Position: position, Score: score, Notes: notes}
}

var (
	thcDNPStatRE  = regexp.MustCompile(`(?i)^(.+?)\s+([A-Z]+)\s+DNP[-=]\s*(\d+)[A-Za-z]+`)
	thcDNPRE      = regexp.MustCompile(`(?i)^(.+?)\s+([A-Z]+)\s*[-–]?\s*DNP[-=]\s*(\d+)`)
	thcMultRE     = regexp.MustCompile(`(?i)^(.+?)\s+([A-Z]+)\s+x(\d+)\s*=\s*(\d+)`)
	thcScoreLineRE = regexp.MustCompile(`^(.+?)\s+([A-Za-z]{2,4})\s*[-=]\s*(\d+)`)
	thcSpaceRE    = regexp.MustCompile(`^(.+?)\s+([A-Za-z]{2,4})\s+(\d+)\s*$`)
	thcNoScoreRE  = regexp.MustCompile(`^(.+?)\s+([A-Za-z]{2,4})\s*$`)
)

func thcNameScore(s string) (name, club string, score *int, notes string) {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile(`(?i)\s*\(AV\)`).ReplaceAllString(s, "")

	if m := thcDNPStatRE.FindStringSubmatch(s); m != nil {
		zero := 0
		return stripInlineNicknames(m[1]), m[2], &zero, fmt.Sprintf("DNP; sub contributed %s pts", m[3])
	}
	if m := thcDNPRE.FindStringSubmatch(s); m != nil {
		zero := 0
		return stripInlineNicknames(m[1]), m[2], &zero, fmt.Sprintf("DNP; sub contributed %s pts", m[3])
	}
	if m := thcMultRE.FindStringSubmatch(s); m != nil {
		v, _ := strconv.Atoi(m[4])
		return stripInlineNicknames(m[1]), m[2], &v, fmt.Sprintf("explicit multiplier x%s shown", m[3])
	}
	if m := thcScoreLineRE.FindStringSubmatch(s); m != nil {
		n := stripInlineNicknames(m[1])
		c := strings.ToUpper(m[2])
		v, _ := strconv.Atoi(m[3])
		n, c = resolveNickname(n, c)
		return n, c, &v, ""
	}
	if m := thcSpaceRE.FindStringSubmatch(s); m != nil {
		n := stripInlineNicknames(m[1])
		c := strings.ToUpper(m[2])
		v, _ := strconv.Atoi(m[3])
		n, c = resolveNickname(n, c)
		return n, c, &v, ""
	}
	if m := thcNoScoreRE.FindStringSubmatch(s); m != nil {
		n := stripInlineNicknames(m[1])
		c := strings.ToUpper(m[2])
		n, c = resolveNickname(n, c)
		return n, c, nil, ""
	}
	n, c := resolveNickname(stripInlineNicknames(s), "")
	return n, c, nil, ""
}

// --- helpers ---

func normalisePosition(raw string) string {
	key := strings.ToUpper(strings.TrimSpace(raw))
	if p, ok := positionAliases[key]; ok {
		return p
	}
	return strings.ToLower(raw)
}

func decodeBenchCode(code string) (backupPositions, interchangePosition string) {
	code = strings.TrimSpace(code)
	upper := strings.ToUpper(code)
	if upper == "*" {
		return "star", ""
	}
	if strings.ReplaceAll(strings.ReplaceAll(upper, " ", ""), "(INT)", "") == "*" {
		return "star", "star"
	}
	parts := strings.Split(upper, "/")
	var positions []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if pos, ok := benchLetter[p]; ok {
			positions = append(positions, pos)
		} else {
			positions = append(positions, strings.ToLower(p))
		}
	}
	return strings.Join(positions, ","), ""
}

func resolveNickname(name, club string) (string, string) {
	lower := strings.ToLower(name)
	for nick, canonical := range nicknames {
		if strings.Contains(lower, nick) {
			c := canonical[1]
			if c == "" {
				c = club
			}
			return canonical[0], c
		}
	}
	return stripInlineNicknames(name), club
}

func stripInlineNicknames(name string) string {
	for _, word := range stripWords {
		name = strings.ReplaceAll(name, word, "")
	}
	return strings.Join(strings.Fields(name), " ")
}

func isArtifact(line string) bool {
	if artifactRE.MatchString(line) || memberNumRE.MatchString(line) {
		return true
	}
	// emoji-only
	onlyEmoji := true
	for _, r := range line {
		if r < 0x2600 || (r > 0x27FF && r < 0x10000) || r > 0x10FFFF {
			if r != ' ' {
				onlyEmoji = false
				break
			}
		}
	}
	return onlyEmoji && strings.TrimSpace(line) != ""
}

func extractScore(s string) *int {
	m := regexp.MustCompile(`\b(\d+)\b`).FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	v, _ := strconv.Atoi(m[1])
	return &v
}

func lastNumber(s string) *int {
	nums := regexp.MustCompile(`\d+`).FindAllString(s, -1)
	if len(nums) == 0 {
		return nil
	}
	v, _ := strconv.Atoi(nums[len(nums)-1])
	return &v
}

func splitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return strings.Split(text, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
