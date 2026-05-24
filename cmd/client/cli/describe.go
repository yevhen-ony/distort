package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	t "dos/internal/common/types"
)

func (app *App) DescribeChunk(ctx context.Context, chunkID string) error {
	placement, err := app.MasterTransport.DescribeChunk(ctx, t.ChunkID(chunkID))
	if err != nil {
		return err
	}

	b := &strings.Builder{}
	RenderChunkMeta(b, placement.Meta)
	RenderObjectSlot(b, placement.Slot)
	RenderSources(b, placement.Sources)

	fmt.Print(b.String())

	return nil
}

func RenderChunkMeta(out io.Writer, meta t. ChunkMeta) {
	
	checksum := ""
	size := int64(0)
	if meta.Digest != nil {
		checksum = string(meta.Digest.Checksum)
		size = meta.Digest.Size
	}
	
	fmt.Fprintln(out, "CHUNK META:")
	fmt.Fprintf(out, "\t * chunk_id: %s\n", meta.ID)
	fmt.Fprintf(out, "\t * checksum: %s\n", checksum)
	fmt.Fprintf(out, "\t * size    : %8.1fMB\n", toMB(size))
}

func RenderObjectSlot(out io.Writer, slot t.ObjectSlot) {
	fmt.Fprintln(out, "OBJECT SLOT:")
	fmt.Fprintf(out, "\t * object_id: %s\n", slot.ObjectID)
	fmt.Fprintf(out, "\t * chunk_key: %s\n", slot.ChunkKey)
}

func RenderSources(out io.Writer, sources []t.NodeRef) {
	fmt.Fprintf(out, "SOURCES (STORAGE NODES) %d:\n", len(sources))
	for _, ref := range sources {
		fmt.Fprintf(out, "\t * node_id: %s | node_addr: %s\n", ref.ID, ref.Addr)
	}
}

