package usecase

import "errors"

var (
	ErrNotFound   = errors.New("resource not found")
	ErrForbidden  = errors.New("forbidden")
	ErrBadRequest = errors.New("bad request")
)
