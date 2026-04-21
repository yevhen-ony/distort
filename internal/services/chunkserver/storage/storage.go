package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dos/internal/libraries/digest"
	cs "dos/internal/services/chunkserver"
)

type FSChunkStorage struct {
	commitDir string
	tempDir string
}

func New(config *ChunkStorageConfig) (*FSChunkStorage, error) {
	commitDir, err := config.GetCommitDir()
	if err != nil {
		return nil, fmt.Errorf("get commit dir: %w", err)
	}

	tempDir, err := config.GetTempDir()
	if err != nil {
		return nil, fmt.Errorf("get temp dir: %w", err)
	}

	s := &FSChunkStorage{
		commitDir: commitDir,
		tempDir: tempDir,
	}
	return s, nil
}

func (s *FSChunkStorage) Get(chunkID cs.ChunkID) (io.ReadCloser, error) {
	chunkPath := filepath.Join(s.commitDir, string(chunkID))
	f, err := os.OpenFile(chunkPath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return f, nil 
}

func (s *FSChunkStorage) GetMeta(chunkID cs.ChunkID) (*cs.ChunkMeta, error) {
	chunkPath := filepath.Join(s.commitDir, string(chunkID))
	
	fd, err := os.Open(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("open chunk: %w", err)
	}
	defer fd.Close()
	
	dg := digest.New()
	if _, err := io.Copy(dg, fd); err != nil {
		return nil, fmt.Errorf("read chunk: %w", err)
	}

	fi, err := os.Stat(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("stat chunk: %w", err)
	}
	meta := &cs.ChunkMeta{
		Digest: dg.Digest(),
		ModifiedAt: fi.ModTime(),
	}
	
	return meta, nil
}

func (s *FSChunkStorage) GetAllIDs() ([]cs.ChunkID, error) {
	entries, err := os.ReadDir(s.commitDir)
	if err != nil {
		return nil, err 
	}
	chunks := []cs.ChunkID{}
	for _, e := range entries  {
		if e.IsDir() {
			continue
		}
		id := filepath.Base(e.Name())
		chunks = append(chunks, cs.ChunkID(id))
	}
	return chunks, nil
}

func (s *FSChunkStorage) NewWriter() (cs.ChunkWriter, error) {
	fd, err := os.CreateTemp(s.tempDir, "chunk-*")
	if err != nil {
		return nil, fmt.Errorf("create temp: %w", err)
	}

	w := &FSChunkWriter{
		fd: fd,
		commitDir: s.commitDir, 
		dg: digest.New(),
	}
	return w, nil
}

