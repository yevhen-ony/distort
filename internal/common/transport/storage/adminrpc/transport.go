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
	Stats  t.NodeStats
	Chunks []t.ChunkStorageView
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
		Stats:  *convert.NodeStatsFromPB(rsp.GetStats()),
		Chunks: utils.Map(rsp.GetChunks(), convert.ChunkStorageViewFromPB),
	}

	return res, nil
}
