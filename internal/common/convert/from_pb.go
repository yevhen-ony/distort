package convert 

import (
	mpb "dos/gen/proto/master/v1"
	t "dos/internal/common/types"
)

type ChunkPlacementLike interface {
	GetChunkId() string
	GetChunkKey() string
	GetNodes() []*mpb.NodeAccess
}

func NodeAccessFromPB(pbNode *mpb.NodeAccess) *t.NodeAccess {
	return &t.NodeAccess{
		NodeID: t.NodeID(pbNode.GetNodeId()),
		Addr: pbNode.GetAddress(),
	}
}

func ChunkPlacementFromPB(pbObj ChunkPlacementLike) *t.ChunkPlacement {
	pbNodes := pbObj.GetNodes()
	nodes := make([]t.NodeAccess, 0, len(pbNodes))
	for _, pbNode := range pbNodes {
		nodes = append(nodes, *NodeAccessFromPB(pbNode))
	}
	return &t.ChunkPlacement{
		ChunkID: t.ChunkID(pbObj.GetChunkId()),
		ChunkKey: t.ChunkKey(pbObj.GetChunkKey()),
		Nodes: nodes,
	}
}

type ObjectAccessLike interface {
	GetObjectId() string
	GetTotalSize() int64
	GetChunks() []*mpb.ChunkPlacement
}

func ObjectAccessFromPB(pbObj ObjectAccessLike) *t.ObjectAccess {
	pbChunks := pbObj.GetChunks()
	chunks := make([]t.ChunkPlacement, 0, len(pbChunks))
	for _, pbChunk := range pbChunks {
		chunks = append(chunks, *ChunkPlacementFromPB(pbChunk))
	}
	return &t.ObjectAccess{
		ObjectID: t.ObjectID(pbObj.GetObjectId()),
		TotalSize: pbObj.GetTotalSize(),
		Chunks: chunks,
	}
}

