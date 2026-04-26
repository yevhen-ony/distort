package client

import (
	t "dos/internal/common/types"
)

type Chunk struct {
	t.ChunkDesc
	Data     []byte
}
