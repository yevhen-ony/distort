package app 

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/services/client/domain/delivery"
	"dos/internal/services/client/io/file"
	"fmt"
)

func (app *App) Upload(ctx context.Context, objectID string, path string) error {

	chunker, err := file.NewObjectChunker(path, app.Config.ChunkSize())
	if err != nil {
		return fmt.Errorf("init chunker: %w", err)
	}
	defer chunker.Close()

	uploader, err := delivery.NewObjectDelivery(delivery.ObjectDeliveryDeps{
		ObjectID: t.ObjectID(objectID),
		MasterT: app.MasterT(),
		ChunkT: app.ChunkT,
		Config: app.Config,
	})
	if err != nil {
		return fmt.Errorf("uploader init: %w", err) 
	}

	if app.onProgress != nil {
		uploader.WithProgress(app.onProgress)
	}

	return uploader.Upload(ctx, chunker)
}
