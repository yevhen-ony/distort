package types

type ObjectID string
type ChunkID string
type ChunkKey string
type NodeID string

type Checksum string

type NodeAccess struct {
	NodeID NodeID
	Addr   string
}

type ChunkPlacement struct {
	ChunkID  ChunkID
	ChunkKey ChunkKey
	Nodes    []NodeAccess
}

type ObjectAccess struct {
	ObjectID  ObjectID
	TotalSize int64
	Chunks    []ChunkPlacement
}

type NodeReport struct {
	Addr       string
	FreeBytes  int64
	UsedBytes  int64
	ChunkCount int
}
