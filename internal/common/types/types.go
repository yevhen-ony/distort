package types

import (
	"dos/internal/common/digest"
)

type ObjectID string
type ChunkID string
type ChunkKey string
type NodeID string

type NodeRef struct {
	ID   NodeID
	Addr string
}

type ChunkDesc struct {
	ChunkID   ChunkID
	ChunkKey  ChunkKey
	ChunkSize int64
}

type ChunkPlacement struct {
	ChunkDesc
	Nodes []NodeRef
}

type ObjectDesc struct {
	ID        ObjectID
	TotalSize int64
}

type ObjectItem struct {
	ID         ObjectID
	ChunkCount int
}

type ObjectAccess struct {
	ObjectDesc
	Chunks []ChunkPlacement
}

type NodeStats struct {
	FreeBytes  int64
	UsedBytes  int64
	ChunkCount int
}

type ChunkMeta struct {
	ID     ChunkID
	Digest *digest.Digest
}

func (meta *ChunkMeta) Clone() *ChunkMeta {
	return &ChunkMeta{
		ID:     meta.ID,
		Digest: meta.Digest.Clone(),
	}
}

type Chunk struct {
	Meta ChunkMeta
	Data []byte 
}

type ReportResult struct {
	Accepted []ChunkID
	Rejected []ChunkID
}

type ReplicaStagedReport struct {
	Chunk ChunkMeta
}

type ReplicaChainFailedReport struct {
	ChunkID ChunkID
	Targets []NodeRef
}

type ReplicaReport struct {
	ReplicaStaged      *ReplicaStagedReport
	ReplicaChainFailed *ReplicaChainFailedReport
}



