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

