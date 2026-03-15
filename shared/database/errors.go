package database

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
	ErrInvalidRef = errors.New("invalid reference")
)

// MapPgError translates pgx/pgconn errors into domain-level sentinel errors.
// Unknown errors are returned unchanged.
func MapPgError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fmt.Errorf("%w: %w", ErrConflict, err)
		case "23503":
			return fmt.Errorf("%w: %w", ErrInvalidRef, err)
		}
	}

	return err
}
