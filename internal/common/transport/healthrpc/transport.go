package healthrpc

import (
	"context"
	"errors"
	"fmt"

	cpb "dos/gen/proto/common/v1"
	"dos/internal/common/connect"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Transport struct {
	conn *connect.ConnCache
}

func NewTransport(conn *connect.ConnCache) (*Transport, error) {
	if conn == nil {
		return nil, errors.New("missing conn")
	}
	health := &Transport{
		conn: conn,
	}
	return health, nil
}

type HealthResult struct {
	Component string
	Ready     bool
}

func (ht *Transport) Ready(ctx context.Context, addr string) (*HealthResult, error) {
	conn, err := ht.conn.Get(addr)
	if err != nil {
		return nil, fmt.Errorf("create connection: %w", err)
	}

	client := cpb.NewHealthServiceClient(conn)
	rsp, err := client.Ready(ctx, &cpb.ReadyRequest{})

	if err != nil {
		if status.Code(err) == codes.Unavailable {
			res := &HealthResult{
				Component: "unknown",
				Ready:     false,
			}
			return res, nil
		}
		return nil, fmt.Errorf("transport: %w", err)
	}

	res := &HealthResult{
		Component: ComponentFromPB(rsp.GetComponent()),
		Ready:     true,
	}
	return res, nil
}

func ComponentFromPB(c cpb.Component) string {
	switch c {
	case cpb.Component_COMPONENT_MASTER:
		return "master"
	case cpb.Component_COMPONENT_STORAGE:
		return "storage"
	default:
		return "unknown"
	}
}
