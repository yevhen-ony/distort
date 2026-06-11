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
	ID   NodeID `json:"node_id"`
	Addr string `json:"address"`
}

type MasterRef struct {
	ID   MasterID `json:"master_id"`
	Addr string   `json:"address"`
}

type ObjectInfo struct {
	ID          ObjectID `json:"object_id"`
	ChunkCount  int      `json:"chunk_count"`
	Replication int      `json:"replication"`
}

type NodeStats struct {
	FreeBytes  int64 `json:"free_bytes"`
	UsedBytes  int64 `json:"used_bytes"`
	ChunkCount int   `json:"chunk_count"`
}

type ChunkMeta struct {
	ID     ChunkID       `json:"chunk_id"`
	Digest digest.Digest `json:"digest"`
}

type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

type ChunkInfo struct {
	ID           ChunkID  `json:"chunk_id"`
	ReplicaCount int      `json:"replica_count"`
	Size         int64    `json:"size"`
	ObjectID     ObjectID `json:"object_id"`
}

type NodeInfo struct {
	ID         NodeID `json:"node_id"`
	Addr       string `json:"address"`
	ChunkCount int    `json:"chunk_count"`
	UsedBytes  int64  `json:"used_bytes"`
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

type ChunkStorageView struct {
	Meta  ChunkMeta `json:"meta"`
	State string    `json:"state"`
}

type HeartbeatView struct {
	Status string `json:"status"`
}
