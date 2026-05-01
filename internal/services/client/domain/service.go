package domain

import (
	"context"
	"fmt"
	"io"

	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

type Service struct {
	master c.MasterTransport
	storage c.StorageTransport

	assembler c.ObjectAssembler
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


// func (* Service) Pull(ctx context.Context, objectID t.ObjectID, sink c.ChunkSink) 
