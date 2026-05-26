package domain

import (
	"context"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"dos/internal/services/master/domain/catalog"
	"dos/internal/services/master/domain/storagenode"
	"errors"
	"fmt"
	"log/slog"
)

type ClientFacadeConfig interface {
	ReplicationCount() int
}

type ReplicationScheduler interface {
	Schedule(context.Context, t.ChunkID)
}

type ClientFacadeDeps struct {
	Catalog     *catalog.CatalogService
	Placement   *storagenode.PlacementService
	Lifecycle   *storagenode.LifecycleService
	Replication ReplicationScheduler
	Config      ClientFacadeConfig
}

type ClientFacadeService struct {
	catalog   *catalog.CatalogService
	placement *storagenode.PlacementService
	lifecycle *storagenode.LifecycleService
	replicate ReplicationScheduler

	config ClientFacadeConfig
}

func NewClientFacadeService(deps ClientFacadeDeps) (*ClientFacadeService, error) {
	if deps.Catalog == nil {
		return nil, errors.New("missing catalog service")
	}
	if deps.Placement == nil {
		return nil, errors.New("missing placement service")
	}
	if deps.Lifecycle == nil {
		return nil, errors.New("missing lifecycle service")
	}
	if deps.Replication == nil {
		return nil, errors.New("missing replication scheduler")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	service := &ClientFacadeService{
		catalog:   deps.Catalog,
		placement: deps.Placement,
		lifecycle: deps.Lifecycle,
		replicate: deps.Replication,
		config:    deps.Config,
	}
	return service, nil
}

func (s *ClientFacadeService) CreateObject(ctx context.Context, oid t.ObjectID) error {

	return s.catalog.CreateObject(ctx, oid, s.config.ReplicationCount())
}

func (s *ClientFacadeService) DescribeObject(
	ctx context.Context,
	objectID t.ObjectID,
) (*t.ObjectDesc1, error) {
	
	obj, err := s.catalog.GetObject(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	
	placements := make([]t.ChunkPlacement1, 0, len(obj.Chunks))
	size := int64(0)
	for _, chunkID := range obj.Chunks {
		desc, err := s.DescribeChunk(ctx, chunkID)
		if err != nil {
			return nil, fmt.Errorf("describe chunk %s: %w", chunkID, err)
		}
		placements = append(placements, desc.Placement)
		size += desc.Placement.Meta.Digest.Size
	}

	objDesc := t.ObjectDesc1 {
		ID: obj.ID,
		Size: size,
		Replication: obj.Replication,
		Chunks: placements,
	}

	return &objDesc, nil
}
 
func (s *ClientFacadeService) DescribeChunk(
	ctx context.Context,
	chunkID t.ChunkID,
) (*t.ChunkDesc1, error) {

	chunk, err := s.catalog.GetChunk(ctx, chunkID)
	if err != nil {
		return nil, fmt.Errorf("get chunk: %w", err)
	}
	nodes, err := s.placement.GetChunkNodes(ctx, chunkID)
	if err != nil {
		return nil, fmt.Errorf("get chunk's nodes: %w", err)
	}
	placement := t.ChunkPlacement1{
		Meta:    chunk.Meta,
		Slot:    chunk.Slot,
		Sources: nodes,
	}
	desc := &t.ChunkDesc1{Placement: placement}
	return desc, nil
}

func (s *ClientFacadeService) AllocateChunk(
	ctx context.Context,
	cmd m.AllocateChunkCommand,
) (*t.ChunkAllocation1, error) {

	exists, err := s.catalog.ExistsChunk(ctx, cmd.Slot)
	if err != nil {
		return nil, fmt.Errorf("exists chunk: %w", err)
	}
	
	replicaCount, err := s.catalog.GetReplication(ctx, cmd.Slot.ObjectID)
	if err != nil {
		return nil, err
	}

	candidates, err := s.placement.GetCandidates(ctx, m.CandidateNodesQuery{
		MinFreeBytes: cmd.Size,
		MaxCount:     replicaCount,
		ExcludeNodes: cmd.ExcludeNodes,
	})
	if err != nil {
		return nil, fmt.Errorf("get candidate nodes: %w", err)
	}
	if len(candidates) == 0 {
		return nil, m.ErrNoCandidateNodes
	}

	var chunkID t.ChunkID
	if exists {

		chunkID, err = s.catalog.GetChunkID(ctx, cmd.Slot)
		if err != nil {
			return nil, fmt.Errorf("get chunk: %w", err)
		}

		desc, err := s.DescribeChunk(ctx, chunkID)
		if err != nil {
			return nil, fmt.Errorf("describe chunk: %w", err)
		}

		if len(desc.Placement.Sources) > 0 {
			return nil, m.ErrChunkKeyOccupied
		}
	} else {

		chunkID, err = s.catalog.AddChunk(ctx, cmd.Slot, cmd.Size)
		if err != nil {
			return nil, fmt.Errorf("add chunk: %w", err)
		}
	}

 
	res := &t.ChunkAllocation1{
		ID:      chunkID,
		Slot:    cmd.Slot,
		Targets: candidates,
	}
	return res, nil
}

func (s *ClientFacadeService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return s.catalog.ListObjects(ctx)
}

func (s *ClientFacadeService) ListChunks(ctx context.Context) []t.ChunkInfo {
	return s.catalog.ListChunks(ctx)
}

func (s *ClientFacadeService) ListNodes(ctx context.Context) []t.NodeInfo {
	return s.lifecycle.ListNodes(ctx)
}

func (s *ClientFacadeService) SetReplication(ctx context.Context, objectID t.ObjectID, count int) error {
	nodesCount := s.lifecycle.GetNodeCount(ctx)
	if count > nodesCount {
		return fmt.Errorf("requested replica count %d exceeds number of nodes %d", count, nodesCount)
	}
	if err := s.catalog.SetReplication(ctx, objectID, count); err != nil {
		return fmt.Errorf("set replication count: %w", err)
	}

	chunkIDs, err := s.catalog.GetObjectChunks(ctx, objectID)
	if err != nil {
		return fmt.Errorf("get object chunk ids: %w", err)
	}
	for _, chunkID := range chunkIDs {
		s.replicate.Schedule(ctx, chunkID)
		slog.DebugContext(ctx, "replication scheduled", "chunk_id", chunkID)
	}
	return nil
}
