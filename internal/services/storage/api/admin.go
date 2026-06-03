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

type AdminDeps struct{
	Inventory Inventory
}

type AdminServer struct{
	inventory Inventory

	spb.UnimplementedAdminServiceServer
}

func NewAdminServer(deps AdminDeps) (*AdminServer, error) {
	if deps.Inventory == nil {
		return nil, errors.New("missing inventory")
	}

	admin := &AdminServer{
		inventory: deps.Inventory,
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

