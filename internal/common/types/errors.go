package types

import "errors"

var (
	ErrChunkMetaMismatch = errors.New("chunk meta mismatch")
	ErrInvalidMasterRef = errors.New("invalid master ref")
)
