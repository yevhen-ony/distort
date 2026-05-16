package storage

import "errors"

var (
	ErrInvalidHeader  = errors.New("invalid header")
	ErrInvalidData    = errors.New("invalid data")
	ErrInvalidRequest = errors.New("invalid request")

	ErrChunkNotFound = errors.New("not found")
	ErrChunkConflict = errors.New("chunk conflict")
	ErrInvalidDigest = errors.New("invalid digest")

	ErrInvalidNodeID     = errors.New("invalid node id")
	ErrNodeNotRegistered = errors.New("node is not registered")

	ErrNoValidTargets = errors.New("no valid targets provided")
)
