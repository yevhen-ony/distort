package storage

import (
	"context"
	t "dos/internal/common/types"
	"io"
	"sync"
)

type UploadSession struct {
	id   t.ChunkID
	data []byte
	n    int

	once sync.Once
	onCommit func(context.Context, t.Chunk) error
	onAbort func() error
}

func NewUploadSession(chunkID t.ChunkID, size int64) *UploadSession {
	return &UploadSession{
		id:   chunkID,
		data: make([]byte, size),
	}
}

func (s *UploadSession) Write(p []byte) (int, error) {
	if s.n+len(p) > len(s.data) {
		return 0, io.ErrShortBuffer
	}

	n := copy(s.data[s.n:], p)
	if n != len(p) {
		return 0, io.ErrShortWrite
	}

	s.n += n
	return n, nil
}

func (s *UploadSession) Close() (err error) {
	s.once.Do(func() {
		if s.onAbort == nil {
			return
		}
		err = s.onAbort()
	})
	return err
}

func (s *UploadSession) Commit(ctx context.Context) (err error) {
	s.once.Do(func() {
		if s.onCommit == nil {
			return 
		}
		chunk := t.NewChunk(s.id, s.data[:s.n])
		err = s.onCommit(ctx, chunk)
	})
	return err
}
