package app

import (
	"context"
	"fmt"

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

func (app *App) ScaleObject(ctx context.Context, objectID string, count int) error {
	err := app.MasterT().SetReplication(ctx, t.ObjectID(objectID), count)
	if err != nil {
		return err
	}
	return nil
}

type DescribeObjectResult struct {
	Description *t.ObjectDesc
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


type CreateObjectResult struct {
  	ObjectID t.ObjectID `json:"object_id"`
}

func (app *App) CreateObject(ctx context.Context, objectID string) (*CreateObjectResult, error) {
  	if err := app.MasterT().CreateObject(ctx, t.ObjectID(objectID)); err != nil {
  		return nil, fmt.Errorf("create object %s: %w", objectID, err)
  	}

  	return &CreateObjectResult{ObjectID: t.ObjectID(objectID)}, nil
}



