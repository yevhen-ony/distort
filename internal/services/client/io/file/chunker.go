package file 

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

type FileChunker struct {
	fd        *os.File
	chunkSize int64
	objectID  t.ObjectID
	key       int
}

func NewFileChunker(path string, chunkSize int64) (*FileChunker, error) {
	if chunkSize <= 0 {
		return nil, c.ErrInvalidChunkSize 
	}
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	chunker := &FileChunker{
		fd:        fd,
		chunkSize: chunkSize,
		objectID:  t.ObjectID(name),
	}
	return chunker, nil
}

func (fc *FileChunker) GetObjectID() t.ObjectID {
	return fc.objectID
}

func (fc *FileChunker) Next() (t.ChunkKey, []byte, error) {
	chunkKey := t.ChunkKey(fmt.Sprintf("%06d", fc.key))
	fc.key++

	buf := make([]byte, fc.chunkSize)
	n, err := io.ReadFull(fc.fd, buf)

	switch {
	case err == nil:
		return chunkKey, buf, nil

	case errors.Is(err, io.ErrUnexpectedEOF):
		return chunkKey, buf[:n], nil
	
	case errors.Is(err, io.EOF):
		return "", nil, io.EOF

	default:
		return "", nil, err
	}
}

func (fc *FileChunker) Close() error {
	return fc.fd.Close()
}
