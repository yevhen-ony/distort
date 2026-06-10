package app

import (
	"context"
	"fmt"
	"io"
	"os"

	t "dos/internal/common/types"
	"dos/internal/services/client/transport"
)

type DownloadChunkQuery struct {
	NodeAddr string
	NodeID   string
	ChunkID  string
	DestPath string
}

type DownloadChunkResult struct {
	Source t.NodeRef   `json:"source"`
	Meta   t.ChunkMeta `json:"meta"`
	Path   string      `json:"path"`
}

func (app *App) DownloadChunk(ctx context.Context, q DownloadChunkQuery) (*DownloadChunkResult, error) {

	source := t.NodeRef{ID: t.NodeID(q.NodeID), Addr: q.NodeAddr}
	session := app.ChunkT.NewDownloadSession([]t.NodeRef{source})

	chunk, err := session.Download(ctx, t.ChunkID(q.ChunkID))
	if err != nil {
		return nil, fmt.Errorf("download chunk %s: %w", q.ChunkID, err)
	}

	destPath := ResolveOutputPath(q.DestPath, q.ChunkID)
	file, err := NewFile(destPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	n, err := file.Write(chunk.Data)
	if err != nil {
		return nil, fmt.Errorf("chunk write: %w", err)
	}
	if n != len(chunk.Data) {
		return nil, io.ErrShortWrite
	}

	res := &DownloadChunkResult{
		Source: source,
		Meta:   chunk.Meta,
		Path:   destPath,
	}
	return res, nil
}

type AllocateChunkQuery struct {
	ObjectID string
	ChunkKey string
}

type AllocateChunkResult struct {
	ChunkID  t.ChunkID   `json:"chunk_id"`
	ObjectID t.ObjectID  `json:"object_id"`
	ChunkKey t.ChunkKey  `json:"chunk_key"`
	Targets  []t.NodeRef `json:"targets"`
}

func (app *App) AllocateChunk(ctx context.Context, q AllocateChunkQuery) (*AllocateChunkResult, error) {
	slot := t.ObjectSlot{
		ObjectID: t.ObjectID(q.ObjectID),
		ChunkKey: t.ChunkKey(q.ChunkKey),
	}

	alloc, err := app.MasterT().AllocateChunk(ctx, &transport.AllocateChunkCommand{
		Slot:      slot,
		ChunkSize: app.Config.ChunkSize(),
	})
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	res := &AllocateChunkResult{
		ChunkID:  alloc.ID,
		ObjectID: alloc.Slot.ObjectID,
		ChunkKey: alloc.Slot.ChunkKey,
		Targets:  alloc.Targets,
	}
	return res, nil
}

type PushChunkQuery struct {
	NodeID   string
	NodeAddr string
	ChunkID  string
	Path     string
}

type PushChunkResult struct {
	Meta   t.ChunkMeta `json:"meta"`
	Target t.NodeRef   `json:"target"`
	File   string      `json:"file"`
}

func (app *App) PushChunk(ctx context.Context, q PushChunkQuery) (*PushChunkResult, error) {
	target := t.NodeRef{ID: t.NodeID(q.NodeID), Addr: q.NodeAddr}
	session := app.ChunkT.NewUploadSession([]t.NodeRef{target})

	data, err := os.ReadFile(q.Path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	chunk := t.NewChunk(t.ChunkID(q.ChunkID), data)
	_, err = session.Upload(ctx, &chunk)
	if err != nil {
		return nil, fmt.Errorf("transport: %w", err)
	}

	res := &PushChunkResult{
		Meta: chunk.Meta,
		Target: target,
		File: q.Path,
	}
	return res, nil
}

type ListChunksResult struct {
	Chunks []t.ChunkInfo
}

func (app *App) ListChunks(ctx context.Context) (*ListChunksResult, error) {
	infos, err := app.MasterT().ListChunks(ctx)
	if err != nil {
		return nil, err
	}
	
	res := &ListChunksResult{
		Chunks: infos,
	}

	return res, nil 
}

type DescribeChunkResult struct {
	Description *t.ChunkDesc1
}

func (app *App) DescribeChunk(ctx context.Context, chunkID string) (*DescribeChunkResult, error) {
	desc, err := app.MasterT().DescribeChunk(ctx, t.ChunkID(chunkID))
	if err != nil {
		return nil, err
	}
	
	res := &DescribeChunkResult{
		Description : desc,
	}

	return res, nil
}
