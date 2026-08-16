package domain

import "errors"

// ErrNotFound signals that a lookup found no matching record. Repositories
// translate their driver's empty-result error into this so callers can tell
// absence apart from a genuine failure.
var ErrNotFound = errors.New("not found")
