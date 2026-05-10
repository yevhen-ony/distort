package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/store"
)

func checksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func newTestService(test *testing.T) (*StorageService, *store.FSChunkStorage) {
  	test.Helper()

	cfg := &store.ChunkStorageConfig{RootDir: test.TempDir()}
  	store, err := store.New(cfg)
  	require.NoError(test, err)

  	svc := &StorageService{
  		diskStore:   store,
  		catalog: make(s.ChunkCatalog),
  	}
  	return svc, store
}

func TestService_CommitUploadSession(test *testing.T) {
	svc, st := newTestService(test)
	payload := []byte("hello")
	dg := digest.New()
	dg.Write(payload)

	desc := &t.ChunkMeta{
		ID: "chunk-1",
		Digest: dg.Digest(),
	}

	test.Run("HappyPath", func(test *testing.T) {
		writer, err := svc.StartUploadSession(desc)
		require.NoError(test, err)

		_, err = writer.Write(payload)
		require.NoError(test, err)

		require.NoError(test, svc.CommitUploadSession(context.Background(), writer, desc))

		record, ok := svc.catalog[desc.ID]
		require.True(test, ok)
		assert.NoError(test, desc.Digest.Match(record.Meta.Digest))

		r, err := st.Get(desc.ID)
		require.NoError(test, err)
		defer r.Close()

		got, err := io.ReadAll(r)
		require.NoError(test, err)
		assert.Equal(test, payload, got)
	})

  	test.Run("CollitionOnStart", func(test *testing.T) {
		_, err := svc.StartUploadSession(desc)
		require.Error(test, err, "chunk id collision")
	})

  	test.Run("CollitionOnCommit", func(test *testing.T) {
		payload := []byte("hello")
		dg := digest.New()
		dg.Write(payload)

		desc := &t.ChunkMeta{
			ID: "chunk-2",
			Digest: dg.Digest(),
		}

		session, err := svc.StartUploadSession(desc)
		require.NoError(test, err)

		_, err = session.Write(payload)
		require.NoError(test, err)

  		// Simulate race: ID becomes taken after session start but before commit.
		svc.catalog[desc.ID] = &s.ChunkRecord{}

		err = svc.CommitUploadSession(context.Background(), session, desc)
		require.Error(test, err)

		// Existing entry must stay untouched.
		record := svc.catalog[desc.ID]
		assert.Equal(test, int64(0), record.Meta.Digest.Size)
		assert.Equal(test, "", string(record.Meta.Digest.Checksum))
	})
}
