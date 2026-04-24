package master

import "errors"

var (
	ErrArgNil = errors.New("argument is nil")

	ErrObjectExists   = errors.New("object already exists")
	ErrObjectNotFound = errors.New("object not found")

	ErrChunkKeyExists      = errors.New("chunk key already exists")
	ErrChunkExists         = errors.New("chunk already exists")
	ErrChunkNotFound       = errors.New("chunk not found")
	ErrChunkDigestNotSet   = errors.New("chunk digest not set")
	ErrChunkDigestConflict = errors.New("chunk digest conflict")

	ErrNodeAddrInUse = errors.New("node address already registered")
	ErrNodeNotFound = errors.New("node not found")
	
	ErrNoCandidateNodes = errors.New("no suitable nodes found")
)
