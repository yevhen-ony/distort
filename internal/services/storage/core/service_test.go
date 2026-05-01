package core 

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	t "dos/internal/common/types"
	"dos/internal/common/digest"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/store"
)

func checksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func newTestService(test *testing.T) (*Service, *store.FSChunkStorage) {
  	test.Helper()

	cfg := &store.ChunkStorageConfig{RootDir: test.TempDir()}
  	store, err := store.New(cfg)
  	require.NoError(test, err)

  	svc := &Service{
  		store:   store,
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

		require.NoError(test, svc.CommitUploadSession(writer, desc))

		meta, ok := svc.catalog[desc.ID]
		require.True(test, ok)
		assert.Equal(test, desc.Digest.Size, meta.Digest.Size)
		assert.Equal(test, desc.Digest.Checksum, meta.Digest.Checksum)

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
		svc.catalog[desc.ID] = t.ChunkMeta{}

		err = svc.CommitUploadSession(session, desc)
		require.Error(test, err)

		// Existing entry must stay untouched.
		meta := svc.catalog[desc.ID]
		assert.Equal(test, int64(0), meta.Digest.Size)
		assert.Equal(test, "", string(meta.Digest.Checksum))
	})
}
