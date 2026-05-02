package repository

import "errors"

// Sentinel errors returned by repository methods.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrLimitExceeded = errors.New("limit exceeded")
)
