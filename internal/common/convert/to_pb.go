package convert

import (
	cpb "dos/gen/proto/common/v1"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

func NodeRefToPB(ref t.NodeRef) *cpb.NodeRef {
	return &cpb.NodeRef{
		NodeId: string(ref.ID),
		Addr:   ref.Addr,
	}
}

func ChunkPlacementToPB(cp t.ChunkPlacement) *mpb.ChunkPlacement {
	return &mpb.ChunkPlacement{
		ChunkId:   string(cp.ChunkID),
		ChunkKey:  string(cp.ChunkKey),
		ChunkSize: cp.ChunkSize,
		Nodes: utils.Map(cp.Nodes, func(ref t.NodeRef) *cpb.NodeRef {
			return NodeRefToPB(ref)
		}),
	}
}

func DigestToPB(d *digest.Digest) *cpb.Digest {
	return &cpb.Digest{
		Checksum: string(d.Checksum),
		Size:     d.Size,
	}
}

func NodeStatsToPB(stat t.NodeStats) *cpb.NodeStats {
	return &cpb.NodeStats{
		FreeBytes:  stat.FreeBytes,
		UsedBytes:  stat.UsedBytes,
		ChunkCount: int32(stat.ChunkCount),
	}
}

func ChunkMetaToPB(meta t.ChunkMeta) *cpb.ChunkMeta {
	return &cpb.ChunkMeta{
		ChunkId: string(meta.ID),
		Digest:  DigestToPB(meta.Digest),
	}
}

func ObjectItemToPB(infos ...t.ObjectItem) []*mpb.ObjectItem {
	pbInfos := make([]*mpb.ObjectItem, len(infos))
	for i, info := range infos {
		pbInfos[i] = &mpb.ObjectItem{
			ObjectId:   string(info.ID),
			ChunkCount: int64(info.ChunkCount),
		}
	}
	return pbInfos
}

func ReportResultToPB(res t.ReportResult) *mpb.ReportStorageResponse {
	accepted := utils.Map(res.Accepted, func(cid t.ChunkID) string { return string(cid) })
	rejected := utils.Map(res.Rejected, func(cid t.ChunkID) string { return string(cid) })
	rsp := &mpb.ReportStorageResponse{
		Accepted: accepted,
		Rejected: rejected,
	}
	return rsp
}

func ReplicaStagedReportToPB(r t.ReplicaStagedReport) *mpb.ReplicaStaged {
	return &mpb.ReplicaStaged{
		Chunk: ChunkMetaToPB(r.Chunk),
	}
}

func ReplicaChainFailedReportToPB(r t.ReplicaChainFailedReport) *mpb.ReplicaChainFailed {
	return &mpb.ReplicaChainFailed{
		ChunkId: string(r.ChunkID),
		Targets: utils.Map(r.Targets, NodeRefToPB),
	}
}

func ReplicaReportToPB(rr t.StorageNodeReport) *mpb.ReplicaReport {
	switch {
	case rr.ReplicaStaged != nil:
		return &mpb.ReplicaReport{
			Record: &mpb.ReplicaReport_Staged{
				Staged: ReplicaStagedReportToPB(*rr.ReplicaStaged),
			},
		}

	case rr.ReplicaChainFailed != nil:
		return &mpb.ReplicaReport{
			Record: &mpb.ReplicaReport_ChainFailed{
				ChainFailed: ReplicaChainFailedReportToPB(*rr.ReplicaChainFailed),
			},
		}

	default:
		return &mpb.ReplicaReport{}
	}
}

