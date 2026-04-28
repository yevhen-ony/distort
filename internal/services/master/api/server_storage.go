package api

import (
	"context"
	pb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	m "dos/internal/services/master"
	"fmt"
)

type StorageServer struct {
	pb.UnimplementedMasterStorageServiceServer
	service m.Service
}

func NewStorageServer(service m.Service) *StorageServer {
	return &StorageServer{service: service}
}

func (s *StorageServer) RegisterStorageNode(
	ctx context.Context,
	req *pb.RegisterStorageNodeRequest,
) (rsp *pb.RegisterStorageNodeResponse, err error) {
	
	defer func() { err = toStatus(err) }()
	
	if err = validateRegisterStorageNodeRequest(req); err != nil {
		return nil, err
	}

	addr := req.GetAddr()
	nodeRef, err := s.service.RegisterStorageNode(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("reg node: %w", err)
	}
	
	rsp.NodeId = string(nodeRef.ID)
	return rsp, nil 
}

func (s *StorageServer) Heartbeat(
	ctx context.Context,
	req *pb.HeartbeatRequest,
) (rsp *pb.HeartbeatResponse, err error) {
	
	defer func() { err = toStatus(err) }()

	if err = validateHeartbeatRequest(req); err != nil {
		return nil, err
	}

	nodeID := t.NodeID(req.GetNodeId())
	stats := convert.NodeStatsFromPB(req.GetStats())

	if err = s.service.Heartbeat(ctx, nodeID, *stats); err != nil {
		return nil, fmt.Errorf("heartbeat: %w", err)
	}

	rsp = &pb.HeartbeatResponse{}
	return rsp, nil 

}

func (s *StorageServer) ReportStorage(
	ctx context.Context,
	req *pb.ReportStorageRequest,
) (rsp *pb.ReportStorageResponse, err error) {

	defer func() { err = toStatus(err) }()
	
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
		return nil, fmt.Errorf("node %s: report chunk storge: %w", nodeID, err)
	}
	
	pbRejects := make([]*pb.StorageReject, 0, len(rejects))
	for _, rej := range rejects {
		pbRejects = append(pbRejects, &pb.StorageReject{
			ChunkId: string(rej.ChunkID),
			Reason: rej.Reason,
		})
	}
	rsp.Rejects = pbRejects
	return rsp, nil
}


