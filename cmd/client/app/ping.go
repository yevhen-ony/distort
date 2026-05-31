package app

import (
	"context"
)

type PingResult struct {
	Address   string `json:"address"`
	Status    string `json:"status"`
	Component string `json:"component"`
}

func (app *App) Ping(ctx context.Context, addr string) (*PingResult, error) {
	health, err := app.HealthT.Ready(ctx, addr)
	if err != nil {
		return nil, err
	}

	res := &PingResult{
		Address: addr,
		Component: health.Component,
		Status: "not ready",
	}
	if health.Ready {
		res.Status = "ready"
	}
	return res, nil
}
