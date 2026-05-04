package api

import (
	"fmt"
	"io"
	"log/slog"

	cpb "dos/gen/proto/common/v1"
	spb "dos/gen/proto/storage/v1"
	"dos/internal/common/convert"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/core"
)

type Server struct {
	spb.UnimplementedChunkServiceServer

	service *core.Service
	config  *ServerConfig
}

func New(service *core.Service, config *ServerConfig) *Server {
	return &Server{
		service: service,
		config: config,
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

	chunkDesc := convert.ChunkDescFromPB(header) 
	session, err := srv.service.StartUploadSession(&chunkDesc)
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

	if err := srv.service.CommitUploadSession(session, &chunkDesc); err != nil {
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

	chunk, err := srv.service.GetChunk(t.ChunkID(req.GetChunkId()))
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	rsp := &spb.GetChunkResponse{
		Header: &spb.GetChunkHeader{
			ChunkId:   string(chunk.Meta.ID),
			Digest: &cpb.Digest{
				Checksum: string(chunk.Meta.Digest.Checksum),
				Size: chunk.Meta.Digest.Size,
			},
		},
	}
	if err = stream.Send(rsp); err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	buf := make([]byte, srv.config.FrameSize)
	for {
		n, readErr := chunk.Data.Read(buf)
		if n > 0 {
			rsp := &spb.GetChunkResponse{Data: buf[:n]}
			if sendErr := stream.Send(rsp); sendErr != nil {
				return fmt.Errorf("send part: %w", sendErr)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("read part: %w", readErr)
		}
	}
	return nil
}

func (srv *Server) validatePutChunkHeader(header *spb.PutChunkHeader) error {
	if header == nil {
		return fmt.Errorf("missing header: %w", s.ErrInvalidHeader)
	}
	nodeID := t.NodeID(header.GetNodeId())
	if err := srv.service.ValidateNodeID(nodeID); err != nil {
		return err 
	}
	return nil
}
