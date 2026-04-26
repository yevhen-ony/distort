package client

import (
	t "dos/internal/common/types"
	"dos/internal/common/digest"
)

type Chunk struct {
	ID       t.ChunkID
	Checksum digest.Checksum 
	Data     []byte
}
