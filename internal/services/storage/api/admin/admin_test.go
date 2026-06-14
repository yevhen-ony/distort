package admin

import (
	"context"
	"testing"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAdminServer_Inspect(tt *testing.T) {
	ctx := context.Background()
	f := newAdminFixture(tt)

	admin, err := NewAdminServer(f.deps())
	require.NoError(tt, err)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	stats := t.NodeStats{
		FreeBytes:  100,
		UsedBytes:  5,
		ChunkCount: 1,
	}
	records := []s.ChunkRecord{
		{Meta: chunk.Meta, State: s.ChunkStateActive},
	}

	f.inventory.EXPECT().GetStats().Return(stats)
	f.inventory.EXPECT().ListRecords().Return(records)
	f.heartbeat.EXPECT().IsPaused().Return(false)

	rsp, err := admin.Inspect(ctx, &spb.InspectRequest{})

	require.NoError(tt, err)

	gotStats := convert.NodeStatsFromPB(rsp.GetStats())
	require.NotNil(tt, stats)
	require.Equal(tt, stats, *gotStats)

	require.Len(tt, rsp.GetChunks(), 1)
	gotChunk := convert.ChunkStorageViewFromPB(rsp.GetChunks()[0])
	require.Equal(tt, chunk.Meta.ID, gotChunk.Meta.ID)
	require.Equal(tt, "active", gotChunk.State)

	require.Equal(tt, "running", rsp.GetHeartbeat().GetStatus())
}

func TestAdminServer_TriggerReport(tt *testing.T) {
	ctx := context.Background()
	f := newAdminFixture(tt)

	admin, err := NewAdminServer(f.deps())
	require.NoError(tt, err)

	req := &spb.TriggerReportRequest{
		ChunkIds: []string{"chunk-1", "chunk-2"},
	}

	result := &s.TriggerReportResult{
		Scheduled: []t.ChunkID{"chunk-1"},
		Failed:    []t.ChunkID{"chunk-2"},
	}

	f.storage.EXPECT().
		StageAndReportMany(gomock.Any(), []t.ChunkID{"chunk-1", "chunk-2"}).
		Return(result)

	rsp, err := admin.TriggerReport(ctx, req)

	require.NoError(tt, err)
	require.Equal(tt, []string{"chunk-1"}, rsp.GetScheduled())
	require.Equal(tt, []string{"chunk-2"}, rsp.GetFailed())
}

func TestAdminServer_PauseHeartbeat(tt *testing.T) {
	ctx := context.Background()
	f := newAdminFixture(tt)

	admin, err := NewAdminServer(f.deps())
	require.NoError(tt, err)

	f.heartbeat.EXPECT().Pause().Return(true)
	f.heartbeat.EXPECT().IsPaused().Return(true)

	rsp, err := admin.PauseHeartbeat(ctx, &spb.HeartbeatControlRequest{})

	require.NoError(tt, err)
	require.Equal(tt, "paused", rsp.GetState().GetStatus())
}

func TestAdminServer_ResumeHeartbeat(tt *testing.T) {
	ctx := context.Background()
	f := newAdminFixture(tt)

	server, err := NewAdminServer(f.deps())
	require.NoError(tt, err)

	f.heartbeat.EXPECT().Resume().Return(true)
	f.heartbeat.EXPECT().IsPaused().Return(false)

	rsp, err := server.ResumeHeartbeat(ctx, &spb.HeartbeatControlRequest{})

	require.NoError(tt, err)
	require.Equal(tt, "running", rsp.GetState().GetStatus())
}

// fixture

type adminFixture struct {
	inventory *MockInventory
	storage   *MockStorage
	heartbeat *MockHeartbeat
}

func newAdminFixture(tt *testing.T) *adminFixture {
	ctrl := gomock.NewController(tt)
	return &adminFixture{
		inventory: NewMockInventory(ctrl),
		storage:   NewMockStorage(ctrl),
		heartbeat: NewMockHeartbeat(ctrl),
	}
}

func (f *adminFixture) deps() AdminDeps {
	return AdminDeps{
		Inventory: f.inventory,
		Storage:   f.storage,
		Heartbeat: f.heartbeat,
	}
}
