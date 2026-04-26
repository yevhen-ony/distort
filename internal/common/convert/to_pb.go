package convert

import (
	t "dos/internal/common/types"
	pb "dos/gen/proto/common/v1"
	mpb "dos/gen/proto/master/v1"
)

func NodeRefToPB(mNodes []t.NodeRef) []*pb.NodeRef {
	pbNodes	:= make([]*pb.NodeRef, 0, len(mNodes))
	for _, mNode := range mNodes {
		pbNodes = append(pbNodes, &pb.NodeRef{
			NodeId: string(mNode.NodeID),
			Addr: mNode.Addr,
		})
	}
	return pbNodes
}

func ChunkPlacementToPB(mChunks []t.ChunkPlacement) []*mpb.ChunkPlacement {
	pbChunks := make([]*mpb.ChunkPlacement, 0, len(mChunks))
	for _, mChunk := range mChunks{
		pbChunks = append(pbChunks, &mpb.ChunkPlacement{
			ChunkId: string(mChunk.ChunkID),
			ChunkKey: string(mChunk.ChunkKey),
			Nodes: NodeRefToPB(mChunk.Nodes),
		})
	}
	return pbChunks 
}
