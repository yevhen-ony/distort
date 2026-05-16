package master

import "errors"

var (
	ErrArgNil          = errors.New("argument is nil")
	ErrInvalidArgument = errors.New("invalid argument")

	ErrObjectExists   = errors.New("object already exists")
	ErrObjectNotFound = errors.New("object not found")
	ErrObjectNotEmpty = errors.New("object still contains chunks")

	ErrChunkNotAvailable     = errors.New("chunk is not available")
	ErrChunkKeyExists        = errors.New("chunk key already exists")
	ErrChunkKeyNotFound      = errors.New("chunk key not found")
	ErrChunkExists           = errors.New("chunk already exists")
	ErrChunkNotFound         = errors.New("chunk not found")
	ErrWrongChunkID          = errors.New("chunk id does not match")
	ErrChunkDigestNotSet     = errors.New("chunk digest not set")
	ErrChunkDigestConflict   = errors.New("chunk digest conflict")
	ErrChunkReplicaUnderflow = errors.New("chunk replication count is already zero")

	ErrNodeAddrInUse = errors.New("node address already registered")
	ErrNodeNotFound  = errors.New("node not found")

	ErrNoCandidateNodes = errors.New("no suitable nodes found")
)
