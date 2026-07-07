package afltables

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// Row is one player-match line from a historical season CSV (as written by the
// afltables-export command). It is the ingest input.
type Row struct {
	Season    int
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

// csvColumns is the exact header the export writes; ReadSeasonCSV validates it
// so a schema drift fails loudly rather than mis-mapping columns.
var csvColumns = []string{
	"season", "round", "date", "venue", "home_club", "away_club",
	"club", "player", "kicks", "marks", "handballs", "goals", "behinds", "hitouts", "tackles",
}

// ReadSeasonCSV parses a historical season CSV into rows.
func ReadSeasonCSV(r io.Reader) ([]Row, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = len(csvColumns)

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	for i, want := range csvColumns {
		if i >= len(header) || header[i] != want {
			return nil, fmt.Errorf("unexpected CSV header: got %v, want %v", header, csvColumns)
		}
	}

	var rows []Row
	line := 1
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return nil, fmt.Errorf("read line %d: %w", line, err)
		}
		row, err := rowFromRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func rowFromRecord(rec []string) (Row, error) {
	atoi := func(i int, field string) (int, error) {
		n, err := strconv.Atoi(rec[i])
		if err != nil {
			return 0, fmt.Errorf("%s: %w", field, err)
		}
		return n, nil
	}
	season, err := atoi(0, "season")
	if err != nil {
		return Row{}, err
	}
	stat := func(i int, field string, dst *int) error {
		v, err := atoi(i, field)
		if err != nil {
			return err
		}
		*dst = v
		return nil
	}
	row := Row{
		Season:   season,
		Round:    rec[1],
		Date:     rec[2],
		Venue:    rec[3],
		HomeClub: rec[4],
		AwayClub: rec[5],
		Club:     rec[6],
		Player:   rec[7],
	}
	for _, f := range []struct {
		i     int
		field string
		dst   *int
	}{
		{8, "kicks", &row.Kicks},
		{9, "marks", &row.Marks},
		{10, "handballs", &row.Handballs},
		{11, "goals", &row.Goals},
		{12, "behinds", &row.Behinds},
		{13, "hitouts", &row.Hitouts},
		{14, "tackles", &row.Tackles},
	} {
		if err := stat(f.i, f.field, f.dst); err != nil {
			return Row{}, err
		}
	}
	return row, nil
}
