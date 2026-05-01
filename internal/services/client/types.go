package client

import (
	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type Chunk struct {
	Meta t.ChunkMeta
	Data []byte
}

func NewChunk(id t.ChunkID, data []byte) Chunk {
	dg := digest.New()
	dg.Write(data)
	return Chunk{
		Meta: t.ChunkMeta{
			ID:     id,
			Digest: dg.Digest(),
		},
		Data: data,
	}
}
