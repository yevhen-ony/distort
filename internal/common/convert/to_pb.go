package convert

import (
	t "dos/internal/common/types"
	mpb "dos/gen/proto/master/v1"
)

func NodeAccessToPB(mNodes []t.NodeAccess) []*mpb.NodeAccess {
	pbNodes	:= make([]*mpb.NodeAccess, 0, len(mNodes))
	for _, mNode := range mNodes {
		pbNodes = append(pbNodes, &mpb.NodeAccess{
			NodeId: string(mNode.NodeID),
			Address: mNode.Addr,
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
			Nodes: NodeAccessToPB(mChunk.Nodes),
		})
	}
	return pbChunks 
}
