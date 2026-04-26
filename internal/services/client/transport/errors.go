package transport

import "errors"
var (
	ErrInputInvalid = errors.New("invalid input")
	ErrChunkInvalid = errors.New("chunk failed validation")

	ErrHeaderInvalid = errors.New("invalid header")
	ErrDataInvalid = errors.New("invalid data")

	ErrChunkDescMismatch = errors.New("chunk description mismatch")
)
