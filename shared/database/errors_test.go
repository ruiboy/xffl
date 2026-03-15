package database

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapPgError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "nil error returns nil",
			err:  nil,
			want: nil,
		},
		{
			name: "no rows maps to ErrNotFound",
			err:  pgx.ErrNoRows,
			want: ErrNotFound,
		},
		{
			name: "unique violation maps to ErrConflict",
			err:  &pgconn.PgError{Code: "23505"},
			want: ErrConflict,
		},
		{
			name: "foreign key violation maps to ErrInvalidRef",
			err:  &pgconn.PgError{Code: "23503"},
			want: ErrInvalidRef,
		},
		{
			name: "wrapped no rows maps to ErrNotFound",
			err:  fmt.Errorf("query failed: %w", pgx.ErrNoRows),
			want: ErrNotFound,
		},
		{
			name: "wrapped unique violation maps to ErrConflict",
			err:  fmt.Errorf("insert failed: %w", &pgconn.PgError{Code: "23505"}),
			want: ErrConflict,
		},
		{
			name: "unknown error passes through",
			err:  errors.New("something else"),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapPgError(tt.err)

			if tt.want == nil {
				if got != tt.err {
					t.Errorf("MapPgError() = %v, want original error %v", got, tt.err)
				}
				return
			}

			if !errors.Is(got, tt.want) {
				t.Errorf("MapPgError() = %v, want error wrapping %v", got, tt.want)
			}

			// mapped errors should still wrap the original
			if tt.err != nil && tt.want != nil {
				var pgErr *pgconn.PgError
				if errors.As(tt.err, &pgErr) {
					if !errors.As(got, &pgErr) {
						t.Error("mapped error should preserve original pgconn.PgError")
					}
				}
			}
		})
	}
}
