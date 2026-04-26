package api

import (
	"fmt"
	"io"

	pb "dos/gen/proto/chunk/v1"
	"dos/internal/common/digest"
	t "dos/internal/common/types"
	s "dos/internal/services/storage"
	"dos/internal/services/storage/core"
)

type Server struct {
	pb.UnimplementedChunkServiceServer

	service *core.Service
	config  *ServerConfig
}

func New(service *core.Service, config *ServerConfig) *Server {
	return &Server{
		service: service,
		config: config,
	}
}

func (srv *Server) PutChunk(stream pb.ChunkService_PutChunkServer) (err error) {
	defer func() { err = toStatus(err) }()

	req, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("receive header request: %w", err)
	}

	header := req.GetHeader()
	if err := srv.validatePutChunkHeader(header); err != nil {
		return err
	}

	chunkDesc := &t.ChunkDesc{
		ID: t.ChunkID(header.GetChunkId()),
		Digest: digest.Digest{
			Size:     header.GetChunkSize(),
			Checksum: digest.Checksum(header.GetChecksum()),
		},
	}

	session, err := srv.service.StartUploadSession(chunkDesc)
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

	if err := srv.service.CommitUploadSession(session, chunkDesc); err != nil {
		return fmt.Errorf("commit upload session: %w", err)
	}
	if err := stream.SendAndClose(&pb.PutChunkResponse{}); err != nil {
		return fmt.Errorf("close stream: %w", err)
	}
	return nil
}

func (srv *Server) GetChunk(req *pb.GetChunkRequest, stream pb.ChunkService_GetChunkServer) (err error) {
	defer func() { err = toStatus(err) }()

	chunk, err := srv.service.GetChunk(t.ChunkID(req.GetChunkId()))
	if err != nil {
		return fmt.Errorf("get chunk: %w", err)
	}

	rsp := &pb.GetChunkResponse{
		Header: &pb.GetChunkHeader{
			ChunkId:   string(chunk.ID),
			ChunkSize: chunk.Digest.Size,
			Checksum:  string(chunk.Digest.Checksum),
		},
	}
	if err = stream.Send(rsp); err != nil {
		return fmt.Errorf("send header: %w", err)
	}

	buf := make([]byte, srv.config.FrameSize)
	for {
		n, readErr := chunk.Data.Read(buf)
		if n > 0 {
			rsp := &pb.GetChunkResponse{Data: buf[:n]}
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

func (srv *Server) validatePutChunkHeader(header *pb.PutChunkHeader) error {
	if header == nil {
		return fmt.Errorf("missing header: %w", s.ErrHeaderInvalid)
	}
	if header.GetNodeId() != srv.service.GetServerID() {
		return fmt.Errorf("invalid server id: %w", s.ErrHeaderInvalid)
	}
	return nil
}
