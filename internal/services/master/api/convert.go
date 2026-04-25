package api

import (
	m "dos/internal/services/master"
	pb "dos/gen/proto/master/v1"
)

func toPBNodeAccess(mNodes []m.NodeAccess) []*pb.NodeAccess {
	pbNodes	:= make([]*pb.NodeAccess, 0, len(mNodes))
	for _, mNode := range mNodes {
		pbNodes = append(pbNodes, &pb.NodeAccess{
			NodeId: string(mNode.NodeID),
			Address: mNode.Addr,
		})
	}
	return pbNodes
}

func toPBChunkPlacement(mChunks []m.ChunkPlacement) []*pb.ChunkPlacement {
	pbChunks := make([]*pb.ChunkPlacement, 0, len(mChunks))
	for _, mChunk := range mChunks{
		pbChunks = append(pbChunks, &pb.ChunkPlacement{
			ChunkId: string(mChunk.ChunkID),
			ChunkKey: string(mChunk.ChunkKey),
			Nodes: toPBNodeAccess(mChunk.Nodes),
		})
	}
	return pbChunks 
}
