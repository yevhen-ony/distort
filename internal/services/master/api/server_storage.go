package api

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"fmt"
	"log/slog"
)

type StorageServer struct {
	mpb.UnimplementedMasterStorageServiceServer
	service m.Service
}

func NewStorageServer(service m.Service) *StorageServer {
	return &StorageServer{service: service}
}

func (s *StorageServer) RegisterStorageNode(
	ctx context.Context,
	req *mpb.RegisterStorageNodeRequest,
) (rsp *mpb.RegisterStorageNodeResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "register storage node failed",
				"addr", req.GetAddr(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "register storage node requested", "addr", req.GetAddr())

	if err = validateRegisterStorageNodeRequest(req); err != nil {
		return nil, err
	}

	addr := req.GetAddr()
	nodeRef, err := s.service.RegisterStorageNode(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("reg node: %w", err)
	}

	rsp = &mpb.RegisterStorageNodeResponse{NodeId: string(nodeRef.ID)}
	return rsp, nil
}

func (s *StorageServer) Heartbeat(
	ctx context.Context,
	req *mpb.HeartbeatRequest,
) (rsp *mpb.HeartbeatResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "heartbeat failed",
				"node_id", req.GetNodeId(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "heartbeat requested", "node_id", req.GetNodeId())

	if err = validateHeartbeatRequest(req); err != nil {
		return nil, err
	}

	nodeID := t.NodeID(req.GetNodeId())
	stats := convert.NodeStatsFromPB(req.GetStats())

	if err = s.service.Heartbeat(ctx, nodeID, *stats); err != nil {
		return nil, err
	}

	rsp = &mpb.HeartbeatResponse{}
	return rsp, nil

}

func (s *StorageServer) ReportStorage(
	ctx context.Context,
	req *mpb.ReportStorageRequest,
) (rsp *mpb.ReportStorageResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "report storage failed",
				"node_id", req.GetNodeId(), "error", err,
			)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "report storage requested", "node_id", req.GetNodeId())

	if err = validateReportStorageRequest(req); err != nil {
		return nil, err
	}

	nodeID := t.NodeID(req.GetNodeId())
	desc := make([]t.ChunkDesc, 0, len(req.GetChunkReports()))
	for _, report := range req.GetChunkReports() {
		d := convert.ChunkDescFromPB(report)
		desc = append(desc, d)
	}

	rejects, err := s.service.ReportChunkStorage(ctx, nodeID, desc)
	if err != nil {
		return nil, err
	}

	pbRejects := make([]*mpb.StorageReject, 0, len(rejects))
	for _, rej := range rejects {
		pbRejects = append(pbRejects, &mpb.StorageReject{
			ChunkId: string(rej.ChunkID),
			Reason:  rej.Reason,
		})
	}
	rsp.Rejects = pbRejects
	return rsp, nil
}
