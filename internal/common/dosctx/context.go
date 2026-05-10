package dosctx

import (
	"context"

	t "dos/internal/common/types"
)

type objectIDKey struct{}
type chunkIDKey struct{}

func WithObjectID(ctx context.Context, objectID t.ObjectID) context.Context {
	return context.WithValue(ctx, objectIDKey{}, objectID)
}

func ObjectID(ctx context.Context) (t.ObjectID, bool) {
	objectID, ok := ctx.Value(objectIDKey{}).(t.ObjectID)
	return objectID, ok
}

func WithChunkID(ctx context.Context, chunkID t.ChunkID) context.Context {
	return context.WithValue(ctx, chunkIDKey{}, chunkID)
}

func ChunkID(ctx context.Context) (t.ChunkID, bool) {
	chunkID, ok := ctx.Value(chunkIDKey{}).(t.ChunkID)
	return chunkID, ok
}




