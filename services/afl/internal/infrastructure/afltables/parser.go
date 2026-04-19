package afltables

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

// PlayerMatchStats represents a single player's stats for one match as
// parsed from the afltables CSV format.
type PlayerMatchStats struct {
	ExternalPlayerID string // afltables' own numeric player ID
	PlayerName       string
	ClubName         string // canonical afl.club.name, resolved from team code
	RoundName        string // e.g. "Round 1", "Opening Round"
	SeasonYear       int
	Kicks            int
	Handballs        int
	Marks            int
	Hitouts          int
	Tackles          int
	Goals            int
	Behinds          int
}

// parseCSV reads the afltables stats CSV and returns one PlayerMatchStats per
// data row. Rows with unrecognised club codes or round labels are skipped with
// a warning — they do not cause the whole parse to fail.
func parseCSV(r io.Reader, year int) ([]PlayerMatchStats, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	idx, err := buildIndex(header)
	if err != nil {
		return nil, err
	}

	var results []PlayerMatchStats
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		row, ok := parseRow(record, idx, year)
		if !ok {
			continue
		}
		results = append(results, row)
	}
	return results, nil
}

// columnIndex maps column names to their position in the header.
type columnIndex struct {
	player, id, team, round                        int
	kicks, marks, handballs, goals, behinds        int
	hitouts, tackles                               int
}

func buildIndex(header []string) (columnIndex, error) {
	pos := make(map[string]int, len(header))
	for i, h := range header {
		pos[h] = i
	}
	required := []string{"Player", "ID", "Team", "Round", "Kicks", "Marks", "Hand Balls", "Goals", "Behinds", "Hit Outs", "Tackles"}
	for _, col := range required {
		if _, ok := pos[col]; !ok {
			return columnIndex{}, fmt.Errorf("CSV missing required column %q", col)
		}
	}
	return columnIndex{
		player:   pos["Player"],
		id:       pos["ID"],
		team:     pos["Team"],
		round:    pos["Round"],
		kicks:    pos["Kicks"],
		marks:    pos["Marks"],
		handballs: pos["Hand Balls"],
		goals:    pos["Goals"],
		behinds:  pos["Behinds"],
		hitouts:  pos["Hit Outs"],
		tackles:  pos["Tackles"],
	}, nil
}

func parseRow(record []string, idx columnIndex, year int) (PlayerMatchStats, bool) {
	clubCode := record[idx.team]
	clubName, ok := ClubNameForCode(clubCode)
	if !ok {
		slog.Warn("afltables: skipping row with unknown club code", "code", clubCode, "player", record[idx.player])
		return PlayerMatchStats{}, false
	}

	roundName, ok := roundNameForCode(record[idx.round])
	if !ok {
		slog.Warn("afltables: skipping row with unrecognised round", "round", record[idx.round], "player", record[idx.player])
		return PlayerMatchStats{}, false
	}

	return PlayerMatchStats{
		ExternalPlayerID: record[idx.id],
		PlayerName:       record[idx.player],
		ClubName:         clubName,
		RoundName:        roundName,
		SeasonYear:       year,
		Kicks:            parseInt(record[idx.kicks]),
		Marks:            parseInt(record[idx.marks]),
		Handballs:        parseInt(record[idx.handballs]),
		Goals:            parseInt(record[idx.goals]),
		Behinds:          parseInt(record[idx.behinds]),
		Hitouts:          parseInt(record[idx.hitouts]),
		Tackles:          parseInt(record[idx.tackles]),
	}, true
}

// roundNameForCode maps the afltables round column to a domain round name.
// Regular rounds are numeric strings ("1"–"24"); Opening Round is "0".
// Finals rounds and any other unrecognised codes return false.
func roundNameForCode(code string) (string, bool) {
	if code == "0" {
		return "Opening Round", true
	}
	n, err := strconv.Atoi(code)
	if err != nil || n < 1 || n > 24 {
		return "", false
	}
	return fmt.Sprintf("Round %d", n), true
}

// parseInt parses a string to int, returning 0 on failure.
func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
