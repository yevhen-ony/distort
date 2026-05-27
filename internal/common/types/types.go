package types

import (
	"dos/internal/common/digest"
)
type MasterID string
type ObjectID string
type ChunkID string
type ChunkKey string
type NodeID string

type NodeRef struct {
	ID   NodeID
	Addr string
}

type MasterRef struct {
	ID MasterID 
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

type ObjectInfo struct {
	ID          ObjectID
	ChunkCount  int
	Replication int
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
	Digest digest.Digest
}

type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

type ChunkInfo struct {
	ID           ChunkID
	ReplicaCount int
	Size         int64
	ObjectID     ObjectID
}

type NodeInfo struct {
	ID         NodeID
	Addr       string
	ChunkCount int
	UsedBytes  int64
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

type ReplicaDeletedReport struct {
	ChunkID ChunkID
}

type StorageNodeReport struct {
	ReplicaStaged      *ReplicaStagedReport
	ReplicaChainFailed *ReplicaChainFailedReport
	ReplicaDeleted     *ReplicaDeletedReport
}


