package route

import (
	"context"
	mpb "dos/gen/proto/master/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"fmt"
)

type MasterDiscoveryService struct {
	conn ConnCache
}

func NewMasterDiscoveryService() *MasterDiscoveryService {
	return &MasterDiscoveryService{
		conn: connect.NewConnCache(),
	}
}

func (mds *MasterDiscoveryService) DiscoverActive(ctx context.Context, masterAddr string) (t.MasterRef, error) {
	conn, err := mds.conn.Get(masterAddr)
	if err != nil {
		return t.MasterRef{}, fmt.Errorf("create connection to %s: %w", masterAddr, err)
	}

	client := mpb.NewMasterDiscoveryServiceClient(conn)
	rsp, err := client.GetActiveMaster(ctx, &mpb.GetActiveMasterRequest{})
	if err != nil {
		return t.MasterRef{}, fmt.Errorf("rpc to %s: %w", masterAddr, err)
	}

	ref := convert.MasterRefFromPB(rsp.GetActive())
	return ref, nil
}

func (mds *MasterDiscoveryService) Close() error {
	return mds.conn.Close()
}
