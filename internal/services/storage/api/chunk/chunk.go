package chunk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	cpb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/convert"
	"dos/internal/common/dosctx"
	t "dos/internal/common/types"
	"dos/internal/common/utils"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/api"
)

type ChunkDeps struct {
	Identity NodeIdentity
	Storage  ChunkStorage
	Config   ChunkConfig
}

type ChunkServer struct {
	spb.UnimplementedChunkServiceServer

	identity NodeIdentity
	storage  ChunkStorage
	config   ChunkConfig
}

func NewChunkServer(deps ChunkDeps) (*ChunkServer, error) {
	if deps.Identity == nil {
		return nil, errors.New("missing identity")
	}
	if deps.Storage == nil {
		return nil, errors.New("missing storage")
	}
	if deps.Config == nil {
		return nil, errors.New("missing config")
	}

	server := &ChunkServer{
		identity: deps.Identity,
		storage:  deps.Storage,
		config:   deps.Config,
	}
	return server, nil
}

func (srv *ChunkServer) PutChunk(stream spb.ChunkService_PutChunkServer) (err error) {

	defer func() {
		if err != nil {
			slog.ErrorContext(stream.Context(), "put chunk failed", "error", err)
			err = api.ToStatus(err)
		}
	}()

	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive header request: %w", err)
	}

	header := req.GetHeader()

	ctx := dosctx.WithChunkID(stream.Context(), t.ChunkID(header.GetChunkId()))

	if err := srv.validatePutChunkHeader(header); err != nil {
		return err
	}

	slog.DebugContext(ctx, "put chunk requested")

	meta := convert.ChunkMetaFromPB(header)
	session, err := srv.storage.StartUpload(ctx, &meta)
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

	if err := session.Commit(ctx); err != nil {
		return fmt.Errorf("commit upload: %w", err)
	}

	if err := stream.SendAndClose(&spb.PutChunkResponse{}); err != nil {
		return fmt.Errorf("close stream: %w", err)
	}
	return nil
}

func (srv *ChunkServer) GetChunk(
	req *spb.GetChunkRequest,
	stream spb.ChunkService_GetChunkServer,
) (err error) {

	ctx := dosctx.WithChunkID(stream.Context(), t.ChunkID(req.GetChunkId()))

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "get chunk failed", "error", err)
			err = api.ToStatus(err)
		}
	}()
	slog.DebugContext(ctx, "get chunk request")

	release, err := srv.storage.AcquireOpSlot(ctx, time.Second)
	if err != nil {
		return err
	}
	defer release()

	if err = srv.identity.Validate(t.NodeID(req.GetNodeId())); err != nil {
		return err
	}

	chunk, err := srv.storage.LoadChunk(t.ChunkID(req.GetChunkId()))
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

	frames := utils.SplitFrames(chunk.Data, srv.config.FrameSize())
	for _, frame := range frames {
		rsp := &spb.GetChunkResponse{Data: frame}
		if err := stream.Send(rsp); err != nil {
			return fmt.Errorf("send part: %w", err)
		}
	}
	return nil
}

func (srv *ChunkServer) ReplicateChunk(
	ctx context.Context,
	req *spb.ReplicateChunkRequest,
) (rsp *spb.ReplicateChunkResponse, err error) {

	ctx = dosctx.WithChunkID(ctx, t.ChunkID(req.GetChunkId()))
	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "replicate chunk failed", "error", err)
			err = api.ToStatus(err)
		}
	}()
	slog.DebugContext(ctx, "replicate chunk requested")

	err = srv.identity.Validate(t.NodeID(req.GetNodeId()))
	if err != nil {
		return nil, err
	}

	chunkID := t.ChunkID(req.GetChunkId())
	targets := utils.Map(req.GetTargets(), convert.NodeRefFromPB)
	err = srv.storage.ScheduleForwardChunk(ctx, chunkID, targets)
	if err != nil {
		return nil, err
	}

	rsp = &spb.ReplicateChunkResponse{}
	return rsp, nil
}

func (srv *ChunkServer) DeleteChunk(
	ctx context.Context, req *spb.DeleteChunkRequest,
) (rsp *spb.DeleteChunkResponse, err error) {

	ctx = dosctx.WithChunkID(ctx, t.ChunkID(req.GetChunkId()))

	defer func() {
		if err != nil {
			slog.ErrorContext(ctx, "delete chunk failed", "error", err)
			err = api.ToStatus(err)
		}
	}()
	slog.WarnContext(ctx, "delete chunk requested")

	err = srv.identity.Validate(t.NodeID(req.GetNodeId()))
	if err != nil {
		return nil, err
	}

	chunkID := t.ChunkID(req.GetChunkId())
	err = srv.storage.DeleteChunk(ctx, chunkID)
	if err != nil {
		return nil, err
	}

	return &spb.DeleteChunkResponse{}, nil
}

func (srv *ChunkServer) validatePutChunkHeader(header *spb.PutChunkHeader) error {
	if header == nil {
		return fmt.Errorf("missing header: %w", s.ErrInvalidHeader)
	}
	if header.GetChunkId() == "" {
		return fmt.Errorf("missing chunk id: %w", s.ErrInvalidHeader)
	}
	d := header.GetDigest()
	if d == nil {
		return fmt.Errorf("missing digest: %w", s.ErrInvalidHeader)
	}
	if d.GetSize() < 0 {
		return fmt.Errorf("invalid digest size: %w", s.ErrInvalidHeader)
	}

	nodeID := t.NodeID(header.GetNodeId())
	if err := srv.identity.Validate(nodeID); err != nil {
		return err
	}
	return nil
}
