package app

import (
	"context"

	"dos/internal/common/transport/storage/adminrpc"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
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
	Addr      string               `json:"addr"`
	Stats     t.NodeStats          `json:"stats"`
	Chunks    []t.ChunkStorageView `json:"chunks"`
	Heartbeat t.HeartbeatView      `json:"heartbeat"`
}

func (app *App) InspectNode(ctx context.Context, addr string) (*InspectNodeResult, error) {
	insp, err := app.StorageAdminT.Inspect(ctx, addr)
	if err != nil {
		return nil, err
	}

	res := &InspectNodeResult{
		Report: InspectReport{
			Addr:      addr,
			Stats:     insp.Stats,
			Chunks:    insp.Chunks,
			Heartbeat: insp.Heartbeat,
		},
	}
	return res, nil
}

type TriggerReportQuery struct {
	Addr     string
	All      bool
	ChunkIDs []string
}

type TriggerReportResult struct {
	Report TriggerReportReport
}

type TriggerReportReport struct {
	Scheduled []t.ChunkID `json:"scheduled"`
	Failed    []t.ChunkID `json:"failed"`
}

func (app *App) TriggerReport(ctx context.Context, q TriggerReportQuery) (*TriggerReportResult, error) {
	out, err := app.StorageAdminT.TriggerReport(ctx, adminrpc.TriggerReportQuery{
		Addr:     q.Addr,
		All:      q.All,
		ChunkIDs: utils.Map(q.ChunkIDs, func(id string) t.ChunkID { return t.ChunkID(id) }),
	})
	if err != nil {
		return nil, err
	}

	res := &TriggerReportResult{
		Report: TriggerReportReport{
			Scheduled: out.Scheduled,
			Failed:    out.Failed,
		},
	}
	return res, nil
}

type HeartbeatControlResult struct {
	Heartbeat t.HeartbeatView `json:"heartbeat"`
}

func (app *App) PauseHeartbeat(ctx context.Context, addr string) (*HeartbeatControlResult, error) {
	pauseRes, err := app.StorageAdminT.PauseHeartbeat(ctx, addr)
	if err != nil {
		return nil, err
	}
	res := &HeartbeatControlResult{
		Heartbeat: pauseRes.State,
	}
	return res, nil
}

func (app *App) ResumeHeartbeat(ctx context.Context, addr string) (*HeartbeatControlResult, error) {
	resumeRes, err := app.StorageAdminT.ResumeHeartbeat(ctx, addr)
	if err != nil {
		return nil, err
	}
	res := &HeartbeatControlResult{
		Heartbeat: resumeRes.State,
	}
	return res, nil
}
