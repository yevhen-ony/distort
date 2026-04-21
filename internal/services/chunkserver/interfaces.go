package chunkserver

import (
	"io"
	"time"

	"dos/internal/libraries/digest"
)

type ChunkWriter interface {
	io.WriteCloser
	Digest() digest.Digest
	Commit(ChunkID) (time.Time, error)
}

type ChunkStorage interface {
	Get(chunkID ChunkID) (io.ReadCloser, error)
	GetMeta(chunkID ChunkID) (*ChunkMeta, error)
	NewWriter() (ChunkWriter, error)
	GetAllIDs() ([]ChunkID, error)
}
