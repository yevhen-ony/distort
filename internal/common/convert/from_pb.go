package convert

import (
	pb "dos/gen/proto/common/v1"
	mpb "dos/gen/proto/master/v1"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type ChunkPlacementLike interface {
	GetChunkId() string
	GetChunkKey() string
	GetNodes() []*pb.NodeRef
}

func NodeRefFromPB(pbNode *pb.NodeRef) *t.NodeRef {

	return &t.NodeRef{
		ID: t.NodeID(pbNode.GetNodeId()),
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
		ID: t.ChunkID(pbObj.GetChunkId()),
		Key: t.ChunkKey(pbObj.GetChunkKey()),
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
		ID: t.ObjectID(pbObj.GetObjectId()),
		TotalSize: pbObj.GetTotalSize(),
		Chunks: chunks,
	}
}

type NodeStatsLike interface {
	GetFreeBytes() int64
	GetUsedBytes() int64
	GetChunkCount() int32
}

func NodeStatsFromPB(pbObj NodeStatsLike) *t.NodeStats {

	return &t.NodeStats{
		FreeBytes: pbObj.GetFreeBytes(),
		UsedBytes: pbObj.GetUsedBytes(),
		ChunkCount: int(pbObj.GetChunkCount()),
	}
}

type DigestLike interface {
	GetChecksum() string
	GetSize() int64
}

func DigestFromPB(pbObj DigestLike) digest.Digest {

	return digest.Digest{
		Checksum: digest.Checksum(pbObj.GetChecksum()),
		Size: pbObj.GetSize(),
	}
}

type ChunkDescLike interface {
	GetDigest() *pb.Digest
	GetChunkId() string
}

func ChunkDescFromPB(pbObj ChunkDescLike) t.ChunkDesc {

	return t.ChunkDesc{
		ID: t.ChunkID(pbObj.GetChunkId()),
		Digest: DigestFromPB(pbObj.GetDigest()),
	}
}

type ChunkStorageRejectLike interface {
	GetChunkId() string
	GetReason() string
}

func ChunkStorageRejectFromPB(pbObj ChunkStorageRejectLike) t.ChunkStorageReject{

	return t.ChunkStorageReject{
		ChunkID: t.ChunkID(pbObj.GetChunkId()),
		Reason: pbObj.GetReason(),
	}
}



