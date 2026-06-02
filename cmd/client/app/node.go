package app

import (
	"context"

	t "dos/internal/common/types"
)


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


