package main

import (
	"context"
	t "dos/internal/common/types"
	"dos/internal/services/client/io/file"
	"fmt"
)

func (app *App) Push(ctx context.Context, objectID string, path string) error {
	app.progressOutput.Start()
	defer app.progressOutput.Stop()

	chunker, err := file.NewFileChunker(path, &app.Config.Chunker)
	if err != nil {
		return fmt.Errorf("init chunker: %w", err)
	}
	defer chunker.Close()

	return app.Service.Push(ctx, t.ObjectID(objectID), chunker)
}
