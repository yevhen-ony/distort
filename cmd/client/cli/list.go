package main

import (
	"context"
	"fmt"
	"strings"
)

func (app *App) ListObjects(ctx context.Context) error {
	infos, err := app.ClientService.ListObjects(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b, "%-20s %10s\n", "OBJECT_ID", "CHUNK_COUNT")
	for _, info := range infos {
		fmt.Fprintf(b, "%-20s %10d\n", info.ID, info.ChunkCount)
	}
	fmt.Print(b.String())
	
	return nil
}

func (app *App) ListChunks(ctx context.Context) error {
	infos, err := app.ClientService.ListChunks(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b,
		"%-18s %-10s %-8s %-20s\n",
		"CHUNK_ID", "SIZE", "REPLICAS", "OBJECT_ID",
	)
	for _, info := range infos {
		fmt.Fprintf(b,
			"%-18s %10d %8d %-20s\n",
			info.ID,
			info.Size,
			info.ReplicaCount,
			info.ObjectID,
		)
	}
	fmt.Print(b.String())

	return nil
}

