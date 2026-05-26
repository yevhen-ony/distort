package file

import (
	"errors"
	"io"
	"os"
)

var (
	ErrChunkSizeMismatch = errors.New("chunk size mismatch")
)

type ObjectWriter struct {
	file *os.File
}

func NewObjectWriter(path string, totalBytes int64) (*ObjectWriter, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
  	if err != nil {
  		return nil, err
  	}
  	if err := file.Truncate(totalBytes); err != nil {
		_ = file.Close()
  		return nil, err
  	}
	writer := &ObjectWriter{
		file: file,
	}
	return writer, nil
}

func (w *ObjectWriter) WriteRegion(region ChunkRegion, data []byte) error {

	if region.Size != int64(len(data)) {
		return ErrChunkSizeMismatch
	}

	n, err := w.file.WriteAt(data, region.Offset)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil 
}

func (w *ObjectWriter) Close() error {

	if err := w.file.Sync(); err != nil {
		return err
	}
	return w.file.Close()
}
