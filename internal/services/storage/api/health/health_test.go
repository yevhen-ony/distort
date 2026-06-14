package health

import (
	"context"
	cpb "dos/gen/proto/common/v1"
	t "dos/internal/common/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthServer_Ready(tt *testing.T) {
  	ctx := context.Background()

  	identity := &fakeIdentity{nodeID: "node-1"}
  	server, err := NewHealthServer(HealthDeps{
  		Identity: identity,
  	})
  	require.NoError(tt, err)

  	rsp, err := server.Ready(ctx, &cpb.ReadyRequest{})

  	require.NoError(tt, err)
  	require.Equal(tt, cpb.Component_COMPONENT_STORAGE, rsp.GetComponent())
}

// fake identity

type fakeIdentity struct {
  	nodeID t.NodeID
  	err    error
}

func (i fakeIdentity) GetID() (t.NodeID, error) { return i.nodeID, i.err }
func (i fakeIdentity) Validate(t.NodeID) error { return i.err }
