package transport

import (
	"errors"

	"dos/internal/common/connect"
	t "dos/internal/common/types"
)

type StorageTransport struct {
	conn   *connect.ConnCache
	config *StorageTransportConfig
}

func NewStorageTransport(conn *connect.ConnCache, config *StorageTransportConfig) (*StorageTransport, error) {
	if conn == nil {
		return nil, errors.New("missing connection pool")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &StorageTransport{conn: conn, config: config}, nil
}

type TransferSessionOption func(*ChunkTransferSession)

func WithProgressHandler(h ChunkProgressHandler) TransferSessionOption {
	return func(s *ChunkTransferSession) {
		s.onProgress = h
	}
}

func (st *StorageTransport) NewTransferSession(
	nodes []t.NodeRef,
	opts ...TransferSessionOption,
) *ChunkTransferSession {
	session := &ChunkTransferSession{
		config: st.config,
		conn: st.conn,
		nodes: nodes,
	}
	for _, opt := range opts {
		opt(session)
	}
	return session
}

