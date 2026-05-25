package app

import (
	"context"
	"fmt"
	"io"
	"strings"

	t "dos/internal/common/types"
)

func (app *App) DescribeChunk(ctx context.Context, chunkID string) error {
	desc, err := app.MasterT.DescribeChunk(ctx, t.ChunkID(chunkID))
	if err != nil {
		return err
	}
	placement := &desc.Placement

	b := &strings.Builder{}
	RenderChunkMeta(b, placement.Meta)
	RenderObjectSlot(b, placement.Slot)
	RenderSources(b, placement.Sources)

	fmt.Print(b.String())

	return nil
}

func RenderChunkMeta(out io.Writer, meta t. ChunkMeta) {
	fmt.Fprintln(out, "CHUNK META:")
	fmt.Fprintf(out, "\t * chunk_id: %s\n", meta.ID)
	fmt.Fprintf(out, "\t * checksum: %s\n", meta.Digest.Checksum)
	fmt.Fprintf(out, "\t * size    : %.1fMB\n", toMB(meta.Digest.Size))
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

func (app *App) DescribeObject(ctx context.Context, objectID string) error {

	desc, err := app.MasterT.DescribeObject(ctx, t.ObjectID(objectID))
	if err != nil {
		return err
	}

	b := &strings.Builder{}

 	fmt.Fprintln(b, "OBJECT:")
  	fmt.Fprintf(b, "\t * object_id  : %s\n", desc.ID)
  	fmt.Fprintf(b, "\t * total_size : %.1fMB\n", toMB(desc.Size))
  	fmt.Fprintf(b, "\t * chunks     : %d\n", len(desc.Chunks))
  	fmt.Fprintf(b, "\t * replication: %d\n", desc.Replication)

	fmt.Fprintln(b, "CHUNKS:")
  	fmt.Fprintf(b, "%-10s %-18s %11s %8s\n",
  		"KEY",
  		"CHUNK_ID",
  		"SIZE",
  		"REPLICAS",
  	)

  	for _, chunk := range desc.Chunks {

  		fmt.Fprintf(b, "%-10s %-18s %8.1fMB %8d\n",
  			chunk.Slot.ChunkKey,
  			chunk.Meta.ID,
  			toMB(chunk.Meta.Digest.Size),
  			len(chunk.Sources),
  		)
  	}

	fmt.Print(b.String())

	return nil
}

