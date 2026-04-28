package storage

import "errors"

var (
	ErrInvalidHeader = errors.New("invalid header")
	ErrInvalidData = errors.New("invalid data")

	ErrChunkNotFound = errors.New("not found")
	ErrChunkConflict = errors.New("chunk conflict")
	ErrInvalidDigest  = errors.New("invalid digest")


	ErrInvalidNodeID = errors.New("invalid node id")
)
