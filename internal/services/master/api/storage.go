package api

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	m "dos/internal/services/master"
	"fmt"
	"log/slog"
)

var _ mpb.MasterStorageServiceServer = (*StorageServer)(nil)

type StorageServer struct {
	mpb.UnimplementedMasterStorageServiceServer
	
	lifecycle m.StorageNodeLifecycle
	report m.StorageNodeReport
}

func NewStorageServer(
	lifecycle m.StorageNodeLifecycle,
	report m.StorageNodeReport,
) *StorageServer {
	return &StorageServer{
		lifecycle: lifecycle,
		report: report,
	}
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
	nodeRef, err := s.lifecycle.Register(ctx, addr)
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

	if err = s.lifecycle.UpdateStats(ctx, nodeID, *stats); err != nil {
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
	slog.DebugContext(
		ctx, "report storage requested",
		"node_id", req.GetNodeId(),
		"count", len(req.GetReports()),
	)

	if err = validateReportStorageRequest(req); err != nil {
		return nil, err
	}

	nodeID := t.NodeID(req.GetNodeId())
	reports := utils.Map(req.GetReports(), convert.ReplicaReportFromPB)

	result, err := s.report.Report(ctx, nodeID, reports)
	if err != nil {
		return nil, err
	}

	rsp = convert.ReportResultToPB(result)
	return rsp, nil
}
