package storage

import "errors"

var (
	ErrHeaderInvalid = errors.New("invalid header")
	ErrDataInvalid = errors.New("invalid data")

	ErrChunkNotFound = errors.New("not found")
	ErrChunkConflict = errors.New("chunk conflict")
	ErrDigestInvalid  = errors.New("digest invalid")
)
