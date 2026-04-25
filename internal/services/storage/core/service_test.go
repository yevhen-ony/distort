package core 

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "dos/internal/services/storage"
	"dos/internal/services/storage/store"
	"dos/internal/common/digest"
)

func checksumHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func newTestService(t *testing.T) (*Service, *store.FSChunkStorage) {
  	t.Helper()

	cfg := &store.ChunkStorageConfig{RootDir: t.TempDir()}
  	store, err := store.New(cfg)
  	require.NoError(t, err)

  	svc := &Service{
  		store:   store,
  		catalog: make(s.ChunkCatalog),
  	}
  	return svc, store
}

func TestService_CommitUploadSession(t *testing.T) {
	svc, st := newTestService(t)
	payload := []byte("hello")
	dg := digest.New()
	dg.Write(payload)

	info := &s.ChunkInfo{
		ID: "chunk-1",
		Digest: dg.Digest(),
	}

	t.Run("HappyPath", func(t *testing.T) {
		writer, err := svc.StartUploadSession(info)
		require.NoError(t, err)

		_, err = writer.Write(payload)
		require.NoError(t, err)

		require.NoError(t, svc.CommitUploadSession(writer, info))

		meta, ok := svc.catalog[info.ID]
		require.True(t, ok)
		assert.Equal(t, info.Digest.Size, meta.Digest.Size)
		assert.Equal(t, info.Digest.Checksum, meta.Digest.Checksum)

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
		dg := digest.New()
		dg.Write(payload)

		info := &s.ChunkInfo{
			ID: "chunk-2",
			Digest: dg.Digest(),
		}

		session, err := svc.StartUploadSession(info)
		require.NoError(t, err)

		_, err = session.Write(payload)
		require.NoError(t, err)

  		// Simulate race: ID becomes taken after session start but before commit.
		svc.catalog[info.ID] = s.ChunkMeta{
			Digest: digest.Digest{Size: int64(999), Checksum: "existing"},
		}

		err = svc.CommitUploadSession(session, info)
		require.Error(t, err)

		// Existing entry must stay untouched.
		meta := svc.catalog[info.ID]
		assert.Equal(t, int64(999), meta.Digest.Size)
		assert.Equal(t, "existing", meta.Digest.Checksum)
	})
}
