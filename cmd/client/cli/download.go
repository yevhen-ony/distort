package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	t "dos/internal/common/types"
	"dos/internal/services/client/domain/delivery"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/io/file"
)

func (app *App) Download(ctx context.Context, objectID string, destPath string) error {
	comparer := func (lhs, rhs t.ChunkKey) int {
		if lhs < rhs { return -1 }
		if lhs > rhs { return 1 }
		return 0

	}
	destPath = ResolveOutputPath(destPath, objectID)
	asm, err := file.NewFileObjectAssembler(destPath, comparer)
	if err != nil {
		return fmt.Errorf("create assembler: %w", err)
	}
	
	downloader, err := delivery.NewObjectDelivery(delivery.ObjectDeliveryDeps{
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

	downloader.WithProgress(func(p *progress.ObjectProgress) {
		render.Update(p)
	})

	go render.RunLoop(ctx)

	if err := downloader.Download(ctx, asm); err != nil {
		return fmt.Errorf("download object %s: %w", objectID, err)
	}
	return nil
}

func ResolveOutputPath(path string, objectID string) string {
	if path == "" || strings.HasSuffix(path, string(os.PathSeparator)) {
		return filepath.Join(path, objectID)
	}

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return filepath.Join(path, objectID)
	}

	return path
}
