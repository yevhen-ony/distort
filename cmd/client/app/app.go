package app

import (
	"errors"
	"fmt"

	"dos/internal/common/connect"
	"dos/internal/common/master/resolve"
	mresolve "dos/internal/common/master/resolve"
	"dos/internal/common/transport/chunkrpc"
	"dos/internal/common/transport/healthrpc"
	"dos/internal/common/transport/masterrouter"
	"dos/internal/services/client/transport"
)

type App struct {
	Config *Config

	Conn    *connect.ConnCache
	Master  *MasterHolder
	ChunkT  *chunkrpc.Transport
	HealthT *healthrpc.HealthTransport
}

func (app *App) Close() error {
	_ = app.Master.router.Close()
	return app.Conn.Close()
}

func NewApp(config *Config) (*App, error) {
	conn := connect.NewConnCache()

	if config == nil {
		return nil, errors.New("missing config")
	}

	master, err := initMasterTransport(config)
	if err != nil {
		return nil, err
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
		Master:  master,
		ChunkT:  chunkT,
		HealthT: healthT,
	}
	return app, nil
}

func (app *App) MasterT() *transport.MasterTransport {
	return app.Master.transport
}

type MasterHolder struct {
	resolver  *resolve.Resolver
	router    *masterrouter.MasterRouter
	transport *transport.MasterTransport
}

func initMasterTransport(config *Config) (*MasterHolder, error) {
	mresolver, err := mresolve.New(&config.Master)
	if err != nil {
		return nil, fmt.Errorf("master resolver init: %w", err)

	}
	mrouter, err := masterrouter.New(mresolver)
	if err != nil {
		return nil, fmt.Errorf("master router init: %w", err)
	}

	mtransport, err := transport.NewMasterTransport(mrouter)
	if err != nil {
		return nil, fmt.Errorf("master transport init: %w", err)
	}

	holder := &MasterHolder{
		resolver:  mresolver,
		router:    mrouter,
		transport: mtransport,
	}
	return holder, nil
}
