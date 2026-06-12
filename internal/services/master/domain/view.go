package domain

import (
	"context"
	"errors"
	"fmt"

	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
)

type ChunkReader interface {
	List(context.Context) []m.Chunk
	Get(context.Context, t.ChunkID) (m.Chunk, error)
}

type ObjectReader interface {
	List(context.Context) []m.Object
	Get(context.Context, t.ObjectID) (m.Object, error)
}

type NodeReader interface {
	Find(context.Context, m.NodeQuery) []m.Node
}

type NodePlacement interface {
	GetChunkNodes(context.Context, t.ChunkID) ([]t.NodeRef, error)
}

type ResourceViewDeps struct {
	ChunkRepo     ChunkReader
	ObjectReader  ObjectReader
	NodeRegistry  NodeReader
	NodePlacement NodePlacement
}

type ResourceViewService struct {
	objects   ObjectReader
	chunks    ChunkReader
	nodes     NodeReader
	placement NodePlacement
}

func NewResourceViewSerivce(deps ResourceViewDeps) (*ResourceViewService, error) {
	if deps.ChunkRepo == nil {
		return nil, errors.New("missing chunk repo")
	}
	if deps.ObjectReader == nil {
		return nil, errors.New("missing object reader")
	}
	if deps.NodeRegistry == nil {
		return nil, errors.New("missing node registry")
	}
	if deps.NodePlacement == nil {
		return nil, errors.New("missing node placement")
	}
	rvs := &ResourceViewService{
		objects: deps.ObjectReader,
		chunks: deps.ChunkRepo,
		nodes: deps.NodeRegistry,
		placement: deps.NodePlacement,
	}
	return rvs, nil
}

func (s *ResourceViewService) ListObjects(ctx context.Context) []t.ObjectInfo {
	return utils.Map(s.objects.List(ctx), func(o m.Object) t.ObjectInfo {
		return t.ObjectInfo{
			ID:          o.ID,
			ChunkCount:  len(o.Chunks),
			Replication: o.Replication,
		}
	})
}

func (s *ResourceViewService) ListChunks(ctx context.Context) []t.ChunkInfo {
	return utils.Map(s.chunks.List(ctx), func(c m.Chunk) t.ChunkInfo {
		size := int64(0)
		if c.ReplicaCount > 0 {
			size = c.Meta.Digest.Size
		}
		return t.ChunkInfo{
			ID:           c.Meta.ID,
			Size:         size,
			ReplicaCount: c.ReplicaCount,
			ObjectID:     c.Slot.ObjectID,
		}
	})
}

func (s *ResourceViewService) ListNodes(ctx context.Context) []t.NodeInfo {
	nodes := s.nodes.Find(ctx, m.NodeQuery{})

	infos := utils.Map(nodes, func(n m.Node) t.NodeInfo {
		return t.NodeInfo{
			ID:         n.ID,
			Addr:       n.Addr,
			ChunkCount: n.Stats.ChunkCount,
			UsedBytes:  n.Stats.UsedBytes,
		}
	})
	return infos
}

func (s *ResourceViewService) DescribeObject(
	ctx context.Context,
	objectID t.ObjectID,
) (*t.ObjectDesc, error) {

	obj, err := s.objects.Get(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	placements := make([]t.ChunkPlacement, 0, len(obj.Chunks))
	size := int64(0)
	for _, chunkID := range obj.Chunks {
		desc, err := s.DescribeChunk(ctx, chunkID)
		if err != nil {
			return nil, fmt.Errorf("describe chunk %s: %w", chunkID, err)
		}
		placements = append(placements, desc.Placement)
		size += desc.Placement.Meta.Digest.Size
	}

	objDesc := t.ObjectDesc{
		ID:          obj.ID,
		Size:        size,
		Replication: obj.Replication,
		Chunks:      placements,
	}

	return &objDesc, nil
}

func (s *ResourceViewService) DescribeChunk(
	ctx context.Context,
	chunkID t.ChunkID,
) (*t.ChunkDesc, error) {

	chunk, err := s.chunks.Get(ctx, chunkID)
	if err != nil {
		return nil, fmt.Errorf("get chunk: %w", err)
	}
	nodes, err := s.placement.GetChunkNodes(ctx, chunkID)
	if err != nil {
		return nil, fmt.Errorf("get chunk's nodes: %w", err)
	}
	placement := t.ChunkPlacement{
		Meta:    chunk.Meta,
		Slot:    chunk.Slot,
		Sources: nodes,
	}
	desc := &t.ChunkDesc{Placement: placement}
	return desc, nil
}
