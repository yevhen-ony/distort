package transport

import (
	"context"
	"fmt"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterConfig struct {
	Addr string `yaml:"addr"`
}

type Master struct {
	client mpb.MasterStorageServiceClient
	config *MasterConfig
}

func NewMaster(conn *connect.ConnCache, cfg *MasterConfig) (*Master, error) {

	c, err := conn.Get(cfg.Addr)	
	if err != nil {
		return nil, fmt.Errorf("get conn %s: %w", cfg.Addr, err)
	}
	mt := &Master{
		client: mpb.NewMasterStorageServiceClient(c),
		config: cfg,
	}
	return mt, nil 
}


func (mt *Master) RegisterNode(ctx context.Context, addr string) (t.NodeID, error) {

	req := &mpb.RegisterStorageNodeRequest{ Addr: addr }
	rsp, err := mt.client.RegisterStorageNode(ctx, req)
	if err != nil {
		return "", fmt.Errorf("register node %s: %w", addr, err)
	}
	if rsp == nil || rsp.NodeId == "" {
		return "", fmt.Errorf("register node %s: empty response", addr)
	}
	return t.NodeID(rsp.NodeId), nil
}

func (mt *Master) Heartbeat(
	ctx context.Context, nodeID t.NodeID, stats t.NodeStats,
) (s.HeartbeatResult, error) {

	req := &mpb.HeartbeatRequest{
		NodeId: string(nodeID),
		Stats: convert.NodeStatsToPB(stats)[0],
	}

	_, err := mt.client.Heartbeat(ctx, req)
	if status.Code(err) == codes.NotFound {
		return s.HeartbeatResult{NodeUnknown: true}, nil
	}
	if err != nil {
		return s.HeartbeatResult{}, fmt.Errorf("heartbeat: %w", err)
	}
	return s.HeartbeatResult{}, nil
}



func (mt *Master) ReportChunks(
	ctx context.Context, nodeID t.NodeID, desc []t.ChunkMeta,
) (t.ReportResult, error) {

	req := &mpb.ReportStorageRequest{
		NodeId: string(nodeID),
		ChunkReports: convert.ChunkDescToPB(desc...),
	}
	rsp, err := mt.client.ReportStorage(ctx, req)
	if err != nil {
		return t.ReportResult{}, fmt.Errorf("report storage: %w", err) 
	}
	rejected := make([]t.ChunkID, len(rsp.GetRejected()))
	for i, idVal := range rsp.GetRejected() {
		rejected[i] = t.ChunkID(idVal)
	}
	accepted := make([]t.ChunkID, len(rsp.GetAccepted()))
	for i, idVal := range rsp.GetAccepted() {
		accepted[i] = t.ChunkID(idVal)
	}
	res := t.ReportResult{
		Accepted: accepted,
		Rejected: rejected,
	}
	return res, nil
}


