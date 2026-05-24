package types

import (
	"dos/internal/common/digest"
	"errors"
	"fmt"
)

func (meta *ChunkMeta) Clone() *ChunkMeta {
	return &ChunkMeta{
		ID:     meta.ID,
		Digest: meta.Digest.Clone(),
	}
}

func (m ChunkMeta) Match(other ChunkMeta) error {
	if other.ID != m.ID {
		return fmt.Errorf("id mismatch: %w", ErrChunkMetaMismatch)
	}

	if err := m.Digest.Match(other.Digest); err != nil {
		return errors.Join(err, ErrChunkMetaMismatch) 
	}

	return nil
}

func NewChunk(id ChunkID, data []byte) Chunk {
	dg := digest.New()
	dg.Write(data)

	return Chunk {
		Meta: ChunkMeta{ID: id, Digest: dg.Digest()},
		Data: data,
	}
}

func (c Chunk) Validate() error {
	got := NewChunk(c.Meta.ID, c.Data)
	return c.Meta.Match(got.Meta)
}

func NewReplicaChainFailed(chunkID ChunkID, targets []NodeRef) *ReplicaChainFailedReport {
	return &ReplicaChainFailedReport{
		ChunkID: chunkID,
		Targets: targets,
	}
}

func (rcf *ReplicaChainFailedReport) ToRecord() StorageNodeReport {
	return StorageNodeReport{ReplicaChainFailed: rcf}
}

func NewReplicaStaged(chunk ChunkMeta) *ReplicaStagedReport {
	return &ReplicaStagedReport{Chunk: chunk}
}

func (rs *ReplicaStagedReport) ToRecord() StorageNodeReport {
	return StorageNodeReport{ReplicaStaged: rs}
}

func NewReplicaDeleted(chunkID ChunkID) *ReplicaDeletedReport {
	return &ReplicaDeletedReport{ChunkID: chunkID}
}

func (rd *ReplicaDeletedReport) ToRecord() StorageNodeReport {
	return StorageNodeReport{ReplicaDeleted: rd}
}
