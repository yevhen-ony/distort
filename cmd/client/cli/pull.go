package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	t "dos/internal/common/types"
	"dos/internal/services/client/io/file"
)

func (app *App) Pull(ctx context.Context, id string, destPath string) error {
	app.progressOutput.Start()
	defer app.progressOutput.Stop()

	comparer := func (lhs, rhs t.ChunkKey) int {
		if lhs < rhs { return -1 }
		if lhs > rhs { return 1 }
		return 0

	}
	destPath = ResolveOutputPath(destPath, id)
	asm, err := file.NewFileObjectAssembler(destPath, comparer)
	if err != nil {
		return fmt.Errorf("create assembler: %w", err)
	}

	if err := app.ClientService.Pull(ctx, t.ObjectID(id), asm); err != nil {
		return fmt.Errorf("pull object %s: %w", id, err)
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
