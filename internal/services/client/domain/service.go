package domain

import (
	"context"
	"errors"
	"fmt"
	"io"

	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

var (
	ErrMissingMasterTransport = errors.New("missing master transport")
	ErrMissingStorageTransport = errors.New("missing storage transport")
)

type Service struct {
	master c.MasterTransport
	storage c.StorageTransport
}

func New(master c.MasterTransport, storage c.StorageTransport) (*Service, error) {
	if master == nil {
		return nil, ErrMissingMasterTransport
	}
	if storage == nil {
		return nil, ErrMissingStorageTransport
	}
	svc := &Service{master: master, storage: storage}
	return svc, nil
}

func (s *Service) Push(ctx context.Context, objectID t.ObjectID, source c.ChunkSource) error {

	if err := s.master.CreateObject(ctx, objectID); err != nil {
		return fmt.Errorf("create object: %w", err)
	}
	
	for {
		key, data, err := source.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read chunk: %w", err)
		}
		allocQuery := &c.AllocateChunkQuery{
			ObjectID: objectID,
			ChunkKey: key,
			ChunkSize: int64(len(data)),
		}
		loc, err := s.master.AllocateChunk(ctx, allocQuery)
		if err != nil {
			return fmt.Errorf("alloc chunk: %w", err)
		}

		chunk := c.NewChunk(loc.ChunkID, data)
		if err := s.storage.PushChunk(ctx, loc.Nodes, &chunk); err != nil {
			return err 
		}
	}
	return nil
}

func (s *Service) Pull(ctx context.Context, objectID t.ObjectID, asm c.ObjectAssembler) error {

	access, err := s.master.GetObjectAccess(ctx, objectID)
	if err != nil {
		return fmt.Errorf("get object access: %w", err)
	}

	chunkDescs := make([]t.ChunkDesc, len(access.Chunks))
	for i, cp := range access.Chunks {
		chunkDescs[i] = cp.ChunkDesc
	}

	ow, err := asm.NewWriter(access.ObjectDesc, chunkDescs)
	if err != nil {
		return fmt.Errorf("new object writer: %w", err)
	}
	defer ow.Close()

	for _, cp := range access.Chunks {
		chunk, err := s.storage.PullChunk(ctx, cp.Nodes, cp.ChunkID)
		if err != nil {
			return fmt.Errorf("pull chunk %s: %w", cp.ChunkID, err) 
		}
		if err := ow.WriteChunk(chunk.Meta.ID, chunk.Data); err != nil {
			return fmt.Errorf("write chunk %s: %w", cp.ChunkID, err)
		}
	}
	return nil
}
