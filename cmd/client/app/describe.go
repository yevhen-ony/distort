package app

import (
	"context"

	t "dos/internal/common/types"
)

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


type DescribeObjectResult struct {
	Description *t.ObjectDesc1
}

func (app *App) DescribeObject(ctx context.Context, objectID string) (*DescribeObjectResult, error) {

	desc, err := app.MasterT().DescribeObject(ctx, t.ObjectID(objectID))
	if err != nil {
		return nil, err
	}
	res := &DescribeObjectResult{
		Description: desc,
	}

	return res, nil 
}

