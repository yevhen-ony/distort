package service

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cs "dos/internal/chunkserver"
	"dos/internal/chunkserver/storage"
)

func checksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func newTestService(t *testing.T) (*Service, *storage.FSChunkStorage) {
  	t.Helper()

	cfg := &storage.ChunkStorageConfig{RootDir: t.TempDir()}
  	store, err := storage.New(cfg)
  	require.NoError(t, err)

  	svc := &Service{
  		store:   store,
  		catalog: make(cs.ChunkCatalog),
  	}
  	return svc, store
}

func TestService_CommitUploadSession(t *testing.T) {
	svc, st := newTestService(t)
	payload := []byte("hello")
	dg := storage.NewDigester()
	dg.Write(payload)

	info := &cs.ChunkInfo{
		ID: "chunk-1",
		ChunkDigest: dg.Digest(),
	}

	t.Run("HappyPath", func(t *testing.T) {
		writer, err := svc.StartUploadSession(info)
		require.NoError(t, err)

		_, err = writer.Write(payload)
		require.NoError(t, err)

		require.NoError(t, svc.CommitUploadSession(writer, info))

		meta, ok := svc.catalog[info.ID]
		require.True(t, ok)
		assert.Equal(t, info.Size, meta.Size)
		assert.Equal(t, info.Checksum, meta.Checksum)

		r, err := st.Get(info.ID)
		require.NoError(t, err)
		defer r.Close()

		got, err := io.ReadAll(r)
		require.NoError(t, err)
		assert.Equal(t, payload, got)
	})

  	t.Run("CollitionOnStart", func(t *testing.T) {
		_, err := svc.StartUploadSession(info)
		require.Error(t, err, "chunk id collision")
	})

  	t.Run("CollitionOnCommit", func(t *testing.T) {
		payload := []byte("hello")
		dg := storage.NewDigester()
		dg.Write(payload)

		info := &cs.ChunkInfo{
			ID: "chunk-2",
			ChunkDigest: dg.Digest(),
		}

		session, err := svc.StartUploadSession(info)
		require.NoError(t, err)

		_, err = session.Write(payload)
		require.NoError(t, err)

  		// Simulate race: ID becomes taken after session start but before commit.
		svc.catalog[info.ID] = cs.ChunkMeta{
			ChunkDigest: cs.ChunkDigest{Size: int64(999), Checksum: "existing"},
		}

		err = svc.CommitUploadSession(session, info)
		require.Error(t, err)

		// Existing entry must stay untouched.
		meta := svc.catalog[info.ID]
		assert.Equal(t, int64(999), meta.Size)
		assert.Equal(t, "existing", meta.Checksum)
	})
}
