package storage

import (
	"fmt"
	"io"

	t "dos/internal/common/types"
)

func (cs *StorageService) LoadChunk(chunkID t.ChunkID) (t.Chunk, error) {

	rec, err := cs.inventory.GetRecord(chunkID)
	if err != nil {
		return t.Chunk{}, err
	}
	reader, err := cs.storageBE.Get(chunkID)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("get from store: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("read chunk: %w", err)
	}
	chunk := t.Chunk{
		Meta: rec.Meta,
		Data: data,
	}
	return chunk, nil
}
