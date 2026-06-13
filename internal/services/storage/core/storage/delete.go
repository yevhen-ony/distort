package storage

import (
	"context"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"fmt"
	"log/slog"
)

func (cs *StorageService) DeleteChunk(ctx context.Context, chunkID t.ChunkID) error {

	ctx = dosctx.WithChunkID(ctx, chunkID)
	ctx = dosctx.WithOperation(ctx, "delete")

	if !cs.inventory.Has(chunkID) {
		slog.WarnContext(ctx, "delete non-existing chunk")
		return nil
	}

	if err := cs.storageBE.Delete(chunkID); err != nil {
		return fmt.Errorf("delete data from disk: %w", err)
	}

	if cs.inventory.Remove(chunkID) {
		cs.reporter.Report(ctx, t.NewReplicaDeleted(chunkID).ToRecord())
	}
	return nil
}
