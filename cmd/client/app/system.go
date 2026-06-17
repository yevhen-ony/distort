package app

import (
	"context"

	t "dos/internal/common/types"
)

type DiscoverMasterResult struct {
	MasterRef t.MasterRef
}

func (app *App) DiscoverMaster(ctx context.Context) (*DiscoverMasterResult, error) {
	ref, err := app.MasterT().DiscoverMaster(ctx)
	if err != nil {
		return nil, err
	}

	res := &DiscoverMasterResult{
		MasterRef: ref,
	}

	return res, nil
}

func (app *App) TransferLeadership(ctx context.Context) error {
	return app.MasterT().TransferLeadership(ctx)
}

type PingResult struct {
	Address   string `json:"address"`
	Status    string `json:"status"`
	Component string `json:"component"`
}

func (app *App) Ping(ctx context.Context, addr string) (*PingResult, error) {
	health, err := app.StorageHealthT.Ready(ctx, addr)
	if err != nil {
		return nil, err
	}

	res := &PingResult{
		Address:   addr,
		Component: health.Component,
		Status:    "not ready",
	}
	if health.Ready {
		res.Status = "ready"
	}
	return res, nil
}
