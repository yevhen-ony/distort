package main

import (
	"context"
	"fmt"
	"strings"
)

func (app *App) List(ctx context.Context) error {
	items, err := app.Service.List(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b, "%-20s %10s\n", "OBJECT_ID", "CHUNK_COUNT")
	for _, item := range items {
		fmt.Fprintf(b, "%-20s %10d\n", item.ID, item.ChunkCount)
	}
	fmt.Print(b.String())
	
	return nil
}
