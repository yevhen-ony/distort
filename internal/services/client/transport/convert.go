package transport

import (
	mpb "dos/gen/proto/master/v1"
	c "dos/internal/services/client"
)


type ChunkPlacementLike interface {
	GetChunkId() string
	GetChunkKey() string
	GetNodes() []*mpb.NodeAccess
}

func NodeAccessFromPB(pbNode *mpb.NodeAccess) *c.NodeAccess {
	return &c.NodeAccess{
		NodeID: pbNode.GetNodeId(),
		Addr: pbNode.GetAddress(),
	}
}

func ChunkPlacementFromPB(pbObj ChunkPlacementLike) *c.ChunkPlacement {
	pbNodes := pbObj.GetNodes()
	nodes := make([]c.NodeAccess, 0, len(pbNodes))
	for _, pbNode := range pbNodes {
		nodes = append(nodes, *NodeAccessFromPB(pbNode))
	}
	return &c.ChunkPlacement{
		ChunkID: c.ChunkID(pbObj.GetChunkId()),
		ChunkKey: c.ChunkKey(pbObj.GetChunkKey()),
		Nodes: nodes,
	}
}

type ObjectAccessLike interface {
	GetObjectId() string
	GetTotalSize() int64
	GetChunks() []*mpb.ChunkPlacement
}

func ObjectAccessFromPB(pbObj ObjectAccessLike) *c.ObjectAccess {
	pbChunks := pbObj.GetChunks()
	chunks := make([]c.ChunkPlacement, 0, len(pbChunks))
	for _, pbChunk := range pbChunks {
		chunks = append(chunks, *ChunkPlacementFromPB(pbChunk))
	}
	return &c.ObjectAccess{
		ObjectID: c.ObjectID(pbObj.GetObjectId()),
		TotalSize: pbObj.GetTotalSize(),
		Chunks: chunks,
	}
}

