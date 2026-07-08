// Command afltables-import ingests the historical CSVs (produced by
// afltables-export) into the AFL schema, creating the season→player_match
// scaffold and resolving player identity interactively.
//
// It processes seasons newest→oldest so a real career chains backward from the
// seeded modern seasons and a genuine gap (same name, non-consecutive seasons)
// stands out. Exact names adjacent to a known career auto-link; brand-new names
// auto-create; and ambiguity — a season gap, duplicate exact names, or a
// high-confidence fuzzy near-match — prompts on stdin. Every new player and
// fuzzy near-miss is appended to a review log.
//
//	cd services/afl
//	DATABASE_URL=... go run ./cmd/afltables-import -from 1998 -to 2023
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"xffl/services/afl/internal/application"
	"xffl/services/afl/internal/infrastructure/afltables"
	"xffl/services/afl/internal/infrastructure/footywire"
	pg "xffl/services/afl/internal/infrastructure/postgres"
)

func main() {
	from := flag.Int("from", 0, "first season (inclusive)")
	to := flag.Int("to", 0, "last season (inclusive); ingested oldest→newest")
	dir := flag.String("dir", "afl-historical", "directory of <season>.csv files")
	reviewPath := flag.String("review", filepath.Join("afl-historical", "import-review.log"), "review log path")
	flag.Parse()

	if *from == 0 || *to == 0 || *from > *to {
		fmt.Fprintln(os.Stderr, "usage: afltables-import -from YYYY -to YYYY [-dir afl-historical] [-review path]")
		fmt.Fprintln(os.Stderr, "seasons are ingested oldest→newest so player identity builds forward")
		os.Exit(2)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xffl?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pg.NewPool(ctx, dbURL)
	if err != nil {
		fatal("connect DB: %v", err)
	}
	defer pool.Close()

	review, err := newReviewLog(*reviewPath)
	if err != nil {
		fatal("open review log: %v", err)
	}
	defer review.Close()

	importer := application.NewHistoricalImporter(
		pg.NewHistoricalRepository(pool),
		footywire.NewLevenshteinResolver(),
		&stdinPrompter{in: bufio.NewReader(os.Stdin)},
		review,
	)

	// Newest→oldest: a real career chains backward from the seeded modern
	// seasons with no false prompts, so a genuine season gap stands out.
	for year := *to; year >= *from; year-- {
		rows, err := loadSeason(*dir, year)
		if err != nil {
			fatal("season %d: %v", year, err)
		}
		fmt.Printf("\n=== Importing %d (%d rows) ===\n", year, len(rows))
		sum, err := importer.ImportSeason(ctx, year, rows)
		if err != nil {
			fatal("import %d: %v", year, err)
		}
		fmt.Printf("season %d: %d matches, %d player-matches, %d new players, %d gaps-logged, %d prompted\n",
			sum.Season, sum.Matches, sum.PlayerMatches, sum.NewPlayers, sum.Gaps, sum.Prompted)
	}
	fmt.Printf("\nDone. Review log: %s\n", *reviewPath)
}

func loadSeason(dir string, year int) ([]application.HistoricalRow, error) {
	path := filepath.Join(dir, fmt.Sprintf("%d.csv", year))
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	csvRows, err := afltables.ReadSeasonCSV(f)
	if err != nil {
		return nil, err
	}
	out := make([]application.HistoricalRow, len(csvRows))
	for i, r := range csvRows {
		out[i] = application.HistoricalRow{
			Round: r.Round, Date: r.Date, Venue: r.Venue,
			// Map afltables short names to canonical club names so historical
			// data attaches to existing club records, not duplicates.
			HomeClub: afltables.CanonicalClub(r.HomeClub),
			AwayClub: afltables.CanonicalClub(r.AwayClub),
			Club:     afltables.CanonicalClub(r.Club),
			Player:   r.Player,
			Kicks:    r.Kicks, Marks: r.Marks, Handballs: r.Handballs,
			Goals: r.Goals, Behinds: r.Behinds, Hitouts: r.Hitouts, Tackles: r.Tackles,
		}
	}
	return out, nil
}

// --- stdin prompter ---

type stdinPrompter struct{ in *bufio.Reader }

func (p *stdinPrompter) Choose(_ context.Context, name, club, season string, candidates []application.PlayerChoice) (int, error) {
	fmt.Printf("\nAMBIGUOUS: %q  (%s, %s)\n", name, club, season)
	for i, c := range candidates {
		suffix := ""
		if c.Detail != "" {
			suffix = " — " + c.Detail
		}
		if c.Confidence > 0 {
			suffix += fmt.Sprintf(" (%.0f%% name match)", c.Confidence*100)
		}
		fmt.Printf("  [%d] %s (id %d)%s\n", i+1, c.Name, c.PlayerID, suffix)
	}
	fmt.Printf("  [n] create new player\n")
	for {
		fmt.Printf("Choose [1-%d/n]: ", len(candidates))
		line, err := p.in.ReadString('\n')
		if err != nil {
			// No input available (non-interactive run): default to a new player,
			// which is always safe — it never silently links the wrong record.
			fmt.Println("(no input — creating new player)")
			return 0, nil
		}
		line = strings.TrimSpace(line)
		if line == "n" || line == "N" {
			return 0, nil
		}
		if n, err := strconv.Atoi(line); err == nil && n >= 1 && n <= len(candidates) {
			return candidates[n-1].PlayerID, nil
		}
		fmt.Println("  invalid choice")
	}
}

// --- review log ---

type reviewLog struct{ f *os.File }

func newReviewLog(path string) (*reviewLog, error) {
	if dir := filepath.Dir(path); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(f, "\n# import run %s\n", time.Now().Format(time.RFC3339))
	return &reviewLog{f: f}, nil
}

func (l *reviewLog) NewPlayer(season, club, name string, playerID int) {
	fmt.Fprintf(l.f, "NEW\t%s\t%s\t%s\t(id %d)\n", season, club, name, playerID)
}

func (l *reviewLog) NearMiss(season, club, name, candidateName string, confidence float64) {
	fmt.Fprintf(l.f, "NEARMISS\t%s\t%s\t%q ~ %q\t(%.0f%%)\n", season, club, name, candidateName, confidence*100)
}

// Gap lines are sortable big-gap-first with: grep '^GAP' log | sort -t$'\t' -k2 -rn
func (l *reviewLog) Gap(name string, year, missedSeasons int, existingSpan string, playerID int) {
	fmt.Fprintf(l.f, "GAP\t%d\t%q\t%d missed → linked to id %d (%s)\n", missedSeasons, name, year, playerID, existingSpan)
}

func (l *reviewLog) Close() error { return l.f.Close() }

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
