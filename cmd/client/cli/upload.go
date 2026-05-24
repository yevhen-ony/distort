package main

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/services/client/domain/delivery"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/io/file"
	"fmt"
)

func (app *App) Upload(ctx context.Context, objectID string, path string) error {

	chunker, err := file.NewObjectChunker(path, app.Config)
	if err != nil {
		return fmt.Errorf("init chunker: %w", err)
	}
	defer chunker.Close()


	uploader, err := delivery.NewObjectDelivery(delivery.ObjectDeliveryDeps{
		ObjectID: t.ObjectID(objectID),
		MasterT: app.MasterTransport,
		ChunkT: app.StorageTransport,
		Config: app.Config,
	})
	if err != nil {
		return fmt.Errorf("uploader init: %w", err) 
	}

	render := NewProgressRender(app.Config.RenderRefreshInterval())	
	defer render.Close()

	go render.RunLoop(ctx)

	uploader.WithProgress(func(p *progress.ObjectProgress) {
		render.Update(p)
	})

	return uploader.Upload(ctx, chunker)
}
