package replication

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)

var (
	ErrReplicationAttemptsExhausted = errors.New("all replication attempts exhausted")
)

type Service struct {
	chunkRepo  m.ChunkRepo
	objectRepo m.ObjectRepo
	nodeReg    m.NodeRegistry

	index m.ChunkNodeIndex

	storageTransport *StorageTransport

	queue *Queue
}

func (s *Service) ReconcileChunk(ctx context.Context, chunkID t.ChunkID)  error {

	ctx = dosctx.WithChunkID(ctx, chunkID)

	chunk, err := s.chunkRepo.Get(ctx, chunkID)
	if err != nil {
		return fmt.Errorf("read chunk %s: %w", chunkID, err)
	}

	wantedReplicaCount, err := s.objectRepo.GetReplication(ctx, chunk.ObjectID)
	if err != nil {
		return fmt.Errorf("read object %s: %w", chunk.ObjectID, err) 
	}

	count := wantedReplicaCount - chunk.ReplicaCount
	if count == 0 {
		return nil
	}

	if chunk.ReplicaCount == 0 {
		return nil
	}

	if count > 0 {
		_, err = s.ReplicateChunk(ctx, chunk.ChunkMeta, count)
		if err != nil {
			return fmt.Errorf("replicate chunk %s: %w", chunkID, err)
		}
		return nil
	}

	// count < 0
	err = s.DeleteReplica(ctx, chunk.ChunkMeta, -count)
	if err != nil {
		return fmt.Errorf("delete chunk %s: %w", chunkID, err)
	}
	return nil
}

func (s *Service) ReplicateChunk(ctx context.Context, meta t.ChunkMeta, count int) (t.NodeID, error) {

	sources := s.index.GetChunkNodes(ctx, meta.ID)
	if len(sources) == 0 {
		slog.ErrorContext(ctx, "no replication source found")
		return "", fmt.Errorf("no replication sources found for %s", meta.ID)
	}

	candidates, err := s.nodeReg.Find(ctx, m.NodeQuery{
		MinFreeBytes: meta.Digest.Size,
		ExcludeIDs:   sources,
	})
	if err != nil {
		slog.ErrorContext(ctx, "find candidate nodes failed", "error", err)
		return "", fmt.Errorf("find candidate nodes: %w", err)
	}
	if len(candidates) == 0 {
		slog.ErrorContext(ctx, "no candidate nodes found")
		return "", fmt.Errorf("no candidates found for %s", meta.ID)
	}

	targetNodes := utils.RandomSelect(candidates, count)
	targetRefs := utils.Map(targetNodes, func(node m.Node) t.NodeRef {
		return node.NodeRef 
	})

	for _, sourceID := range utils.RandomSelect(sources, len(sources)) {
		sourceNode, err := s.nodeReg.Get(ctx, sourceID)
		if err != nil {
			slog.ErrorContext(ctx, "read source node failed", "source", sourceID, "error", err)
			continue
		}

		err = s.storageTransport.ReplicateChunk(ctx, meta.ID, sourceNode.NodeRef, targetRefs)
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "source", sourceID, "error", err)
			continue
		}
		return sourceID, nil
	}
	return "", ErrReplicationAttemptsExhausted
}

func (s *Service) DeleteReplica(ctx context.Context, meta t.ChunkMeta, count int) error {

	nodeIDs := s.index.GetChunkNodes(ctx, meta.ID)
	if len(nodeIDs) == 0 {
		slog.ErrorContext(ctx, "no replicas found")
		return fmt.Errorf("no replicas found for chunk %s", meta.ID)
	}

	var errs []error
	for _, nodeID := range utils.RandomSelect(nodeIDs, count) {

		node, err := s.nodeReg.Get(ctx, nodeID)
		if err != nil {
			slog.ErrorContext(ctx, "read source node failed", "node_id", nodeID, "error", err)
			errs = append(errs, fmt.Errorf("get node %s: %w", nodeID, err))
			continue
		}

		err = s.storageTransport.DeleteChunk(ctx, meta.ID, node.NodeRef)
		if err != nil {
			slog.ErrorContext(ctx, "delete chunk failed", "source", nodeID, "error", err)
			errs = append(errs, fmt.Errorf(
				"delete chunk %s from node %s: %w",
				meta.ID, node.ID, err,
			))
		}
	}

	return errors.Join(errs...)
}

func (s *Service) Run(ctx context.Context) {
	for {
		chunkID, err := s.queue.Pop(ctx)	
		if err != nil {
			return
		}
		_ = s.ReconcileChunk(ctx, chunkID)
	}
}

func (s *Service) Enqueue(ctx context.Context, chunkID t.ChunkID) {
	s.queue.Enqueue(ctx, chunkID)
}

