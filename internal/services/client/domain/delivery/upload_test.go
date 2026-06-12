package delivery

import (
	"context"
	"errors"
	"iter"
	"testing"

	chunkrpcmock "dos/internal/common/transport/chunkrpc/mock"
	t "dos/internal/common/types"
	"dos/internal/services/client/domain/progress"
	"dos/internal/services/client/transport"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type uploadDeliveryMocks struct {
	MasterT *MockMasterTransport
	ChunkT  *MockChunkTransport
	Session *chunkrpcmock.MockUploadSession
}

func newUploadDeliveryMocks(tt *testing.T) uploadDeliveryMocks {
	ctrl := gomock.NewController(tt)

	return uploadDeliveryMocks{
		MasterT: NewMockMasterTransport(ctrl),
		ChunkT:  NewMockChunkTransport(ctrl),
		Session: chunkrpcmock.NewMockUploadSession(ctrl),
	}
}

func TestObjectDelivery_uploadChunk_AllocateSucceeded(tt *testing.T) {
	mocks := newUploadDeliveryMocks(tt)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	targets := []t.NodeRef{{ID: "node-1", Addr: "node-1:10000"}}

	mocks.MasterT.EXPECT().
		AllocateChunk(gomock.Any(), &transport.AllocateChunkCommand{
			Slot:      slot,
			ChunkSize: chunk.Meta.Digest.Size,
		}).
		Return(&t.ChunkAllocation{
			ID:      chunk.Meta.ID,
			Slot:    slot,
			Targets: targets,
		}, nil)

	mocks.ChunkT.EXPECT().
		NewUploadSession(targets, gomock.Any()).
		Return(mocks.Session)

	mocks.Session.EXPECT().
		Upload(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, cnk *t.Chunk) (t.NodeRef, error) {
			require.Equal(tt, chunk.Meta.ID, cnk.Meta.ID)
			require.Equal(tt, chunk.Data, cnk.Data)
			return targets[0], nil
		})

	d := newUploadObjectDelivery(slot.ObjectID, mocks)
	err := d.uploadChunk(context.Background(), slot.ChunkKey, chunk.Data)

	require.NoError(tt, err)
}

func TestObjectDelivery_uploadChunk_AllocateFailed(tt *testing.T) {
	mocks := newUploadDeliveryMocks(tt)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	expectedErr := errors.New("allocation failed")

	mocks.MasterT.EXPECT().
		AllocateChunk(gomock.Any(), &transport.AllocateChunkCommand{
			Slot:      slot,
			ChunkSize: chunk.Meta.Digest.Size,
		}).
		Return(nil, expectedErr)

	d := newUploadObjectDelivery(slot.ObjectID, mocks)
	err := d.uploadChunk(context.Background(), slot.ChunkKey, chunk.Data)

	require.ErrorIs(tt, err, expectedErr)
}

func TestObjectDelivery_uploadChunk_OnUploadError(tt *testing.T) {

	mocks := newUploadDeliveryMocks(tt)

	chunk := t.NewChunk("chunk-1", []byte("hello"))
	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	targets := []t.NodeRef{{ID: "node-1", Addr: "node-1:10000"}}

	expectedErr := errors.New("upload failed")

	mocks.MasterT.EXPECT().
		AllocateChunk(gomock.Any(), &transport.AllocateChunkCommand{
			Slot:      slot,
			ChunkSize: chunk.Meta.Digest.Size,
		}).
		Return(&t.ChunkAllocation{
			ID:      chunk.Meta.ID,
			Slot:    slot,
			Targets: targets,
		}, nil)

	mocks.ChunkT.EXPECT().
		NewUploadSession(targets, gomock.Any()).
		Return(mocks.Session)

	mocks.Session.EXPECT().
		Upload(gomock.Any(), gomock.Any()).
		Return(t.NodeRef{}, expectedErr)

	d := newUploadObjectDelivery(slot.ObjectID, mocks)
	err := d.uploadChunk(context.Background(), slot.ChunkKey, chunk.Data)

	require.ErrorIs(tt, err, expectedErr)
}

func TestObjectDelivery_Upload_CreatesObjectAndUploadsChunks(tt *testing.T) {
	mocks := newUploadDeliveryMocks(tt)

	slot := t.ObjectSlot{ObjectID: "object-1", ChunkKey: "chunk-key-1"}
	chunk := t.NewChunk("chunk-1", []byte("hello"))
	targets := []t.NodeRef{{ID: "node-1", Addr: "node-1:10000"}}

	mocks.MasterT.EXPECT().
		CreateObject(gomock.Any(), slot.ObjectID).
		Return(nil)

	mocks.MasterT.EXPECT().
		AllocateChunk(gomock.Any(), &transport.AllocateChunkCommand{
			Slot:      slot,
			ChunkSize: chunk.Meta.Digest.Size,
		}).
		Return(&t.ChunkAllocation{
			ID:      chunk.Meta.ID,
			Slot:    slot,
			Targets: targets,
		}, nil)

	mocks.ChunkT.EXPECT().
		NewUploadSession(targets, gomock.Any()).
		Return(mocks.Session)

	mocks.Session.EXPECT().
		Upload(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, got *t.Chunk) (t.NodeRef, error) {
			require.Equal(tt, chunk.Meta.ID, got.Meta.ID)
			require.Equal(tt, chunk.Data, got.Data)
			return targets[0], nil
		})

	d := newUploadObjectDelivery(slot.ObjectID, mocks)
	d.config = testDeliveryConfig{concurrency: 1}

	err := d.Upload(context.Background(), testChunkSource{
		chunks: []testChunk{{key: slot.ChunkKey, data: chunk.Data}},
	})

	require.NoError(tt, err)
}

func TestObjectDelivery_Upload_OnSourceError(tt *testing.T) {
	mocks := newUploadDeliveryMocks(tt)

	objectID := t.ObjectID("object-1")
	expectedErr := errors.New("read failed")

	mocks.MasterT.EXPECT().
		CreateObject(gomock.Any(), objectID).
		Return(nil)

	d := newUploadObjectDelivery(objectID, mocks)
	d.config = testDeliveryConfig{concurrency: 1}

	err := d.Upload(context.Background(), testChunkSource{
		err: expectedErr,
	})

	require.ErrorIs(tt, err, expectedErr)
	require.Equal(tt, progress.ObjectFailed, d.progress.Status)
}

func newUploadObjectDelivery(oid t.ObjectID, mocks uploadDeliveryMocks) *ObjectDelivery {
	return &ObjectDelivery{
		objectID:   oid,
		masterT:    mocks.MasterT,
		chunkT:     mocks.ChunkT,
		progress:   progress.NewObjectProgress(oid),
		onProgress: func(*progress.ObjectProgress) {},
	}
}

// test chunk source
type testChunkSource struct {
	chunks []testChunk
	err    error
}

type testChunk struct {
	key  t.ChunkKey
	data []byte
}

func (s testChunkSource) Chunks() iter.Seq2[t.ChunkKey, []byte] {
	return func(yield func(t.ChunkKey, []byte) bool) {
		for _, chunk := range s.chunks {
			if !yield(chunk.key, chunk.data) {
				return
			}
		}
	}
}

func (s testChunkSource) Err() error {
	return s.err
}

// test config
type testDeliveryConfig struct {
	concurrency int
}

func (c testDeliveryConfig) TransferConcurrency() int {
	return c.concurrency
}
