package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ChunkStorageConfig struct {
	RootDir string `yaml:"root_dir"`
}

func (c *ChunkStorageConfig) GetCommitDir() (string, error) {
	commitDir := filepath.Join(c.RootDir, "chunks")
	if err := ensureDirExists(commitDir); err != nil {
		return "", err
	}
	return commitDir, nil
}

func (c *ChunkStorageConfig) GetTempDir() (string, error) {
	tempDir := filepath.Join(c.RootDir, "temp")
	if err := ensureDirExists(tempDir); err != nil {
		return "", err
	}
	return tempDir, nil
}

func ensureDirExists(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create dir: %w", err)
	}
	return nil
}
