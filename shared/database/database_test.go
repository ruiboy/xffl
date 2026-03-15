package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeDBTX is a minimal DBTX for testing.
type fakeDBTX struct {
	tag string
}

func (f *fakeDBTX) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag(""), nil
}
func (f *fakeDBTX) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}
func (f *fakeDBTX) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

// fakeQueries tracks which DBTX it was created from.
type fakeQueries struct {
	source DBTX
}

func TestDB_Queries(t *testing.T) {
	pool := &fakeDBTX{tag: "pool"}
	db := &DB[fakeQueries]{
		dbtx: pool,
		newQ: func(d DBTX) *fakeQueries {
			return &fakeQueries{source: d}
		},
	}

	q := db.Queries()
	if q.source.(*fakeDBTX).tag != "pool" {
		t.Error("Queries() should create queries from the pool")
	}
}

func TestWithTx_Commit(t *testing.T) {
	var committed bool
	tx := &fakeTx{
		commitFn:   func() { committed = true },
		rollbackFn: func() {},
	}

	db := &DB[fakeQueries]{
		dbtx: &fakeDBTX{},
		newQ: func(d DBTX) *fakeQueries { return &fakeQueries{source: d} },
		begin: func(context.Context) (Tx, error) {
			return tx, nil
		},
	}

	var called bool
	err := db.WithTx(context.Background(), func(q *fakeQueries) error {
		called = true
		// Verify the queries got the tx, not the pool
		if q.source != DBTX(tx) {
			t.Error("callback should receive queries bound to the transaction")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("WithTx() error = %v", err)
	}
	if !called {
		t.Error("callback was not called")
	}
	if !committed {
		t.Error("transaction was not committed")
	}
}

func TestWithTx_Rollback(t *testing.T) {
	var rolledBack bool
	tx := &fakeTx{
		commitFn:   func() {},
		rollbackFn: func() { rolledBack = true },
	}

	db := &DB[fakeQueries]{
		dbtx: &fakeDBTX{},
		newQ: func(d DBTX) *fakeQueries { return &fakeQueries{} },
		begin: func(context.Context) (Tx, error) {
			return tx, nil
		},
	}

	wantErr := errors.New("something failed")
	err := db.WithTx(context.Background(), func(q *fakeQueries) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx() error = %v, want %v", err, wantErr)
	}
	if !rolledBack {
		t.Error("transaction was not rolled back")
	}
}

func TestWithTx_BeginError(t *testing.T) {
	wantErr := errors.New("connection failed")

	db := &DB[fakeQueries]{
		dbtx: &fakeDBTX{},
		newQ: func(d DBTX) *fakeQueries { return &fakeQueries{} },
		begin: func(context.Context) (Tx, error) {
			return nil, wantErr
		},
	}

	err := db.WithTx(context.Background(), func(q *fakeQueries) error {
		t.Fatal("callback should not be called when Begin fails")
		return nil
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx() error = %v, want %v", err, wantErr)
	}
}

// --- test doubles ---

type fakeTx struct {
	fakeDBTX
	commitFn   func()
	rollbackFn func()
}

func (f *fakeTx) Commit(_ context.Context) error {
	f.commitFn()
	return nil
}

func (f *fakeTx) Rollback(_ context.Context) error {
	f.rollbackFn()
	return nil
}
