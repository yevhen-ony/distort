package transport

import (
	"context"
	"fmt"

	pb "dos/gen/proto/master/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

func (mt *MasterTransport) ListObjects(ctx context.Context) ([]t.ObjectInfo, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()
	
	admin, err := mt.admin(ctx)
	if err != nil {
		return nil, err
	}
	
	rsp, err := admin.ListObjects(ctx, &pb.ListObjectsRequest{})
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}
	infos := utils.Map(rsp.GetObjects(), convert.ObjectInfoFromPB)
	return infos, nil
}

func (mt *MasterTransport) ListChunks(ctx context.Context) ([]t.ChunkInfo, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	admin, err := mt.admin(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := admin.ListChunks(ctx, &pb.ListChunksRequest{})
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}
	infos := utils.Map(rsp.GetChunks(), convert.ChunkInfoFromPB)
	return infos, nil
}

func (mt *MasterTransport) ListNodes(ctx context.Context) ([]t.NodeInfo, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	admin, err := mt.admin(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := admin.ListNodes(ctx, &pb.ListNodesRequest{})
	if err != nil {
		return nil, fmt.Errorf("rpc: %w", err)
	}
	infos := utils.Map(rsp.GetNodes(), convert.NodeInfoFromPB)
	return infos, nil
}

func (mt *MasterTransport) DiscoverMaster(ctx context.Context) (t.MasterRef, error) {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	return mt.mrouter.Rediscover(ctx)
}

func (mt *MasterTransport) TransferLeadership(ctx context.Context) error {

	ctx, cancel := context.WithTimeout(ctx, mt.config.RPCTimeout())
	defer cancel()

	admin, err := mt.admin(ctx)
	if err != nil {
		return err
	}
	_, err = admin.TransferLeadership(ctx, &pb.TransferLeadershipRequest{})
	if err != nil {
		return fmt.Errorf("rpc: %w", err)
	}
	return nil
}
