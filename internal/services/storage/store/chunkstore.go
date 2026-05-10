package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
)

type FSChunkStorage struct {
	commitDir string
	tempDir   string
}

type ChunkStorageConfig interface {
	StorageRootDir() string
}

func NewChunkStorage(config ChunkStorageConfig) (*FSChunkStorage, error) {

	commitDir, err := getCommitDir(config.StorageRootDir())
	if err != nil {
		return nil, fmt.Errorf("get commit dir: %w", err)
	}

	tempDir, err := getTempDir(config.StorageRootDir())
	if err != nil {
		return nil, fmt.Errorf("get temp dir: %w", err)
	}

	s := &FSChunkStorage{
		commitDir: commitDir,
		tempDir:   tempDir,
	}
	return s, nil
}

func (stg *FSChunkStorage) Get(chunkID t.ChunkID) (io.ReadCloser, error) {

	chunkPath := filepath.Join(stg.commitDir, string(chunkID))
	f, err := os.OpenFile(chunkPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (stg *FSChunkStorage) Delete(chunkID t.ChunkID) error {
	chunkPath := filepath.Join(stg.commitDir, string(chunkID))
	err := os.Remove(chunkPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (stg *FSChunkStorage) GetMeta(chunkID t.ChunkID) (t.ChunkMeta, error) {

	chunkPath := filepath.Join(stg.commitDir, string(chunkID))

	fd, err := os.Open(chunkPath)
	if err != nil {
		return t.ChunkMeta{}, fmt.Errorf("open chunk: %w", err)
	}
	defer fd.Close()

	dg := digest.New()
	if _, err := io.Copy(dg, fd); err != nil {
		return t.ChunkMeta{}, fmt.Errorf("read chunk: %w", err)
	}

	meta := t.ChunkMeta{
		ID:     chunkID,
		Digest: dg.Digest(),
	}

	return meta, nil
}

func (stg *FSChunkStorage) List() ([]t.ChunkID, error) {
	entries, err := os.ReadDir(stg.commitDir)
	if err != nil {
		return nil, err
	}
	chunks := []t.ChunkID{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		id := filepath.Base(e.Name())
		chunks = append(chunks, t.ChunkID(id))
	}
	return chunks, nil
}

func (stg *FSChunkStorage) NewWriter() (s.ChunkWriter, error) {

	fd, err := os.CreateTemp(stg.tempDir, "chunk-*")
	if err != nil {
		return nil, fmt.Errorf("create temp: %w", err)
	}

	w := &FSChunkWriter{
		fd:        fd,
		commitDir: stg.commitDir,
		dg:        digest.New(),
	}
	return w, nil
}
