package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dos/internal/common/digest"
	t "dos/internal/common/types"
)

type FSChunkStorage struct {
	commitDir string
}

type ChunkStorageConfig interface {
	StorageRootDir() string
}

func NewChunkStorage(config ChunkStorageConfig) (*FSChunkStorage, error) {

	commitDir, err := getCommitDir(config.StorageRootDir())
	if err != nil {
		return nil, fmt.Errorf("get commit dir: %w", err)
	}

	s := &FSChunkStorage{
		commitDir: commitDir,
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

func (stg *FSChunkStorage) Store(chunk t.Chunk) (err error) {
	
	chunkPath := filepath.Join(stg.commitDir, string(chunk.Meta.ID))
	defer func() {
		if err != nil {
			_ = os.Remove(chunkPath)	
		}
	}()

	f, err := os.OpenFile(chunkPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("new chunk file: %w", err)
	}
	defer func() {
		syncErr := f.Sync()
		closeErr := f.Close() 
  		if err == nil {
  			if syncErr != nil {
  				err = fmt.Errorf("sync chunk file: %w", syncErr)
  			} else if closeErr != nil {
  				err = fmt.Errorf("close chunk file: %w", closeErr)
  			}
  		}
  	}()

	n, err := f.Write(chunk.Data)
	if err != nil {
		return fmt.Errorf("write chunk file: %w", err) 
	}
	if n != len(chunk.Data) {
		return fmt.Errorf("write chunk file: %w", io.ErrShortWrite)
	}

	return nil
}

func getCommitDir(rootDir string) (string, error) {
	commitDir := filepath.Join(rootDir, "chunks")
	if err := ensureDirExists(commitDir); err != nil {
		return "", err
	}
	return commitDir, nil
}

func ensureDirExists(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create dir: %w", err)
	}
	return nil
}
