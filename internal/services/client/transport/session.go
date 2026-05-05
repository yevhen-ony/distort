package transport

import (
	"bytes"
	"context"
	pb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	c "dos/internal/services/client"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type ChunkTransferSession struct {
	config     *StorageTransportConfig
	conn       *connect.ConnCache
	nodes      []t.NodeRef

	onProgress ChunkProgressHandler
	progress   ChunkProgress
}

func (s *ChunkTransferSession) Upload(ctx context.Context, chunk *c.Chunk) error {
	var errs []error
	for _, node := range s.nodes {
		err := s.uploadToNode(ctx, node, chunk)
		if err == nil {
			return nil
		}
		slog.WarnContext(ctx, "send chunk failed", "addr", node.Addr, "chunk", chunk.Meta.ID, "error", err)
		errs = append(errs, fmt.Errorf("send chunk %s to %s failed: %w", chunk.Meta.ID, node.Addr, err))
	}
	return fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))
}


func (s *ChunkTransferSession) uploadToNode(ctx context.Context, nodeRef t.NodeRef, chunk *c.Chunk) error {
	conn, err := s.conn.Get(nodeRef.Addr)
	if err != nil {
		return fmt.Errorf("get conn: %w", err)
	}

	client := spb.NewChunkServiceClient(conn)

	stream, err := client.PutChunk(ctx)
	if err != nil {
		return fmt.Errorf("open put stream: %w", err)
	}
	header := &spb.PutChunkHeader{
		NodeId:  string(nodeRef.ID),
		ChunkId: string(chunk.Meta.ID),
		Digest: &pb.Digest{
			Size:     int64(chunk.Meta.Digest.Size),
			Checksum: string(chunk.Meta.Digest.Checksum),
		},
	}

	s.progress = NewChunkProgress(chunk.Meta, nodeRef)
	s.emitProgress()

	err = stream.Send(&spb.PutChunkRequest{Header: header})
	if err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	if err = s.uploadData(stream, chunk.Data); err != nil {
		return fmt.Errorf("send data: %w", err)
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("close stream: %w", err)
	}

	s.progress.Done = true
	s.emitProgress()

	return nil
}

func (s *ChunkTransferSession) uploadData(stream spb.ChunkService_PutChunkClient, data []byte) error {
	for len(data) > 0 {
		n := min(int64(s.config.FrameSize), int64(len(data)))
		err := stream.Send(&spb.PutChunkRequest{Data: data[:n]})
		if err != nil {
			return err
		}
		
		s.progress.SentBytes += n
		s.emitProgress()

		data = data[n:]
	}
	return nil
}

func (s *ChunkTransferSession) Download(ctx context.Context, chunkID t.ChunkID) (c.Chunk, error) {
	var errs []error
	for _, node := range s.nodes {
		chunk, err := s.downloadFromNode(ctx, node, chunkID)
		if err == nil {
			return chunk, nil
		}
		slog.WarnContext(ctx, "pull chunk failed", "addr", node.Addr, "chunk", chunkID, "error", err)
		errs = append(errs, fmt.Errorf("send chunk %s to %s: %w", chunkID, node.Addr, err))
	}
	return c.Chunk{}, fmt.Errorf("all candidate nodes failed: %w", errors.Join(errs...))

}

func (s *ChunkTransferSession) downloadFromNode(
	ctx context.Context, nodeRef t.NodeRef, chunkID t.ChunkID,
) (c.Chunk, error) {

	conn, err := s.conn.Get(nodeRef.Addr)
	if err != nil {
		return c.Chunk{}, fmt.Errorf("get conn: %w", err)
	}

	client := spb.NewChunkServiceClient(conn)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stream, err := client.GetChunk(ctx, &spb.GetChunkRequest{
		NodeId:  string(nodeRef.ID),
		ChunkId: string(chunkID),
	})
	if err != nil {
		return c.Chunk{}, fmt.Errorf("send request: %w", err)
	}

	rsp, err := stream.Recv()
	if err != nil {
		return c.Chunk{}, fmt.Errorf("recv header: %w", err)
	}

	header := rsp.GetHeader()
	if header == nil {
		return c.Chunk{}, ErrHeaderInvalid
	}

	headerMeta := convert.ChunkMetaFromPB(header)

	s.progress = NewChunkProgress(headerMeta, nodeRef)
	s.emitProgress()

	data, digest, err := s.downloadData(stream)
	if err != nil {
		return c.Chunk{}, fmt.Errorf("recv data: %w", err)
	}

	meta := t.ChunkMeta{
		ID:     chunkID,
		Digest: digest,
	}

	err = matchChunkMeta(headerMeta, meta)
	if err != nil {
		return c.Chunk{}, err
	}

	chunk := c.Chunk{
		Meta: meta,
		Data: data,
	}

	s.progress.Done = true
	s.emitProgress()

	return chunk, nil
}

func (s *ChunkTransferSession) downloadData(
	stream spb.ChunkService_GetChunkClient,
) ([]byte, *digest.Digest, error) {

	var buf bytes.Buffer
	dg := digest.New()

	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		data := rsp.GetData()
		if data == nil {
			return nil, nil, ErrDataInvalid
		}

		buf.Write(rsp.Data)
		dg.Write(rsp.Data)

		s.progress.SentBytes += int64(len(rsp.Data))
		s.emitProgress()
	}
	return buf.Bytes(), dg.Digest(), nil
}

func (s *ChunkTransferSession) emitProgress() {
	if s.onProgress != nil {
		s.onProgress(s.progress)
	}
}

func matchChunkMeta(want, got t.ChunkMeta) error {
	if err := got.Digest.Match(want.Digest); err != nil {
		return err
	}

	if want.ID != got.ID {
		return fmt.Errorf("id mismatch: %w", ErrChunkMetaMismatch)
	}
	return nil
}
