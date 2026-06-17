package domain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	t "dos/internal/common/types"
	m "dos/internal/services/master"
)

//go:generate mockgen -source=$GOFILE -destination=facade_mocks_test.go -package=domain

type ClientFacadeConfig interface {
	ReplicationCount() int
}

type ReplicationScheduler interface {
	Schedule(context.Context, t.ChunkID)
}

type Catalog interface {
	CreateObject(context.Context, t.ObjectID, int) error
	GetObject(context.Context, t.ObjectID) (m.Object, error)
	SetReplication(context.Context, t.ObjectID, int) error

	AddChunk(context.Context, t.ObjectSlot, int64) (t.ChunkID, error)
	GetChunk(context.Context, t.ChunkID) (m.Chunk, error)
	ExistsChunk(context.Context, t.ObjectSlot) (bool, error)
	GetObjectChunks(context.Context, t.ObjectID) ([]t.ChunkID, error)
	GetChunkID(context.Context, t.ObjectSlot) (t.ChunkID, error)
	GetReplication(context.Context, t.ObjectID) (int, error)
}

type Placement interface {
	GetChunkNodes(context.Context, t.ChunkID) ([]t.NodeRef, error)
	GetCandidates(context.Context, m.CandidateNodesQuery) ([]t.NodeRef, error)
}

type Lifecycle interface {
	Register(context.Context, string) (t.NodeRef, error)
	ListNodes(context.Context) []t.NodeInfo
	GetNodeCount(context.Context) int
}

type ClientFacadeDeps struct {
	Catalog     Catalog
	Placement   Placement
	Lifecycle   Lifecycle
	Replication ReplicationScheduler
	Config      ClientFacadeConfig
}

type ClientFacadeService struct {
	catalog   Catalog
	placement Placement
	lifecycle Lifecycle
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

func (s *ClientFacadeService) AllocateChunk(
	ctx context.Context,
	cmd m.AllocateChunkCommand,
) (*t.ChunkAllocation, error) {

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

		sources, err := s.placement.GetChunkNodes(ctx, chunkID)
		if err != nil {
			return nil, fmt.Errorf("get chunk sources: %w", err)
		}
		if len(sources) > 0 {
			return nil, m.ErrChunkKeyOccupied
		}
	} else {

		chunkID, err = s.catalog.AddChunk(ctx, cmd.Slot, cmd.Size)
		if err != nil {
			return nil, fmt.Errorf("add chunk: %w", err)
		}
	}

	res := &t.ChunkAllocation{
		ID:      chunkID,
		Slot:    cmd.Slot,
		Targets: candidates,
	}
	return res, nil
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
