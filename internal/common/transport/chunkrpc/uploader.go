package chunkrpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
)

type ChunkUploader struct {
	conn   *connect.ConnCache
	config Config

	progress   Progress
	onProgress func(Progress)
}

func (cu *ChunkUploader) Upload(
	ctx context.Context,
	target t.NodeRef,
	chunk *t.Chunk,
	others []t.NodeRef,
) error {

	stream, err := cu.startUpload(ctx, target.Addr)
	if err != nil {
		return err
	}

	cu.progress = NewProgress(*chunk.Meta.Clone(), target)
	cu.emitProgress()
	defer cu.emitProgress()

	header := cu.buildHeader(target, &chunk.Meta, others)
	err = cu.uploadHeader(stream, header)
	if err != nil {
		return err
	}

	err = cu.uploadData(stream, chunk.Data)
	if err != nil && !errors.Is(err, io.EOF) {
		cu.progress.Fail("send data failed")
		return fmt.Errorf("send data: %w", err)
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		cu.progress.Fail("close stream failed")
		return fmt.Errorf("close stream: %w", err)
	}

	cu.progress.Done()

	return nil
}

func (cu *ChunkUploader) uploadData(stream spb.ChunkService_PutChunkClient, data []byte) error {

	frames := utils.SplitFrames(data, cu.config.FrameSize())
	for _, frame := range frames {

		err := stream.Send(&spb.PutChunkRequest{Data: frame})
		if err != nil {
			return fmt.Errorf("send frame: %w", err)
		}

		cu.progress.SentBytes += int64(len(frame))
		cu.emitProgress()
	}
	return nil
}

func (cu *ChunkUploader) startUpload(ctx context.Context, addr string) (spb.ChunkService_PutChunkClient, error) {
	conn, err := cu.conn.Get(addr)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	client := spb.NewChunkServiceClient(conn)
	stream, err := client.PutChunk(ctx)
	if err != nil {
		return nil, fmt.Errorf("put chunk rpc: %w", err)
	}
	return stream, nil
}

func (cu *ChunkUploader) uploadHeader(
	stream spb.ChunkService_PutChunkClient,
	header *spb.PutChunkHeader,
) error {
	err := stream.Send(&spb.PutChunkRequest{Header: header})
	if err != nil {
		cu.progress.Fail("send header failed")
		return fmt.Errorf("send header: %w", err)
	}
	return nil
}

func (cu *ChunkUploader) buildHeader(
	target t.NodeRef,
	meta *t.ChunkMeta,
	others []t.NodeRef,
) *spb.PutChunkHeader {
	return &spb.PutChunkHeader{
		NodeId:  string(target.ID),
		ChunkId: string(meta.ID),
		Digest:  convert.DigestToPB(meta.Digest),
	}
}

func (cu *ChunkUploader) emitProgress() {
	if cu.onProgress != nil {
		cu.onProgress(cu.progress)
	}
}
