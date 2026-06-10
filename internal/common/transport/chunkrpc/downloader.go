package chunkrpc

import (
	"bytes"
	"context"
	"fmt"
	"io"

	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/connect"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
)

type ChunkDownloader struct {
	conn *connect.ConnCache

	progress   Progress
	onProgress func(Progress)
}

func (cd *ChunkDownloader) Download(
	ctx context.Context,
	nodeRef t.NodeRef,
	chunkID t.ChunkID,
) (*t.Chunk, error) {

	stream, err := cd.startDownload(ctx, nodeRef, chunkID)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	meta, err := cd.receiveHeader(stream)
	if err != nil {
		return nil, err
	}

	cd.progress = NewProgress(meta, nodeRef)
	cd.emitProgress()
	defer cd.emitProgress()

	data, err := cd.downloadData(stream)
	if err != nil {
		cd.progress.Fail(err.Error())
		return nil, fmt.Errorf("recv data: %w", err)
	}

	chunk := t.NewChunk(chunkID, data)
	if err := meta.Match(chunk.Meta); err != nil {
		cd.progress.Fail(err.Error())
		return nil, err
	}

	cd.progress.Done()
	return &chunk, nil
}

func (cd *ChunkDownloader) startDownload(
	ctx context.Context,
	nodeRef t.NodeRef,
	chunkID t.ChunkID,
) (spb.ChunkService_GetChunkClient, error) {
	conn, err := cd.conn.Get(nodeRef.Addr)
	if err != nil {
		return nil, fmt.Errorf("get conn: %w", err)
	}

	client := spb.NewChunkServiceClient(conn)
	stream, err := client.GetChunk(ctx, &spb.GetChunkRequest{
		NodeId:  string(nodeRef.ID),
		ChunkId: string(chunkID),
	})
	if err != nil {
		return nil, fmt.Errorf("get chunk rpc: %w", err)
	}
	return stream, nil
}

func (cd *ChunkDownloader) receiveHeader(stream spb.ChunkService_GetChunkClient) (t.ChunkMeta, error) {
	rsp, err := stream.Recv()
	if err != nil {
		return t.ChunkMeta{}, fmt.Errorf("recv header: %w", err)
	}
	header := rsp.GetHeader()
	if header == nil {
		return t.ChunkMeta{}, ErrInvalidHeader
	}
	meta := convert.ChunkMetaFromPB(header)
	return meta, nil
}

func (cd *ChunkDownloader) downloadData(stream spb.ChunkService_GetChunkClient) ([]byte, error) {

	var buf bytes.Buffer
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("receive frame: %w", err)
		}

		data := rsp.GetData()
		if data == nil {
			return nil, ErrInvalidData
		}

		buf.Write(rsp.Data)

		cd.progress.SentBytes += int64(len(rsp.Data))
		cd.emitProgress()
	}

	return buf.Bytes(), nil
}

func (cd *ChunkDownloader) emitProgress() {
	if cd.onProgress != nil {
		cd.onProgress(cd.progress)
	}
}
