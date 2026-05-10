package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func getCommitDir(rootDir string) (string, error) {
	commitDir := filepath.Join(rootDir, "chunks")
	if err := ensureDirExists(commitDir); err != nil {
		return "", err
	}
	return commitDir, nil
}

func getTempDir(rootDir string) (string, error) {
	tempDir := filepath.Join(rootDir, "temp")
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
