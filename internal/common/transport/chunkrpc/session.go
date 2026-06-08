package chunkrpc

import (
	"bytes"
	"context"
	pb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type Session struct {
	config  Config
	conn    *connect.ConnCache
	targets []t.NodeRef

	onProgress ProgressHandler
	progress   Progress
}

func (s *Session) Upload(ctx context.Context, chunk *t.Chunk) (t.NodeRef, error) {
	var errs []error
	others := make([]t.NodeRef, 0, len(s.targets))
	for i, target := range s.targets {
		others = others[:0]
		others = append(others, s.targets[:i]...)
		others = append(others, s.targets[i+1:]...)

		uploadCtx, cancel := context.WithTimeout(ctx, s.config.RPCTimeout())
		err := s.uploadToNode(uploadCtx, target, chunk, others)
		cancel()
		if err == nil {
			return target, nil
		}
		slog.WarnContext(ctx,
			"send chunk failed",
			"addr", target.Addr,
			"chunk", chunk.Meta.ID,
			"error", err,
		)
		errs = append(errs, fmt.Errorf("send chunk %s to %s failed: %w", chunk.Meta.ID, target.Addr, err))
	}
	return t.NodeRef{}, fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))
}

func (s *Session) uploadToNode(
	ctx context.Context, target t.NodeRef, chunk *t.Chunk, others []t.NodeRef,
) error {

	conn, err := s.conn.Get(target.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}

	client := spb.NewChunkServiceClient(conn)

	stream, err := client.PutChunk(ctx)
	if err != nil {
		return fmt.Errorf("open put stream: %w", err)
	}
	header := &spb.PutChunkHeader{
		NodeId:  string(target.ID),
		ChunkId: string(chunk.Meta.ID),
		Digest: &pb.Digest{
			Size:     int64(chunk.Meta.Digest.Size),
			Checksum: string(chunk.Meta.Digest.Checksum),
		},
	}

	s.progress = NewProgress(chunk.Meta, target)
	s.emitProgress()
	defer s.emitProgress()

	err = stream.Send(&spb.PutChunkRequest{Header: header})
	if err != nil {
		s.progress.Fail("send header failed")
		return fmt.Errorf("send header: %w", err)
	}

	err = s.uploadData(stream, chunk.Data)
	if err != nil && !errors.Is(err, io.EOF) {
		s.progress.Fail("send data failed")
		return fmt.Errorf("send data: %w", err)
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		s.progress.Fail("close stream failed")
		return fmt.Errorf("close stream: %w", err)
	}

	s.progress.Done()

	return nil
}

func (s *Session) uploadData(stream spb.ChunkService_PutChunkClient, data []byte) error {

	frames := utils.SplitFrames(data, s.config.FrameSize())
	for _, frame := range frames {

		err := stream.Send(&spb.PutChunkRequest{Data: frame})
		if err != nil {
			return err
		}

		s.progress.SentBytes += int64(len(frame))
		s.emitProgress()
	}
	return nil
}

func (s *Session) Download(ctx context.Context, chunkID t.ChunkID) (t.Chunk, error) {
	var errs []error
	for _, node := range s.targets {
		dlCtx, cancel := context.WithTimeout(ctx, s.config.RPCTimeout())
		chunk, err := s.downloadFromNode(dlCtx, node, chunkID)
		cancel()

		if err == nil {
			return chunk, nil
		}
		slog.WarnContext(ctx, "pull chunk failed", "addr", node.Addr, "chunk", chunkID, "error", err)
		errs = append(errs, fmt.Errorf("receive chunk %s from %s: %w", chunkID, node.Addr, err))
	}
	return t.Chunk{}, fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))
}

func (s *Session) downloadFromNode(
	ctx context.Context, nodeRef t.NodeRef, chunkID t.ChunkID,
) (t.Chunk, error) {

	conn, err := s.conn.Get(nodeRef.Addr)
	if err != nil {
		return t.Chunk{}, fmt.Errorf("get conn: %w", err)
	}

	client := spb.NewChunkServiceClient(conn)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := client.GetChunk(ctx, &spb.GetChunkRequest{
		NodeId:  string(nodeRef.ID),
		ChunkId: string(chunkID),
	})
	if err != nil {
		return t.Chunk{}, fmt.Errorf("send request: %w", err)
	}

	rsp, err := stream.Recv()
	if err != nil {
		return t.Chunk{}, fmt.Errorf("recv header: %w", err)
	}

	header := rsp.GetHeader()
	if header == nil {
		return t.Chunk{}, ErrInvalidHeader
	}

	headerMeta := convert.ChunkMetaFromPB(header)

	s.progress = NewProgress(headerMeta, nodeRef)
	s.emitProgress()
	defer s.emitProgress()

	chunk, err := s.downloadData(stream, chunkID)
	if err != nil {
		s.progress.Fail(err.Error())
		return t.Chunk{}, fmt.Errorf("recv data: %w", err)
	}

	if err := headerMeta.Match(chunk.Meta); err != nil {
		s.progress.Fail(err.Error())
		return t.Chunk{}, err
	}

	s.progress.Done()
	return chunk, nil
}

func (s *Session) downloadData(
	stream spb.ChunkService_GetChunkClient, chunkID t.ChunkID,
) (t.Chunk, error) {

	var buf bytes.Buffer

	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return t.Chunk{}, err
		}

		data := rsp.GetData()
		if data == nil {
			return t.Chunk{}, ErrInvalidData
		}

		buf.Write(rsp.Data)

		s.progress.SentBytes += int64(len(rsp.Data))
		s.emitProgress()
	}

	chunk := t.NewChunk(chunkID, buf.Bytes())
	return chunk, nil
}

func (s *Session) emitProgress() {
	if s.onProgress != nil {
		s.onProgress(s.progress)
	}
}
