package dosctx

import (
	"context"

	t "dos/internal/common/types"
)

type objectIDKey struct{}
type chunkIDKey struct{}
type serviceKey struct{}
type operationKey struct{}

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

func WithService(ctx context.Context, service string)  context.Context {
	return context.WithValue(ctx, serviceKey{}, service)
}

func Service(ctx context.Context) (string, bool) {
	service, ok := ctx.Value(serviceKey{}).(string)
	return service, ok
}

func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, operationKey{}, operation) 
}

func Operation(ctx context.Context) (string, bool) {
	operation, ok := ctx.Value(operationKey{}).(string)
	return operation, ok
}




