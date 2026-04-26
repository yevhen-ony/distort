package convert 

import (
	pb "dos/gen/proto/common/v1"
	mpb "dos/gen/proto/master/v1"

	t "dos/internal/common/types"
)

type ChunkPlacementLike interface {
	GetChunkId() string
	GetChunkKey() string
	GetNodes() []*pb.NodeRef
}

func NodeRefFromPB(pbNode *pb.NodeRef) *t.NodeRef {
	return &t.NodeRef{
		NodeID: t.NodeID(pbNode.GetNodeId()),
		Addr: pbNode.GetAddr(),
	}
}

func ChunkPlacementFromPB(pbObj ChunkPlacementLike) *t.ChunkPlacement {
	pbNodes := pbObj.GetNodes()
	nodes := make([]t.NodeRef, 0, len(pbNodes))
	for _, pbNode := range pbNodes {
		nodes = append(nodes, *NodeRefFromPB(pbNode))
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

type NodeReportLike interface {
	GetAddr() string
	GetFreeBytes() int64
	GetUsedBytes() int64
	GetChunkCount() int32
}

func NodeReportFromPB(pbObj NodeReportLike) *t.NodeReport {
	return &t.NodeReport{
		Addr: pbObj.GetAddr(),
		FreeBytes: pbObj.GetFreeBytes(),
		UsedBytes: pbObj.GetUsedBytes(),
		ChunkCount: int(pbObj.GetChunkCount()),
	}
}

