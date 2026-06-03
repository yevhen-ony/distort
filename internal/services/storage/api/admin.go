package api

import (
	"context"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/convert"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"
	"errors"
	"log/slog"
)

type Inventory interface {
	GetStats() t.NodeStats
	ListRecords() []s.ChunkRecord
}

type Storage interface {
	StageAndReportMany(context.Context, []t.ChunkID) *s.TriggerReportResult
	StageAndReportAll(context.Context) *s.TriggerReportResult
}

type AdminDeps struct{
	Inventory Inventory
	Storage Storage
}

type AdminServer struct{
	inventory Inventory
	storage Storage

	spb.UnimplementedAdminServiceServer
}

func NewAdminServer(deps AdminDeps) (*AdminServer, error) {
	if deps.Inventory == nil {
		return nil, errors.New("missing inventory")
	}
	if deps.Storage == nil {
		return nil, errors.New("missing storage")
	}

	admin := &AdminServer{
		inventory: deps.Inventory,
		storage: deps.Storage,
	}
	return admin, nil
}

func (as *AdminServer) Inspect(ctx context.Context, _ *spb.InspectRequest) (*spb.InspectResponse, error) {
	
	ctx = dosctx.WithService(ctx, "admin")
	ctx = dosctx.WithOperation(ctx, "inspect")

	slog.DebugContext(ctx, "inspect requested")

	stats := as.inventory.GetStats()
	recs := as.inventory.ListRecords()
	views := utils.Map(recs, func(r s.ChunkRecord) t.ChunkStorageView {
		return t.ChunkStorageView{
			Meta: r.Meta,
			State: r.State.String(),
		}
	})

	rsp := &spb.InspectResponse{
		Stats: convert.NodeStatsToPB(stats),
		Chunks: utils.Map(views, convert.ChunkStorageViewToPB),
	}
	return rsp, nil
}

func (as *AdminServer) TriggerReport(
	ctx context.Context,
	req *spb.TriggerReportRequest,
) (*spb.TriggerReportResponse, error) {

	ctx = dosctx.WithService(ctx, "admin")
	ctx = dosctx.WithOperation(ctx, "trigger_reports")

	slog.DebugContext(ctx, "trigger report requested")

	var trr *s.TriggerReportResult
	if req.GetAll() {
		trr = as.storage.StageAndReportAll(ctx)
	} else {
		chunkIDs := utils.Map(req.GetChunkIds(), func(id string) t.ChunkID {
			return t.ChunkID(id)
		})
		trr = as.storage.StageAndReportMany(ctx, chunkIDs)
	}

	toStr := func(id t.ChunkID) string { return string(id) }
	res := &spb.TriggerReportResponse{
		Scheduled: utils.Map(trr.Scheduled, toStr),
		Failed: utils.Map(trr.Failed, toStr),
	}
	return res, nil
}

