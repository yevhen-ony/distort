package storagenode

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"errors"
	"fmt"
	"log/slog"
)

type ReportDeps struct {
	ChunkRepo      m.ChunkRepo
	NodeRegistry   m.NodeRegistry
	ChunkNodeIndex m.ChunkNodeIndex
	Replication    m.ReplicaScheduler
	Metrics        *ReportMetrics
}

type ReportService struct {
	chunkNodeIndex m.ChunkNodeIndex
	chunkRepo      m.ChunkRepo
	nodeRegistry   m.NodeRegistry

	replication m.ReplicaScheduler

	metrics *ReportMetrics
}

func NewReportService(deps ReportDeps) (*ReportService, error) {
	if deps.ChunkNodeIndex == nil {
		return nil, errors.New("missing chunk-node index")
	}
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repository")
	}
	if deps.NodeRegistry == nil {
		return nil, errors.New("missing node registry")
	}
	if deps.Replication == nil {
		return nil, errors.New("missing replication scheduler")
	}
	if deps.Metrics == nil {
		return nil, errors.New("missing metrics")
	}
	service := &ReportService{
		chunkNodeIndex: deps.ChunkNodeIndex,
		chunkRepo:      deps.ChunkRepo,
		nodeRegistry:   deps.NodeRegistry,
		replication:    deps.Replication,
		metrics:        deps.Metrics,
	}
	return service, nil
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
			s.metrics.ReplicaChainFailedTotal.Inc()
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

	if err := s.chunkRepo.SetDigest(ctx, meta.ID, meta.Digest); err != nil {
		slog.WarnContext(ctx, "reject chunk report", "chunk_id", meta.ID, "reason", err)
		s.metrics.StagedReplicasRejectedTotal.Inc()
		return err
	}
	if s.chunkNodeIndex.AttachChunk(ctx, nodeID, meta.ID) {
		_ = s.chunkRepo.IncReplicaCount(ctx, meta.ID)
		s.metrics.StagedReplicasAcceptedTotal.Inc()
	}
	return nil
}

func (s *ReportService) reportDeletedReplica(
	ctx context.Context, nodeID t.NodeID, report *t.ReplicaDeletedReport,
) {
	chunkID := report.ChunkID
	slog.DebugContext(ctx, "deleted replica reported", "chunk_id", chunkID)
	s.metrics.DeletedReplicasTotal.Inc()

	if s.chunkNodeIndex.DetachChunk(ctx, nodeID, chunkID) {
		s.chunkRepo.DecReplicaCount(ctx, chunkID)
	}
}
