package delivery

import (
	"context"
	"errors"
	"testing"

	chunkrpcmock "dos/internal/common/transport/chunkrpc/mock"
	t "dos/internal/common/types"
	"dos/internal/services/client/domain/progress"
	iomock "dos/internal/services/client/io/mock"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type downloadDeliveryMocks struct {
	ChunkT  *MockChunkTransport
	Session *chunkrpcmock.MockDownloadSession
	Sink    *iomock.MockObjectSink
}

func newDownloadDeliveryMocks(tt *testing.T) downloadDeliveryMocks {
	ctrl := gomock.NewController(tt)
	return downloadDeliveryMocks{
		ChunkT:  NewMockChunkTransport(ctrl),
		Session: chunkrpcmock.NewMockDownloadSession(ctrl),
		Sink:    iomock.NewMockObjectSink(ctrl),
	}
}

func TestObjectDelivery_Download_WritesToSink(tt *testing.T) {
	chunk := t.NewChunk("chunk-1", []byte("hello"))
	placement := testPlacement("chunk-1", "chunk-key-1")

	mocks := newDownloadDeliveryMocks(tt)

	mocks.ChunkT.EXPECT().
		NewDownloadSession(placement.Sources, gomock.Any()).
		Return(mocks.Session)

	mocks.Session.EXPECT().
		Download(gomock.Any(), placement.Meta.ID).
		Return(chunk, nil)

	mocks.Sink.EXPECT().
		WriteChunk(placement.Slot.ChunkKey, chunk.Data).
		Return(nil)

	d := newTestDelivery(mocks.ChunkT)
	err := d.downloadChunk(context.Background(), placement, mocks.Sink)

	require.NoError(tt, err)
}

func TestObjectDelivery_Download_OnDownloadError(tt *testing.T) {
	mocks := newDownloadDeliveryMocks(tt)

	placement := testPlacement("chunk-1", "chunk-key-1")
	expectedErr := errors.New("download failed")

	mocks.ChunkT.EXPECT().
		NewDownloadSession(placement.Sources, gomock.Any()).
		Return(mocks.Session)

	mocks.Session.EXPECT().
		Download(gomock.Any(), placement.Meta.ID).
		Return(t.Chunk{}, expectedErr)

	d := newTestDelivery(mocks.ChunkT)
	err := d.downloadChunk(context.Background(), placement, mocks.Sink)

	require.ErrorIs(tt, err, expectedErr)
	require.ErrorContains(tt, err, "download chunk chunk-1")
}

func TestObjectDelivery_Download_OnSinkWriteError(tt *testing.T) {
	mocks := newDownloadDeliveryMocks(tt)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	placement := testPlacement("chunk-1", "chunk-key-1")
	expectedErr := errors.New("sink write error")

	mocks.ChunkT.EXPECT().
		NewDownloadSession(placement.Sources, gomock.Any()).
		Return(mocks.Session)

	mocks.Session.EXPECT().
		Download(gomock.Any(), placement.Meta.ID).
		Return(chunk, nil)

	mocks.Sink.EXPECT().
		WriteChunk(placement.Slot.ChunkKey, chunk.Data).
		Return(expectedErr)

	d := newTestDelivery(mocks.ChunkT)
	err := d.downloadChunk(context.Background(), placement, mocks.Sink)

	require.ErrorIs(tt, err, expectedErr)
	require.ErrorContains(tt, err, "write chunk chunk-1")
}

func newTestDelivery(chunkT ChunkTransport) *ObjectDelivery {
	return &ObjectDelivery{
		objectID:   "object-1",
		chunkT:     chunkT,
		progress:   progress.NewObjectProgress("object-1"),
		onProgress: func(*progress.ObjectProgress) {},
	}
}

func testPlacement(chunkID t.ChunkID, chunkKey t.ChunkKey) t.ChunkPlacement {
	return t.ChunkPlacement{
		Meta: t.ChunkMeta{ID: chunkID},
		Slot: t.ObjectSlot{
			ObjectID: "object-1",
			ChunkKey: chunkKey,
		},
		Sources: []t.NodeRef{{ID: "node-1", Addr: "node-1:10000"}},
	}
}
