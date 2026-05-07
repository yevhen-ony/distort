package chunkrpc

import "errors"

var (
	ErrInvalidHeader = errors.New("missing or invalid header")
	ErrInvalidData = errors.New("missing or invalid data")
	
)
	
