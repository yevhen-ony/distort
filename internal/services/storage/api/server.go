package api

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	cpb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/core"
)

type Config interface {
	FrameSize() int64
}

type Server struct {
	spb.UnimplementedChunkServiceServer

	identity *core.IdentityService
	storage  *core.StorageService
	config   *ServerConfig
}

func New(identity *core.IdentityService, storage *core.StorageService, config *ServerConfig) *Server {
	return &Server{
		identity: identity,
		storage:  storage,
		config:   config,
	}
}

func (srv *Server) PutChunk(stream spb.ChunkService_PutChunkServer) (err error) {
	defer func() {
		if err != nil {
			slog.ErrorContext(stream.Context(), "put chunk failed", "error", err)
			err = toStatus(err)
		}
	}()

	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive header request: %w", err)
	}

	header := req.GetHeader()
	if err := srv.validatePutChunkHeader(header); err != nil {
		return err
	}

	slog.DebugContext(stream.Context(), "put chunk request", "chunk_id", header.GetChunkId())

	chunkDesc := convert.ChunkMetaFromPB(header)
	session, err := srv.storage.StartUploadSession(&chunkDesc)
	if err != nil {
		return fmt.Errorf("start upload session: %w", err)
	}
	defer session.Close()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("receive data request: %w", err)
		}
		if _, err = session.Write(req.GetData()); err != nil {
			return fmt.Errorf("write to upload session: %w", err)
		}
	}

	err = srv.storage.CommitUploadSession(stream.Context(), session, &chunkDesc)
	if err != nil {
		return fmt.Errorf("commit upload session: %w", err)
	}
	if err := stream.SendAndClose(&spb.PutChunkResponse{}); err != nil {
		return fmt.Errorf("close stream: %w", err)
	}
	return nil
}

func (srv *Server) GetChunk(req *spb.GetChunkRequest, stream spb.ChunkService_GetChunkServer) (err error) {
	defer func() {
		if err != nil {
			slog.ErrorContext(
				stream.Context(), "get chunk failed", "chunk_id", req.GetChunkId(), "error", err)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(stream.Context(), "get chunk request", "chunk_id", req.GetChunkId())

	chunk, err := srv.storage.GetChunk(t.ChunkID(req.GetChunkId()))
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	rsp := &spb.GetChunkResponse{
		Header: &spb.GetChunkHeader{
			ChunkId: string(chunk.Meta.ID),
			Digest: &cpb.Digest{
				Checksum: string(chunk.Meta.Digest.Checksum),
				Size:     chunk.Meta.Digest.Size,
			},
		},
	}
	if err = stream.Send(rsp); err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	frames := utils.SplitFrames(chunk.Data, int64(srv.config.FrameSize))
	for _, frame := range frames {
		rsp := &spb.GetChunkResponse{Data: frame}
		if err := stream.Send(rsp); err != nil {
			return fmt.Errorf("send part: %w", err)
		}
	}
	return nil
}

func (srv *Server) ReplicateChunk(
	ctx context.Context, req *spb.ReplicateChunkRequest,
) (rsp *spb.ReplicateChunkResponse, err error) {
	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "chunk_id", req.GetChunkId(), "error", err)
			err = toStatus(err)
		}
	}()
	slog.DebugContext(ctx, "replicate chunk requested", "chunk_id", req.GetChunkId())

	err = srv.identity.Validate(t.NodeID(req.GetNodeId()))
	if err != nil {
		return nil, err
	}

	chunkID := t.ChunkID(req.GetChunkId())
	targets := utils.Map(req.GetTargets(), convert.NodeRefFromPB)
	replCtx := context.WithoutCancel(ctx)	
	go func() {
		_ = srv.storage.ReplicateChunk(replCtx, chunkID, targets)
	}()

	rsp = &spb.ReplicateChunkResponse{}
	return rsp, nil
}

func (srv *Server) DeleteChunk(
	ctx context.Context, req *spb.DeleteChunkRequest,
) (rsp *spb.DeleteChunkResponse, err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "delete chunk failed", "chunk_id", req.GetChunkId(), "error", err)
			err = toStatus(err)
		}
	}()
	slog.WarnContext(ctx, "delete chunk requested", "chunk_id", req.GetChunkId())

	err = srv.identity.Validate(t.NodeID(req.GetNodeId()))
	if err != nil {
		return nil, err
	}

	chunkID := t.ChunkID(req.GetChunkId())
	err = srv.storage.DeleteChunk(ctx, chunkID)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func (srv *Server) validatePutChunkHeader(header *spb.PutChunkHeader) error {
	if header == nil {
		return fmt.Errorf("missing header: %w", s.ErrInvalidHeader)
	}
	nodeID := t.NodeID(header.GetNodeId())
	if err := srv.identity.Validate(nodeID); err != nil {
		return err
	}
	return nil
}
