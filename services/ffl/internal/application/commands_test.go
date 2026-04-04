package application_test

import (
	"context"
	"errors"
	"testing"

	"xffl/services/ffl/internal/application"
	"xffl/services/ffl/internal/domain"
)

// --- mock TxManager ---

type mockTxManager struct {
	repos application.WriteRepos
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(repos application.WriteRepos) error) error {
	return fn(m.repos)
}

// --- mock PlayerRepository ---

type mockPlayerRepo struct {
	players map[int]domain.Player
	nextID  int
}

func newMockPlayerRepo() *mockPlayerRepo {
	return &mockPlayerRepo{players: make(map[int]domain.Player), nextID: 1}
}

func (r *mockPlayerRepo) FindAll(_ context.Context) ([]domain.Player, error) {
	var out []domain.Player
	for _, p := range r.players {
		out = append(out, p)
	}
	return out, nil
}

func (r *mockPlayerRepo) FindByID(_ context.Context, id int) (domain.Player, error) {
	p, ok := r.players[id]
	if !ok {
		return domain.Player{}, errors.New("not found")
	}
	return p, nil
}

func (r *mockPlayerRepo) FindByAFLPlayerID(_ context.Context, aflPlayerID int) (domain.Player, error) {
	for _, p := range r.players {
		if p.AFLPlayerID == aflPlayerID {
			return p, nil
		}
	}
	return domain.Player{}, errors.New("not found")
}

func (r *mockPlayerRepo) Create(_ context.Context, name string, aflPlayerID int) (domain.Player, error) {
	p := domain.Player{ID: r.nextID, Name: name, AFLPlayerID: aflPlayerID}
	r.players[r.nextID] = p
	r.nextID++
	return p, nil
}

func (r *mockPlayerRepo) Update(_ context.Context, id int, name string) (domain.Player, error) {
	p, ok := r.players[id]
	if !ok {
		return domain.Player{}, errors.New("not found")
	}
	p.Name = name
	r.players[id] = p
	return p, nil
}

func (r *mockPlayerRepo) Delete(_ context.Context, id int) error {
	if _, ok := r.players[id]; !ok {
		return errors.New("not found")
	}
	delete(r.players, id)
	return nil
}

// --- mock PlayerSeasonRepository ---

type mockPlayerSeasonRepo struct {
	seasons map[int]domain.PlayerSeason
	nextID  int
}

func newMockPlayerSeasonRepo() *mockPlayerSeasonRepo {
	return &mockPlayerSeasonRepo{seasons: make(map[int]domain.PlayerSeason), nextID: 1}
}

func (r *mockPlayerSeasonRepo) FindByClubSeasonID(_ context.Context, clubSeasonID int) ([]domain.PlayerSeason, error) {
	var out []domain.PlayerSeason
	for _, ps := range r.seasons {
		if ps.ClubSeasonID == clubSeasonID {
			out = append(out, ps)
		}
	}
	return out, nil
}

func (r *mockPlayerSeasonRepo) FindByID(_ context.Context, id int) (domain.PlayerSeason, error) {
	ps, ok := r.seasons[id]
	if !ok {
		return domain.PlayerSeason{}, errors.New("not found")
	}
	return ps, nil
}

func (r *mockPlayerSeasonRepo) Create(_ context.Context, playerID int, clubSeasonID int) (domain.PlayerSeason, error) {
	ps := domain.PlayerSeason{ID: r.nextID, PlayerID: playerID, ClubSeasonID: clubSeasonID}
	r.seasons[r.nextID] = ps
	r.nextID++
	return ps, nil
}

func (r *mockPlayerSeasonRepo) Delete(_ context.Context, id int) error {
	if _, ok := r.seasons[id]; !ok {
		return errors.New("not found")
	}
	delete(r.seasons, id)
	return nil
}

// --- mock PlayerMatchRepository ---

type mockPlayerMatchRepo struct {
	matches map[int]domain.PlayerMatch
	nextID  int
}

func newMockPlayerMatchRepo() *mockPlayerMatchRepo {
	return &mockPlayerMatchRepo{matches: make(map[int]domain.PlayerMatch), nextID: 1}
}

func (r *mockPlayerMatchRepo) FindByClubMatchID(_ context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	var out []domain.PlayerMatch
	for _, pm := range r.matches {
		if pm.ClubMatchID == clubMatchID {
			out = append(out, pm)
		}
	}
	return out, nil
}

func (r *mockPlayerMatchRepo) FindByID(_ context.Context, id int) (domain.PlayerMatch, error) {
	pm, ok := r.matches[id]
	if !ok {
		return domain.PlayerMatch{}, errors.New("not found")
	}
	return pm, nil
}

func (r *mockPlayerMatchRepo) Upsert(_ context.Context, params domain.UpsertPlayerMatchParams) (domain.PlayerMatch, error) {
	// Find existing by ClubMatchID + PlayerSeasonID, or create new.
	for id, pm := range r.matches {
		if pm.ClubMatchID == params.ClubMatchID && pm.PlayerSeasonID == params.PlayerSeasonID {
			pm.Position = params.Position
			pm.Status = params.Status
			pm.BackupPositions = params.BackupPositions
			pm.InterchangePosition = params.InterchangePosition
			if params.Score != nil {
				pm.Score = *params.Score
			}
			r.matches[id] = pm
			return pm, nil
		}
	}
	pm := domain.PlayerMatch{
		ID:                  r.nextID,
		ClubMatchID:         params.ClubMatchID,
		PlayerSeasonID:      params.PlayerSeasonID,
		Position:            params.Position,
		Status:              params.Status,
		BackupPositions:     params.BackupPositions,
		InterchangePosition: params.InterchangePosition,
	}
	if params.Score != nil {
		pm.Score = *params.Score
	}
	r.matches[r.nextID] = pm
	r.nextID++
	return pm, nil
}

// --- mock ClubMatchRepository ---

type mockClubMatchRepo struct {
	matches map[int]domain.ClubMatch
}

func newMockClubMatchRepo() *mockClubMatchRepo {
	return &mockClubMatchRepo{matches: make(map[int]domain.ClubMatch)}
}

func (r *mockClubMatchRepo) FindByMatchID(_ context.Context, matchID int) ([]domain.ClubMatch, error) {
	var out []domain.ClubMatch
	for _, cm := range r.matches {
		if cm.MatchID == matchID {
			out = append(out, cm)
		}
	}
	return out, nil
}

func (r *mockClubMatchRepo) FindByID(_ context.Context, id int) (domain.ClubMatch, error) {
	cm, ok := r.matches[id]
	if !ok {
		return domain.ClubMatch{}, errors.New("not found")
	}
	return cm, nil
}

func (r *mockClubMatchRepo) UpdateScore(_ context.Context, id int, score int) error {
	cm, ok := r.matches[id]
	if !ok {
		return errors.New("not found")
	}
	cm.StoredScore = score
	r.matches[id] = cm
	return nil
}

// --- helper ---

func setupCommands() (*application.Commands, *mockPlayerRepo, *mockPlayerSeasonRepo, *mockPlayerMatchRepo, *mockClubMatchRepo) {
	playerRepo := newMockPlayerRepo()
	playerSeasonRepo := newMockPlayerSeasonRepo()
	playerMatchRepo := newMockPlayerMatchRepo()
	clubMatchRepo := newMockClubMatchRepo()

	tx := &mockTxManager{
		repos: application.WriteRepos{
			Players:       playerRepo,
			PlayerSeasons: playerSeasonRepo,
			PlayerMatches: playerMatchRepo,
			ClubMatches:   clubMatchRepo,
		},
	}
	return application.NewCommands(tx), playerRepo, playerSeasonRepo, playerMatchRepo, clubMatchRepo
}

// --- tests ---

func TestCreatePlayer(t *testing.T) {
	cmds, playerRepo, _, _, _ := setupCommands()
	ctx := context.Background()

	p, err := cmds.CreatePlayer(ctx, "Alice", 100)
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if p.Name != "Alice" {
		t.Errorf("got name %q, want %q", p.Name, "Alice")
	}
	if p.AFLPlayerID != 100 {
		t.Errorf("got AFLPlayerID %d, want 100", p.AFLPlayerID)
	}
	if p.ID == 0 {
		t.Error("expected non-zero ID")
	}

	// Verify stored.
	stored, err := playerRepo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if stored.Name != "Alice" {
		t.Errorf("stored name %q, want %q", stored.Name, "Alice")
	}
}

func TestUpdatePlayer(t *testing.T) {
	cmds, playerRepo, _, _, _ := setupCommands()
	ctx := context.Background()

	p, _ := cmds.CreatePlayer(ctx, "Alice", 100)
	updated, err := cmds.UpdatePlayer(ctx, p.ID, "Bob")
	if err != nil {
		t.Fatalf("UpdatePlayer: %v", err)
	}
	if updated.Name != "Bob" {
		t.Errorf("got name %q, want %q", updated.Name, "Bob")
	}

	stored, _ := playerRepo.FindByID(ctx, p.ID)
	if stored.Name != "Bob" {
		t.Errorf("stored name %q, want %q", stored.Name, "Bob")
	}
}

func TestDeletePlayer(t *testing.T) {
	cmds, playerRepo, _, _, _ := setupCommands()
	ctx := context.Background()

	p, _ := cmds.CreatePlayer(ctx, "Alice", 100)
	if err := cmds.DeletePlayer(ctx, p.ID); err != nil {
		t.Fatalf("DeletePlayer: %v", err)
	}

	_, err := playerRepo.FindByID(ctx, p.ID)
	if err == nil {
		t.Error("expected error after delete, got nil")
	}
}

func TestAddPlayerToSeason(t *testing.T) {
	cmds, _, psRepo, _, _ := setupCommands()
	ctx := context.Background()

	ps, err := cmds.AddPlayerToSeason(ctx, 10, 20)
	if err != nil {
		t.Fatalf("AddPlayerToSeason: %v", err)
	}
	if ps.PlayerID != 10 || ps.ClubSeasonID != 20 {
		t.Errorf("got playerID=%d clubSeasonID=%d, want 10/20", ps.PlayerID, ps.ClubSeasonID)
	}

	stored, _ := psRepo.FindByID(ctx, ps.ID)
	if stored.PlayerID != 10 {
		t.Errorf("stored playerID %d, want 10", stored.PlayerID)
	}
}

func TestRemovePlayerFromSeason(t *testing.T) {
	cmds, _, psRepo, _, _ := setupCommands()
	ctx := context.Background()

	ps, _ := cmds.AddPlayerToSeason(ctx, 10, 20)
	if err := cmds.RemovePlayerFromSeason(ctx, ps.ID); err != nil {
		t.Fatalf("RemovePlayerFromSeason: %v", err)
	}

	_, err := psRepo.FindByID(ctx, ps.ID)
	if err == nil {
		t.Error("expected error after remove, got nil")
	}
}

func TestAddAFLPlayerToRoster_CreatesNewPlayer(t *testing.T) {
	cmds, playerRepo, psRepo, _, _ := setupCommands()
	ctx := context.Background()

	ps, err := cmds.AddAFLPlayerToRoster(ctx, 42, "Lachie Neale", 1)
	if err != nil {
		t.Fatalf("AddAFLPlayerToRoster: %v", err)
	}

	// Should have created a new FFL player linked to AFL player 42
	player, err := playerRepo.FindByAFLPlayerID(ctx, 42)
	if err != nil {
		t.Fatalf("FindByAFLPlayerID: %v", err)
	}
	if player.Name != "Lachie Neale" {
		t.Errorf("player name = %q, want %q", player.Name, "Lachie Neale")
	}
	if player.AFLPlayerID != 42 {
		t.Errorf("player AFLPlayerID = %d, want 42", player.AFLPlayerID)
	}

	// Should have created a player season
	stored, err := psRepo.FindByID(ctx, ps.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if stored.PlayerID != player.ID || stored.ClubSeasonID != 1 {
		t.Errorf("player season playerID=%d clubSeasonID=%d, want %d/1", stored.PlayerID, stored.ClubSeasonID, player.ID)
	}
}

func TestAddAFLPlayerToRoster_ReusesExistingPlayer(t *testing.T) {
	cmds, playerRepo, _, _, _ := setupCommands()
	ctx := context.Background()

	// Pre-create an FFL player linked to AFL player 42
	existing, _ := playerRepo.Create(ctx, "Lachie Neale", 42)

	ps, err := cmds.AddAFLPlayerToRoster(ctx, 42, "Lachie Neale", 1)
	if err != nil {
		t.Fatalf("AddAFLPlayerToRoster: %v", err)
	}

	// Should reuse the existing player, not create a new one
	if ps.PlayerID != existing.ID {
		t.Errorf("playerID = %d, want %d (existing)", ps.PlayerID, existing.ID)
	}

	// Should still only have one player with that AFL ID
	count := 0
	for _, p := range playerRepo.players {
		if p.AFLPlayerID == 42 {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 player with AFL ID 42, got %d", count)
	}
}

func TestCalculateFantasyScore(t *testing.T) {
	cmds, _, _, pmRepo, cmRepo := setupCommands()
	ctx := context.Background()

	// Seed a club match.
	cmRepo.matches[1] = domain.ClubMatch{ID: 1, MatchID: 1, ClubSeasonID: 1}

	// Seed a player match at the "goals" position (5 pts/goal).
	pmRepo.matches[1] = domain.PlayerMatch{
		ID:             1,
		ClubMatchID:    1,
		PlayerSeasonID: 1,
		Position:       domain.PositionPtr(domain.PositionGoals),
		Status:         domain.PlayerMatchStatusPtr(domain.PlayerMatchStatusPlayed),
		Score:          0,
	}

	stats := domain.AFLStats{Goals: 3, Kicks: 10, Handballs: 5, Marks: 4, Tackles: 2, Hitouts: 0}
	pm, err := cmds.CalculateFantasyScore(ctx, 1, stats)
	if err != nil {
		t.Fatalf("CalculateFantasyScore: %v", err)
	}

	wantScore := 3 * domain.GoalsMultiplier // 15
	if pm.Score != wantScore {
		t.Errorf("player score = %d, want %d", pm.Score, wantScore)
	}

	// Club match stored score should be updated.
	cm := cmRepo.matches[1]
	if cm.StoredScore != wantScore {
		t.Errorf("club match stored score = %d, want %d", cm.StoredScore, wantScore)
	}
}

func TestCalculateFantasyScore_Star(t *testing.T) {
	cmds, _, _, pmRepo, cmRepo := setupCommands()
	ctx := context.Background()

	cmRepo.matches[1] = domain.ClubMatch{ID: 1, MatchID: 1, ClubSeasonID: 1}

	pmRepo.matches[1] = domain.PlayerMatch{
		ID:             1,
		ClubMatchID:    1,
		PlayerSeasonID: 1,
		Position:       domain.PositionPtr(domain.PositionStar),
		Status:         domain.PlayerMatchStatusPtr(domain.PlayerMatchStatusPlayed),
		Score:          0,
	}

	stats := domain.AFLStats{Goals: 2, Kicks: 10, Handballs: 5, Marks: 4, Tackles: 3, Hitouts: 0}
	pm, err := cmds.CalculateFantasyScore(ctx, 1, stats)
	if err != nil {
		t.Fatalf("CalculateFantasyScore: %v", err)
	}

	// Star: 2*5 + 10*1 + 5*1 + 4*2 + 3*4 = 10+10+5+8+12 = 45
	wantScore := 45
	if pm.Score != wantScore {
		t.Errorf("star score = %d, want %d", pm.Score, wantScore)
	}
}

func TestCalculateFantasyScore_RecalculatesClubTotal(t *testing.T) {
	cmds, _, _, pmRepo, cmRepo := setupCommands()
	ctx := context.Background()

	cmRepo.matches[1] = domain.ClubMatch{ID: 1, MatchID: 1, ClubSeasonID: 1}

	// Two starters in the same club match.
	pmRepo.matches[1] = domain.PlayerMatch{
		ID: 1, ClubMatchID: 1, PlayerSeasonID: 1,
		Position: domain.PositionPtr(domain.PositionGoals), Status: domain.PlayerMatchStatusPtr(domain.PlayerMatchStatusPlayed), Score: 10,
	}
	pmRepo.matches[2] = domain.PlayerMatch{
		ID: 2, ClubMatchID: 1, PlayerSeasonID: 2,
		Position: domain.PositionPtr(domain.PositionKicks), Status: domain.PlayerMatchStatusPtr(domain.PlayerMatchStatusPlayed), Score: 0,
	}

	// Calculate score for player 2 (kicks position, 1 pt/kick).
	stats := domain.AFLStats{Kicks: 20}
	_, err := cmds.CalculateFantasyScore(ctx, 2, stats)
	if err != nil {
		t.Fatalf("CalculateFantasyScore: %v", err)
	}

	// Club total should be player1(10) + player2(20) = 30.
	cm := cmRepo.matches[1]
	if cm.StoredScore != 30 {
		t.Errorf("club match total = %d, want 30", cm.StoredScore)
	}
}
