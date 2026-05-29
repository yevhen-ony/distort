package app

import (
	"context"
	"fmt"
	"strings"
)

func (app *App) ListObjects(ctx context.Context) error {
	infos, err := app.MasterT().ListObjects(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b, "%-20s %11s %11s\n", "OBJECT_ID", "CHUNK_COUNT", "REPLICATION")
	for _, info := range infos {
		fmt.Fprintf(b, "%-20s %11d %11d\n", info.ID, info.ChunkCount, info.Replication)
	}

	fmt.Print(b.String())
	return nil
}

func (app *App) ListChunks(ctx context.Context) error {
	infos, err := app.MasterT().ListChunks(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b,
		"%-18s %-8s %-8s %-20s\n",
		"CHUNK_ID", "SIZE", "REPLICAS", "OBJECT_ID",
	)
	for _, info := range infos {
		fmt.Fprintf(b,
			"%-18s %8s %8d %-20s\n",
			info.ID,
			ToMBStr(info.Size),
			info.ReplicaCount,
			info.ObjectID,
		)
	}

	fmt.Print(b.String())
	return nil
}

func (app *App) ListNodes(ctx context.Context) error {
	infos, err := app.MasterT().ListNodes(ctx)
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	fmt.Fprintf(b,
		"%-18s %-18s %-6s %-8s\n",
		"NODE_ID", "ADDR", "CHUNKS", "SIZE",
	)
	for _, info := range infos {
		fmt.Fprintf(b,
			"%-18s %-18s %6d %8s\n",
			info.ID,
			info.Addr,
			info.ChunkCount,
			ToMBStr(info.UsedBytes),
		)
	}
	fmt.Print(b.String())
	return nil
}

func ToMBStr(bytes int64) string {
	mb := float64(bytes) / float64(1024 * 1024)
	return fmt.Sprintf("%.1fMB", mb)
}

