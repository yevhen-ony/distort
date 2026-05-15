package storagenode

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"fmt"
	"log/slog"
)

type ReportService struct {
	chunkNodeIndex  m.ChunkNodeIndex
	chunkRepository m.ChunkRepo
	nodeRegistry    m.NodeRegistry

	replication m.ReplicaScheduler
}

func NewReportService(
	chunkNodeIndex m.ChunkNodeIndex,
	chunkRepository m.ChunkRepo,
	nodeRegistry m.NodeRegistry,
	replication m.ReplicaScheduler,
) *ReportService {
	return &ReportService{
		chunkNodeIndex:  chunkNodeIndex,
		chunkRepository: chunkRepository,
		nodeRegistry:    nodeRegistry,
		replication:     replication,
	}
}

func (s *ReportService) Report(
	ctx context.Context, nodeID t.NodeID, reports []t.StorageNodeReport,
) (t.ReportResult, error) {

	ctx = dosctx.WithService(ctx, "report")
	ctx = dosctx.WithNodeID(ctx, nodeID)

	if _, err := s.nodeRegistry.Get(ctx, nodeID); err != nil {
		return t.ReportResult{}, fmt.Errorf("get node %s: %w", nodeID, err)
	}

	result := t.ReportResult{}

	for _, report := range reports {
		if report.ReplicaStaged != nil {
			r := report.ReplicaStaged
			err := s.reportStagedReplica(ctx, nodeID, r)
			if err != nil {
				result.Rejected = append(result.Rejected, report.ReplicaStaged.Chunk.ID)
			}
			continue
		}
		if report.ReplicaChainFailed != nil {
			r := report.ReplicaChainFailed
			slog.WarnContext(ctx, "replica chain failed",
				"chunk_id", r.ChunkID,
				"targets", r.Targets,
			)
			s.replication.Schedule(ctx, r.ChunkID)
			continue
		}
		if report.ReplicaDeleted != nil {
			r := report.ReplicaDeleted
			s.reportDeletedReplica(ctx, nodeID, r)
		}
	}
	return result, nil
}

func (s *ReportService) reportStagedReplica(
	ctx context.Context, nodeID t.NodeID, report *t.ReplicaStagedReport,
) error {
	meta := report.Chunk
	slog.DebugContext(ctx, "staged replica reported", "chunk_id", meta.ID)

	if err := s.chunkRepository.SetDigest(ctx, meta.ID, meta.Digest); err != nil {
		slog.WarnContext(ctx, "reject chunk report", "chunk_id", meta.ID, "reason", err)
		return err
	}
	if s.chunkNodeIndex.AttachChunk(ctx, nodeID, meta.ID) {
		s.chunkRepository.IncReplication(ctx, meta.ID)
	}
	return nil
}

func (s *ReportService) reportDeletedReplica(
	ctx context.Context, nodeID t.NodeID, report *t.ReplicaDeletedReport,
) {
	chunkID := report.ChunkID 
	slog.DebugContext(ctx, "deleted replica reported", "chunk_id", chunkID)

	if s.chunkNodeIndex.DetachChunk(ctx, nodeID, chunkID) {
		s.chunkRepository.DecReplication(ctx, chunkID)
	}
}
