package repository

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidReference = errors.New("invalid reference")
	ErrInvalidData      = errors.New("invalid data")
	ErrConflict         = errors.New("conflict")
)
