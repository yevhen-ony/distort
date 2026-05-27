package convert

import (
	pb "dos/gen/proto/common/v1"
	mpb "dos/gen/proto/master/v1"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type ChunkPlacementLike interface {
	GetChunkId() string
	GetChunkKey() string
	GetChunkSize() int64
	GetNodes() []*pb.NodeRef
}

func NodeRefFromPB(pbNode *pb.NodeRef) t.NodeRef {
	return t.NodeRef{
		ID:   t.NodeID(pbNode.GetNodeId()),
		Addr: pbNode.GetAddr(),
	}
}

func ObjectSlotFromPB(pbSlot *pb.ObjectSlot) t.ObjectSlot {
	return t.ObjectSlot {
		ObjectID: t.ObjectID(pbSlot.GetObjectId()),
		ChunkKey: t.ChunkKey(pbSlot.GetChunkKey()),
	}
}

func ChunkPlacement1FromPB(pbP *mpb.ChunkPlacement1) t.ChunkPlacement1 {
	return t.ChunkPlacement1{
		Meta: ChunkMetaFromPB(pbP.GetChunkMeta()),
		Slot: ObjectSlotFromPB(pbP.GetObjectSlot()),
		Sources: utils.Map(pbP.GetSources(), NodeRefFromPB),
	}
}

func ChunkDesc1FromPB(pbD *mpb.ChunkDesc1) t.ChunkDesc1 {
	return t.ChunkDesc1 {
		Placement: ChunkPlacement1FromPB(pbD.GetPlacement()),
	}
}

func ObjectDesc1FromPB(pbD *mpb.ObjectDesc1) t.ObjectDesc1 {
	return t.ObjectDesc1{
		ID: t.ObjectID(pbD.GetObjectId()), 
		Size: pbD.GetSize(),
		Replication: int(pbD.GetReplication()),
		Chunks: utils.Map(pbD.GetChunks(), ChunkPlacement1FromPB),
	}
}

func ChunkPlacementFromPB(pbObj ChunkPlacementLike) *t.ChunkPlacement {
	pbNodes := pbObj.GetNodes()
	nodes := make([]t.NodeRef, 0, len(pbNodes))
	for _, pbNode := range pbNodes {
		nodes = append(nodes, NodeRefFromPB(pbNode))
	}
	return &t.ChunkPlacement{
		ChunkDesc: t.ChunkDesc{
			ChunkID:   t.ChunkID(pbObj.GetChunkId()),
			ChunkKey:  t.ChunkKey(pbObj.GetChunkKey()),
			ChunkSize: pbObj.GetChunkSize(),
		},
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
		ObjectDesc: t.ObjectDesc{
			ID:        t.ObjectID(pbObj.GetObjectId()),
			TotalSize: pbObj.GetTotalSize(),
		},
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
		FreeBytes:  pbObj.GetFreeBytes(),
		UsedBytes:  pbObj.GetUsedBytes(),
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
		Size:     pbObj.GetSize(),
	}
}

type ChunkDescLike interface {
	GetDigest() *pb.Digest
	GetChunkId() string
}

func ChunkMetaFromPB(pbObj ChunkDescLike) t.ChunkMeta {
	digest := DigestFromPB(pbObj.GetDigest())
	return t.ChunkMeta{
		ID:     t.ChunkID(pbObj.GetChunkId()),
		Digest: digest,
	}
}

type ChunkStorageRejectLike interface {
	GetChunkId() string
	GetReason() string
}


func ObjectInfoFromPB(pbInfo *mpb.ObjectInfo) t.ObjectInfo {
	return t.ObjectInfo{
		ID:         t.ObjectID(pbInfo.GetObjectId()),
		ChunkCount: int(pbInfo.GetChunkCount()),
		Replication: int(pbInfo.GetReplication()),
	}
}

func ChunkInfoFromPB(pbInfo *mpb.ChunkInfo) t.ChunkInfo {
	return t.ChunkInfo{
		ID:           t.ChunkID(pbInfo.GetChunkId()),
		Size:         pbInfo.GetChunkSize(),
		ReplicaCount: int(pbInfo.GetReplicaCount()),
		ObjectID:     t.ObjectID(pbInfo.GetObjectId()),
	}
}

func NodeInfoFromPB(pbInfo *mpb.NodeInfo) t.NodeInfo {
	return t.NodeInfo{
		ID: t.NodeID(pbInfo.GetNodeId()),
		Addr: pbInfo.GetAddr(),
		ChunkCount: int(pbInfo.GetChunkCount()),
		UsedBytes: pbInfo.GetUsedBytes(),
	}
}

func ReplicaStagedReportFromPB(pb *mpb.ReplicaStaged) *t.ReplicaStagedReport {
	if pb == nil {
		return nil
	}

	return &t.ReplicaStagedReport{
		Chunk: ChunkMetaFromPB(pb.GetChunk()),
	}
}

func ReplicaChainFailedReportFromPB(pb *mpb.ReplicaChainFailed) *t.ReplicaChainFailedReport {
	if pb == nil {
		return nil
	}

	return &t.ReplicaChainFailedReport{
		ChunkID: t.ChunkID(pb.GetChunkId()),
		Targets: utils.Map(pb.GetTargets(), NodeRefFromPB),
	}
}

func ReplicaDeletedReportFromPB(pb *mpb.ReplicaDeleted) *t.ReplicaDeletedReport {
	if pb == nil {
		return nil
	}
	return &t.ReplicaDeletedReport{
		ChunkID: t.ChunkID(pb.GetChunkId()),
	}
}

func ReplicaReportFromPB(pb *mpb.ReplicaReport) t.StorageNodeReport {
	if pb == nil {
		return t.StorageNodeReport{}
	}

	switch rec := pb.GetReport().(type) {
	case *mpb.ReplicaReport_Staged:
		return ReplicaStagedReportFromPB(rec.Staged).ToRecord()

	case *mpb.ReplicaReport_ChainFailed:
		return ReplicaChainFailedReportFromPB(rec.ChainFailed).ToRecord()
	
	case *mpb.ReplicaReport_Deleted:
		return ReplicaDeletedReportFromPB(rec.Deleted).ToRecord()

	default:
		return t.StorageNodeReport{}
	}
}

func MasterRefFromPB(pb *mpb.MasterRef) t.MasterRef {
	return t.MasterRef{
		ID: t.MasterID(pb.GetMasterId()),
		Addr: pb.GetAddr(),
	}
}
