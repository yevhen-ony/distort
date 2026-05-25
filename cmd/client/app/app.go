package app

import (
	"errors"
	"fmt"

	"dos/internal/common/connect"
	"dos/internal/common/transport/chunkrpc"
	"dos/internal/common/transport/healthrpc"
	"dos/internal/services/client/transport"
)

type App struct {
	Config *Config

	Conn    *connect.ConnCache
	MasterT *transport.MasterTransport
	ChunkT  *chunkrpc.Transport
	HealthT *healthrpc.HealthTransport
}

func (app *App) Close() error {
	if app == nil || app.Conn == nil {
		return nil
	}
	return app.Conn.Close()
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	if config == nil {
		return nil, errors.New("missing config")
	}

	masterT, err := transport.NewMasterTransport(conn, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init master transport: %w", err)
	}

	healthT, err := healthrpc.NewHealthTransport(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init health transport: %w", err)
	}

	chunkT, err := chunkrpc.NewTransport(conn, config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init storage transport: %w", err)
	}

	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init service: %w", err)
	}

	app := &App{
		Config:  config,
		Conn:    conn,
		MasterT: masterT,
		ChunkT:  chunkT,
		HealthT: healthT,
	}
	return app, nil
}
