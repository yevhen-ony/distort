package client

import "errors"

var (
	ErrObjectSizeMismatch = errors.New("object size mismatch")
	ErrChunkSizeMismatch  = errors.New("chunk size mismatch")
	ErrChunkNotFound      = errors.New("chunk not found")
	ErrInvalidChunkSize   = errors.New("chunk size must be greater than zero")
)
