package chunk

import (
	"bytes"
	"context"
	"testing"
	"time"

	cpb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestChunkServer_DeleteChunk(tt *testing.T) {
	ctx := context.Background()
	f := newChunkFixture(tt)

	api, err := NewChunkServer(f.deps())
	require.NoError(tt, err)

	nodeID := t.NodeID("node-1")
	chunkID := t.ChunkID("chunk-1")

	f.identity.EXPECT().Validate(nodeID).Return(nil)
	f.storage.EXPECT().DeleteChunk(gomock.Any(), chunkID).Return(nil)

	rsp, err := api.DeleteChunk(ctx, &spb.DeleteChunkRequest{
		NodeId:  string(nodeID),
		ChunkId: string(chunkID),
	})

	require.NoError(tt, err)
	require.NotNil(tt, rsp)
}

func TestChunkServer_ReplicateChunk(tt *testing.T) {
	ctx := context.Background()
	f := newChunkFixture(tt)

	api, err := NewChunkServer(f.deps())
	require.NoError(tt, err)

	chunkID := t.ChunkID("chunk-1")
	self := t.NodeID("node-1")
	targets := []t.NodeRef{
		{ID: "node-2", Addr: "addr-2"},
		{ID: "node-3", Addr: "addr-3"},
	}

	req := &spb.ReplicateChunkRequest{
		NodeId:  string(self),
		ChunkId: string(chunkID),
		Targets: utils.Map(targets, convert.NodeRefToPB),
	}

	f.identity.EXPECT().Validate(self).Return(nil)
	f.storage.EXPECT().
		ScheduleForwardChunk(gomock.Any(), chunkID, targets).
		Return(nil)

	rsp, err := api.ReplicateChunk(ctx, req)

	require.NoError(tt, err)
	require.NotNil(tt, rsp)
}

func TestChunkServer_GetChunk_StreamsHeaderAndDataFrames(tt *testing.T) {
	ctx := context.Background()
	f := newChunkFixture(tt)

	api, err := NewChunkServer(f.deps())
	require.NoError(tt, err)

	self := t.NodeID("node-1")
	chunk := t.NewChunk("chunk-1", []byte("hello"))
	released := false
	release := func() { released = true }

	f.storage.EXPECT().
		AcquireOpSlot(gomock.Any(), time.Second).
		Return(release, nil)

	f.identity.EXPECT().
		Validate(self).
		Return(nil)

	f.storage.EXPECT().
		LoadChunk(chunk.Meta.ID).
		Return(chunk, nil)

	f.config.EXPECT().FrameSize().Return(int64(2))

	stream := &fakeGetChunkStream{ctx: ctx}
	req := &spb.GetChunkRequest{
		NodeId:  string(self),
		ChunkId: string(chunk.Meta.ID),
	}

	err = api.GetChunk(req, stream)

	require.NoError(tt, err)
	require.True(tt, released)
	require.Len(tt, stream.sent, 4)

	require.Equal(tt, "chunk-1", stream.sent[0].GetHeader().GetChunkId())
	require.Equal(tt, chunk.Meta.Digest.Size, stream.sent[0].GetHeader().GetDigest().GetSize())

	frames := utils.Map(stream.sent[1:], func(res *spb.GetChunkResponse) []byte {
		return res.GetData()
	})
	require.Equal(tt, chunk.Data, bytes.Join(frames, nil))
}

func TestChunkServer_PutChunk_WritesFramesAndCommitsUpload(tt *testing.T) {
	ctx := context.Background()
	f := newChunkFixture(tt)
	session := &fakeUploadSession{}

	api, err := NewChunkServer(f.deps())
	require.NoError(tt, err)

	self := t.NodeID("node-1")
	chunk := t.NewChunk("chunk-1", []byte("hello"))

	f.identity.EXPECT().Validate(self).Return(nil)
	f.storage.EXPECT().
		StartUpload(gomock.Any(), &chunk.Meta).
		Return(session, nil)

	header := &spb.PutChunkHeader{
		NodeId:  string(self),
		ChunkId: string(chunk.Meta.ID),
		Digest: &cpb.Digest{
			Checksum: string(chunk.Meta.Digest.Checksum),
			Size:     chunk.Meta.Digest.Size,
		},
	}
	stream := &fakePutChunkStream{
		ctx: ctx,
		recv: []*spb.PutChunkRequest{
			{Header: header},
			{Data: []byte("he")},
			{Data: []byte("llo")},
		},
	}

	err = api.PutChunk(stream)

	require.NoError(tt, err)
	require.Equal(tt, chunk.Data, session.written)
	require.Equal(tt, 1, session.commits)
	require.Equal(tt, 1, session.closes)
	require.NotNil(tt, stream.response)
}
