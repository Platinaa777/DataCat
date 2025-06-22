package entity

import "errors"

var (
	ErrUpstream  = errors.New("upstream unavailable")
	ErrNotFound  = errors.New("resource not found")
	ErrConflict  = errors.New("resource already exists or was modified")
	ErrForbidden = errors.New("access denied")
	ErrInvalid   = errors.New("invalid input")
)
