package logger

import (
	"context"
	"dos/internal/common/dosctx"
	"log/slog"
)

type ContextHandler struct {
	next slog.Handler
}

func NewContextHandler(next slog.Handler) slog.Handler {
	return &ContextHandler{
		next: next,
	}
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if objectID, ok := dosctx.ObjectID(ctx); ok {
		r.AddAttrs(slog.String("object_id", string(objectID)))
	}
	if chunkID, ok := dosctx.ChunkID(ctx); ok {
		r.AddAttrs(slog.String("chunk_id", string(chunkID)))
	}
	if nodeID, ok := dosctx.NodeID(ctx); ok {
		r.AddAttrs(slog.String("node_id", string(nodeID)))
	}
	if service, ok := dosctx.Service(ctx); ok {
		r.AddAttrs(slog.String("service", service))
	}
	if operation, ok := dosctx.Operation(ctx); ok {
		r.AddAttrs(slog.String("operation", operation))
	}
	return h.next.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
  	return &ContextHandler{next: h.next.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
  	return &ContextHandler{next: h.next.WithGroup(name)}
}
