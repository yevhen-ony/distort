package file

import (
	t "dos/internal/common/types"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
)

type ObjectChunker struct {
	file  *os.File
	size int
	err error
}

func NewObjectChunker(path string, chunkSize int64) (*ObjectChunker, error) {
	if chunkSize <= 0 {
		return nil, errors.New("invalid chunk size") 
	}
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	chunker := &ObjectChunker{file: fd, size: int(chunkSize)}
	return chunker, nil
}

func (oc *ObjectChunker) Chunks() iter.Seq2[t.ChunkKey, []byte] {

	return func(yield func(t.ChunkKey, []byte) bool) {
		key := 0	
		for {
			buf := make([]byte, oc.size)
			n, err := io.ReadFull(oc.file, buf)

			switch {
			case err == nil:
				if !yield(toChunkKey(key), buf) {
					return
				}
				key++

			case errors.Is(err, io.ErrUnexpectedEOF):
				if n > 0 {
					yield(toChunkKey(key), buf[:n])
				}
				return

			case errors.Is(err, io.EOF):
				return

			default:
				oc.err = err
				return
			}
		}
	}
}

func toChunkKey(key int) t.ChunkKey {
	return t.ChunkKey(fmt.Sprintf("%06d", key))
}

func (oc *ObjectChunker) Close() error {
	return oc.file.Close()
}

func (oc *ObjectChunker) Err() error {
	return oc.err
}
