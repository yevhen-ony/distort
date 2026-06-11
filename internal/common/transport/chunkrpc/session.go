package chunkrpc

import (
	"context"
	t "dos/internal/common/types"
	"errors"
	"fmt"
	"log/slog"
)

type Uploader interface {
	Upload(context.Context, t.NodeRef, *t.Chunk, []t.NodeRef) error
}

type Downloader interface {
	Download(context.Context, t.NodeRef, t.ChunkID) (*t.Chunk, error)
}

type uploadSession struct {
	config   Config
	targets  []t.NodeRef
	uploader Uploader
}

func (s *uploadSession) Upload(ctx context.Context, chunk *t.Chunk) (t.NodeRef, error) {
	var errs []error
	others := make([]t.NodeRef, 0, len(s.targets))
	for i, target := range s.targets {
		others = others[:0]
		others = append(others, s.targets[:i]...)
		others = append(others, s.targets[i+1:]...)

		uploadCtx, cancel := context.WithTimeout(ctx, s.config.RPCTimeout())
		err := s.uploader.Upload(uploadCtx, target, chunk, others)
		cancel()
		if err == nil {
			return target, nil
		}
		slog.WarnContext(ctx,
			"send chunk failed",
			"addr", target.Addr,
			"chunk", chunk.Meta.ID,
			"error", err,
		)
		errs = append(errs, fmt.Errorf("send chunk %s to %s failed: %w", chunk.Meta.ID, target.Addr, err))
	}
	return t.NodeRef{}, fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))
}

type downloadSession struct {
	config     Config
	targets    []t.NodeRef
	downloader Downloader
}

func (s *downloadSession) Download(ctx context.Context, chunkID t.ChunkID) (t.Chunk, error) {
	var errs []error
	for _, node := range s.targets {
		dlCtx, cancel := context.WithTimeout(ctx, s.config.RPCTimeout())
		chunk, err := s.downloader.Download(dlCtx, node, chunkID)
		cancel()

		if err == nil {
			return *chunk, nil
		}
		slog.WarnContext(ctx, "pull chunk failed", "addr", node.Addr, "chunk", chunkID, "error", err)
		errs = append(errs, fmt.Errorf("receive chunk %s from %s: %w", chunkID, node.Addr, err))
	}
	return t.Chunk{}, fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))
}
