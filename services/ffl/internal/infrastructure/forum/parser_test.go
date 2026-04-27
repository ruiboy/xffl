package forum

import (
	"context"
	"os"
	"testing"

	"xffl/services/ffl/internal/application"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read testdata/%s: %v", name, err)
	}
	return string(b)
}

func findRow(rows []application.ParsedPlayerRow, name string) *application.ParsedPlayerRow {
	for i := range rows {
		if rows[i].Name == name {
			return &rows[i]
		}
	}
	return nil
}

func TestParseRuiboys(t *testing.T) {
	post := readTestdata(t, "ruiboys.txt")
	p := NewParser()
	rows, err := p.Parse(context.Background(), "", post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 22 {
		t.Errorf("row count: got %d, want 22", len(rows))
	}

	// starter: goals player
	r := findRow(rows, "Jeremy Cameron")
	if r == nil {
		t.Fatal("Jeremy Cameron not found")
	}
	if r.Position != "goals" {
		t.Errorf("position: got %q, want %q", r.Position, "goals")
	}
	if r.ClubHint != "Geel" {
		t.Errorf("club: got %q, want %q", r.ClubHint, "Geel")
	}
	if r.Score == nil || *r.Score != 15 {
		t.Errorf("score: got %v, want 15", r.Score)
	}

	// star player substituted mid-game
	r = findRow(rows, "Jye Caldwell")
	if r == nil {
		t.Fatal("Jye Caldwell not found")
	}
	if r.Score == nil || *r.Score != 52 {
		t.Errorf("Jye Caldwell score: got %v, want 52 (post-sub score)", r.Score)
	}

	// interchange bench player: * (INT) → backupPositions=star, interchangePosition=star
	r = findRow(rows, "Hugh McCluggage")
	if r == nil {
		t.Fatal("Hugh McCluggage not found")
	}
	if r.BackupPositions != "star" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "star")
	}
	if r.InterchangePosition != "star" {
		t.Errorf("interchangePosition: got %q, want %q", r.InterchangePosition, "star")
	}
	if r.Score == nil || *r.Score != 52 {
		t.Errorf("Hugh McCluggage score: got %v, want 52", r.Score)
	}

	// bench with position code: K/M → kicks,marks
	r = findRow(rows, "Karl Amon")
	if r == nil {
		t.Fatal("Karl Amon not found")
	}
	if r.BackupPositions != "kicks,marks" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "kicks,marks")
	}
	if r.InterchangePosition != "" {
		t.Errorf("interchangePosition: got %q, want %q", r.InterchangePosition, "")
	}
}

func TestParseSlashers(t *testing.T) {
	post := readTestdata(t, "slashers.txt")
	p := NewParser()
	rows, err := p.Parse(context.Background(), "", post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 22 {
		t.Errorf("row count: got %d, want 22", len(rows))
	}

	// kicks starter
	r := findRow(rows, "Z Merrett")
	if r == nil {
		t.Fatal("Z Merrett not found")
	}
	if r.Position != "kicks" {
		t.Errorf("position: got %q, want %q", r.Position, "kicks")
	}
	if r.ClubHint != "Ess" {
		t.Errorf("club: got %q, want %q", r.ClubHint, "Ess")
	}
	if r.Score == nil || *r.Score != 10 {
		t.Errorf("score: got %v, want 10", r.Score)
	}

	// RUCK → hitouts
	r = findRow(rows, "B Grundy")
	if r == nil {
		t.Fatal("B Grundy not found")
	}
	if r.Position != "hitouts" {
		t.Errorf("position: got %q, want %q", r.Position, "hitouts")
	}

	// star player bumped by interchange score
	r = findRow(rows, "M Holmes")
	if r == nil {
		t.Fatal("M Holmes not found")
	}
	if r.Score == nil || *r.Score != 70 {
		t.Errorf("M Holmes score: got %v, want 70 (interchanged score)", r.Score)
	}

	// interchange player: ***A Brayshaw*** → star + star
	r = findRow(rows, "A Brayshaw")
	if r == nil {
		t.Fatal("A Brayshaw not found")
	}
	if r.BackupPositions != "star" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "star")
	}
	if r.InterchangePosition != "star" {
		t.Errorf("interchangePosition: got %q, want %q", r.InterchangePosition, "star")
	}
	if r.Score == nil || *r.Score != 70 {
		t.Errorf("A Brayshaw score: got %v, want 70", r.Score)
	}

	// bench code R/T → hitouts,tackles
	r = findRow(rows, "S Flanders")
	if r == nil {
		t.Fatal("S Flanders not found")
	}
	if r.BackupPositions != "hitouts,tackles" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "hitouts,tackles")
	}
}

func TestParseCheetahs(t *testing.T) {
	post := readTestdata(t, "cheetahs.txt")
	p := NewParser()
	rows, err := p.Parse(context.Background(), "", post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 22 {
		t.Errorf("row count: got %d, want 22", len(rows))
	}

	// goals raw stat multiplied ×5
	r := findRow(rows, "Ben King")
	if r == nil {
		t.Fatal("Ben King not found")
	}
	if r.Position != "goals" {
		t.Errorf("position: got %q, want %q", r.Position, "goals")
	}
	if r.Score == nil || *r.Score != 10 {
		t.Errorf("Ben King score: got %v, want 10 (2 goals × 5)", r.Score)
	}

	// marks raw stat multiplied ×2
	r = findRow(rows, "Nick Haynes")
	if r == nil {
		t.Fatal("Nick Haynes not found")
	}
	if r.Score == nil || *r.Score != 14 {
		t.Errorf("Nick Haynes score: got %v, want 14 (7 marks × 2)", r.Score)
	}

	// star player
	r = findRow(rows, "Will Ashcroft")
	if r == nil {
		t.Fatal("Will Ashcroft not found")
	}
	if r.Score == nil || *r.Score != 65 {
		t.Errorf("Will Ashcroft score: got %v, want 65", r.Score)
	}

	// bench * → star; then "Interchange = * *" sets interchangePosition
	r = findRow(rows, "Harry Sheezel")
	if r == nil {
		t.Fatal("Harry Sheezel not found")
	}
	if r.BackupPositions != "star" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "star")
	}
	if r.InterchangePosition != "star" {
		t.Errorf("interchangePosition: got %q, want %q", r.InterchangePosition, "star")
	}

	// bench code T/H → tackles,handballs
	r = findRow(rows, "Hugo Garcia")
	if r == nil {
		t.Fatal("Hugo Garcia not found")
	}
	if r.BackupPositions != "tackles,handballs" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "tackles,handballs")
	}
}

func TestParseTHC(t *testing.T) {
	post := readTestdata(t, "thc.txt")
	p := NewParser()
	rows, err := p.Parse(context.Background(), "", post)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 22 {
		t.Errorf("row count: got %d, want 22", len(rows))
	}

	// HB section → handballs position
	r := findRow(rows, "Touk Miller")
	if r == nil {
		t.Fatal("Touk Miller not found")
	}
	if r.Position != "handballs" {
		t.Errorf("position: got %q, want %q", r.Position, "handballs")
	}
	if r.ClubHint != "GCS" {
		t.Errorf("club: got %q, want %q", r.ClubHint, "GCS")
	}
	if r.Score == nil || *r.Score != 11 {
		t.Errorf("score: got %v, want 11", r.Score)
	}

	// star player
	r = findRow(rows, "Marcus Bontempelli")
	if r == nil {
		t.Fatal("Marcus Bontempelli not found")
	}
	if r.Score == nil || *r.Score != 65 {
		t.Errorf("Marcus Bontempelli score: got %v, want 65", r.Score)
	}

	// I/C- Star → "Star- Toby Greene GWS- 60" → bench, bp=star, ic=star
	r = findRow(rows, "Toby Greene")
	if r == nil {
		t.Fatal("Toby Greene not found")
	}
	if r.Position != "bench" {
		t.Errorf("position: got %q, want %q", r.Position, "bench")
	}
	if r.BackupPositions != "star" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "star")
	}
	if r.InterchangePosition != "star" {
		t.Errorf("interchangePosition: got %q, want %q", r.InterchangePosition, "star")
	}
	if r.Score == nil || *r.Score != 60 {
		t.Errorf("Toby Greene score: got %v, want 60", r.Score)
	}

	// bench code K/HB → kicks,handballs
	r = findRow(rows, "Nick Blakey")
	if r == nil {
		t.Fatal("Nick Blakey not found")
	}
	if r.BackupPositions != "kicks,handballs" {
		t.Errorf("backupPositions: got %q, want %q", r.BackupPositions, "kicks,handballs")
	}
}
