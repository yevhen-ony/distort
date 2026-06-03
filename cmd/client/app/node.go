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

type InspectNodeResult struct {
	Report InspectReport
}

type InspectReport struct {
	Addr   string               `json:"addr"`
	Stats  t.NodeStats          `json:"stats"`
	Chunks []t.ChunkStorageView `json:"chunks"`
}

func (app *App) InspectNode(ctx context.Context, addr string) (*InspectNodeResult, error) {
	insp, err := app.StorageAdminT.Inspect(ctx, addr)
	if err != nil {
		return nil, err
	}

	res := &InspectNodeResult{
		Report: InspectReport{
			Addr: addr,
			Stats: insp.Stats,
			Chunks: insp.Chunks,
		},
	}
	return res, nil
	
}
