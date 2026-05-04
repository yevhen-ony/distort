package file

import (
	"errors"
	"fmt"
	"io"
	"os"

	"dos/internal/common/config"
	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

type FileChunker struct {
	fd  *os.File
	cfg *FileChunkerConfig
	key int
}

type FileChunkerConfig struct {
	ChunkSize config.Size `yaml:"chunk_size"`
}

func NewFileChunker(path string, cfg *FileChunkerConfig) (*FileChunker, error) {
	if cfg == nil {
		return nil, errors.New("missing config") 
	}
	if cfg.ChunkSize <= 0 {
		return nil, c.ErrInvalidChunkSize
	}
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	chunker := &FileChunker{fd: fd, cfg: cfg}
	return chunker, nil
}

func (fc *FileChunker) Next() (t.ChunkKey, []byte, error) {
	chunkKey := t.ChunkKey(fmt.Sprintf("%06d", fc.key))
	fc.key++

	buf := make([]byte, fc.cfg.ChunkSize)
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
