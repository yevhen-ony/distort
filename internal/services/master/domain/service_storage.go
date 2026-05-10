package domain

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"errors"
	"fmt"
	"log/slog"
)

func (s *MasterService) RegisterStorageNode(ctx context.Context, addr string) (t.NodeRef, error) {

	nref, err := s.nodeReg.Register(ctx, addr)
	if err != nil {
		return t.NodeRef{}, fmt.Errorf("register node: %w", err)
	}
	return nref, err
}

func (s *MasterService) ReportReplicas(
	ctx context.Context, nodeID t.NodeID, reports []t.ReplicaReport,
) (t.ReportResult, error) {

	if _, err := s.nodeReg.Get(ctx, nodeID); err != nil {
		return t.ReportResult{}, fmt.Errorf("get node %s: %w", nodeID, err)
	}

	result := t.ReportResult{}

	for _, report:= range reports {
		if report.ReplicaStaged != nil {
			err := s.reportStagedReplica(ctx, nodeID, report.ReplicaStaged)
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
			s.reconcileSink.Enqueue(ctx, report.ReplicaChainFailed.ChunkID)
		}
	}
	return result, nil
}

func (s *MasterService) reportStagedReplica(
	ctx context.Context, nodeID t.NodeID, report *t.ReplicaStagedReport,
) error {
	meta := report.Chunk

	if err := s.chunkRepo.SetDigest(ctx, meta.ID, meta.Digest); err != nil {
		slog.WarnContext(ctx, "reject chunk report", "chunk_id", meta.ID, "reason", err)
		return err
	}
	if s.index.AttachChunk(ctx, nodeID, meta.ID) {
		s.chunkRepo.IncReplication(ctx, meta.ID)
	}
	return nil
}

func (s *MasterService) Heartbeat(ctx context.Context, nodeID t.NodeID, stats t.NodeStats) error {
	if _, err := s.nodeReg.Get(ctx, nodeID); err != nil {
		return fmt.Errorf("get node %s: %w", nodeID, err)
	}
	if err := s.nodeReg.UpdateStats(ctx, nodeID, stats); err != nil {
		return fmt.Errorf("update stats for node %s: %w", nodeID, err)
	}
	return nil 
}

func (s *MasterService) GetCandidateNodes(
	ctx context.Context, query m.CandidateNodesQuery,
) ([]m.Node, error) {

	nodesToExclude := s.index.GetChunkNodes(ctx, query.ExcludeChunk)
	nodes, err := s.nodeReg.Find(ctx, m.NodeQuery{
		MinFreeBytes: query.MinFreeBytes,
		ExcludeIDs: nodesToExclude,
	})
	if err != nil {
		return []m.Node{}, err
	}
	return nodes, nil	
}

func (s *MasterService) EvictStorageNode(ctx context.Context, nodeID t.NodeID) error {
	if _, err := s.nodeReg.Get(ctx, nodeID); err != nil {
		return fmt.Errorf("get node %s: %w", nodeID, err)
	}
	
	var errs []error
	chunks := s.index.GetNodeChunks(ctx, nodeID)
	for _, chunk := range chunks {
		if err := s.chunkRepo.DecReplication(ctx, chunk); err != nil {
			errs = append(errs, fmt.Errorf("dec replica for chunk %s: %w", chunk, err))
		}
	}

	s.index.DetachNode(ctx, nodeID)

	if err := s.nodeReg.Unregister(ctx, nodeID); err != nil {
		errs = append(errs, fmt.Errorf("unregister node %s: %w", nodeID, err))
	}
	return errors.Join(errs...)
}

