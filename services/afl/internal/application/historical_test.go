package application

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHistoricalRepo is an in-memory HistoricalRepo for exercising the
// resolution branches without a database.
type fakeHistoricalRepo struct {
	players    []PlayerRef
	nextPlayer int
	nextID     int
	xref       map[string]int // "source|season|club|player" -> player_season_id
	// captured writes
	playerMatches []PlayerMatchInput
}

func newFakeRepo(existing ...PlayerRef) *fakeHistoricalRepo {
	f := &fakeHistoricalRepo{xref: map[string]int{}, nextPlayer: 100, nextID: 1}
	f.players = append(f.players, existing...)
	return f
}

func (f *fakeHistoricalRepo) id() int { f.nextID++; return f.nextID }

func (f *fakeHistoricalRepo) UpsertLeague(context.Context, string) (int, error)      { return 1, nil }
func (f *fakeHistoricalRepo) GetOrCreateSeason(context.Context, int, string) (int, error) { return 1, nil }
func (f *fakeHistoricalRepo) GetOrCreateRound(context.Context, int, string) (int, error)  { return f.id(), nil }
func (f *fakeHistoricalRepo) UpsertClub(context.Context, string) (int, error)             { return f.id(), nil }
func (f *fakeHistoricalRepo) GetOrCreateClubSeason(context.Context, int, int) (int, error) { return f.id(), nil }
func (f *fakeHistoricalRepo) FindMatchByRoundAndHomeClubSeason(context.Context, int, int) (int, bool, error) {
	return 0, false, nil
}
func (f *fakeHistoricalRepo) InsertMatch(context.Context, int, string, time.Time) (int, error) {
	return f.id(), nil
}
func (f *fakeHistoricalRepo) UpsertClubMatch(context.Context, int, int, string) (int, error) {
	return f.id(), nil
}
func (f *fakeHistoricalRepo) FindPlayersByExactName(_ context.Context, name string) ([]PlayerRef, error) {
	var out []PlayerRef
	for _, p := range f.players {
		if p.Name == name {
			out = append(out, p)
		}
	}
	return out, nil
}
func (f *fakeHistoricalRepo) AllPlayers(context.Context) ([]PlayerRef, error) { return f.players, nil }
func (f *fakeHistoricalRepo) CreatePlayer(_ context.Context, name string) (int, error) {
	f.nextPlayer++
	f.players = append(f.players, PlayerRef{ID: f.nextPlayer, Name: name})
	return f.nextPlayer, nil
}
func (f *fakeHistoricalRepo) GetOrCreatePlayerSeason(_ context.Context, playerID, _ int) (int, error) {
	return 900000 + playerID, nil // deterministic ps id from player id
}
func (f *fakeHistoricalRepo) UpsertPlayerMatch(_ context.Context, p PlayerMatchInput) error {
	f.playerMatches = append(f.playerMatches, p)
	return nil
}
func (f *fakeHistoricalRepo) FindXref(_ context.Context, source, season, club, player string) (int, bool, error) {
	id, ok := f.xref[source+"|"+season+"|"+club+"|"+player]
	return id, ok, nil
}
func (f *fakeHistoricalRepo) StoreXref(_ context.Context, source, season, club, player string, psID int) error {
	f.xref[source+"|"+season+"|"+club+"|"+player] = psID
	return nil
}

// recordingPrompter returns a fixed decision and records how many times it was asked.
type recordingPrompter struct {
	returnID int
	calls    int
	lastCand []PlayerChoice
}

func (p *recordingPrompter) Choose(_ context.Context, _, _, _ string, cands []PlayerChoice) (int, error) {
	p.calls++
	p.lastCand = cands
	return p.returnID, nil
}

type capturingLog struct {
	newPlayers []string
	nearMisses []string
}

func (l *capturingLog) NewPlayer(_, _, name string, _ int) { l.newPlayers = append(l.newPlayers, name) }
func (l *capturingLog) NearMiss(_, _, name, cand string, _ float64) {
	l.nearMisses = append(l.nearMisses, name+"~"+cand)
}

func row(club, player string) HistoricalRow {
	return HistoricalRow{Round: "1", Date: "2020-03-20", Venue: "MCG", HomeClub: "Richmond", AwayClub: "Carlton", Club: club, Player: player, Kicks: 10}
}

func runImport(t *testing.T, repo *fakeHistoricalRepo, prompter PlayerPrompter, log HistoricalReviewLogger, rows []HistoricalRow) ImportSeasonSummary {
	t.Helper()
	imp := NewHistoricalImporter(repo, resolverStub{}, prompter, log)
	sum, err := imp.ImportSeason(context.Background(), 2020, rows)
	require.NoError(t, err)
	return sum
}

// resolverStub scores only exact (normalised) name equality — enough to test
// that fuzzy does NOT fire on unrelated names.
type resolverStub struct{}

func (resolverStub) Resolve(_ context.Context, name, _ string, cands []PlayerCandidate) ([]PlayerMatch, error) {
	out := make([]PlayerMatch, 0, len(cands))
	for _, c := range cands {
		conf := 0.0
		if c.Name == name {
			conf = 1.0
		}
		out = append(out, PlayerMatch{Candidate: c, Confidence: conf})
	}
	// crude sort: put any 1.0 first
	for i := range out {
		if out[i].Confidence == 1.0 {
			out[0], out[i] = out[i], out[0]
			break
		}
	}
	return out, nil
}

func TestImport_ExactUniqueAutoLinks(t *testing.T) {
	repo := newFakeRepo(PlayerRef{ID: 50, Name: "Dustin Martin"})
	prompter := &recordingPrompter{}
	log := &capturingLog{}
	sum := runImport(t, repo, prompter, log, []HistoricalRow{row("Richmond", "Dustin Martin")})

	assert.Equal(t, 0, prompter.calls, "exact-unique must not prompt")
	assert.Equal(t, 0, sum.NewPlayers)
	require.Len(t, repo.playerMatches, 1)
	assert.Equal(t, 900000+50, repo.playerMatches[0].PlayerSeasonID)
}

func TestImport_NoMatchAutoCreates(t *testing.T) {
	repo := newFakeRepo()
	prompter := &recordingPrompter{}
	log := &capturingLog{}
	sum := runImport(t, repo, prompter, log, []HistoricalRow{row("Richmond", "Nobody Known")})

	assert.Equal(t, 0, prompter.calls, "brand-new player must not prompt")
	assert.Equal(t, 1, sum.NewPlayers)
	assert.Equal(t, []string{"Nobody Known"}, log.newPlayers)
}

func TestImport_DuplicateExactNamesPrompt(t *testing.T) {
	repo := newFakeRepo(
		PlayerRef{ID: 60, Name: "Josh Kennedy"},
		PlayerRef{ID: 61, Name: "Josh Kennedy"},
	)
	prompter := &recordingPrompter{returnID: 61}
	log := &capturingLog{}
	sum := runImport(t, repo, prompter, log, []HistoricalRow{row("Sydney", "Josh Kennedy")})

	assert.Equal(t, 1, prompter.calls, "duplicate exact names must prompt")
	assert.Len(t, prompter.lastCand, 2)
	assert.Equal(t, 1, sum.Prompted)
	assert.Equal(t, 0, sum.NewPlayers)
	require.Len(t, repo.playerMatches, 1)
	assert.Equal(t, 900000+61, repo.playerMatches[0].PlayerSeasonID, "links to operator's choice")
}

func TestImport_XrefShortCircuits(t *testing.T) {
	repo := newFakeRepo()
	repo.xref["afltables|AFL 2020|Richmond|Prior Decision"] = 12345
	prompter := &recordingPrompter{}
	sum := runImport(t, repo, prompter, &capturingLog{}, []HistoricalRow{row("Richmond", "Prior Decision")})

	assert.Equal(t, 0, prompter.calls)
	assert.Equal(t, 0, sum.NewPlayers)
	require.Len(t, repo.playerMatches, 1)
	assert.Equal(t, 12345, repo.playerMatches[0].PlayerSeasonID)
}

func TestImport_SameNameResolvedOncePerRun(t *testing.T) {
	repo := newFakeRepo()
	prompter := &recordingPrompter{}
	log := &capturingLog{}
	// Same new player appears in two matches; should be created once.
	sum := runImport(t, repo, prompter, log, []HistoricalRow{
		{Round: "1", Club: "Richmond", HomeClub: "Richmond", AwayClub: "Carlton", Player: "Repeat Guy"},
		{Round: "2", Club: "Richmond", HomeClub: "Richmond", AwayClub: "Geelong", Player: "Repeat Guy"},
	})
	assert.Equal(t, 1, sum.NewPlayers)
	assert.Equal(t, 2, sum.PlayerMatches)
	assert.Len(t, log.newPlayers, 1)
}
