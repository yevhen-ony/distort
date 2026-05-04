package file 

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
	destPath string
	compare ChunkKeyComparer 
}

func NewFileObjectAssembler(destPath string, compare ChunkKeyComparer) (*FileObjectAssembler, error) {
	destDir := filepath.Dir(destPath)
	err := os.MkdirAll(destDir, 0o755)
	if err != nil {
		return nil, err
	}

	if compare == nil {
		return nil, errors.New("missing compare func")
	}
	return &FileObjectAssembler{destPath: destPath, compare: compare}, nil
}

func (a *FileObjectAssembler) NewWriter(obj t.ObjectDesc, chunks []t.ChunkDesc) (c.ObjectWriter, error) {

	slices.SortFunc(chunks, func(lhs, rhs t.ChunkDesc) int {
		return a.compare(lhs.ChunkKey, rhs.ChunkKey)
	})

	layout := make(map[t.ChunkID]region, len(chunks))
	offset := int64(0)
	for _, chunk := range chunks {
		layout[chunk.ChunkID] = region{offset: offset, size: chunk.ChunkSize}
		offset += chunk.ChunkSize
	}
	if offset != obj.TotalSize {
		return nil, c.ErrObjectSizeMismatch	
	}

	fd, err := os.OpenFile(a.destPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
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

	if err := w.fd.Sync(); err != nil {
		return err
	}
	return w.fd.Close()
}

type region struct {
	offset int64
	size int64
}
