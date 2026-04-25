package client

type ObjectID string
type ChunkID string
type ChunkKey string

type Chunk struct {
	ID       string
	Checksum string
	Data     []byte
}

type NodeAccess struct {
	NodeID   string
	Addr string
}

type ChunkPlacement struct {
	ChunkID ChunkID
	ChunkKey ChunkKey
	Nodes []NodeAccess
}

type ObjectAccess struct {
	ObjectID ObjectID
	TotalSize int64
	Chunks []ChunkPlacement
}
