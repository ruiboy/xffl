package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is the interface satisfied by both *pgxpool.Pool and pgx.Tx.
// sqlc generates this same interface for its Queries constructor.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Tx extends DBTX with commit/rollback.
type Tx interface {
	DBTX
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// beginFunc abstracts transaction creation for testing.
type beginFunc func(ctx context.Context) (Tx, error)

// DB wraps a connection pool and a sqlc Queries constructor.
// Q is the sqlc-generated Queries type for a given service.
type DB[Q any] struct {
	dbtx  DBTX
	newQ  func(DBTX) *Q
	begin beginFunc
}

// New creates a DB that uses pool for queries and transactions.
// newQ is the sqlc-generated New function (e.g., sqlcgen.New).
func New[Q any](pool *pgxpool.Pool, newQ func(DBTX) *Q) *DB[Q] {
	return &DB[Q]{
		dbtx: pool,
		newQ: newQ,
		begin: func(ctx context.Context) (Tx, error) {
			return pool.Begin(ctx)
		},
	}
}

// Connect creates a connection pool and returns a DB.
func Connect[Q any](ctx context.Context, connString string, newQ func(DBTX) *Q) (*DB[Q], error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("database connect: %w", err)
	}
	return New(pool, newQ), nil
}

// Queries returns a Queries instance for the read path (no transaction).
func (db *DB[Q]) Queries() *Q {
	return db.newQ(db.dbtx)
}

// WithTx runs fn inside a transaction. It commits on success and rolls back
// on error or panic. The Queries instance passed to fn is bound to the transaction.
func (db *DB[Q]) WithTx(ctx context.Context, fn func(*Q) error) error {
	tx, err := db.begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(db.newQ(tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// Pool returns the underlying DBTX (typically *pgxpool.Pool).
func (db *DB[Q]) Pool() DBTX {
	return db.dbtx
}
