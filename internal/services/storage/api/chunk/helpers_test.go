package chunk

import (
	"context"
	"io"
	"testing"

	spb "dos/gen/proto/storage/v1"

	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/metadata"
)

// fixture

type chunkFixture struct {
	identity *MockNodeIdentity
	storage  *MockChunkStorage
	config   *MockChunkConfig
}

func newChunkFixture(tt *testing.T) *chunkFixture {
	ctrl := gomock.NewController(tt)
	return &chunkFixture{
		identity: NewMockNodeIdentity(ctrl),
		storage:  NewMockChunkStorage(ctrl),
		config:   NewMockChunkConfig(ctrl),
	}
}

func (f *chunkFixture) deps() ChunkDeps {
	return ChunkDeps{
		Identity: f.identity,
		Storage:  f.storage,
		Config:   f.config,
	}
}

// fake get chunk stream

type fakeGetChunkStream struct {
	ctx  context.Context
	sent []*spb.GetChunkResponse
}

func (s *fakeGetChunkStream) Send(rsp *spb.GetChunkResponse) error {
	s.sent = append(s.sent, rsp)
	return nil
}

func (s *fakeGetChunkStream) Context() context.Context     { return s.ctx }
func (s *fakeGetChunkStream) SendHeader(metadata.MD) error { return nil }
func (s *fakeGetChunkStream) SetHeader(metadata.MD) error  { return nil }
func (s *fakeGetChunkStream) SetTrailer(metadata.MD)       {}
func (s *fakeGetChunkStream) SendMsg(any) error            { return nil }
func (s *fakeGetChunkStream) RecvMsg(any) error            { return nil }

// fake put chunk stream

type fakePutChunkStream struct {
	ctx      context.Context
	recv     []*spb.PutChunkRequest
	response *spb.PutChunkResponse
}

func (s *fakePutChunkStream) Recv() (*spb.PutChunkRequest, error) {
	if len(s.recv) == 0 {
		return nil, io.EOF
	}
	req := s.recv[0]
	s.recv = s.recv[1:]
	return req, nil
}

func (s *fakePutChunkStream) SendAndClose(rsp *spb.PutChunkResponse) error {
	s.response = rsp
	return nil
}

func (s *fakePutChunkStream) Context() context.Context     { return s.ctx }
func (s *fakePutChunkStream) SetHeader(metadata.MD) error  { return nil }
func (s *fakePutChunkStream) SendHeader(metadata.MD) error { return nil }
func (s *fakePutChunkStream) SetTrailer(metadata.MD)       {}
func (s *fakePutChunkStream) SendMsg(any) error            { return nil }
func (s *fakePutChunkStream) RecvMsg(any) error            { return nil }

// fake upload session

type fakeUploadSession struct {
	written []byte
	commits int
	closes  int
}

func (s *fakeUploadSession) Write(p []byte) (int, error) {
	s.written = append(s.written, p...)
	return len(p), nil
}

func (s *fakeUploadSession) Commit(context.Context) error {
	s.commits++
	return nil
}

func (s *fakeUploadSession) Close() error {
	s.closes++
	return nil
}
