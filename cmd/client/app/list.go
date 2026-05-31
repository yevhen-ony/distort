package app

import (
	"context"

	t "dos/internal/common/types"
)

type ListObjectsResult struct {
	Objects []t.ObjectInfo
}

func (app *App) ListObjects(ctx context.Context) (*ListObjectsResult, error) {

	infos, err := app.MasterT().ListObjects(ctx)
	if err != nil {
		return nil, err
	}

	res := &ListObjectsResult{
		Objects: infos, 
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

type ListNodesResult struct {
	Nodes []t.NodeInfo
}

func (app *App) ListNodes(ctx context.Context) (*ListNodesResult, error) {
	infos, err := app.MasterT().ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	res := &ListNodesResult{
		Nodes: infos,
	}
	return res, nil
}


