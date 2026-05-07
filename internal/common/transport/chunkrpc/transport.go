package chunkrpc 

import (
	"errors"

	"dos/internal/common/connect"
	t "dos/internal/common/types"
)

type Transport struct {
	conn   *connect.ConnCache
	config *Config
}

func NewTransport(conn *connect.ConnCache, config *Config) (*Transport, error) {
	if conn == nil {
		return nil, errors.New("missing connection pool")
	}
	if config == nil {
		return nil, errors.New("missing config")
	}
	return &Transport{conn: conn, config: config}, nil
}

type SessionOption func(*Session)

func WithProgressHandler(h ProgressHandler) SessionOption {
	return func(s *Session) {
		s.onProgress = h
	}
}

func (st *Transport) NewTransferSession(nodes []t.NodeRef, opts ...SessionOption) *Session {

	session := &Session{
		config: st.config,
		conn: st.conn,
		nodes: nodes,
	}
	for _, opt := range opts {
		opt(session)
	}
	return session
}

