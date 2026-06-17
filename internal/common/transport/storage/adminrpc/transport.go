package adminrpc

import (
	"context"
	"errors"
	"fmt"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type Transport struct {
	conn *connect.ConnCache
}

func NewTransport(conn *connect.ConnCache) (*Transport, error) {
	if conn == nil {
		return nil, errors.New("missing conn")
	}
	admin := &Transport{
		conn: conn,
	}
	return admin, nil
}

func (at *Transport) admin(addr string) (spb.AdminServiceClient, error) {
	conn, err := at.conn.Get(addr)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	return spb.NewAdminServiceClient(conn), nil
}

type InspectResult struct {
	Stats     t.NodeStats
	Chunks    []t.ChunkStorageView
	Heartbeat t.HeartbeatView
}

func (at *Transport) Inspect(ctx context.Context, addr string) (*InspectResult, error) {
	admin, err := at.admin(addr)
	if err != nil {
		return nil, err
	}

	rsp, err := admin.Inspect(ctx, &spb.InspectRequest{})
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}

	res := &InspectResult{
		Stats:     *convert.NodeStatsFromPB(rsp.GetStats()),
		Chunks:    utils.Map(rsp.GetChunks(), convert.ChunkStorageViewFromPB),
		Heartbeat: convert.HeatbeatViewFromPB(rsp.GetHeartbeat()),
	}

	return res, nil
}

type TriggerReportQuery struct {
	Addr     string
	All      bool
	ChunkIDs []t.ChunkID
}

type TriggerReportResult struct {
	Scheduled []t.ChunkID
	Failed    []t.ChunkID
}

func (at *Transport) TriggerReport(
	ctx context.Context,
	q TriggerReportQuery,
) (*TriggerReportResult, error) {

	admin, err := at.admin(q.Addr)
	if err != nil {
		return nil, err
	}

	req := &spb.TriggerReportRequest{All: q.All}
	if !q.All {
		toStr := func(id t.ChunkID) string { return string(id) }
		req.ChunkIds = utils.Map(q.ChunkIDs, toStr)
	}

	rsp, err := admin.TriggerReport(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}

	toChunkID := func(s string) t.ChunkID { return t.ChunkID(s) }
	res := &TriggerReportResult{
		Scheduled: utils.Map(rsp.GetScheduled(), toChunkID),
		Failed:    utils.Map(rsp.GetFailed(), toChunkID),
	}
	return res, nil
}

type HeartbeatControlResult struct {
	State t.HeartbeatView
}

func (at *Transport) PauseHeartbeat(ctx context.Context, addr string) (*HeartbeatControlResult, error) {

	admin, err := at.admin(addr)
	if err != nil {
		return nil, err
	}

	rsp, err := admin.PauseHeartbeat(ctx, &spb.HeartbeatControlRequest{})
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}

	res := &HeartbeatControlResult{
		State: convert.HeatbeatViewFromPB(rsp.GetState()),
	}
	return res, nil
}

func (at *Transport) ResumeHeartbeat(ctx context.Context, addr string) (*HeartbeatControlResult, error) {

	admin, err := at.admin(addr)
	if err != nil {
		return nil, err
	}

	rsp, err := admin.ResumeHeartbeat(ctx, &spb.HeartbeatControlRequest{})
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}

	res := &HeartbeatControlResult{
		State: convert.HeatbeatViewFromPB(rsp.GetState()),
	}
	return res, nil
}
