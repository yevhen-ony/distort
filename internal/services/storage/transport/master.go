package transport

import (
	"context"
	"errors"
	"fmt"
	"time"

	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	"dos/internal/common/master/route"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MasterTransportConfig interface {
	RPCTimeout() time.Duration
}

type Master struct {
	router *route.MasterRouter
	config MasterTransportConfig
}

type MasterTransportDeps struct {
	Router *route.MasterRouter
	Config MasterTransportConfig
}

func NewMaster(deps MasterTransportDeps) (*Master, error) {
	if deps.Router == nil {
		return nil, errors.New("missing master router")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}
	mt := &Master{
		router: deps.Router,
		config: deps.Config,
	}
	return mt, nil
}

func (mt *Master) client(ctx context.Context) (mpb.MasterStorageServiceClient, error) {

	conn, err := mt.router.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get master conn: %w", err)
	}

	return mpb.NewMasterStorageServiceClient(conn), nil
}

func (mt *Master) RegisterNode(ctx context.Context, addr string) (t.NodeID, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	client, err := mt.client(ctx)
	if err != nil {
		return "", err
	}

	req := &mpb.RegisterStorageNodeRequest{Addr: addr}
	rsp, err := client.RegisterStorageNode(ctx, req)
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

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	client, err := mt.client(ctx)
	if err != nil {
		return s.HeartbeatResult{}, err
	}

	req := &mpb.HeartbeatRequest{
		NodeId: string(nodeID),
		Stats:  convert.NodeStatsToPB(stats),
	}

	_, err = client.Heartbeat(ctx, req)
	if status.Code(err) == codes.NotFound {
		return s.HeartbeatResult{NodeUnknown: true}, nil
	}
	if err != nil {
		return s.HeartbeatResult{}, fmt.Errorf("heartbeat: %w", err)
	}
	return s.HeartbeatResult{}, nil
}

func (mt *Master) ReportChunks(
	ctx context.Context, nodeID t.NodeID, reports []t.StorageNodeReport,
) (t.ReportResult, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	client, err := mt.client(ctx)
	if err != nil {
		return t.ReportResult{}, err
	}

	req := &mpb.ReportStorageRequest{
		NodeId:  string(nodeID),
		Reports: utils.Map(reports, convert.ReplicaReportToPB),
	}
	rsp, err := client.ReportStorage(ctx, req)
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
