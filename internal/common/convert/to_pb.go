package convert

import (
	cpb "dos/gen/proto/common/v1"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

func NodeRefToPB(mNode ...t.NodeRef) []*cpb.NodeRef {

	pbNodes	:= make([]*cpb.NodeRef, 0, len(mNode))
	for _, mn := range mNode {
		pbNodes = append(pbNodes, &cpb.NodeRef{
			NodeId: string(mn.ID),
			Addr: mn.Addr,
		})
	}
	return pbNodes
}

func ChunkPlacementToPB(mChunk ...t.ChunkPlacement) []*mpb.ChunkPlacement {

	pbChunks := make([]*mpb.ChunkPlacement, 0, len(mChunk))
	for _, mc := range mChunk{
		pbChunks = append(pbChunks, &mpb.ChunkPlacement{
			ChunkId: string(mc.ChunkID),
			ChunkKey: string(mc.ChunkKey),
			ChunkSize: mc.ChunkSize,
			Nodes: NodeRefToPB(mc.Nodes...),
		})
	}
	return pbChunks 
}

func DigestToPB(d ...*digest.Digest) []*cpb.Digest {

	pbDigests := make([]*cpb.Digest, 0, len(d))
	for _, dgt := range d {
		pbDigests = append(pbDigests, &cpb.Digest{
			Checksum: string(dgt.Checksum),
			Size: dgt.Size,
		})
	}
	return pbDigests
}

func NodeStatsToPB(s ...t.NodeStats) []*cpb.NodeStats {

	pbStats := make([]*cpb.NodeStats, 0, len(s))
	for _, stats := range s {
		pbStats = append(pbStats, &cpb.NodeStats{
			FreeBytes: stats.FreeBytes,
			UsedBytes: stats.UsedBytes,
			ChunkCount: int32(stats.ChunkCount),
		})
	}
	return pbStats
}

func ChunkDescToPB(d ...t.ChunkMeta) []*mpb.ChunkDesc {

	pbDesc := make([]*mpb.ChunkDesc, 0, len(d))
	for _, desc := range d {
		pbDesc = append(pbDesc, &mpb.ChunkDesc{
			ChunkId: string(desc.ID),
			Digest: DigestToPB(desc.Digest)[0],
		})
	}
	return pbDesc
}

func ObjectItemToPB(infos ...t.ObjectItem) []*mpb.ObjectItem {
	pbInfos := make([]*mpb.ObjectItem, len(infos)) 
	for i, info := range infos {
		pbInfos[i] = &mpb.ObjectItem{
			ObjectId: string(info.ID),
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
