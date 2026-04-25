package client

import (
	t "dos/internal/common/types"
)

type Chunk struct {
	ID       t.ChunkID
	Checksum t.Checksum 
	Data     []byte
}
