package forum

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"xffl/services/ffl/internal/application"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
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
	rows, err := NewParser().Parse(context.Background(), "", post)
	require.NoError(t, err)
	assert.Len(t, rows, 22)

	r := findRow(rows, "Jeremy Cameron")
	require.NotNil(t, r, "Jeremy Cameron not found")
	assert.Equal(t, "goals", r.Position)
	assert.Equal(t, "Geel", r.ClubHint)
	require.NotNil(t, r.Score)
	assert.Equal(t, 15, *r.Score)

	r = findRow(rows, "Jye Caldwell")
	require.NotNil(t, r, "Jye Caldwell not found")
	require.NotNil(t, r.Score)
	assert.Equal(t, 52, *r.Score, "post-sub score")

	r = findRow(rows, "Hugh McCluggage")
	require.NotNil(t, r, "Hugh McCluggage not found")
	assert.Equal(t, "star", r.BackupPositions)
	assert.Equal(t, "star", r.InterchangePosition)
	require.NotNil(t, r.Score)
	assert.Equal(t, 52, *r.Score)

	r = findRow(rows, "Karl Amon")
	require.NotNil(t, r, "Karl Amon not found")
	assert.Equal(t, "kicks,marks", r.BackupPositions)
	assert.Equal(t, "", r.InterchangePosition)
}

func TestParseSlashers(t *testing.T) {
	post := readTestdata(t, "slashers.txt")
	rows, err := NewParser().Parse(context.Background(), "", post)
	require.NoError(t, err)
	assert.Len(t, rows, 22)

	r := findRow(rows, "Z Merrett")
	require.NotNil(t, r, "Z Merrett not found")
	assert.Equal(t, "kicks", r.Position)
	assert.Equal(t, "Ess", r.ClubHint)
	require.NotNil(t, r.Score)
	assert.Equal(t, 10, *r.Score)

	r = findRow(rows, "B Grundy")
	require.NotNil(t, r, "B Grundy not found")
	assert.Equal(t, "hitouts", r.Position, "RUCK → hitouts")

	r = findRow(rows, "M Holmes")
	require.NotNil(t, r, "M Holmes not found")
	require.NotNil(t, r.Score)
	assert.Equal(t, 70, *r.Score, "star score bumped by interchange")

	r = findRow(rows, "A Brayshaw")
	require.NotNil(t, r, "A Brayshaw not found")
	assert.Equal(t, "star", r.BackupPositions)
	assert.Equal(t, "star", r.InterchangePosition)
	require.NotNil(t, r.Score)
	assert.Equal(t, 70, *r.Score)

	r = findRow(rows, "S Flanders")
	require.NotNil(t, r, "S Flanders not found")
	assert.Equal(t, "hitouts,tackles", r.BackupPositions)
}

func TestParseCheetahs(t *testing.T) {
	post := readTestdata(t, "cheetahs.txt")
	rows, err := NewParser().Parse(context.Background(), "", post)
	require.NoError(t, err)
	assert.Len(t, rows, 22)

	r := findRow(rows, "Ben King")
	require.NotNil(t, r, "Ben King not found")
	assert.Equal(t, "goals", r.Position)
	require.NotNil(t, r.Score)
	assert.Equal(t, 10, *r.Score, "2 goals × 5")

	r = findRow(rows, "Nick Haynes")
	require.NotNil(t, r, "Nick Haynes not found")
	require.NotNil(t, r.Score)
	assert.Equal(t, 14, *r.Score, "7 marks × 2")

	r = findRow(rows, "Will Ashcroft")
	require.NotNil(t, r, "Will Ashcroft not found")
	require.NotNil(t, r.Score)
	assert.Equal(t, 65, *r.Score)

	r = findRow(rows, "Harry Sheezel")
	require.NotNil(t, r, "Harry Sheezel not found")
	assert.Equal(t, "star", r.BackupPositions)
	assert.Equal(t, "star", r.InterchangePosition, "set by Interchange = * *")

	r = findRow(rows, "Hugo Garcia")
	require.NotNil(t, r, "Hugo Garcia not found")
	assert.Equal(t, "tackles,handballs", r.BackupPositions)
}

func TestParseTHC(t *testing.T) {
	post := readTestdata(t, "thc.txt")
	rows, err := NewParser().Parse(context.Background(), "", post)
	require.NoError(t, err)
	assert.Len(t, rows, 22)

	r := findRow(rows, "Touk Miller")
	require.NotNil(t, r, "Touk Miller not found")
	assert.Equal(t, "handballs", r.Position, "HB section → handballs")
	assert.Equal(t, "GCS", r.ClubHint)
	require.NotNil(t, r.Score)
	assert.Equal(t, 11, *r.Score)

	r = findRow(rows, "Marcus Bontempelli")
	require.NotNil(t, r, "Marcus Bontempelli not found")
	require.NotNil(t, r.Score)
	assert.Equal(t, 65, *r.Score)

	r = findRow(rows, "Toby Greene")
	require.NotNil(t, r, "Toby Greene not found")
	assert.Equal(t, "bench", r.Position)
	assert.Equal(t, "star", r.BackupPositions)
	assert.Equal(t, "star", r.InterchangePosition)
	require.NotNil(t, r.Score)
	assert.Equal(t, 60, *r.Score)

	r = findRow(rows, "Nick Blakey")
	require.NotNil(t, r, "Nick Blakey not found")
	assert.Equal(t, "kicks,handballs", r.BackupPositions)
}
