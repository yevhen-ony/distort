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

type MasterTransportConfig struct {
	MasterAddr string `yaml:"master_addr"`
}

type MasterTransport struct {
	client mpb.MasterStorageServiceClient
	config *MasterTransportConfig
}

func NewMasterTransport(conn *connect.ConnCache, config *MasterTransportConfig) (*MasterTransport, error) {

	c, err := conn.Get(config.MasterAddr)	
	if err != nil {
		return nil, fmt.Errorf("get conn %s: %w", config.MasterAddr, err)
	}
	mt := &MasterTransport{
		client: mpb.NewMasterStorageServiceClient(c),
		config: config,
	}
	return mt, nil 
}


func (mt *MasterTransport) RegisterStorageNode(ctx context.Context, addr string) (t.NodeID, error) {

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

func (mt *MasterTransport) Heartbeat(
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

func (mt *MasterTransport) ReportChunkStorage(
	ctx context.Context, nodeID t.NodeID, desc []t.ChunkDesc,
) ([]t.ChunkStorageReject, error) {

	req := &mpb.ReportStorageRequest{
		NodeId: string(nodeID),
		ChunkReports: convert.ChunkDescToPB(desc...),
	}
	rsp, err := mt.client.ReportStorage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("report storage: %w", err) 
	}
	
	rejects := make([]t.ChunkStorageReject, 0, len(rsp.GetRejects()))
	for _, pbReject := range rsp.GetRejects() {
		rejects = append(rejects, convert.ChunkStorageRejectFromPB(pbReject))
	}
	return rejects, nil
}


