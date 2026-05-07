package domain

import (
	"context"
	"errors"
	"fmt"
	"io"

	"dos/internal/common/transport/chunkrpc"
	t "dos/internal/common/types"
	"dos/internal/services/client/transport"
	"dos/internal/services/client/io/file"
)

var (
	ErrMissingMasterTransport  = errors.New("missing master transport")
	ErrMissingStorageTransport = errors.New("missing storage transport")
)

type Service struct {
	master  *transport.MasterTransport
	storage *chunkrpc.Transport

	onProgress func(*ObjectProgress)
}

type ServiceOption func(*Service)

func NewService(
	master *transport.MasterTransport,
	storage *chunkrpc.Transport,
	opts ...ServiceOption,
) (*Service, error) {

	if master == nil {
		return nil, ErrMissingMasterTransport
	}
	if storage == nil {
		return nil, ErrMissingStorageTransport
	}
	svc := &Service{master: master, storage: storage}
	svc.onProgress = func(*ObjectProgress) {} // nop by default

	for _, opt := range opts {
		opt(svc)
	}
	return svc, nil
}

func (s *Service) Push(ctx context.Context, objectID t.ObjectID, source *file.ObjectChunker) error {

	if err := s.master.CreateObject(ctx, objectID); err != nil {
		return fmt.Errorf("create object: %w", err)
	}
	progress := NewObjectProgress(objectID)
	s.onProgress(progress)

	for {
		key, data, err := source.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read chunk: %w", err)
		}
		allocQuery := &transport.AllocateChunkQuery{
			ObjectID:  objectID,
			ChunkKey:  key,
			ChunkSize: int64(len(data)),
		}
		loc, err := s.master.AllocateChunk(ctx, allocQuery)
		if err != nil {
			return fmt.Errorf("alloc chunk: %w", err)
		}

		chunk := t.NewChunk(loc.ChunkID, data)
		opt := chunkrpc.WithProgressHandler(func(cp chunkrpc.Progress) {
			progress.UpdateChunk(key, cp)		
			s.onProgress(progress)
		})

		session := s.storage.NewTransferSession(loc.Nodes, opt) 
		if err := session.Upload(ctx, &chunk); err != nil {
			return err
		}
	}
	progress.Done = true
	return nil
}

func (s *Service) Pull(ctx context.Context, objectID t.ObjectID, asm *file.ObjectAssembler) error {

	access, err := s.master.GetObjectAccess(ctx, objectID)
	if err != nil {
		return fmt.Errorf("get object access: %w", err)
	}

	progress := NewObjectProgress(objectID)
	s.onProgress(progress)

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
		
		opt := chunkrpc.WithProgressHandler(func(prog chunkrpc.Progress) {
			progress.UpdateChunk(cp.ChunkKey, prog)		
			s.onProgress(progress)
		})
		session := s.storage.NewTransferSession(cp.Nodes, opt)
		chunk, err := session.Download(ctx, cp.ChunkID)
		if err != nil {
			return fmt.Errorf("pull chunk %s: %w", cp.ChunkID, err)
		}
		if err := ow.WriteChunk(chunk.Meta.ID, chunk.Data); err != nil {
			return fmt.Errorf("write chunk %s: %w", cp.ChunkID, err)
		}
	}
	progress.Done = true
	s.onProgress(progress)
	return nil
}

func (s *Service) List(ctx context.Context) ([]t.ObjectItem, error) {
	items, err := s.master.ListObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err) 
	}
	return items, nil
}

func WithProgressHandler(h func(*ObjectProgress)) ServiceOption {
	return func(s *Service) {
		s.onProgress = h
	}
}
