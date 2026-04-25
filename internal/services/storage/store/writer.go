package store 

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	s "dos/internal/services/storage"
	"dos/internal/common/digest"
)


type FSChunkWriter struct {
	fd *os.File
	dg *digest.Digester
	commitDir string
	closed bool
}

func (w *FSChunkWriter) Close() error {
	w.closed = true

	var errs []error
	var err error

	err = w.fd.Close()
	if err != nil && !errors.Is(err, os.ErrClosed) {
		errs = append(errs, fmt.Errorf("close: %w", err))
	}

	err = os.Remove(w.fd.Name())
	if err != nil && !errors.Is(err, os.ErrNotExist)  {
		errs = append(errs, fmt.Errorf("remove: %w", err))
	}
	return errors.Join(errs...)
}

func (w *FSChunkWriter) Commit(chunkID s.ChunkID) (time.Time, error) {
	if w.closed {
		return time.Time{}, errors.New("cannot commit closed writer")
	}

	if err := w.fd.Sync(); err != nil {
		return time.Time{}, fmt.Errorf("sync: %w", err)
	}
	
	if err := w.fd.Close(); err != nil {
		return time.Time{}, fmt.Errorf("close: %w", err)
	}	
	
	commitPath := filepath.Join(w.commitDir, string(chunkID))
	
	if err := os.Link(w.fd.Name(), commitPath); err != nil {
		return time.Time{}, fmt.Errorf("create link: %w", err)
	}
	
	if err := w.sync(w.commitDir); err != nil {
		return time.Time{}, fmt.Errorf("sync dir: %w", err)
	}

	fi, err := os.Stat(commitPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("stat chunk: %w", err)
	}
	return fi.ModTime(), nil
}

func (w *FSChunkWriter) Write(data []byte) (int, error) {
	n, err := w.fd.Write(data)
	if n > 0 {
		w.dg.Write(data[:n])
	}
	if err != nil {
		return n, fmt.Errorf("write: %w", err)
	}
	if n != len(data) {
		return n, errors.New("partial write")
	}

	return n, nil 
}

func (w *FSChunkWriter) Digest() digest.Digest {
	return w.dg.Digest()
}

func (w *FSChunkWriter) sync(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	if err := f.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}
	return nil
}
