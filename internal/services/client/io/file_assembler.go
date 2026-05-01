package io

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"

	t "dos/internal/common/types"
	c "dos/internal/services/client"
)

type ChunkKeyComparer func(t.ChunkKey, t.ChunkKey) int

type FileObjectAssembler struct {
	baseDir string
	compare ChunkKeyComparer 
}

func NewFileObjectAssembler(baseDir string, compare ChunkKeyComparer) (*FileObjectAssembler, error) {

	err := os.MkdirAll(baseDir, 0o755)
	if err != nil {
		return nil, err
	}

	if compare == nil {
		return nil, errors.New("missing compare func")
	}
	return &FileObjectAssembler{baseDir: baseDir, compare: compare}, nil
}

func (a *FileObjectAssembler) NewWriter(obj c.ObjectInfo) (c.ObjectWriter, error) {
	slices.SortFunc(obj.Chunks, func(lhs, rhs c.ChunkInfo) int {
		return a.compare(lhs.Key, rhs.Key)
	})

	layout := make(map[t.ChunkID]region, len(obj.Chunks))
	offset := int64(0)
	for _, chunk := range obj.Chunks {
		layout[chunk.ID] = region{offset: offset, size: chunk.Size}
		offset += chunk.Size
	}
	if offset != obj.TotalSize {
		return nil, c.ErrObjectSizeMismatch	
	}

	path := filepath.Join(a.baseDir, string(obj.ID))

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	if err := fd.Truncate(obj.TotalSize); err != nil {
		_ = fd.Close()
		return nil, err
	}
	writer := &FileObjectWriter{fd: fd, layout: layout}

	return writer, nil
}

type FileObjectWriter struct {
	fd *os.File
	layout map[t.ChunkID]region
}

func (w *FileObjectWriter) WriteChunk(id t.ChunkID, data []byte) error {

	reg, ok := w.layout[id]	
	if !ok {
		return c.ErrChunkNotFound
	}
	if reg.size != int64(len(data)) {
		return c.ErrChunkSizeMismatch
	}

	n, err := w.fd.WriteAt(data, reg.offset)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	return nil 
}

func (w *FileObjectWriter) Close() error {
	return w.fd.Close()
}

type region struct {
	offset int64
	size int64
}
