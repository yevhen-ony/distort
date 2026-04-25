package client

type ObjectID string
type ChunkID string
type ChunkKey int64

type Chunk struct {
	ID       string
	Checksum string
	Data     []byte
}

type NodeAccess struct {
	ID   string
	Addr string
}
