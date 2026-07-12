// Command afltables-export scrapes historical AFL match stats from
// afltables.com and writes one CSV per season into an output directory.
//
// It is the "produce CSV" half of the Phase 24 historical import (Option B: a
// CSV intermediate the ingest CLI later reads). Player names are normalised to
// "First Last"; empty stats are 0. Reconciling names to existing AFL players is
// the ingest step's job, not this tool's.
//
//	cd services/afl
//	go run ./cmd/afltables-export -from 2023 -to 2023 -out ../../afl-historical
//
// -from/-to are seasons; the tool walks from -from down to -to (inclusive) so
// you can work backwards. Be polite: it sleeps between requests.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"xffl/services/afl/internal/infrastructure/afltables"
)

const baseURL = "https://afltables.com/afl/"

func main() {
	from := flag.Int("from", 0, "start season (walks backwards to -to)")
	to := flag.Int("to", 0, "end season (inclusive)")
	out := flag.String("out", "afl-historical", "output directory for <season>.csv files")
	delay := flag.Duration("delay", 400*time.Millisecond, "delay between HTTP requests")
	flag.Parse()

	if *from == 0 || *to == 0 || *to > *from {
		fmt.Fprintln(os.Stderr, "usage: afltables-export -from YYYY -to YYYY [-out dir] [-delay 400ms]")
		os.Exit(2)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fatal("create out dir: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	for season := *from; season >= *to; season-- {
		if err := exportSeason(client, season, *out, *delay); err != nil {
			fatal("season %d: %v", season, err)
		}
	}
}

func exportSeason(client *http.Client, season int, outDir string, delay time.Duration) error {
	indexURL := fmt.Sprintf("%sseas/%d.html", baseURL, season)
	fmt.Printf("season %d: fetching index %s\n", season, indexURL)
	body, err := get(client, indexURL)
	if err != nil {
		return err
	}
	paths, err := afltables.ParseSeasonIndex(body)
	body.Close()
	if err != nil {
		return fmt.Errorf("parse index: %w", err)
	}
	fmt.Printf("season %d: %d games\n", season, len(paths))

	path := filepath.Join(outDir, fmt.Sprintf("%d.csv", season))
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(header()); err != nil {
		return err
	}

	games, rows := 0, 0
	for _, p := range paths {
		time.Sleep(delay)
		gameURL := baseURL + p
		gb, err := get(client, gameURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skip %s: %v\n", p, err)
			continue
		}
		g, err := afltables.ParseGamePage(gb)
		gb.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skip %s: %v\n", p, err)
			continue
		}
		for _, pl := range g.Players {
			if err := w.Write(record(season, g, pl)); err != nil {
				return err
			}
			rows++
		}
		games++
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	fmt.Printf("season %d: wrote %d rows from %d games → %s\n", season, rows, games, path)
	return nil
}

func header() []string {
	return []string{
		"season", "round", "date", "venue", "home_club", "away_club",
		"club", "player", "kicks", "marks", "handballs", "goals", "behinds", "hitouts", "tackles",
	}
}

func record(season int, g afltables.Game, p afltables.PlayerLine) []string {
	return []string{
		strconv.Itoa(season), g.Round, g.Date, g.Venue, g.HomeClub, g.AwayClub,
		p.Club, p.Name,
		strconv.Itoa(p.Kicks), strconv.Itoa(p.Marks), strconv.Itoa(p.Handballs),
		strconv.Itoa(p.Goals), strconv.Itoa(p.Behinds), strconv.Itoa(p.Hitouts), strconv.Itoa(p.Tackles),
	}
}

func get(client *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "xffl-historical-import/1.0 (one-time backfill; contact repo owner)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
